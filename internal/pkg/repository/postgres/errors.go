package postgres

import "errors"

var (
	// ErrNoUser is returned when a user is not found.
	ErrNoUser = errors.New("user not found")

	// ErrDuplicateEmail is returned when email already exists.
	ErrDuplicateEmail = errors.New("email already exists")
)
