package user

import (
	"errors"
	"fmt"
	"strings"

	"github.com/your-org/go-backend-template/internal/pkg/entity"
)

// ========== Request DTOs ==========

// CreateUserRequest represents the request body for creating a user.
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=100"`
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Role     string `json:"role" binding:"required"`
}

func (r *CreateUserRequest) Validate() error {
	if !entity.IsValidRole(r.Role) {
		return fmt.Errorf("invalid role: %s, must be one of: admin, user, viewer", r.Role)
	}
	return nil
}

// UpdateUserRequest represents the request body for updating a user.
type UpdateUserRequest struct {
	Email    *string `json:"email" binding:"omitempty,email"`
	Username *string `json:"username" binding:"omitempty,min=3,max=50"`
	Name     *string `json:"name" binding:"omitempty,min=1,max=100"`
	Role     *string `json:"role"`
	IsActive *bool   `json:"is_active"`
}

func (r *UpdateUserRequest) Validate() error {
	if r.Role != nil && !entity.IsValidRole(*r.Role) {
		return fmt.Errorf("invalid role: %s, must be one of: admin, user, viewer", *r.Role)
	}
	return nil
}

// ChangePasswordRequest represents the request body for changing password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=100"`
}

func (r *ChangePasswordRequest) Validate() error {
	if r.CurrentPassword == r.NewPassword {
		return errors.New("new password must be different from current password")
	}
	return nil
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// GetUsersQuery represents query parameters for listing users.
type GetUsersQuery struct {
	Page       *int  `form:"page" binding:"omitempty,min=1"`
	Size       *int  `form:"size" binding:"omitempty,min=1,max=100"`
	OnlyActive *bool `form:"only_active"`
}

func (q *GetUsersQuery) GetPage() int {
	if q.Page == nil || *q.Page < 1 {
		return 1
	}
	return *q.Page
}

func (q *GetUsersQuery) GetSize() int {
	if q.Size == nil || *q.Size < 1 {
		return 20
	}
	return *q.Size
}

func (q *GetUsersQuery) GetOnlyActive() bool {
	if q.OnlyActive == nil {
		return false
	}
	return *q.OnlyActive
}

// ========== Response DTOs ==========

// UserResponse represents a user in API responses.
type UserResponse struct {
	Id        int    `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt int64  `json:"created_at"` // Unix timestamp
	UpdatedAt int64  `json:"updated_at"` // Unix timestamp
}

// ToUserResponse converts an entity.User to UserResponse.
func ToUserResponse(user *entity.User) *UserResponse {
	return &UserResponse{
		Id:        user.Id,
		Email:     user.Email,
		Username:  user.Username,
		Name:      user.Name,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}
}

// ToUserResponseList converts a list of entity.User to UserResponse list.
func ToUserResponseList(users []*entity.User) []*UserResponse {
	result := make([]*UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, ToUserResponse(user))
	}
	return result
}

// CreateUserResponse represents the response for user creation.
type CreateUserResponse struct {
	Id int `json:"id"`
}

// GetUsersResponse represents the response for listing users.
type GetUsersResponse struct {
	TotalCount int             `json:"total_count"`
	Count      int             `json:"count"`
	Data       []*UserResponse `json:"data"`
}

// LoginResponse represents the response for user login.
type LoginResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

// MessageResponse represents a simple message response.
type MessageResponse struct {
	Message string `json:"message"`
}

// ========== Validation Helpers ==========

// ValidateEmail performs additional email validation if needed.
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email is required")
	}
	if len(email) > 255 {
		return errors.New("email must be less than 255 characters")
	}
	return nil
}

