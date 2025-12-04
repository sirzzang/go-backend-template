package repository

import "errors"

// Repository layer errors
// These errors are used by repository implementations and consumed by service layer.
// This allows service layer to handle repository errors without depending on specific implementations.

var (
	// User repository errors
	ErrUserNotFound   = errors.New("user not found")
	ErrDuplicateEmail = errors.New("email already exists")
)

