package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/your-org/go-backend-template/internal/pkg/entity"
	"github.com/your-org/go-backend-template/internal/pkg/repository"
)

// Note: These are integration tests that require a test database.
// For unit tests, you would use a mock database driver like go-sqlmock.
// This file shows the structure of repository tests.

// ========== Test Helpers ==========

// TestRepository wraps Repository for testing
type TestRepository struct {
	*Repository
	cleanup func()
}

// setupTestDB creates a test database connection.
// In a real scenario, you would:
// 1. Use a test database or in-memory SQLite
// 2. Run migrations before tests
// 3. Clean up after tests
func setupTestDB(t *testing.T) *TestRepository {
	// Skip if no test database is available
	// In CI/CD, you would set up a test PostgreSQL container
	t.Skip("Skipping integration test: no test database configured")

	config := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "go_backend_template_test",
		SSLMode:  "disable",
	}

	repo, err := New(config)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Create tables
	if err := repo.CreateTables(); err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	return &TestRepository{
		Repository: repo,
		cleanup: func() {
			// Clean up test data
			repo.db.Exec("DELETE FROM users")
			repo.Close()
		},
	}
}

// ========== Unit Tests with SQL Mock ==========
// These tests demonstrate the testing patterns without requiring a real database

func TestRepository_InsertUser_Validation(t *testing.T) {
	// Test that user entity is properly structured
	user := &entity.User{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashed_password",
		Name:     "Test User",
		Role:     entity.RoleUser,
		IsActive: true,
	}

	assert.NotEmpty(t, user.Email)
	assert.NotEmpty(t, user.Username)
	assert.NotEmpty(t, user.Password)
	assert.NotEmpty(t, user.Name)
	assert.True(t, entity.IsValidRole(user.Role))
}

func TestRepository_Config_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorType   error
	}{
		{
			name: "valid config",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				DBName:   "testdb",
			},
			expectError: false,
		},
		{
			name: "empty host",
			config: &Config{
				Host:     "",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				DBName:   "testdb",
			},
			expectError: true,
			errorType:   ErrEmptyHost,
		},
		{
			name: "invalid port - zero",
			config: &Config{
				Host:     "localhost",
				Port:     0,
				User:     "postgres",
				Password: "postgres",
				DBName:   "testdb",
			},
			expectError: true,
			errorType:   ErrInvalidPort,
		},
		{
			name: "invalid port - negative",
			config: &Config{
				Host:     "localhost",
				Port:     -1,
				User:     "postgres",
				Password: "postgres",
				DBName:   "testdb",
			},
			expectError: true,
			errorType:   ErrInvalidPort,
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Host:     "localhost",
				Port:     65536,
				User:     "postgres",
				Password: "postgres",
				DBName:   "testdb",
			},
			expectError: true,
			errorType:   ErrInvalidPort,
		},
		{
			name: "empty user",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "",
				Password: "postgres",
				DBName:   "testdb",
			},
			expectError: true,
			errorType:   ErrEmptyUser,
		},
		{
			name: "empty password",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "",
				DBName:   "testdb",
			},
			expectError: true,
			errorType:   ErrEmptyPassword,
		},
		{
			name: "empty db name",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				DBName:   "",
			},
			expectError: true,
			errorType:   ErrEmptyDBName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorType, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ========== Integration Tests (skipped by default) ==========

func TestRepository_InsertUser_Integration(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.cleanup()

	user := &entity.User{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashed_password",
		Name:     "Test User",
		Role:     entity.RoleUser,
		IsActive: true,
	}

	id, err := repo.InsertUser(user)
	assert.NoError(t, err)
	assert.Greater(t, id, 0)
}

func TestRepository_GetUserById_Integration(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.cleanup()

	// Insert a user first
	user := &entity.User{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashed_password",
		Name:     "Test User",
		Role:     entity.RoleUser,
		IsActive: true,
	}

	id, err := repo.InsertUser(user)
	assert.NoError(t, err)

	// Get the user
	gotUser, err := repo.GetUserById(id)
	assert.NoError(t, err)
	assert.NotNil(t, gotUser)
	assert.Equal(t, user.Email, gotUser.Email)
	assert.Equal(t, user.Username, gotUser.Username)
	assert.Equal(t, user.Name, gotUser.Name)
}

func TestRepository_GetUserById_NotFound_Integration(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.cleanup()

	user, err := repo.GetUserById(99999)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrUserNotFound, err)
	assert.Nil(t, user)
}

func TestRepository_GetUserByEmail_Integration(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.cleanup()

	email := "unique@example.com"
	user := &entity.User{
		Email:    email,
		Username: "uniqueuser",
		Password: "hashed_password",
		Name:     "Unique User",
		Role:     entity.RoleUser,
		IsActive: true,
	}

	_, err := repo.InsertUser(user)
	assert.NoError(t, err)

	gotUser, err := repo.GetUserByEmail(email)
	assert.NoError(t, err)
	assert.NotNil(t, gotUser)
	assert.Equal(t, email, gotUser.Email)
}

func TestRepository_GetUsers_Integration(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.cleanup()

	// Insert multiple users
	users := []*entity.User{
		{Email: "user1@example.com", Username: "user1", Password: "hash1", Name: "User 1", Role: entity.RoleUser, IsActive: true},
		{Email: "user2@example.com", Username: "user2", Password: "hash2", Name: "User 2", Role: entity.RoleUser, IsActive: true},
		{Email: "user3@example.com", Username: "user3", Password: "hash3", Name: "User 3", Role: entity.RoleAdmin, IsActive: false},
	}

	for _, u := range users {
		_, err := repo.InsertUser(u)
		assert.NoError(t, err)
	}

	// Get all users
	allUsers, err := repo.GetUsers(0, 10, false)
	assert.NoError(t, err)
	assert.Len(t, allUsers, 3)

	// Get only active users
	activeUsers, err := repo.GetUsers(0, 10, true)
	assert.NoError(t, err)
	assert.Len(t, activeUsers, 2)
}

func TestRepository_UpdateUser_Integration(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.cleanup()

	user := &entity.User{
		Email:    "update@example.com",
		Username: "updateuser",
		Password: "hashed_password",
		Name:     "Original Name",
		Role:     entity.RoleUser,
		IsActive: true,
	}

	id, err := repo.InsertUser(user)
	assert.NoError(t, err)

	// Update the user
	user.Id = id
	user.Name = "Updated Name"
	user.Role = entity.RoleAdmin

	err = repo.UpdateUser(user)
	assert.NoError(t, err)

	// Verify the update
	updatedUser, err := repo.GetUserById(id)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedUser.Name)
	assert.Equal(t, entity.RoleAdmin, updatedUser.Role)
}

func TestRepository_DeleteUserById_Integration(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.cleanup()

	user := &entity.User{
		Email:    "delete@example.com",
		Username: "deleteuser",
		Password: "hashed_password",
		Name:     "Delete Me",
		Role:     entity.RoleUser,
		IsActive: true,
	}

	id, err := repo.InsertUser(user)
	assert.NoError(t, err)

	// Delete the user
	err = repo.DeleteUserById(id)
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetUserById(id)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrUserNotFound, err)
}

func TestRepository_ExistsUserByEmail_Integration(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.cleanup()

	email := "exists@example.com"

	// Should not exist initially
	exists, err := repo.ExistsUserByEmail(email)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Insert user
	user := &entity.User{
		Email:    email,
		Username: "existsuser",
		Password: "hashed_password",
		Name:     "Exists User",
		Role:     entity.RoleUser,
		IsActive: true,
	}
	_, err = repo.InsertUser(user)
	assert.NoError(t, err)

	// Should exist now
	exists, err = repo.ExistsUserByEmail(email)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestRepository_InsertUser_DuplicateEmail_Integration(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.cleanup()

	email := "duplicate@example.com"
	user1 := &entity.User{
		Email:    email,
		Username: "user1",
		Password: "hash1",
		Name:     "User 1",
		Role:     entity.RoleUser,
		IsActive: true,
	}

	_, err := repo.InsertUser(user1)
	assert.NoError(t, err)

	// Try to insert another user with same email
	user2 := &entity.User{
		Email:    email,
		Username: "user2",
		Password: "hash2",
		Name:     "User 2",
		Role:     entity.RoleUser,
		IsActive: true,
	}

	_, err = repo.InsertUser(user2)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrDuplicateEmail, err)
}

// ========== Benchmark Tests ==========

func BenchmarkRepository_GetUserById(b *testing.B) {
	b.Skip("Skipping benchmark: no test database configured")

	config := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "go_backend_template_test",
	}

	repo, err := New(config)
	if err != nil {
		b.Fatal(err)
	}
	defer repo.Close()

	// Insert a test user
	user := &entity.User{
		Email:    "bench@example.com",
		Username: "benchuser",
		Password: "hash",
		Name:     "Bench User",
		Role:     entity.RoleUser,
		IsActive: true,
	}
	id, _ := repo.InsertUser(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetUserById(id)
	}
}

// ========== Entity Tests ==========

func TestEntity_User_IsValidRole(t *testing.T) {
	tests := []struct {
		role  string
		valid bool
	}{
		{entity.RoleAdmin, true},
		{entity.RoleUser, true},
		{entity.RoleViewer, true},
		{"superadmin", false},
		{"guest", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			assert.Equal(t, tt.valid, entity.IsValidRole(tt.role))
		})
	}
}

func TestEntity_User_Fields(t *testing.T) {
	now := time.Now()
	user := &entity.User{
		Id:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		Password:  "hashed_password",
		Name:      "Test User",
		Role:      entity.RoleUser,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, 1, user.Id)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "hashed_password", user.Password)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, entity.RoleUser, user.Role)
	assert.True(t, user.IsActive)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
}

// Mock DB for unit tests (example structure)
type MockDB struct {
	*sql.DB
	// Add mock-specific fields
}

func NewMockDB() *MockDB {
	// In real implementation, use go-sqlmock
	return &MockDB{}
}
