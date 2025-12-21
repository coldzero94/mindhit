# Phase 3: 세션 관리 API

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | 브라우징 세션 CRUD API 구현 |
| **선행 조건** | Phase 2 완료 |
| **예상 소요** | 3 Steps |
| **결과물** | 세션 시작/일시정지/종료 API 동작 |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 3.1 | Session 서비스 구현 | ⬜ |
| 3.2 | Session 컨트롤러 구현 | ⬜ |
| 3.3 | Session API 테스트 | ⬜ |

---

## Step 3.1: Session 서비스 구현

### 목표
세션 생성, 상태 변경, 조회 비즈니스 로직

### 체크리스트

- [ ] **Session 서비스 작성**
  - [ ] `internal/service/session_service.go`
    ```go
    package service

    import (
        "context"
        "errors"
        "time"

        "github.com/google/uuid"
        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/highlight"
        "github.com/mindhit/api/ent/pagevisit"
        "github.com/mindhit/api/ent/session"
        "github.com/mindhit/api/ent/user"
    )

    var (
        ErrSessionNotFound     = errors.New("session not found")
        ErrSessionNotOwned     = errors.New("session not owned by user")
        ErrInvalidSessionState = errors.New("invalid session state transition")
    )

    type SessionService struct {
        client *ent.Client
    }

    func NewSessionService(client *ent.Client) *SessionService {
        return &SessionService{client: client}
    }

    // activeSessions returns a query filtered to active (non-deleted) sessions only
    func (s *SessionService) activeSessions() *ent.SessionQuery {
        return s.client.Session.Query().Where(session.StatusNEQ("deleted"))
    }

    // Start creates a new recording session
    func (s *SessionService) Start(ctx context.Context, userID uuid.UUID) (*ent.Session, error) {
        return s.client.Session.
            Create().
            SetUserID(userID).
            SetSessionStatus(session.SessionStatusRecording).
            SetStartedAt(time.Now()).
            Save(ctx)
    }

    // Pause pauses a recording session
    func (s *SessionService) Pause(ctx context.Context, sessionID, userID uuid.UUID) (*ent.Session, error) {
        sess, err := s.getOwnedSession(ctx, sessionID, userID)
        if err != nil {
            return nil, err
        }

        if sess.SessionStatus != session.SessionStatusRecording {
            return nil, ErrInvalidSessionState
        }

        return s.client.Session.
            UpdateOneID(sessionID).
            SetSessionStatus(session.SessionStatusPaused).
            Save(ctx)
    }

    // Resume resumes a paused session
    func (s *SessionService) Resume(ctx context.Context, sessionID, userID uuid.UUID) (*ent.Session, error) {
        sess, err := s.getOwnedSession(ctx, sessionID, userID)
        if err != nil {
            return nil, err
        }

        if sess.SessionStatus != session.SessionStatusPaused {
            return nil, ErrInvalidSessionState
        }

        return s.client.Session.
            UpdateOneID(sessionID).
            SetSessionStatus(session.SessionStatusRecording).
            Save(ctx)
    }

    // Stop stops a session and marks it for processing
    func (s *SessionService) Stop(ctx context.Context, sessionID, userID uuid.UUID) (*ent.Session, error) {
        sess, err := s.getOwnedSession(ctx, sessionID, userID)
        if err != nil {
            return nil, err
        }

        if sess.SessionStatus != session.SessionStatusRecording && sess.SessionStatus != session.SessionStatusPaused {
            return nil, ErrInvalidSessionState
        }

        now := time.Now()
        return s.client.Session.
            UpdateOneID(sessionID).
            SetSessionStatus(session.SessionStatusProcessing).
            SetEndedAt(now).
            Save(ctx)
    }

    // Get retrieves a session by ID with ownership check
    func (s *SessionService) Get(ctx context.Context, sessionID, userID uuid.UUID) (*ent.Session, error) {
        return s.getOwnedSession(ctx, sessionID, userID)
    }

    // GetWithDetails retrieves a session with all related data
    func (s *SessionService) GetWithDetails(ctx context.Context, sessionID, userID uuid.UUID) (*ent.Session, error) {
        sess, err := s.activeSessions().
            Where(session.IDEQ(sessionID)).
            WithUser().
            WithPageVisits(func(q *ent.PageVisitQuery) {
                q.Where(pagevisit.StatusEQ("active")).WithURL()
            }).
            WithHighlights(func(q *ent.HighlightQuery) {
                q.Where(highlight.StatusEQ("active"))
            }).
            WithMindmap().
            Only(ctx)

        if err != nil {
            if ent.IsNotFound(err) {
                return nil, ErrSessionNotFound
            }
            return nil, err
        }

        if sess.Edges.User.ID != userID {
            return nil, ErrSessionNotOwned
        }

        return sess, nil
    }

    // ListByUser retrieves all active sessions for a user
    func (s *SessionService) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*ent.Session, error) {
        return s.activeSessions().
            Where(session.HasUserWith(user.IDEQ(userID))).
            Order(ent.Desc(session.FieldCreatedAt)).
            Limit(limit).
            Offset(offset).
            All(ctx)
    }

    // Update updates session metadata (title, description)
    func (s *SessionService) Update(ctx context.Context, sessionID, userID uuid.UUID, title, description *string) (*ent.Session, error) {
        sess, err := s.getOwnedSession(ctx, sessionID, userID)
        if err != nil {
            return nil, err
        }

        update := s.client.Session.UpdateOneID(sess.ID)

        if title != nil {
            update.SetTitle(*title)
        }
        if description != nil {
            update.SetDescription(*description)
        }

        return update.Save(ctx)
    }

    // Delete soft-deletes a session (sets status to "deleted")
    func (s *SessionService) Delete(ctx context.Context, sessionID, userID uuid.UUID) error {
        sess, err := s.getOwnedSession(ctx, sessionID, userID)
        if err != nil {
            return err
        }

        now := time.Now()
        _, err = s.client.Session.
            UpdateOneID(sess.ID).
            SetStatus("deleted").
            SetDeletedAt(now).
            Save(ctx)
        return err
    }

    // getOwnedSession retrieves an active session and verifies ownership
    func (s *SessionService) getOwnedSession(ctx context.Context, sessionID, userID uuid.UUID) (*ent.Session, error) {
        sess, err := s.activeSessions().
            Where(session.IDEQ(sessionID)).
            WithUser().
            Only(ctx)

        if err != nil {
            if ent.IsNotFound(err) {
                return nil, ErrSessionNotFound
            }
            return nil, err
        }

        if sess.Edges.User.ID != userID {
            return nil, ErrSessionNotOwned
        }

        return sess, nil
    }
    ```

### 검증
```bash
go build ./...
# 컴파일 에러 없음
```

---

## Step 3.2: Session 컨트롤러 구현

### 목표
세션 API 엔드포인트 구현

### 체크리스트

- [ ] **Session 컨트롤러 작성**
  - [ ] `internal/controller/session_controller.go`
    ```go
    package controller

    import (
        "errors"
        "net/http"

        "github.com/gin-gonic/gin"
        "github.com/google/uuid"

        "github.com/mindhit/api/internal/infrastructure/middleware"
        "github.com/mindhit/api/internal/service"
    )

    type SessionController struct {
        sessionService *service.SessionService
    }

    func NewSessionController(sessionService *service.SessionService) *SessionController {
        return &SessionController{sessionService: sessionService}
    }

    func (c *SessionController) Start(ctx *gin.Context) {
        userID, ok := middleware.GetUserID(ctx)
        if !ok {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{"message": "unauthorized"},
            })
            return
        }

        session, err := c.sessionService.Start(ctx.Request.Context(), userID)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{"message": "failed to create session"},
            })
            return
        }

        ctx.JSON(http.StatusCreated, gin.H{
            "session": mapSession(session),
        })
    }

    func (c *SessionController) Pause(ctx *gin.Context) {
        userID, ok := middleware.GetUserID(ctx)
        if !ok {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{"message": "unauthorized"},
            })
            return
        }

        sessionID, err := uuid.Parse(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{"message": "invalid session id"},
            })
            return
        }

        session, err := c.sessionService.Pause(ctx.Request.Context(), sessionID, userID)
        if err != nil {
            c.handleError(ctx, err)
            return
        }

        ctx.JSON(http.StatusOK, gin.H{
            "session": mapSession(session),
        })
    }

    func (c *SessionController) Resume(ctx *gin.Context) {
        userID, ok := middleware.GetUserID(ctx)
        if !ok {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{"message": "unauthorized"},
            })
            return
        }

        sessionID, err := uuid.Parse(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{"message": "invalid session id"},
            })
            return
        }

        session, err := c.sessionService.Resume(ctx.Request.Context(), sessionID, userID)
        if err != nil {
            c.handleError(ctx, err)
            return
        }

        ctx.JSON(http.StatusOK, gin.H{
            "session": mapSession(session),
        })
    }

    func (c *SessionController) Stop(ctx *gin.Context) {
        userID, ok := middleware.GetUserID(ctx)
        if !ok {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{"message": "unauthorized"},
            })
            return
        }

        sessionID, err := uuid.Parse(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{"message": "invalid session id"},
            })
            return
        }

        session, err := c.sessionService.Stop(ctx.Request.Context(), sessionID, userID)
        if err != nil {
            c.handleError(ctx, err)
            return
        }

        ctx.JSON(http.StatusOK, gin.H{
            "session": mapSession(session),
        })
    }

    func (c *SessionController) Get(ctx *gin.Context) {
        userID, ok := middleware.GetUserID(ctx)
        if !ok {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{"message": "unauthorized"},
            })
            return
        }

        sessionID, err := uuid.Parse(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{"message": "invalid session id"},
            })
            return
        }

        session, err := c.sessionService.GetWithDetails(ctx.Request.Context(), sessionID, userID)
        if err != nil {
            c.handleError(ctx, err)
            return
        }

        ctx.JSON(http.StatusOK, gin.H{
            "session": mapSessionWithDetails(session),
        })
    }

    func (c *SessionController) List(ctx *gin.Context) {
        userID, ok := middleware.GetUserID(ctx)
        if !ok {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{"message": "unauthorized"},
            })
            return
        }

        // TODO: Parse pagination from query params
        limit := 20
        offset := 0

        sessions, err := c.sessionService.ListByUser(ctx.Request.Context(), userID, limit, offset)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{"message": "failed to list sessions"},
            })
            return
        }

        result := make([]map[string]interface{}, len(sessions))
        for i, s := range sessions {
            result[i] = mapSession(s)
        }

        ctx.JSON(http.StatusOK, gin.H{
            "sessions": result,
        })
    }

    type UpdateSessionRequest struct {
        Title       *string `json:"title"`
        Description *string `json:"description"`
    }

    func (c *SessionController) Update(ctx *gin.Context) {
        userID, ok := middleware.GetUserID(ctx)
        if !ok {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{"message": "unauthorized"},
            })
            return
        }

        sessionID, err := uuid.Parse(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{"message": "invalid session id"},
            })
            return
        }

        var req UpdateSessionRequest
        if err := ctx.ShouldBindJSON(&req); err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{"message": "invalid request body"},
            })
            return
        }

        session, err := c.sessionService.Update(ctx.Request.Context(), sessionID, userID, req.Title, req.Description)
        if err != nil {
            c.handleError(ctx, err)
            return
        }

        ctx.JSON(http.StatusOK, gin.H{
            "session": mapSession(session),
        })
    }

    func (c *SessionController) Delete(ctx *gin.Context) {
        userID, ok := middleware.GetUserID(ctx)
        if !ok {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{"message": "unauthorized"},
            })
            return
        }

        sessionID, err := uuid.Parse(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{"message": "invalid session id"},
            })
            return
        }

        err = c.sessionService.Delete(ctx.Request.Context(), sessionID, userID)
        if err != nil {
            c.handleError(ctx, err)
            return
        }

        ctx.JSON(http.StatusNoContent, nil)
    }

    func (c *SessionController) handleError(ctx *gin.Context, err error) {
        switch {
        case errors.Is(err, service.ErrSessionNotFound):
            ctx.JSON(http.StatusNotFound, gin.H{
                "error": gin.H{"message": "session not found"},
            })
        case errors.Is(err, service.ErrSessionNotOwned):
            ctx.JSON(http.StatusForbidden, gin.H{
                "error": gin.H{"message": "access denied"},
            })
        case errors.Is(err, service.ErrInvalidSessionState):
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{"message": "invalid session state for this operation"},
            })
        default:
            ctx.JSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{"message": "internal server error"},
            })
        }
    }

    func mapSession(s *ent.Session) map[string]interface{} {
        result := map[string]interface{}{
            "id":            s.ID.String(),
            "sessionStatus": string(s.SessionStatus),
            "startedAt":     s.StartedAt,
            "createdAt":     s.CreatedAt,
            "updatedAt":     s.UpdatedAt,
        }
        if s.Title != nil {
            result["title"] = *s.Title
        }
        if s.Description != nil {
            result["description"] = *s.Description
        }
        if s.EndedAt != nil {
            result["endedAt"] = *s.EndedAt
        }
        return result
    }

    func mapSessionWithDetails(s *ent.Session) map[string]interface{} {
        result := mapSession(s)
        // Add page visits, highlights, mindmap if loaded
        // ... (상세 매핑 로직)
        return result
    }
    ```

- [ ] **main.go에 라우트 추가**
  ```go
  // Session routes (protected)
  sessionController := controller.NewSessionController(sessionService)

  sessions := v1.Group("/sessions")
  sessions.Use(middleware.Auth(jwtService))
  {
      sessions.POST("/start", sessionController.Start)
      sessions.GET("", sessionController.List)
      sessions.GET("/:id", sessionController.Get)
      sessions.PUT("/:id", sessionController.Update)
      sessions.PATCH("/:id/pause", sessionController.Pause)
      sessions.PATCH("/:id/resume", sessionController.Resume)
      sessions.POST("/:id/stop", sessionController.Stop)
      sessions.DELETE("/:id", sessionController.Delete)
  }
  ```

### 검증
```bash
# 서버 실행
go run ./cmd/server

# 로그인하여 토큰 획득
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' | jq -r '.token')

# 세션 시작
curl -X POST http://localhost:8080/v1/sessions/start \
  -H "Authorization: Bearer $TOKEN"
```

---

## Step 3.3: Session API 테스트

### 목표
E2E 테스트로 전체 플로우 검증

### 체크리스트

- [ ] **E2E 테스트 시나리오**
  ```bash
  # 1. 사용자 생성
  curl -X POST http://localhost:8080/v1/auth/signup \
    -H "Content-Type: application/json" \
    -d '{"email":"session-test@example.com","password":"password123"}'

  # 2. 로그인
  TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"session-test@example.com","password":"password123"}' | jq -r '.token')

  # 3. 세션 시작
  SESSION_ID=$(curl -s -X POST http://localhost:8080/v1/sessions/start \
    -H "Authorization: Bearer $TOKEN" | jq -r '.session.id')
  echo "Session ID: $SESSION_ID"

  # 4. 세션 목록 조회
  curl -X GET http://localhost:8080/v1/sessions \
    -H "Authorization: Bearer $TOKEN"

  # 5. 세션 상세 조회
  curl -X GET http://localhost:8080/v1/sessions/$SESSION_ID \
    -H "Authorization: Bearer $TOKEN"

  # 6. 세션 제목/설명 업데이트
  curl -X PUT http://localhost:8080/v1/sessions/$SESSION_ID \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"My Research Session","description":"Researching AI topics"}'

  # 7. 세션 일시정지
  curl -X PATCH http://localhost:8080/v1/sessions/$SESSION_ID/pause \
    -H "Authorization: Bearer $TOKEN"

  # 8. 세션 재개
  curl -X PATCH http://localhost:8080/v1/sessions/$SESSION_ID/resume \
    -H "Authorization: Bearer $TOKEN"

  # 9. 세션 종료
  curl -X POST http://localhost:8080/v1/sessions/$SESSION_ID/stop \
    -H "Authorization: Bearer $TOKEN"

  # 10. 종료된 세션 상태 확인
  curl -X GET http://localhost:8080/v1/sessions/$SESSION_ID \
    -H "Authorization: Bearer $TOKEN"
  # status: "processing"
  ```

- [ ] **에러 케이스 테스트**
  ```bash
  # 인증 없이 요청
  curl -X POST http://localhost:8080/v1/sessions/start
  # 401 Unauthorized

  # 잘못된 세션 ID
  curl -X GET http://localhost:8080/v1/sessions/invalid-uuid \
    -H "Authorization: Bearer $TOKEN"
  # 400 Bad Request

  # 존재하지 않는 세션
  curl -X GET http://localhost:8080/v1/sessions/00000000-0000-0000-0000-000000000000 \
    -H "Authorization: Bearer $TOKEN"
  # 404 Not Found

  # 잘못된 상태 전환 (이미 정지된 세션 다시 정지)
  curl -X PATCH http://localhost:8080/v1/sessions/$SESSION_ID/pause \
    -H "Authorization: Bearer $TOKEN"
  # 400 Bad Request (이미 processing 상태)
  ```

- [ ] **자동화 테스트 작성** (선택)
  - [ ] `test/e2e/session_test.go`

### 검증
모든 테스트 시나리오가 예상대로 동작하는지 확인

---

## Phase 3 완료 확인

### 전체 검증 체크리스트

- [ ] `POST /v1/sessions/start` - 세션 생성
- [ ] `GET /v1/sessions` - 세션 목록
- [ ] `GET /v1/sessions/:id` - 세션 상세
- [ ] `PUT /v1/sessions/:id` - 세션 업데이트 (제목/설명)
- [ ] `PATCH /v1/sessions/:id/pause` - 일시정지
- [ ] `PATCH /v1/sessions/:id/resume` - 재개
- [ ] `POST /v1/sessions/:id/stop` - 종료
- [ ] `DELETE /v1/sessions/:id` - 삭제
- [ ] 상태 전환 규칙 적용됨

### 산출물 요약

| 항목 | 위치 |
|-----|------|
| Session 서비스 | `internal/service/session_service.go` |
| Session 컨트롤러 | `internal/controller/session_controller.go` |

---

## 다음 Phase

Phase 3 완료 후 [Phase 4: 이벤트 수집 API](./phase-4-events.md)로 진행하세요.
