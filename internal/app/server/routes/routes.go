package routes

import (
	"github.com/gin-gonic/gin"
	userHandler "github.com/your-org/go-backend-template/internal/app/server/handler/user"
)

// Handlers holds all domain-specific handlers.
type Handlers struct {
	User *userHandler.Handler
}

// AuthMiddleware defines the auth middleware interface.
type AuthMiddleware interface {
	RequireAuth() gin.HandlerFunc
	RequireAdmin() gin.HandlerFunc
	RequireRole(roles ...string) gin.HandlerFunc
}

// SetupRoutes configures all API routes.
func SetupRoutes(r *gin.RouterGroup, h *Handlers, auth AuthMiddleware) {
	// Public routes (no authentication required)
	SetupAuthRoutes(r, h.User)

	// Protected routes (authentication required)
	protected := r.Group("")
	protected.Use(auth.RequireAuth())
	{
		SetupUserRoutes(protected, h.User, auth)
	}
}

