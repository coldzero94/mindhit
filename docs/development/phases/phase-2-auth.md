# Phase 2: 인증 시스템

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | JWT 기반 인증 시스템 구현 (Access + Refresh Token) |
| **선행 조건** | Phase 1, Phase 1.5 완료 |
| **예상 소요** | 5 Steps |
| **결과물** | 회원가입, 로그인, 토큰 갱신, 사용자 정보 API 동작 |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 2.1 | JWT 서비스 구현 (Access + Refresh) | ⬜ |
| 2.2 | Auth 서비스 구현 | ⬜ |
| 2.3 | Auth 컨트롤러 구현 | ⬜ |
| 2.4 | Auth 미들웨어 구현 | ⬜ |
| 2.5 | 토큰 갱신 및 사용자 정보 API | ⬜ |

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
  cd apps/api
  go get github.com/golang-jwt/jwt/v5
  ```

- [ ] **JWT 서비스 작성**
  - [ ] `internal/service/jwt_service.go`
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
  - [ ] `internal/service/jwt_service_test.go`
    ```go
    package service_test

    import (
        "testing"

        "github.com/google/uuid"
        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"

        "github.com/mindhit/api/internal/service"
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
cd apps/api
go test ./internal/service/... -v -run TestJWT
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
  - [ ] `internal/service/auth_service.go`
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
  - [ ] `internal/service/auth_service_test.go`
    ```go
    package service_test

    import (
        "context"
        "testing"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
        "github.com/stretchr/testify/suite"

        "github.com/mindhit/api/ent/enttest"
        "github.com/mindhit/api/internal/service"

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
cd apps/api
go test ./internal/service/... -v -run TestAuthService
```

---

## Step 2.3: Auth 컨트롤러 구현

### 목표
HTTP 핸들러로 API 엔드포인트 구현

### 체크리스트

- [ ] **Auth 컨트롤러 작성**
  - [ ] `internal/controller/auth_controller.go`
    ```go
    package controller

    import (
        "errors"
        "net/http"

        "github.com/gin-gonic/gin"

        "github.com/mindhit/api/internal/generated"
        "github.com/mindhit/api/internal/service"
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
cd apps/api && go run ./cmd/server

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
  - [ ] `internal/infrastructure/middleware/auth.go`
    ```go
    package middleware

    import (
        "net/http"
        "strings"

        "github.com/gin-gonic/gin"

        "github.com/mindhit/api/internal/service"
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
  - [ ] `internal/infrastructure/middleware/cors.go`
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
  - [ ] `internal/infrastructure/middleware/logging.go`
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
  - [ ] `internal/controller/auth_controller.go`에 메서드 추가
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

### 산출물 요약

| 항목 | 위치 |
|-----|------|
| JWT 서비스 | `internal/service/jwt_service.go` |
| Auth 서비스 | `internal/service/auth_service.go` |
| Auth 컨트롤러 | `internal/controller/auth_controller.go` |
| Auth 미들웨어 | `internal/infrastructure/middleware/auth.go` |
| 테스트 | `internal/service/*_test.go` |

### API 요약

| 메서드 | 엔드포인트 | 인증 | 설명 |
|-------|-----------|------|------|
| POST | `/v1/auth/signup` | - | 회원가입 |
| POST | `/v1/auth/login` | - | 로그인 |
| POST | `/v1/auth/refresh` | - | 토큰 갱신 |
| GET | `/v1/auth/me` | Bearer | 사용자 정보 |
| POST | `/v1/auth/logout` | Bearer | 로그아웃 |

---

## 다음 Phase

Phase 2 완료 후 [Phase 3: 세션 관리 API](./phase-3-sessions.md)로 진행하세요.
