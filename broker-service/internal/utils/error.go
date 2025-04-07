package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// handleGRPCError converts gRPC errors to HTTP responses
func HandleGRPCError(c *gin.Context, err error) {
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
