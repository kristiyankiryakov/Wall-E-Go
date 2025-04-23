package db

import (
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
	"time"
)

// NewDB creates a new sqlx.DB connection that can be injected into services
func NewDB(dbUrl string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return nil, err
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)

	return db, nil
}
