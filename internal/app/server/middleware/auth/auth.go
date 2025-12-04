package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/your-org/go-backend-template/internal/app/server/handler"
	"github.com/your-org/go-backend-template/internal/pkg/entity"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
)

// IJWTValidator defines the interface for JWT validation.
type IJWTValidator interface {
	ValidateToken(tokenString string) (*Claims, error)
}

// Claims represents JWT claims.
type Claims struct {
	UserId int
	Role   string
}

// Middleware provides authentication middleware.
type Middleware struct {
	jwtValidator IJWTValidator
}

// New creates a new auth middleware.
func New(jwtValidator IJWTValidator) (*Middleware, error) {
	if jwtValidator == nil {
		return nil, errors.New("jwt validator is required")
	}
	return &Middleware{jwtValidator: jwtValidator}, nil
}

// RequireAuth returns a middleware that requires a valid JWT token.
func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set Vary header for caching
		c.Header("Vary", authorizationHeader)

		// Get Authorization header
		authHeader := c.GetHeader(authorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "authorization header is required",
			})
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid authorization header format",
			})
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "token is required",
			})
			return
		}

		// Validate token
		claims, err := m.jwtValidator.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid token",
				"data":    err.Error(),
			})
			return
		}

		// Set user info in context
		handler.SetUserId(c, claims.UserId)
		handler.SetUserRole(c, claims.Role)

		c.Next()
	}
}

// RequireRole returns a middleware that requires a specific role.
func (m *Middleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := handler.GetUserRole(c)

		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "insufficient permissions",
		})
	}
}

// RequireAdmin is a shortcut for RequireRole(entity.RoleAdmin).
func (m *Middleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole(entity.RoleAdmin)
}

// RequireAdminOrSelf returns a middleware that allows admin or the user themselves.
func (m *Middleware) RequireAdminOrSelf(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := handler.GetUserRole(c)
		userId := handler.GetUserId(c)

		// Admin can access any resource
		if userRole == entity.RoleAdmin {
			c.Next()
			return
		}

		// Check if accessing own resource
		paramId := c.Param(paramName)
		if paramId != "" {
			// Simple string comparison since we set it as int in context
			if paramId == string(rune(userId)) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "insufficient permissions",
		})
	}
}

