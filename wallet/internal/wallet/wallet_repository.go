package wallet

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	CreateWallet(ctx context.Context, wallet *Wallet) (string, error)
	GetByUserIdAndWalletName(ctx context.Context, userID int, walletName string) (*Wallet, error)
	GetByUserIdAndWalletID(ctx context.Context, userID int, walletID string) (*Wallet, error)
}

type PostgresWalletRepository struct {
	db  *pgxpool.Pool
	log *logrus.Logger
}

func NewPostgresWalletRepository(db *pgxpool.Pool, log *logrus.Logger) *PostgresWalletRepository {
	return &PostgresWalletRepository{
		db:  db,
		log: log,
	}
}

// GetByUserIdAndWalletName returns one wallet by wallet name and UserID
func (r *PostgresWalletRepository) GetByUserIdAndWalletName(ctx context.Context, userID int, walletName string) (*Wallet, error) {
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
func (r *PostgresWalletRepository) GetByUserIdAndWalletID(ctx context.Context, userID int, walletID string) (*Wallet, error) {
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

// CreateWallet creates a new wallet in the database
func (r *PostgresWalletRepository) CreateWallet(ctx context.Context, wallet *Wallet) (string, error) {
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
