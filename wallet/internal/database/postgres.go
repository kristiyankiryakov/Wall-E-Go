package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	db.SetMaxOpenConns(25)

	err = db.Ping()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return db, nil
}

func ConnectToDB() *sql.DB {
	dsn := os.Getenv("DSN")
	counts := 0

	for {
		connection, err := OpenDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready...")
			counts++
		} else {
			log.Println("Connected to Postgres")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds...")
		time.Sleep(2 * time.Second)
		continue
	}
}
