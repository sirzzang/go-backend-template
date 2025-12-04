package user

import (
	"errors"

	"github.com/your-org/go-backend-template/internal/pkg/domain"
	"github.com/your-org/go-backend-template/internal/pkg/entity"
	"github.com/your-org/go-backend-template/internal/pkg/repository"
)

var (
	errNilRepository     = errors.New("user repository is nil")
	errNilPasswordHasher = errors.New("password hasher is nil")
)

// Service handles user business logic.
type Service struct {
	userRepo       IUserRepository
	passwordHasher IPasswordHasher
}

// NewService creates a new user service.
func NewService(userRepo IUserRepository, passwordHasher IPasswordHasher) (*Service, error) {
	if userRepo == nil {
		return nil, domain.InternalServerError{Msg: "failed to create user service", Err: errNilRepository}
	}
	if passwordHasher == nil {
		return nil, domain.InternalServerError{Msg: "failed to create user service", Err: errNilPasswordHasher}
	}

	return &Service{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
	}, nil
}

// ========== Create User ==========

func (s *Service) CreateUser(input *CreateUserInput) (int, error) {
	// Validate role
	if !entity.IsValidRole(input.Role) {
		return 0, domain.InvalidRoleError{Role: input.Role}
	}

	// Check if email already exists
	exists, err := s.userRepo.ExistsUserByEmail(input.Email)
	if err != nil {
		return 0, domain.InternalServerError{Msg: "failed to check email existence", Err: err}
	}
	if exists {
		return 0, domain.UserAlreadyExistsError{Email: input.Email}
	}

	// Hash password
	hashedPassword, err := s.passwordHasher.Hash(input.Password)
	if err != nil {
		return 0, domain.InternalServerError{Msg: "failed to hash password", Err: err}
	}

	// Create user entity
	user := &entity.User{
		Email:    input.Email,
		Username: input.Username,
		Password: hashedPassword,
		Name:     input.Name,
		Role:     input.Role,
		IsActive: true,
	}

	// Insert user
	userId, err := s.userRepo.InsertUser(user)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return 0, domain.UserAlreadyExistsError{Email: input.Email}
		}
		return 0, domain.InternalServerError{Msg: "failed to create user", Err: err}
	}

	return userId, nil
}

// ========== Get User ==========

func (s *Service) GetUserById(id int) (*entity.User, error) {
	user, err := s.userRepo.GetUserById(id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, domain.UserNotFoundError{Id: id}
		}
		return nil, domain.InternalServerError{Msg: "failed to get user", Err: err}
	}
	return user, nil
}

func (s *Service) GetUserByEmail(email string) (*entity.User, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, domain.UserNotFoundError{Email: email}
		}
		return nil, domain.InternalServerError{Msg: "failed to get user", Err: err}
	}
	return user, nil
}

// ========== Get Users ==========

type GetUsersResult struct {
	Users      []*entity.User
	TotalCount int
}

func (s *Service) GetUsers(input *GetUsersInput) (*GetUsersResult, error) {
	offset := input.Size * (input.Page - 1)

	users, err := s.userRepo.GetUsers(offset, input.Size, input.OnlyActive)
	if err != nil {
		return nil, domain.InternalServerError{Msg: "failed to get users", Err: err}
	}

	totalCount, err := s.userRepo.GetUserCount(input.OnlyActive)
	if err != nil {
		return nil, domain.InternalServerError{Msg: "failed to get user count", Err: err}
	}

	return &GetUsersResult{
		Users:      users,
		TotalCount: totalCount,
	}, nil
}

// ========== Update User ==========

func (s *Service) UpdateUser(input *UpdateUserInput) error {
	// Get existing user
	user, err := s.userRepo.GetUserById(input.Id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return domain.UserNotFoundError{Id: input.Id}
		}
		return domain.InternalServerError{Msg: "failed to get user", Err: err}
	}

	hasChanges := false

	// Update email if provided
	if input.Email != nil && *input.Email != user.Email {
		// Check if new email already exists
		exists, err := s.userRepo.ExistsUserByEmail(*input.Email)
		if err != nil {
			return domain.InternalServerError{Msg: "failed to check email existence", Err: err}
		}
		if exists {
			return domain.UserAlreadyExistsError{Email: *input.Email}
		}
		user.Email = *input.Email
		hasChanges = true
	}

	// Update username if provided
	if input.Username != nil && *input.Username != user.Username {
		user.Username = *input.Username
		hasChanges = true
	}

	// Update name if provided
	if input.Name != nil && *input.Name != user.Name {
		user.Name = *input.Name
		hasChanges = true
	}

	// Update role if provided
	if input.Role != nil && *input.Role != user.Role {
		if !entity.IsValidRole(*input.Role) {
			return domain.InvalidRoleError{Role: *input.Role}
		}
		user.Role = *input.Role
		hasChanges = true
	}

	// Update is_active if provided
	if input.IsActive != nil && *input.IsActive != user.IsActive {
		user.IsActive = *input.IsActive
		hasChanges = true
	}

	// Only update if there are changes
	if hasChanges {
		if err := s.userRepo.UpdateUser(user); err != nil {
			if errors.Is(err, repository.ErrDuplicateEmail) {
				return domain.UserAlreadyExistsError{Email: user.Email}
			}
			return domain.InternalServerError{Msg: "failed to update user", Err: err}
		}
	}

	return nil
}

// ========== Change Password ==========

func (s *Service) ChangePassword(input *ChangePasswordInput) error {
	// Get user
	user, err := s.userRepo.GetUserById(input.UserId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return domain.UserNotFoundError{Id: input.UserId}
		}
		return domain.InternalServerError{Msg: "failed to get user", Err: err}
	}

	// Verify current password
	if err := s.passwordHasher.Compare(user.Password, input.CurrentPassword); err != nil {
		return domain.InvalidCredentialsError{}
	}

	// Hash new password
	hashedPassword, err := s.passwordHasher.Hash(input.NewPassword)
	if err != nil {
		return domain.InternalServerError{Msg: "failed to hash password", Err: err}
	}

	// Update password
	if err := s.userRepo.UpdateUserPassword(input.UserId, hashedPassword); err != nil {
		return domain.InternalServerError{Msg: "failed to update password", Err: err}
	}

	return nil
}

// ========== Delete User ==========

func (s *Service) DeleteUser(id int) error {
	if err := s.userRepo.DeleteUserById(id); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return domain.UserNotFoundError{Id: id}
		}
		return domain.InternalServerError{Msg: "failed to delete user", Err: err}
	}
	return nil
}

// ========== Login ==========

func (s *Service) Login(input *LoginInput) (*entity.User, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, domain.InvalidCredentialsError{}
		}
		return nil, domain.InternalServerError{Msg: "failed to get user", Err: err}
	}

	// Check if user is active
	if !user.IsActive {
		return nil, domain.InvalidCredentialsError{}
	}

	// Verify password
	if err := s.passwordHasher.Compare(user.Password, input.Password); err != nil {
		return nil, domain.InvalidCredentialsError{}
	}

	return user, nil
}
