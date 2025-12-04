package domain

import (
	"fmt"
	"net/http"
)

// DomainError is an interface that domain errors should implement.
// It provides HTTP status code mapping for error handling in handlers.
type DomainError interface {
	error
	HTTPStatus() int
}

// ========== Common Errors ==========

// InternalServerError represents an unexpected server error.
type InternalServerError struct {
	Msg string
	Err error
}

func (e InternalServerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func (e InternalServerError) HTTPStatus() int {
	return http.StatusInternalServerError
}

// UnauthorizedError represents an authentication failure.
type UnauthorizedError struct {
	Reason string
}

func (e UnauthorizedError) Error() string {
	return fmt.Sprintf("unauthorized: %s", e.Reason)
}

func (e UnauthorizedError) HTTPStatus() int {
	return http.StatusUnauthorized
}

// ForbiddenError represents an authorization failure.
type ForbiddenError struct {
	Reason string
}

func (e ForbiddenError) Error() string {
	return fmt.Sprintf("forbidden: %s", e.Reason)
}

func (e ForbiddenError) HTTPStatus() int {
	return http.StatusForbidden
}

// ValidationError represents a validation failure.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

func (e ValidationError) HTTPStatus() int {
	return http.StatusBadRequest
}

// ========== User Domain Errors ==========

// UserNotFoundError represents a user not found error.
type UserNotFoundError struct {
	Id    int
	Email string
}

func (e UserNotFoundError) Error() string {
	if e.Id != 0 {
		return fmt.Sprintf("user not found with id: %d", e.Id)
	}
	if e.Email != "" {
		return fmt.Sprintf("user not found with email: %s", e.Email)
	}
	return "user not found"
}

func (e UserNotFoundError) HTTPStatus() int {
	return http.StatusNotFound
}

// UserAlreadyExistsError represents a duplicate user error.
type UserAlreadyExistsError struct {
	Email string
}

func (e UserAlreadyExistsError) Error() string {
	return fmt.Sprintf("user already exists with email: %s", e.Email)
}

func (e UserAlreadyExistsError) HTTPStatus() int {
	return http.StatusConflict
}

// InvalidCredentialsError represents an invalid login attempt.
type InvalidCredentialsError struct{}

func (e InvalidCredentialsError) Error() string {
	return "invalid email or password"
}

func (e InvalidCredentialsError) HTTPStatus() int {
	return http.StatusUnauthorized
}

// InvalidRoleError represents an invalid role error.
type InvalidRoleError struct {
	Role string
}

func (e InvalidRoleError) Error() string {
	return fmt.Sprintf("invalid role: %s", e.Role)
}

func (e InvalidRoleError) HTTPStatus() int {
	return http.StatusBadRequest
}

