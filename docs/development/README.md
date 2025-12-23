# MindHit 개발 문서

## 핵심 기술 문서

개발을 시작하기 전 아래 문서를 순서대로 읽어주세요.

| 순서 | 문서 | 설명 |
|------|------|------|
| 1 | [01-architecture.md](./01-architecture.md) | 시스템 아키텍처, 개발 환경, 배포 전략 |
| 2 | [02-data-structure.md](./02-data-structure.md) | 데이터베이스 스키마, ER 다이어그램 |
| 3 | [03-tracking-pipeline.md](./03-tracking-pipeline.md) | Extension → API → Worker 이벤트 파이프라인 |
| 4 | [04-api-spec.md](./04-api-spec.md) | REST API 엔드포인트 상세 명세 |
| 5 | [07-api-spec-workflow.md](./07-api-spec-workflow.md) | TypeSpec 기반 API 코드 자동 생성 |
| 6 | [08-naming-conventions.md](./08-naming-conventions.md) | 네이밍 규칙, 코드 스타일 가이드 |
| 7 | [09-error-handling.md](./09-error-handling.md) | 에러 처리, 로깅, 모니터링 표준 |

## Phase별 구현 가이드

상세한 구현 체크리스트는 [phases/](./phases/) 폴더를 참조하세요.

```
phases/
├── README.md                    # 전체 진행 상황
├── phase-0-dev-environment.md   # 개발 환경 설정
├── phase-1-project-init.md      # 프로젝트 초기화
├── phase-1.5-api-spec.md        # API 스펙 공통화
├── phase-2-auth.md              # 인증 시스템
├── phase-3-sessions.md          # 세션 관리
├── phase-4-events.md            # 이벤트 수집
├── phase-5-infra.md             # 모니터링 인프라
├── phase-6-worker.md            # Worker 및 Job Queue
├── phase-7-webapp.md            # Next.js 웹앱
├── phase-8-extension.md         # Chrome Extension
├── phase-9-plan-usage.md        # 플랜 및 사용량
├── phase-10-ai.md               # AI 마인드맵
├── phase-11-dashboard.md        # 대시보드
├── phase-12-monitoring.md       # 모니터링 시스템
├── phase-13-deployment.md       # 배포 및 운영
└── phase-14-billing.md          # Stripe 결제
```

## 문서 관계

```
01-architecture.md      ← 전체 시스템 설계
        ↓
02-data-structure.md    ← DB 스키마 정의
        ↓
03-tracking-pipeline.md ← 데이터 흐름 상세
        ↓
04-api-spec.md          ← API 인터페이스
        ↓
07-api-spec-workflow.md ← API 코드 생성 방법
        ↓
08-naming-conventions.md ← 네이밍 규칙
        ↓
09-error-handling.md    ← 에러 처리 표준
        ↓
phases/                 ← 단계별 구현 가이드
```

## Deprecated 문서

| 문서 | 상태 | 대체 문서 |
|------|------|----------|
| [05-milestones.md](./05-milestones.md) | Deprecated | [phases/README.md](./phases/README.md) |
| [06-implementation-phases.md](./06-implementation-phases.md) | Deprecated | [phases/README.md](./phases/README.md) |

## 시작하기

```bash
# 1. 개발 환경 설정
cd /path/to/mindhit
pnpm install

# 2. 인프라 실행 (PostgreSQL, Redis)
docker compose up -d

# 3. 백엔드 개발 서버
moonx backend:dev-api

# 4. 웹 개발 서버
moonx web:dev
```

자세한 내용은 [Phase 0: 개발 환경](./phases/phase-0-dev-environment.md)을 참조하세요.
