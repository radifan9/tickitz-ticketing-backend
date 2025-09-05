package configs

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Initialization of DB
func InitDB() (*pgxpool.Pool, error) {
	pgUser := os.Getenv("POSTGRES_USER")
	pgPass := os.Getenv("POSTGRES_PASSWORD")
	pgHost := os.Getenv("POSTGRES_HOST")
	pgPort := os.Getenv("POSTGRES_PORT")
	pgDB := os.Getenv("POSTGRES_DB")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pgUser, pgPass, pgHost, pgPort, pgDB)

	return pgxpool.New(context.Background(), connString)
}

// Test DB Connection
func TestDB(db *pgxpool.Pool) error {
	return db.Ping(context.Background())
}
