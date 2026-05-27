// internal/auth/middleware.go
package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// ContextKey é um tipo customizado. É uma boa prática em Go usar tipos customizados
// para chaves de contexto para evitar colisões com pacotes de terceiros.
type ContextKey string

const UserIDKey ContextKey = "userID"

// AuthMiddleware é o nosso segurança. Recebe um Handler e devolve outro Handler.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Ler o cabeçalho "Authorization"
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Acesso negado: Token não fornecido", http.StatusUnauthorized)
			return
		}

		// 2. O formato standard é "Bearer <token>". Vamos separar a string.
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Acesso negado: Formato de token inválido", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		secretKey := os.Getenv("JWT_SECRET")

		// 3. Fazer o Parse e Validar a assinatura do Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Garantir que o algoritmo usado é o que esperamos (Segurança)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de assinatura inesperado")
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Acesso negado: Token inválido ou expirado", http.StatusUnauthorized)
			return
		}

		// 4. Extrair os dados (Claims) do interior do Token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Erro ao ler dados do token", http.StatusUnauthorized)
			return
		}

		// O pacote JWT lê todos os números como float64, por isso temos de converter para int
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "ID de utilizador inválido no token", http.StatusUnauthorized)
			return
		}

		// 5. MÁGICA: Guardar o userID no Contexto do request e passar para a próxima função!
		ctx := context.WithValue(r.Context(), UserIDKey, int(userIDFloat))
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
