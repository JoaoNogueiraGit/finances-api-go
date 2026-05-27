// internal/database/postgres.go
package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDB funciona como um "wrapper" (invólucro) para o pool de conexões.
// Decisão: Em vez de expor o 'pgxpool.Pool' diretamente para toda a aplicação,
// colocamo-lo dentro desta struct. Isto dá-nos flexibilidade. Se no futuro
// quisermos adicionar métricas ou mudar algo na base de dados, o resto da app não sofre.
type PostgresDB struct {
	Pool *pgxpool.Pool
}

// NewConnection inicializa, configura e testa a conexão com o PostgreSQL.
func NewConnection(databaseURL string) (*PostgresDB, error) {
	// Decisão: Usamos Context com Timeout.
	// Na web, nunca queremos que uma operação fique "bloqueada" para sempre.
	// Se a base de dados estiver em baixo, queremos que a app desista após 5 segundos.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Boa prática: limpa os recursos do contexto assim que a função terminar.

	// Decisão: Configuração via URL string (lida do .env).
	// O pgxpool.ParseConfig analisa a string e cria um objeto de configuração padrão.
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar string de conexão: %w", err)
	}

	// Criar o Pool de conexões.
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar pool de conexões: %w", err)
	}

	// Decisão: Fazer um Ping explícito.
	// O comando 'NewWithConfig' apenas cria a estrutura do pool, ele não garante
	// que a password está certa ou que o servidor está online. O Ping força uma
	// comunicação real para validar a ligação antes da API começar a aceitar utilizadores.
	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("não foi possível comunicar com a base de dados: %w", err)
	}

	log.Println("Conexão com a base de dados estabelecida com sucesso!")

	// Retornamos um ponteiro (*PostgresDB) para evitar copiar a struct na memória
	return &PostgresDB{Pool: pool}, nil
}

// Close fecha o pool de conexões de forma segura quando a aplicação for desligada.
func (db *PostgresDB) Close() {
	db.Pool.Close()
}
