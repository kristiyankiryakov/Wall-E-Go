package middleware

import (
	"broker-service/internal/clients"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

var secretKey = os.Getenv("JWT_KEY")

func AuthenticateWalletOwner(walletClient *clients.WalletClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := c.Request.Header.Get("Authorization")
		if bearerToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		log.Printf("received: %v", bearerToken)

		var token string
		if len(bearerToken) > 7 && bearerToken[:7] == "Bearer " {
			token = bearerToken[7:]
		}

		userID, err := validateToken(token) // Local JWT parsing
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		walletIDParam := c.Query("walletID")
		if walletIDParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid or missing walletID: %v", walletIDParam)})
			c.Abort()
			return
		}
		walletID, err := strconv.Atoi(walletIDParam)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "converting walletID"})
			c.Abort()
		}

		if ok, err := walletClient.IsWalletOwner(c, userID, walletID); !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

func AppendUserIDToGrpcContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		// Convert userID to string based on its actual type
		var userIDStr string
		switch v := userID.(type) {
		case int:
			userIDStr = strconv.Itoa(v)
		case int64:
			userIDStr = strconv.FormatInt(v, 10)
		case string:
			userIDStr = v
		default:
			log.Printf("Unsupported userID type: %T", userID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID format"})
			c.Abort()
			return
		}

		// Create context with userID metadata
		ctx := metadata.AppendToOutgoingContext(c.Request.Context(), "userID", userIDStr)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func AuthenticateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := c.Request.Header.Get("Authorization")
		if bearerToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		var token string
		if len(bearerToken) > 7 && bearerToken[:7] == "Bearer " {
			token = bearerToken[7:]
		}

		userID, err := validateToken(token) // Local JWT parsing
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

func validateToken(token string) (int, error) {
	claims := &jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
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
