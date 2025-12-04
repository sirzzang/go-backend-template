package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrEmptyHost     = errors.New("empty host")
	ErrInvalidPort   = errors.New("invalid port")
	ErrEmptyUser     = errors.New("empty user")
	ErrEmptyPassword = errors.New("empty password")
	ErrEmptyDBName   = errors.New("empty db name")

	connectTimeout   = 10 // seconds
	operationTimeout = 20 // seconds
	driverPostgres   = "pgx"
)

// Config holds database connection configuration.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string // disable, require, verify-ca, verify-full
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Host == "" {
		return ErrEmptyHost
	}
	if c.Port <= 0 || c.Port > 65535 {
		return ErrInvalidPort
	}
	if c.User == "" {
		return ErrEmptyUser
	}
	if c.Password == "" {
		return ErrEmptyPassword
	}
	if c.DBName == "" {
		return ErrEmptyDBName
	}
	return nil
}

// Repository provides database access methods.
type Repository struct {
	db *sql.DB
}

// New creates a new Repository instance with the given configuration.
func New(config *Config) (*Repository, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	sslMode := config.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		sslMode,
		connectTimeout,
	)

	db, err := sql.Open(driverPostgres, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(connectTimeout)*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Repository{db: db}, nil
}

// Close closes the database connection.
func (r *Repository) Close() error {
	return r.db.Close()
}

// GetContext returns a context with operation timeout.
func (r *Repository) GetContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(operationTimeout)*time.Second)
}

// DB returns the underlying database connection for transactions.
func (r *Repository) DB() *sql.DB {
	return r.db
}

// CreateTables creates all required tables.
func (r *Repository) CreateTables() error {
	queries := []string{
		createUsersTableQuery,
	}

	ctx, cancel := r.GetContext()
	defer cancel()

	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}
