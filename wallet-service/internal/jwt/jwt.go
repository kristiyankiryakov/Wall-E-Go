package jwt

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

type JWTUtil interface {
	ValidateToken(token string) (int, error)
}

type JWTUtilImpl struct {
	secretKey string
}

func NewJWTUtil(secretKey string) *JWTUtilImpl {
	return &JWTUtilImpl{
		secretKey: secretKey,
	}
}

func (j *JWTUtilImpl) ValidateToken(token string) (int, error) {
	claims := &jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		log.Println("Parse error:", err)
		return 0, err
	}

	if !parsedToken.Valid {
		log.Println("Token invalid")
		return 0, errors.New("invalid token")
	}

	log.Println("Claims:", *claims)
	sub, ok := (*claims)["sub"]
	if !ok {
		log.Println("Missing sub claim")
		return 0, errors.New("missing subject claim")
	}

	switch v := sub.(type) {
	case float64:
		return int(v), nil
	case string:
		userID, err := strconv.Atoi(v)
		if err != nil {
			log.Println("String conversion error:", err)
			return 0, errors.New("invalid user_id format")
		}
		return userID, nil
	default:
		log.Println("Unexpected sub type:", fmt.Sprintf("%T", v))
		return 0, errors.New("invalid user_id type")
	}
}
