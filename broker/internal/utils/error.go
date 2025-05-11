package utils

import (
	"google.golang.org/grpc/status"
	"log"
	"net/http"

	"google.golang.org/grpc/codes"
)

// HandleGRPCError handles errors from gRPC service calls
func HandleGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		Respond(w, http.StatusInternalServerError, "", nil, err)
		return
	}

	code := st.Code()
	message := st.Message()
	httpStatus := GrpcToHTTPStatus(code)

	// Log the full error details for debugging
	log.Printf("gRPC error: code=%s, message=%s, mapped to HTTP=%d",
		code.String(),
		message,
		httpStatus,
	)

	// For security, we may want to sanitize some error messages before sending to clients
	clientMessage := getSafeErrorMessage(code, message)

	Respond(w, httpStatus, clientMessage, nil, nil)
}

// getSafeErrorMessage ensures we don't leak sensitive information in error messages
func getSafeErrorMessage(code codes.Code, originalMessage string) string {
	// For some error types, we might want to sanitize or customize the message
	switch code {
	case codes.Internal:
		return "An internal error occurred"
	case codes.Unavailable:
		return "Service temporarily unavailable"
	case codes.DeadlineExceeded:
		return "Request timed out"
	default:
		return originalMessage
	}
}

// GrpcToHTTPStatus maps gRPC codes to HTTP status codes
func GrpcToHTTPStatus(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusRequestTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
