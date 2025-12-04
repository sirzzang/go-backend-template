package user

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/your-org/go-backend-template/internal/pkg/domain"
	"github.com/your-org/go-backend-template/internal/pkg/entity"
	"github.com/your-org/go-backend-template/internal/pkg/repository/postgres"
)

// ========== Mock Repository ==========

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) InsertUser(user *entity.User) (int, error) {
	args := m.Called(user)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) GetUserById(id int) (*entity.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*entity.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetUsers(offset, limit int, onlyActive bool) ([]*entity.User, error) {
	args := m.Called(offset, limit, onlyActive)
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetUserCount(onlyActive bool) (int, error) {
	args := m.Called(onlyActive)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) ExistsUserByEmail(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(user *entity.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserPassword(id int, hashedPassword string) error {
	args := m.Called(id, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUserById(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// ========== Mock Password Hasher ==========

type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) Compare(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

// ========== Test Helper ==========

func setupTestService() (*Service, *MockUserRepository, *MockPasswordHasher) {
	mockRepo := new(MockUserRepository)
	mockHasher := new(MockPasswordHasher)
	service, _ := NewService(mockRepo, mockHasher)
	return service, mockRepo, mockHasher
}

// ========== CreateUser Tests ==========

func TestCreateUser_Success(t *testing.T) {
	svc, mockRepo, mockHasher := setupTestService()

	input := &CreateUserInput{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
		Name:     "Test User",
		Role:     entity.RoleUser,
	}

	mockRepo.On("ExistsUserByEmail", input.Email).Return(false, nil)
	mockHasher.On("Hash", input.Password).Return("hashed_password", nil)
	mockRepo.On("InsertUser", mock.AnythingOfType("*entity.User")).Return(1, nil)

	userId, err := svc.CreateUser(input)

	assert.NoError(t, err)
	assert.Equal(t, 1, userId)
	mockRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
}

func TestCreateUser_InvalidRole(t *testing.T) {
	svc, _, _ := setupTestService()

	input := &CreateUserInput{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
		Name:     "Test User",
		Role:     "invalid_role",
	}

	userId, err := svc.CreateUser(input)

	assert.Error(t, err)
	assert.Equal(t, 0, userId)
	assert.IsType(t, domain.InvalidRoleError{}, err)
}

func TestCreateUser_EmailAlreadyExists(t *testing.T) {
	svc, mockRepo, _ := setupTestService()

	input := &CreateUserInput{
		Email:    "existing@example.com",
		Username: "testuser",
		Password: "password123",
		Name:     "Test User",
		Role:     entity.RoleUser,
	}

	mockRepo.On("ExistsUserByEmail", input.Email).Return(true, nil)

	userId, err := svc.CreateUser(input)

	assert.Error(t, err)
	assert.Equal(t, 0, userId)
	assert.IsType(t, domain.UserAlreadyExistsError{}, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateUser_HashError(t *testing.T) {
	svc, mockRepo, mockHasher := setupTestService()

	input := &CreateUserInput{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
		Name:     "Test User",
		Role:     entity.RoleUser,
	}

	mockRepo.On("ExistsUserByEmail", input.Email).Return(false, nil)
	mockHasher.On("Hash", input.Password).Return("", errors.New("hash error"))

	userId, err := svc.CreateUser(input)

	assert.Error(t, err)
	assert.Equal(t, 0, userId)
	assert.IsType(t, domain.InternalServerError{}, err)
}

// ========== GetUserById Tests ==========

func TestGetUserById_Success(t *testing.T) {
	svc, mockRepo, _ := setupTestService()

	expectedUser := &entity.User{
		Id:       1,
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
		Role:     entity.RoleUser,
	}

	mockRepo.On("GetUserById", 1).Return(expectedUser, nil)

	user, err := svc.GetUserById(1)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestGetUserById_NotFound(t *testing.T) {
	svc, mockRepo, _ := setupTestService()

	mockRepo.On("GetUserById", 999).Return(nil, postgres.ErrNoUser)

	user, err := svc.GetUserById(999)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.IsType(t, domain.UserNotFoundError{}, err)
	mockRepo.AssertExpectations(t)
}

// ========== GetUsers Tests ==========

func TestGetUsers_Success(t *testing.T) {
	svc, mockRepo, _ := setupTestService()

	input := &GetUsersInput{
		Page:       1,
		Size:       10,
		OnlyActive: false,
	}

	expectedUsers := []*entity.User{
		{Id: 1, Email: "user1@example.com"},
		{Id: 2, Email: "user2@example.com"},
	}

	mockRepo.On("GetUsers", 0, 10, false).Return(expectedUsers, nil)
	mockRepo.On("GetUserCount", false).Return(2, nil)

	result, err := svc.GetUsers(input)

	assert.NoError(t, err)
	assert.Equal(t, 2, result.TotalCount)
	assert.Len(t, result.Users, 2)
	mockRepo.AssertExpectations(t)
}

func TestGetUsers_Pagination(t *testing.T) {
	svc, mockRepo, _ := setupTestService()

	input := &GetUsersInput{
		Page:       2,
		Size:       10,
		OnlyActive: true,
	}

	expectedUsers := []*entity.User{
		{Id: 11, Email: "user11@example.com"},
	}

	mockRepo.On("GetUsers", 10, 10, true).Return(expectedUsers, nil)
	mockRepo.On("GetUserCount", true).Return(11, nil)

	result, err := svc.GetUsers(input)

	assert.NoError(t, err)
	assert.Equal(t, 11, result.TotalCount)
	assert.Len(t, result.Users, 1)
	mockRepo.AssertExpectations(t)
}

// ========== UpdateUser Tests ==========

func TestUpdateUser_Success(t *testing.T) {
	svc, mockRepo, _ := setupTestService()

	existingUser := &entity.User{
		Id:       1,
		Email:    "old@example.com",
		Username: "olduser",
		Name:     "Old Name",
		Role:     entity.RoleUser,
		IsActive: true,
	}

	newEmail := "new@example.com"
	newName := "New Name"
	input := &UpdateUserInput{
		Id:    1,
		Email: &newEmail,
		Name:  &newName,
	}

	mockRepo.On("GetUserById", 1).Return(existingUser, nil)
	mockRepo.On("ExistsUserByEmail", newEmail).Return(false, nil)
	mockRepo.On("UpdateUser", mock.AnythingOfType("*entity.User")).Return(nil)

	err := svc.UpdateUser(input)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_NotFound(t *testing.T) {
	svc, mockRepo, _ := setupTestService()

	newName := "New Name"
	input := &UpdateUserInput{
		Id:   999,
		Name: &newName,
	}

	mockRepo.On("GetUserById", 999).Return(nil, postgres.ErrNoUser)

	err := svc.UpdateUser(input)

	assert.Error(t, err)
	assert.IsType(t, domain.UserNotFoundError{}, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_EmailConflict(t *testing.T) {
	svc, mockRepo, _ := setupTestService()

	existingUser := &entity.User{
		Id:    1,
		Email: "old@example.com",
	}

	newEmail := "existing@example.com"
	input := &UpdateUserInput{
		Id:    1,
		Email: &newEmail,
	}

	mockRepo.On("GetUserById", 1).Return(existingUser, nil)
	mockRepo.On("ExistsUserByEmail", newEmail).Return(true, nil)

	err := svc.UpdateUser(input)

	assert.Error(t, err)
	assert.IsType(t, domain.UserAlreadyExistsError{}, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_NoChanges(t *testing.T) {
	svc, mockRepo, _ := setupTestService()

	existingUser := &entity.User{
		Id:       1,
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
		Role:     entity.RoleUser,
		IsActive: true,
	}

	// Input with same values - no changes
	sameName := "Test User"
	input := &UpdateUserInput{
		Id:   1,
		Name: &sameName,
	}

	mockRepo.On("GetUserById", 1).Return(existingUser, nil)
	// UpdateUser should NOT be called since there are no changes

	err := svc.UpdateUser(input)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "UpdateUser")
}

// ========== DeleteUser Tests ==========

func TestDeleteUser_Success(t *testing.T) {
	svc, mockRepo, _ := setupTestService()

	mockRepo.On("DeleteUserById", 1).Return(nil)

	err := svc.DeleteUser(1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteUser_NotFound(t *testing.T) {
	svc, mockRepo, _ := setupTestService()

	mockRepo.On("DeleteUserById", 999).Return(postgres.ErrNoUser)

	err := svc.DeleteUser(999)

	assert.Error(t, err)
	assert.IsType(t, domain.UserNotFoundError{}, err)
	mockRepo.AssertExpectations(t)
}

// ========== Login Tests ==========

func TestLogin_Success(t *testing.T) {
	svc, mockRepo, mockHasher := setupTestService()

	input := &LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedUser := &entity.User{
		Id:       1,
		Email:    "test@example.com",
		Password: "hashed_password",
		IsActive: true,
	}

	mockRepo.On("GetUserByEmail", input.Email).Return(expectedUser, nil)
	mockHasher.On("Compare", "hashed_password", "password123").Return(nil)

	user, err := svc.Login(input)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, mockRepo, _ := setupTestService()

	input := &LoginInput{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	mockRepo.On("GetUserByEmail", input.Email).Return(nil, postgres.ErrNoUser)

	user, err := svc.Login(input)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.IsType(t, domain.InvalidCredentialsError{}, err)
	mockRepo.AssertExpectations(t)
}

func TestLogin_InactiveUser(t *testing.T) {
	svc, mockRepo, _ := setupTestService()

	input := &LoginInput{
		Email:    "inactive@example.com",
		Password: "password123",
	}

	inactiveUser := &entity.User{
		Id:       1,
		Email:    "inactive@example.com",
		IsActive: false,
	}

	mockRepo.On("GetUserByEmail", input.Email).Return(inactiveUser, nil)

	user, err := svc.Login(input)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.IsType(t, domain.InvalidCredentialsError{}, err)
	mockRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, mockRepo, mockHasher := setupTestService()

	input := &LoginInput{
		Email:    "test@example.com",
		Password: "wrong_password",
	}

	existingUser := &entity.User{
		Id:       1,
		Email:    "test@example.com",
		Password: "hashed_password",
		IsActive: true,
	}

	mockRepo.On("GetUserByEmail", input.Email).Return(existingUser, nil)
	mockHasher.On("Compare", "hashed_password", "wrong_password").Return(errors.New("password mismatch"))

	user, err := svc.Login(input)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.IsType(t, domain.InvalidCredentialsError{}, err)
	mockRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
}

// ========== ChangePassword Tests ==========

func TestChangePassword_Success(t *testing.T) {
	svc, mockRepo, mockHasher := setupTestService()

	input := &ChangePasswordInput{
		UserId:          1,
		CurrentPassword: "old_password",
		NewPassword:     "new_password",
	}

	existingUser := &entity.User{
		Id:       1,
		Password: "hashed_old_password",
	}

	mockRepo.On("GetUserById", 1).Return(existingUser, nil)
	mockHasher.On("Compare", "hashed_old_password", "old_password").Return(nil)
	mockHasher.On("Hash", "new_password").Return("hashed_new_password", nil)
	mockRepo.On("UpdateUserPassword", 1, "hashed_new_password").Return(nil)

	err := svc.ChangePassword(input)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	svc, mockRepo, mockHasher := setupTestService()

	input := &ChangePasswordInput{
		UserId:          1,
		CurrentPassword: "wrong_password",
		NewPassword:     "new_password",
	}

	existingUser := &entity.User{
		Id:       1,
		Password: "hashed_old_password",
	}

	mockRepo.On("GetUserById", 1).Return(existingUser, nil)
	mockHasher.On("Compare", "hashed_old_password", "wrong_password").Return(errors.New("mismatch"))

	err := svc.ChangePassword(input)

	assert.Error(t, err)
	assert.IsType(t, domain.InvalidCredentialsError{}, err)
	mockRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
}

// ========== NewService Tests ==========

func TestNewService_NilRepository(t *testing.T) {
	mockHasher := new(MockPasswordHasher)

	svc, err := NewService(nil, mockHasher)

	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestNewService_NilPasswordHasher(t *testing.T) {
	mockRepo := new(MockUserRepository)

	svc, err := NewService(mockRepo, nil)

	assert.Error(t, err)
	assert.Nil(t, svc)
}

