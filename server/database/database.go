package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

type Database struct {
	Client *sqlx.DB
}

func NewDatabase() (*Database, error) {
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("SSL_MODE"),
	)

	db, err := sqlx.Connect("postgres", connectionString)

	if err != nil {
		println("DB_CONNECTION_ERROR", err)
		return &Database{}, fmt.Errorf("error connecting to database: %w", err)
	}
	return &Database{Client: db}, nil
}

func (d *Database) Ping(ctx context.Context) error {
	return d.Client.DB.PingContext(ctx)
}
