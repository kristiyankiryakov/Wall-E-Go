package middlewares

import (
	"broker/internal/utils"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"net/http"
	"strconv"
	"time"
)

type customClaims struct {
	jwt.RegisteredClaims
}

// Authenticate middleware for validating JWT tokens and appending user ID to the context
func Authenticate(secret string, log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			bearerToken := r.Header.Get("Authorization")
			if bearerToken == "" {
				utils.Respond(w, http.StatusUnauthorized, "missing token", nil, errors.New("unauthorized access"))
				return
			}

			var token string
			if len(bearerToken) > 7 && bearerToken[:7] == "Bearer " {
				token = bearerToken[7:]
			} else {
				utils.Respond(w, http.StatusUnauthorized, "invalid token format", nil, errors.New("unauthorized access"))
				return
			}

			userID, err := validateToken(token, secret, log)
			if err != nil {
				utils.Respond(w, http.StatusUnauthorized, "invalid token", nil, err)
				return
			}

			md := metadata.New(map[string]string{
				"userID": strconv.Itoa(userID),
			})
			grpcCtx := metadata.NewOutgoingContext(r.Context(), md)

			next.ServeHTTP(w, r.WithContext(grpcCtx))
		}
		return http.HandlerFunc(fn)
	}
}

func validateToken(token string, secret string, log *logrus.Logger) (int, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %v", err)
	}

	if !parsedToken.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := parsedToken.Claims.(*customClaims)
	if !ok {
		return 0, fmt.Errorf("failed to extract claims from token")
	}
	log.Info("Claims: ", claims, "subject: ", claims.Subject)

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, fmt.Errorf("failed to convert user ID to int: %v", err)
	}
	expiration := claims.ExpiresAt
	if expiration == nil || expiration.Time.Before(time.Now()) {
		return 0, fmt.Errorf("token has expired")
	}

	if userID == 0 {
		return 0, fmt.Errorf("user ID not found in token claims")
	}

	return userID, nil
}
