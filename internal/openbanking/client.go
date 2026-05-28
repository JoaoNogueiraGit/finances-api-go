package openbanking

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/plaid/plaid-go/plaid"
)

type PlaidClient struct {
	Client *plaid.PlaidApiService
}

func NewPlaidClient() (*PlaidClient, error) {
	clientID := os.Getenv("PLAID_CLIENT_ID")
	secret := os.Getenv("PLAID_SECRET")
	env := os.Getenv("PLAID_ENV")

	if clientID == "" || secret == "" {
		return nil, fmt.Errorf("credenciais do Plaid não configuradas no .env")
	}

	var plaidEnv plaid.Environment
	switch env {
	case "sandbox":
		plaidEnv = plaid.Sandbox
	case "development":
		plaidEnv = plaid.Development
	case "production":
		plaidEnv = plaid.Production
	default:
		return nil, fmt.Errorf("ambiente Plaid inválido: usa 'sandbox', 'development' ou 'production'")
	}

	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	configuration.AddDefaultHeader("PLAID-SECRET", secret)
	configuration.UseEnvironment(plaidEnv)

	client := plaid.NewAPIClient(configuration)
	return &PlaidClient{Client: client.PlaidApi}, nil
}

func countryCodesFromEnv() []plaid.CountryCode {
	raw := os.Getenv("PLAID_COUNTRY_CODES")
	if raw == "" {
		raw = "PT"
	}
	parts := strings.Split(raw, ",")
	codes := make([]plaid.CountryCode, 0, len(parts))
	for _, part := range parts {
		code := strings.TrimSpace(strings.ToUpper(part))
		if code != "" {
			codes = append(codes, plaid.CountryCode(code))
		}
	}
	return codes
}

func linkClientName() string {
	if name := os.Getenv("PLAID_CLIENT_NAME"); name != "" {
		return name
	}
	return "Finances"
}

func linkLanguage() string {
	if lang := os.Getenv("PLAID_LANGUAGE"); lang != "" {
		return lang
	}
	return "pt"
}

// CreateLinkToken generates the token used to open Plaid Link in the browser.
func (pc *PlaidClient) CreateLinkToken(userID string) (string, error) {
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: userID,
	}

	request := plaid.NewLinkTokenCreateRequest(
		linkClientName(),
		linkLanguage(),
		countryCodesFromEnv(),
		user,
	)
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})

	if redirect := os.Getenv("PLAID_REDIRECT_URI"); redirect != "" {
		request.SetRedirectUri(redirect)
	}

	ctx := context.Background()
	resp, _, err := pc.Client.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		return "", fmt.Errorf("erro ao gerar Link Token do Plaid: %w", err)
	}

	return resp.GetLinkToken(), nil
}

// ExchangePublicToken converts the one-time public token from Link into a permanent access token.
func (pc *PlaidClient) ExchangePublicToken(publicToken string) (accessToken, itemID string, err error) {
	req := plaid.NewItemPublicTokenExchangeRequest(publicToken)

	ctx := context.Background()
	resp, _, err := pc.Client.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*req).Execute()
	if err != nil {
		return "", "", fmt.Errorf("erro ao trocar public token: %w", err)
	}

	return resp.GetAccessToken(), resp.GetItemId(), nil
}

// SyncTransactions pulls added transactions from Plaid using /transactions/sync.
func (pc *PlaidClient) SyncTransactions(accessToken, cursor string) (added []plaid.Transaction, nextCursor string, err error) {
	ctx := context.Background()
	req := plaid.NewTransactionsSyncRequest(accessToken)
	if cursor != "" {
		req.SetCursor(cursor)
	}

	for {
		resp, _, err := pc.Client.TransactionsSync(ctx).TransactionsSyncRequest(*req).Execute()
		if err != nil {
			return nil, "", fmt.Errorf("erro ao sincronizar transações: %w", err)
		}

		added = append(added, resp.GetAdded()...)

		if !resp.GetHasMore() {
			return added, resp.GetNextCursor(), nil
		}

		req.SetCursor(resp.GetNextCursor())
	}
}
