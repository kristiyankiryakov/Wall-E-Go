package data

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/kristiyankiryakov/Wall-E-Go-Common/dto"
)

const DB_TIMEOUT = time.Second * 10

type TransactionRepository interface {
	InsertOne(deposit dto.DepositRequest) (int64, error)
	GetByKey(key string) (int64, error)
}

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewPostgresTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func (r *PostgresTransactionRepository) InsertOne(deposit dto.DepositRequest) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	query := `INSERT INTO transactions (wallet_id, amount, type, idempotency_key) VALUES ($1, $2, 'DEPOSIT', $3) RETURNING id`

	var newID int
	err := r.db.QueryRowContext(ctx, query,
		deposit.WalletID,
		deposit.Amount,
		deposit.IdempotencyKey,
	).Scan(&newID)

	if err != nil {
		log.Println(err)
		return 0, err
	}

	return int64(newID), nil
}

func (r *PostgresTransactionRepository) GetByKey(key string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	query := `SELECT id FROM transactions WHERE idempotency_key = $1`

	var existingID int
	err := r.db.QueryRowContext(ctx, query, key).Scan(&existingID)
	if err != sql.ErrNoRows {
		log.Println(err)
		return 0, err
	}

	return int64(existingID), nil
}
