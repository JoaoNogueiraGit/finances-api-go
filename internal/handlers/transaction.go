// internal/handlers/transaction.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/JoaoNogueiraGit/finances-api/internal/auth"
	"github.com/JoaoNogueiraGit/finances-api/internal/database"
	"github.com/JoaoNogueiraGit/finances-api/internal/models"
)

// TransactionHandler guarda a conexão à base de dados para ser usada nas rotas
type TransactionHandler struct {
	DB *database.PostgresDB
}

// NewTransactionHandler é o construtor
func NewTransactionHandler(db *database.PostgresDB) *TransactionHandler {
	return &TransactionHandler{DB: db}
}

// CreateTransaction lida com o POST /transactions
// @Summary Criar uma nova transação
// @Description Adiciona uma receita ou despesa associada ao utilizador logado
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transaction body models.Transaction true "Dados da transação"
// @Success 201 {object} models.Transaction
// @Failure 400 {string} string "JSON inválido"
// @Failure 401 {string} string "Não autorizado"
// @Failure 500 {string} string "Erro interno"
// @Router /transactions [post]
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	// 1. Dizer ao cliente que vamos responder em JSON
	w.Header().Set("Content-Type", "application/json")

	// 2. Ler o JSON que vem no corpo do pedido e convertê-lo para a nossa struct
	var tx models.Transaction
	err := json.NewDecoder(r.Body).Decode(&tx)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 3. Inserir na Base de Dados usando a nossa conexão injetada (h.DB)
	// A cláusula RETURNING id permite-nos saber qual foi o ID gerado pelo PostgreSQL
	query := `
		INSERT INTO transactions (user_id, amount, currency, type, category, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, date
	`

	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		http.Error(w, "Erro ao identificar utilizador", http.StatusInternalServerError)
		return
	}

	tx.UserID = userID

	err = h.DB.Pool.QueryRow(r.Context(), query,
		tx.UserID, tx.Amount, tx.Currency, tx.Type, tx.Category, tx.Description,
	).Scan(&tx.ID, &tx.Date)

	if err != nil {
		log.Printf("ERRO NA BASE DE DADOS: %v\n", err)
		http.Error(w, "Erro ao guardar transação", http.StatusInternalServerError)
		return
	}

	// 4. Devolver a transação criada (com o novo ID e Data) como resposta com status 201 (Created)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tx)
}

// GetTransactions lida com o GET /transactions
func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		http.Error(w, "Erro ao identificar utilizador", http.StatusInternalServerError)
		return
	}

	// 1. Fazer a query à base de dados (Nota: usamos Query e não QueryRow porque esperamos vários resultados)
	query := `
		SELECT id, amount, currency, type, category, description, date 
		FROM transactions 
		WHERE user_id = $1 
		ORDER BY date DESC
	`

	rows, err := h.DB.Pool.Query(r.Context(), query, userID)
	if err != nil {
		http.Error(w, "Erro ao buscar transações", http.StatusInternalServerError)
		return
	}
	// IMPORTANTE: Garantir que as rows são fechadas no fim para não esgotar as conexões do Pool!
	defer rows.Close()

	// 2. Criar uma slice (lista) vazia de transações
	var transactions []models.Transaction

	// 3. Iterar sobre as linhas devolvidas pela base de dados
	for rows.Next() {
		var tx models.Transaction
		tx.UserID = userID // Já sabemos qual é o user

		// O Scan copia os dados da linha atual para as variáveis da nossa struct
		err := rows.Scan(
			&tx.ID, &tx.Amount, &tx.Currency, &tx.Type,
			&tx.Category, &tx.Description, &tx.Date,
		)
		if err != nil {
			http.Error(w, "Erro ao processar dados", http.StatusInternalServerError)
			return
		}

		// Adicionar a transação processada à nossa lista
		transactions = append(transactions, tx)
	}

	// 4. Verificar se ocorreu algum erro durante a iteração (boa prática em Go)
	if err = rows.Err(); err != nil {
		http.Error(w, "Erro na leitura dos dados", http.StatusInternalServerError)
		return
	}

	// 5. Se a lista estiver vazia, garantir que devolvemos [] em vez de null no JSON
	if transactions == nil {
		transactions = []models.Transaction{}
	}

	// 6. Devolver a resposta em JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}

// DeleteTransaction lida com o DELETE /transactions/{id}
func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	// 1. Capturar o ID que vem no URL (ex: /transactions/5)
	// Graças ao Go 1.22, isto é super simples usando r.PathValue
	txID := r.PathValue("id")

	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		http.Error(w, "Erro ao identificar utilizador", http.StatusInternalServerError)
		return
	}

	// 2. Apagar da base de dados garantindo que a transação pertence a este utilizador
	query := `DELETE FROM transactions WHERE id = $1 AND user_id = $2`

	// Usamos Exec porque não esperamos que devolva linhas, apenas confirme a ação
	commandTag, err := h.DB.Pool.Exec(r.Context(), query, txID, userID)
	if err != nil {
		log.Printf("ERRO NA BD (DELETE): %v\n", err)
		http.Error(w, "Erro ao apagar transação", http.StatusInternalServerError)
		return
	}

	// 3. Verificar se alguma linha foi realmente apagada
	if commandTag.RowsAffected() == 0 {
		http.Error(w, "Transação não encontrada", http.StatusNotFound)
		return
	}

	// 4. Sucesso: Devolver Status 204 (No Content) - Padrão REST para DELETE
	w.WriteHeader(http.StatusNoContent)
}

// UpdateTransaction lida com o PUT /transactions/{id}
func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	// 1. Capturar o ID do URL
	txID := r.PathValue("id")
	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		http.Error(w, "Erro ao identificar utilizador", http.StatusInternalServerError)
		return
	}
	// 2. Ler o JSON com os novos dados
	var tx models.Transaction
	err := json.NewDecoder(r.Body).Decode(&tx)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 3. Atualizar na base de dados
	// Usamos RETURNING para o PostgreSQL nos devolver a data original e confirmarmos o ID
	query := `
		UPDATE transactions
		SET amount = $1, currency = $2, type = $3, category = $4, description = $5
		WHERE id = $6 AND user_id = $7
		RETURNING id, date
	`

	err = h.DB.Pool.QueryRow(r.Context(), query,
		tx.Amount, tx.Currency, tx.Type, tx.Category, tx.Description,
		txID, userID,
	).Scan(&tx.ID, &tx.Date)

	if err != nil {
		log.Printf("ERRO NA BD (UPDATE): %v\n", err)
		// Se o ID não existir, o PostgreSQL não devolve linhas e o Scan dá erro
		http.Error(w, "Erro ao atualizar ou transação não encontrada", http.StatusNotFound)
		return
	}

	// 4. Preencher o userID na struct para a resposta ficar completa
	tx.UserID = userID

	// 5. Devolver a transação atualizada
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tx)
}
