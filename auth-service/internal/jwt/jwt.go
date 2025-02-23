package jwt

import (
	"log"
	"time"
	"wall-e-go/internal/errors"

	"github.com/golang-jwt/jwt/v5"
)

type JWTUtil interface {
	GenerateToken(user_id int) (string, error)
}

type JWTUtilImpl struct {
	secretKey string
}

func NewJWTUtil(secretKey string) *JWTUtilImpl {
	return &JWTUtilImpl{secretKey: secretKey}
}

func (j *JWTUtilImpl) GenerateToken(user_id string) (string, error) {
	if j.secretKey == "" {
		log.Println("secret key is missing")
		return "", errors.WrapError(errors.ErrInternal, "secret key is missing")
	}

	claims := jwt.MapClaims{
		"iss": "auth-server",
		"sub": user_id,
		"exp": time.Now().Add(25 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", errors.WrapError(errors.ErrInternal, "Error signing token")
	}

	return signedToken, nil
}
