package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewJWTService_Success(t *testing.T) {
	config := JWTConfig{
		SecretKey:     "test-secret-key",
		TokenDuration: time.Hour,
		Issuer:        "test-issuer",
	}

	service, err := NewJWTService(config)

	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestNewJWTService_EmptySecretKey(t *testing.T) {
	config := JWTConfig{
		SecretKey: "",
	}

	service, err := NewJWTService(config)

	assert.Error(t, err)
	assert.Nil(t, service)
}

func TestNewJWTService_DefaultValues(t *testing.T) {
	config := JWTConfig{
		SecretKey: "test-secret-key",
		// No TokenDuration or Issuer provided
	}

	service, err := NewJWTService(config)

	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, 24*time.Hour, service.tokenDuration)
	assert.Equal(t, "go-backend-template", service.issuer)
}

func TestJWTService_GenerateToken(t *testing.T) {
	service, _ := NewJWTService(JWTConfig{
		SecretKey:     "test-secret-key",
		TokenDuration: time.Hour,
	})

	token, err := service.GenerateToken(1, "user")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTService_ValidateToken_Success(t *testing.T) {
	service, _ := NewJWTService(JWTConfig{
		SecretKey:     "test-secret-key",
		TokenDuration: time.Hour,
	})

	token, _ := service.GenerateToken(123, "admin")

	claims, err := service.ValidateToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, 123, claims.UserId)
	assert.Equal(t, "admin", claims.Role)
}

func TestJWTService_ValidateToken_InvalidToken(t *testing.T) {
	service, _ := NewJWTService(JWTConfig{
		SecretKey:     "test-secret-key",
		TokenDuration: time.Hour,
	})

	claims, err := service.ValidateToken("invalid.token.here")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestJWTService_ValidateToken_WrongSignature(t *testing.T) {
	service1, _ := NewJWTService(JWTConfig{
		SecretKey:     "secret-key-1",
		TokenDuration: time.Hour,
	})

	service2, _ := NewJWTService(JWTConfig{
		SecretKey:     "secret-key-2",
		TokenDuration: time.Hour,
	})

	// Generate token with service1
	token, _ := service1.GenerateToken(1, "user")

	// Try to validate with service2 (different secret key)
	claims, err := service2.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTService_ValidateToken_ExpiredToken(t *testing.T) {
	service, _ := NewJWTService(JWTConfig{
		SecretKey:     "test-secret-key",
		TokenDuration: -time.Hour, // Already expired
	})

	token, _ := service.GenerateToken(1, "user")

	claims, err := service.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrExpiredToken, err)
}

func TestJWTService_ValidateToken_MalformedToken(t *testing.T) {
	service, _ := NewJWTService(JWTConfig{
		SecretKey:     "test-secret-key",
		TokenDuration: time.Hour,
	})

	testCases := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"single part", "onlyonepart"},
		{"two parts", "two.parts"},
		{"garbage", "not-a-jwt-at-all"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tc.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestJWTService_TokenContainsClaims(t *testing.T) {
	service, _ := NewJWTService(JWTConfig{
		SecretKey:     "test-secret-key",
		TokenDuration: time.Hour,
		Issuer:        "test-issuer",
	})

	userId := 42
	role := "viewer"

	token, err := service.GenerateToken(userId, role)
	assert.NoError(t, err)

	claims, err := service.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userId, claims.UserId)
	assert.Equal(t, role, claims.Role)
}

func BenchmarkJWTService_GenerateToken(b *testing.B) {
	service, _ := NewJWTService(JWTConfig{
		SecretKey:     "benchmark-secret-key",
		TokenDuration: time.Hour,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GenerateToken(i, "user")
	}
}

func BenchmarkJWTService_ValidateToken(b *testing.B) {
	service, _ := NewJWTService(JWTConfig{
		SecretKey:     "benchmark-secret-key",
		TokenDuration: time.Hour,
	})

	token, _ := service.GenerateToken(1, "user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ValidateToken(token)
	}
}

