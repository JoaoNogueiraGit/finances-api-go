// internal/handlers/user.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/JoaoNogueiraGit/finances-api/internal/auth"
	"github.com/JoaoNogueiraGit/finances-api/internal/database"
	"github.com/JoaoNogueiraGit/finances-api/internal/models"
)

type UserHandler struct {
	DB *database.PostgresDB
}

func NewUserHandler(db *database.PostgresDB) *UserHandler {
	return &UserHandler{DB: db}
}

// Register lida com a criação de uma nova conta (POST /register)
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 1. Encriptar a password usando o pacote que criámos
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Erro ao processar password", http.StatusInternalServerError)
		return
	}

	// 2. Guardar na base de dados
	query := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`
	err = h.DB.Pool.QueryRow(r.Context(), query, user.Name, user.Email, hashedPassword).Scan(&user.ID)
	if err != nil {
		log.Printf("ERRO NA BD (REGISTER): %v\n", err)
		http.Error(w, "Erro ao criar utilizador (email já existe?)", http.StatusConflict)
		return
	}

	// 3. Limpar a password da struct antes de devolver a resposta ao cliente
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Login verifica as credenciais e devolve o Token JWT (POST /login)
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Usamos uma struct anónima rápida só para ler o pedido
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 1. Ir à base de dados buscar o ID e a Password Encriptada associada a este email
	var dbUser models.User
	query := `SELECT id, password FROM users WHERE email = $1`
	err := h.DB.Pool.QueryRow(r.Context(), query, credentials.Email).Scan(&dbUser.ID, &dbUser.Password)
	if err != nil {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	// 2. Comparar a password que o utilizador enviou com o Hash guardado
	if !auth.CheckPassword(credentials.Password, dbUser.Password) {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	// 3. Sucesso! Gerar o Token JWT
	token, err := auth.GenerateToken(dbUser.ID)
	if err != nil {
		http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}

	// 4. Devolver o token num JSON simples
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
