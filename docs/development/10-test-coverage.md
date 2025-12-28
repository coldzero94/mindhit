# Test Coverage Report

이 문서는 프로젝트의 테스트 커버리지를 추적합니다.

> **Last Updated**: 2025-12-28 (Phase 9 Plan & Usage 추가, url_service 버그 수정)

---

## Extension Coverage Summary (Phase 8)

### Test Files

| 파일 | 테스트 수 | 설명 |
| ---- | --------- | ---- |
| `auth-store.test.ts` | 3 | Auth Zustand store |
| `session-store.test.ts` | 10 | Session Zustand store |
| `events.test.ts` | 6 | Event queue logic |
| `api.test.ts` | 12 | API 클라이언트 (MSW 통합) |
| **Total** | **31** | - |

### Test Coverage Details

#### Stores

| Store | 테스트 항목 |
| ----- | ----------- |
| `useAuthStore` | 초기 상태, setAuth, logout |
| `useSessionStore` | 초기 상태, startSession, pauseSession, resumeSession, stopSession, incrementPageCount, incrementHighlightCount, updateElapsedTime, reset |

#### API Client (Integration Tests with MSW)

| Endpoint | 테스트 항목 |
| -------- | ----------- |
| `login` | 성공, 잘못된 자격 증명 |
| `startSession` | 성공, 인증 없음 |
| `pauseSession` | 성공, 세션 없음 |
| `resumeSession` | 성공, 세션 없음 |
| `stopSession` | 성공, 세션 없음 |
| `sendEvents` | 성공, 인증 없음 |

#### Event Queue Logic

| 테스트 항목 |
| ----------- |
| 이벤트 배치 처리 (10개 단위) |
| 빈 이벤트 리스트 처리 |
| page_visit 이벤트 생성 |
| page_leave 이벤트 생성 |
| highlight 이벤트 생성 |
| scroll 이벤트 생성 |

### Test Infrastructure

| 파일 | 설명 |
| ---- | ---- |
| `vitest.config.ts` | Vitest 설정 (happy-dom, 경로 alias) |
| `src/test/setup.ts` | 테스트 셋업 (MSW, Chrome API mock) |
| `src/test/mocks/handlers.ts` | MSW API 핸들러 (auth, sessions, events) |
| `src/test/mocks/server.ts` | MSW 서버 설정 |

### Test Commands

```bash
# 테스트 실행
moonx extension:test

# Watch 모드
pnpm test:watch
```

---

## Frontend Coverage Summary (Phase 7)

### Test Files

| 파일 | 테스트 수 | 설명 |
| ---- | --------- | ---- |
| `auth-store.test.ts` | 7 | Auth Zustand store |
| `use-sessions.test.ts` | 11 | Session hooks (React Query) |
| `auth.test.ts` | 6 | Auth API 클라이언트 |
| `sessions.test.ts` | 6 | Sessions API 클라이언트 |
| `login-form.test.tsx` | 5 | 로그인 폼 컴포넌트 |
| `signup-form.test.tsx` | 5 | 회원가입 폼 컴포넌트 |
| `session-card.test.tsx` | 11 | 세션 카드 컴포넌트 |
| `session-list.test.tsx` | 6 | 세션 목록 컴포넌트 |
| **Total** | **57** | - |

### Test Coverage Details

#### Stores

| Store | 테스트 항목 |
| ----- | ----------- |
| `useAuthStore` | 초기 상태, setAuth, setTokens, logout |

#### Hooks

| Hook | 테스트 항목 |
| ---- | ----------- |
| `sessionKeys` | 쿼리 키 생성 (all, lists, list, detail, events, stats) |
| `useSessions` | 세션 목록 fetch, 페이지네이션 |
| `useSession` | 단일 세션 fetch, empty id 처리 |

#### API Clients

| Client | 테스트 항목 |
| ------ | ----------- |
| `authApi` | login 성공/실패, signup 성공/실패, me, logout |
| `sessionsApi` | list, get 성공/실패, delete 성공/실패 |

#### Auth Components

| 컴포넌트 | 테스트 항목 |
| -------- | ----------- |
| `LoginForm` | 렌더링, 빈 필드 검증, 짧은 비밀번호 검증, 성공 제출, 에러 클리어 |
| `SignupForm` | 렌더링, 짧은 비밀번호 검증, 비밀번호 불일치 검증, 성공 제출, 에러 클리어 |

#### Session Components

| 컴포넌트 | 테스트 항목 |
| -------- | ----------- |
| `SessionCard` | 렌더링, 상태별 배지(active/paused/completed), 날짜 포맷, 설명 표시/숨김, 링크 생성, 이벤트 수 표시 |
| `SessionList` | 로딩 상태, 세션 카드 렌더링, 빈 상태, 페이지네이션, 이전 버튼 클릭, 첫 페이지 비활성화 |

### Test Infrastructure

| 파일 | 설명 |
| ---- | ---- |
| `vitest.config.ts` | Vitest 설정 (jsdom, 경로 alias) |
| `src/test/setup.ts` | 테스트 셋업 (MSW, Next.js mocks) |
| `src/test/mocks/handlers.ts` | MSW API 핸들러 (auth, sessions) |
| `src/test/mocks/server.ts` | MSW 서버 설정 |
| `src/test/utils.tsx` | 커스텀 render (QueryClientProvider) + 헬퍼 re-export |
| `src/test/helpers/auth.ts` | 인증 상태 preset 헬퍼 |
| `src/test/helpers/router.ts` | Next.js 라우터 mock 헬퍼 |

### Test Commands

```bash
# 테스트 실행
pnpm test

# Watch 모드
pnpm test:watch

# 커버리지 리포트
pnpm test:coverage
```

---

## Backend Coverage Summary

### Overall Coverage

| 범위 | 커버리지 |
| ---- | -------- |
| Internal 패키지 전체 | 30.6% |

> Note: Ent 생성 코드를 제외한 `internal/` 패키지만 측정

### Package-level Coverage

| 패키지 | 커버리지 | Phase |
| ------ | -------- | ----- |
| `internal/infrastructure/queue` | 81.4% | Phase 6 |
| `internal/worker/handler` | 80.0% | Phase 6 |
| `internal/service` | 76.0% | Phase 2-4 |
| `internal/controller` | 76.6% | Phase 2-4 |
| `internal/infrastructure/config` | 0.0% | - |
| `internal/infrastructure/logger` | 0.0% | - |
| `internal/infrastructure/middleware` | 0.0% | - |
| `internal/controller/response` | 0.0% | - |

---

## Coverage by Phase

### Phase 2: Authentication

| 파일 | 함수 | 커버리지 |
| ---- | ---- | -------- |
| `service/auth_service.go` | `NewAuthService` | 100.0% |
| | `Signup` | 77.8% |
| | `Login` | 87.5% |
| | `GetUserByID` | 83.3% |
| | `GetUserByEmail` | 83.3% |
| | `generateSecureToken` | 75.0% |
| | `RequestPasswordReset` | 73.3% |
| | `ResetPassword` | 66.7% |
| `service/jwt_service.go` | `NewJWTService` | 100.0% |
| | `GenerateTokenPair` | 71.4% |
| | `GenerateAccessToken` | 75.0% |
| | `generateToken` | 100.0% |
| | `ValidateToken` | 77.8% |
| | `ValidateRefreshToken` | 83.3% |
| | `ValidateAccessToken` | 80.0% |
| `controller/auth_controller.go` | `NewAuthController` | 100.0% |
| | `RoutesSignup` | 69.2% |
| | `RoutesLogin` | 69.2% |
| | `RoutesRefresh` | 83.3% |
| | `RoutesMe` | 70.6% |
| | `RoutesLogout` | 91.7% |
| | `RoutesForgotPassword` | 87.5% |
| | `RoutesResetPassword` | 58.3% |

### Phase 3: Sessions

| 파일 | 함수 | 커버리지 |
| ---- | ---- | -------- |
| `service/session_service.go` | `NewSessionService` | 100.0% |
| | `activeSessions` | 100.0% |
| | `Start` | 100.0% |
| | `Pause` | 100.0% |
| | `Resume` | 83.3% |
| | `Stop` | 45.0% |
| | `Get` | 100.0% |
| | `GetWithDetails` | 88.9% |
| | `ListByUser` | 100.0% |
| | `Update` | 100.0% |
| | `Delete` | 100.0% |
| | `getOwnedSession` | 87.5% |
| `controller/session_controller.go` | `NewSessionController` | 100.0% |
| | `extractUserID` | 88.9% |
| | `RoutesStart` | 75.0% |
| | `RoutesList` | 88.2% |
| | `RoutesGet` | 100.0% |
| | `RoutesUpdate` | 90.0% |
| | `RoutesPause` | 90.0% |
| | `RoutesResume` | 90.0% |
| | `RoutesStop` | 90.0% |
| | `RoutesDelete` | 90.0% |
| | `mapSession` | 100.0% |

### Phase 4: Events

| 파일 | 함수 | 커버리지 |
| ---- | ---- | -------- |
| `service/event_service.go` | `NewEventService` | 100.0% |
| | `ProcessBatchEvents` | 84.6% |
| | `processEvent` | 70.0% |
| | `processPageVisit` | 66.7% |
| | `processHighlight` | 81.8% |
| | `ProcessBatchEventsFromJSON` | 0.0% |
| | `GetEventsBySession` | 71.4% |
| | `GetEventStats` | 69.2% |
| | `toJSON` | 75.0% |
| `service/url_service.go` | `NewURLService` | 100.0% |
| | `GetOrCreate` | 93.3% |
| | `GetByHash` | 100.0% |
| | `UpdateSummary` | 100.0% |
| | `GetURLsWithoutSummary` | 100.0% |
| | `normalizeURL` | 90.9% |
| | `hashURL` | 100.0% |
| `controller/event_controller.go` | `NewEventController` | 100.0% |
| | `extractUserID` | 100.0% |
| | `RoutesBatchEvents` | 84.6% |
| | `RoutesListEvents` | 90.5% |
| | `RoutesGetEventStats` | 78.9% |
| | `ptrToString` | 100.0% |
| | `getStringFromPayload` | 100.0% |

### Phase 6: Worker & Queue

| 파일 | 함수 | 커버리지 |
| ---- | ---- | -------- |
| `queue/client.go` | `NewClient` | 100.0% |
| | `Enqueue` | 100.0% |
| | `Close` | 100.0% |
| `queue/server.go` | `NewServer` | 85.7% |
| | `HandleFunc` | 100.0% |
| | `Run` | 100.0% |
| | `Shutdown` | 100.0% |
| `queue/scheduler.go` | `NewScheduler` | 100.0% |
| | `RegisterPeriodicTasks` | 75.0% |
| | `Run` | 0.0% |
| | `Shutdown` | 100.0% |
| `queue/tasks.go` | `NewSessionProcessTask` | 75.0% |
| | `NewSessionCleanupTask` | 75.0% |
| | `NewURLSummarizeTask` | 75.0% |
| | `NewMindmapGenerateTask` | 75.0% |
| `handler/session.go` | `HandleSessionProcess` | 88.9% |
| `handler/cleanup.go` | `HandleSessionCleanup` | 85.7% |

### Phase 9: Plan & Usage (NEW)

> **Note**: Phase 9 서비스는 API 엔드포인트까지 구현되었으나, 단위 테스트는 아직 미작성 상태입니다.
> Phase 문서에 따라 향후 테스트 작성이 필요합니다.

| 파일 | 함수 | 커버리지 | 비고 |
| ---- | ---- | -------- | ---- |
| `service/subscription_service.go` | `NewSubscriptionService` | 0.0% | 테스트 미작성 |
| | `GetSubscription` | 0.0% | 테스트 미작성 |
| | `GetAvailablePlans` | 0.0% | 테스트 미작성 |
| | `CreateFreeSubscription` | 0.0% | 테스트 미작성 |
| | `GetUserPlan` | 0.0% | 테스트 미작성 |
| | `HasFeature` | 0.0% | 테스트 미작성 |
| | `GetSubscriptionInfo` | 0.0% | 테스트 미작성 |
| | `planToInfo` | 0.0% | 테스트 미작성 |
| `service/usage_service.go` | `NewUsageService` | 0.0% | 테스트 미작성 |
| | `RecordUsage` | 0.0% | 테스트 미작성 |
| | `CheckLimit` | 0.0% | 테스트 미작성 |
| | `GetCurrentUsage` | 0.0% | 테스트 미작성 |
| | `GetUsageHistory` | 0.0% | 테스트 미작성 |
| | `getCurrentPeriodStart` | 0.0% | 테스트 미작성 |
| | `calculatePeriodStartForDate` | 0.0% | 테스트 미작성 |
| | `calculateFreePlanPeriodStart` | 0.0% | 테스트 미작성 |
| `controller/subscription_controller.go` | `NewSubscriptionController` | 0.0% | 테스트 미작성 |
| | `SubscriptionRoutesGetSubscription` | 0.0% | 테스트 미작성 |
| | `SubscriptionRoutesListPlans` | 0.0% | 테스트 미작성 |
| `controller/usage_controller.go` | `NewUsageController` | 0.0% | 테스트 미작성 |
| | `UsageRoutesGetUsage` | 0.0% | 테스트 미작성 |
| | `UsageRoutesGetUsageHistory` | 0.0% | 테스트 미작성 |

**향후 테스트 계획:**

- `service/subscription_service_test.go`: 구독 서비스 단위 테스트
- `service/usage_service_test.go`: 사용량 서비스 단위 테스트
- `controller/subscription_controller_test.go`: 구독 API 통합 테스트
- `controller/usage_controller_test.go`: 사용량 API 통합 테스트

---

## 미테스트 영역 (0% Coverage)

### Infrastructure

| 파일 | 비고 |
| ---- | ---- |
| `config/config.go` | 환경 변수 로드, 테스트 불필요 |
| `logger/logger.go` | 로거 초기화, 테스트 불필요 |
| `middleware/*.go` | 통합 테스트에서 간접 검증 |
| `controller/response/*.go` | 에러 응답 헬퍼, 간접 검증 |
| `controller/handler.go` | 라우터 바인딩, 통합 테스트에서 검증 |

### 향후 테스트 필요

| 파일 | 함수 | 우선순위 |
| ---- | ---- | -------- |
| `session_service.go` | `Stop` (queue 통합) | Medium |
| `event_service.go` | `ProcessBatchEventsFromJSON` | Low |
| `jwt_service.go` | `IsTestToken` | Low |

---

## Backend Test Files

| 위치 | 설명 | Phase |
| ---- | ---- | ----- |
| `internal/controller/auth_controller_test.go` | Auth API 테스트 | Phase 2 |
| `internal/controller/session_controller_test.go` | Session API 테스트 | Phase 3 |
| `internal/controller/event_controller_test.go` | Event API 테스트 | Phase 4 |
| `internal/controller/subscription_controller_test.go` | Subscription API 테스트 | Phase 9 (TODO) |
| `internal/controller/usage_controller_test.go` | Usage API 테스트 | Phase 9 (TODO) |
| `internal/service/auth_service_test.go` | Auth 서비스 테스트 | Phase 2 |
| `internal/service/session_service_test.go` | Session 서비스 테스트 | Phase 3 |
| `internal/service/event_service_test.go` | Event 서비스 테스트 | Phase 4 |
| `internal/service/url_service_test.go` | URL 서비스 테스트 | Phase 4 |
| `internal/service/jwt_service_test.go` | JWT 서비스 테스트 | Phase 2 |
| `internal/service/subscription_service_test.go` | Subscription 서비스 테스트 | Phase 9 (TODO) |
| `internal/service/usage_service_test.go` | Usage 서비스 테스트 | Phase 9 (TODO) |
| `internal/infrastructure/queue/*_test.go` | Queue 테스트 | Phase 6 |
| `internal/worker/handler/handler_test.go` | Worker 핸들러 테스트 | Phase 6 |

---

## How to Measure Coverage

### Full Coverage (including generated code)

```bash
cd apps/backend
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```

### Internal Packages Only (recommended)

```bash
cd apps/backend
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```

### HTML Report

```bash
cd apps/backend
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

### Specific Package

```bash
# Queue 패키지만
go test ./internal/infrastructure/queue/... -cover

# Handler 패키지만
go test ./internal/worker/handler/... -cover

# Service 패키지만
go test ./internal/service/... -cover
```

### Function-level Coverage

```bash
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep "service/"
```

---

## Coverage Goals

| 영역 | 목표 | 현재 | 상태 |
| ---- | ---- | ---- | ---- |
| Core Services | 60%+ | 76.0% | ✅ |
| Controllers | 50%+ | 76.6% | ✅ |
| New Code (Phase 6+) | 80%+ | 80%+ | ✅ |
| Event Controller | 50%+ | 84.6% | ✅ |
| URL Service | 80%+ | 100.0% | ✅ |

### Guidelines

1. **새로 추가되는 코드**는 80% 이상 커버리지 목표
2. **Critical path** (인증, 결제)는 90% 이상 권장
3. **Generated code** (Ent)는 커버리지 측정에서 제외
4. **Integration tests**는 외부 의존성(Redis, DB)이 필요하므로 CI에서 별도 실행
5. **Config/Logger**는 환경 의존적이므로 테스트 제외 가능

---

## CI Integration

```yaml
# .github/workflows/test.yml (예시)
- name: Run tests with coverage
  run: |
    cd apps/backend
    go test ./internal/... -coverprofile=coverage.out -covermode=atomic

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./apps/backend/coverage.out
```

---

## Integration & E2E Test Strategy

### 테스트 피라미드

```mermaid
graph TB
    subgraph "Test Pyramid"
        E2E["🔺 E2E Tests<br/>Phase 8+"]
        INT["🔶 Integration Tests<br/>Phase 7+"]
        UNIT["🟢 Unit Tests<br/>Phase 2-6 ✅ 76%+"]
    end

    E2E --> INT --> UNIT

    style UNIT fill:#22c55e,color:#fff
    style INT fill:#f59e0b,color:#fff
    style E2E fill:#ef4444,color:#fff
```

```mermaid
timeline
    title Test Strategy Timeline
    Phase 2-6 : Unit Tests : Service 76% : Controller 76% : Queue 81%
    Phase 7 : Backend Integration : Auth Flow : Session Flow : API 안정화
    Phase 8 : E2E Tests : Playwright : Extension 연동
    Phase 10 : Worker Integration : AI Pipeline : Full Flow
```

### 도입 시점

| 테스트 유형 | 도입 시점 | 트리거 조건 |
| ------------ | ---------- | ------------ |
| **Unit Tests** | Phase 2-6 | ✅ 완료 |
| **Backend Integration** | Phase 7 이후 | Web App 완성, API 안정화 |
| **E2E (Playwright)** | Phase 8 이후 | Extension 완성, 전체 플로우 구현 |
| **Worker Integration** | Phase 10 이후 | AI 연동 완료, 파이프라인 검증 필요 |

### 왜 지금이 아닌가?

1. **API 스펙 변경 가능성**: Phase 7-8에서 프론트엔드 요구사항에 따라 API 변경 가능
2. **유지보수 비용**: Integration test는 변경에 취약 - 안정화 전 작성 시 지속적 수정 필요
3. **현재 Unit Test 충분**: 76%+ 커버리지로 핵심 비즈니스 로직 검증 완료
4. **외부 의존성**: Redis, PostgreSQL 연동 테스트는 CI 환경 구성 필요

### Integration Test 계획 (Phase 7+)

```text
tests/integration/
├── auth_flow_test.go      # 회원가입 → 로그인 → 토큰 갱신 → 로그아웃
├── session_flow_test.go   # 세션 시작 → 이벤트 수집 → 종료 → Worker 처리
└── worker_flow_test.go    # Queue Enqueue → Worker 처리 → DB 업데이트
```

**필요 인프라:**

- Docker Compose (PostgreSQL + Redis)
- Test fixtures (seed data)
- CI workflow 수정

### E2E Test 계획 (Phase 8+)

```text
tests/e2e/
├── auth.spec.ts           # 로그인/회원가입 UI 플로우
├── dashboard.spec.ts      # 대시보드 세션 목록/상세
├── extension.spec.ts      # Extension ↔ Web 연동
└── mindmap.spec.ts        # 마인드맵 생성/조회
```

**도구:**

- Playwright (크로스 브라우저)
- Chrome Extension testing
- Visual regression (optional)

### 현재 미테스트 영역 분석

| 함수 | 커버리지 | 테스트 방법 | 우선순위 |
| ------ | ---------- | ------------ | ---------- |
| `Session.Stop` (queue) | 45% | Integration Test (Redis) | Phase 7 |
| `ProcessBatchEventsFromJSON` | 0% | Unit Test 가능 | Low |
| `scheduler.Run` | 0% | Skip (blocking operation) | N/A |
| `middleware/*.go` | 0% | Integration Test | Phase 7 |

### Integration Test 환경 (예정)

```yaml
# docker-compose.test.yml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: mindhit_test
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test

  redis:
    image: redis:7-alpine
```

```bash
# 실행 명령 (Phase 7 이후)
docker-compose -f docker-compose.test.yml up -d
go test ./tests/integration/... -tags=integration
```

---

## History

| 날짜 | Phase | 변경사항 |
| ---- | ----- | -------- |
| 2025-12-28 | Phase 9 | Plan & Usage 서비스/컨트롤러 추가 (테스트 미작성, 향후 작성 예정) |
| 2025-12-28 | - | url_service.go 버그 수정: GetURLsWithoutSummary에 빈 content 제외 조건 추가 |
| 2025-12-28 | Phase 8 | Extension 테스트 추가: API 통합 테스트 (MSW), stores, events (31개 테스트) |
| 2025-12-27 | Phase 7 | Frontend 테스트 확장: stores, hooks, API 테스트 추가 (57개 테스트) |
| 2025-12-26 | - | 테스트 커버리지 개선: Service 76.0%, Controller 76.6% |
| 2025-12-26 | Phase 6 | Queue 81.4%, Handler 80.0% 달성 |
| 2025-12-26 | Phase 2-4 | 상세 함수별 커버리지 문서화 |
