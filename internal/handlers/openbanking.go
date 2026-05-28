package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/JoaoNogueiraGit/finances-api/internal/auth"
	"github.com/JoaoNogueiraGit/finances-api/internal/openbanking"
)

type OpenBankingHandler struct {
	Repo *openbanking.Repository
}

func NewOpenBankingHandler(repo *openbanking.Repository) *OpenBankingHandler {
	return &OpenBankingHandler{Repo: repo}
}

func (h *OpenBankingHandler) userID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	return userID, ok
}

// CreateLinkToken handles POST /banking/link-token
func (h *OpenBankingHandler) CreateLinkToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		http.Error(w, "Erro ao identificar utilizador", http.StatusInternalServerError)
		return
	}

	linkToken, err := h.Repo.Plaid.CreateLinkToken(strconv.Itoa(userID))
	if err != nil {
		log.Printf("ERRO NO PLAID (link-token): %v\n", err)
		http.Error(w, "Erro ao gerar token do banco", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"link_token": linkToken,
	})
}

// ExchangeToken handles POST /banking/exchange-token
func (h *OpenBankingHandler) ExchangeToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		http.Error(w, "Erro ao identificar utilizador", http.StatusInternalServerError)
		return
	}

	var body struct {
		PublicToken string `json:"public_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PublicToken == "" {
		http.Error(w, "public_token é obrigatório", http.StatusBadRequest)
		return
	}

	accessToken, itemID, err := h.Repo.Plaid.ExchangePublicToken(body.PublicToken)
	if err != nil {
		log.Printf("ERRO NO PLAID (exchange): %v\n", err)
		http.Error(w, "Erro ao ligar conta bancária", http.StatusInternalServerError)
		return
	}

	if err := h.Repo.SaveItem(r.Context(), userID, itemID, accessToken); err != nil {
		log.Printf("ERRO NA BD (plaid_items): %v\n", err)
		http.Error(w, "Erro ao guardar ligação bancária", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "connected",
		"item_id": itemID,
	})
}

// SyncTransactions handles POST /banking/sync
func (h *OpenBankingHandler) SyncTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		http.Error(w, "Erro ao identificar utilizador", http.StatusInternalServerError)
		return
	}

	result, err := h.Repo.SyncUserTransactions(r.Context(), userID)
	if err != nil {
		log.Printf("ERRO NO PLAID (sync): %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetStatus handles GET /banking/status
func (h *OpenBankingHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(r)
	if !ok {
		http.Error(w, "Erro ao identificar utilizador", http.StatusInternalServerError)
		return
	}

	connected, err := h.Repo.HasConnection(r.Context(), userID)
	if err != nil {
		http.Error(w, "Erro ao verificar ligação bancária", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{
		"connected": connected,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
