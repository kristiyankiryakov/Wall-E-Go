package db

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
	"time"
)

var DB *sql.DB

func InitDB(dbUrl string) error {
	var err error
	DB, err = sql.Open("pgx", dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return err
	}

	// Set connection pool parameters
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(1 * time.Hour)
	DB.SetConnMaxIdleTime(30 * time.Minute)

	if err := DB.Ping(); err != nil {
		return err
	}
	return nil
}
