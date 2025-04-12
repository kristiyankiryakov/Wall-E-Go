package repositories

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"
	"transaction/internal/domain/entities"
)

const DB_TIMEOUT = time.Second * 10

const TRANSACTION_STATUS_PENDING entities.TransactionStatus = "PENDING"

type TransactionRepository interface {
	InsertOne(tx *sql.Tx, deposit entities.Transaction) (string, error)
	GetTxByKey(tx *sql.Tx, key string) (string, error)
	BeginTx(ctx context.Context) (*sql.Tx, error)

	UpdateStatusBatch(ctx context.Context, transactionIDs []string, status entities.TransactionStatus) error
	UpdateStatusConcurrently(ctx context.Context, transactionIDs []string, status entities.TransactionStatus) error
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

func (r *PostgresTransactionRepository) InsertOne(tx *sql.Tx, deposit entities.Transaction) (string, error) {
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

// UpdateStatusBatch updates status for multiple transactions in a single database operation
func (r *PostgresTransactionRepository) UpdateStatusBatch(ctx context.Context, transactionIDs []string, status entities.TransactionStatus) error {
	if len(transactionIDs) == 0 {
		return nil
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, DB_TIMEOUT)
	defer cancel()

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin transaction for batch status update: %v", err)
		return err
	}
	defer tx.Rollback()

	// Create a parameterized query for batch update
	// This uses PostgreSQL's ANY operator with an array
	query := `UPDATE transactions SET status = $1, updated_at = $2 WHERE id = ANY($3) AND status = $4`

	// Execute update
	_, err = tx.ExecContext(ctx, query, string(status), time.Now(), transactionIDs, TRANSACTION_STATUS_PENDING)
	if err != nil {
		log.Printf("Failed to batch update transaction statuses: %v", err)
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit batch update transaction: %v", err)
		return err
	}

	return nil
}

// UpdateStatusConcurrently updates transaction statuses concurrently using multiple goroutines
func (r *PostgresTransactionRepository) UpdateStatusConcurrently(ctx context.Context, transactionIDs []string, status entities.TransactionStatus) error {
	if len(transactionIDs) == 0 {
		return nil
	}

	// If only a few transactions, just use the batch method
	if len(transactionIDs) < 5 {

		return r.UpdateStatusBatch(
			ctx,
			transactionIDs,
			status,
		)

	}

	// For larger batches, process in parallel chunks
	const chunkSize = 50

	// Create chunks of transaction IDs
	chunks := make([][]string, 0)
	for i := 0; i < len(transactionIDs); i += chunkSize {
		end := min(i+chunkSize, len(transactionIDs))
		chunks = append(chunks, transactionIDs[i:end])
	}

	// Process chunks concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, len(chunks))

	for _, chunk := range chunks {
		wg.Add(1)
		go func(ids []string) {
			defer wg.Done()

			// Create a new context for this goroutine
			chunkCtx, cancel := context.WithTimeout(ctx, DB_TIMEOUT)
			defer cancel()

			if err := r.UpdateStatusBatch(chunkCtx, ids, status); err != nil {
				errChan <- err
			}
		}(chunk)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// Helper function for determining min value
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
