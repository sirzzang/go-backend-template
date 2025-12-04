package user

import (
	"github.com/your-org/go-backend-template/internal/pkg/entity"
)

// ========== Service Dependencies ==========
// Interfaces that the user service depends on (injected from outside)

// IUserRepository defines the interface for user data access.
// This interface should contain only the methods needed by the user service.
type IUserRepository interface {
	// Create
	InsertUser(user *entity.User) (int, error)

	// Read
	GetUserById(id int) (*entity.User, error)
	GetUserByEmail(email string) (*entity.User, error)
	GetUsers(offset, limit int, onlyActive bool) ([]*entity.User, error)
	GetUserCount(onlyActive bool) (int, error)
	ExistsUserByEmail(email string) (bool, error)

	// Update
	UpdateUser(user *entity.User) error
	UpdateUserPassword(id int, hashedPassword string) error

	// Delete
	DeleteUserById(id int) error
}

// IPasswordHasher defines the interface for password hashing.
type IPasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

