package data

import (
	"context"
	"database/sql"
	"log"
	"time"
)

const DB_TIMEOUT = time.Second * 3

type UserRepository interface {
	GetAll() ([]*User, error)
	GetByUsername(username string) (*User, error)
	GetOne(id int) (*User, error)
	DeleteByID(id int) error
	Insert(user User) (int, error)
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// GetAll returns a slice of all users
func (r *PostgresUserRepository) GetAll() ([]*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	query := `SELECT id, username, password, created_at, updated_at FROM users`

	rows, err := r.db.QueryContext(ctx, query)
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
func (r *PostgresUserRepository) GetByUsername(username string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	query := `select * from users where username = $1`

	var user User
	row := r.db.QueryRowContext(ctx, query, username)

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return nil, err
	}

	return &user, nil
}

// GetOne returns one user by id
func (r *PostgresUserRepository) GetOne(id int) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	query := `select * from users where id = $1`

	var user User
	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteByID deletes one user from the database, by ID
func (r *PostgresUserRepository) DeleteByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	stmt := `DELETE FROM users WHERE id = $1`

	_, err := r.db.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}

// Insert inserts a new user into the database, and returns the ID of the newly inserted row
func (r *PostgresUserRepository) Insert(user User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	var newID int
	stmt := `insert into users (username, password, created_at, updated_at)
		values ($1, $2, $3, $4) returning id`

	err := r.db.QueryRowContext(ctx, stmt,
		user.Username,
		user.Password,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}
