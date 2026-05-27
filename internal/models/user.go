// internal/models/user.go
package models

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	// O 'omitempty' garante que, ao devolvermos o utilizador em JSON,
	// a password não é enviada acidentalmente para o cliente!
	Password string `json:"password,omitempty"`
}
