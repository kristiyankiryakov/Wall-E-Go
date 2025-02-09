package util

import (
	"os"
	"time"
	errors "wall-e-go/internal/error"

	"github.com/golang-jwt/jwt/v5"
)

type JWTUtil struct{}

func (util JWTUtil) GenerateToken(username string) (string, error) {
	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		return "", errors.WrapError(errors.ErrInternal, "JWT_KEY env variable not set")
	}

	claims := jwt.MapClaims{
		"iss": "auth-server",
		"sub": username,
		"exp": time.Now().Add(25 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", errors.WrapError(errors.ErrInternal, "Error signing token")
	}

	return signedToken, nil
}
