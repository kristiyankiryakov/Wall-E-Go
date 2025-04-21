package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const COMPLETED = "COMPLETED"
const PROCESSING = "PROCESSING"

type Repository interface {
	GetAuthorizedTransactionsForPeriod(ctx context.Context, tx *sql.Tx, start, end string) ([]*Transaction, error)
	UpdateTransactionsStatus(ctx context.Context, tx *sql.Tx, ids []string, status string) error
	GetTransactionsForProcessing(ctx context.Context, start, end string) ([]*Transaction, error)
}

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewPostgresTransactionRepository(db *sql.DB) Repository {
	return &PostgresTransactionRepository{db: db}
}

// GetAuthorizedTransactionsForPeriod fetches transactions with locking
func (r *PostgresTransactionRepository) GetAuthorizedTransactionsForPeriod(ctx context.Context, tx *sql.Tx, start, end string) ([]*Transaction, error) {
	rows, err := tx.QueryContext(
		ctx,
		"SELECT id, amount, status, type FROM transactions WHERE status = $1 AND created_at BETWEEN $2 AND $3 FOR UPDATE",
		COMPLETED,
		start,
		end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.Amount,
			&transaction.Status,
			&transaction.Type,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, &transaction)
	}

	return transactions, nil
}

// UpdateTransactionsStatus updates the status of specified transactions
func (r *PostgresTransactionRepository) UpdateTransactionsStatus(ctx context.Context, tx *sql.Tx, ids []string, status string) error {
	if len(ids) == 0 {
		return nil
	}

	// Build query with individual placeholders
	query := "UPDATE transactions SET status = $1 WHERE id IN ("
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, status) // $1 is status

	for i, id := range ids {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("$%d", i+2) // $2, $3, etc.
		args = append(args, id)
	}
	query += ")"

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

// GetTransactionsForProcessing combines both operations in a single transaction and returns the transactions
func (r *PostgresTransactionRepository) GetTransactionsForProcessing(ctx context.Context, start, end string) ([]*Transaction, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Step 1: Get transactions with lock
	transactions, err := r.GetAuthorizedTransactionsForPeriod(ctx, tx, start, end)
	if err != nil {
		return nil, err
	}

	// Step 2: Update their status if we have transactions
	if len(transactions) > 0 {
		ids := make([]string, len(transactions))
		for i, t := range transactions {
			ids[i] = t.ID
			// Update in-memory objects too
			t.Status = PROCESSING
		}

		err = r.UpdateTransactionsStatus(ctx, tx, ids, PROCESSING)
		if err != nil {
			return nil, err
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return transactions, nil
}
