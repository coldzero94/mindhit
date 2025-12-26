# Test Coverage Report

이 문서는 프로젝트의 테스트 커버리지를 추적합니다.

> **Last Updated**: 2025-12-26 (Phase 6 완료 후)

---

## Backend Coverage Summary

### Overall Coverage

| 범위 | 커버리지 |
| ---- | -------- |
| Internal 패키지 전체 | 23.3% |

> Note: Ent 생성 코드를 제외한 `internal/` 패키지만 측정

### Package-level Coverage

| 패키지 | 커버리지 | Phase |
| ------ | -------- | ----- |
| `internal/infrastructure/queue` | 81.4% | Phase 6 |
| `internal/worker/handler` | 80.0% | Phase 6 |
| `internal/service` | 62.5% | Phase 2-4 |
| `internal/controller` | 50.1% | Phase 2-4 |
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
| | `RequestPasswordReset` | 0.0% |
| | `ResetPassword` | 0.0% |
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
| | `Start` | 100.0% |
| | `Pause` | 100.0% |
| | `Resume` | 83.3% |
| | `Stop` | 45.0% |
| | `Get` | 100.0% |
| | `GetWithDetails` | 0.0% |
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
| `service/url_service.go` | `NewURLService` | 100.0% |
| | `GetOrCreate` | 93.3% |
| | `GetByHash` | 0.0% |
| | `UpdateSummary` | 0.0% |
| | `GetURLsWithoutSummary` | 0.0% |
| | `normalizeURL` | 90.9% |
| | `hashURL` | 100.0% |
| `controller/event_controller.go` | 전체 | 0.0% |

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

---

## 미테스트 영역 (0% Coverage)

### Infrastructure

| 파일 | 비고 |
| ---- | ---- |
| `config/config.go` | 환경 변수 로드, 테스트 불필요 |
| `logger/logger.go` | 로거 초기화, 테스트 불필요 |
| `middleware/*.go` | 통합 테스트에서 간접 검증 |
| `controller/response/*.go` | 에러 응답 헬퍼, 간접 검증 |

### 향후 테스트 필요

| 파일 | 함수 | 우선순위 |
| ---- | ---- | -------- |
| `auth_service.go` | `RequestPasswordReset` | Medium |
| `auth_service.go` | `ResetPassword` | Medium |
| `session_service.go` | `GetWithDetails` | Low |
| `event_controller.go` | 전체 | High |
| `url_service.go` | `GetByHash`, `UpdateSummary` | Low |

---

## Test Files

| 위치 | 설명 | Phase |
| ---- | ---- | ----- |
| `internal/controller/auth_controller_test.go` | Auth API 테스트 | Phase 2 |
| `internal/controller/session_controller_test.go` | Session API 테스트 | Phase 3 |
| `internal/service/auth_service_test.go` | Auth 서비스 테스트 | Phase 2 |
| `internal/service/session_service_test.go` | Session 서비스 테스트 | Phase 3 |
| `internal/service/event_service_test.go` | Event 서비스 테스트 | Phase 4 |
| `internal/service/url_service_test.go` | URL 서비스 테스트 | Phase 4 |
| `internal/service/jwt_service_test.go` | JWT 서비스 테스트 | Phase 2 |
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
| Core Services | 60%+ | 62.5% | ✅ |
| Controllers | 50%+ | 50.1% | ✅ |
| New Code (Phase 6+) | 80%+ | 80%+ | ✅ |
| Event Controller | 50%+ | 0.0% | ⚠️ |

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

## History

| 날짜 | Phase | 변경사항 |
| ---- | ----- | -------- |
| 2025-12-26 | Phase 6 | Queue 81.4%, Handler 80.0% 달성 |
| 2025-12-26 | Phase 2-4 | 상세 함수별 커버리지 문서화 |
