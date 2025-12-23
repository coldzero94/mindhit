# 구현 단계 상세

> **Note**: 이 문서는 더 이상 사용되지 않습니다.
> 상세 구현 가이드는 [phases/](./phases/) 폴더를 참조하세요.

## 새로운 문서 구조

각 Phase는 별도의 문서로 분리되어 더 상세한 체크리스트와 코드 예제를 제공합니다.

| Phase | 문서 | 설명 |
|-------|------|------|
| 0 | [개발 환경](./phases/phase-0-dev-environment.md) | K8s + Helm 로컬 개발 환경 |
| 1 | [프로젝트 초기화](./phases/phase-1-project-init.md) | Monorepo, Go 백엔드 기본 설정 |
| 1.5 | [API 스펙 공통화](./phases/phase-1.5-api-spec.md) | OpenAPI, 공통 응답 포맷 |
| 2 | [인증 시스템](./phases/phase-2-auth.md) | JWT, OAuth2 |
| 3 | [세션 관리 API](./phases/phase-3-sessions.md) | 세션 CRUD API |
| 4 | [이벤트 수집 API](./phases/phase-4-events.md) | 이벤트 배치 수집 |
| 5 | [모니터링 및 인프라](./phases/phase-5-infra.md) | Prometheus, Grafana |
| 6 | [Worker 및 Job Queue](./phases/phase-6-worker.md) | Asynq + Redis 비동기 작업 |
| 7 | [Next.js 웹앱](./phases/phase-7-webapp.md) | 웹 프론트엔드 |
| 8 | [Chrome Extension](./phases/phase-8-extension.md) | 브라우저 확장 프로그램 |
| 9 | [플랜 및 사용량 시스템](./phases/phase-9-plan-usage.md) | 구독 플랜, 토큰 사용량 추적 |
| 10 | [AI 마인드맵](./phases/phase-10-ai.md) | 다중 AI Provider, Worker 연동 |
| 11 | [웹앱 대시보드](./phases/phase-11-dashboard.md) | 마인드맵 시각화 |
| 12 | [모니터링 시스템](./phases/phase-12-monitoring.md) | 알림, 로깅 |
| 13 | [배포 및 운영](./phases/phase-13-deployment.md) | Terraform, EKS, CI/CD |
| 14 | [Stripe 결제 연동](./phases/phase-14-billing.md) | 결제 시스템 |

## 아키텍처 개요

### 인프라

- **로컬 (go run)**: Docker Compose (DB, Redis)
- **로컬 (K8s)**: kind + Helm
- **프로덕션**: EKS + Terraform + 관리형 서비스 (RDS, ElastiCache)

### 백엔드 구조

```
apps/backend/
├── cmd/
│   ├── api/main.go       # API 서버 엔트리포인트
│   └── worker/main.go    # Worker 서버 엔트리포인트
├── internal/
│   ├── api/              # API 전용 코드 (controller, middleware, router)
│   └── worker/           # Worker 전용 코드 (handler)
├── pkg/                  # 공유 코드
│   ├── config/           # 설정
│   ├── ent/              # Ent ORM 스키마
│   ├── service/          # 비즈니스 로직
│   └── infra/            # 인프라 (queue, ai, etc.)
├── Dockerfile.api
├── Dockerfile.worker
└── go.mod
```

### 주요 기술 스택

| 영역 | 기술 |
|------|------|
| 백엔드 | Go, Gin, Ent |
| 프론트엔드 | Next.js, React, TypeScript |
| 데이터베이스 | PostgreSQL |
| 캐시/큐 | Redis + Asynq |
| AI | OpenAI, Gemini, Claude (다중 Provider) |
| 인프라 | Kubernetes, Helm, Terraform |
| CI/CD | GitHub Actions |

## 시작하기

1. [Phase 0: 개발 환경](./phases/phase-0-dev-environment.md)부터 시작하세요.
2. 각 Phase 문서의 체크리스트를 따라 진행하세요.
3. [phases/README.md](./phases/README.md)에서 전체 진행 상황을 관리하세요.
