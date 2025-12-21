# Phase 0: 개발 환경 구성

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | Docker 기반 로컬 개발 환경 구성 |
| **선행 조건** | Docker, Docker Compose 설치 |
| **예상 소요** | 2 Steps |
| **결과물** | PostgreSQL, Redis, API 서버가 Docker로 동작 |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 0.1 | Docker Compose 구성 | ⬜ |
| 0.2 | 개발 환경 스크립트 | ⬜ |

---

## Step 0.1: Docker Compose 구성

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

## Step 0.2: 개발 환경 스크립트

### 목표
개발 환경 관리를 위한 편의 스크립트 생성

### 체크리스트

- [ ] **Makefile 생성** (루트)
  ```makefile
  .PHONY: help dev dev-db dev-full down logs clean migrate test lint

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
  	@echo "개발:"
  	@echo "  make api         - API 서버 로컬 실행 (DB는 Docker)"
  	@echo "  make migrate     - 마이그레이션 적용"
  	@echo "  make migrate-new - 새 마이그레이션 생성"
  	@echo "  make generate    - Ent 코드 생성"
  	@echo ""
  	@echo "테스트:"
  	@echo "  make test        - 전체 테스트 실행"
  	@echo "  make lint        - 린트 실행"

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

  # API 서버 로컬 실행 (Hot reload)
  api:
  	cd apps/api && air

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

  # 마이그레이션 적용
  migrate:
  	cd apps/api && atlas migrate apply \
  		--dir "file://ent/migrate/migrations" \
  		--url "postgres://postgres:password@localhost:5432/mindhit?sslmode=disable"

  # 새 마이그레이션 생성
  migrate-new:
  	@read -p "Migration name: " name; \
  	cd apps/api && atlas migrate diff $$name \
  		--dir "file://ent/migrate/migrations" \
  		--to "ent://ent/schema" \
  		--dev-url "postgres://postgres:password@localhost:5432/mindhit_dev?sslmode=disable"

  # Ent 코드 생성
  generate:
  	cd apps/api && go generate ./ent

  # 테스트 실행
  test:
  	cd apps/api && go test -v -race -cover ./...

  # 린트 실행
  lint:
  	cd apps/api && golangci-lint run
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

## Phase 0 완료 확인

### 전체 검증 체크리스트

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

### 산출물 요약

| 항목 | 위치 |
|-----|------|
| Docker Compose | `docker-compose.yml` |
| API Dockerfile | `apps/api/Dockerfile.dev` |
| Air 설정 | `apps/api/.air.toml` |
| Makefile | `Makefile` |
| 환경 변수 | `apps/api/.env.example` |
| 설정 스크립트 | `scripts/dev-setup.sh` |

### 환경 구성 요약

| 서비스 | 포트 | 용도 |
|-------|------|------|
| PostgreSQL | 5432 | 메인 데이터베이스 |
| Redis | 6379 | 캐시, Rate Limiting |
| API | 8080 | Go 백엔드 서버 |

---

## 다음 Phase

Phase 0 완료 후 [Phase 1: 프로젝트 초기화](./phase-1-project-init.md)로 진행하세요.
