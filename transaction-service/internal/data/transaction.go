package data

import (
	"context"
	"database/sql"
	"log"
	"time"
)

const DB_TIMEOUT = time.Second * 10

type TransactionRepository interface {
	InsertOne(deposit TransactionRequest) (string, error)
	GetByKey(key string) (string, error)
}

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewPostgresTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func (r *PostgresTransactionRepository) InsertOne(deposit TransactionRequest) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	query := `INSERT INTO transactions (wallet_id, amount, type, idempotency_key) VALUES ($1, $2, 'DEPOSIT', $3) RETURNING id`

	var newID string
	err := r.db.QueryRowContext(ctx, query,
		deposit.WalletID,
		deposit.Amount,
		deposit.IdempotencyKey,
	).Scan(&newID)

	if err != nil {
		log.Println(err)
		return "", err
	}

	return newID, nil
}

func (r *PostgresTransactionRepository) GetByKey(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	query := `SELECT id FROM transactions WHERE idempotency_key = $1`

	var existingID string
	err := r.db.QueryRowContext(ctx, query, key).Scan(&existingID)
	if err != sql.ErrNoRows {
		log.Println(err)
		return "", err
	}

	return existingID, nil
}
