package service

import (
	"auth/internal/data"
	"auth/proto/gen"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

// InMemoryUserRepository is a real implementation using in-memory storage
type InMemoryUserRepository struct {
	users map[string]data.User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]data.User),
	}
}

func (r *InMemoryUserRepository) GetByUsername(username string) (*data.User, error) {
	user, exists := r.users[username]
	if !exists {
		return &data.User{}, nil
	}
	return &user, nil
}

func (r *InMemoryUserRepository) Insert(user data.User) (int, error) {
	if _, exists := r.users[user.Username]; exists {
		return 0, errors.New("user already exists")
	}

	user.ID = len(r.users) + 1
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	r.users[user.Username] = user
	return user.ID, nil
}

func (r *InMemoryUserRepository) GetAll() ([]*data.User, error) {
	var users []*data.User
	for _, user := range r.users {
		userCopy := user
		users = append(users, &userCopy)
	}
	return users, nil
}

func (r *InMemoryUserRepository) GetOne(id int) (*data.User, error) {
	for _, user := range r.users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *InMemoryUserRepository) DeleteByID(id int) error {
	for username, user := range r.users {
		if user.ID == id {
			delete(r.users, username)
			return nil
		}
	}
	return errors.New("user not found")
}

// SimpleJWTUtil for testing
type SimpleJWTUtil struct{}

func (j *SimpleJWTUtil) GenerateToken(userID int) (string, error) {
	return "test-token", nil
}

func TestRegisterUser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		existingUsers map[string]data.User
		request       *gen.RegisterUserRequest
		expected      string
		expectedError bool
	}{
		{
			name:          "successful registration",
			existingUsers: map[string]data.User{},
			request: &gen.RegisterUserRequest{
				Username: "newuser",
				Password: "password123",
			},
			expected:      "test-token",
			expectedError: false,
		},
		{
			name: "duplicate user",
			existingUsers: map[string]data.User{
				"existinguser": {
					ID:       1,
					Username: "existinguser",
					Password: "hashedpassword",
				},
			},
			request: &gen.RegisterUserRequest{
				Username: "existinguser",
				Password: "password123",
			},
			expected:      "",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		// Setup repository with existing users
		repo := NewInMemoryUserRepository()
		repo.users = tc.existingUsers

		// Setup service
		jwtUtil := &SimpleJWTUtil{}
		service := NewAuthService(jwtUtil, repo)

		// Execute
		resp, err := service.RegisterUser(context.Background(), tc.request)

		if tc.expectedError {
			assert.Error(t, err)
			assert.Nil(t, resp)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tc.expected, resp.Token)
		}
	}
}

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	// Create a hashed password for testing
	correctPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), 12)

	testCases := []struct {
		name          string
		existingUsers map[string]data.User
		request       *gen.AuthenticateRequest
		expected      string
		expectError   bool
	}{
		{
			name: "successful authentication",
			existingUsers: map[string]data.User{
				"testuser": {
					ID:        1,
					Username:  "testuser",
					Password:  string(correctPasswordHash),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			request: &gen.AuthenticateRequest{
				Username: "testuser",
				Password: "correct_password",
			},
			expected:    "test-token",
			expectError: false,
		},
		{
			name: "incorrect password",
			existingUsers: map[string]data.User{
				"testuser": {
					ID:        1,
					Username:  "testuser",
					Password:  string(correctPasswordHash),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			request: &gen.AuthenticateRequest{
				Username: "testuser",
				Password: "wrong_password",
			},
			expected:    "",
			expectError: true,
		},
		{
			name:          "non-existent user",
			existingUsers: map[string]data.User{},
			request: &gen.AuthenticateRequest{
				Username: "nonexistentuser",
				Password: "anypassword",
			},
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		// Setup repository with existing users
		repo := NewInMemoryUserRepository()
		repo.users = tc.existingUsers

		// Setup service
		jwtUtil := &SimpleJWTUtil{}
		service := NewAuthService(jwtUtil, repo)

		// Execute
		resp, err := service.Authenticate(context.Background(), tc.request)

		// Verify expectations
		if tc.expectError {
			assert.Error(t, err)
			assert.Nil(t, resp)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tc.expected, resp.Token)
		}
	}
}
