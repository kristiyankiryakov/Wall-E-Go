package data

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

type UserRepository interface {
	GetAll(ctx context.Context) ([]*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetOne(ctx context.Context, id int) (*User, error)
	DeleteByID(ctx context.Context, id int) error
	Insert(ctx context.Context, user User) (int, error)
}

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// GetAll returns a slice of all users
func (r *PostgresUserRepository) GetAll(ctx context.Context) ([]*User, error) {
	query := `SELECT id, username, password, created_at, updated_at FROM users`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Password,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			log.Println("Error scanning user:", err)
			return nil, err
		}
		users = append(users, &user)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		log.Println("Error iterating rows:", err)
		return nil, err
	}

	return users, nil
}

// GetByUsername returns one user by username
func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := `select * from users where username = $1`

	var user User
	row := r.db.QueryRow(ctx, query, username)

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("error getting user by username: %v", err)
	}

	return &user, nil
}

// GetOne returns one user by id
func (r *PostgresUserRepository) GetOne(ctx context.Context, id int) (*User, error) {
	query := `select * from users where id = $1`

	var user User
	row := r.db.QueryRow(ctx, query, id)

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("error getting user by id: %v", err)
	}

	return &user, nil
}

// DeleteByID deletes one user from the database, by ID
func (r *PostgresUserRepository) DeleteByID(ctx context.Context, id int) error {
	stmt := `DELETE FROM users WHERE id = $1`

	_, err := r.db.Exec(ctx, stmt, id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("error deleting user by id: %v", err)
	}

	return nil
}

// Insert inserts a new user into the database, and returns the ID of the newly inserted row
func (r *PostgresUserRepository) Insert(ctx context.Context, user User) (int, error) {
	var newID int
	stmt := `insert into users (username, password, created_at, updated_at)
		values ($1, $2, $3, $4) returning id`

	err := r.db.QueryRow(ctx, stmt,
		user.Username,
		user.Password,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("error inserting user: %v", err)
	}

	return newID, nil
}
