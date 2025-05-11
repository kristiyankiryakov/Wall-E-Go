package jwt

import (
	"auth/internal/errors"
	"log"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTUtil interface {
	GenerateToken(user_id int) (string, error)
}

type JWTUtilImpl struct {
	secretKey string
}

type customClaims struct {
	jwt.RegisteredClaims
}

func NewJWTUtil(secretKey string) *JWTUtilImpl {
	return &JWTUtilImpl{
		secretKey: secretKey,
	}
}

func (j *JWTUtilImpl) GenerateToken(userID int) (string, error) {
	if j.secretKey == "" {
		log.Println("secret key is missing")
		return "", errors.WrapError(errors.ErrInternal, "secret key is missing")
	}

	claims := &customClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth-service",
			Subject:   strconv.Itoa(userID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", errors.WrapError(errors.ErrInternal, "Error signing token")
	}

	return signedToken, nil
}
