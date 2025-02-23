package handlers

import (
	"broker-service/internal/clients"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type BrokerHandler struct {
	authClient   *clients.AuthClient
	walletClient *clients.WalletClient
}

func NewBrokerHandler(authClient *clients.AuthClient, walletClient *clients.WalletClient) *BrokerHandler {
	return &BrokerHandler{
		authClient:   authClient,
		walletClient: walletClient,
	}
}

func (h *BrokerHandler) CreateWallet(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	// Extract JWT from HTTP Header
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	ctx := metadata.AppendToOutgoingContext(c.Request.Context(), "authorization", authHeader)

	walletID, err := h.walletClient.CreateWallet(ctx, req.Name)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"wallet_id": walletID})
}

func (h *BrokerHandler) RegisterUser(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, err := h.authClient.RegisterUser(req.Username, req.Password)
	if err != nil {
		handleGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *BrokerHandler) Authenticate(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, err := h.authClient.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func SetupRouter(h *BrokerHandler) *gin.Engine {
	r := gin.Default()
	r.POST("/register", h.RegisterUser)
	r.POST("/authenticate", h.Authenticate)
	r.POST("/create", h.CreateWallet)
	return r
}

// handleGRPCError converts gRPC errors to HTTP responses
func handleGRPCError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	httpCode := grpcToHTTPStatus(st.Code())
	c.JSON(httpCode, gin.H{"error": st.Message()})
}

// grpcToHTTPStatus maps gRPC codes to HTTP status codes
func grpcToHTTPStatus(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.Internal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
