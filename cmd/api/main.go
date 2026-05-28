// cmd/api/main.go

// @title Finances API
// @version 1.0
// @description API para gestão de finanças pessoais em Go.
// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Escreve 'Bearer O_TEU_TOKEN' na caixa abaixo.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/JoaoNogueiraGit/finances-api/docs"
	"github.com/JoaoNogueiraGit/finances-api/internal/auth"
	"github.com/JoaoNogueiraGit/finances-api/internal/database"
	"github.com/JoaoNogueiraGit/finances-api/internal/handlers"
	"github.com/JoaoNogueiraGit/finances-api/internal/middleware"
	"github.com/JoaoNogueiraGit/finances-api/internal/openbanking"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {

	// 1. Carregar variáveis de ambiente
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: Ficheiro .env não encontrado. A usar variáveis de sistema.")
	}

	// Buscar a string de conexão do .env
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL não está definida no ficheiro .env")
	}

	db, err := database.NewConnection(dbURL)
	if err != nil {
		log.Fatalf("Erro critico na base de dados: %v", err)
	}

	plaidClient, err := openbanking.NewPlaidClient()
	if err != nil {
		log.Fatalf("Falha ao iniciar o cliente Plaid: %v", err)
	}

	defer db.Close()

	txHandler := handlers.NewTransactionHandler(db)
	userHandler := handlers.NewUserHandler(db)
	bankingRepo := openbanking.NewRepository(db, plaidClient)
	bankingHandler := handlers.NewOpenBankingHandler(bankingRepo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. Inicializar o Router (Go 1.22+)
	mux := http.NewServeMux()

	// 3. Definir as Rotas (Endpoints)
	mux.HandleFunc("GET /health", healthCheckHandler)

	// Rotas de Autenticação
	mux.HandleFunc("POST /register", userHandler.Register)
	mux.HandleFunc("POST /login", userHandler.Login)

	// Rotas de Transação
	mux.HandleFunc("POST /transactions", auth.AuthMiddleware(txHandler.CreateTransaction))
	mux.HandleFunc("GET /transactions", auth.AuthMiddleware(txHandler.GetTransactions))
	mux.HandleFunc("PUT /transactions/{id}", auth.AuthMiddleware(txHandler.UpdateTransaction))
	mux.HandleFunc("DELETE /transactions/{id}", auth.AuthMiddleware(txHandler.DeleteTransaction))

	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	mux.HandleFunc("POST /banking/link-token", auth.AuthMiddleware(bankingHandler.CreateLinkToken))
	mux.HandleFunc("POST /banking/exchange-token", auth.AuthMiddleware(bankingHandler.ExchangeToken))
	mux.HandleFunc("POST /banking/sync", auth.AuthMiddleware(bankingHandler.SyncTransactions))
	mux.HandleFunc("GET /banking/status", auth.AuthMiddleware(bankingHandler.GetStatus))

	// 4. Iniciar o Servidor
	fmt.Printf("Servidor a correr na porta %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, middleware.CORS(mux)))
}

// Handler de teste para verificar se a API está viva
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("API está online e saudável!"))
}
