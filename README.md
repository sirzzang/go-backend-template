# Go Backend Template

[English](./README.en.md) | **한국어**

> Go로 백엔드 서버 만들 때마다 "이번엔 구조를 좀 더 예쁘게 짜봐야지" 하고 시작해서,
> 결국 프로젝트 중반에 갈아엎기를 반복하다, **드디어 정착한 구조**.
> 
> 더 이상 리팩토링 지옥에 빠지지 않길 바라며, 그리고 미래의 내가 "그때 어떻게 했더라? 할 때 참고하려고 만든 레포.

레이어드 아키텍처 기반의 Go 백엔드 템플릿입니다.

## 목차

- [아키텍처](#아키텍처)
- [레이어별 책임](#레이어별-책임)
- [주요 구현 내용](#주요-구현-내용)
- [시작하기](#시작하기)
- [테스트](#테스트)
- [새 도메인 추가하기](#새-도메인-추가하기)
- [구조 확장 시 고려사항](#구조-확장-시-고려사항)

## 아키텍처

이 템플릿은 레이어드 아키텍처 패턴을 따릅니다:

```
Route → Middleware → Handler → Service → Repository
```

### 디렉토리 구조

```
go-backend-template/
├── cmd/
│   └── server/              # 애플리케이션 진입점
│       ├── main.go          # 메인 함수
│       └── config.go        # 설정 로딩
├── internal/
│   ├── app/
│   │   └── server/          # HTTP 서버 구현
│   │       ├── server.go    # 서버 설정 및 구성
│   │       ├── handler/     # HTTP 핸들러 (컨트롤러)
│   │       │   ├── base.go  # 공통 메서드를 가진 베이스 핸들러
│   │       │   ├── context.go # 컨텍스트 헬퍼
│   │       │   └── user/    # User 도메인 핸들러
│   │       │       ├── handler.go
│   │       │       └── dto.go
│   │       ├── middleware/  # HTTP 미들웨어
│   │       │   └── auth/
│   │       │       └── auth.go
│   │       ├── routes/      # 라우트 정의
│   │       │   ├── routes.go
│   │       │   └── user.go
│   │       └── service/     # 비즈니스 로직 레이어
│   │           └── user/
│   │               ├── service.go
│   │               ├── input.go
│   │               └── dependencies.go
│   └── pkg/                 # 공유 내부 패키지
│       ├── auth/            # 인증 유틸리티
│       │   ├── jwt.go
│       │   └── password.go
│       ├── domain/          # 도메인 에러
│       │   └── errors.go
│       ├── entity/          # 도메인 엔티티
│       │   └── user.go
│       └── repository/      # 데이터 접근 레이어
│           └── postgres/
│               ├── repository.go
│               ├── errors.go
│               └── user.go
├── build/
│   └── Dockerfile           # Docker 빌드 파일
├── deployments/
│   ├── docker-compose.yml   # 로컬 개발용 Docker Compose
│   └── .env.example         # 환경 변수 예시
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 레이어별 책임

### 1. Handler 레이어 (`handler/`)
- HTTP 요청 파싱 (경로 파라미터, 쿼리, 바디)
- 요청 형식 유효성 검증
- 요청 DTO를 서비스 입력으로 변환
- 서비스 메서드 호출
- 서비스 출력을 응답 DTO로 변환
- 에러 처리 및 HTTP 응답 전송

### 2. Service 레이어 (`service/`)
- 비즈니스 로직 구현
- 레포지토리 호출 조율
- 도메인 에러 반환
- HTTP 관련 코드 없음

### 3. Repository 레이어 (`repository/`)
- 데이터베이스 접근
- SQL 쿼리
- 데이터베이스 에러 반환

### 4. Domain 레이어 (`domain/`, `entity/`)
- 도메인 엔티티
- HTTP 상태 매핑이 포함된 도메인 에러

## 주요 구현 내용

### Service Input

Handler의 DTO와 Service의 Input을 분리. Handler는 HTTP 요청/응답에 집중하고, Service는 비즈니스 로직에 집중.

```go
// handler/user/dto.go - HTTP 요청용
type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

// service/user/input.go - 비즈니스 로직용
type CreateUserInput struct {
    Email    string
    Username string
    Password string
}
```

### 도메인 에러

도메인 에러에 HTTP 상태 코드 매핑을 포함. Handler에서 일관된 에러 처리 가능.

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

### 의존성 주입

Service는 구체적인 구현이 아닌 인터페이스에 의존. 테스트 시 모킹 용이.

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

공통 에러 처리 로직을 BaseHandler에 구현. 도메인별 Handler에서 임베딩하여 재사용.

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
    handler.BaseHandler  // 임베딩
    userService *user.Service
}
```

## 시작하기

### 사전 요구사항

- Go 1.21+
- PostgreSQL 16+
- Docker & Docker Compose (선택)

### 빌드

```bash
# 바이너리 빌드
make build

# Docker 이미지 빌드
make docker-build
```

### 테스트

```bash
# 모든 테스트 실행
make test

# 커버리지와 함께 테스트 실행
make test-coverage

# 특정 패키지 테스트 실행
go test -v ./internal/app/server/service/user/...
go test -v ./internal/app/server/handler/user/...
go test -v ./internal/pkg/auth/...
go test -v ./internal/pkg/repository/postgres/...

# 레이스 디텍션과 함께 테스트 실행
go test -race ./...
```

## 테스트

### Service 레이어 테스트 (`service/*_test.go`)
- 레포지토리 인터페이스 모킹
- 패스워드 해셔 모킹
- 비즈니스 로직 격리 테스트
- 도메인 에러 반환 검증

```go
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetUserById(id int) (*entity.User, error) {
    args := m.Called(id)
    return args.Get(0).(*entity.User), args.Error(1)
}
```

### Handler 레이어 테스트 (`handler/*_test.go`)
- `httptest`를 사용한 HTTP 테스트
- 서비스 레이어 모킹
- 요청 파싱 및 유효성 검증 테스트
- 응답 포맷팅 테스트

```go
router := gin.New()
router.POST("/users", handler.CreateUser)

req := httptest.NewRequest(http.MethodPost, "/users", body)
w := httptest.NewRecorder()
router.ServeHTTP(w, req)

assert.Equal(t, http.StatusCreated, w.Code)
```

### Repository 레이어 테스트 (`repository/*_test.go`)
- 쿼리 빌딩 단위 테스트
- 테스트 데이터베이스를 사용한 통합 테스트 (기본적으로 스킵)
- 설정 유효성 검증 테스트

### Auth 패키지 테스트 (`auth/*_test.go`)
- JWT 토큰 생성 및 검증
- 패스워드 해싱 및 비교
- 엣지 케이스 (만료된 토큰, 잘못된 패스워드)

## 새 도메인 추가하기

1. `internal/pkg/entity/`에 엔티티 생성
2. `internal/pkg/domain/errors.go`에 도메인 에러 추가
3. `internal/pkg/repository/postgres/`에 레포지토리 메서드 생성
4. `internal/app/server/service/<domain>/`에 서비스 생성
   - `dependencies.go` - 레포지토리 인터페이스
   - `input.go` - 서비스 입력 타입
   - `service.go` - 비즈니스 로직
5. `internal/app/server/handler/<domain>/`에 핸들러 생성
   - `dto.go` - 요청/응답 DTO
   - `handler.go` - HTTP 핸들러
6. `internal/app/server/routes/<domain>.go`에 라우트 추가
7. `server.go`에서 연결

## 구조 확장 시 고려사항

프로젝트가 커질 때 참고할 수 있는 가이드입니다.

### 도메인이 많아질 때

도메인별로 완전히 분리하는 구조 고려:

```
internal/
├── user/           # user 도메인 전체
│   ├── handler/
│   ├── service/
│   ├── repository/
│   └── entity/
├── order/          # order 도메인 전체
│   └── ...
```

### 서비스 간 의존성

서비스끼리 호출해야 할 때 순환 의존성을 피하려면:

```go
// 인터페이스로 분리
type IUserGetter interface {
    GetUserById(id int) (*entity.User, error)
}

type OrderService struct {
    userGetter IUserGetter  // 구체적인 서비스가 아닌 인터페이스
}
```

### 트랜잭션 관리

여러 Repository를 하나의 트랜잭션으로 묶어야 할 때 고려할 수 있는 Unit of Work 패턴:

```go
type UnitOfWork interface {
    Begin() error
    Commit() error
    Rollback() error
    Users() IUserRepository
    Orders() IOrderRepository
}
```

### ORM / Query Builder 도입

현재 템플릿은 raw SQL을 사용. 쿼리가 복잡해지면 다음 도구들을 고려:

| 도구 | 특징 |
|------|------|
| [sqlx](https://github.com/jmoiron/sqlx) | `database/sql` 확장. 구조체 매핑, Named Query 지원 |
| [sqlc](https://sqlc.dev/) | SQL → Go 코드 생성. 타입 안전, 컴파일 타임 검증 |
| [squirrel](https://github.com/Masterminds/squirrel) | SQL 쿼리 빌더. Fluent API, 동적 쿼리 생성에 강함 |
| [goqu](https://github.com/doug-martin/goqu) | 쿼리 빌더. 다양한 DB 방언 지원, 활발한 유지보수 |
| [GORM](https://gorm.io/) | 풀 ORM. 마이그레이션, 관계 매핑, 훅 지원 |
| [ent](https://entgo.io/) | Facebook 개발. 스키마 기반 코드 생성, 그래프 순회 |
| [Bun](https://bun.uptrace.dev/) | 경량 ORM. PostgreSQL 기능 잘 지원 |

선택 기준:
- 단순 CRUD 위주 → sqlx
- 동적 쿼리 생성 → squirrel, goqu
- 타입 안전성 중시 → sqlc
- 복잡한 관계/마이그레이션 필요 → GORM, ent

### 기타 고려사항

| 상황 | 고려할 패턴/도구 |
|------|------------------|
| 조회 성능 이슈 | 캐싱 레이어 (Redis) |
| 비동기 처리 필요 | 이벤트 기반 아키텍처 |
| API 하위 호환성 | API 버저닝 (`/api/v1/`, `/api/v2/`) |
| 서비스 규모 폭발 | 마이크로서비스 분리 검토 |
| 환경별 설정 관리 | 환경별 config 파일 |
| 분산 추적 필요 | OpenTelemetry |
