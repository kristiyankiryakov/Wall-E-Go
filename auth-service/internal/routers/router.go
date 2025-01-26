package router

import (
	"wall-e-go/auth-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

type Router struct {
	AuthHandler *handlers.AuthHandler
}

func NewRouter(authHandler *handlers.AuthHandler) *Router {
	return &Router{AuthHandler: authHandler}
}

func (r *Router) RegisterRoutes(router *gin.Engine) {
	router.POST("/register", r.AuthHandler.Register)
	router.POST("/login", r.AuthHandler.Login)
}
