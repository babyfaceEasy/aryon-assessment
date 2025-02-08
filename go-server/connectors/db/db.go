package db

import (
	"database/sql"
	"log"

	"github.com/jackc/pgx/v5"
)

/*
// NewPostgresStorage initializes a PostgreSQL database connection
func NewPostgresStorage(connString string) (*pgx.Conn, error) {
	db, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		return nil, err
	}

	return db, nil
}
*/

func NewPGXStorage(cfg pgx.ConnConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.ConnString())
	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}
