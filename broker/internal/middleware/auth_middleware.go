package middleware

import (
	"broker/internal/clients"
	"broker/internal/config"
	"broker/internal/models"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

type AuthMiddleware interface {
	AuthenticateWalletOwner() gin.HandlerFunc
	AppendUserIDToGrpcContext() gin.HandlerFunc
	AuthenticateUser() gin.HandlerFunc
}

type AuthMiddlewareImpl struct {
	config       *config.ServerCfg
	walletClient *clients.WalletClient
	log          *logrus.Logger
}

func NewAuthMiddleware(config *config.ServerCfg, walletClient *clients.WalletClient, log *logrus.Logger) AuthMiddleware {
	return &AuthMiddlewareImpl{
		config:       config,
		walletClient: walletClient,
		log:          log,
	}
}

func (m *AuthMiddlewareImpl) AuthenticateWalletOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := m.extractAndValidateToken(c)
		if err != nil {
			return // Error already handled in the function
		}

		var req models.TransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			m.log.WithError(err).Error("Invalid transaction request format")
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request: %v", err)})
			c.Abort()
			return
		}

		if ok, err := m.walletClient.IsWalletOwner(c, int64(userID), req.WalletID); !ok {
			m.log.WithFields(logrus.Fields{
				"userID":   userID,
				"walletID": req.WalletID,
				"error":    err.Error(),
			}).Warn("Unauthorized wallet access attempt")
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("transactionRequest", req)
		c.Set("userID", userID)
		c.Next()
	}
}

func (m *AuthMiddlewareImpl) AppendUserIDToGrpcContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists || userID == nil {
			m.log.Warn("User not authenticated in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		userIDStr, err := m.convertUserIDToString(userID)
		if err != nil {
			m.log.WithError(err).Error("Failed to convert user ID")
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

func (m *AuthMiddlewareImpl) AuthenticateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := m.extractAndValidateToken(c)
		if err != nil {
			return // Error already handled in the function
		}

		c.Set("userID", userID)
		c.Next()
	}
}

// extractAndValidateToken extracts bearer token from request and validates it
func (m *AuthMiddlewareImpl) extractAndValidateToken(c *gin.Context) (int, error) {
	bearerToken := c.Request.Header.Get("Authorization")
	if bearerToken == "" {
		m.log.Warn("Missing authorization token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		c.Abort()
		return 0, errors.New("missing token")
	}

	var token string
	if len(bearerToken) > 7 && bearerToken[:7] == "Bearer " {
		token = bearerToken[7:]
	} else {
		token = bearerToken // Try to be flexible if token format is different
	}

	userID, err := m.validateToken(token)
	if err != nil {
		m.log.WithError(err).Warn("Invalid token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return 0, err
	}

	return userID, nil
}

func (m *AuthMiddlewareImpl) validateToken(token string) (int, error) {
	claims := &jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(m.config.JWTSecret), nil
	})

	if err != nil {
		m.log.WithError(err).Debug("Token parse error")
		return 0, err
	}

	if !parsedToken.Valid {
		m.log.Debug("Token invalid")
		return 0, errors.New("invalid token")
	}

	sub, ok := (*claims)["sub"]
	if !ok {
		m.log.Debug("Missing sub claim in token")
		return 0, errors.New("missing subject claim")
	}

	return m.parseUserIDFromClaim(sub)
}

func (m *AuthMiddlewareImpl) parseUserIDFromClaim(sub interface{}) (int, error) {
	switch v := sub.(type) {
	case float64:
		return int(v), nil
	case string:
		userID, err := strconv.Atoi(v)
		if err != nil {
			m.log.WithError(err).Debug("String conversion error for user ID")
			return 0, errors.New("invalid user_id format")
		}
		return userID, nil
	default:
		m.log.WithField("type", fmt.Sprintf("%T", v)).Debug("Unexpected sub type")
		return 0, errors.New("invalid user_id type")
	}
}

func (m *AuthMiddlewareImpl) convertUserIDToString(userID interface{}) (string, error) {
	switch v := userID.(type) {
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("unsupported userID type: %T", userID)
	}
}
