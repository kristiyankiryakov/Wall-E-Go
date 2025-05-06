package wallet

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

const DB_TIMEOUT = time.Second * 10

type WalletRepository interface {
	CreateWallet(wallet Wallet) (string, error)
	GetByUserIdAndWalletName(user_id int64, walletName string) (*Wallet, error)
	GetByUserIdAndWalletID(user_id int64, walletID string) (*Wallet, error)
}

type PostgresWalletRepository struct {
	db *pgxpool.Pool
}

func NewPostgresWalletRepository(db *pgxpool.Pool) *PostgresWalletRepository {
	return &PostgresWalletRepository{
		db: db,
	}
}

// GetByUserIDAndWalletName returns one wallet by wallet Name and UserID
func (r *PostgresWalletRepository) GetByUserIdAndWalletName(userID int64, walletName string) (*Wallet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	query := `select * from wallets where user_id = $1 AND name = $2`

	var wallet Wallet
	row := r.db.QueryRow(ctx, query, userID, walletName)

	err := row.Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Name,
		&wallet.Balance,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return &wallet, nil
}

// GetByUserIdAndWalletID returns one wallet by wallet ID and UserID
func (r *PostgresWalletRepository) GetByUserIdAndWalletID(userID int64, walletID string) (*Wallet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	query := `select * from wallets where user_id = $1 AND id = $2`

	var wallet Wallet
	row := r.db.QueryRow(ctx, query, userID, walletID)

	err := row.Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Name,
		&wallet.Balance,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return &wallet, nil
}

func (r *PostgresWalletRepository) CreateWallet(wallet Wallet) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	query := `insert into wallets (user_id, name, created_at, updated_at)
		values ($1, $2, $3, $4) returning id`

	var newID string
	err := r.db.QueryRow(ctx, query,
		wallet.UserID,
		wallet.Name,
		wallet.CreatedAt,
		wallet.UpdatedAt,
	).Scan(&newID)

	if err != nil {
		return "", fmt.Errorf("failed to create wallet: %w", err)
	}

	return newID, nil
}
