package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/your-org/go-backend-template/internal/app/server/handler"
	"github.com/your-org/go-backend-template/internal/pkg/entity"
)

// ========== Mock JWT Validator ==========

type MockJWTValidator struct {
	mock.Mock
}

func (m *MockJWTValidator) ValidateToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Claims), args.Error(1)
}

// ========== Test Helpers ==========

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// ========== New Middleware Tests ==========

func TestNew_Success(t *testing.T) {
	mockValidator := new(MockJWTValidator)

	middleware, err := New(mockValidator)

	assert.NoError(t, err)
	assert.NotNil(t, middleware)
}

func TestNew_NilValidator(t *testing.T) {
	middleware, err := New(nil)

	assert.Error(t, err)
	assert.Nil(t, middleware)
}

// ========== RequireAuth Tests ==========

func TestRequireAuth_ValidToken(t *testing.T) {
	mockValidator := new(MockJWTValidator)
	middleware, _ := New(mockValidator)

	expectedClaims := &Claims{
		UserId: 123,
		Role:   "admin",
	}
	mockValidator.On("ValidateToken", "valid-token").Return(expectedClaims, nil)

	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		userId := handler.GetUserId(c)
		role := handler.GetUserRole(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id": userId,
			"role":    role,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockValidator.AssertExpectations(t)
}

func TestRequireAuth_NoAuthorizationHeader(t *testing.T) {
	mockValidator := new(MockJWTValidator)
	middleware, _ := New(mockValidator)

	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authorization header is required")
}

func TestRequireAuth_InvalidHeaderFormat(t *testing.T) {
	mockValidator := new(MockJWTValidator)
	middleware, _ := New(mockValidator)

	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	testCases := []struct {
		name        string
		headerValue string
		expected    string
	}{
		{
			name:        "no Bearer prefix",
			headerValue: "just-a-token",
			expected:    "invalid authorization header format",
		},
		{
			name:        "Basic auth instead of Bearer",
			headerValue: "Basic dXNlcjpwYXNz",
			expected:    "invalid authorization header format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tc.headerValue)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), tc.expected)
		})
	}
}

func TestRequireAuth_EmptyToken(t *testing.T) {
	mockValidator := new(MockJWTValidator)
	middleware, _ := New(mockValidator)

	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "token is required")
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	mockValidator := new(MockJWTValidator)
	middleware, _ := New(mockValidator)

	mockValidator.On("ValidateToken", "invalid-token").Return(nil, errors.New("token expired"))

	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
	mockValidator.AssertExpectations(t)
}

func TestRequireAuth_SetsVaryHeader(t *testing.T) {
	mockValidator := new(MockJWTValidator)
	middleware, _ := New(mockValidator)

	mockValidator.On("ValidateToken", "valid-token").Return(&Claims{UserId: 1, Role: "user"}, nil)

	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, "Authorization", w.Header().Get("Vary"))
}

// ========== RequireRole Tests ==========

func TestRequireRole_AllowedRole(t *testing.T) {
	mockValidator := new(MockJWTValidator)
	middleware, _ := New(mockValidator)

	mockValidator.On("ValidateToken", "admin-token").Return(&Claims{UserId: 1, Role: "admin"}, nil)

	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/admin-only", middleware.RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_NotAllowedRole(t *testing.T) {
	mockValidator := new(MockJWTValidator)
	middleware, _ := New(mockValidator)

	mockValidator.On("ValidateToken", "user-token").Return(&Claims{UserId: 1, Role: "user"}, nil)

	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/admin-only", middleware.RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "insufficient permissions")
}

func TestRequireRole_MultipleAllowedRoles(t *testing.T) {
	testCases := []struct {
		name         string
		role         string
		expectedCode int
	}{
		{"admin allowed", "admin", http.StatusOK},
		{"user allowed", "user", http.StatusOK},
		{"viewer not allowed", "viewer", http.StatusForbidden},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockValidator := new(MockJWTValidator)
			middleware, _ := New(mockValidator)

			mockValidator.On("ValidateToken", "token").Return(&Claims{UserId: 1, Role: tc.role}, nil)

			router := setupTestRouter()
			router.Use(middleware.RequireAuth())
			router.GET("/endpoint", middleware.RequireRole("admin", "user"), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/endpoint", nil)
			req.Header.Set("Authorization", "Bearer token")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}
}

// ========== RequireAdmin Tests ==========

func TestRequireAdmin_AdminUser(t *testing.T) {
	mockValidator := new(MockJWTValidator)
	middleware, _ := New(mockValidator)

	mockValidator.On("ValidateToken", "admin-token").Return(&Claims{UserId: 1, Role: entity.RoleAdmin}, nil)

	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/admin", middleware.RequireAdmin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAdmin_NonAdminUser(t *testing.T) {
	mockValidator := new(MockJWTValidator)
	middleware, _ := New(mockValidator)

	mockValidator.On("ValidateToken", "user-token").Return(&Claims{UserId: 1, Role: entity.RoleUser}, nil)

	router := setupTestRouter()
	router.Use(middleware.RequireAuth())
	router.GET("/admin", middleware.RequireAdmin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Context Helpers Tests ==========

func TestContextHelpers_SetAndGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		handler.SetUserId(c, 42)
		handler.SetUserRole(c, "admin")

		userId := handler.GetUserId(c)
		role := handler.GetUserRole(c)

		c.JSON(http.StatusOK, gin.H{
			"user_id": userId,
			"role":    role,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "42")
	assert.Contains(t, w.Body.String(), "admin")
}

// ========== Chain Middleware Tests ==========

func TestMiddlewareChain_AuthThenRole(t *testing.T) {
	mockValidator := new(MockJWTValidator)
	middleware, _ := New(mockValidator)

	mockValidator.On("ValidateToken", "admin-token").Return(&Claims{UserId: 1, Role: "admin"}, nil)

	router := setupTestRouter()

	// Chain: Auth -> RequireAdmin -> Handler
	adminGroup := router.Group("/admin")
	adminGroup.Use(middleware.RequireAuth())
	adminGroup.Use(middleware.RequireAdmin())
	adminGroup.GET("/dashboard", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": handler.GetUserId(c),
			"role":    handler.GetUserRole(c),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":1`)
	assert.Contains(t, w.Body.String(), `"role":"admin"`)
}
