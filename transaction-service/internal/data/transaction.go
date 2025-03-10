package data

import (
	"context"
	"database/sql"
	"log"
	"time"
)

const DB_TIMEOUT = time.Second * 10

type TransactionRepository interface {
	InsertOne(tx *sql.Tx, deposit TransactionRequest) (string, error)
	GetTxByKey(tx *sql.Tx, key string) (string, error)
	BeginTx(ctx context.Context) (*sql.Tx, error)
}

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewPostgresTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func (r *PostgresTransactionRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	return tx, err
}

func (r *PostgresTransactionRepository) InsertOne(tx *sql.Tx, deposit TransactionRequest) (string, error) {
	query := `INSERT INTO transactions (wallet_id, amount, type, idempotency_key) VALUES ($1, $2, 'DEPOSIT', $3) RETURNING id`

	var newID string
	err := tx.QueryRow(query,
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

func (r *PostgresTransactionRepository) GetTxByKey(tx *sql.Tx, key string) (string, error) {
	query := `SELECT id FROM transactions WHERE idempotency_key = $1`

	var existingID string
	err := tx.QueryRow(query, key).Scan(&existingID)
	if err != sql.ErrNoRows {
		log.Println(err)
		return "", err
	}

	return existingID, nil
}
