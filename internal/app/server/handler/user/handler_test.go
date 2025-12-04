package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/your-org/go-backend-template/internal/app/server/handler"
	"github.com/your-org/go-backend-template/internal/app/server/service/user"
	"github.com/your-org/go-backend-template/internal/pkg/domain"
	"github.com/your-org/go-backend-template/internal/pkg/entity"
)

// ========== Mock Service ==========

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(input *user.CreateUserInput) (int, error) {
	args := m.Called(input)
	return args.Int(0), args.Error(1)
}

func (m *MockUserService) GetUserById(id int) (*entity.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(email string) (*entity.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserService) GetUsers(input *user.GetUsersInput) (*user.GetUsersResult, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.GetUsersResult), args.Error(1)
}

func (m *MockUserService) UpdateUser(input *user.UpdateUserInput) error {
	args := m.Called(input)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) ChangePassword(input *user.ChangePasswordInput) error {
	args := m.Called(input)
	return args.Error(0)
}

func (m *MockUserService) Login(input *user.LoginInput) (*entity.User, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

// ========== Mock JWT Service ==========

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(userId int, role string) (string, error) {
	args := m.Called(userId, role)
	return args.String(0), args.Error(1)
}

// ========== Test Helpers ==========

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func setupTestHandler(mockSvc *MockUserService, mockJWT *MockJWTService) *Handler {
	return &Handler{
		BaseHandler: handler.BaseHandler{},
		userService: &user.Service{}, // We'll use mock directly in tests
		jwtService:  mockJWT,
	}
}

// testHandlerWithMock creates a test handler that uses mocked service
type testHandler struct {
	handler.BaseHandler
	mockService *MockUserService
	mockJWT     *MockJWTService
}

func newTestHandler() (*testHandler, *MockUserService, *MockJWTService) {
	mockSvc := new(MockUserService)
	mockJWT := new(MockJWTService)
	return &testHandler{
		BaseHandler: handler.BaseHandler{},
		mockService: mockSvc,
		mockJWT:     mockJWT,
	}, mockSvc, mockJWT
}

// ========== CreateUser Tests ==========

func TestHandler_CreateUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mockSvc, _ := newTestHandler()

	router := gin.New()
	router.POST("/users", func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.HandleBindingError(c, err)
			return
		}

		if err := req.Validate(); err != nil {
			h.HandleValidationError(c, err.Error())
			return
		}

		input := &user.CreateUserInput{
			Email:    req.Email,
			Username: req.Username,
			Password: req.Password,
			Name:     req.Name,
			Role:     req.Role,
		}

		userId, err := mockSvc.CreateUser(input)
		if err != nil {
			h.HandleDomainError(c, err)
			return
		}

		h.HandleSuccess(c, http.StatusCreated, &CreateUserResponse{Id: userId})
	})

	mockSvc.On("CreateUser", mock.AnythingOfType("*user.CreateUserInput")).Return(1, nil)

	reqBody := CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
		Name:     "Test User",
		Role:     "user",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp CreateUserResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Id)
	mockSvc.AssertExpectations(t)
}

func TestHandler_CreateUser_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := newTestHandler()

	router := gin.New()
	router.POST("/users", func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.HandleBindingError(c, err)
			return
		}
	})

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_CreateUser_InvalidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := newTestHandler()

	router := gin.New()
	router.POST("/users", func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.HandleBindingError(c, err)
			return
		}

		if err := req.Validate(); err != nil {
			h.HandleValidationError(c, err.Error())
			return
		}
	})

	reqBody := CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
		Name:     "Test User",
		Role:     "invalid_role",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_CreateUser_EmailAlreadyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mockSvc, _ := newTestHandler()

	router := gin.New()
	router.POST("/users", func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.HandleBindingError(c, err)
			return
		}

		input := &user.CreateUserInput{
			Email:    req.Email,
			Username: req.Username,
			Password: req.Password,
			Name:     req.Name,
			Role:     req.Role,
		}

		userId, err := mockSvc.CreateUser(input)
		if err != nil {
			h.HandleDomainError(c, err)
			return
		}

		h.HandleSuccess(c, http.StatusCreated, &CreateUserResponse{Id: userId})
	})

	mockSvc.On("CreateUser", mock.AnythingOfType("*user.CreateUserInput")).
		Return(0, domain.UserAlreadyExistsError{Email: "existing@example.com"})

	reqBody := CreateUserRequest{
		Email:    "existing@example.com",
		Username: "testuser",
		Password: "password123",
		Name:     "Test User",
		Role:     "user",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockSvc.AssertExpectations(t)
}

// ========== GetUser Tests ==========

func TestHandler_GetUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mockSvc, _ := newTestHandler()

	router := gin.New()
	router.GET("/users/:id", func(c *gin.Context) {
		userId := 1 // simplified for test

		gotUser, err := mockSvc.GetUserById(userId)
		if err != nil {
			h.HandleDomainError(c, err)
			return
		}

		h.HandleSuccess(c, http.StatusOK, ToUserResponse(gotUser))
	})

	now := time.Now()
	expectedUser := &entity.User{
		Id:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		Name:      "Test User",
		Role:      "user",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockSvc.On("GetUserById", 1).Return(expectedUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Id)
	assert.Equal(t, "test@example.com", resp.Email)
	mockSvc.AssertExpectations(t)
}

func TestHandler_GetUser_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mockSvc, _ := newTestHandler()

	router := gin.New()
	router.GET("/users/:id", func(c *gin.Context) {
		userId := 999

		gotUser, err := mockSvc.GetUserById(userId)
		if err != nil {
			h.HandleDomainError(c, err)
			return
		}

		h.HandleSuccess(c, http.StatusOK, ToUserResponse(gotUser))
	})

	mockSvc.On("GetUserById", 999).Return(nil, domain.UserNotFoundError{Id: 999})

	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

// ========== GetUsers Tests ==========

func TestHandler_GetUsers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mockSvc, _ := newTestHandler()

	router := gin.New()
	router.GET("/users", func(c *gin.Context) {
		input := &user.GetUsersInput{
			Page:       1,
			Size:       20,
			OnlyActive: false,
		}

		result, err := mockSvc.GetUsers(input)
		if err != nil {
			h.HandleDomainError(c, err)
			return
		}

		resp := &GetUsersResponse{
			TotalCount: result.TotalCount,
			Count:      len(result.Users),
			Data:       ToUserResponseList(result.Users),
		}

		h.HandleSuccess(c, http.StatusOK, resp)
	})

	now := time.Now()
	mockResult := &user.GetUsersResult{
		Users: []*entity.User{
			{Id: 1, Email: "user1@example.com", Username: "user1", Name: "User 1", Role: "user", CreatedAt: now, UpdatedAt: now},
			{Id: 2, Email: "user2@example.com", Username: "user2", Name: "User 2", Role: "admin", CreatedAt: now, UpdatedAt: now},
		},
		TotalCount: 2,
	}

	mockSvc.On("GetUsers", mock.AnythingOfType("*user.GetUsersInput")).Return(mockResult, nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp GetUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 2, resp.TotalCount)
	assert.Equal(t, 2, resp.Count)
	assert.Len(t, resp.Data, 2)
	mockSvc.AssertExpectations(t)
}

// ========== Login Tests ==========

func TestHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mockSvc, mockJWT := newTestHandler()

	router := gin.New()
	router.POST("/auth/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.HandleBindingError(c, err)
			return
		}

		input := &user.LoginInput{
			Email:    req.Email,
			Password: req.Password,
		}

		loggedInUser, err := mockSvc.Login(input)
		if err != nil {
			h.HandleDomainError(c, err)
			return
		}

		token, err := mockJWT.GenerateToken(loggedInUser.Id, loggedInUser.Role)
		if err != nil {
			h.HandleDomainError(c, err)
			return
		}

		resp := &LoginResponse{
			Token: token,
			User:  ToUserResponse(loggedInUser),
		}

		h.HandleSuccess(c, http.StatusOK, resp)
	})

	now := time.Now()
	mockUser := &entity.User{
		Id:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		Name:      "Test User",
		Role:      "user",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockSvc.On("Login", mock.AnythingOfType("*user.LoginInput")).Return(mockUser, nil)
	mockJWT.On("GenerateToken", 1, "user").Return("mock_jwt_token", nil)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "mock_jwt_token", resp.Token)
	assert.Equal(t, 1, resp.User.Id)
	mockSvc.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mockSvc, _ := newTestHandler()

	router := gin.New()
	router.POST("/auth/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.HandleBindingError(c, err)
			return
		}

		input := &user.LoginInput{
			Email:    req.Email,
			Password: req.Password,
		}

		_, err := mockSvc.Login(input)
		if err != nil {
			h.HandleDomainError(c, err)
			return
		}
	})

	mockSvc.On("Login", mock.AnythingOfType("*user.LoginInput")).
		Return(nil, domain.InvalidCredentialsError{})

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "wrong_password",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertExpectations(t)
}

// ========== DeleteUser Tests ==========

func TestHandler_DeleteUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mockSvc, _ := newTestHandler()

	router := gin.New()
	router.DELETE("/users/:id", func(c *gin.Context) {
		userId := 1

		if err := mockSvc.DeleteUser(userId); err != nil {
			h.HandleDomainError(c, err)
			return
		}

		h.HandleSuccess(c, http.StatusOK, &MessageResponse{Message: "user deleted successfully"})
	})

	mockSvc.On("DeleteUser", 1).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "user deleted successfully", resp.Message)
	mockSvc.AssertExpectations(t)
}

// ========== DTO Validation Tests ==========

func TestCreateUserRequest_Validate_ValidRole(t *testing.T) {
	req := &CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
		Name:     "Test User",
		Role:     "admin",
	}

	err := req.Validate()
	assert.NoError(t, err)
}

func TestCreateUserRequest_Validate_InvalidRole(t *testing.T) {
	req := &CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
		Name:     "Test User",
		Role:     "superadmin",
	}

	err := req.Validate()
	assert.Error(t, err)
}

func TestUpdateUserRequest_Validate_ValidRole(t *testing.T) {
	role := "viewer"
	req := &UpdateUserRequest{
		Role: &role,
	}

	err := req.Validate()
	assert.NoError(t, err)
}

func TestUpdateUserRequest_Validate_InvalidRole(t *testing.T) {
	role := "invalid"
	req := &UpdateUserRequest{
		Role: &role,
	}

	err := req.Validate()
	assert.Error(t, err)
}

func TestChangePasswordRequest_Validate_SamePassword(t *testing.T) {
	req := &ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "password123",
	}

	err := req.Validate()
	assert.Error(t, err)
}

func TestChangePasswordRequest_Validate_DifferentPassword(t *testing.T) {
	req := &ChangePasswordRequest{
		CurrentPassword: "old_password",
		NewPassword:     "new_password",
	}

	err := req.Validate()
	assert.NoError(t, err)
}

func TestGetUsersQuery_GetPage(t *testing.T) {
	tests := []struct {
		name     string
		page     *int
		expected int
	}{
		{"nil page returns 1", nil, 1},
		{"page 0 returns 1", intPtr(0), 1},
		{"page 1 returns 1", intPtr(1), 1},
		{"page 5 returns 5", intPtr(5), 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &GetUsersQuery{Page: tt.page}
			assert.Equal(t, tt.expected, q.GetPage())
		})
	}
}

func TestGetUsersQuery_GetSize(t *testing.T) {
	tests := []struct {
		name     string
		size     *int
		expected int
	}{
		{"nil size returns 20", nil, 20},
		{"size 0 returns 20", intPtr(0), 20},
		{"size 10 returns 10", intPtr(10), 10},
		{"size 100 returns 100", intPtr(100), 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &GetUsersQuery{Size: tt.size}
			assert.Equal(t, tt.expected, q.GetSize())
		})
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}

