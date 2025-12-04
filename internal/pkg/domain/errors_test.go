package domain

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ========== InternalServerError Tests ==========

func TestInternalServerError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      InternalServerError
		expected string
	}{
		{
			name:     "with underlying error",
			err:      InternalServerError{Msg: "database error", Err: errors.New("connection refused")},
			expected: "database error: connection refused",
		},
		{
			name:     "without underlying error",
			err:      InternalServerError{Msg: "something went wrong", Err: nil},
			expected: "something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestInternalServerError_HTTPStatus(t *testing.T) {
	err := InternalServerError{}
	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus())
}

// ========== UnauthorizedError Tests ==========

func TestUnauthorizedError_Error(t *testing.T) {
	err := UnauthorizedError{Reason: "token expired"}
	assert.Equal(t, "unauthorized: token expired", err.Error())
}

func TestUnauthorizedError_HTTPStatus(t *testing.T) {
	err := UnauthorizedError{}
	assert.Equal(t, http.StatusUnauthorized, err.HTTPStatus())
}

// ========== ForbiddenError Tests ==========

func TestForbiddenError_Error(t *testing.T) {
	err := ForbiddenError{Reason: "insufficient permissions"}
	assert.Equal(t, "forbidden: insufficient permissions", err.Error())
}

func TestForbiddenError_HTTPStatus(t *testing.T) {
	err := ForbiddenError{}
	assert.Equal(t, http.StatusForbidden, err.HTTPStatus())
}

// ========== ValidationError Tests ==========

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ValidationError
		expected string
	}{
		{
			name:     "with field",
			err:      ValidationError{Field: "email", Message: "must be a valid email"},
			expected: "validation error on field 'email': must be a valid email",
		},
		{
			name:     "without field",
			err:      ValidationError{Message: "invalid request"},
			expected: "validation error: invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestValidationError_HTTPStatus(t *testing.T) {
	err := ValidationError{}
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus())
}

// ========== UserNotFoundError Tests ==========

func TestUserNotFoundError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      UserNotFoundError
		expected string
	}{
		{
			name:     "with id",
			err:      UserNotFoundError{Id: 123},
			expected: "user not found with id: 123",
		},
		{
			name:     "with email",
			err:      UserNotFoundError{Email: "test@example.com"},
			expected: "user not found with email: test@example.com",
		},
		{
			name:     "with both (id takes precedence)",
			err:      UserNotFoundError{Id: 123, Email: "test@example.com"},
			expected: "user not found with id: 123",
		},
		{
			name:     "without id or email",
			err:      UserNotFoundError{},
			expected: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestUserNotFoundError_HTTPStatus(t *testing.T) {
	err := UserNotFoundError{}
	assert.Equal(t, http.StatusNotFound, err.HTTPStatus())
}

// ========== UserAlreadyExistsError Tests ==========

func TestUserAlreadyExistsError_Error(t *testing.T) {
	err := UserAlreadyExistsError{Email: "existing@example.com"}
	assert.Equal(t, "user already exists with email: existing@example.com", err.Error())
}

func TestUserAlreadyExistsError_HTTPStatus(t *testing.T) {
	err := UserAlreadyExistsError{}
	assert.Equal(t, http.StatusConflict, err.HTTPStatus())
}

// ========== InvalidCredentialsError Tests ==========

func TestInvalidCredentialsError_Error(t *testing.T) {
	err := InvalidCredentialsError{}
	assert.Equal(t, "invalid email or password", err.Error())
}

func TestInvalidCredentialsError_HTTPStatus(t *testing.T) {
	err := InvalidCredentialsError{}
	assert.Equal(t, http.StatusUnauthorized, err.HTTPStatus())
}

// ========== InvalidRoleError Tests ==========

func TestInvalidRoleError_Error(t *testing.T) {
	err := InvalidRoleError{Role: "superadmin"}
	assert.Equal(t, "invalid role: superadmin", err.Error())
}

func TestInvalidRoleError_HTTPStatus(t *testing.T) {
	err := InvalidRoleError{}
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus())
}

// ========== DomainError Interface Tests ==========

func TestDomainError_Interface(t *testing.T) {
	// Verify all domain errors implement DomainError interface
	var _ DomainError = InternalServerError{}
	var _ DomainError = UnauthorizedError{}
	var _ DomainError = ForbiddenError{}
	var _ DomainError = ValidationError{}
	var _ DomainError = UserNotFoundError{}
	var _ DomainError = UserAlreadyExistsError{}
	var _ DomainError = InvalidCredentialsError{}
	var _ DomainError = InvalidRoleError{}
}

func TestDomainError_TypeAssertion(t *testing.T) {
	var err error = UserNotFoundError{Id: 1}

	domainErr, ok := err.(DomainError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusNotFound, domainErr.HTTPStatus())
}

// ========== Error Wrapping Tests ==========

func TestInternalServerError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("database connection failed")
	err := InternalServerError{Msg: "query failed", Err: underlyingErr}

	// While InternalServerError doesn't implement Unwrap, we can still check the Err field
	assert.Equal(t, underlyingErr, err.Err)
}

