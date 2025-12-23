# Phase 2: 인증 시스템

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | JWT 기반 인증 시스템 구현 (Access + Refresh Token) |
| **선행 조건** | Phase 1, Phase 1.5 완료 |
| **예상 소요** | 7 Steps |
| **결과물** | 회원가입, 로그인, 토큰 갱신, 비밀번호 재설정, Google OAuth, 사용자 정보 API 동작 |

---

## 아키텍처

```mermaid
sequenceDiagram
    participant C as Client
    participant A as API Server
    participant DB as Database

    Note over C,DB: 회원가입/로그인 플로우
    C->>A: POST /auth/register or /auth/login
    A->>DB: 사용자 생성/조회
    DB-->>A: User
    A->>A: Access Token (15분) 생성
    A->>A: Refresh Token (7일) 생성
    A-->>C: { accessToken, refreshToken }

    Note over C,DB: 인증된 API 요청
    C->>A: GET /api/* (Authorization: Bearer {accessToken})
    A->>A: JWT 검증 (Middleware)
    A->>DB: 데이터 조회
    DB-->>A: Data
    A-->>C: Response

    Note over C,DB: 토큰 갱신
    C->>A: POST /auth/refresh (refreshToken)
    A->>A: Refresh Token 검증
    A->>A: 새 Access Token 생성
    A-->>C: { accessToken }
```

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 2.1 | JWT 서비스 구현 (Access + Refresh) | ⬜ |
| 2.2 | Auth 서비스 구현 | ⬜ |
| 2.3 | Auth 컨트롤러 구현 | ⬜ |
| 2.4 | Auth 미들웨어 구현 | ⬜ |
| 2.5 | 토큰 갱신 및 사용자 정보 API | ⬜ |
| 2.6 | 비밀번호 재설정 | ⬜ |
| 2.7 | Google OAuth | ⬜ |

---

## 개발 테스트용 계정/토큰

### 목적

개발 및 테스트 환경에서 일관된 인증 테스트를 위해 고정된 테스트 계정과 토큰을 제공합니다.

### 테스트 계정

| 항목 | 값 |
|-----|-----|
| **Email** | `test@mindhit.dev` |
| **Password** | `test1234!` |

> Note: User ID는 seed 스크립트 실행 시 데이터베이스에서 자동 생성됩니다.

### 테스트 토큰

개발 환경에서만 유효한 고정 토큰 (실제 JWT 서명 검증 통과):

| 토큰 유형 | 환경 변수 | 만료 |
|----------|----------|------|
| **Access Token** | `TEST_ACCESS_TOKEN` | 만료 없음 (개발 환경만) |
| **Refresh Token** | `TEST_REFRESH_TOKEN` | 만료 없음 (개발 환경만) |

### 구현 요구사항

- [ ] **Seed 데이터 생성**
  - [ ] `scripts/seed_test_user.go` - 테스트 사용자 생성 스크립트

  ```go
  const TestUserEmail = "test@mindhit.dev"

  func SeedTestUser(ctx context.Context, client *ent.Client) (*ent.User, error) {
      // 이미 존재하면 기존 사용자 반환
      existing, err := client.User.Query().
          Where(user.EmailEQ(TestUserEmail)).
          Only(ctx)
      if err == nil {
          return existing, nil
      }

      hashedPassword, _ := bcrypt.GenerateFromPassword(
          []byte("test1234!"),
          bcrypt.DefaultCost,
      )

      return client.User.Create().
          SetEmail(TestUserEmail).
          SetPasswordHash(string(hashedPassword)).
          Save(ctx)
  }
  ```

- [ ] **환경별 토큰 검증**
  - [ ] `pkg/service/jwt_service.go`에 테스트 토큰 지원 추가

  ```go
  func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
      // 개발 환경에서 테스트 토큰 허용
      if s.isDev && tokenString == os.Getenv("TEST_ACCESS_TOKEN") {
          // 테스트 사용자 조회
          testUser, err := s.userRepo.GetByEmail(ctx, "test@mindhit.dev")
          if err != nil {
              return nil, err
          }
          return &Claims{
              UserID:    testUser.ID,
              TokenType: AccessToken,
          }, nil
      }

      // 일반 JWT 검증
      // ...
  }
  ```

- [ ] **환경 변수 설정** (`.env.development`)
  ```bash
  # 테스트 인증
  TEST_ACCESS_TOKEN=mindhit-test-access-token-dev-only
  TEST_REFRESH_TOKEN=mindhit-test-refresh-token-dev-only
  ```

- [ ] **moonrepo task 추가** (`moon.yml`)
  ```yaml
  tasks:
    seed-test-user:
      command: go run ./scripts/seed_test_user.go
      deps:
        - migrate
  ```

### 사용 예시

```bash
# 테스트 토큰으로 API 호출
curl -X GET http://localhost:8080/v1/auth/me \
  -H "Authorization: Bearer mindhit-test-access-token-dev-only"

# 또는 환경 변수 사용
curl -X GET http://localhost:8080/v1/auth/me \
  -H "Authorization: Bearer $TEST_ACCESS_TOKEN"
```

### 보안 주의사항

> **WARNING**: 테스트 토큰은 반드시 개발/테스트 환경에서만 활성화되어야 합니다.
> 프로덕션 환경에서는 `APP_ENV=production`으로 설정하여 테스트 토큰이 무효화됩니다.

```go
// 프로덕션에서는 테스트 토큰 비활성화
if os.Getenv("APP_ENV") == "production" {
    s.isDev = false
}
```

---

## Step 2.1: JWT 서비스 구현 (Access + Refresh Token)

### 목표

Access Token (15분) + Refresh Token (7일) 기반 JWT 인증

### 토큰 전략

| 토큰 | 만료 시간 | 용도 | 저장 위치 |
|-----|----------|------|----------|
| Access Token | 15분 | API 인증 | 메모리 (클라이언트) |
| Refresh Token | 7일 | Access Token 갱신 | HttpOnly Cookie / Storage |

### 체크리스트

- [ ] **의존성 추가**

  ```bash
  cd apps/backend
  go get github.com/golang-jwt/jwt/v5
  ```

- [ ] **JWT 서비스 작성**
  - [ ] `pkg/service/jwt_service.go`

    ```go
    package service

    import (
        "fmt"
        "time"

        "github.com/golang-jwt/jwt/v5"
        "github.com/google/uuid"
    )

    type TokenType string

    const (
        AccessToken  TokenType = "access"
        RefreshToken TokenType = "refresh"
    )

    type JWTService struct {
        secret            []byte
        accessExpiration  time.Duration
        refreshExpiration time.Duration
    }

    type Claims struct {
        UserID    uuid.UUID `json:"user_id"`
        TokenType TokenType `json:"token_type"`
        jwt.RegisteredClaims
    }

    type TokenPair struct {
        AccessToken  string `json:"access_token"`
        RefreshToken string `json:"refresh_token"`
        ExpiresIn    int64  `json:"expires_in"` // Access token expiry in seconds
    }

    func NewJWTService(secret string) *JWTService {
        return &JWTService{
            secret:            []byte(secret),
            accessExpiration:  15 * time.Minute,
            refreshExpiration: 7 * 24 * time.Hour,
        }
    }

    // GenerateTokenPair creates both access and refresh tokens
    func (s *JWTService) GenerateTokenPair(userID uuid.UUID) (*TokenPair, error) {
        accessToken, err := s.generateToken(userID, AccessToken, s.accessExpiration)
        if err != nil {
            return nil, fmt.Errorf("generate access token: %w", err)
        }

        refreshToken, err := s.generateToken(userID, RefreshToken, s.refreshExpiration)
        if err != nil {
            return nil, fmt.Errorf("generate refresh token: %w", err)
        }

        return &TokenPair{
            AccessToken:  accessToken,
            RefreshToken: refreshToken,
            ExpiresIn:    int64(s.accessExpiration.Seconds()),
        }, nil
    }

    // GenerateAccessToken creates only access token (for refresh)
    func (s *JWTService) GenerateAccessToken(userID uuid.UUID) (string, int64, error) {
        token, err := s.generateToken(userID, AccessToken, s.accessExpiration)
        if err != nil {
            return "", 0, err
        }
        return token, int64(s.accessExpiration.Seconds()), nil
    }

    func (s *JWTService) generateToken(userID uuid.UUID, tokenType TokenType, expiration time.Duration) (string, error) {
        claims := Claims{
            UserID:    userID,
            TokenType: tokenType,
            RegisteredClaims: jwt.RegisteredClaims{
                ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
                IssuedAt:  jwt.NewNumericDate(time.Now()),
                Issuer:    "mindhit",
                Subject:   userID.String(),
            },
        }

        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
        return token.SignedString(s.secret)
    }

    // ValidateToken validates any token type
    func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
        token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return s.secret, nil
        })
        if err != nil {
            return nil, err
        }

        if claims, ok := token.Claims.(*Claims); ok && token.Valid {
            return claims, nil
        }

        return nil, jwt.ErrSignatureInvalid
    }

    // ValidateRefreshToken validates specifically refresh token
    func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
        claims, err := s.ValidateToken(tokenString)
        if err != nil {
            return nil, err
        }

        if claims.TokenType != RefreshToken {
            return nil, fmt.Errorf("invalid token type: expected refresh token")
        }

        return claims, nil
    }

    // ValidateAccessToken validates specifically access token
    func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
        claims, err := s.ValidateToken(tokenString)
        if err != nil {
            return nil, err
        }

        if claims.TokenType != AccessToken {
            return nil, fmt.Errorf("invalid token type: expected access token")
        }

        return claims, nil
    }
    ```

- [ ] **테스트 작성**
  - [ ] `pkg/service/jwt_service_test.go`

    ```go
    package service_test

    import (
        "testing"

        "github.com/google/uuid"
        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"

        "github.com/mindhit/api/pkg/service"
    )

    func TestJWTService_GenerateAndValidate(t *testing.T) {
        jwtService := service.NewJWTService("test-secret")
        userID := uuid.New()

        token, err := jwtService.GenerateToken(userID)
        require.NoError(t, err)
        assert.NotEmpty(t, token)

        claims, err := jwtService.ValidateToken(token)
        require.NoError(t, err)
        assert.Equal(t, userID, claims.UserID)
    }

    func TestJWTService_InvalidToken(t *testing.T) {
        jwtService := service.NewJWTService("test-secret")

        _, err := jwtService.ValidateToken("invalid-token")
        assert.Error(t, err)
    }
    ```

### 검증

```bash
cd apps/backend
go test ./pkg/service/... -v -run TestJWT
```

---

## Step 2.2: Auth 서비스 구현

### 목표

회원가입, 로그인 비즈니스 로직

### 체크리스트

- [ ] **bcrypt 의존성 추가**

  ```bash
  go get golang.org/x/crypto/bcrypt
  ```

- [ ] **Auth 서비스 작성**
  - [ ] `pkg/service/auth_service.go`

    ```go
    package service

    import (
        "context"
        "errors"

        "github.com/google/uuid"
        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/user"
        "golang.org/x/crypto/bcrypt"
    )

    var (
        ErrUserNotFound       = errors.New("user not found")
        ErrInvalidCredentials = errors.New("invalid credentials")
        ErrEmailExists        = errors.New("email already exists")
        ErrUserInactive       = errors.New("user account is inactive")
    )

    type AuthService struct {
        client *ent.Client
    }

    func NewAuthService(client *ent.Client) *AuthService {
        return &AuthService{client: client}
    }

    // activeUsers returns a query filtered to active users only
    func (s *AuthService) activeUsers() *ent.UserQuery {
        return s.client.User.Query().Where(user.StatusEQ("active"))
    }

    func (s *AuthService) Signup(ctx context.Context, email, password string) (*ent.User, error) {
        // 이메일 중복 체크 (활성 사용자만)
        exists, err := s.activeUsers().
            Where(user.EmailEQ(email)).
            Exist(ctx)
        if err != nil {
            return nil, err
        }
        if exists {
            return nil, ErrEmailExists
        }

        // 비밀번호 해싱
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if err != nil {
            return nil, err
        }

        // 사용자 생성 (status는 SoftDeleteMixin 기본값 "active")
        return s.client.User.
            Create().
            SetEmail(email).
            SetPasswordHash(string(hashedPassword)).
            Save(ctx)
    }

    func (s *AuthService) Login(ctx context.Context, email, password string) (*ent.User, error) {
        // 활성 사용자만 조회
        u, err := s.activeUsers().
            Where(user.EmailEQ(email)).
            Only(ctx)
        if err != nil {
            if ent.IsNotFound(err) {
                return nil, ErrInvalidCredentials
            }
            return nil, err
        }

        // 비밀번호 검증
        if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
            return nil, ErrInvalidCredentials
        }

        return u, nil
    }

    func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*ent.User, error) {
        // 활성 사용자만 조회
        u, err := s.activeUsers().
            Where(user.IDEQ(id)).
            Only(ctx)
        if err != nil {
            if ent.IsNotFound(err) {
                return nil, ErrUserNotFound
            }
            return nil, err
        }
        return u, nil
    }
    ```

- [ ] **import 추가**

  ```go
  import "github.com/google/uuid"
  ```

- [ ] **테스트 작성**
  - [ ] `pkg/service/auth_service_test.go`

    ```go
    package service_test

    import (
        "context"
        "testing"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
        "github.com/stretchr/testify/suite"

        "github.com/mindhit/api/ent/enttest"
        "github.com/mindhit/api/pkg/service"

        _ "github.com/mattn/go-sqlite3"
    )

    type AuthServiceTestSuite struct {
        suite.Suite
        client      *ent.Client
        authService *service.AuthService
    }

    func (s *AuthServiceTestSuite) SetupTest() {
        s.client = enttest.Open(s.T(), "sqlite3", "file:ent?mode=memory&_fk=1")
        s.authService = service.NewAuthService(s.client)
    }

    func (s *AuthServiceTestSuite) TearDownTest() {
        s.client.Close()
    }

    func (s *AuthServiceTestSuite) TestSignup_Success() {
        ctx := context.Background()

        user, err := s.authService.Signup(ctx, "test@example.com", "password123")

        require.NoError(s.T(), err)
        assert.NotNil(s.T(), user)
        assert.Equal(s.T(), "test@example.com", user.Email)
    }

    func (s *AuthServiceTestSuite) TestSignup_DuplicateEmail() {
        ctx := context.Background()

        _, err := s.authService.Signup(ctx, "test@example.com", "password123")
        require.NoError(s.T(), err)

        _, err = s.authService.Signup(ctx, "test@example.com", "password456")
        assert.ErrorIs(s.T(), err, service.ErrEmailExists)
    }

    func (s *AuthServiceTestSuite) TestLogin_Success() {
        ctx := context.Background()

        _, err := s.authService.Signup(ctx, "test@example.com", "password123")
        require.NoError(s.T(), err)

        user, err := s.authService.Login(ctx, "test@example.com", "password123")

        require.NoError(s.T(), err)
        assert.Equal(s.T(), "test@example.com", user.Email)
    }

    func (s *AuthServiceTestSuite) TestLogin_InvalidCredentials() {
        ctx := context.Background()

        _, err := s.authService.Signup(ctx, "test@example.com", "password123")
        require.NoError(s.T(), err)

        _, err = s.authService.Login(ctx, "test@example.com", "wrongpassword")
        assert.ErrorIs(s.T(), err, service.ErrInvalidCredentials)
    }

    func TestAuthServiceTestSuite(t *testing.T) {
        suite.Run(t, new(AuthServiceTestSuite))
    }
    ```

- [ ] **SQLite 드라이버 추가** (테스트용)

  ```bash
  go get github.com/mattn/go-sqlite3
  go get github.com/stretchr/testify
  ```

### 검증

```bash
cd apps/backend
go test ./pkg/service/... -v -run TestAuthService
```

---

## Step 2.3: Auth 컨트롤러 구현

### 목표

HTTP 핸들러로 API 엔드포인트 구현

### 체크리스트

- [ ] **Auth 컨트롤러 작성**
  - [ ] `internal/api/controller/auth_controller.go`

    ```go
    package controller

    import (
        "errors"
        "net/http"

        "github.com/gin-gonic/gin"

        "github.com/mindhit/api/internal/generated"
        "github.com/mindhit/api/pkg/service"
    )

    type AuthController struct {
        authService *service.AuthService
        jwtService  *service.JWTService
    }

    func NewAuthController(authService *service.AuthService, jwtService *service.JWTService) *AuthController {
        return &AuthController{
            authService: authService,
            jwtService:  jwtService,
        }
    }

    func (c *AuthController) Signup(ctx *gin.Context) {
        var req generated.SignupRequest
        if err := ctx.ShouldBindJSON(&req); err != nil {
            ctx.JSON(http.StatusBadRequest, generated.ValidationError{
                Error: struct {
                    Message string                              `json:"message"`
                    Details *[]generated.ValidationDetail       `json:"details,omitempty"`
                }{
                    Message: err.Error(),
                },
            })
            return
        }

        user, err := c.authService.Signup(ctx.Request.Context(), req.Email, req.Password)
        if err != nil {
            if errors.Is(err, service.ErrEmailExists) {
                ctx.JSON(http.StatusConflict, generated.ErrorResponse{
                    Error: struct {
                        Message string  `json:"message"`
                        Code    *string `json:"code,omitempty"`
                    }{
                        Message: "email already exists",
                    },
                })
                return
            }
            ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{
                Error: struct {
                    Message string  `json:"message"`
                    Code    *string `json:"code,omitempty"`
                }{
                    Message: "internal server error",
                },
            })
            return
        }

        token, err := c.jwtService.GenerateToken(user.ID)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{
                Error: struct {
                    Message string  `json:"message"`
                    Code    *string `json:"code,omitempty"`
                }{
                    Message: "failed to generate token",
                },
            })
            return
        }

        ctx.JSON(http.StatusCreated, generated.AuthResponse{
            User: generated.User{
                Id:        user.ID.String(),
                Email:     user.Email,
                CreatedAt: user.CreatedAt,
                UpdatedAt: user.UpdatedAt,
            },
            Token: token,
        })
    }

    func (c *AuthController) Login(ctx *gin.Context) {
        var req generated.LoginRequest
        if err := ctx.ShouldBindJSON(&req); err != nil {
            ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{
                Error: struct {
                    Message string  `json:"message"`
                    Code    *string `json:"code,omitempty"`
                }{
                    Message: err.Error(),
                },
            })
            return
        }

        user, err := c.authService.Login(ctx.Request.Context(), req.Email, req.Password)
        if err != nil {
            if errors.Is(err, service.ErrInvalidCredentials) {
                ctx.JSON(http.StatusUnauthorized, generated.ErrorResponse{
                    Error: struct {
                        Message string  `json:"message"`
                        Code    *string `json:"code,omitempty"`
                    }{
                        Message: "invalid credentials",
                    },
                })
                return
            }
            ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{
                Error: struct {
                    Message string  `json:"message"`
                    Code    *string `json:"code,omitempty"`
                }{
                    Message: "internal server error",
                },
            })
            return
        }

        token, err := c.jwtService.GenerateToken(user.ID)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{
                Error: struct {
                    Message string  `json:"message"`
                    Code    *string `json:"code,omitempty"`
                }{
                    Message: "failed to generate token",
                },
            })
            return
        }

        ctx.JSON(http.StatusOK, generated.AuthResponse{
            User: generated.User{
                Id:        user.ID.String(),
                Email:     user.Email,
                CreatedAt: user.CreatedAt,
                UpdatedAt: user.UpdatedAt,
            },
            Token: token,
        })
    }
    ```

- [ ] **main.go 업데이트**
  - [ ] `cmd/server/main.go`에 라우트 등록

    ```go
    // Ent Client 초기화
    client, err := ent.Open("postgres", cfg.DatabaseURL)
    if err != nil {
        slog.Error("failed to connect to database", "error", err)
        os.Exit(1)
    }
    defer client.Close()

    // Services
    jwtService := service.NewJWTService(cfg.JWTSecret)
    authService := service.NewAuthService(client)

    // Controllers
    authController := controller.NewAuthController(authService, jwtService)

    // Routes
    v1 := r.Group("/v1")
    {
        auth := v1.Group("/auth")
        {
            auth.POST("/signup", authController.Signup)
            auth.POST("/login", authController.Login)
        }
    }
    ```

### 검증

```bash
# 서버 실행
cd apps/backend && go run ./cmd/server

# 회원가입 테스트
curl -X POST http://localhost:8080/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# 로그인 테스트
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

---

## Step 2.4: Auth 미들웨어 구현

### 목표

JWT 토큰 검증 미들웨어

### 체크리스트

- [ ] **Auth 미들웨어 작성**
  - [ ] `pkg/infra/middleware/auth.go`

    ```go
    package middleware

    import (
        "net/http"
        "strings"

        "github.com/gin-gonic/gin"

        "github.com/mindhit/api/pkg/service"
    )

    const (
        UserIDKey = "userID"
    )

    func Auth(jwtService *service.JWTService) gin.HandlerFunc {
        return func(c *gin.Context) {
            authHeader := c.GetHeader("Authorization")
            if authHeader == "" {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                    "error": gin.H{"message": "missing authorization header"},
                })
                return
            }

            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                    "error": gin.H{"message": "invalid authorization header format"},
                })
                return
            }

            claims, err := jwtService.ValidateToken(parts[1])
            if err != nil {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                    "error": gin.H{"message": "invalid or expired token"},
                })
                return
            }

            c.Set(UserIDKey, claims.UserID)
            c.Next()
        }
    }

    // GetUserID extracts user ID from context
    func GetUserID(c *gin.Context) (uuid.UUID, bool) {
        userID, exists := c.Get(UserIDKey)
        if !exists {
            return uuid.UUID{}, false
        }
        id, ok := userID.(uuid.UUID)
        return id, ok
    }
    ```

- [ ] **import 추가**

  ```go
  import "github.com/google/uuid"
  ```

- [ ] **CORS 미들웨어 업데이트**
  - [ ] `pkg/infra/middleware/cors.go`

    ```go
    package middleware

    import (
        "github.com/gin-contrib/cors"
        "github.com/gin-gonic/gin"
    )

    func CORS() gin.HandlerFunc {
        return cors.New(cors.Config{
            AllowOrigins:     []string{"http://localhost:3000", "chrome-extension://*"},
            AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
            AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
            ExposeHeaders:    []string{"Content-Length"},
            AllowCredentials: true,
        })
    }
    ```

- [ ] **로깅 미들웨어**
  - [ ] `pkg/infra/middleware/logging.go`

    ```go
    package middleware

    import (
        "log/slog"
        "time"

        "github.com/gin-gonic/gin"
    )

    func Logging() gin.HandlerFunc {
        return func(c *gin.Context) {
            start := time.Now()
            path := c.Request.URL.Path
            query := c.Request.URL.RawQuery

            c.Next()

            slog.Info("request",
                "method", c.Request.Method,
                "path", path,
                "query", query,
                "status", c.Writer.Status(),
                "latency", time.Since(start),
                "ip", c.ClientIP(),
                "user-agent", c.Request.UserAgent(),
            )
        }
    }
    ```

- [ ] **main.go에 미들웨어 적용**

  ```go
  r.Use(gin.Recovery())
  r.Use(middleware.Logging())
  r.Use(middleware.CORS())

  // Protected routes example
  protected := v1.Group("/")
  protected.Use(middleware.Auth(jwtService))
  {
      // 인증이 필요한 라우트들
  }
  ```

### 검증

```bash
# 토큰 없이 요청 (401 예상)
curl -X GET http://localhost:8080/v1/sessions \
  -H "Content-Type: application/json"

# 토큰과 함께 요청 (로그인 후)
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' | jq -r '.token')

curl -X GET http://localhost:8080/v1/sessions \
  -H "Authorization: Bearer $TOKEN"
```

---

---

## Step 2.5: 토큰 갱신 및 사용자 정보 API

### 목표

- `POST /v1/auth/refresh` - Refresh Token으로 Access Token 갱신
- `GET /v1/auth/me` - 현재 사용자 정보 조회

### 체크리스트

- [ ] **Auth 컨트롤러에 추가**
  - [ ] `internal/api/controller/auth_controller.go`에 메서드 추가

    ```go
    // RefreshRequest for token refresh
    type RefreshRequest struct {
        RefreshToken string `json:"refresh_token" binding:"required"`
    }

    // Refresh generates new access token using refresh token
    func (c *AuthController) Refresh(ctx *gin.Context) {
        var req RefreshRequest
        if err := ctx.ShouldBindJSON(&req); err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{"message": err.Error()},
            })
            return
        }

        // Validate refresh token
        claims, err := c.jwtService.ValidateRefreshToken(req.RefreshToken)
        if err != nil {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{"message": "invalid or expired refresh token"},
            })
            return
        }

        // Verify user still exists
        user, err := c.authService.GetUserByID(ctx.Request.Context(), claims.UserID)
        if err != nil {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{"message": "user not found"},
            })
            return
        }

        // Generate new access token
        accessToken, expiresIn, err := c.jwtService.GenerateAccessToken(user.ID)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{"message": "failed to generate token"},
            })
            return
        }

        ctx.JSON(http.StatusOK, gin.H{
            "access_token": accessToken,
            "token_type":   "Bearer",
            "expires_in":   expiresIn,
        })
    }

    // Me returns current authenticated user information
    func (c *AuthController) Me(ctx *gin.Context) {
        userID, exists := middleware.GetUserID(ctx)
        if !exists {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{"message": "unauthorized"},
            })
            return
        }

        user, err := c.authService.GetUserByID(ctx.Request.Context(), userID)
        if err != nil {
            if errors.Is(err, service.ErrUserNotFound) {
                ctx.JSON(http.StatusNotFound, gin.H{
                    "error": gin.H{"message": "user not found"},
                })
                return
            }
            ctx.JSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{"message": "internal server error"},
            })
            return
        }

        ctx.JSON(http.StatusOK, gin.H{
            "user": gin.H{
                "id":         user.ID.String(),
                "email":      user.Email,
                "created_at": user.CreatedAt,
                "updated_at": user.UpdatedAt,
            },
        })
    }

    // Logout invalidates the current session (client-side token removal)
    // Note: For stateless JWT, logout is handled client-side by removing tokens
    // This endpoint can be used for audit logging or future token blacklisting
    func (c *AuthController) Logout(ctx *gin.Context) {
        userID, exists := middleware.GetUserID(ctx)
        if !exists {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{"message": "unauthorized"},
            })
            return
        }

        // Log the logout event (optional: implement token blacklisting here)
        slog.Info("user logged out", "user_id", userID.String())

        ctx.JSON(http.StatusOK, gin.H{
            "message": "successfully logged out",
        })
    }
    ```

- [ ] **Signup/Login 응답 수정** (TokenPair 반환)

    ```go
    func (c *AuthController) Signup(ctx *gin.Context) {
        // ... validation code ...

        user, err := c.authService.Signup(ctx.Request.Context(), req.Email, req.Password)
        if err != nil {
            // ... error handling ...
        }

        // Generate token pair
        tokenPair, err := c.jwtService.GenerateTokenPair(user.ID)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{"message": "failed to generate tokens"},
            })
            return
        }

        ctx.JSON(http.StatusCreated, gin.H{
            "user": gin.H{
                "id":         user.ID.String(),
                "email":      user.Email,
                "created_at": user.CreatedAt,
                "updated_at": user.UpdatedAt,
            },
            "access_token":  tokenPair.AccessToken,
            "refresh_token": tokenPair.RefreshToken,
            "token_type":    "Bearer",
            "expires_in":    tokenPair.ExpiresIn,
        })
    }
    ```

- [ ] **라우트 등록**

  ```go
  // In main.go
  auth := v1.Group("/auth")
  {
      auth.POST("/signup", authController.Signup)
      auth.POST("/login", authController.Login)
      auth.POST("/refresh", authController.Refresh)
  }

  // Protected auth routes
  authProtected := v1.Group("/auth")
  authProtected.Use(middleware.Auth(jwtService))
  {
      authProtected.GET("/me", authController.Me)
      authProtected.POST("/logout", authController.Logout)
  }
  ```

- [ ] **Auth 미들웨어 수정** (Access Token만 허용)

  ```go
  func Auth(jwtService *service.JWTService) gin.HandlerFunc {
      return func(c *gin.Context) {
          authHeader := c.GetHeader("Authorization")
          if authHeader == "" {
              c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                  "error": gin.H{"message": "missing authorization header"},
              })
              return
          }

          parts := strings.Split(authHeader, " ")
          if len(parts) != 2 || parts[0] != "Bearer" {
              c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                  "error": gin.H{"message": "invalid authorization header format"},
              })
              return
          }

          // Only accept access tokens for API authentication
          claims, err := jwtService.ValidateAccessToken(parts[1])
          if err != nil {
              c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                  "error": gin.H{"message": "invalid or expired token"},
              })
              return
          }

          c.Set(UserIDKey, claims.UserID)
          c.Next()
      }
  }
  ```

### 검증

```bash
# 1. 회원가입 (access + refresh token 반환)
curl -X POST http://localhost:8080/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# 2. 토큰 갱신
curl -X POST http://localhost:8080/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token_here>"}'

# 3. 사용자 정보 조회
curl -X GET http://localhost:8080/v1/auth/me \
  -H "Authorization: Bearer <access_token_here>"
```

---

## Step 2.6: 비밀번호 재설정

### 목표

- `POST /v1/auth/forgot-password` - 비밀번호 재설정 이메일 요청
- `POST /v1/auth/reset-password` - 비밀번호 재설정 완료

### 아키텍처

```mermaid
sequenceDiagram
    participant C as Client
    participant A as API Server
    participant DB as Database
    participant E as Email Service

    Note over C,E: 비밀번호 재설정 요청
    C->>A: POST /auth/forgot-password {email}
    A->>DB: 이메일로 사용자 조회
    DB-->>A: User
    A->>A: Reset Token 생성 (1시간 만료)
    A->>DB: Reset Token 저장
    A->>E: 재설정 링크 이메일 발송
    A-->>C: { message: "이메일 발송됨" }

    Note over C,E: 비밀번호 재설정 완료
    C->>A: POST /auth/reset-password {token, password}
    A->>DB: Token 검증 및 사용자 조회
    DB-->>A: User
    A->>A: 비밀번호 해싱
    A->>DB: 비밀번호 업데이트, Token 삭제
    A-->>C: { message: "비밀번호 변경됨" }
```

### 체크리스트

- [ ] **Ent 스키마 추가**
  - [ ] `pkg/ent/schema/password_reset_token.go`

    ```go
    package schema

    import (
        "time"

        "entgo.io/ent"
        "entgo.io/ent/schema/edge"
        "entgo.io/ent/schema/field"
        "entgo.io/ent/schema/index"
        "github.com/google/uuid"
    )

    // PasswordResetToken holds the schema definition for password reset tokens.
    type PasswordResetToken struct {
        ent.Schema
    }

    func (PasswordResetToken) Fields() []ent.Field {
        return []ent.Field{
            field.UUID("id", uuid.UUID{}).
                Default(uuid.New).
                Immutable(),
            field.String("token").
                Unique().
                NotEmpty(),
            field.UUID("user_id", uuid.UUID{}),
            field.Time("expires_at"),
            field.Time("created_at").
                Default(time.Now).
                Immutable(),
            field.Bool("used").
                Default(false),
        }
    }

    func (PasswordResetToken) Edges() []ent.Edge {
        return []ent.Edge{
            edge.From("user", User.Type).
                Ref("password_reset_tokens").
                Field("user_id").
                Unique().
                Required(),
        }
    }

    func (PasswordResetToken) Indexes() []ent.Index {
        return []ent.Index{
            index.Fields("token"),
            index.Fields("user_id"),
            index.Fields("expires_at"),
        }
    }
    ```

  - [ ] `pkg/ent/schema/user.go`에 edge 추가

    ```go
    func (User) Edges() []ent.Edge {
        return []ent.Edge{
            // ... 기존 edges ...
            edge.To("password_reset_tokens", PasswordResetToken.Type),
        }
    }
    ```

- [ ] **Auth 서비스에 메서드 추가**
  - [ ] `pkg/service/auth_service.go`

    ```go
    import (
        "crypto/rand"
        "encoding/hex"
    )

    var (
        ErrTokenExpired = errors.New("token expired")
        ErrTokenUsed    = errors.New("token already used")
        ErrTokenInvalid = errors.New("invalid token")
    )

    // generateSecureToken creates a cryptographically secure random token
    func generateSecureToken() (string, error) {
        bytes := make([]byte, 32)
        if _, err := rand.Read(bytes); err != nil {
            return "", err
        }
        return hex.EncodeToString(bytes), nil
    }

    // RequestPasswordReset creates a password reset token and returns it
    // The caller is responsible for sending the email
    func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) (string, error) {
        // 사용자 조회 (활성 사용자만)
        u, err := s.activeUsers().
            Where(user.EmailEQ(email)).
            Only(ctx)
        if err != nil {
            if ent.IsNotFound(err) {
                // 보안: 존재하지 않는 이메일도 성공으로 응답 (enumeration 방지)
                return "", nil
            }
            return "", err
        }

        // 기존 미사용 토큰 무효화
        _, err = s.client.PasswordResetToken.
            Update().
            Where(
                passwordresettoken.UserIDEQ(u.ID),
                passwordresettoken.UsedEQ(false),
            ).
            SetUsed(true).
            Save(ctx)
        if err != nil {
            return "", err
        }

        // 새 토큰 생성
        token, err := generateSecureToken()
        if err != nil {
            return "", err
        }

        // 토큰 저장 (1시간 만료)
        _, err = s.client.PasswordResetToken.
            Create().
            SetToken(token).
            SetUserID(u.ID).
            SetExpiresAt(time.Now().Add(1 * time.Hour)).
            Save(ctx)
        if err != nil {
            return "", err
        }

        return token, nil
    }

    // ResetPassword validates the token and updates the password
    func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
        // 토큰 조회
        resetToken, err := s.client.PasswordResetToken.
            Query().
            Where(
                passwordresettoken.TokenEQ(token),
                passwordresettoken.UsedEQ(false),
            ).
            WithUser().
            Only(ctx)
        if err != nil {
            if ent.IsNotFound(err) {
                return ErrTokenInvalid
            }
            return err
        }

        // 만료 확인
        if time.Now().After(resetToken.ExpiresAt) {
            return ErrTokenExpired
        }

        // 사용자 활성 상태 확인
        if resetToken.Edges.User.Status != "active" {
            return ErrUserInactive
        }

        // 비밀번호 해싱
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
        if err != nil {
            return err
        }

        // 트랜잭션으로 비밀번호 업데이트 및 토큰 사용 처리
        tx, err := s.client.Tx(ctx)
        if err != nil {
            return err
        }

        // 비밀번호 업데이트
        _, err = tx.User.
            UpdateOneID(resetToken.Edges.User.ID).
            SetPasswordHash(string(hashedPassword)).
            Save(ctx)
        if err != nil {
            tx.Rollback()
            return err
        }

        // 토큰 사용 처리
        _, err = tx.PasswordResetToken.
            UpdateOneID(resetToken.ID).
            SetUsed(true).
            Save(ctx)
        if err != nil {
            tx.Rollback()
            return err
        }

        return tx.Commit()
    }
    ```

- [ ] **Auth 컨트롤러에 메서드 추가**
  - [ ] `internal/api/controller/auth_controller.go`

    ```go
    // ForgotPasswordRequest for password reset request
    type ForgotPasswordRequest struct {
        Email string `json:"email" binding:"required,email"`
    }

    // ResetPasswordRequest for password reset
    type ResetPasswordRequest struct {
        Token       string `json:"token" binding:"required"`
        NewPassword string `json:"new_password" binding:"required,min=8"`
    }

    // ForgotPassword initiates password reset flow
    func (c *AuthController) ForgotPassword(ctx *gin.Context) {
        var req ForgotPasswordRequest
        if err := ctx.ShouldBindJSON(&req); err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{"message": err.Error()},
            })
            return
        }

        token, err := c.authService.RequestPasswordReset(ctx.Request.Context(), req.Email)
        if err != nil {
            // 내부 오류만 로깅, 클라이언트에는 성공 응답
            slog.Error("failed to create reset token", "error", err)
        }

        // 토큰이 있으면 이메일 발송 (이메일 서비스 구현 필요)
        if token != "" {
            // TODO: 이메일 발송 (Phase 이후 구현)
            // emailService.SendPasswordReset(req.Email, token)
            slog.Info("password reset requested", "email", req.Email)
        }

        // 보안: 항상 같은 응답 (이메일 존재 여부 노출 방지)
        ctx.JSON(http.StatusOK, gin.H{
            "message": "If the email exists, a password reset link has been sent.",
        })
    }

    // ResetPassword completes password reset
    func (c *AuthController) ResetPassword(ctx *gin.Context) {
        var req ResetPasswordRequest
        if err := ctx.ShouldBindJSON(&req); err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{"message": err.Error()},
            })
            return
        }

        err := c.authService.ResetPassword(ctx.Request.Context(), req.Token, req.NewPassword)
        if err != nil {
            switch {
            case errors.Is(err, service.ErrTokenInvalid):
                ctx.JSON(http.StatusBadRequest, gin.H{
                    "error": gin.H{"message": "invalid or expired token"},
                })
            case errors.Is(err, service.ErrTokenExpired):
                ctx.JSON(http.StatusBadRequest, gin.H{
                    "error": gin.H{"message": "token has expired"},
                })
            case errors.Is(err, service.ErrTokenUsed):
                ctx.JSON(http.StatusBadRequest, gin.H{
                    "error": gin.H{"message": "token has already been used"},
                })
            case errors.Is(err, service.ErrUserInactive):
                ctx.JSON(http.StatusBadRequest, gin.H{
                    "error": gin.H{"message": "user account is inactive"},
                })
            default:
                ctx.JSON(http.StatusInternalServerError, gin.H{
                    "error": gin.H{"message": "internal server error"},
                })
            }
            return
        }

        ctx.JSON(http.StatusOK, gin.H{
            "message": "Password has been reset successfully.",
        })
    }
    ```

- [ ] **라우트 등록**

  ```go
  // In main.go - Public auth routes
  auth := v1.Group("/auth")
  {
      auth.POST("/signup", authController.Signup)
      auth.POST("/login", authController.Login)
      auth.POST("/refresh", authController.Refresh)
      auth.POST("/forgot-password", authController.ForgotPassword)
      auth.POST("/reset-password", authController.ResetPassword)
  }
  ```

- [ ] **마이그레이션 생성**

  ```bash
  cd apps/backend
  go generate ./pkg/ent
  moonx backend:migrate-diff -- password_reset_tokens
  moonx backend:migrate
  ```

### 검증

```bash
# 1. 비밀번호 재설정 요청
curl -X POST http://localhost:8080/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
# 200 OK + message (이메일 존재 여부와 무관하게 동일 응답)

# 2. 비밀번호 재설정 (토큰은 DB 또는 로그에서 확인)
curl -X POST http://localhost:8080/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{"token":"<reset_token>","new_password":"newpassword123"}'
# 200 OK + message

# 3. 새 비밀번호로 로그인 확인
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"newpassword123"}'
# 200 OK + tokens
```

### 보안 고려사항

1. **이메일 열거 방지**: 존재하지 않는 이메일에도 동일한 성공 응답
2. **토큰 보안**: 32바이트 암호학적 난수 사용
3. **토큰 만료**: 1시간 후 자동 만료
4. **일회성 토큰**: 사용 후 재사용 불가
5. **Rate Limiting**: (향후) 요청 횟수 제한 필요

---

## Step 2.7: Google OAuth

### 목표

- `GET /v1/auth/google` - Google OAuth 로그인 시작 (리다이렉트)
- `GET /v1/auth/google/callback` - Google OAuth 콜백 처리

### 아키텍처

```mermaid
sequenceDiagram
    participant C as Client (Browser)
    participant A as API Server
    participant G as Google OAuth
    participant DB as Database

    Note over C,DB: Google OAuth 로그인 플로우
    C->>A: GET /auth/google
    A->>A: state 토큰 생성 (CSRF 방지)
    A-->>C: Redirect to Google OAuth

    C->>G: Google 로그인 페이지
    G-->>C: 사용자 인증 후 callback으로 리다이렉트

    C->>A: GET /auth/google/callback?code=xxx&state=xxx
    A->>A: state 검증 (CSRF 방지)
    A->>G: code로 access token 교환
    G-->>A: access_token + id_token

    A->>G: 사용자 정보 요청 (userinfo)
    G-->>A: { email, name, picture, sub }

    A->>DB: 이메일로 사용자 조회 또는 생성
    DB-->>A: User

    A->>A: Access Token + Refresh Token 생성
    A-->>C: Redirect to frontend with tokens
```

### 체크리스트

- [ ] **의존성 추가**

  ```bash
  cd apps/backend
  go get golang.org/x/oauth2
  go get golang.org/x/oauth2/google
  ```

- [ ] **User 스키마 수정**
  - [ ] `pkg/ent/schema/user.go`에 OAuth 필드 추가

    ```go
    func (User) Fields() []ent.Field {
        return []ent.Field{
            // ... 기존 필드 ...
            field.String("password_hash").
                Optional(), // OAuth 사용자는 비밀번호 없음
            field.String("google_id").
                Optional().
                Unique(),
            field.String("avatar_url").
                Optional(),
            field.Enum("auth_provider").
                Values("email", "google").
                Default("email"),
        }
    }
    ```

- [ ] **OAuth 설정**
  - [ ] `pkg/infra/config/oauth.go`

    ```go
    package config

    import (
        "os"

        "golang.org/x/oauth2"
        "golang.org/x/oauth2/google"
    )

    type OAuthConfig struct {
        Google *oauth2.Config
    }

    func NewOAuthConfig() *OAuthConfig {
        return &OAuthConfig{
            Google: &oauth2.Config{
                ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
                ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
                RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
                Scopes: []string{
                    "https://www.googleapis.com/auth/userinfo.email",
                    "https://www.googleapis.com/auth/userinfo.profile",
                },
                Endpoint: google.Endpoint,
            },
        }
    }
    ```

- [ ] **OAuth 서비스 작성**
  - [ ] `pkg/service/oauth_service.go`

    ```go
    package service

    import (
        "context"
        "crypto/rand"
        "encoding/base64"
        "encoding/json"
        "errors"
        "fmt"
        "io"
        "net/http"

        "golang.org/x/oauth2"

        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/user"
    )

    var (
        ErrInvalidState = errors.New("invalid oauth state")
    )

    type GoogleUserInfo struct {
        ID            string `json:"id"`
        Email         string `json:"email"`
        VerifiedEmail bool   `json:"verified_email"`
        Name          string `json:"name"`
        Picture       string `json:"picture"`
    }

    type OAuthService struct {
        client       *ent.Client
        googleConfig *oauth2.Config
    }

    func NewOAuthService(client *ent.Client, googleConfig *oauth2.Config) *OAuthService {
        return &OAuthService{
            client:       client,
            googleConfig: googleConfig,
        }
    }

    // GenerateState creates a random state for CSRF protection
    func (s *OAuthService) GenerateState() (string, error) {
        bytes := make([]byte, 32)
        if _, err := rand.Read(bytes); err != nil {
            return "", err
        }
        return base64.URLEncoding.EncodeToString(bytes), nil
    }

    // GetGoogleAuthURL returns the Google OAuth authorization URL
    func (s *OAuthService) GetGoogleAuthURL(state string) string {
        return s.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
    }

    // ExchangeGoogleCode exchanges the authorization code for tokens
    func (s *OAuthService) ExchangeGoogleCode(ctx context.Context, code string) (*oauth2.Token, error) {
        return s.googleConfig.Exchange(ctx, code)
    }

    // GetGoogleUserInfo fetches user info from Google
    func (s *OAuthService) GetGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
        client := s.googleConfig.Client(ctx, token)
        resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
        if err != nil {
            return nil, fmt.Errorf("failed to get user info: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            body, _ := io.ReadAll(resp.Body)
            return nil, fmt.Errorf("google api error: %s", string(body))
        }

        var userInfo GoogleUserInfo
        if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
            return nil, fmt.Errorf("failed to decode user info: %w", err)
        }

        return &userInfo, nil
    }

    // FindOrCreateGoogleUser finds or creates a user from Google OAuth
    func (s *OAuthService) FindOrCreateGoogleUser(ctx context.Context, info *GoogleUserInfo) (*ent.User, error) {
        // 먼저 Google ID로 조회
        u, err := s.client.User.Query().
            Where(user.GoogleIDEQ(info.ID)).
            Only(ctx)
        if err == nil {
            return u, nil
        }
        if !ent.IsNotFound(err) {
            return nil, err
        }

        // Google ID가 없으면 이메일로 조회
        u, err = s.client.User.Query().
            Where(
                user.EmailEQ(info.Email),
                user.StatusEQ("active"),
            ).
            Only(ctx)
        if err == nil {
            // 기존 이메일 계정에 Google ID 연결
            return s.client.User.
                UpdateOneID(u.ID).
                SetGoogleID(info.ID).
                SetAvatarURL(info.Picture).
                Save(ctx)
        }
        if !ent.IsNotFound(err) {
            return nil, err
        }

        // 새 사용자 생성
        return s.client.User.
            Create().
            SetEmail(info.Email).
            SetGoogleID(info.ID).
            SetAvatarURL(info.Picture).
            SetAuthProvider("google").
            Save(ctx)
    }
    ```

- [ ] **OAuth 컨트롤러 작성**
  - [ ] `internal/api/controller/oauth_controller.go`

    ```go
    package controller

    import (
        "net/http"
        "os"
        "time"

        "github.com/gin-gonic/gin"

        "github.com/mindhit/api/pkg/service"
    )

    type OAuthController struct {
        oauthService *service.OAuthService
        jwtService   *service.JWTService
        frontendURL  string
    }

    func NewOAuthController(
        oauthService *service.OAuthService,
        jwtService *service.JWTService,
    ) *OAuthController {
        return &OAuthController{
            oauthService: oauthService,
            jwtService:   jwtService,
            frontendURL:  os.Getenv("FRONTEND_URL"),
        }
    }

    // GoogleLogin initiates Google OAuth flow
    func (c *OAuthController) GoogleLogin(ctx *gin.Context) {
        state, err := c.oauthService.GenerateState()
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{"message": "failed to generate state"},
            })
            return
        }

        // state를 쿠키에 저장 (CSRF 방지)
        ctx.SetCookie(
            "oauth_state",
            state,
            int(10*time.Minute.Seconds()),
            "/",
            "",
            true,  // secure
            true,  // httpOnly
        )

        url := c.oauthService.GetGoogleAuthURL(state)
        ctx.Redirect(http.StatusTemporaryRedirect, url)
    }

    // GoogleCallback handles Google OAuth callback
    func (c *OAuthController) GoogleCallback(ctx *gin.Context) {
        // state 검증
        state := ctx.Query("state")
        storedState, err := ctx.Cookie("oauth_state")
        if err != nil || state != storedState {
            ctx.Redirect(http.StatusTemporaryRedirect, c.frontendURL+"/auth/error?error=invalid_state")
            return
        }

        // 쿠키 삭제
        ctx.SetCookie("oauth_state", "", -1, "/", "", true, true)

        // code로 token 교환
        code := ctx.Query("code")
        if code == "" {
            ctx.Redirect(http.StatusTemporaryRedirect, c.frontendURL+"/auth/error?error=no_code")
            return
        }

        token, err := c.oauthService.ExchangeGoogleCode(ctx.Request.Context(), code)
        if err != nil {
            ctx.Redirect(http.StatusTemporaryRedirect, c.frontendURL+"/auth/error?error=exchange_failed")
            return
        }

        // 사용자 정보 가져오기
        userInfo, err := c.oauthService.GetGoogleUserInfo(ctx.Request.Context(), token)
        if err != nil {
            ctx.Redirect(http.StatusTemporaryRedirect, c.frontendURL+"/auth/error?error=userinfo_failed")
            return
        }

        // 사용자 찾기/생성
        user, err := c.oauthService.FindOrCreateGoogleUser(ctx.Request.Context(), userInfo)
        if err != nil {
            ctx.Redirect(http.StatusTemporaryRedirect, c.frontendURL+"/auth/error?error=user_failed")
            return
        }

        // JWT 토큰 쌍 생성
        tokenPair, err := c.jwtService.GenerateTokenPair(user.ID)
        if err != nil {
            ctx.Redirect(http.StatusTemporaryRedirect, c.frontendURL+"/auth/error?error=token_failed")
            return
        }

        // 프론트엔드로 리다이렉트 (토큰 포함)
        redirectURL := fmt.Sprintf(
            "%s/auth/callback?access_token=%s&refresh_token=%s&expires_in=%d",
            c.frontendURL,
            tokenPair.AccessToken,
            tokenPair.RefreshToken,
            tokenPair.ExpiresIn,
        )
        ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
    }
    ```

- [ ] **라우트 등록**

  ```go
  // In main.go
  auth := v1.Group("/auth")
  {
      // ... 기존 라우트 ...
      auth.GET("/google", oauthController.GoogleLogin)
      auth.GET("/google/callback", oauthController.GoogleCallback)
  }
  ```

- [ ] **환경 변수 설정**

  ```bash
  # .env.development
  GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
  GOOGLE_CLIENT_SECRET=your-client-secret
  GOOGLE_REDIRECT_URL=http://localhost:8080/v1/auth/google/callback
  FRONTEND_URL=http://localhost:3000
  ```

- [ ] **마이그레이션 생성**

  ```bash
  cd apps/backend
  go generate ./pkg/ent
  moonx backend:migrate-diff -- add_oauth_fields
  moonx backend:migrate
  ```

### 검증

```bash
# 1. Google OAuth 시작 (브라우저에서 접근)
open "http://localhost:8080/v1/auth/google"
# Google 로그인 페이지로 리다이렉트됨

# 2. 로그인 후 프론트엔드로 리다이렉트
# http://localhost:3000/auth/callback?access_token=...&refresh_token=...

# 3. 받은 토큰으로 API 호출
curl -X GET http://localhost:8080/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
# 200 OK + user (Google 프로필 정보 포함)
```

### Google Cloud Console 설정

1. [Google Cloud Console](https://console.cloud.google.com/) 접속
2. 프로젝트 생성 또는 선택
3. APIs & Services > Credentials > Create Credentials > OAuth client ID
4. Application type: Web application
5. Authorized redirect URIs 추가:
   - 개발: `http://localhost:8080/v1/auth/google/callback`
   - 프로덕션: `https://api.mindhit.io/v1/auth/google/callback`
6. Client ID와 Client Secret을 환경 변수에 설정

### 보안 고려사항

1. **state 파라미터**: CSRF 공격 방지를 위해 랜덤 state 사용
2. **HTTPS**: 프로덕션에서는 반드시 HTTPS 사용
3. **HttpOnly 쿠키**: state를 HttpOnly 쿠키에 저장
4. **토큰 저장**: 프론트엔드에서 안전하게 토큰 저장 필요
5. **이메일 검증**: Google은 이미 이메일 검증 완료

---

## Phase 2 완료 확인

### 전체 검증 체크리스트

- [ ] **회원가입 API**

  ```bash
  curl -X POST http://localhost:8080/v1/auth/signup \
    -H "Content-Type: application/json" \
    -d '{"email":"new@example.com","password":"password123"}'
  # 201 Created + user + access_token + refresh_token
  ```

- [ ] **로그인 API**

  ```bash
  curl -X POST http://localhost:8080/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"new@example.com","password":"password123"}'
  # 200 OK + user + access_token + refresh_token
  ```

- [ ] **토큰 갱신 API**

  ```bash
  curl -X POST http://localhost:8080/v1/auth/refresh \
    -H "Content-Type: application/json" \
    -d '{"refresh_token":"<refresh_token>"}'
  # 200 OK + access_token + expires_in
  ```

- [ ] **사용자 정보 API**

  ```bash
  curl -X GET http://localhost:8080/v1/auth/me \
    -H "Authorization: Bearer <access_token>"
  # 200 OK + user
  ```

- [ ] **잘못된 자격증명**

  ```bash
  curl -X POST http://localhost:8080/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"new@example.com","password":"wrongpassword"}'
  # 401 Unauthorized
  ```

- [ ] **중복 이메일**

  ```bash
  curl -X POST http://localhost:8080/v1/auth/signup \
    -H "Content-Type: application/json" \
    -d '{"email":"new@example.com","password":"password123"}'
  # 409 Conflict
  ```

- [ ] **만료된 토큰**

  ```bash
  curl -X GET http://localhost:8080/v1/auth/me \
    -H "Authorization: Bearer <expired_access_token>"
  # 401 Unauthorized
  ```

### 테스트 요구사항

| 테스트 유형 | 대상 | 파일 |
| ----------- | ---- | ---- |
| 단위 테스트 | JWT 토큰 생성/검증 | `jwt_service_test.go` |
| 단위 테스트 | 회원가입/로그인 로직 | `auth_service_test.go` |
| 통합 테스트 | Auth API 엔드포인트 | `auth_controller_test.go` |

```bash
# Phase 2 테스트 실행
moonx backend:test -- -run "TestJWT|TestAuth"
```

> **Note**: 모든 테스트가 통과해야 Phase 2 완료로 인정됩니다.

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| JWT 서비스 | `pkg/service/jwt_service.go` |
| Auth 서비스 | `pkg/service/auth_service.go` |
| Auth 컨트롤러 | `internal/api/controller/auth_controller.go` |
| Auth 미들웨어 | `pkg/infra/middleware/auth.go` |
| 테스트 | `pkg/service/*_test.go` |

### API 요약

| 메서드 | 엔드포인트 | 인증 | 설명 |
|-------|-----------|------|------|
| POST | `/v1/auth/signup` | - | 회원가입 |
| POST | `/v1/auth/login` | - | 로그인 |
| POST | `/v1/auth/refresh` | - | 토큰 갱신 |
| POST | `/v1/auth/forgot-password` | - | 비밀번호 재설정 요청 |
| POST | `/v1/auth/reset-password` | - | 비밀번호 재설정 완료 |
| GET | `/v1/auth/google` | - | Google OAuth 로그인 시작 |
| GET | `/v1/auth/google/callback` | - | Google OAuth 콜백 |
| GET | `/v1/auth/me` | Bearer | 사용자 정보 |
| POST | `/v1/auth/logout` | Bearer | 로그아웃 |

---

## 다음 Phase

Phase 2 완료 후 [Phase 3: 세션 관리 API](./phase-3-sessions.md)로 진행하세요.
