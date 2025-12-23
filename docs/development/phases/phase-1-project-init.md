# Phase 1: 프로젝트 초기화

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | 모노레포 구조 설정 및 Go 백엔드 기반 구축 |
| **선행 조건** | Phase 0 완료 |
| **예상 소요** | 10 Steps |
| **결과물** | 동작하는 Go 서버 + PostgreSQL 연결 + 테스트 인프라 |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 1.1 | 모노레포 구조 설정 | ⬜ |
| 1.2 | Go 백엔드 프로젝트 초기화 | ⬜ |
| 1.3 | Ent ORM 설정 | ⬜ |
| 1.4 | Atlas Migration 설정 | ⬜ |
| 1.5 | PostgreSQL Docker Compose 설정 | ⬜ |
| 1.6 | Gin 서버 기본 설정 | ⬜ |
| 1.7 | Ent 스키마 정의 - 핵심 엔티티 | ⬜ |
| 1.8 | Ent 스키마 정의 - 보조 엔티티 | ⬜ |
| 1.9 | 첫 번째 Migration 생성 및 적용 | ⬜ |
| 1.10 | 테스트 인프라 설정 | ⬜ |

---

## Step 1.1: 모노레포 구조 확인

### 목표

Phase 0에서 설정한 Moon + pnpm 모노레포 구조 확인 및 검증

> **참고**: Moon 설치 및 기본 설정은 [Phase 0: Moon + Docker 개발 환경](./phase-0-dev-environment.md)에서 완료됩니다.

### 체크리스트

- [ ] **Phase 0 완료 확인**
  - [ ] Moon이 설치되어 있는지 확인
  - [ ] Docker Compose가 동작하는지 확인
  - [ ] GitHub Actions CI 워크플로우가 설정되어 있는지 확인

- [ ] **디렉토리 구조 확인**

  ```bash
  ls -la .moon/
  ls -la apps/
  ls -la packages/
  ```

- [ ] **Moon 프로젝트 확인**

  ```bash
  pnpm moon query projects
  ```

### 검증

```bash
# Moon 버전 확인
pnpm moon --version

# 프로젝트 목록 확인
pnpm moon query projects

# Docker 서비스 확인
docker-compose ps
```

### 결과물 (Phase 0에서 생성됨)

```text
mindhit/
├── .moon/
│   ├── workspace.yml     # 워크스페이스 설정
│   ├── toolchain.yml     # 도구 버전 설정
│   └── tasks.yml         # 전역 태스크 설정
├── .github/
│   └── workflows/
│       └── ci.yml        # Moon CI 워크플로우
├── apps/
│   ├── backend/          # Go 백엔드 (API + Worker)
│   ├── web/              # Next.js 웹앱
│   └── extension/        # Chrome Extension
├── packages/
│   ├── shared/           # 공유 유틸
│   └── protocol/         # API 타입 정의
├── infra/
│   ├── docker/           # Docker Compose (go run 모드)
│   ├── kind/             # kind 클러스터 설정
│   ├── helm/             # Helm Charts
│   └── modules/          # Terraform 모듈
├── .moon/                # Moon 설정
├── package.json
├── pnpm-workspace.yaml
└── pnpm-lock.yaml
```

---

## Step 1.2: Go 백엔드 프로젝트 초기화

### 목표

Go 모듈 및 기본 프로젝트 구조 설정

### 체크리스트

- [ ] **Go 모듈 초기화**

  ```bash
  cd apps/api
  go mod init github.com/mindhit/api
  ```

- [ ] **디렉토리 구조 생성**

  ```bash
  mkdir -p cmd/server
  mkdir -p internal/{controller,service,infrastructure/{config,middleware,logger}}
  mkdir -p ent/schema
  mkdir -p test/e2e
  ```

- [ ] **기본 파일 생성**
  - [ ] `cmd/server/main.go`

    ```go
    package main

    import "fmt"

    func main() {
        fmt.Println("MindHit API Server")
    }
    ```

- [ ] **golangci-lint 설정**
  - [ ] `.golangci.yml` 생성

    ```yaml
    run:
      timeout: 5m

    linters:
      enable:
        - errcheck
        - gosimple
        - govet
        - ineffassign
        - staticcheck
        - unused
        - gofmt
        - goimports
        - misspell
        - gocritic
        - revive

    linters-settings:
      gofmt:
        simplify: true
      goimports:
        local-prefixes: github.com/mindhit

    issues:
      exclude-use-default: false
    ```

- [ ] **moon.yml 설정**
  - [ ] `apps/api/moon.yml` 생성

    ```yaml
    language: go
    type: application

    tasks:
      build:
        command: go build -o ./bin/server ./cmd/server
        inputs:
          - "**/*.go"
          - "go.mod"
          - "go.sum"
        outputs:
          - "bin/server"

      test:
        command: go test -v -race -coverprofile=coverage.out ./...
        inputs:
          - "**/*.go"

      lint:
        command: golangci-lint run
        inputs:
          - "**/*.go"

      run:
        command: go run ./cmd/server
        local: true
    ```

### 검증

```bash
cd apps/api
go build ./cmd/server
go run ./cmd/server
# 출력: MindHit API Server
```

### 결과물

```
apps/api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── controller/
│   ├── service/
│   └── infrastructure/
│       ├── config/
│       ├── middleware/
│       └── logger/
├── ent/
│   └── schema/
├── test/
│   └── e2e/
├── go.mod
├── moon.yml
└── .golangci.yml
```

---

## Step 1.3: Ent ORM 설정

### 목표

Ent ORM 의존성 추가 및 기본 설정

### 체크리스트

- [ ] **Ent 의존성 추가**

  ```bash
  cd apps/api
  go get entgo.io/ent/cmd/ent@latest
  go get entgo.io/ent@latest
  ```

- [ ] **Ent 초기화**

  ```bash
  go run -mod=mod entgo.io/ent/cmd/ent new User
  ```

- [ ] **generate.go 생성**
  - [ ] `ent/generate.go` 확인/생성

    ```go
    package ent

    //go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./schema
    ```

- [ ] **moon.yml에 generate 태스크 추가**

  ```yaml
  generate:
    command: go generate ./ent
    inputs:
      - "ent/schema/**/*.go"
  ```

### 검증

```bash
cd apps/api
go generate ./ent
# ent/ 디렉토리에 생성된 파일 확인
ls ent/
```

### 결과물

```
apps/api/ent/
├── schema/
│   └── user.go        # 빈 스키마
├── client.go          # 생성됨
├── ent.go             # 생성됨
├── generate.go
└── ...                # 기타 생성 파일
```

---

## Step 1.4: Atlas Migration 설정

### 목표

Atlas CLI 설정 및 migration 워크플로우 구성

### 체크리스트

- [ ] **Atlas 설치** (로컬)

  ```bash
  # macOS
  brew install ariga/tap/atlas

  # 또는 curl
  curl -sSf https://atlasgo.sh | sh
  ```

- [ ] **atlas.hcl 생성**
  - [ ] `apps/api/atlas.hcl`

    ```hcl
    env "local" {
      src = "ent://ent/schema"
      dev = "postgres://postgres:password@localhost:5432/mindhit_dev?sslmode=disable"
      migration {
        dir = "file://ent/migrate/migrations"
      }
    }

    env "prod" {
      src = "ent://ent/schema"
      migration {
        dir = "file://ent/migrate/migrations"
      }
    }
    ```

- [ ] **migration 디렉토리 생성**

  ```bash
  mkdir -p apps/api/ent/migrate/migrations
  ```

- [ ] **moon.yml에 migration 태스크 추가**

  ```yaml
  migrate-diff:
    command: atlas migrate diff ${MIGRATION_NAME} --dir "file://ent/migrate/migrations" --to "ent://ent/schema" --dev-url "${DEV_DATABASE_URL}"
    local: true

  migrate-apply:
    command: atlas migrate apply --dir "file://ent/migrate/migrations" --url "${DATABASE_URL}"
    local: true

  migrate-status:
    command: atlas migrate status --dir "file://ent/migrate/migrations" --url "${DATABASE_URL}"
    local: true
  ```

### 검증

```bash
atlas version
# Atlas CLI version x.x.x
```

### 결과물

```
apps/api/
├── atlas.hcl
└── ent/
    └── migrate/
        └── migrations/     # 빈 디렉토리
```

---

## Step 1.5: PostgreSQL Docker Compose 설정

### 목표

로컬 개발용 PostgreSQL/Redis 컨테이너 설정

> **Note**: Docker Compose 설정은 [Phase 0](./phase-0-dev-environment.md)에서 이미 완료됩니다.
> 이 단계에서는 Phase 0에서 생성한 인프라가 정상 동작하는지 확인합니다.

### 체크리스트

- [ ] **Phase 0 완료 확인**
  - Docker Compose 파일 위치: `infra/docker/docker-compose.yml`

- [ ] **환경 변수 파일 생성**
  - [ ] `apps/backend/.env.local`

    ```
    ENVIRONMENT=local
    API_PORT=8081
    DATABASE_URL=postgres://postgres:password@localhost:5432/mindhit?sslmode=disable
    REDIS_URL=redis://localhost:6379
    JWT_SECRET=your-secret-key-change-in-production
    WORKER_CONCURRENCY=10
    ```

### 검증

```bash
# Docker Compose 실행 (Phase 0에서 설정)
moon run infra:dev-up

# 연결 테스트
docker exec -it mindhit-postgres psql -U postgres -d mindhit -c "SELECT 1;"

# Redis 테스트
docker exec -it mindhit-redis redis-cli ping
```

### 결과물

```
mindhit/
├── infra/
│   └── docker/
│       └── docker-compose.yml  # Phase 0에서 생성
└── apps/backend/
    └── .env.local
```

---

## Step 1.6: Gin 서버 기본 설정

### 목표

Gin 프레임워크로 기본 HTTP 서버 구성

### 체크리스트

- [ ] **의존성 추가**

  ```bash
  cd apps/api
  go get github.com/gin-gonic/gin
  go get github.com/gin-contrib/cors
  go get github.com/joho/godotenv
  ```

- [ ] **config 패키지 작성**
  - [ ] `internal/infrastructure/config/config.go`

    ```go
    package config

    import (
        "os"

        "github.com/joho/godotenv"
    )

    type Config struct {
        Port        string
        Environment string
        DatabaseURL string
        JWTSecret   string
        RedisURL    string
    }

    func Load() *Config {
        // .env 파일 로드 (있으면)
        _ = godotenv.Load()

        return &Config{
            Port:        getEnv("PORT", "8080"),
            Environment: getEnv("ENVIRONMENT", "development"),
            DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/mindhit?sslmode=disable"),
            JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
            RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
        }
    }

    func getEnv(key, defaultValue string) string {
        if value := os.Getenv(key); value != "" {
            return value
        }
        return defaultValue
    }
    ```

- [ ] **main.go 업데이트**
  - [ ] `cmd/server/main.go`

    ```go
    package main

    import (
        "log/slog"
        "net/http"
        "os"

        "github.com/gin-contrib/cors"
        "github.com/gin-gonic/gin"

        "github.com/mindhit/api/internal/infrastructure/config"
    )

    func main() {
        logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
        slog.SetDefault(logger)

        cfg := config.Load()

        if cfg.Environment == "production" {
            gin.SetMode(gin.ReleaseMode)
        }

        r := gin.New()
        r.Use(gin.Recovery())
        r.Use(cors.Default())

        // Health check
        r.GET("/health", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"status": "ok"})
        })

        slog.Info("starting server", "port", cfg.Port, "env", cfg.Environment)
        if err := r.Run(":" + cfg.Port); err != nil {
            slog.Error("server error", "error", err)
            os.Exit(1)
        }
    }
    ```

### 검증

```bash
cd apps/api
go run ./cmd/server

# 다른 터미널에서
curl http://localhost:8080/health
# {"status":"ok"}
```

### 결과물

- 동작하는 Gin 서버
- `/health` 엔드포인트 응답

---

## Step 1.7: Ent 스키마 정의 - 핵심 엔티티

### 목표

User, Session, URL 핵심 엔티티 스키마 정의

### 공통 Mixin 정의

모든 엔티티에서 공유하는 공통 필드를 Mixin으로 정의합니다.

#### 공통 필드 규칙

| 필드 | 타입 | 설명 | 적용 대상 |
|-----|------|------|----------|
| `id` | UUID | Primary Key | 모든 엔티티 |
| `status` | Enum(active, inactive) | 운영 상태 (soft delete) | 주요 엔티티 |
| `created_at` | Time | 생성 시각 | 모든 엔티티 |
| `updated_at` | Time | 수정 시각 | 모든 엔티티 |
| `deleted_at` | Time (nullable) | 삭제 시각 (soft delete) | 주요 엔티티 |

### 체크리스트

- [ ] **BaseMixin 정의** (모든 엔티티 공통)
  - [ ] `ent/schema/mixin/base.go`

    ```go
    package mixin

    import (
        "time"

        "entgo.io/ent"
        "entgo.io/ent/schema/field"
        "entgo.io/ent/schema/mixin"
        "github.com/google/uuid"
    )

    // BaseMixin defines common fields for all entities
    type BaseMixin struct {
        mixin.Schema
    }

    func (BaseMixin) Fields() []ent.Field {
        return []ent.Field{
            field.UUID("id", uuid.UUID{}).
                Default(uuid.New).
                Immutable().
                Comment("Primary key"),
            field.Time("created_at").
                Default(time.Now).
                Immutable().
                Comment("Record creation timestamp"),
            field.Time("updated_at").
                Default(time.Now).
                UpdateDefault(time.Now).
                Comment("Record last update timestamp"),
        }
    }
    ```

- [ ] **SoftDeleteMixin 정의** (주요 엔티티용)
  - [ ] `ent/schema/mixin/soft_delete.go`

    ```go
    package mixin

    import (
        "entgo.io/ent"
        "entgo.io/ent/schema/field"
        "entgo.io/ent/schema/mixin"
    )

    // SoftDeleteMixin adds soft delete capability
    type SoftDeleteMixin struct {
        mixin.Schema
    }

    func (SoftDeleteMixin) Fields() []ent.Field {
        return []ent.Field{
            field.Enum("status").
                Values("active", "inactive").
                Default("active").
                Comment("Record status for soft delete"),
            field.Time("deleted_at").
                Optional().
                Nillable().
                Comment("Soft delete timestamp"),
        }
    }
    ```

- [ ] **AuditMixin 정의** (감사 추적용, 선택적)
  - [ ] `ent/schema/mixin/audit.go`

    ```go
    package mixin

    import (
        "entgo.io/ent"
        "entgo.io/ent/schema/field"
        "entgo.io/ent/schema/mixin"
        "github.com/google/uuid"
    )

    // AuditMixin adds audit trail fields
    type AuditMixin struct {
        mixin.Schema
    }

    func (AuditMixin) Fields() []ent.Field {
        return []ent.Field{
            field.UUID("created_by", uuid.UUID{}).
                Optional().
                Nillable().
                Comment("User who created this record"),
            field.UUID("updated_by", uuid.UUID{}).
                Optional().
                Nillable().
                Comment("User who last updated this record"),
        }
    }
    ```

- [ ] **User 스키마** (BaseMixin + SoftDeleteMixin 적용)
  - [ ] `ent/schema/user.go`

    ```go
    package schema

    import (
        "entgo.io/ent"
        "entgo.io/ent/schema/edge"
        "entgo.io/ent/schema/field"
        "entgo.io/ent/schema/index"

        "github.com/mindhit/api/ent/schema/mixin"
    )

    type User struct {
        ent.Schema
    }

    func (User) Mixin() []ent.Mixin {
        return []ent.Mixin{
            mixin.BaseMixin{},
            mixin.SoftDeleteMixin{},
        }
    }

    func (User) Fields() []ent.Field {
        return []ent.Field{
            field.String("email").
                Unique().
                NotEmpty().
                Comment("User email address"),
            field.String("password_hash").
                Sensitive().
                Comment("Hashed password"),
        }
    }

    func (User) Edges() []ent.Edge {
        return []ent.Edge{
            edge.To("sessions", Session.Type),
            edge.To("settings", UserSettings.Type).
                Unique(),
        }
    }

    func (User) Indexes() []ent.Index {
        return []ent.Index{
            index.Fields("email"),
            index.Fields("status"),
        }
    }
    ```

- [ ] **Session 스키마** (BaseMixin 적용, 별도 session_status 사용)
  - [ ] `ent/schema/session.go`

    ```go
    package schema

    import (
        "time"

        "entgo.io/ent"
        "entgo.io/ent/schema/edge"
        "entgo.io/ent/schema/field"
        "entgo.io/ent/schema/index"

        "github.com/mindhit/api/ent/schema/mixin"
    )

    type Session struct {
        ent.Schema
    }

    func (Session) Mixin() []ent.Mixin {
        return []ent.Mixin{
            mixin.BaseMixin{},
        }
    }

    func (Session) Fields() []ent.Field {
        return []ent.Field{
            field.String("title").
                Optional().
                Nillable().
                Comment("Session title"),
            field.Text("description").
                Optional().
                Nillable().
                Comment("Session description"),
            field.Enum("session_status").
                Values("recording", "paused", "processing", "completed", "failed").
                Default("recording").
                Comment("Session workflow status"),
            field.Enum("status").
                Values("active", "inactive").
                Default("active").
                Comment("Record status for soft delete"),
            field.Time("started_at").
                Default(time.Now).
                Comment("Session start time"),
            field.Time("ended_at").
                Optional().
                Nillable().
                Comment("Session end time"),
            field.Time("deleted_at").
                Optional().
                Nillable().
                Comment("Soft delete timestamp"),
        }
    }

    func (Session) Edges() []ent.Edge {
        return []ent.Edge{
            edge.From("user", User.Type).
                Ref("sessions").
                Unique().
                Required(),
            edge.To("page_visits", PageVisit.Type),
            edge.To("highlights", Highlight.Type),
            edge.To("raw_events", RawEvent.Type),
            edge.To("mindmap", MindmapGraph.Type).
                Unique(),
        }
    }

    func (Session) Indexes() []ent.Index {
        return []ent.Index{
            index.Fields("session_status"),
            index.Fields("status"),
        }
    }
    ```

- [ ] **URL 스키마** (BaseMixin + SoftDeleteMixin 적용)
  - [ ] `ent/schema/url.go`

    ```go
    package schema

    import (
        "entgo.io/ent"
        "entgo.io/ent/schema/edge"
        "entgo.io/ent/schema/field"
        "entgo.io/ent/schema/index"

        "github.com/mindhit/api/ent/schema/mixin"
    )

    type URL struct {
        ent.Schema
    }

    func (URL) Mixin() []ent.Mixin {
        return []ent.Mixin{
            mixin.BaseMixin{},
            mixin.SoftDeleteMixin{},
        }
    }

    func (URL) Fields() []ent.Field {
        return []ent.Field{
            field.String("url").
                NotEmpty().
                Comment("Original URL"),
            field.String("url_hash").
                Unique().
                NotEmpty().
                Comment("SHA256 hash of normalized URL"),
            field.String("title").
                Optional().
                Comment("Page title"),
            field.Text("content").
                Optional().
                Comment("Extracted page content"),
            field.Text("summary").
                Optional().
                Comment("AI-generated summary"),
            field.JSON("keywords", []string{}).
                Optional().
                Comment("AI-extracted keywords"),
            field.Time("crawled_at").
                Optional().
                Nillable().
                Comment("Last time the URL content was crawled"),
        }
    }

    func (URL) Edges() []ent.Edge {
        return []ent.Edge{
            edge.From("page_visits", PageVisit.Type).
                Ref("url"),
        }
    }

    func (URL) Indexes() []ent.Index {
        return []ent.Index{
            index.Fields("url_hash"),
            index.Fields("status"),
        }
    }
    ```

- [ ] **uuid 의존성 추가**

  ```bash
  go get github.com/google/uuid
  ```

### 검증

```bash
# 아직 다른 엔티티 참조로 인해 generate 실패할 수 있음
# Step 1.8 완료 후 검증
```

---

## Step 1.8: Ent 스키마 정의 - 보조 엔티티

### 목표

PageVisit, Highlight, RawEvent, MindmapGraph, UserSettings 스키마 정의

### Mixin 적용 규칙

| 엔티티 | BaseMixin | SoftDeleteMixin | 비고 |
|--------|-----------|-----------------|------|
| PageVisit | ✅ | ❌ | 이벤트 데이터, soft delete 불필요 |
| Highlight | ✅ | ✅ | 사용자 생성 데이터, soft delete 필요 |
| RawEvent | ✅ | ❌ | 이벤트 로그, 삭제 불가 |
| MindmapGraph | ✅ | ✅ | AI 생성 결과물 |
| UserSettings | ✅ | ❌ | 1:1 관계, User와 함께 관리 |

### 체크리스트

- [ ] **PageVisit 스키마** (BaseMixin 적용)
  - [ ] `ent/schema/pagevisit.go`

    ```go
    package schema

    import (
        "time"

        "entgo.io/ent"
        "entgo.io/ent/schema/edge"
        "entgo.io/ent/schema/field"
        "entgo.io/ent/schema/index"

        "github.com/mindhit/api/ent/schema/mixin"
    )

    type PageVisit struct {
        ent.Schema
    }

    func (PageVisit) Mixin() []ent.Mixin {
        return []ent.Mixin{
            mixin.BaseMixin{},
        }
    }

    func (PageVisit) Fields() []ent.Field {
        return []ent.Field{
            field.Time("entered_at").
                Default(time.Now).
                Comment("Page entry time"),
            field.Time("left_at").
                Optional().
                Nillable().
                Comment("Page leave time"),
            field.Int("dwell_time_seconds").
                Optional().
                Nillable().
                Comment("Total time spent on page"),
            field.Float("max_scroll_depth").
                Default(0).
                Comment("Maximum scroll depth (0-1)"),
        }
    }

    func (PageVisit) Edges() []ent.Edge {
        return []ent.Edge{
            edge.From("session", Session.Type).
                Ref("page_visits").
                Unique().
                Required(),
            edge.To("url", URL.Type).
                Unique().
                Required(),
        }
    }

    func (PageVisit) Indexes() []ent.Index {
        return []ent.Index{
            index.Fields("entered_at"),
        }
    }
    ```

- [ ] **Highlight 스키마** (BaseMixin + SoftDeleteMixin 적용)
  - [ ] `ent/schema/highlight.go`

    ```go
    package schema

    import (
        "entgo.io/ent"
        "entgo.io/ent/schema/edge"
        "entgo.io/ent/schema/field"
        "entgo.io/ent/schema/index"

        "github.com/mindhit/api/ent/schema/mixin"
    )

    type Highlight struct {
        ent.Schema
    }

    func (Highlight) Mixin() []ent.Mixin {
        return []ent.Mixin{
            mixin.BaseMixin{},
            mixin.SoftDeleteMixin{},
        }
    }

    func (Highlight) Fields() []ent.Field {
        return []ent.Field{
            field.Text("text").
                NotEmpty().
                Comment("Highlighted text content"),
            field.String("selector").
                Optional().
                Comment("CSS selector for highlight position"),
            field.String("color").
                Default("#FFFF00").
                Comment("Highlight color (hex)"),
            field.String("note").
                Optional().
                Comment("User note for this highlight"),
        }
    }

    func (Highlight) Edges() []ent.Edge {
        return []ent.Edge{
            edge.From("session", Session.Type).
                Ref("highlights").
                Unique().
                Required(),
            edge.To("page_visit", PageVisit.Type).
                Unique(),
        }
    }

    func (Highlight) Indexes() []ent.Index {
        return []ent.Index{
            index.Fields("status"),
        }
    }
    ```

- [ ] **RawEvent 스키마** (BaseMixin 적용)
  - [ ] `ent/schema/rawevent.go`

    ```go
    package schema

    import (
        "entgo.io/ent"
        "entgo.io/ent/schema/edge"
        "entgo.io/ent/schema/field"
        "entgo.io/ent/schema/index"

        "github.com/mindhit/api/ent/schema/mixin"
    )

    type RawEvent struct {
        ent.Schema
    }

    func (RawEvent) Mixin() []ent.Mixin {
        return []ent.Mixin{
            mixin.BaseMixin{},
        }
    }

    func (RawEvent) Fields() []ent.Field {
        return []ent.Field{
            field.String("event_type").
                NotEmpty().
                Comment("Event type (page_visit, highlight, scroll, etc.)"),
            field.Time("timestamp").
                Comment("Client-side event timestamp"),
            field.Text("payload").
                Comment("Raw JSON event payload"),
            field.Bool("processed").
                Default(false).
                Comment("Whether event has been processed"),
            field.Time("processed_at").
                Optional().
                Nillable().
                Comment("When event was processed"),
        }
    }

    func (RawEvent) Edges() []ent.Edge {
        return []ent.Edge{
            edge.From("session", Session.Type).
                Ref("raw_events").
                Unique().
                Required(),
        }
    }

    func (RawEvent) Indexes() []ent.Index {
        return []ent.Index{
            index.Fields("event_type"),
            index.Fields("processed"),
            index.Fields("timestamp"),
        }
    }
    ```

- [ ] **MindmapGraph 스키마** (BaseMixin + SoftDeleteMixin 적용)
  - [ ] `ent/schema/mindmapgraph.go`

    ```go
    package schema

    import (
        "time"

        "entgo.io/ent"
        "entgo.io/ent/schema/edge"
        "entgo.io/ent/schema/field"
        "entgo.io/ent/schema/index"

        "github.com/mindhit/api/ent/schema/mixin"
    )

    type MindmapGraph struct {
        ent.Schema
    }

    func (MindmapGraph) Mixin() []ent.Mixin {
        return []ent.Mixin{
            mixin.BaseMixin{},
            mixin.SoftDeleteMixin{},
        }
    }

    func (MindmapGraph) Fields() []ent.Field {
        return []ent.Field{
            field.JSON("nodes", []map[string]interface{}{}).
                Optional().
                Comment("Mindmap node data"),
            field.JSON("edges", []map[string]interface{}{}).
                Optional().
                Comment("Mindmap edge data"),
            field.JSON("layout", map[string]interface{}{}).
                Optional().
                Comment("Layout configuration"),
            field.Time("generated_at").
                Default(time.Now).
                Comment("AI generation timestamp"),
            field.Int("version").
                Default(1).
                Comment("Mindmap version for regeneration tracking"),
        }
    }

    func (MindmapGraph) Edges() []ent.Edge {
        return []ent.Edge{
            edge.From("session", Session.Type).
                Ref("mindmap").
                Unique().
                Required(),
        }
    }

    func (MindmapGraph) Indexes() []ent.Index {
        return []ent.Index{
            index.Fields("status"),
        }
    }
    ```

- [ ] **UserSettings 스키마** (BaseMixin 적용)
  - [ ] `ent/schema/usersettings.go`

    ```go
    package schema

    import (
        "entgo.io/ent"
        "entgo.io/ent/schema/edge"
        "entgo.io/ent/schema/field"

        "github.com/mindhit/api/ent/schema/mixin"
    )

    type UserSettings struct {
        ent.Schema
    }

    func (UserSettings) Mixin() []ent.Mixin {
        return []ent.Mixin{
            mixin.BaseMixin{},
        }
    }

    func (UserSettings) Fields() []ent.Field {
        return []ent.Field{
            field.Enum("theme").
                Values("light", "dark", "system").
                Default("system").
                Comment("UI theme preference"),
            field.Bool("email_notifications").
                Default(true).
                Comment("Email notification preference"),
            field.Bool("browser_notifications").
                Default(true).
                Comment("Browser notification preference"),
            field.String("language").
                Default("ko").
                Comment("Preferred language"),
            field.Int("session_timeout_minutes").
                Default(60).
                Comment("Auto-stop session after inactivity"),
            field.Bool("auto_summarize").
                Default(true).
                Comment("Auto-generate summary when session ends"),
            field.JSON("extension_settings", map[string]interface{}{}).
                Optional().
                Comment("Chrome extension specific settings"),
        }
    }

    func (UserSettings) Edges() []ent.Edge {
        return []ent.Edge{
            edge.From("user", User.Type).
                Ref("settings").
                Unique().
                Required(),
        }
    }
    ```

> **Note**: User 스키마의 settings edge는 Step 1.7에서 이미 정의되어 있습니다.

- [ ] **코드 생성**

  ```bash
  cd apps/api
  go generate ./ent
  ```

### 검증

```bash
cd apps/api
go generate ./ent
# 에러 없이 완료

ls ent/
# client.go, ent.go, user.go, session.go, url.go, pagevisit.go, highlight.go, rawevent.go, mindmapgraph.go 등
```

### 결과물

```
apps/api/ent/
├── schema/
│   ├── mixin/
│   │   ├── base.go           # BaseMixin (id, created_at, updated_at)
│   │   └── soft_delete.go    # SoftDeleteMixin (status, deleted_at)
│   ├── user.go
│   ├── usersettings.go
│   ├── session.go
│   ├── url.go
│   ├── pagevisit.go
│   ├── highlight.go
│   ├── rawevent.go
│   └── mindmapgraph.go
├── client.go
├── user.go
├── session.go
└── ... (생성된 파일들)
```

### Soft Delete 쿼리 가이드

Soft delete를 사용하는 엔티티는 조회 시 `status` 필터링이 필요합니다.

#### 기본 쿼리 패턴

```go
// ❌ 잘못된 방법 - 삭제된 데이터도 포함됨
users, err := client.User.Query().All(ctx)

// ✅ 올바른 방법 - 활성 데이터만 조회
users, err := client.User.Query().
    Where(user.StatusEQ("active")).
    All(ctx)

// ✅ 삭제된 데이터만 조회 (복구/관리용)
deletedUsers, err := client.User.Query().
    Where(user.StatusEQ("inactive")).
    All(ctx)

// ✅ 모든 데이터 조회 (관리자 전용)
allUsers, err := client.User.Query().All(ctx)
```

#### Soft Delete 수행

```go
import "time"

// Soft delete 수행
now := time.Now()
_, err := client.User.UpdateOneID(userID).
    SetStatus("inactive").
    SetDeletedAt(now).
    Save(ctx)

// 복구
_, err := client.User.UpdateOneID(userID).
    SetStatus("active").
    ClearDeletedAt().
    Save(ctx)
```

#### 엔티티별 쿼리 예시

```go
// Session: session_status(워크플로우)와 status(soft delete) 구분
activeSessions, err := client.Session.Query().
    Where(
        session.StatusEQ("active"),                    // soft delete 필터
        session.SessionStatusEQ("recording"),          // 워크플로우 필터
    ).
    All(ctx)

// URL: 활성 URL만 조회
urls, err := client.URL.Query().
    Where(url.StatusEQ("active")).
    All(ctx)

// Highlight: 활성 하이라이트만 조회
highlights, err := client.Highlight.Query().
    Where(highlight.StatusEQ("active")).
    All(ctx)

// PageVisit, RawEvent: soft delete 없음 (필터 불필요)
pageVisits, err := client.PageVisit.Query().All(ctx)
rawEvents, err := client.RawEvent.Query().All(ctx)
```

#### Edge 쿼리 시 주의사항

```go
// User의 활성 세션만 조회
user, err := client.User.Query().
    Where(user.IDEQ(userID)).
    WithSessions(func(q *ent.SessionQuery) {
        q.Where(session.StatusEQ("active"))
    }).
    Only(ctx)

// Session의 활성 하이라이트만 조회
sess, err := client.Session.Query().
    Where(session.IDEQ(sessionID)).
    WithHighlights(func(q *ent.HighlightQuery) {
        q.Where(highlight.StatusEQ("active"))
    }).
    Only(ctx)
```

#### 헬퍼 함수 권장

서비스 레이어에서 반복을 줄이기 위한 헬퍼 패턴:

```go
// internal/service/helpers.go

// ActiveUsers returns query for active users only
func (s *UserService) ActiveUsers() *ent.UserQuery {
    return s.client.User.Query().Where(user.StatusEQ("active"))
}

// ActiveSessions returns query for active sessions only
func (s *SessionService) ActiveSessions() *ent.SessionQuery {
    return s.client.Session.Query().Where(session.StatusEQ("active"))
}

// 사용 예시
users, err := s.ActiveUsers().
    Where(user.EmailContains("@example.com")).
    All(ctx)
```

> **중요**: 모든 조회 쿼리에서 soft delete 필터를 적용하는 것을 잊지 마세요.
> 향후 Ent Interceptor를 사용하여 자동 필터링을 구현할 수 있습니다.

---

## Step 1.9: 첫 번째 Migration 생성 및 적용

### 목표

Atlas로 초기 migration 생성 및 PostgreSQL에 적용

### 체크리스트

- [ ] **PostgreSQL 실행 확인**

  ```bash
  docker-compose up -d postgres
  docker exec -it mindhit-postgres psql -U postgres -c "SELECT 1;"
  ```

- [ ] **PostgreSQL 드라이버 추가**

  ```bash
  cd apps/api
  go get github.com/lib/pq
  go get ariga.io/atlas-provider-ent
  ```

- [ ] **Migration 생성**

  ```bash
  cd apps/api
  atlas migrate diff initial_schema \
    --dir "file://ent/migrate/migrations" \
    --to "ent://ent/schema" \
    --dev-url "postgres://postgres:password@localhost:5432/mindhit_dev?sslmode=disable"
  ```

- [ ] **생성된 SQL 확인**

  ```bash
  cat ent/migrate/migrations/*.sql
  ```

- [ ] **Migration 적용**

  ```bash
  atlas migrate apply \
    --dir "file://ent/migrate/migrations" \
    --url "postgres://postgres:password@localhost:5432/mindhit?sslmode=disable"
  ```

- [ ] **테이블 확인**

  ```bash
  docker exec -it mindhit-postgres psql -U postgres -d mindhit -c "\dt"
  ```

### 검증

```bash
# 테이블 목록 확인
docker exec -it mindhit-postgres psql -U postgres -d mindhit -c "\dt"

# 예상 출력:
#          List of relations
#  Schema |      Name       | Type  |  Owner
# --------+-----------------+-------+----------
#  public | atlas_schema_revisions | table | postgres
#  public | highlights      | table | postgres
#  public | mindmap_graphs  | table | postgres
#  public | page_visits     | table | postgres
#  public | raw_events      | table | postgres
#  public | sessions        | table | postgres
#  public | urls            | table | postgres
#  public | users           | table | postgres
```

### 결과물

```
apps/api/ent/migrate/migrations/
├── 20241221000000_initial_schema.sql
└── atlas.sum
```

### 마이그레이션 전략

#### 원칙

1. **Backward-compatible 변경 우선**: 가능하면 롤백 가능한 마이그레이션 작성
2. **Phase별 마이그레이션 분리**: 각 Phase의 변경사항을 별도 마이그레이션으로 생성
3. **CI에서 마이그레이션 검증**: PR 시 마이그레이션 dry-run 검증

#### 마이그레이션 파일 네이밍

```
{YYYYMMDD}_{sequence}_{phase}_{description}.sql
```

예시:

- `20241221_001_phase1_initial_schema.sql`
- `20241222_001_phase2_add_oauth_fields.sql`
- `20241223_001_phase13_add_billing_tables.sql`

#### 안전한 스키마 변경 가이드

| 변경 유형 | 안전한 방법 | 피해야 할 방법 |
| --------- | ----------- | -------------- |
| 컬럼 추가 | `ADD COLUMN ... DEFAULT` | `ADD COLUMN ... NOT NULL` (기본값 없이) |
| 컬럼 삭제 | 2단계: 코드 제거 → 컬럼 삭제 | 즉시 삭제 |
| 컬럼 이름 변경 | 새 컬럼 추가 → 데이터 복사 → 이전 컬럼 삭제 | `RENAME COLUMN` |
| 타입 변경 | 호환 가능한 타입으로만 | 데이터 손실 가능한 변경 |
| 인덱스 추가 | `CREATE INDEX CONCURRENTLY` | `CREATE INDEX` (락 발생) |

#### 롤백 계획

각 마이그레이션에 대해 롤백 SQL을 문서화:

```sql
-- Migration: 20241222_001_phase2_add_oauth_fields.sql
-- Rollback:
-- ALTER TABLE users DROP COLUMN IF EXISTS oauth_provider;
-- ALTER TABLE users DROP COLUMN IF EXISTS oauth_id;
```

---

## Step 1.10: 테스트 인프라 설정

### 목표

테스트 헬퍼, fixture, CI 통합을 통해 지속적인 테스트 기반 마련

### 체크리스트

- [ ] **testutil 패키지 생성**
  - [ ] `internal/testutil/db.go`

    ```go
    package testutil

    import (
        "context"
        "testing"

        "entgo.io/ent/dialect"
        _ "github.com/lib/pq"

        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/enttest"
    )

    // SetupTestDB creates a test database client
    // Uses SQLite in-memory for fast unit tests
    func SetupTestDB(t *testing.T) *ent.Client {
        t.Helper()
        client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&_fk=1")
        return client
    }

    // SetupTestDBWithPostgres creates a test client with PostgreSQL
    // Use for integration tests that need PostgreSQL-specific features
    func SetupTestDBWithPostgres(t *testing.T, dsn string) *ent.Client {
        t.Helper()
        client, err := ent.Open(dialect.Postgres, dsn)
        if err != nil {
            t.Fatalf("failed to open postgres: %v", err)
        }
        // Run migrations
        if err := client.Schema.Create(context.Background()); err != nil {
            t.Fatalf("failed to create schema: %v", err)
        }
        return client
    }

    // CleanupTestDB closes the test database client
    func CleanupTestDB(t *testing.T, client *ent.Client) {
        t.Helper()
        if err := client.Close(); err != nil {
            t.Errorf("failed to close client: %v", err)
        }
    }
    ```

- [ ] **fixture 패키지 생성**
  - [ ] `internal/testutil/fixture/user.go`

    ```go
    package fixture

    import (
        "context"
        "testing"

        "github.com/mindhit/api/ent"
    )

    // CreateTestUser creates a user for testing
    func CreateTestUser(t *testing.T, client *ent.Client, email string) *ent.User {
        t.Helper()
        user, err := client.User.Create().
            SetEmail(email).
            SetPasswordHash("$2a$10$testhashedpassword").
            Save(context.Background())
        if err != nil {
            t.Fatalf("failed to create test user: %v", err)
        }
        return user
    }

    // CreateTestSession creates a session for testing
    func CreateTestSession(t *testing.T, client *ent.Client, user *ent.User) *ent.Session {
        t.Helper()
        session, err := client.Session.Create().
            SetUser(user).
            Save(context.Background())
        if err != nil {
            t.Fatalf("failed to create test session: %v", err)
        }
        return session
    }
    ```

- [ ] **SQLite 의존성 추가**

  ```bash
  cd apps/api
  go get github.com/mattn/go-sqlite3
  go get entgo.io/ent/dialect/sql
  ```

- [ ] **예제 단위 테스트 작성**
  - [ ] `internal/service/user_test.go`

    ```go
    package service_test

    import (
        "context"
        "testing"

        "github.com/mindhit/api/ent/user"
        "github.com/mindhit/api/internal/testutil"
    )

    func TestUserCreate(t *testing.T) {
        client := testutil.SetupTestDB(t)
        defer testutil.CleanupTestDB(t, client)

        ctx := context.Background()

        // Create user
        u, err := client.User.Create().
            SetEmail("test@example.com").
            SetPasswordHash("hashedpassword").
            Save(ctx)
        if err != nil {
            t.Fatalf("failed to create user: %v", err)
        }

        // Verify
        if u.Email != "test@example.com" {
            t.Errorf("expected email test@example.com, got %s", u.Email)
        }
        if u.Status != user.StatusActive {
            t.Errorf("expected status active, got %s", u.Status)
        }
    }

    func TestUserSoftDelete(t *testing.T) {
        client := testutil.SetupTestDB(t)
        defer testutil.CleanupTestDB(t, client)

        ctx := context.Background()

        // Create user
        u, err := client.User.Create().
            SetEmail("delete@example.com").
            SetPasswordHash("hashedpassword").
            Save(ctx)
        if err != nil {
            t.Fatalf("failed to create user: %v", err)
        }

        // Soft delete
        _, err = client.User.UpdateOneID(u.ID).
            SetStatus(user.StatusInactive).
            Save(ctx)
        if err != nil {
            t.Fatalf("failed to soft delete user: %v", err)
        }

        // Query active users - should not find deleted user
        activeUsers, err := client.User.Query().
            Where(user.StatusEQ(user.StatusActive)).
            All(ctx)
        if err != nil {
            t.Fatalf("failed to query users: %v", err)
        }
        if len(activeUsers) != 0 {
            t.Errorf("expected 0 active users, got %d", len(activeUsers))
        }
    }
    ```

- [ ] **moon.yml에 테스트 태스크 확인**

  ```yaml
  # apps/backend/moon.yml
  test:
    command: go test -v -race -coverprofile=coverage.out ./...
    inputs:
      - "**/*.go"

  test-coverage:
    command: bash
    args: [-c, "go test -v -race -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html"]
    deps: [test]

  test-short:
    command: go test -v -short ./...
    inputs:
      - "**/*.go"
  ```

- [ ] **CI 워크플로우에 테스트 추가**
  - [ ] `.github/workflows/ci.yml` 업데이트

    ```yaml
    # backend 테스트 job 추가
    test-backend:
      runs-on: ubuntu-latest
      services:
        postgres:
          image: postgres:16
          env:
            POSTGRES_USER: postgres
            POSTGRES_PASSWORD: password
            POSTGRES_DB: mindhit_test
          ports:
            - 5432:5432
          options: >-
            --health-cmd pg_isready
            --health-interval 10s
            --health-timeout 5s
            --health-retries 5
      steps:
        - uses: actions/checkout@v4
        - uses: actions/setup-go@v5
          with:
            go-version: '1.22'
        - name: Run tests
          run: moon run backend:test
          env:
            DATABASE_URL: postgres://postgres:password@localhost:5432/mindhit_test?sslmode=disable
    ```

### 검증

```bash
# 단위 테스트 실행
cd apps/api
go test -v ./internal/service/...

# 전체 테스트 실행
moon run backend:test

# 커버리지 리포트
moon run backend:test-coverage
open coverage.html
```

### 결과물

```
apps/api/
├── internal/
│   ├── testutil/
│   │   ├── db.go              # 테스트 DB 헬퍼
│   │   └── fixture/
│   │       └── user.go        # 테스트 fixture
│   └── service/
│       └── user_test.go       # 예제 테스트
└── coverage.out               # (생성됨)
```

---

## Phase 1 완료 확인

### 전체 검증 체크리스트

- [ ] **모노레포 구조**

  ```bash
  ls -la apps/ packages/
  ```

- [ ] **Go 서버 실행**

  ```bash
  cd apps/api && go run ./cmd/server
  curl http://localhost:8080/health
  ```

- [ ] **Ent 코드 생성**

  ```bash
  cd apps/api && go generate ./ent
  ```

- [ ] **PostgreSQL 테이블**

  ```bash
  docker exec -it mindhit-postgres psql -U postgres -d mindhit -c "\dt"
  ```

- [ ] **테스트 통과**

  ```bash
  moon run backend:test
  # 모든 테스트 PASS
  ```

### 테스트 요구사항

| 테스트 유형 | 대상 | 최소 커버리지 |
| ----------- | ---- | ------------- |
| 단위 테스트 | Ent 스키마 CRUD | 기본 동작 검증 |
| 단위 테스트 | Soft delete 동작 | status 필터링 검증 |

> **Note**: Phase 1에서는 테스트 인프라 설정이 목표입니다.
> 이후 Phase에서 각 기능에 대한 테스트가 점진적으로 추가됩니다.

### 산출물 요약

| 항목 | 위치 |
|-----|------|
| 모노레포 설정 | `pnpm-workspace.yaml`, `.moon/` |
| Go 프로젝트 | `apps/backend/` |
| Ent 스키마 | `apps/backend/pkg/ent/schema/` |
| Migration | `apps/backend/pkg/ent/migrate/migrations/` |
| Docker Compose | `infra/docker/docker-compose.yml` |

---

## 다음 Phase

Phase 1 완료 후 [Phase 1.5: API 스펙 공통화](./phase-1.5-api-spec.md)로 진행하세요.
