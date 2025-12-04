package entity

import "time"

// User represents a user entity in the system.
type User struct {
	Id        int       `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // never expose password in JSON
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRole constants
const (
	RoleAdmin  = "admin"
	RoleUser   = "user"
	RoleViewer = "viewer"
)

// IsValidRole checks if the given role is valid.
func IsValidRole(role string) bool {
	switch role {
	case RoleAdmin, RoleUser, RoleViewer:
		return true
	default:
		return false
	}
}

