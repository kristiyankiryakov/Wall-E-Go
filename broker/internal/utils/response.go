package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// Response represents a standardized API response structure
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Respond writes a JSON response with the given parameters
func Respond(w http.ResponseWriter, statusCode int, message string, data interface{}, err error) {
	var r Response
	r.Status = "success"
	if message != "" {
		r.Message = message
	}

	if data != nil {
		r.Data = data
	}

	if err != nil {
		r.Status = "error"
		r.Error = err.Error()
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(&r); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
