# Go Backend Template

**English** | [한국어](./README.md)

> Every time I built a backend server with Go, I'd start thinking "let me structure this more elegantly,"
> only to end up rewriting everything halfway through the project. This is the structure I **finally settled on**.
>
> Created to avoid refactoring hell, and for future me to reference when asking "how did I do that again?"

A layered architecture Go backend template.

## Table of Contents

- [Architecture](#architecture)
- [Layer Responsibilities](#layer-responsibilities)
- [Key Implementation Details](#key-implementation-details)
- [Getting Started](#getting-started)
- [Testing](#testing)
- [Adding a New Domain](#adding-a-new-domain)
- [Scaling Considerations](#scaling-considerations)

## Architecture

This template follows a layered architecture pattern:

```
Route → Middleware → Handler → Service → Repository
```

### Directory Structure

```
go-backend-template/
├── cmd/
│   └── server/              # Application entry point
│       ├── main.go          # Main function
│       └── config.go        # Configuration loading
├── internal/
│   ├── app/
│   │   └── server/          # HTTP server implementation
│   │       ├── server.go    # Server setup and configuration
│   │       ├── handler/     # HTTP handlers (controllers)
│   │       │   ├── base.go  # Base handler with common methods
│   │       │   ├── context.go # Context helpers
│   │       │   └── user/    # User domain handlers
│   │       │       ├── handler.go
│   │       │       └── dto.go
│   │       ├── middleware/  # HTTP middlewares
│   │       │   └── auth/
│   │       │       └── auth.go
│   │       ├── routes/      # Route definitions
│   │       │   ├── routes.go
│   │       │   └── user.go
│   │       └── service/     # Business logic layer
│   │           └── user/
│   │               ├── service.go
│   │               ├── input.go
│   │               └── dependencies.go
│   └── pkg/                 # Shared internal packages
│       ├── auth/            # Authentication utilities
│       │   ├── jwt.go
│       │   └── password.go
│       ├── domain/          # Domain errors
│       │   └── errors.go
│       ├── entity/          # Domain entities
│       │   └── user.go
│       └── repository/      # Data access layer
│           └── postgres/
│               ├── repository.go
│               ├── errors.go
│               └── user.go
├── build/
│   └── Dockerfile           # Docker build file
├── deployments/
│   ├── docker-compose.yml   # Docker Compose for local dev
│   └── .env.example         # Environment variables example
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Layer Responsibilities

### 1. Handler Layer (`handler/`)
- Parse HTTP requests (path params, query, body)
- Validate request format
- Convert request DTO to service input
- Call service methods
- Convert service output to response DTO
- Handle errors and send HTTP responses

### 2. Service Layer (`service/`)
- Implement business logic
- Orchestrate repository calls
- Return domain errors
- No HTTP-related code

### 3. Repository Layer (`repository/`)
- Database access
- SQL queries
- Return database errors

### 4. Domain Layer (`domain/`, `entity/`)
- Domain entities
- Domain errors with HTTP status mapping

## Key Implementation Details

### Service Input

Separate Handler DTOs from Service Inputs. Handler focuses on HTTP request/response, Service focuses on business logic.

```go
// handler/user/dto.go - for HTTP requests
type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

// service/user/input.go - for business logic
type CreateUserInput struct {
    Email    string
    Username string
    Password string
}
```

### Domain Errors

Domain errors include HTTP status code mapping. Enables consistent error handling in Handler.

```go
// domain/errors.go
type DomainError interface {
    error
    HTTPStatus() int
}

type UserNotFoundError struct {
    Id int
}

func (e UserNotFoundError) Error() string {
    return fmt.Sprintf("user not found with id: %d", e.Id)
}

func (e UserNotFoundError) HTTPStatus() int {
    return http.StatusNotFound
}
```

### Dependency Injection

Service depends on interfaces, not concrete implementations. Facilitates mocking in tests.

```go
// service/user/dependencies.go
type IUserRepository interface {
    GetUserById(id int) (*entity.User, error)
    InsertUser(user *entity.User) (int, error)
    // ...
}

type IPasswordHasher interface {
    Hash(password string) (string, error)
    Compare(hashedPassword, password string) error
}
```

### BaseHandler

Common error handling logic in BaseHandler. Domain-specific handlers embed and reuse.

```go
// handler/base.go
type BaseHandler struct{}

func (b *BaseHandler) HandleDomainError(c *gin.Context, err error) {
    if domainErr, ok := err.(domain.DomainError); ok {
        c.AbortWithStatusJSON(domainErr.HTTPStatus(), gin.H{"message": domainErr.Error()})
        return
    }
    c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
}

// handler/user/handler.go
type UserHandler struct {
    handler.BaseHandler  // embedding
    userService *user.Service
}
```

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 16+
- Docker & Docker Compose (optional)

### Build

```bash
# Build binary
make build

# Build Docker image
make docker-build
```

### Test

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/app/server/service/user/...
go test -v ./internal/app/server/handler/user/...
go test -v ./internal/pkg/auth/...
go test -v ./internal/pkg/repository/postgres/...

# Run tests with race detection
go test -race ./...
```

## Testing

### Service Layer Tests (`service/*_test.go`)
- Mock repository interface
- Mock password hasher
- Test business logic in isolation
- Verify domain error returns

```go
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetUserById(id int) (*entity.User, error) {
    args := m.Called(id)
    return args.Get(0).(*entity.User), args.Error(1)
}
```

### Handler Layer Tests (`handler/*_test.go`)
- Use `httptest` for HTTP testing
- Mock service layer
- Test request parsing and validation
- Test response formatting

```go
router := gin.New()
router.POST("/users", handler.CreateUser)

req := httptest.NewRequest(http.MethodPost, "/users", body)
w := httptest.NewRecorder()
router.ServeHTTP(w, req)

assert.Equal(t, http.StatusCreated, w.Code)
```

### Repository Layer Tests (`repository/*_test.go`)
- Unit tests for query building
- Integration tests with test database (skipped by default)
- Config validation tests

### Auth Package Tests (`auth/*_test.go`)
- JWT token generation and validation
- Password hashing and comparison
- Edge cases (expired tokens, wrong passwords)

## Adding a New Domain

1. Create entity in `internal/pkg/entity/`
2. Add domain errors in `internal/pkg/domain/errors.go`
3. Create repository methods in `internal/pkg/repository/postgres/`
4. Create service in `internal/app/server/service/<domain>/`
   - `dependencies.go` - Repository interface
   - `input.go` - Service input types
   - `service.go` - Business logic
5. Create handler in `internal/app/server/handler/<domain>/`
   - `dto.go` - Request/Response DTOs
   - `handler.go` - HTTP handlers
6. Add routes in `internal/app/server/routes/<domain>.go`
7. Wire up in `server.go`

## Scaling Considerations

Guide for when the project grows.

### When Domains Multiply

Consider fully separating by domain:

```
internal/
├── user/           # entire user domain
│   ├── handler/
│   ├── service/
│   ├── repository/
│   └── entity/
├── order/          # entire order domain
│   └── ...
```

### Service-to-Service Dependencies

To avoid circular dependencies when services need to call each other:

```go
// Separate with interfaces
type IUserGetter interface {
    GetUserById(id int) (*entity.User, error)
}

type OrderService struct {
    userGetter IUserGetter  // interface, not concrete service
}
```

### Transaction Management

Unit of Work pattern for wrapping multiple Repositories in a single transaction:

```go
type UnitOfWork interface {
    Begin() error
    Commit() error
    Rollback() error
    Users() IUserRepository
    Orders() IOrderRepository
}
```

### ORM / Query Builder

This template uses raw SQL. As queries get complex, consider:

| Tool | Characteristics |
|------|-----------------|
| [sqlx](https://github.com/jmoiron/sqlx) | `database/sql` extension. Struct mapping, Named Query |
| [sqlc](https://sqlc.dev/) | SQL → Go code generation. Type-safe, compile-time verification |
| [squirrel](https://github.com/Masterminds/squirrel) | SQL query builder. Fluent API, good for dynamic queries |
| [goqu](https://github.com/doug-martin/goqu) | Query builder. Multiple DB dialects, actively maintained |
| [GORM](https://gorm.io/) | Full ORM. Migrations, relations, hooks |
| [ent](https://entgo.io/) | By Facebook. Schema-based codegen, graph traversal |
| [Bun](https://bun.uptrace.dev/) | Lightweight ORM. Good PostgreSQL support |

Selection criteria:
- Simple CRUD → sqlx
- Dynamic query generation → squirrel, goqu
- Type safety priority → sqlc
- Complex relations/migrations → GORM, ent

### Other Considerations

| Situation | Pattern/Tool to Consider |
|-----------|--------------------------|
| Query performance issues | Caching layer (Redis) |
| Async processing needed | Event-driven architecture |
| API backward compatibility | API versioning (`/api/v1/`, `/api/v2/`) |
| Service scale explosion | Microservices decomposition |
| Environment-specific config | Per-environment config files |
| Distributed tracing needed | OpenTelemetry |
