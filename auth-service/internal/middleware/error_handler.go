package middleware

import (
	"log"
	"net/http"
	errors "wall-e-go/common"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() //Process request

		// Check if there were errors during request processing
		if len(c.Errors) > 0 {
			// Use the first error for simplicity
			err := c.Errors[0].Err

			switch e := err.(type) {
			case *errors.AppError:
				// Handling known common errors
				c.JSON(e.Code, gin.H{"error": e.Message, "details": e.Details})
			default:
				// Log and handle unknown/internal errors
				log.Printf("Unknown error: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
		}
	}
}
