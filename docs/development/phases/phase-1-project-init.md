# Phase 1: 프로젝트 초기화

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | 모노레포 구조 설정 및 Go 백엔드 기반 구축 |
| **선행 조건** | 없음 (첫 번째 Phase) |
| **예상 소요** | 9 Steps |
| **결과물** | 동작하는 Go 서버 + PostgreSQL 연결 |

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

---

## Step 1.1: 모노레포 구조 설정

### 목표
pnpm workspace + moon 기반 모노레포 구조 생성

### 체크리스트

- [ ] **루트 디렉토리 설정**
  - [ ] `pnpm-workspace.yaml` 생성
    ```yaml
    packages:
      - 'apps/*'
      - 'packages/*'
    ```
  - [ ] 루트 `package.json` 생성
    ```json
    {
      "name": "mindhit",
      "private": true,
      "scripts": {
        "dev": "moon run :dev",
        "build": "moon run :build",
        "test": "moon run :test",
        "lint": "moon run :lint",
        "generate": "moon run :generate"
      },
      "devDependencies": {
        "@moonrepo/cli": "^1.28.0"
      }
    }
    ```

- [ ] **디렉토리 구조 생성**
  ```bash
  mkdir -p apps/{api,web,extension}
  mkdir -p packages/{protocol,shared}
  ```

- [ ] **moon 설정**
  - [ ] `.moon/workspace.yml` 생성
    ```yaml
    $schema: 'https://moonrepo.dev/schemas/workspace.json'
    projects:
      - 'apps/*'
      - 'packages/*'
    vcs:
      manager: git
      defaultBranch: main
    ```
  - [ ] `.moon/toolchain.yml` 생성
    ```yaml
    $schema: 'https://moonrepo.dev/schemas/toolchain.json'
    node:
      version: '20.10.0'
      packageManager: pnpm
    ```

- [ ] **Git 설정**
  - [ ] `.gitignore` 업데이트
    ```
    node_modules/
    .moon/cache/
    .moon/docker/
    dist/
    build/
    .env
    .env.local
    ```

### 검증
```bash
# pnpm workspace 확인
pnpm install

# moon 동작 확인
moon --version
```

### 결과물
```
mindhit/
├── .moon/
│   ├── workspace.yml
│   └── toolchain.yml
├── apps/
│   ├── api/
│   ├── web/
│   └── extension/
├── packages/
│   ├── protocol/
│   └── shared/
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
로컬 개발용 PostgreSQL 컨테이너 설정

### 체크리스트

- [ ] **docker-compose.yml 생성** (루트)
  ```yaml
  version: '3.8'

  services:
    postgres:
      image: postgres:16-alpine
      container_name: mindhit-postgres
      environment:
        POSTGRES_USER: postgres
        POSTGRES_PASSWORD: password
        POSTGRES_DB: mindhit
      ports:
        - "5432:5432"
      volumes:
        - postgres_data:/var/lib/postgresql/data
        - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql
      healthcheck:
        test: ["CMD-SHELL", "pg_isready -U postgres"]
        interval: 5s
        timeout: 5s
        retries: 5

    redis:
      image: redis:7-alpine
      container_name: mindhit-redis
      ports:
        - "6379:6379"
      volumes:
        - redis_data:/data

  volumes:
    postgres_data:
    redis_data:
  ```

- [ ] **초기화 스크립트 생성**
  - [ ] `scripts/init-db.sql`
    ```sql
    -- 개발용 데이터베이스 생성
    CREATE DATABASE mindhit_dev;
    CREATE DATABASE mindhit_test;
    ```

- [ ] **환경 변수 파일 생성**
  - [ ] `apps/api/.env.example`
    ```
    PORT=8080
    ENVIRONMENT=development
    DATABASE_URL=postgres://postgres:password@localhost:5432/mindhit?sslmode=disable
    DEV_DATABASE_URL=postgres://postgres:password@localhost:5432/mindhit_dev?sslmode=disable
    JWT_SECRET=your-secret-key-change-in-production
    REDIS_URL=redis://localhost:6379
    ```
  - [ ] `apps/api/.env` (복사 후 수정)

### 검증
```bash
# 컨테이너 실행
docker-compose up -d

# 연결 테스트
docker exec -it mindhit-postgres psql -U postgres -d mindhit -c "SELECT 1;"

# 데이터베이스 목록 확인
docker exec -it mindhit-postgres psql -U postgres -c "\l"
```

### 결과물
```
mindhit/
├── docker-compose.yml
├── scripts/
│   └── init-db.sql
└── apps/api/
    ├── .env.example
    └── .env
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

### 산출물 요약

| 항목 | 위치 |
|-----|------|
| 모노레포 설정 | `pnpm-workspace.yaml`, `.moon/` |
| Go 프로젝트 | `apps/api/` |
| Ent 스키마 | `apps/api/ent/schema/` |
| Migration | `apps/api/ent/migrate/migrations/` |
| Docker | `docker-compose.yml` |

---

## 다음 Phase

Phase 1 완료 후 [Phase 1.5: API 스펙 공통화](./phase-1.5-api-spec.md)로 진행하세요.
