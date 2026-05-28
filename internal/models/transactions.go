// internal/models/transaction.go
package models

import "time"

// Transaction representa uma transação financeira.
// As tags `json:"..."` dizem ao Go como traduzir esta struct para JSON
// quando respondemos a um pedido HTTP.
type Transaction struct {
	ID                  int       `json:"id"`
	UserID              int       `json:"user_id"`
	Amount              float64   `json:"amount"`
	Currency            string    `json:"currency"`
	Type                string    `json:"type"` // "income" ou "expense"
	Category            string    `json:"category"`
	Description         string    `json:"description"`
	Date                time.Time `json:"date"`
	PlaidTransactionID  *string   `json:"plaid_transaction_id,omitempty"`
}
