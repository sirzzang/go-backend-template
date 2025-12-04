package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	authMiddleware "github.com/your-org/go-backend-template/internal/app/server/middleware/auth"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	SecretKey     string
	TokenDuration time.Duration
	Issuer        string
}

// JWTService handles JWT operations.
type JWTService struct {
	secretKey     []byte
	tokenDuration time.Duration
	issuer        string
}

// CustomClaims represents JWT claims.
type CustomClaims struct {
	UserId int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service.
func NewJWTService(config JWTConfig) (*JWTService, error) {
	if config.SecretKey == "" {
		return nil, errors.New("secret key is required")
	}

	tokenDuration := config.TokenDuration
	if tokenDuration == 0 {
		tokenDuration = 24 * time.Hour // default 24 hours
	}

	issuer := config.Issuer
	if issuer == "" {
		issuer = "go-backend-template"
	}

	return &JWTService{
		secretKey:     []byte(config.SecretKey),
		tokenDuration: tokenDuration,
		issuer:        issuer,
	}, nil
}

// GenerateToken generates a new JWT token for the given user.
func (s *JWTService) GenerateToken(userId int, role string) (string, error) {
	now := time.Now()

	claims := CustomClaims{
		UserId: userId,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken validates a JWT token and returns the claims.
func (s *JWTService) ValidateToken(tokenString string) (*authMiddleware.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return &authMiddleware.Claims{
		UserId: claims.UserId,
		Role:   claims.Role,
	}, nil
}

