// internal/auth/auth.go
package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword recebe uma password em texto e devolve o hash seguro
func HashPassword(password string) (string, error) {
	// O número 14 é o "custo". Quanto maior, mais lento é a gerar (e mais difícil de hackear).
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPassword compara uma password em texto com o hash guardado na base de dados
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil // Se err for nulo, a password está correta
}

// GenerateToken cria o token JWT para um utilizador específico
func GenerateToken(userID int) (string, error) {
	// Buscar o nosso segredo ao ficheiro .env
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "chave_de_backup_insegura" // Apenas para evitar crashes se esqueceres do .env
	}

	// 1. Criar os "Claims" (os dados que vão dentro do token)
	// Registamos o ID do user e a data de expiração (ex: válido por 24 horas)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	// 2. Criar o token usando o algoritmo de assinatura HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 3. Assinar o token com a nossa chave secreta
	return token.SignedString([]byte(secretKey))
}
