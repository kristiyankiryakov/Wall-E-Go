package errors

import "net/http"

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

var (
	ErrNotFound      = &AppError{Code: http.StatusNotFound, Message: "Resource not found"}
	ErrAlreadyExists = &AppError{Code: http.StatusConflict, Message: "Resource already exists"}
	ErrBadRequest    = &AppError{Code: http.StatusBadRequest, Message: "Invalid request"}
	ErrInternal      = &AppError{Code: http.StatusInternalServerError, Message: "Internal server error"}
)

func WrapError(err *AppError, details string) *AppError {
	return &AppError{
		Code:    err.Code,
		Message: err.Message,
		Details: details,
	}
}
