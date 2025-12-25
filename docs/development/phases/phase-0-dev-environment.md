# Phase 0: 개발 환경 구성

## 개요

| 항목 | 내용 |
| --- | --- |
| **목표** | 3단계 개발 환경 구성 (go run → 로컬 K8s → 프로덕션) |
| **선행 조건** | Docker, kind, kubectl, Helm, Go, Node.js, pnpm 설치 |
| **예상 소요** | 6 Steps |
| **결과물** | 빠른 개발(go run)과 프로덕션 검증(kind) 환경 동작 |

---

## 3단계 개발 환경 개요

| 단계 | 용도 | 실행 방식 | Database | Redis |
| --- | --- | --- | --- | --- |
| **1. go run** | 빠른 개발/디버깅 | `moonx backend:dev-api` | Docker Compose | Docker Compose |
| **2. 로컬 K8s** | 프로덕션 환경 검증 | `moonx infra:kind-deploy` | K8s StatefulSet | K8s StatefulSet |
| **3. 프로덕션** | 실서비스 | GitHub Actions + Helm | RDS (관리형) | ElastiCache (관리형) |

### 왜 3단계인가?

- **go run 모드**: Hot reload, 디버거 연결 가능, 빠른 피드백 루프
- **로컬 K8s (kind)**: Dockerfile, Helm Chart, K8s 매니페스트 검증
- **프로덕션**: 관리형 서비스 활용, 고가용성

```mermaid
flowchart LR
    subgraph Stage1[1. go run 모드]
        DEV[개발자]
        GO[go run / air]
        DC[Docker Compose<br/>PostgreSQL + Redis]
    end

    subgraph Stage2[2. 로컬 K8s]
        KIND[kind Cluster]
        HELM1[Helm Charts]
        LOCAL_DB[(PostgreSQL<br/>StatefulSet)]
    end

    subgraph Stage3[3. 프로덕션]
        EKS[EKS Cluster]
        HELM2[Helm Charts]
        RDS[(RDS)]
        ELASTICACHE[(ElastiCache)]
    end

    DEV -->|빠른 개발| GO
    GO --> DC
    KIND -->|검증| HELM1
    HELM1 --> LOCAL_DB
    EKS -->|배포| HELM2
    HELM2 --> RDS
    HELM2 --> ELASTICACHE

    Stage1 -->|"Dockerfile 검증"| Stage2
    Stage2 -->|"Helm Chart 검증"| Stage3
```

---

## 진행 상황

| Step | 이름 | 상태 |
| --- | --- | --- |
| 0.1 | 필수 도구 설치 | ✅ |
| 0.2 | Moon 설치 및 모노레포 설정 | ✅ |
| 0.3 | Docker Compose 설정 (go run 모드용) | ✅ |
| 0.4 | 로컬 K8s (kind) 설정 | ✅ |
| 0.5 | Helm Chart 구성 | ✅ |
| 0.6 | Moon 태스크 설정 | ✅ |

---

## Step 0.1: 필수 도구 설치

### 목표

개발에 필요한 모든 도구 설치 및 버전 확인

### 필수 도구

| 도구 | 버전 | 용도 |
| --- | --- | --- |
| Docker | 24+ | 컨테이너 런타임 |
| kind | 0.22+ | 로컬 K8s 클러스터 (Kubernetes IN Docker) |
| kubectl | 1.29+ | K8s CLI |
| Helm | 3.14+ | K8s 패키지 매니저 |
| Go | 1.22+ | Backend 개발 |
| Node.js | 20+ | Web/Extension 개발 |
| pnpm | 9+ | Node.js 패키지 매니저 |

### 체크리스트

- [x] **Docker 설치**

  ```bash
  # macOS
  brew install --cask docker

  # 또는 Docker Desktop 설치
  # https://www.docker.com/products/docker-desktop

  docker --version
  ```

- [x] **kind 설치**

  ```bash
  # macOS
  brew install kind

  # Linux
  curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.22.0/kind-linux-amd64
  chmod +x ./kind
  sudo mv ./kind /usr/local/bin/kind

  kind version
  ```

- [x] **kubectl 설치**

  ```bash
  # macOS
  brew install kubectl

  # Linux
  curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
  sudo install kubectl /usr/local/bin/kubectl

  kubectl version --client
  ```

- [x] **Helm 설치**

  ```bash
  # macOS
  brew install helm

  # Linux
  curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

  helm version
  ```

- [x] **Go 설치**

  ```bash
  # macOS
  brew install go

  # Linux - https://go.dev/dl/
  wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
  sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
  export PATH=$PATH:/usr/local/go/bin

  go version
  ```

- [x] **Node.js + pnpm 설치**

  ```bash
  # Node.js (nvm 권장)
  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
  nvm install 20
  nvm use 20

  # pnpm
  npm install -g pnpm

  node --version
  pnpm --version
  ```

- [x] **추가 도구 설치**

  ```bash
  # golangci-lint (Go 린터)
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

  # Atlas (DB 마이그레이션)
  curl -sSf https://atlasgo.sh | sh

  # Air (Go Hot Reload) - go run 모드용
  go install github.com/air-verse/air@latest
  ```

### 검증

```bash
# 모든 도구 버전 확인
docker --version
kind version
kubectl version --client
helm version
go version
node --version
pnpm --version
```

---

## Step 0.2: Moon 설치 및 모노레포 설정

### 목표

Moon 태스크 러너 설치 및 모노레포 기본 설정

### 체크리스트

- [x] **pnpm-workspace.yaml 생성** (루트)

  ```yaml
  packages:
    - 'apps/*'
    - 'packages/*'
  ```

- [x] **루트 package.json 생성**

  ```json
  {
    "name": "mindhit",
    "private": true,
    "packageManager": "pnpm@9.0.0",
    "scripts": {
      "dev": "moonx :dev",
      "build": "moonx :build",
      "test": "moonx :test",
      "lint": "moonx :lint",
      "ci": "moon ci"
    },
    "devDependencies": {
      "@moonrepo/cli": "^1.29.0"
    }
  }
  ```

- [x] **Moon 설치**

  ```bash
  pnpm install
  ```

- [x] **.moon/workspace.yml 생성**

  ```yaml
  $schema: 'https://moonrepo.dev/schemas/workspace.json'

  projects:
    - 'apps/*'
    - 'packages/*'
    - 'infra'

  vcs:
    manager: git
    defaultBranch: main
    remoteCandidates:
      - origin

  versionConstraint: '>=1.28.0'
  ```

- [x] **.moon/toolchain.yml 생성**

  ```yaml
  $schema: 'https://moonrepo.dev/schemas/toolchain.json'

  node:
    version: '20.10.0'
    packageManager: pnpm
    pnpm:
      version: '9.0.0'

  typescript:
    syncProjectReferences: true
  ```

- [x] **디렉토리 구조 생성**

  ```bash
  mkdir -p apps/{backend,web,extension}
  mkdir -p apps/backend/{cmd,internal,pkg}
  mkdir -p apps/backend/cmd/{api,worker}
  mkdir -p packages/{shared,protocol}
  mkdir -p infra/{helm,docker,kind,modules,environments,scripts}
  mkdir -p .github/workflows
  ```

### 결과물

```text
mindhit/
├── .moon/
│   ├── workspace.yml
│   └── toolchain.yml
├── apps/
│   ├── backend/          # Go 백엔드
│   ├── web/              # Next.js
│   └── extension/        # Chrome Extension
├── packages/
│   ├── shared/           # TS 공유
│   └── protocol/         # API 타입
├── infra/
│   ├── docker/           # Docker Compose (go run 모드)
│   ├── kind/             # kind 클러스터 설정
│   ├── helm/             # Helm Charts
│   ├── modules/          # Terraform 모듈
│   └── scripts/
├── package.json
└── pnpm-workspace.yaml
```

---

## Step 0.3: Docker Compose 설정 (go run 모드용)

### 목표

`go run` 모드에서 사용할 PostgreSQL, Redis를 Docker Compose로 구성

### 체크리스트

- [x] **docker-compose.yml 생성**

  ```yaml
  # infra/docker/docker-compose.yml
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
      healthcheck:
        test: ["CMD", "redis-cli", "ping"]
        interval: 5s
        timeout: 5s
        retries: 5

    # Asynqmon (선택사항) - Asynq 모니터링 UI
    asynqmon:
      image: hibiken/asynqmon:latest
      container_name: mindhit-asynqmon
      ports:
        - "8080:8080"
      command:
        - "--redis-addr=redis:6379"
      depends_on:
        - redis

  volumes:
    postgres_data:
    redis_data:
  ```

- [x] **.env.local 생성** (apps/backend/)

  ```bash
  # apps/backend/.env.local
  ENVIRONMENT=local

  # Database
  DATABASE_URL=postgres://postgres:password@localhost:5432/mindhit?sslmode=disable

  # Redis
  REDIS_URL=redis://localhost:6379

  # Server
  API_PORT=8081
  WORKER_CONCURRENCY=10

  # AI (개발용 - 필요시 설정)
  # OPENAI_API_KEY=sk-...
  # GEMINI_API_KEY=...
  # ANTHROPIC_API_KEY=sk-...
  ```

### 검증

```bash
# Docker Compose 시작
cd infra/docker
docker compose up -d

# 상태 확인
docker compose ps

# PostgreSQL 접속 테스트
docker exec -it mindhit-postgres psql -U postgres -d mindhit

# Redis 접속 테스트
docker exec -it mindhit-redis redis-cli ping

# Asynqmon 접속 (브라우저)
# http://localhost:8080
```

---

## Step 0.4: 로컬 K8s (kind) 설정

### 목표

kind로 로컬 K8s 클러스터 구성 (프로덕션 환경 검증용)

### 왜 kind인가?

| 항목 | kind | minikube | k3d |
| --- | --- | --- | --- |
| 기반 | 표준 K8s (Docker) | VM 또는 Docker | k3s (경량) |
| EKS 호환성 | 높음 | 보통 | 낮음 |
| 시작 속도 | 빠름 | 느림 | 빠름 |
| 리소스 사용 | 적음 | 많음 | 적음 |
| 안정성 | 높음 | 높음 | 보통 |

**kind 선택 이유**: EKS와 동일한 표준 Kubernetes이므로 프로덕션 환경 검증에 가장 적합

### 체크리스트

- [x] **kind 클러스터 설정 파일 생성**

  ```yaml
  # infra/kind/kind-config.yaml
  kind: Cluster
  apiVersion: kind.x-k8s.io/v1alpha4
  nodes:
    - role: control-plane
      kubeadmConfigPatches:
        - |
          kind: InitConfiguration
          nodeRegistration:
            kubeletExtraArgs:
              node-labels: "ingress-ready=true"
      extraPortMappings:
        - containerPort: 80
          hostPort: 80
          protocol: TCP
        - containerPort: 443
          hostPort: 443
          protocol: TCP
  ```

- [x] **kind 클러스터 생성 스크립트**

  ```bash
  # infra/kind/setup.sh
  #!/bin/bash
  set -e

  CLUSTER_NAME="mindhit"

  echo "=== MindHit Local K8s (kind) Setup ==="

  # 기존 클러스터 삭제 (있으면)
  kind delete cluster --name $CLUSTER_NAME 2>/dev/null || true

  # 클러스터 생성
  echo "Creating kind cluster..."
  kind create cluster --name $CLUSTER_NAME --config infra/kind/kind-config.yaml

  # kubectl 컨텍스트 설정
  kubectl cluster-info --context kind-$CLUSTER_NAME

  # NGINX Ingress Controller 설치
  echo "Installing NGINX Ingress Controller..."
  kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

  # Ingress 준비 대기
  echo "Waiting for Ingress Controller to be ready..."
  kubectl wait --namespace ingress-nginx \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=90s

  echo ""
  echo "=== Kind cluster ready! ==="
  echo ""
  echo "Add to /etc/hosts:"
  echo "127.0.0.1 api.mindhit.local"
  echo ""
  echo "Run 'moonx infra:kind-deploy' to deploy"
  ```

- [x] **실행 권한 부여**

  ```bash
  chmod +x infra/kind/setup.sh
  ```

### 검증

```bash
# 클러스터 생성
./infra/kind/setup.sh

# 클러스터 상태 확인
kubectl get nodes
kubectl get pods -A

# 클러스터 삭제
kind delete cluster --name mindhit
```

---

## Step 0.5: Helm Chart 구성

### 목표

MindHit 애플리케이션용 Helm Chart 생성

### 체크리스트

- [x] **Helm Chart 디렉토리 생성**

  ```bash
  mkdir -p infra/helm/mindhit/templates
  ```

- [x] **Chart.yaml 생성**

  ```yaml
  # infra/helm/mindhit/Chart.yaml
  apiVersion: v2
  name: mindhit
  description: MindHit - AI-powered browsing tracker
  type: application
  version: 0.1.0
  appVersion: "1.0.0"

  dependencies:
    - name: postgresql
      version: "15.2.0"
      repository: "https://charts.bitnami.com/bitnami"
      condition: postgresql.enabled
    - name: redis
      version: "19.0.0"
      repository: "https://charts.bitnami.com/bitnami"
      condition: redis.enabled
  ```

- [x] **values.yaml 생성 (기본값)**

  ```yaml
  # infra/helm/mindhit/values.yaml
  global:
    environment: development

  api:
    replicaCount: 1
    image:
      repository: mindhit/api
      tag: latest
      pullPolicy: IfNotPresent
    service:
      type: ClusterIP
      port: 8080
    resources:
      requests:
        memory: "256Mi"
        cpu: "250m"
      limits:
        memory: "512Mi"
        cpu: "500m"

  worker:
    replicaCount: 1
    image:
      repository: mindhit/worker
      tag: latest
      pullPolicy: IfNotPresent
    resources:
      requests:
        memory: "512Mi"
        cpu: "500m"
      limits:
        memory: "1Gi"
        cpu: "1000m"

  postgresql:
    enabled: true
    auth:
      database: mindhit
      username: postgres
      password: password
    primary:
      persistence:
        size: 1Gi

  redis:
    enabled: true
    auth:
      enabled: false
    master:
      persistence:
        size: 1Gi

  ingress:
    enabled: true
    className: nginx
    hosts:
      - host: api.mindhit.local
        paths:
          - path: /
            pathType: Prefix
  ```

- [x] **values-local.yaml 생성 (로컬 K8s 오버라이드)**

  ```yaml
  # infra/helm/mindhit/values-local.yaml
  global:
    environment: local

  api:
    replicaCount: 1
    image:
      pullPolicy: Never  # kind에 로드된 로컬 이미지 사용

  worker:
    replicaCount: 1
    image:
      pullPolicy: Never
  ```

- [x] **API Deployment 템플릿 생성**

  ```yaml
  # infra/helm/mindhit/templates/api-deployment.yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: {{ .Release.Name }}-api
    labels:
      app: {{ .Release.Name }}-api
  spec:
    replicas: {{ .Values.api.replicaCount }}
    selector:
      matchLabels:
        app: {{ .Release.Name }}-api
    template:
      metadata:
        labels:
          app: {{ .Release.Name }}-api
      spec:
        containers:
          - name: api
            image: "{{ .Values.api.image.repository }}:{{ .Values.api.image.tag }}"
            imagePullPolicy: {{ .Values.api.image.pullPolicy }}
            ports:
              - containerPort: 8080
            env:
              - name: ENVIRONMENT
                value: {{ .Values.global.environment }}
              - name: DATABASE_URL
                value: "postgres://{{ .Values.postgresql.auth.username }}:{{ .Values.postgresql.auth.password }}@{{ .Release.Name }}-postgresql:5432/{{ .Values.postgresql.auth.database }}?sslmode=disable"
              - name: REDIS_URL
                value: "redis://{{ .Release.Name }}-redis-master:6379"
            resources:
              {{- toYaml .Values.api.resources | nindent 14 }}
  ```

- [x] **Worker Deployment 템플릿 생성**

  ```yaml
  # infra/helm/mindhit/templates/worker-deployment.yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: {{ .Release.Name }}-worker
    labels:
      app: {{ .Release.Name }}-worker
  spec:
    replicas: {{ .Values.worker.replicaCount }}
    selector:
      matchLabels:
        app: {{ .Release.Name }}-worker
    template:
      metadata:
        labels:
          app: {{ .Release.Name }}-worker
      spec:
        containers:
          - name: worker
            image: "{{ .Values.worker.image.repository }}:{{ .Values.worker.image.tag }}"
            imagePullPolicy: {{ .Values.worker.image.pullPolicy }}
            env:
              - name: ENVIRONMENT
                value: {{ .Values.global.environment }}
              - name: DATABASE_URL
                value: "postgres://{{ .Values.postgresql.auth.username }}:{{ .Values.postgresql.auth.password }}@{{ .Release.Name }}-postgresql:5432/{{ .Values.postgresql.auth.database }}?sslmode=disable"
              - name: REDIS_URL
                value: "redis://{{ .Release.Name }}-redis-master:6379"
            resources:
              {{- toYaml .Values.worker.resources | nindent 14 }}
  ```

- [x] **API Service 템플릿 생성**

  ```yaml
  # infra/helm/mindhit/templates/api-service.yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: {{ .Release.Name }}-api
  spec:
    type: {{ .Values.api.service.type }}
    ports:
      - port: {{ .Values.api.service.port }}
        targetPort: 8080
        protocol: TCP
    selector:
      app: {{ .Release.Name }}-api
  ```

- [x] **Ingress 템플릿 생성**

  ```yaml
  # infra/helm/mindhit/templates/ingress.yaml
  {{- if .Values.ingress.enabled }}
  apiVersion: networking.k8s.io/v1
  kind: Ingress
  metadata:
    name: {{ .Release.Name }}-ingress
    annotations:
      nginx.ingress.kubernetes.io/rewrite-target: /
  spec:
    ingressClassName: {{ .Values.ingress.className }}
    rules:
      {{- range .Values.ingress.hosts }}
      - host: {{ .host }}
        http:
          paths:
            {{- range .paths }}
            - path: {{ .path }}
              pathType: {{ .pathType }}
              backend:
                service:
                  name: {{ $.Release.Name }}-api
                  port:
                    number: {{ $.Values.api.service.port }}
            {{- end }}
      {{- end }}
  {{- end }}
  ```

- [x] **Helm 의존성 업데이트**

  ```bash
  cd infra/helm/mindhit
  helm dependency update
  ```

### 검증

```bash
# Chart 검증
helm lint ./infra/helm/mindhit

# 템플릿 렌더링 테스트
helm template mindhit ./infra/helm/mindhit -f ./infra/helm/mindhit/values-local.yaml
```

---

## Step 0.6: Moon 태스크 설정

### 목표

Moon을 이용한 개발 워크플로우 자동화

### 체크리스트

- [x] **infra/moon.yml 생성**

  ```yaml
  # infra/moon.yml
  $schema: 'https://moonrepo.dev/schemas/project.json'

  type: tool

  tasks:
    # ===================
    # go run 모드 (빠른 개발)
    # ===================
    dev-up:
      command: docker
      args:
        - compose
        - -f
        - infra/docker/docker-compose.yml
        - up
        - -d
      options:
        runFromWorkspaceRoot: true

    dev-down:
      command: docker
      args:
        - compose
        - -f
        - infra/docker/docker-compose.yml
        - down
      options:
        runFromWorkspaceRoot: true

    dev-logs:
      command: docker
      args:
        - compose
        - -f
        - infra/docker/docker-compose.yml
        - logs
        - -f
      options:
        runFromWorkspaceRoot: true

    # ===================
    # 로컬 K8s (kind)
    # ===================
    kind-up:
      command: bash
      args:
        - infra/kind/setup.sh
      options:
        runFromWorkspaceRoot: true

    kind-down:
      command: kind
      args:
        - delete
        - cluster
        - --name
        - mindhit

    kind-build:
      command: bash
      args:
        - -c
        - |
          set -e
          echo "Building API image..."
          docker build -f apps/backend/Dockerfile.api -t mindhit/api:latest apps/backend
          echo "Building Worker image..."
          docker build -f apps/backend/Dockerfile.worker -t mindhit/worker:latest apps/backend
          echo "Loading images to kind..."
          kind load docker-image mindhit/api:latest --name mindhit
          kind load docker-image mindhit/worker:latest --name mindhit
          echo "Done!"
      options:
        runFromWorkspaceRoot: true

    kind-deploy:
      command: bash
      args:
        - -c
        - |
          set -e
          cd infra/helm/mindhit
          helm dependency update
          cd ../../..
          helm upgrade --install mindhit ./infra/helm/mindhit -f ./infra/helm/mindhit/values-local.yaml
      options:
        runFromWorkspaceRoot: true
      deps:
        - kind-build

    kind-redeploy:
      command: bash
      args:
        - -c
        - |
          set -e
          echo "Building and loading images..."
          docker build -f apps/backend/Dockerfile.api -t mindhit/api:latest apps/backend
          docker build -f apps/backend/Dockerfile.worker -t mindhit/worker:latest apps/backend
          kind load docker-image mindhit/api:latest --name mindhit
          kind load docker-image mindhit/worker:latest --name mindhit
          echo "Restarting deployments..."
          kubectl rollout restart deployment/mindhit-api deployment/mindhit-worker
          kubectl rollout status deployment/mindhit-api
          kubectl rollout status deployment/mindhit-worker
      options:
        runFromWorkspaceRoot: true

    kind-logs-api:
      command: kubectl
      args:
        - logs
        - -f
        - -l
        - app=mindhit-api

    kind-logs-worker:
      command: kubectl
      args:
        - logs
        - -f
        - -l
        - app=mindhit-worker

    kind-shell-db:
      command: bash
      args:
        - -c
        - kubectl exec -it $(kubectl get pod -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}') -- psql -U postgres -d mindhit
      options:
        shell: true

    kind-shell-redis:
      command: bash
      args:
        - -c
        - kubectl exec -it $(kubectl get pod -l app.kubernetes.io/name=redis -o jsonpath='{.items[0].metadata.name}') -- redis-cli
      options:
        shell: true
  ```

- [x] **apps/backend/moon.yml 생성**

  ```yaml
  # apps/backend/moon.yml
  $schema: 'https://moonrepo.dev/schemas/project.json'

  type: application
  language: go

  tasks:
    # go run 모드
    dev-api:
      command: air
      args:
        - -c
        - .air.api.toml
      env:
        ENVIRONMENT: local

    dev-worker:
      command: air
      args:
        - -c
        - .air.worker.toml
      env:
        ENVIRONMENT: local

    # 빌드
    build-api:
      command: go
      args:
        - build
        - -o
        - bin/api
        - ./cmd/api

    build-worker:
      command: go
      args:
        - build
        - -o
        - bin/worker
        - ./cmd/worker

    # 테스트
    test:
      command: go
      args:
        - test
        - -v
        - ./...

    lint:
      command: golangci-lint
      args:
        - run
        - ./...

    # DB 마이그레이션
    migrate:
      command: atlas
      args:
        - migrate
        - apply
        - --env
        - local
  ```

- [x] **Air 설정 파일 생성** (Hot reload)

  ```toml
  # apps/backend/.air.api.toml
  root = "."
  tmp_dir = "tmp"

  [build]
  cmd = "go build -o ./tmp/api ./cmd/api"
  bin = "./tmp/api"
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["tmp", "vendor", "node_modules"]
  delay = 1000

  [log]
  time = false

  [color]
  main = "magenta"
  watcher = "cyan"
  build = "yellow"
  runner = "green"
  ```

  ```toml
  # apps/backend/.air.worker.toml
  root = "."
  tmp_dir = "tmp"

  [build]
  cmd = "go build -o ./tmp/worker ./cmd/worker"
  bin = "./tmp/worker"
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["tmp", "vendor", "node_modules"]
  delay = 1000

  [log]
  time = false

  [color]
  main = "magenta"
  watcher = "cyan"
  build = "yellow"
  runner = "green"
  ```

### 검증

```bash
# go run 모드 테스트
moonx infra:dev-up         # Docker Compose 시작
moonx backend:dev-api      # API 서버 시작 (터미널 1)
moonx backend:dev-worker   # Worker 시작 (터미널 2)
moonx infra:dev-down       # Docker Compose 종료

# kind 모드 테스트
moonx infra:kind-up        # kind 클러스터 생성
moonx infra:kind-deploy    # 빌드 + 배포
moonx infra:kind-logs-api  # 로그 확인
moonx infra:kind-down      # 클러스터 삭제
```

---

## Phase 0 완료 확인

### 전체 검증 체크리스트

#### go run 모드

- [x] **Docker Compose 동작**

  ```bash
  moonx infra:dev-up
  docker compose -f infra/docker/docker-compose.yml ps
  # postgres, redis 모두 running
  ```

- [ ] **API 서버 실행** (Go 코드 작성 후)

  ```bash
  moonx backend:dev-api
  curl http://localhost:8081/health
  ```

- [ ] **Worker 실행**

  ```bash
  moonx backend:dev-worker
  ```

#### 로컬 K8s (kind) 모드

- [ ] **kind 클러스터 동작**

  ```bash
  moonx infra:kind-up
  kubectl get nodes
  ```

- [ ] **모든 Pod Running 상태**

  ```bash
  kubectl get pods
  # mindhit-api, mindhit-worker, postgresql, redis 모두 Running
  ```

- [ ] **API 서버 응답**

  ```bash
  curl http://api.mindhit.local/health
  # {"status":"ok"}
  ```

### 테스트 요구사항

| 테스트 유형 | 대상 | 검증 방법 |
| ----------- | ---- | --------- |
| 인프라 검증 | Docker Compose | `docker compose ps` (모든 서비스 running) |
| 인프라 검증 | kind 클러스터 | `kubectl get nodes` (Ready 상태) |
| 연결 테스트 | PostgreSQL | `psql` 접속 확인 |
| 연결 테스트 | Redis | `redis-cli ping` |

```bash
# Phase 0 환경 검증
moonx infra:dev-up
docker compose -f infra/docker/docker-compose.yml ps
```

> **Note**: Phase 0은 개발 환경 설정이므로 애플리케이션 테스트가 아닌 인프라 동작 검증에 집중합니다.

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| Moon Workspace | `.moon/workspace.yml` |
| Moon Toolchain | `.moon/toolchain.yml` |
| Infra Moon Tasks | `infra/moon.yml` |
| Backend Moon Tasks | `apps/backend/moon.yml` |
| Docker Compose | `infra/docker/docker-compose.yml` |
| kind 설정 | `infra/kind/kind-config.yaml` |
| kind 설정 스크립트 | `infra/kind/setup.sh` |
| Helm Chart | `infra/helm/mindhit/` |
| API Dockerfile | `apps/backend/Dockerfile.api` |
| Worker Dockerfile | `apps/backend/Dockerfile.worker` |
| Air 설정 (Hot reload) | `apps/backend/.air.*.toml` |

### 주요 명령어

| 명령어 | 설명 |
| --- | --- |
| **go run 모드** | |
| `moonx infra:dev-up` | Docker Compose 시작 (DB, Redis) |
| `moonx backend:dev-api` | API 서버 실행 (Hot reload) |
| `moonx backend:dev-worker` | Worker 실행 (Hot reload) |
| `moonx infra:dev-down` | Docker Compose 종료 |
| **로컬 K8s (kind)** | |
| `moonx infra:kind-up` | kind 클러스터 생성 |
| `moonx infra:kind-deploy` | 빌드 + Helm 배포 |
| `moonx infra:kind-redeploy` | 코드 변경 후 재배포 |
| `moonx infra:kind-logs-api` | API 로그 확인 |
| `moonx infra:kind-down` | 클러스터 삭제 |

### 개발 흐름 권장 사항

1. **일상적인 개발**: `go run` 모드 사용
   - Hot reload로 빠른 피드백
   - 디버거 연결 용이
   - 가벼운 리소스 사용

2. **배포 전 검증**: kind 환경에서 테스트
   - Dockerfile 검증
   - Helm Chart 검증
   - K8s 매니페스트 검증
   - 프로덕션과 유사한 환경

3. **PR 머지 후**: GitHub Actions에서 자동 프로덕션 배포

---

## 다음 Phase

Phase 0 완료 후 [Phase 1: 프로젝트 초기화](./phase-1-project-init.md)로 진행하세요.
