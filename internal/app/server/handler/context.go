package handler

import "github.com/gin-gonic/gin"

// Context key constants
const (
	ContextKeyUserId   = "user_id"
	ContextKeyUserRole = "user_role"
)

// GetUserId retrieves the user ID from the gin context.
func GetUserId(c *gin.Context) int {
	return c.GetInt(ContextKeyUserId)
}

// GetUserRole retrieves the user role from the gin context.
func GetUserRole(c *gin.Context) string {
	return c.GetString(ContextKeyUserRole)
}

// SetUserId sets the user ID in the gin context.
func SetUserId(c *gin.Context, userId int) {
	c.Set(ContextKeyUserId, userId)
}

// SetUserRole sets the user role in the gin context.
func SetUserRole(c *gin.Context, role string) {
	c.Set(ContextKeyUserRole, role)
}

