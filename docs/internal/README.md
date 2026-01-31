# MindHit 문서

MindHit은 브라우저 히스토리를 AI 마인드맵으로 변환하는 서비스입니다.

## 문서 구조

### [product/](./product/)

제품 및 사용자 관점의 문서입니다.

| 문서                                               | 설명                   |
| -------------------------------------------------- | ---------------------- |
| [01-overview.md](./product/01-overview.md)         | 제품 개요 및 문제 정의 |
| [02-user-flow.md](./product/02-user-flow.md)       | 사용자 여정 및 플로우  |
| [03-features.md](./product/03-features.md)         | 기능 명세 및 우선순위  |
| [04-ui-wireframe.md](./product/04-ui-wireframe.md) | UI/UX 와이어프레임     |
| [05-pricing.md](./product/05-pricing.md)           | 가격 정책 및 빌링      |

### [development/](./development/)

기술 및 구현 관점의 문서입니다.

| 문서 | 설명 |
| ---- | ---- |
| [01-architecture.md](./development/01-architecture.md) | 시스템 아키텍처 (필독) |
| [02-data-structure.md](./development/02-data-structure.md) | 데이터베이스 설계 |
| [03-tracking-pipeline.md](./development/03-tracking-pipeline.md) | 이벤트 수집 파이프라인 |
| [04-api-spec.md](./development/04-api-spec.md) | REST API 명세 |
| [07-api-spec-workflow.md](./development/07-api-spec-workflow.md) | API 자동 생성 워크플로우 |
| [08-naming-conventions.md](./development/08-naming-conventions.md) | 네이밍 규칙, 코드 스타일 |
| [09-error-handling.md](./development/09-error-handling.md) | 에러 처리, 로깅 표준 |
| [phases/](./development/phases/) | Phase별 구현 가이드 |

## 시작하기

### 제품 이해

1. [product/01-overview.md](./product/01-overview.md) - 제품이 해결하는 문제
2. [product/02-user-flow.md](./product/02-user-flow.md) - 사용자 경험 흐름

### 개발 시작

1. [development/01-architecture.md](./development/01-architecture.md) - 아키텍처 파악
2. [development/phases/README.md](./development/phases/README.md) - Phase별 구현 진행

## 기술 스택

| 영역         | 기술                                    |
| ------------ | --------------------------------------- |
| 백엔드       | Go, Gin, Ent                            |
| 프론트엔드   | Next.js, React, TypeScript              |
| Extension    | Chrome Extension (Manifest V3)          |
| 데이터베이스 | PostgreSQL                              |
| 캐시/큐      | Redis + Asynq                           |
| AI           | OpenAI, Gemini, Claude (다중 Provider)  |
| 인프라       | Kubernetes, Helm, Terraform             |
| CI/CD        | GitHub Actions                          |
