# Phase 0: 개발 환경 구성

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | Moon + Docker 기반 로컬 개발 환경 구성 |
| **선행 조건** | Node.js, pnpm, Docker, Docker Compose 설치 |
| **예상 소요** | 4 Steps |
| **결과물** | Moon 태스크 러너 + PostgreSQL, Redis, API 서버가 동작 |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 0.1 | Moon 설치 및 초기 설정 | ⬜ |
| 0.2 | Docker Compose 구성 | ⬜ |
| 0.3 | 개발 환경 스크립트 | ⬜ |
| 0.4 | Moon CI 워크플로우 구성 | ⬜ |

---

## Step 0.1: Moon 설치 및 초기 설정

### 목표
Moon 태스크 러너 설치 및 모노레포 기본 설정

### Moon 소개

Moon은 모노레포를 위한 고성능 태스크 러너입니다.

| 특징 | 설명 |
|-----|------|
| **빠른 실행** | 변경된 파일만 감지하여 필요한 태스크만 실행 |
| **캐싱** | 이전 실행 결과를 캐싱하여 중복 작업 방지 |
| **의존성 관리** | 프로젝트 간 의존성 자동 감지 |
| **CI 최적화** | `moon ci` 명령으로 PR 변경 사항만 테스트 |

### 체크리스트

- [ ] **Node.js 및 pnpm 설치 확인**
  ```bash
  node --version  # v20.x 이상
  pnpm --version  # v8.x 이상
  ```

- [ ] **pnpm-workspace.yaml 생성** (루트)
  ```yaml
  packages:
    - 'apps/*'
    - 'packages/*'
  ```

- [ ] **루트 package.json 생성**
  ```json
  {
    "name": "mindhit",
    "private": true,
    "packageManager": "pnpm@9.0.0",
    "scripts": {
      "dev": "moon run :dev",
      "build": "moon run :build",
      "test": "moon run :test",
      "lint": "moon run :lint",
      "ci": "moon ci"
    },
    "devDependencies": {
      "@moonrepo/cli": "^1.29.0"
    }
  }
  ```

- [ ] **Moon 설치**
  ```bash
  pnpm install
  ```

- [ ] **Moon workspace 설정**
  - [ ] `.moon/workspace.yml` 생성
    ```yaml
    $schema: 'https://moonrepo.dev/schemas/workspace.json'

    # 프로젝트 위치 설정
    projects:
      - 'apps/*'
      - 'packages/*'

    # VCS 설정
    vcs:
      manager: git
      defaultBranch: main
      remoteCandidates:
        - origin

    # 버전 관리 및 제약
    versionConstraint: '>=1.28.0'

    # 실험적 기능 (선택)
    experiments:
      actionPipelineV2: true

    # 원격 캐싱 (선택, 향후 설정)
    # runner:
    #   cacheLifetime: '7 days'
    ```

- [ ] **Moon toolchain 설정**
  - [ ] `.moon/toolchain.yml` 생성
    ```yaml
    $schema: 'https://moonrepo.dev/schemas/toolchain.json'

    # Node.js 설정
    node:
      version: '20.10.0'
      packageManager: pnpm
      pnpm:
        version: '9.0.0'

    # TypeScript 설정 (웹앱/익스텐션용)
    typescript:
      syncProjectReferences: true
    ```

- [ ] **Moon tasks 설정 (전역)**
  - [ ] `.moon/tasks.yml` 생성
    ```yaml
    $schema: 'https://moonrepo.dev/schemas/tasks.json'

    # 전역 태스크 설정
    fileGroups:
      sources:
        - 'src/**/*'
        - 'lib/**/*'
      tests:
        - 'test/**/*'
        - '**/*.test.*'
      configs:
        - '*.config.*'
        - 'tsconfig.json'

    # 환경별 설정
    implicitDeps:
      - '^:build'

    # CI에서 실행하지 않을 태스크
    implicitInputs:
      - '/.moon/*.yml'
    ```

- [ ] **디렉토리 구조 생성**
  ```bash
  mkdir -p apps/{api,web,extension}
  mkdir -p packages/{shared,protocol}
  mkdir -p .github/workflows
  ```

### 검증
```bash
# Moon 버전 확인
pnpm moon --version

# 프로젝트 목록 확인
pnpm moon project-graph

# 설정 검증
pnpm moon query projects
```

### 결과물

```text
mindhit/
├── .moon/
│   ├── workspace.yml     # 워크스페이스 설정
│   ├── toolchain.yml     # 도구 버전 설정
│   └── tasks.yml         # 전역 태스크 설정
├── apps/
│   ├── api/              # Go 백엔드
│   ├── web/              # Next.js 웹앱
│   └── extension/        # Chrome Extension
├── packages/
│   ├── shared/           # 공유 유틸
│   └── protocol/         # API 타입 정의
├── package.json
├── pnpm-workspace.yaml
└── pnpm-lock.yaml
```

---

## Step 0.2: Docker Compose 구성

### 목표
PostgreSQL, Redis, API 서버를 Docker Compose로 구성

### 체크리스트

- [ ] **docker-compose.yml 생성** (루트)
  ```yaml
  version: '3.8'

  services:
    # PostgreSQL 데이터베이스
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
      networks:
        - mindhit-network

    # Redis 캐시/세션
    redis:
      image: redis:7-alpine
      container_name: mindhit-redis
      ports:
        - "6379:6379"
      volumes:
        - redis_data:/data
      command: redis-server --appendonly yes
      healthcheck:
        test: ["CMD", "redis-cli", "ping"]
        interval: 5s
        timeout: 5s
        retries: 5
      networks:
        - mindhit-network

    # API 서버 (개발용)
    api:
      build:
        context: ./apps/api
        dockerfile: Dockerfile.dev
      container_name: mindhit-api
      ports:
        - "8080:8080"
      environment:
        - PORT=8080
        - ENVIRONMENT=development
        - DATABASE_URL=postgres://postgres:password@postgres:5432/mindhit?sslmode=disable
        - DEV_DATABASE_URL=postgres://postgres:password@postgres:5432/mindhit_dev?sslmode=disable
        - REDIS_URL=redis://redis:6379
        - JWT_SECRET=dev-secret-key-change-in-production
      volumes:
        - ./apps/api:/app
        - go_mod_cache:/go/pkg/mod
      depends_on:
        postgres:
          condition: service_healthy
        redis:
          condition: service_healthy
      networks:
        - mindhit-network

  volumes:
    postgres_data:
    redis_data:
    go_mod_cache:

  networks:
    mindhit-network:
      driver: bridge
  ```

- [ ] **docker-compose.override.yml 생성** (로컬 개발용)
  ```yaml
  version: '3.8'

  # 로컬 개발 시 API 서버 없이 DB만 실행할 때 사용
  # docker-compose up postgres redis

  services:
    api:
      profiles:
        - full  # docker-compose --profile full up
  ```

- [ ] **API Dockerfile.dev 생성**
  - [ ] `apps/api/Dockerfile.dev`
    ```dockerfile
    FROM golang:1.22-alpine

    # 필수 패키지 설치
    RUN apk add --no-cache git curl

    # Air (Hot reload) 설치
    RUN go install github.com/air-verse/air@latest

    WORKDIR /app

    # 의존성 먼저 복사 (캐시 활용)
    COPY go.mod go.sum ./
    RUN go mod download

    # 소스 복사
    COPY . .

    # Air 설정 파일이 없으면 기본 설정 사용
    CMD ["air", "-c", ".air.toml"]
    ```

- [ ] **Air 설정 파일 생성**
  - [ ] `apps/api/.air.toml`
    ```toml
    root = "."
    tmp_dir = "tmp"

    [build]
    cmd = "go build -o ./tmp/main ./cmd/server"
    bin = "tmp/main"
    full_bin = "./tmp/main"
    include_ext = ["go", "tpl", "tmpl", "html"]
    exclude_dir = ["assets", "tmp", "vendor", "ent/migrate"]
    exclude_regex = ["_test.go"]
    delay = 1000  # ms
    stop_on_error = true
    log = "air.log"

    [log]
    time = true

    [color]
    main = "magenta"
    watcher = "cyan"
    build = "yellow"
    runner = "green"

    [misc]
    clean_on_exit = true
    ```

- [ ] **초기화 스크립트 생성**
  - [ ] `scripts/init-db.sql`
    ```sql
    -- 개발용 데이터베이스 생성
    CREATE DATABASE mindhit_dev;
    CREATE DATABASE mindhit_test;

    -- 권한 부여
    GRANT ALL PRIVILEGES ON DATABASE mindhit_dev TO postgres;
    GRANT ALL PRIVILEGES ON DATABASE mindhit_test TO postgres;
    ```

- [ ] **.dockerignore 생성**
  - [ ] `apps/api/.dockerignore`
    ```
    .git
    .gitignore
    .env
    .env.local
    tmp/
    bin/
    *.md
    Dockerfile*
    docker-compose*
    .air.toml
    ```

### 검증
```bash
# 전체 스택 실행
docker-compose up -d

# 상태 확인
docker-compose ps

# 로그 확인
docker-compose logs -f api

# PostgreSQL 연결 테스트
docker exec -it mindhit-postgres psql -U postgres -d mindhit -c "SELECT 1;"

# Redis 연결 테스트
docker exec -it mindhit-redis redis-cli ping
```

---

## Step 0.3: 개발 환경 스크립트

### 목표
개발 환경 관리를 위한 편의 스크립트 생성 (Moon 통합)

### 체크리스트

- [ ] **Makefile 생성** (루트) - Moon 명령어와 Docker 통합
  ```makefile
  .PHONY: help dev dev-db dev-full down logs clean migrate test lint build ci

  # 기본 도움말
  help:
  	@echo "MindHit Development Commands"
  	@echo ""
  	@echo "환경 관리:"
  	@echo "  make dev-db      - PostgreSQL, Redis만 실행"
  	@echo "  make dev-full    - 전체 스택 실행 (API 포함)"
  	@echo "  make down        - 모든 컨테이너 중지"
  	@echo "  make logs        - 로그 보기"
  	@echo "  make clean       - 볼륨 포함 완전 삭제"
  	@echo ""
  	@echo "Moon 태스크:"
  	@echo "  make build       - 모든 프로젝트 빌드 (moon run :build)"
  	@echo "  make test        - 모든 프로젝트 테스트 (moon run :test)"
  	@echo "  make lint        - 모든 프로젝트 린트 (moon run :lint)"
  	@echo "  make ci          - CI 파이프라인 실행 (moon ci)"
  	@echo ""
  	@echo "개발:"
  	@echo "  make api         - API 서버 로컬 실행 (Hot reload)"
  	@echo "  make migrate     - 마이그레이션 적용"
  	@echo "  make migrate-new - 새 마이그레이션 생성"
  	@echo "  make generate    - Ent 코드 생성"

  # ===================
  # Docker 환경 관리
  # ===================

  # DB만 실행 (로컬 개발용)
  dev-db:
  	docker-compose up -d postgres redis
  	@echo "Waiting for PostgreSQL..."
  	@sleep 3
  	@docker exec mindhit-postgres pg_isready -U postgres
  	@echo "PostgreSQL is ready!"
  	@echo ""
  	@echo "Connection strings:"
  	@echo "  DATABASE_URL=postgres://postgres:password@localhost:5432/mindhit?sslmode=disable"
  	@echo "  REDIS_URL=redis://localhost:6379"

  # 전체 스택 실행
  dev-full:
  	docker-compose --profile full up -d
  	@echo "All services started!"

  # 컨테이너 중지
  down:
  	docker-compose down

  # 로그 확인
  logs:
  	docker-compose logs -f

  # 완전 삭제 (볼륨 포함)
  clean:
  	docker-compose down -v --remove-orphans
  	@echo "All containers and volumes removed!"

  # ===================
  # Moon 태스크
  # ===================

  # 모든 프로젝트 빌드
  build:
  	pnpm moon run :build

  # 모든 프로젝트 테스트
  test:
  	pnpm moon run :test

  # 모든 프로젝트 린트
  lint:
  	pnpm moon run :lint

  # CI 파이프라인 (변경된 프로젝트만)
  ci:
  	pnpm moon ci

  # ===================
  # API 개발
  # ===================

  # API 서버 로컬 실행 (Hot reload)
  api:
  	pnpm moon run api:dev

  # 마이그레이션 적용
  migrate:
  	pnpm moon run api:migrate-apply

  # 새 마이그레이션 생성
  migrate-new:
  	@read -p "Migration name: " name; \
  	cd apps/api && atlas migrate diff $$name \
  		--dir "file://ent/migrate/migrations" \
  		--to "ent://ent/schema" \
  		--dev-url "postgres://postgres:password@localhost:5432/mindhit_dev?sslmode=disable"

  # Ent 코드 생성
  generate:
  	pnpm moon run api:generate
  ```

- [ ] **환경 변수 파일 템플릿**
  - [ ] `apps/api/.env.example`
    ```env
    # Server
    PORT=8080
    ENVIRONMENT=development

    # Database
    DATABASE_URL=postgres://postgres:password@localhost:5432/mindhit?sslmode=disable
    DEV_DATABASE_URL=postgres://postgres:password@localhost:5432/mindhit_dev?sslmode=disable

    # Redis
    REDIS_URL=redis://localhost:6379

    # JWT
    JWT_SECRET=your-secret-key-change-in-production

    # OpenAI (Phase 9에서 사용)
    OPENAI_API_KEY=
    OPENAI_MODEL=gpt-4-turbo-preview
    ```

- [ ] **.env 파일 생성**
  ```bash
  cp apps/api/.env.example apps/api/.env
  ```

- [ ] **scripts/dev-setup.sh 생성**
  ```bash
  #!/bin/bash
  set -e

  echo "=== MindHit Development Setup ==="
  echo ""

  # 1. 환경 변수 파일 확인
  if [ ! -f "apps/api/.env" ]; then
      echo "Creating .env file..."
      cp apps/api/.env.example apps/api/.env
      echo "Please update apps/api/.env with your settings"
  fi

  # 2. Docker 확인
  if ! command -v docker &> /dev/null; then
      echo "ERROR: Docker is not installed"
      exit 1
  fi

  if ! command -v docker-compose &> /dev/null; then
      echo "ERROR: Docker Compose is not installed"
      exit 1
  fi

  # 3. Go 확인
  if ! command -v go &> /dev/null; then
      echo "ERROR: Go is not installed"
      exit 1
  fi

  echo "Go version: $(go version)"

  # 4. 의존성 설치
  echo ""
  echo "Installing Go dependencies..."
  cd apps/api
  go mod download
  cd ../..

  # 5. Air 설치 (Hot reload)
  echo ""
  echo "Installing Air for hot reload..."
  go install github.com/air-verse/air@latest

  # 6. golangci-lint 설치
  echo ""
  echo "Installing golangci-lint..."
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

  # 7. Atlas 설치 확인
  if ! command -v atlas &> /dev/null; then
      echo ""
      echo "Installing Atlas CLI..."
      curl -sSf https://atlasgo.sh | sh
  fi

  # 8. Docker 컨테이너 시작
  echo ""
  echo "Starting Docker containers..."
  docker-compose up -d postgres redis

  # 9. 대기
  echo ""
  echo "Waiting for PostgreSQL to be ready..."
  sleep 5

  # 10. 마이그레이션 실행
  echo ""
  echo "Running migrations..."
  cd apps/api
  atlas migrate apply \
      --dir "file://ent/migrate/migrations" \
      --url "postgres://postgres:password@localhost:5432/mindhit?sslmode=disable" || true
  cd ../..

  echo ""
  echo "=== Setup Complete! ==="
  echo ""
  echo "Next steps:"
  echo "  1. Start API server:  make api"
  echo "  2. Or run full stack: make dev-full"
  echo ""
  echo "Useful commands:"
  echo "  make help     - Show all commands"
  echo "  make logs     - View container logs"
  echo "  make test     - Run tests"
  ```

- [ ] **실행 권한 부여**
  ```bash
  chmod +x scripts/dev-setup.sh
  ```

### 검증
```bash
# 개발 환경 설정
./scripts/dev-setup.sh

# DB만 실행하고 로컬에서 API 개발
make dev-db
make api

# 또는 전체 Docker 스택
make dev-full
```

---

## Step 0.4: Moon CI 워크플로우 구성

### 목표
GitHub Actions에서 Moon CI를 활용한 자동화 파이프라인 구성

### Moon CI 핵심 개념

Moon CI는 PR에서 변경된 파일만 감지하여 필요한 태스크만 실행합니다.

```mermaid
flowchart LR
    A[PR 생성] --> B[moon ci 실행]
    B --> C{변경 감지}
    C --> D[apps/api 변경]
    C --> E[apps/web 변경]
    C --> F[packages/shared 변경]
    D --> G[api:test, api:lint]
    E --> H[web:test, web:lint]
    F --> I[shared 의존 프로젝트 모두 실행]
```

### 체크리스트

- [ ] **GitHub Actions 워크플로우 생성**
  - [ ] `.github/workflows/ci.yml`
    ```yaml
    name: CI

    on:
      push:
        branches:
          - main
      pull_request:
        branches:
          - main

    env:
      MOON_TOOLCHAIN_FORCE_GLOBALS: true

    jobs:
      ci:
        name: CI Pipeline
        runs-on: ubuntu-latest

        services:
          postgres:
            image: postgres:16-alpine
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

          redis:
            image: redis:7-alpine
            ports:
              - 6379:6379
            options: >-
              --health-cmd "redis-cli ping"
              --health-interval 10s
              --health-timeout 5s
              --health-retries 5

        steps:
          # 1. 체크아웃 (전체 히스토리 필요)
          - name: Checkout repository
            uses: actions/checkout@v4
            with:
              fetch-depth: 0

          # 2. Go 설치
          - name: Setup Go
            uses: actions/setup-go@v5
            with:
              go-version: '1.22'
              cache: true
              cache-dependency-path: apps/api/go.sum

          # 3. Moon toolchain 설치 (pnpm + node 자동 설치)
          - name: Setup Moon toolchain
            uses: moonrepo/setup-toolchain@v0
            with:
              auto-install: true

          # 4. pnpm 의존성 설치
          - name: Install dependencies
            run: pnpm install --frozen-lockfile

          # 5. Go 의존성 설치
          - name: Install Go dependencies
            run: |
              cd apps/api
              go mod download

          # 6. Moon CI 실행 (변경된 프로젝트만)
          - name: Run Moon CI
            run: moon ci
            env:
              DATABASE_URL: postgres://postgres:password@localhost:5432/mindhit_test?sslmode=disable
              REDIS_URL: redis://localhost:6379
              JWT_SECRET: test-secret-key

          # 7. 결과 리포트 (PR 코멘트)
          - name: Report results
            uses: moonrepo/run-report-action@v1
            if: success() || failure()
            with:
              access-token: ${{ secrets.GITHUB_TOKEN }}
    ```

- [ ] **병렬 실행 워크플로우** (대규모 프로젝트용, 선택)
  - [ ] `.github/workflows/ci-parallel.yml`
    ```yaml
    name: CI (Parallel)

    on:
      push:
        branches:
          - main
      pull_request:
        branches:
          - main

    jobs:
      ci:
        name: CI Job ${{ matrix.index }}
        runs-on: ubuntu-latest
        strategy:
          matrix:
            index: [0, 1, 2]  # 3개 병렬 작업

        steps:
          - uses: actions/checkout@v4
            with:
              fetch-depth: 0

          - uses: moonrepo/setup-toolchain@v0
            with:
              auto-install: true

          - run: pnpm install --frozen-lockfile

          # 작업을 3개로 분할하여 병렬 실행
          - name: Run Moon CI (parallel)
            run: moon ci --job ${{ matrix.index }} --jobTotal 3
    ```

- [ ] **API 전용 moon.yml 태스크 정의**
  - [ ] `apps/api/moon.yml` 업데이트
    ```yaml
    $schema: 'https://moonrepo.dev/schemas/project.json'

    language: go
    type: application

    fileGroups:
      sources:
        - '**/*.go'
        - 'go.mod'
        - 'go.sum'
      tests:
        - '**/*_test.go'
      schemas:
        - 'ent/schema/**/*.go'

    tasks:
      # 빌드
      build:
        command: go build -o ./bin/server ./cmd/server
        inputs:
          - '@group(sources)'
        outputs:
          - bin/server

      # 테스트
      test:
        command: go test -v -race -coverprofile=coverage.out ./...
        inputs:
          - '@group(sources)'
          - '@group(tests)'

      # 린트
      lint:
        command: golangci-lint run
        inputs:
          - '@group(sources)'

      # 개발 서버 (로컬 전용)
      dev:
        command: air -c .air.toml
        local: true
        options:
          runInCI: false

      # Ent 코드 생성
      generate:
        command: go generate ./ent
        inputs:
          - '@group(schemas)'
        outputs:
          - ent/client.go
          - ent/ent.go

      # 마이그레이션 적용
      migrate-apply:
        command: atlas migrate apply --dir "file://ent/migrate/migrations" --url "${DATABASE_URL}"
        local: true
        options:
          runInCI: false

      # 마이그레이션 상태
      migrate-status:
        command: atlas migrate status --dir "file://ent/migrate/migrations" --url "${DATABASE_URL}"
        local: true
        options:
          runInCI: false
    ```

- [ ] **Web 프로젝트 moon.yml** (Next.js용)
  - [ ] `apps/web/moon.yml`
    ```yaml
    $schema: 'https://moonrepo.dev/schemas/project.json'

    language: typescript
    type: application

    dependsOn:
      - 'shared'
      - 'protocol'

    fileGroups:
      sources:
        - 'src/**/*'
        - 'app/**/*'
      tests:
        - '**/*.test.*'
        - '**/*.spec.*'
      configs:
        - 'next.config.*'
        - 'tsconfig.json'
        - 'tailwind.config.*'

    tasks:
      build:
        command: next build
        inputs:
          - '@group(sources)'
          - '@group(configs)'
        outputs:
          - .next

      dev:
        command: next dev
        local: true
        options:
          runInCI: false

      lint:
        command: next lint
        inputs:
          - '@group(sources)'

      test:
        command: vitest run
        inputs:
          - '@group(sources)'
          - '@group(tests)'

      typecheck:
        command: tsc --noEmit
        inputs:
          - '@group(sources)'
          - '@group(configs)'
    ```

- [ ] **Extension 프로젝트 moon.yml**
  - [ ] `apps/extension/moon.yml`
    ```yaml
    $schema: 'https://moonrepo.dev/schemas/project.json'

    language: typescript
    type: application

    dependsOn:
      - 'shared'
      - 'protocol'

    fileGroups:
      sources:
        - 'src/**/*'
      tests:
        - '**/*.test.*'
      configs:
        - 'manifest.json'
        - 'tsconfig.json'
        - 'vite.config.*'

    tasks:
      build:
        command: vite build
        inputs:
          - '@group(sources)'
          - '@group(configs)'
        outputs:
          - dist

      dev:
        command: vite build --watch
        local: true
        options:
          runInCI: false

      lint:
        command: eslint src --ext .ts,.tsx
        inputs:
          - '@group(sources)'

      test:
        command: vitest run
        inputs:
          - '@group(sources)'
          - '@group(tests)'

      typecheck:
        command: tsc --noEmit
        inputs:
          - '@group(sources)'
    ```

### Moon CI 명령어 참조

| 명령어 | 설명 |
|-------|------|
| `moon ci` | 변경된 프로젝트만 빌드/테스트/린트 실행 |
| `moon ci :build` | 변경된 프로젝트의 build 태스크만 실행 |
| `moon ci :test :lint` | 테스트와 린트만 실행 |
| `moon ci --base main` | main 브랜치 대비 변경 감지 |
| `moon ci --job 0 --jobTotal 3` | 병렬 작업 분할 (3개 중 첫 번째) |

### 환경 변수

| 변수 | 설명 |
|-----|------|
| `MOON_BASE` | 베이스 브랜치 (기본: vcs.defaultBranch) |
| `MOON_HEAD` | 비교 대상 커밋 (기본: HEAD) |
| `MOON_TOOLCHAIN_FORCE_GLOBALS` | 시스템 도구 사용 (CI 최적화) |

### 검증

```bash
# 로컬에서 CI 시뮬레이션
moon ci --dry-run

# 특정 브랜치 대비 변경 확인
moon ci --base main --dry-run

# 변경된 프로젝트 확인
moon query touched-files
moon query projects --affected
```

### CI 구성 파일 구조

```text
.github/
└── workflows/
    ├── ci.yml              # 기본 CI 워크플로우
    └── ci-parallel.yml     # 병렬 CI (선택)

apps/
├── api/
│   └── moon.yml           # Go API 태스크
├── web/
│   └── moon.yml           # Next.js 태스크
└── extension/
    └── moon.yml           # Extension 태스크
```

---

## Phase 0 완료 확인

### 전체 검증 체크리스트

- [ ] **Moon 설정 확인**
  ```bash
  pnpm moon --version
  pnpm moon query projects
  ```

- [ ] **Docker Compose 실행**
  ```bash
  docker-compose up -d postgres redis
  docker-compose ps
  # postgres, redis 모두 healthy
  ```

- [ ] **PostgreSQL 연결**
  ```bash
  docker exec -it mindhit-postgres psql -U postgres -d mindhit -c "\dt"
  ```

- [ ] **Redis 연결**
  ```bash
  docker exec -it mindhit-redis redis-cli ping
  # PONG
  ```

- [ ] **API 서버 실행 (로컬)**
  ```bash
  make api
  curl http://localhost:8080/health
  # {"status":"ok"}
  ```

- [ ] **Hot Reload 동작**
  - main.go 수정 시 자동 재시작 확인

- [ ] **Moon CI 테스트**
  ```bash
  # 로컬에서 CI 시뮬레이션
  pnpm moon ci --dry-run
  ```

### 산출물 요약

| 항목 | 위치 |
|-----|------|
| Moon Workspace | `.moon/workspace.yml` |
| Moon Toolchain | `.moon/toolchain.yml` |
| Moon Tasks (전역) | `.moon/tasks.yml` |
| Docker Compose | `docker-compose.yml` |
| API Dockerfile | `apps/api/Dockerfile.dev` |
| Air 설정 | `apps/api/.air.toml` |
| Makefile | `Makefile` |
| 환경 변수 | `apps/api/.env.example` |
| 설정 스크립트 | `scripts/dev-setup.sh` |
| GitHub Actions CI | `.github/workflows/ci.yml` |

### 환경 구성 요약

| 서비스 | 포트 | 용도 |
|-------|------|------|
| PostgreSQL | 5432 | 메인 데이터베이스 |
| Redis | 6379 | 캐시, Rate Limiting |
| API | 8080 | Go 백엔드 서버 |

### Moon 태스크 요약

| 명령어 | 설명 |
|-------|------|
| `pnpm moon run :build` | 모든 프로젝트 빌드 |
| `pnpm moon run :test` | 모든 프로젝트 테스트 |
| `pnpm moon run :lint` | 모든 프로젝트 린트 |
| `pnpm moon ci` | CI 파이프라인 (변경 감지) |
| `pnpm moon run api:dev` | API 개발 서버 |

---

## 다음 Phase

Phase 0 완료 후 [Phase 1: 프로젝트 초기화](./phase-1-project-init.md)로 진행하세요.
