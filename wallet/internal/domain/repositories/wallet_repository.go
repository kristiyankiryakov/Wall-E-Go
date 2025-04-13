package repositories

import (
	"context"
	"database/sql"
	"log"
	"time"
	"wallet/internal/domain/entities"
)

const DB_TIMEOUT = time.Second * 10

type WalletRepository interface {
	CreateWallet(wallet entities.Wallet) (string, error)
	GetByUserIdAndWalletName(user_id int64, walletName string) (*entities.Wallet, error)
	GetByUserIdAndWalletID(user_id int64, walletID string) (*entities.Wallet, error)
}

type PostgresWalletRepository struct {
	db *sql.DB
}

func NewPostgresWalletRepository(db *sql.DB) *PostgresWalletRepository {
	return &PostgresWalletRepository{db: db}
}

// GetByUserIDAndWalletName returns one wallet by wallet Name and UserID
func (r *PostgresWalletRepository) GetByUserIdAndWalletName(user_id int64, walletName string) (*entities.Wallet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	query := `select * from wallets where user_id = $1 AND name = $2`

	var wallet entities.Wallet
	row := r.db.QueryRowContext(ctx, query, user_id, walletName)

	err := row.Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Name,
		&wallet.Balance,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return nil, err
	}

	return &wallet, nil
}

// GetByUserIdAndWalletID returns one wallet by wallet ID and UserID
func (r *PostgresWalletRepository) GetByUserIdAndWalletID(user_id int64, walletID string) (*entities.Wallet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	query := `select * from wallets where user_id = $1 AND id = $2`

	var wallet entities.Wallet
	row := r.db.QueryRowContext(ctx, query, user_id, walletID)

	err := row.Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Name,
		&wallet.Balance,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return nil, err
	}

	return &wallet, nil
}

func (r *PostgresWalletRepository) CreateWallet(wallet entities.Wallet) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	query := `insert into wallets (user_id, name, created_at, updated_at)
		values ($1, $2, $3, $4) returning id`

	var newID string
	err := r.db.QueryRowContext(ctx, query,
		wallet.UserID,
		wallet.Name,
		wallet.CreatedAt,
		wallet.UpdatedAt,
	).Scan(&newID)

	if err != nil {
		log.Println(err)
		return "", err
	}

	return newID, nil
}
