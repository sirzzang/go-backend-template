package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/your-org/go-backend-template/internal/app/server"
	"github.com/your-org/go-backend-template/internal/pkg/auth"
	"github.com/your-org/go-backend-template/internal/pkg/repository/postgres"
)

func main() {
	// Load configuration
	config := LoadConfig()
	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize database repository
	repo, err := postgres.New(&postgres.Config{
		Host:     config.DBHost,
		Port:     config.DBPort,
		User:     config.DBUser,
		Password: config.DBPassword,
		DBName:   config.DBName,
		SSLMode:  config.DBSSLMode,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repo.Close()

	log.Println("Connected to database")

	// Create tables if not exists
	if err := repo.CreateTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Initialize JWT service
	jwtService, err := auth.NewJWTService(auth.JWTConfig{
		SecretKey:     config.JWTSecretKey,
		TokenDuration: config.JWTTokenDuration,
	})
	if err != nil {
		log.Fatalf("Failed to create JWT service: %v", err)
	}

	// Initialize password hasher
	passwordHasher := auth.NewPasswordHasher(12) // bcrypt cost 12

	// Create server
	srv, err := server.New(
		&server.Config{
			Host:             config.ServerHost,
			Port:             config.ServerPort,
			Mode:             config.ServerMode,
			CORSAllowOrigins: config.CORSAllowOrigins,
		},
		&server.Dependencies{
			Repository:     repo,
			JWTService:     jwtService,
			PasswordHasher: passwordHasher,
		},
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup routes
	srv.SetupRoutes()

	// Start server in a goroutine
	go func() {
		if err := srv.Run(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}

