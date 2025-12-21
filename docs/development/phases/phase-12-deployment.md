# Phase 12: 배포 및 운영

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | 프로덕션 배포 파이프라인 및 운영 절차 구축 |
| **선행 조건** | Phase 11 (모니터링) 완료 |
| **예상 소요** | 4 Steps |
| **결과물** | CI/CD 파이프라인, 배포 자동화, 운영 매뉴얼 |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 12.1 | CI/CD 파이프라인 구성 | ⬜ |
| 12.2 | 프로덕션 인프라 설정 | ⬜ |
| 12.3 | 배포 자동화 | ⬜ |
| 12.4 | 운영 매뉴얼 및 장애 대응 | ⬜ |

---

## Step 12.1: CI/CD 파이프라인 구성

### 목표
GitHub Actions 기반 CI/CD 파이프라인 구축

### 체크리스트

- [ ] **API CI 워크플로우**
  - [ ] `.github/workflows/api-ci.yml`
    ```yaml
    name: API CI

    on:
      push:
        branches: [main, develop]
        paths:
          - 'apps/api/**'
          - '.github/workflows/api-ci.yml'
      pull_request:
        branches: [main]
        paths:
          - 'apps/api/**'

    defaults:
      run:
        working-directory: apps/api

    jobs:
      lint:
        runs-on: ubuntu-latest
        steps:
          - uses: actions/checkout@v4

          - name: Set up Go
            uses: actions/setup-go@v5
            with:
              go-version: '1.22'

          - name: golangci-lint
            uses: golangci/golangci-lint-action@v4
            with:
              version: latest
              working-directory: apps/api

      test:
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
          - uses: actions/checkout@v4

          - name: Set up Go
            uses: actions/setup-go@v5
            with:
              go-version: '1.22'

          - name: Cache Go modules
            uses: actions/cache@v4
            with:
              path: ~/go/pkg/mod
              key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
              restore-keys: |
                ${{ runner.os }}-go-

          - name: Install dependencies
            run: go mod download

          - name: Run tests
            env:
              DATABASE_URL: postgres://postgres:password@localhost:5432/mindhit_test?sslmode=disable
              REDIS_URL: redis://localhost:6379
            run: go test -v -race -coverprofile=coverage.out ./...

          - name: Upload coverage
            uses: codecov/codecov-action@v4
            with:
              file: apps/api/coverage.out
              flags: api

      build:
        runs-on: ubuntu-latest
        needs: [lint, test]
        steps:
          - uses: actions/checkout@v4

          - name: Set up Go
            uses: actions/setup-go@v5
            with:
              go-version: '1.22'

          - name: Build
            run: CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/server

          - name: Upload artifact
            uses: actions/upload-artifact@v4
            with:
              name: api-binary
              path: apps/api/bin/server

      docker:
        runs-on: ubuntu-latest
        needs: build
        if: github.ref == 'refs/heads/main'
        steps:
          - uses: actions/checkout@v4

          - name: Set up Docker Buildx
            uses: docker/setup-buildx-action@v3

          - name: Login to Container Registry
            uses: docker/login-action@v3
            with:
              registry: ghcr.io
              username: ${{ github.actor }}
              password: ${{ secrets.GITHUB_TOKEN }}

          - name: Build and push
            uses: docker/build-push-action@v5
            with:
              context: apps/api
              push: true
              tags: |
                ghcr.io/${{ github.repository }}/api:latest
                ghcr.io/${{ github.repository }}/api:${{ github.sha }}
              cache-from: type=gha
              cache-to: type=gha,mode=max
    ```

- [ ] **프로덕션용 Dockerfile**
  - [ ] `apps/api/Dockerfile`
    ```dockerfile
    # Build stage
    FROM golang:1.22-alpine AS builder

    RUN apk add --no-cache git ca-certificates

    WORKDIR /app

    # 의존성 먼저 복사 (캐시 활용)
    COPY go.mod go.sum ./
    RUN go mod download

    # 소스 복사 및 빌드
    COPY . .
    RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/server

    # Runtime stage
    FROM alpine:3.19

    RUN apk add --no-cache ca-certificates tzdata

    # 비root 사용자 생성
    RUN adduser -D -u 1000 appuser
    USER appuser

    WORKDIR /app

    COPY --from=builder /server .

    EXPOSE 8080

    CMD ["./server"]
    ```

- [ ] **Database Migration 워크플로우**
  - [ ] `.github/workflows/db-migrate.yml`
    ```yaml
    name: Database Migration

    on:
      push:
        branches: [main]
        paths:
          - 'apps/api/ent/migrate/migrations/**'
      workflow_dispatch:
        inputs:
          environment:
            description: 'Target environment'
            required: true
            default: 'staging'
            type: choice
            options:
              - staging
              - production

    jobs:
      migrate:
        runs-on: ubuntu-latest
        environment: ${{ github.event.inputs.environment || 'staging' }}

        steps:
          - uses: actions/checkout@v4

          - name: Install Atlas
            run: |
              curl -sSf https://atlasgo.sh | sh

          - name: Run migrations
            env:
              DATABASE_URL: ${{ secrets.DATABASE_URL }}
            run: |
              cd apps/api
              atlas migrate apply \
                --dir "file://ent/migrate/migrations" \
                --url "$DATABASE_URL"

          - name: Notify on failure
            if: failure()
            uses: slackapi/slack-github-action@v1
            with:
              payload: |
                {
                  "text": "❌ Database migration failed on ${{ github.event.inputs.environment || 'staging' }}"
                }
            env:
              SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
    ```

- [ ] **Web App CI/CD**
  - [ ] `.github/workflows/webapp-ci.yml`
    ```yaml
    name: Web App CI

    on:
      push:
        branches: [main, develop]
        paths:
          - 'apps/web/**'
          - '.github/workflows/webapp-ci.yml'
      pull_request:
        branches: [main]
        paths:
          - 'apps/web/**'

    defaults:
      run:
        working-directory: apps/web

    jobs:
      lint-and-test:
        runs-on: ubuntu-latest
        steps:
          - uses: actions/checkout@v4

          - name: Setup Node.js
            uses: actions/setup-node@v4
            with:
              node-version: '20'

          - name: Setup pnpm
            uses: pnpm/action-setup@v2
            with:
              version: 8

          - name: Get pnpm store directory
            id: pnpm-cache
            shell: bash
            run: |
              echo "STORE_PATH=$(pnpm store path)" >> $GITHUB_OUTPUT

          - name: Setup pnpm cache
            uses: actions/cache@v4
            with:
              path: ${{ steps.pnpm-cache.outputs.STORE_PATH }}
              key: ${{ runner.os }}-pnpm-store-${{ hashFiles('**/pnpm-lock.yaml') }}
              restore-keys: |
                ${{ runner.os }}-pnpm-store-

          - name: Install dependencies
            run: pnpm install

          - name: Lint
            run: pnpm lint

          - name: Type check
            run: pnpm typecheck

          - name: Test
            run: pnpm test

          - name: Build
            run: pnpm build
    ```

### 검증
```bash
# GitHub Actions 워크플로우 문법 검사
act -l  # act 도구 사용 시

# 로컬에서 테스트
cd apps/api && go test -v ./...
cd apps/web && pnpm test
```

---

## Step 12.2: 프로덕션 인프라 설정

### 목표
AWS/Railway/Vercel 기반 프로덕션 인프라 구성

### 체크리스트

#### Option A: Railway (권장 - 초기)

- [ ] **Railway 프로젝트 설정**
  - [ ] `railway.json`
    ```json
    {
      "$schema": "https://railway.app/railway.schema.json",
      "build": {
        "builder": "DOCKERFILE",
        "dockerfilePath": "apps/api/Dockerfile"
      },
      "deploy": {
        "startCommand": "./server",
        "healthcheckPath": "/health",
        "healthcheckTimeout": 30,
        "restartPolicyType": "ON_FAILURE",
        "restartPolicyMaxRetries": 3
      }
    }
    ```

- [ ] **Railway 환경 변수 설정**
  ```
  PORT=8080
  ENVIRONMENT=production
  DATABASE_URL=${{Postgres.DATABASE_URL}}
  REDIS_URL=${{Redis.REDIS_URL}}
  JWT_SECRET=${{secret(JWT_SECRET)}}
  ```

#### Option B: AWS ECS (고가용성)

- [ ] **Terraform 기본 구성**
  - [ ] `infrastructure/terraform/main.tf`
    ```hcl
    terraform {
      required_version = ">= 1.5"
      required_providers {
        aws = {
          source  = "hashicorp/aws"
          version = "~> 5.0"
        }
      }

      backend "s3" {
        bucket = "mindhit-terraform-state"
        key    = "production/terraform.tfstate"
        region = "ap-northeast-2"
      }
    }

    provider "aws" {
      region = var.aws_region
    }

    module "vpc" {
      source  = "terraform-aws-modules/vpc/aws"
      version = "~> 5.0"

      name = "mindhit-vpc"
      cidr = "10.0.0.0/16"

      azs             = ["ap-northeast-2a", "ap-northeast-2c"]
      public_subnets  = ["10.0.1.0/24", "10.0.2.0/24"]
      private_subnets = ["10.0.10.0/24", "10.0.20.0/24"]

      enable_nat_gateway = true
      single_nat_gateway = true
    }

    module "rds" {
      source  = "terraform-aws-modules/rds/aws"
      version = "~> 6.0"

      identifier = "mindhit-postgres"

      engine         = "postgres"
      engine_version = "16"
      instance_class = "db.t3.small"

      allocated_storage     = 20
      max_allocated_storage = 100

      db_name  = "mindhit"
      username = var.db_username
      password = var.db_password

      vpc_security_group_ids = [aws_security_group.rds.id]
      subnet_ids             = module.vpc.private_subnets

      backup_retention_period = 7
      skip_final_snapshot     = false
    }

    module "elasticache" {
      source  = "terraform-aws-modules/elasticache/aws"
      version = "~> 1.0"

      cluster_id         = "mindhit-redis"
      engine             = "redis"
      node_type          = "cache.t3.micro"
      num_cache_nodes    = 1
      port               = 6379
      subnet_group_name  = aws_elasticache_subnet_group.main.name
      security_group_ids = [aws_security_group.redis.id]
    }
    ```

- [ ] **ECS 서비스 정의**
  - [ ] `infrastructure/terraform/ecs.tf`
    ```hcl
    resource "aws_ecs_cluster" "main" {
      name = "mindhit-cluster"

      setting {
        name  = "containerInsights"
        value = "enabled"
      }
    }

    resource "aws_ecs_task_definition" "api" {
      family                   = "mindhit-api"
      network_mode             = "awsvpc"
      requires_compatibilities = ["FARGATE"]
      cpu                      = 256
      memory                   = 512
      execution_role_arn       = aws_iam_role.ecs_execution.arn
      task_role_arn            = aws_iam_role.ecs_task.arn

      container_definitions = jsonencode([
        {
          name  = "api"
          image = "${aws_ecr_repository.api.repository_url}:latest"
          portMappings = [
            {
              containerPort = 8080
              protocol      = "tcp"
            }
          ]
          environment = [
            { name = "PORT", value = "8080" },
            { name = "ENVIRONMENT", value = "production" }
          ]
          secrets = [
            { name = "DATABASE_URL", valueFrom = aws_secretsmanager_secret.db_url.arn },
            { name = "REDIS_URL", valueFrom = aws_secretsmanager_secret.redis_url.arn },
            { name = "JWT_SECRET", valueFrom = aws_secretsmanager_secret.jwt_secret.arn }
          ]
          logConfiguration = {
            logDriver = "awslogs"
            options = {
              "awslogs-group"         = aws_cloudwatch_log_group.api.name
              "awslogs-region"        = var.aws_region
              "awslogs-stream-prefix" = "api"
            }
          }
          healthCheck = {
            command     = ["CMD-SHELL", "wget -q --spider http://localhost:8080/health || exit 1"]
            interval    = 30
            timeout     = 5
            retries     = 3
            startPeriod = 60
          }
        }
      ])
    }

    resource "aws_ecs_service" "api" {
      name            = "mindhit-api"
      cluster         = aws_ecs_cluster.main.id
      task_definition = aws_ecs_task_definition.api.arn
      desired_count   = 2
      launch_type     = "FARGATE"

      network_configuration {
        subnets          = module.vpc.private_subnets
        security_groups  = [aws_security_group.api.id]
        assign_public_ip = false
      }

      load_balancer {
        target_group_arn = aws_lb_target_group.api.arn
        container_name   = "api"
        container_port   = 8080
      }

      deployment_circuit_breaker {
        enable   = true
        rollback = true
      }
    }
    ```

#### Vercel (Web App)

- [ ] **Vercel 설정**
  - [ ] `apps/web/vercel.json`
    ```json
    {
      "framework": "nextjs",
      "buildCommand": "pnpm build",
      "devCommand": "pnpm dev",
      "installCommand": "pnpm install",
      "regions": ["icn1"],
      "env": {
        "NEXT_PUBLIC_API_URL": "@api_url"
      }
    }
    ```

### 검증
```bash
# Railway CLI
railway status
railway logs

# Terraform
cd infrastructure/terraform
terraform plan
terraform apply

# Vercel
vercel --prod
```

---

## Step 12.3: 배포 자동화

### 목표
Blue-Green 또는 Rolling 배포 전략 구현

### 체크리스트

- [ ] **배포 워크플로우**
  - [ ] `.github/workflows/deploy.yml`
    ```yaml
    name: Deploy

    on:
      push:
        branches: [main]
      workflow_dispatch:
        inputs:
          environment:
            description: 'Target environment'
            required: true
            default: 'staging'
            type: choice
            options:
              - staging
              - production

    jobs:
      deploy-api:
        runs-on: ubuntu-latest
        environment: ${{ github.event.inputs.environment || 'staging' }}

        steps:
          - uses: actions/checkout@v4

          - name: Configure AWS credentials
            uses: aws-actions/configure-aws-credentials@v4
            with:
              aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
              aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
              aws-region: ap-northeast-2

          - name: Login to Amazon ECR
            id: login-ecr
            uses: aws-actions/amazon-ecr-login@v2

          - name: Build and push Docker image
            env:
              ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
              ECR_REPOSITORY: mindhit-api
              IMAGE_TAG: ${{ github.sha }}
            run: |
              docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG apps/api
              docker push $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG
              docker tag $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG $ECR_REGISTRY/$ECR_REPOSITORY:latest
              docker push $ECR_REGISTRY/$ECR_REPOSITORY:latest

          - name: Update ECS service
            run: |
              aws ecs update-service \
                --cluster mindhit-cluster \
                --service mindhit-api \
                --force-new-deployment

          - name: Wait for deployment
            run: |
              aws ecs wait services-stable \
                --cluster mindhit-cluster \
                --services mindhit-api

          - name: Notify success
            if: success()
            uses: slackapi/slack-github-action@v1
            with:
              payload: |
                {
                  "text": "✅ API deployed successfully to ${{ github.event.inputs.environment || 'staging' }}"
                }
            env:
              SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}

          - name: Notify failure
            if: failure()
            uses: slackapi/slack-github-action@v1
            with:
              payload: |
                {
                  "text": "❌ API deployment failed to ${{ github.event.inputs.environment || 'staging' }}"
                }
            env:
              SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}

      deploy-web:
        runs-on: ubuntu-latest
        environment: ${{ github.event.inputs.environment || 'staging' }}
        needs: deploy-api

        steps:
          - uses: actions/checkout@v4

          - name: Deploy to Vercel
            uses: amondnet/vercel-action@v25
            with:
              vercel-token: ${{ secrets.VERCEL_TOKEN }}
              vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
              vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
              vercel-args: '--prod'
              working-directory: apps/web
    ```

- [ ] **롤백 스크립트**
  - [ ] `scripts/rollback.sh`
    ```bash
    #!/bin/bash
    set -e

    CLUSTER="mindhit-cluster"
    SERVICE="mindhit-api"
    REGION="ap-northeast-2"

    echo "=== MindHit API Rollback ==="

    # 현재 태스크 정의 조회
    CURRENT_TASK=$(aws ecs describe-services \
      --cluster $CLUSTER \
      --services $SERVICE \
      --region $REGION \
      --query 'services[0].taskDefinition' \
      --output text)

    echo "Current task definition: $CURRENT_TASK"

    # 이전 태스크 정의 조회
    TASK_FAMILY=$(echo $CURRENT_TASK | cut -d'/' -f2 | cut -d':' -f1)
    CURRENT_REVISION=$(echo $CURRENT_TASK | cut -d':' -f2)
    PREVIOUS_REVISION=$((CURRENT_REVISION - 1))

    if [ $PREVIOUS_REVISION -lt 1 ]; then
      echo "No previous revision available for rollback"
      exit 1
    fi

    PREVIOUS_TASK="$TASK_FAMILY:$PREVIOUS_REVISION"
    echo "Rolling back to: $PREVIOUS_TASK"

    # 서비스 업데이트
    aws ecs update-service \
      --cluster $CLUSTER \
      --service $SERVICE \
      --task-definition $PREVIOUS_TASK \
      --region $REGION

    # 배포 완료 대기
    echo "Waiting for rollback to complete..."
    aws ecs wait services-stable \
      --cluster $CLUSTER \
      --services $SERVICE \
      --region $REGION

    echo "Rollback completed successfully!"
    ```

- [ ] **실행 권한 부여**
  ```bash
  chmod +x scripts/rollback.sh
  ```

### 검증
```bash
# 배포 상태 확인
aws ecs describe-services --cluster mindhit-cluster --services mindhit-api

# 롤백 테스트
./scripts/rollback.sh
```

---

## Step 12.4: 운영 매뉴얼 및 장애 대응

### 목표
운영 절차 및 장애 대응 가이드라인 문서화

### 체크리스트

- [ ] **운영 매뉴얼**
  - [ ] `docs/operations/runbook.md`
    ```markdown
    # MindHit 운영 매뉴얼

    ## 서비스 개요

    | 서비스 | URL | 담당 |
    |-------|-----|------|
    | API | api.mindhit.io | Backend |
    | Web | app.mindhit.io | Frontend |
    | Grafana | grafana.mindhit.io | DevOps |

    ## 일상 점검

    ### 매일
    - [ ] Grafana 대시보드 확인
    - [ ] 에러 로그 확인
    - [ ] 리소스 사용량 확인

    ### 매주
    - [ ] 백업 상태 확인
    - [ ] 보안 업데이트 확인
    - [ ] 비용 리포트 확인

    ## 일반적인 운영 절차

    ### 배포
    1. `develop` 브랜치에서 `main` 브랜치로 PR 생성
    2. 코드 리뷰 완료
    3. PR 머지 → 자동 배포 트리거
    4. Slack 알림 확인
    5. 배포 후 모니터링 (15분)

    ### 롤백
    1. 문제 발생 시 즉시 롤백 결정
    2. `./scripts/rollback.sh` 실행
    3. 롤백 완료 확인
    4. 원인 분석 및 수정

    ### 데이터베이스 마이그레이션
    1. 스테이징에서 먼저 테스트
    2. 프로덕션 백업 확인
    3. GitHub Actions 워크플로우 트리거
    4. 마이그레이션 완료 확인
    ```

- [ ] **장애 대응 가이드**
  - [ ] `docs/operations/incident-response.md`
    ```markdown
    # 장애 대응 가이드

    ## 심각도 정의

    | 레벨 | 정의 | 응답 시간 | 해결 시간 |
    |------|------|---------|---------|
    | P1 (Critical) | 서비스 완전 중단 | 5분 | 1시간 |
    | P2 (High) | 주요 기능 장애 | 15분 | 4시간 |
    | P3 (Medium) | 일부 기능 장애 | 1시간 | 24시간 |
    | P4 (Low) | 경미한 이슈 | 4시간 | 1주 |

    ## 장애 대응 절차

    ### 1. 감지 (Detection)
    - Alertmanager 알림 수신
    - 사용자 리포트
    - 모니터링 대시보드

    ### 2. 분류 (Triage)
    - 심각도 판단 (P1-P4)
    - 영향 범위 파악
    - 담당자 지정

    ### 3. 대응 (Response)
    - 즉시 조치 (롤백, 스케일링 등)
    - 고객 커뮤니케이션
    - 실시간 상황 공유

    ### 4. 복구 (Recovery)
    - 서비스 정상화 확인
    - 모니터링 강화
    - 임시 조치 안정성 확인

    ### 5. 사후 분석 (Post-mortem)
    - 타임라인 정리
    - 근본 원인 분석
    - 재발 방지 대책 수립

    ## 일반적인 장애 시나리오

    ### API 서버 무응답

    **증상**: `/health` 엔드포인트 타임아웃

    **진단**:
    ```bash
    # ECS 태스크 상태 확인
    aws ecs describe-tasks --cluster mindhit-cluster --tasks <task-id>

    # 로그 확인
    aws logs tail /ecs/mindhit-api --follow
    ```

    **조치**:
    1. ECS 서비스 강제 재배포
       ```bash
       aws ecs update-service --cluster mindhit-cluster --service mindhit-api --force-new-deployment
       ```
    2. 로드밸런서 헬스체크 확인
    3. 필요시 롤백

    ### 데이터베이스 연결 실패

    **증상**: `connection refused` 또는 `too many connections`

    **진단**:
    ```bash
    # RDS 상태 확인
    aws rds describe-db-instances --db-instance-identifier mindhit-postgres

    # 연결 수 확인 (psql)
    SELECT count(*) FROM pg_stat_activity;
    ```

    **조치**:
    1. 연결 풀 설정 확인
    2. 불필요한 연결 종료
    3. RDS 인스턴스 재시작 (최후 수단)

    ### Redis 메모리 부족

    **증상**: OOM 에러, 캐시 실패

    **진단**:
    ```bash
    # Redis INFO 확인
    redis-cli INFO memory
    ```

    **조치**:
    1. 불필요한 키 삭제
    2. TTL 정책 검토
    3. 인스턴스 스케일업

    ### AI 처리 지연

    **증상**: 세션 완료 지연, 큐 백로그 증가

    **진단**:
    ```bash
    # 큐 상태 확인
    # Prometheus 메트릭 확인
    curl -s 'http://prometheus:9090/api/v1/query?query=mindhit_ai_processing_duration_seconds'
    ```

    **조치**:
    1. OpenAI API 상태 확인
    2. 워커 수 증가
    3. 처리 속도 제한 적용
    ```

- [ ] **연락처 및 에스컬레이션**
  - [ ] `docs/operations/contacts.md`
    ```markdown
    # 비상 연락처

    ## 팀 연락처

    | 역할 | 담당자 | 연락처 |
    |------|--------|--------|
    | Backend Lead | - | - |
    | Frontend Lead | - | - |
    | DevOps | - | - |

    ## 외부 서비스

    | 서비스 | 지원 페이지 |
    |--------|-----------|
    | AWS | https://console.aws.amazon.com/support |
    | Vercel | https://vercel.com/support |
    | OpenAI | https://help.openai.com |

    ## 에스컬레이션

    1. **1차**: 담당 개발자 (10분 내 응답 없으면)
    2. **2차**: 팀 리드 (20분 내 해결 없으면)
    3. **3차**: CTO (P1 장애 시)
    ```

- [ ] **백업 및 복구**
  - [ ] `docs/operations/backup-recovery.md`
    ```markdown
    # 백업 및 복구 절차

    ## 자동 백업

    ### PostgreSQL (RDS)
    - 일일 자동 백업 (7일 보관)
    - Point-in-time Recovery 활성화

    ### Redis
    - AOF persistence 활성화
    - 일일 스냅샷

    ## 수동 백업

    ### PostgreSQL
    ```bash
    # pg_dump 백업
    pg_dump -h $DB_HOST -U postgres mindhit > backup_$(date +%Y%m%d).sql

    # S3 업로드
    aws s3 cp backup_$(date +%Y%m%d).sql s3://mindhit-backups/postgres/
    ```

    ## 복구 절차

    ### PostgreSQL 복구
    1. RDS 스냅샷에서 새 인스턴스 생성
    2. 또는 pg_dump 파일에서 복원:
       ```bash
       psql -h $NEW_DB_HOST -U postgres mindhit < backup_20240101.sql
       ```

    ### Redis 복구
    1. ElastiCache 스냅샷에서 복원
    2. 또는 AOF 파일에서 복원

    ## 복구 테스트

    - 분기별 복구 테스트 실시
    - 테스트 환경에서 실제 복원 수행
    - 복원 시간 및 데이터 정합성 확인
    ```

### 검증
- [ ] 운영 매뉴얼 문서 완성
- [ ] 장애 대응 시나리오별 절차 확인
- [ ] 백업/복구 테스트 실행

---

## Phase 12 완료 확인

### 전체 검증 체크리스트

- [ ] **CI/CD 파이프라인**
  - [ ] API CI 워크플로우 동작
  - [ ] Web App CI 워크플로우 동작
  - [ ] DB Migration 워크플로우 동작

- [ ] **프로덕션 인프라**
  - [ ] API 서버 배포 완료
  - [ ] 데이터베이스 접근 가능
  - [ ] Redis 접근 가능

- [ ] **배포 자동화**
  - [ ] 자동 배포 동작
  - [ ] 롤백 스크립트 동작
  - [ ] Slack 알림 수신

- [ ] **운영 문서**
  - [ ] 운영 매뉴얼 작성 완료
  - [ ] 장애 대응 가이드 작성 완료
  - [ ] 백업/복구 절차 문서화

### 산출물 요약

| 항목 | 위치 |
|-----|------|
| API CI 워크플로우 | `.github/workflows/api-ci.yml` |
| Web App CI 워크플로우 | `.github/workflows/webapp-ci.yml` |
| DB Migration 워크플로우 | `.github/workflows/db-migrate.yml` |
| 배포 워크플로우 | `.github/workflows/deploy.yml` |
| 프로덕션 Dockerfile | `apps/api/Dockerfile` |
| Terraform 설정 | `infrastructure/terraform/` |
| 롤백 스크립트 | `scripts/rollback.sh` |
| 운영 매뉴얼 | `docs/operations/runbook.md` |
| 장애 대응 가이드 | `docs/operations/incident-response.md` |
| 백업/복구 절차 | `docs/operations/backup-recovery.md` |

### 배포 환경 요약

| 환경 | 용도 | 배포 트리거 |
|-----|------|-----------|
| Staging | 테스트 | `develop` 브랜치 push |
| Production | 실서비스 | `main` 브랜치 push |

---

## 프로젝트 완료

축하합니다! Phase 12까지 완료하면 MindHit 프로젝트의 모든 개발 및 운영 체계가 구축됩니다.

### 전체 Phase 요약

| Phase | 내용 | 상태 |
|-------|------|------|
| 0 | Docker 개발 환경 | ⬜ |
| 1 | 프로젝트 초기화 | ⬜ |
| 1.5 | API 스펙 공통화 | ⬜ |
| 2 | 인증 시스템 | ⬜ |
| 3 | 세션 관리 API | ⬜ |
| 4 | 이벤트 수집 API | ⬜ |
| 5 | 모니터링 및 인프라 | ⬜ |
| 6 | 스케줄러 | ⬜ |
| 7 | Next.js 웹앱 | ⬜ |
| 8 | Chrome Extension | ⬜ |
| 9 | AI 마인드맵 | ⬜ |
| 10 | 웹앱 대시보드 | ⬜ |
| 11 | 모니터링 시스템 | ⬜ |
| 12 | 배포 및 운영 | ⬜ |
