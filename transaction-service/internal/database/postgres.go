package database

import (
	"database/sql"
	"log"
	"time"
)

func OpenDb(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	db.SetMaxOpenConns(10)

	err = db.Ping()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return db, nil
}

func ConnectToDB(dsn string) *sql.DB {
	counts := 0

	for {
		connection, err := OpenDb(dsn)
		if err != nil {
			log.Println(err)
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
