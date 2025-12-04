package server

import (
	"errors"
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	userHandler "github.com/your-org/go-backend-template/internal/app/server/handler/user"
	"github.com/your-org/go-backend-template/internal/app/server/middleware/auth"
	"github.com/your-org/go-backend-template/internal/app/server/routes"
	userService "github.com/your-org/go-backend-template/internal/app/server/service/user"
	pkgAuth "github.com/your-org/go-backend-template/internal/pkg/auth"
	"github.com/your-org/go-backend-template/internal/pkg/repository/postgres"
)

const (
	prefixAPI = "/api"
)

// Config holds server configuration.
type Config struct {
	Host             string
	Port             int
	Mode             string   // debug, release, test
	CORSAllowOrigins []string // allowed CORS origins
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("invalid port")
	}
	return nil
}

// Dependencies holds all external dependencies for the server.
type Dependencies struct {
	Repository     *postgres.Repository
	JWTService     *pkgAuth.JWTService
	PasswordHasher *pkgAuth.PasswordHasher
}

// Validate checks if all required dependencies are provided.
func (d *Dependencies) Validate() error {
	if d.Repository == nil {
		return errors.New("repository is nil")
	}
	if d.JWTService == nil {
		return errors.New("jwt service is nil")
	}
	if d.PasswordHasher == nil {
		return errors.New("password hasher is nil")
	}
	return nil
}

// Server represents the HTTP server.
type Server struct {
	config         *Config
	router         *gin.Engine
	handlers       *routes.Handlers
	authMiddleware *auth.Middleware
}

// New creates a new Server with the given configuration and dependencies.
func New(config *Config, deps *Dependencies) (*Server, error) {
	// Validate configuration
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Validate dependencies
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid dependencies: %w", err)
	}

	// Set Gin mode
	switch config.Mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		log.Printf("unknown server mode: %s, using debug mode\n", config.Mode)
		gin.SetMode(gin.DebugMode)
	}

	// Initialize auth middleware
	authMiddleware, err := auth.New(deps.JWTService)
	if err != nil {
		return nil, fmt.Errorf("failed to init auth middleware: %w", err)
	}

	// Initialize user service
	userSvc, err := userService.NewService(deps.Repository, deps.PasswordHasher)
	if err != nil {
		return nil, fmt.Errorf("failed to init user service: %w", err)
	}

	// Initialize handlers
	userH := userHandler.NewHandler(userSvc, deps.JWTService)

	handlers := &routes.Handlers{
		User: userH,
	}

	// Setup Gin router
	router := gin.Default()

	// Configure CORS
	allowOrigins := config.CORSAllowOrigins
	if len(allowOrigins) == 0 {
		allowOrigins = []string{"*"} // default to allow all (change in production)
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	return &Server{
		config:         config,
		router:         router,
		handlers:       handlers,
		authMiddleware: authMiddleware,
	}, nil
}

// SetupRoutes configures all routes.
func (s *Server) SetupRoutes() {
	// Health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	apiRoutes := s.router.Group(prefixAPI)
	apiRoutes.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Go Backend Template API",
			"version": "1.0.0",
		})
	})

	routes.SetupRoutes(apiRoutes, s.handlers, s.authMiddleware)
}

// Run starts the HTTP server.
func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	log.Printf("Starting server on %s\n", addr)
	return s.router.Run(addr)
}

// Router returns the underlying gin router (useful for testing).
func (s *Server) Router() *gin.Engine {
	return s.router
}

