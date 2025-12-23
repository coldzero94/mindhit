# Phase 13: 배포 및 운영

## 개요

| 항목 | 내용 |
|------|------|
| **목표** | Terraform + EKS 기반 프로덕션 배포 및 운영 체계 구축 |
| **선행 조건** | Phase 12 (모니터링) 완료 |
| **예상 소요** | 4 Steps |
| **결과물** | IaC 기반 인프라, CI/CD 파이프라인, 운영 매뉴얼 |

---

## 아키텍처 개요

```mermaid
flowchart TB
    subgraph AWS["AWS Cloud"]
        subgraph VPC["VPC (10.0.0.0/16)"]
            subgraph Public["Public Subnet"]
                ALB[ALB<br/>Ingress]
            end

            subgraph Private["Private Subnet"]
                subgraph EKS["EKS Cluster"]
                    API[API Server]
                    WORKER[Worker]
                end
            end

            subgraph Managed["Managed Services"]
                RDS[(RDS<br/>PostgreSQL)]
                ELASTICACHE[(ElastiCache<br/>Redis)]
                ECR[(ECR<br/>Registry)]
            end
        end
    end

    ALB --> API
    API --> RDS
    API --> ELASTICACHE
    WORKER --> RDS
    WORKER --> ELASTICACHE

    style AWS fill:#f5f5f5
    style VPC fill:#e3f2fd
    style Public fill:#fff3e0
    style Private fill:#e8f5e9
    style EKS fill:#c8e6c9
    style Managed fill:#fce4ec
```

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 13.1 | Terraform 인프라 구성 | ⬜ |
| 13.2 | CI/CD 파이프라인 | ⬜ |
| 13.3 | 배포 자동화 | ⬜ |
| 13.4 | 운영 매뉴얼 | ⬜ |

---

## Step 13.1: Terraform 인프라 구성

### 목표

Terraform으로 AWS 인프라 전체를 코드로 관리합니다.

### 디렉토리 구조

```
infra/
├── modules/                    # Terraform 모듈
│   ├── vpc/
│   ├── eks/
│   ├── rds/
│   ├── elasticache/
│   └── ecr/
├── environments/               # 환경별 Terraform
│   ├── staging/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── terraform.tfvars
│   └── prod/
│       ├── main.tf
│       ├── variables.tf
│       └── terraform.tfvars
├── helm/                       # Helm Charts
│   └── mindhit/
├── docker/                     # Docker Compose (go run 모드)
├── kind/                       # kind 클러스터 설정 (로컬 K8s)
└── scripts/
```

### 체크리스트

- [ ] **VPC 모듈**

  `infra/modules/vpc/main.tf`:

  ```hcl
  variable "name" {
    type = string
  }

  variable "cidr" {
    type    = string
    default = "10.0.0.0/16"
  }

  variable "azs" {
    type    = list(string)
    default = ["ap-northeast-2a", "ap-northeast-2c"]
  }

  module "vpc" {
    source  = "terraform-aws-modules/vpc/aws"
    version = "~> 5.0"

    name = var.name
    cidr = var.cidr

    azs             = var.azs
    public_subnets  = ["10.0.1.0/24", "10.0.2.0/24"]
    private_subnets = ["10.0.10.0/24", "10.0.20.0/24"]

    enable_nat_gateway     = true
    single_nat_gateway     = true
    enable_dns_hostnames   = true
    enable_dns_support     = true

    public_subnet_tags = {
      "kubernetes.io/role/elb" = 1
    }

    private_subnet_tags = {
      "kubernetes.io/role/internal-elb" = 1
    }

    tags = {
      Environment = var.name
      Terraform   = "true"
    }
  }

  output "vpc_id" {
    value = module.vpc.vpc_id
  }

  output "private_subnets" {
    value = module.vpc.private_subnets
  }

  output "public_subnets" {
    value = module.vpc.public_subnets
  }
  ```

- [ ] **EKS 모듈**

  `infra/modules/eks/main.tf`:

  ```hcl
  variable "cluster_name" {
    type = string
  }

  variable "cluster_version" {
    type    = string
    default = "1.29"
  }

  variable "vpc_id" {
    type = string
  }

  variable "subnet_ids" {
    type = list(string)
  }

  variable "node_instance_types" {
    type    = list(string)
    default = ["t3.medium"]
  }

  variable "node_desired_size" {
    type    = number
    default = 2
  }

  module "eks" {
    source  = "terraform-aws-modules/eks/aws"
    version = "~> 20.0"

    cluster_name    = var.cluster_name
    cluster_version = var.cluster_version

    cluster_endpoint_public_access = true

    vpc_id     = var.vpc_id
    subnet_ids = var.subnet_ids

    eks_managed_node_groups = {
      default = {
        name           = "${var.cluster_name}-nodes"
        instance_types = var.node_instance_types

        min_size     = 1
        max_size     = 5
        desired_size = var.node_desired_size

        labels = {
          Environment = var.cluster_name
        }
      }
    }

    # OIDC provider for IAM roles
    enable_irsa = true

    tags = {
      Environment = var.cluster_name
      Terraform   = "true"
    }
  }

  output "cluster_endpoint" {
    value = module.eks.cluster_endpoint
  }

  output "cluster_name" {
    value = module.eks.cluster_name
  }

  output "cluster_certificate_authority_data" {
    value = module.eks.cluster_certificate_authority_data
  }

  output "oidc_provider_arn" {
    value = module.eks.oidc_provider_arn
  }
  ```

- [ ] **RDS 모듈**

  `infra/modules/rds/main.tf`:

  ```hcl
  variable "identifier" {
    type = string
  }

  variable "vpc_id" {
    type = string
  }

  variable "subnet_ids" {
    type = list(string)
  }

  variable "allowed_security_group_ids" {
    type = list(string)
  }

  variable "db_name" {
    type    = string
    default = "mindhit"
  }

  variable "db_username" {
    type      = string
    sensitive = true
  }

  variable "db_password" {
    type      = string
    sensitive = true
  }

  variable "instance_class" {
    type    = string
    default = "db.t3.small"
  }

  resource "aws_db_subnet_group" "main" {
    name       = var.identifier
    subnet_ids = var.subnet_ids

    tags = {
      Name = var.identifier
    }
  }

  resource "aws_security_group" "rds" {
    name   = "${var.identifier}-rds-sg"
    vpc_id = var.vpc_id

    ingress {
      from_port       = 5432
      to_port         = 5432
      protocol        = "tcp"
      security_groups = var.allowed_security_group_ids
    }

    egress {
      from_port   = 0
      to_port     = 0
      protocol    = "-1"
      cidr_blocks = ["0.0.0.0/0"]
    }

    tags = {
      Name = "${var.identifier}-rds-sg"
    }
  }

  module "rds" {
    source  = "terraform-aws-modules/rds/aws"
    version = "~> 6.0"

    identifier = var.identifier

    engine         = "postgres"
    engine_version = "16"
    instance_class = var.instance_class

    allocated_storage     = 20
    max_allocated_storage = 100

    db_name  = var.db_name
    username = var.db_username
    password = var.db_password

    vpc_security_group_ids = [aws_security_group.rds.id]
    db_subnet_group_name   = aws_db_subnet_group.main.name

    backup_retention_period = 7
    skip_final_snapshot     = false

    performance_insights_enabled = true

    tags = {
      Environment = var.identifier
      Terraform   = "true"
    }
  }

  output "endpoint" {
    value = module.rds.db_instance_endpoint
  }

  output "database_name" {
    value = var.db_name
  }
  ```

- [ ] **ElastiCache 모듈**

  `infra/modules/elasticache/main.tf`:

  ```hcl
  variable "cluster_id" {
    type = string
  }

  variable "vpc_id" {
    type = string
  }

  variable "subnet_ids" {
    type = list(string)
  }

  variable "allowed_security_group_ids" {
    type = list(string)
  }

  variable "node_type" {
    type    = string
    default = "cache.t3.micro"
  }

  resource "aws_elasticache_subnet_group" "main" {
    name       = var.cluster_id
    subnet_ids = var.subnet_ids
  }

  resource "aws_security_group" "redis" {
    name   = "${var.cluster_id}-redis-sg"
    vpc_id = var.vpc_id

    ingress {
      from_port       = 6379
      to_port         = 6379
      protocol        = "tcp"
      security_groups = var.allowed_security_group_ids
    }

    egress {
      from_port   = 0
      to_port     = 0
      protocol    = "-1"
      cidr_blocks = ["0.0.0.0/0"]
    }

    tags = {
      Name = "${var.cluster_id}-redis-sg"
    }
  }

  resource "aws_elasticache_cluster" "main" {
    cluster_id           = var.cluster_id
    engine               = "redis"
    node_type            = var.node_type
    num_cache_nodes      = 1
    parameter_group_name = "default.redis7"
    port                 = 6379
    subnet_group_name    = aws_elasticache_subnet_group.main.name
    security_group_ids   = [aws_security_group.redis.id]

    tags = {
      Environment = var.cluster_id
      Terraform   = "true"
    }
  }

  output "endpoint" {
    value = aws_elasticache_cluster.main.cache_nodes[0].address
  }

  output "port" {
    value = aws_elasticache_cluster.main.port
  }
  ```

- [ ] **ECR 모듈**

  `infra/modules/ecr/main.tf`:

  ```hcl
  variable "repository_names" {
    type = list(string)
  }

  resource "aws_ecr_repository" "main" {
    for_each = toset(var.repository_names)

    name                 = each.value
    image_tag_mutability = "MUTABLE"

    image_scanning_configuration {
      scan_on_push = true
    }

    tags = {
      Terraform = "true"
    }
  }

  resource "aws_ecr_lifecycle_policy" "main" {
    for_each   = toset(var.repository_names)
    repository = aws_ecr_repository.main[each.key].name

    policy = jsonencode({
      rules = [
        {
          rulePriority = 1
          description  = "Keep last 10 images"
          selection = {
            tagStatus     = "any"
            countType     = "imageCountMoreThan"
            countNumber   = 10
          }
          action = {
            type = "expire"
          }
        }
      ]
    })
  }

  output "repository_urls" {
    value = { for k, v in aws_ecr_repository.main : k => v.repository_url }
  }
  ```

- [ ] **Production 환경 구성**

  `infra/environments/prod/main.tf`:

  ```hcl
  terraform {
    required_version = ">= 1.5"

    required_providers {
      aws = {
        source  = "hashicorp/aws"
        version = "~> 5.0"
      }
      kubernetes = {
        source  = "hashicorp/kubernetes"
        version = "~> 2.0"
      }
      helm = {
        source  = "hashicorp/helm"
        version = "~> 2.0"
      }
    }

    backend "s3" {
      bucket         = "mindhit-terraform-state"
      key            = "production/terraform.tfstate"
      region         = "ap-northeast-2"
      encrypt        = true
      dynamodb_table = "mindhit-terraform-locks"
    }
  }

  provider "aws" {
    region = var.aws_region
  }

  # VPC
  module "vpc" {
    source = "../../modules/vpc"
    name   = "mindhit-production"
  }

  # EKS
  module "eks" {
    source = "../../modules/eks"

    cluster_name        = "mindhit-production"
    vpc_id              = module.vpc.vpc_id
    subnet_ids          = module.vpc.private_subnets
    node_instance_types = ["t3.medium"]
    node_desired_size   = 2
  }

  # RDS
  module "rds" {
    source = "../../modules/rds"

    identifier                 = "mindhit-production"
    vpc_id                     = module.vpc.vpc_id
    subnet_ids                 = module.vpc.private_subnets
    allowed_security_group_ids = [module.eks.cluster_security_group_id]
    db_username                = var.db_username
    db_password                = var.db_password
    instance_class             = "db.t3.small"
  }

  # ElastiCache
  module "elasticache" {
    source = "../../modules/elasticache"

    cluster_id                 = "mindhit-production"
    vpc_id                     = module.vpc.vpc_id
    subnet_ids                 = module.vpc.private_subnets
    allowed_security_group_ids = [module.eks.cluster_security_group_id]
    node_type                  = "cache.t3.micro"
  }

  # ECR
  module "ecr" {
    source = "../../modules/ecr"

    repository_names = ["mindhit/api", "mindhit/worker"]
  }

  # Kubernetes Provider
  provider "kubernetes" {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)

    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
    }
  }

  # Helm Provider
  provider "helm" {
    kubernetes {
      host                   = module.eks.cluster_endpoint
      cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)

      exec {
        api_version = "client.authentication.k8s.io/v1beta1"
        command     = "aws"
        args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
      }
    }
  }

  # AWS Load Balancer Controller
  module "aws_load_balancer_controller" {
    source = "git::https://github.com/DNXLabs/terraform-aws-eks-lb-controller.git"

    cluster_identity_oidc_issuer     = module.eks.oidc_provider_arn
    cluster_identity_oidc_issuer_arn = module.eks.oidc_provider_arn
    cluster_name                     = module.eks.cluster_name
  }

  # Outputs
  output "eks_cluster_endpoint" {
    value = module.eks.cluster_endpoint
  }

  output "rds_endpoint" {
    value     = module.rds.endpoint
    sensitive = true
  }

  output "redis_endpoint" {
    value = "${module.elasticache.endpoint}:${module.elasticache.port}"
  }

  output "ecr_repositories" {
    value = module.ecr.repository_urls
  }
  ```

  `infra/environments/prod/variables.tf`:

  ```hcl
  variable "aws_region" {
    type    = string
    default = "ap-northeast-2"
  }

  variable "db_username" {
    type      = string
    sensitive = true
  }

  variable "db_password" {
    type      = string
    sensitive = true
  }
  ```

### 검증

```bash
cd infra/environments/prod

# 초기화
terraform init

# 계획 확인
terraform plan

# 적용 (주의: 비용 발생)
terraform apply
```

---

## Step 13.2: CI/CD 파이프라인

### 목표

GitHub Actions 기반 CI/CD 파이프라인 구축

### 체크리스트

- [ ] **Backend CI 워크플로우**

  `.github/workflows/backend-ci.yml`:

  ```yaml
  name: Backend CI

  on:
    push:
      branches: [main, develop]
      paths:
        - 'apps/backend/**'
        - '.github/workflows/backend-ci.yml'
    pull_request:
      branches: [main]
      paths:
        - 'apps/backend/**'

  defaults:
    run:
      working-directory: apps/backend

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
            working-directory: apps/backend

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

        - name: Run tests
          env:
            DATABASE_URL: postgres://postgres:password@localhost:5432/mindhit_test?sslmode=disable
            REDIS_ADDR: localhost:6379
          run: go test -v -race -coverprofile=coverage.out ./...

        - name: Upload coverage
          uses: codecov/codecov-action@v4
          with:
            file: apps/backend/coverage.out
            flags: backend

    build:
      runs-on: ubuntu-latest
      needs: [lint, test]
      steps:
        - uses: actions/checkout@v4

        - name: Set up Go
          uses: actions/setup-go@v5
          with:
            go-version: '1.22'

        - name: Build API
          run: CGO_ENABLED=0 GOOS=linux go build -o bin/api ./cmd/api

        - name: Build Worker
          run: CGO_ENABLED=0 GOOS=linux go build -o bin/worker ./cmd/worker

        - name: Upload artifacts
          uses: actions/upload-artifact@v4
          with:
            name: backend-binaries
            path: apps/backend/bin/
  ```

- [ ] **Docker 빌드 및 푸시 워크플로우**

  `.github/workflows/docker-build.yml`:

  ```yaml
  name: Docker Build

  on:
    push:
      branches: [main]
      paths:
        - 'apps/backend/**'

  env:
    AWS_REGION: ap-northeast-2

  jobs:
    build-and-push:
      runs-on: ubuntu-latest
      permissions:
        id-token: write
        contents: read

      strategy:
        matrix:
          include:
            - name: api
              dockerfile: apps/backend/Dockerfile.api
            - name: worker
              dockerfile: apps/backend/Dockerfile.worker

      steps:
        - uses: actions/checkout@v4

        - name: Configure AWS credentials
          uses: aws-actions/configure-aws-credentials@v4
          with:
            role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
            aws-region: ${{ env.AWS_REGION }}

        - name: Login to Amazon ECR
          id: login-ecr
          uses: aws-actions/amazon-ecr-login@v2

        - name: Set up Docker Buildx
          uses: docker/setup-buildx-action@v3

        - name: Build and push
          uses: docker/build-push-action@v5
          with:
            context: apps/backend
            file: ${{ matrix.dockerfile }}
            push: true
            tags: |
              ${{ steps.login-ecr.outputs.registry }}/mindhit/${{ matrix.name }}:latest
              ${{ steps.login-ecr.outputs.registry }}/mindhit/${{ matrix.name }}:${{ github.sha }}
            cache-from: type=gha
            cache-to: type=gha,mode=max
  ```

- [ ] **Web App CI 워크플로우**

  `.github/workflows/webapp-ci.yml`:

  ```yaml
  name: Web App CI

  on:
    push:
      branches: [main, develop]
      paths:
        - 'apps/web/**'
    pull_request:
      branches: [main]
      paths:
        - 'apps/web/**'

  defaults:
    run:
      working-directory: apps/web

  jobs:
    build-and-test:
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
            version: 9

        - name: Get pnpm store directory
          id: pnpm-cache
          shell: bash
          run: echo "STORE_PATH=$(pnpm store path)" >> $GITHUB_OUTPUT

        - name: Setup pnpm cache
          uses: actions/cache@v4
          with:
            path: ${{ steps.pnpm-cache.outputs.STORE_PATH }}
            key: ${{ runner.os }}-pnpm-store-${{ hashFiles('**/pnpm-lock.yaml') }}

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
gh workflow view backend-ci

# 로컬 테스트
cd apps/backend && go test -v ./...
cd apps/web && pnpm test
```

---

## Step 13.3: 배포 자동화

### 목표

Helm을 통한 K8s 배포 자동화

### 체크리스트

- [ ] **Helm Values (Production)**

  `infra/helm/mindhit/values-prod.yaml`:

  ```yaml
  global:
    environment: production
    imageRegistry: "<AWS_ACCOUNT_ID>.dkr.ecr.ap-northeast-2.amazonaws.com"

  api:
    replicas: 2
    image:
      repository: mindhit/api
      tag: latest
    resources:
      requests:
        cpu: 250m
        memory: 256Mi
      limits:
        cpu: 500m
        memory: 512Mi
    env:
      - name: ENVIRONMENT
        value: production
      - name: DATABASE_URL
        valueFrom:
          secretKeyRef:
            name: mindhit-secrets
            key: database-url
      - name: REDIS_ADDR
        valueFrom:
          secretKeyRef:
            name: mindhit-secrets
            key: redis-addr
      - name: JWT_SECRET
        valueFrom:
          secretKeyRef:
            name: mindhit-secrets
            key: jwt-secret
    hpa:
      enabled: true
      minReplicas: 2
      maxReplicas: 10
      targetCPUUtilization: 70

  worker:
    replicas: 2
    image:
      repository: mindhit/worker
      tag: latest
    resources:
      requests:
        cpu: 500m
        memory: 512Mi
      limits:
        cpu: 1000m
        memory: 1Gi
    env:
      - name: ENVIRONMENT
        value: production
      - name: DATABASE_URL
        valueFrom:
          secretKeyRef:
            name: mindhit-secrets
            key: database-url
      - name: REDIS_ADDR
        valueFrom:
          secretKeyRef:
            name: mindhit-secrets
            key: redis-addr
      - name: OPENAI_API_KEY
        valueFrom:
          secretKeyRef:
            name: mindhit-secrets
            key: openai-api-key

  ingress:
    enabled: true
    className: alb
    annotations:
      alb.ingress.kubernetes.io/scheme: internet-facing
      alb.ingress.kubernetes.io/target-type: ip
      alb.ingress.kubernetes.io/certificate-arn: "<ACM_CERTIFICATE_ARN>"
      alb.ingress.kubernetes.io/ssl-policy: ELBSecurityPolicy-TLS-1-2-2017-01
    host: api.mindhit.io
  ```

- [ ] **배포 워크플로우**

  `.github/workflows/deploy.yml`:

  ```yaml
  name: Deploy

  on:
    workflow_run:
      workflows: ["Docker Build"]
      types: [completed]
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

  env:
    AWS_REGION: ap-northeast-2

  jobs:
    deploy:
      runs-on: ubuntu-latest
      if: ${{ github.event.workflow_run.conclusion == 'success' || github.event_name == 'workflow_dispatch' }}
      environment: ${{ github.event.inputs.environment || 'staging' }}

      steps:
        - uses: actions/checkout@v4

        - name: Configure AWS credentials
          uses: aws-actions/configure-aws-credentials@v4
          with:
            role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
            aws-region: ${{ env.AWS_REGION }}

        - name: Update kubeconfig
          run: |
            aws eks update-kubeconfig \
              --name mindhit-${{ github.event.inputs.environment || 'staging' }} \
              --region ${{ env.AWS_REGION }}

        - name: Install Helm
          uses: azure/setup-helm@v3
          with:
            version: v3.14.0

        - name: Deploy with Helm
          run: |
            helm upgrade --install mindhit ./infra/helm/mindhit \
              --namespace mindhit \
              --create-namespace \
              -f ./infra/helm/mindhit/values-${{ github.event.inputs.environment || 'staging' }}.yaml \
              --set api.image.tag=${{ github.sha }} \
              --set worker.image.tag=${{ github.sha }} \
              --wait \
              --timeout 10m

        - name: Verify deployment
          run: |
            kubectl rollout status deployment/mindhit-api -n mindhit
            kubectl rollout status deployment/mindhit-worker -n mindhit

        - name: Notify success
          if: success()
          uses: slackapi/slack-github-action@v1
          with:
            payload: |
              {
                "text": "✅ Deployed to ${{ github.event.inputs.environment || 'staging' }} successfully"
              }
          env:
            SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}

        - name: Notify failure
          if: failure()
          uses: slackapi/slack-github-action@v1
          with:
            payload: |
              {
                "text": "❌ Deploy to ${{ github.event.inputs.environment || 'staging' }} failed"
              }
          env:
            SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
  ```

- [ ] **롤백 스크립트**

  `infra/scripts/rollback.sh`:

  ```bash
  #!/bin/bash
  set -e

  NAMESPACE="mindhit"
  RELEASE="mindhit"

  echo "=== MindHit Rollback ==="
  echo "Current revision:"
  helm history $RELEASE -n $NAMESPACE | tail -3

  read -p "Enter revision number to rollback to: " REVISION

  helm rollback $RELEASE $REVISION -n $NAMESPACE --wait

  echo "Rollback completed!"
  kubectl rollout status deployment/mindhit-api -n $NAMESPACE
  kubectl rollout status deployment/mindhit-worker -n $NAMESPACE
  ```

### 검증

```bash
# Helm 배포 상태 확인
helm list -n mindhit

# Pod 상태 확인
kubectl get pods -n mindhit

# 로그 확인
kubectl logs -f deployment/mindhit-api -n mindhit
```

---

## Step 13.4: 운영 매뉴얼

### 체크리스트

- [ ] **운영 매뉴얼**

  `docs/operations/runbook.md`:

  ```markdown
  # MindHit 운영 매뉴얼

  ## 서비스 개요

  | 서비스 | URL | 설명 |
  |--------|-----|------|
  | API | api.mindhit.io | REST API |
  | Web | app.mindhit.io | Next.js 웹앱 |
  | Grafana | grafana.mindhit.io | 모니터링 |

  ## 인프라 구성

  - **EKS**: mindhit-production 클러스터
  - **RDS**: PostgreSQL 16, db.t3.small
  - **ElastiCache**: Redis 7, cache.t3.micro
  - **ECR**: mindhit/api, mindhit/worker

  ## 일상 점검

  ### 매일
  - Grafana 대시보드 확인
  - Slack 알림 확인
  - 리소스 사용량 확인

  ### 매주
  - 백업 상태 확인
  - 보안 업데이트 확인
  - 비용 리포트 확인

  ## 배포 절차

  1. `main` 브랜치에 PR 머지
  2. Docker Build 워크플로우 자동 실행
  3. Deploy 워크플로우 자동 실행
  4. Slack 알림 확인
  5. 배포 후 모니터링 (15분)

  ## 롤백 절차

  ```bash
  ./scripts/rollback.sh
  ```

  ## kubectl 명령어

  ```bash
  # kubeconfig 설정
  aws eks update-kubeconfig --name mindhit-production --region ap-northeast-2

  # Pod 확인
  kubectl get pods -n mindhit

  # 로그 확인
  kubectl logs -f deployment/mindhit-api -n mindhit

  # Pod 재시작
  kubectl rollout restart deployment/mindhit-api -n mindhit
  ```

  ```

- [ ] **장애 대응 가이드**

  `docs/operations/incident-response.md`:

  ```markdown
  # 장애 대응 가이드

  ## 심각도 정의

  | 레벨 | 정의 | 응답 시간 | 해결 시간 |
  |------|------|---------|---------|
  | P1 | 서비스 완전 중단 | 5분 | 1시간 |
  | P2 | 주요 기능 장애 | 15분 | 4시간 |
  | P3 | 일부 기능 장애 | 1시간 | 24시간 |
  | P4 | 경미한 이슈 | 4시간 | 1주 |

  ## 장애 대응 절차

  ### 1. 감지
  - Alertmanager 알림
  - 사용자 리포트
  - 모니터링 대시보드

  ### 2. 분류
  - 심각도 판단
  - 영향 범위 파악
  - 담당자 지정

  ### 3. 대응
  - 즉시 조치 (롤백, 스케일링)
  - 고객 커뮤니케이션
  - 실시간 상황 공유

  ### 4. 복구
  - 서비스 정상화 확인
  - 모니터링 강화

  ### 5. 사후 분석
  - 타임라인 정리
  - 근본 원인 분석
  - 재발 방지 대책

  ## 일반적인 장애 시나리오

  ### API Pod 무응답

  ```bash
  # Pod 상태 확인
  kubectl get pods -n mindhit

  # 로그 확인
  kubectl logs -f deployment/mindhit-api -n mindhit

  # Pod 재시작
  kubectl rollout restart deployment/mindhit-api -n mindhit
  ```

  ### 데이터베이스 연결 실패

  ```bash
  # RDS 상태 확인
  aws rds describe-db-instances --db-instance-identifier mindhit-production

  # 연결 테스트
  kubectl run -it --rm debug --image=postgres:16 -- psql "$DATABASE_URL"
  ```

  ### Worker 처리 지연

  ```bash
  # Worker 로그 확인
  kubectl logs -f deployment/mindhit-worker -n mindhit

  # Worker 스케일 업
  kubectl scale deployment/mindhit-worker -n mindhit --replicas=5
  ```

  ```

---

## Phase 13 완료 확인

### 전체 검증 체크리스트

- [ ] Terraform으로 인프라 프로비저닝 완료
- [ ] EKS 클러스터 접근 가능
- [ ] RDS, ElastiCache 연결 가능
- [ ] CI/CD 파이프라인 동작
- [ ] Helm 배포 성공
- [ ] 롤백 테스트 완료
- [ ] 운영 문서 작성 완료

### 테스트 요구사항

| 테스트 유형 | 대상 | 도구 |
| ----------- | ---- | ---- |
| IaC 검증 | Terraform 문법 | `terraform validate` |
| IaC 계획 | 인프라 변경 사항 | `terraform plan` |
| Helm 검증 | Chart 문법 | `helm lint` |
| E2E 테스트 | 배포 파이프라인 | GitHub Actions (staging) |
| Smoke 테스트 | 배포 후 헬스 체크 | `kubectl get pods` |

```bash
# Phase 13 검증
# 1. Terraform 검증
cd infra/environments/prod
terraform validate
terraform plan

# 2. Helm Chart 검증
helm lint infra/helm/mindhit

# 3. 전체 테스트 통과 확인
moonx backend:test
moonx web:test
moonx extension:test
```

> **Note**: Phase 13은 프로덕션 배포 전 모든 Phase의 테스트가 통과해야 합니다.

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| Terraform 모듈 | `infra/modules/` |
| Production 환경 | `infra/environments/prod/` |
| Backend CI | `.github/workflows/backend-ci.yml` |
| Docker Build | `.github/workflows/docker-build.yml` |
| Deploy | `.github/workflows/deploy.yml` |
| Helm Values | `infra/helm/mindhit/values-prod.yaml` |
| 롤백 스크립트 | `infra/scripts/rollback.sh` |
| 운영 매뉴얼 | `docs/operations/runbook.md` |
| 장애 대응 | `docs/operations/incident-response.md` |

---

## 다음 Phase

Phase 13 완료 후 [Phase 14: Stripe 결제 연동](./phase-14-billing.md)으로 진행하세요.
