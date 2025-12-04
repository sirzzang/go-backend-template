package routes

import (
	"github.com/gin-gonic/gin"
	userHandler "github.com/your-org/go-backend-template/internal/app/server/handler/user"
)

// SetupAuthRoutes sets up authentication routes (public).
func SetupAuthRoutes(r *gin.RouterGroup, h *userHandler.Handler) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
	}
}

// SetupUserRoutes sets up user routes (protected).
func SetupUserRoutes(r *gin.RouterGroup, h *userHandler.Handler, auth AuthMiddleware) {
	// Current user endpoints
	r.GET("/me", h.GetMe)

	// User CRUD endpoints
	users := r.Group("/users")
	{
		// List users - requires admin role
		users.GET("", auth.RequireAdmin(), h.GetUsers)

		// Create user - requires admin role
		users.POST("", auth.RequireAdmin(), h.CreateUser)

		// Get user by ID - authenticated users can access
		users.GET("/:id", h.GetUser)

		// Update user - requires admin role
		users.PATCH("/:id", auth.RequireAdmin(), h.UpdateUser)

		// Delete user - requires admin role
		users.DELETE("/:id", auth.RequireAdmin(), h.DeleteUser)

		// Change password - user can change their own password
		users.POST("/:id/change-password", h.ChangePassword)
	}
}

