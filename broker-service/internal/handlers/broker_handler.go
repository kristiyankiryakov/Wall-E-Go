package handlers

import (
	"broker-service/internal/authclient"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BrokerHandler struct {
	authClient *authclient.AuthClient
}

func NewBrokerHandler(authClient *authclient.AuthClient) *BrokerHandler {
	return &BrokerHandler{authClient: authClient}
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
