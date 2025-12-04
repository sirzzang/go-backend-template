package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher handles password hashing using bcrypt.
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new password hasher.
func NewPasswordHasher(cost int) *PasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &PasswordHasher{cost: cost}
}

// Hash hashes a password using bcrypt.
func (h *PasswordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Compare compares a hashed password with a plain text password.
// Returns nil if they match, error otherwise.
func (h *PasswordHasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

