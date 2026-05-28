package openbanking

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/JoaoNogueiraGit/finances-api/internal/database"
	"github.com/plaid/plaid-go/plaid"
)

type BankItem struct {
	ID          int
	UserID      int
	ItemID      string
	AccessToken string
	SyncCursor  string
}

type SyncResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
}

type Repository struct {
	DB     *database.PostgresDB
	Plaid  *PlaidClient
}

func NewRepository(db *database.PostgresDB, plaidClient *PlaidClient) *Repository {
	return &Repository{DB: db, Plaid: plaidClient}
}

func (r *Repository) SaveItem(ctx context.Context, userID int, itemID, accessToken string) error {
	query := `
		INSERT INTO plaid_items (user_id, item_id, access_token)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, item_id) DO UPDATE SET access_token = EXCLUDED.access_token
	`
	_, err := r.DB.Pool.Exec(ctx, query, userID, itemID, accessToken)
	return err
}

func (r *Repository) ListItems(ctx context.Context, userID int) ([]BankItem, error) {
	query := `
		SELECT id, user_id, item_id, access_token, COALESCE(sync_cursor, '')
		FROM plaid_items
		WHERE user_id = $1
	`
	rows, err := r.DB.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []BankItem
	for rows.Next() {
		var item BankItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.ItemID, &item.AccessToken, &item.SyncCursor); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) HasConnection(ctx context.Context, userID int) (bool, error) {
	var exists bool
	err := r.DB.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM plaid_items WHERE user_id = $1)`,
		userID,
	).Scan(&exists)
	return exists, err
}

func (r *Repository) SyncUserTransactions(ctx context.Context, userID int) (*SyncResult, error) {
	items, err := r.ListItems(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("nenhum banco ligado")
	}

	result := &SyncResult{}
	for _, item := range items {
		plaidTxs, cursor, err := r.Plaid.SyncTransactions(item.AccessToken, item.SyncCursor)
		if err != nil {
			return nil, err
		}

		imported, skipped, err := r.importTransactions(ctx, userID, plaidTxs)
		if err != nil {
			return nil, err
		}
		result.Imported += imported
		result.Skipped += skipped

		_, err = r.DB.Pool.Exec(ctx,
			`UPDATE plaid_items SET sync_cursor = $1 WHERE id = $2`,
			cursor, item.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (r *Repository) importTransactions(ctx context.Context, userID int, plaidTxs []plaid.Transaction) (imported, skipped int, err error) {
	for _, ptx := range plaidTxs {
		if ptx.GetPending() {
			skipped++
			continue
		}

		txType, amount := mapPlaidAmount(ptx.GetAmount())
		category := mapPlaidCategory(ptx)
		description := ptx.GetName()
		if ptx.MerchantName.IsSet() && ptx.GetMerchantName() != "" {
			description = ptx.GetMerchantName()
		}

		currency := "EUR"
		if ptx.IsoCurrencyCode.IsSet() && ptx.GetIsoCurrencyCode() != "" {
			currency = ptx.GetIsoCurrencyCode()
		}

		txDate, err := time.Parse("2006-01-02", ptx.GetDate())
		if err != nil {
			txDate = time.Now()
		}

		query := `
			INSERT INTO transactions (user_id, amount, currency, type, category, description, date, plaid_transaction_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (plaid_transaction_id) DO NOTHING
		`
		tag, err := r.DB.Pool.Exec(ctx, query,
			userID,
			amount,
			currency,
			txType,
			category,
			description,
			txDate,
			ptx.GetTransactionId(),
		)
		if err != nil {
			return imported, skipped, err
		}
		if tag.RowsAffected() > 0 {
			imported++
		} else {
			skipped++
		}
	}
	return imported, skipped, nil
}

// Plaid: positive amount = money out (expense), negative = money in (income).
func mapPlaidAmount(plaidAmount float32) (txType string, amount float64) {
	abs := math.Abs(float64(plaidAmount))
	if plaidAmount > 0 {
		return "expense", abs
	}
	return "income", abs
}

func mapPlaidCategory(ptx plaid.Transaction) string {
	if ptx.PersonalFinanceCategory.IsSet() {
		pfc := ptx.GetPersonalFinanceCategory()
		if primary, ok := pfc.GetPrimaryOk(); ok && primary != nil && *primary != "" {
			return strings.ReplaceAll(*primary, "_", " ")
		}
	}
	if len(ptx.Category) > 0 {
		return ptx.Category[len(ptx.Category)-1]
	}
	return "Banco"
}
