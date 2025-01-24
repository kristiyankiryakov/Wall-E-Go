package repository

import (
	"database/sql"
	"wall-e-go/auth-service/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) CreateUser(user models.User) error {
	query := `INSERT INTO users (username, password, created_at) VALUES ($1, $2, $3)`

	_, err := r.DB.Exec(query, user.Username, user.Password, user.CreatedAt)

	return err
}

func (r *UserRepository) GetUserByUsername(username string) (models.User, error) {
	user := models.User{}

	query := `SELECT id, username, password, created_at FROM users WHERE username=$1`

	err := r.DB.QueryRow(query, username).Scan(&user.ID, &user.Password, &user.Username, &user.CreatedAt)

	return user, err
}
