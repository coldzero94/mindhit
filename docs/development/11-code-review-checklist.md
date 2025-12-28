# Code Review Checklist (Phase 8 완료 후)

> **목적**: Phase 8 완료 후 전체 코드베이스의 잠재적 문제점 점검
> **방식**: 영역별 순차 리뷰 (Backend → Extension → Web)
> **날짜**: 2025-12-28

---

## 1. Backend (`apps/backend/`)

### 1.1 보안

- [x] JWT 토큰 검증 로직 확인 - ✅ HMAC 알고리즘 검증, 서명 확인 정상
- [x] 비밀번호 해싱 알고리즘 (bcrypt cost factor) - ✅ bcrypt.DefaultCost 사용
- [x] SQL Injection 방지 (Ent ORM 사용 확인) - ✅ Raw SQL 없음, Ent ORM만 사용
- [x] 민감 정보 로깅 여부 - ⚠️ 이메일 로깅됨 (Medium)
- [x] Rate limiting 구현 여부 - ❌ 미구현 (High)
- [x] CORS 설정 확인 - ⚠️ `chrome-extension://*` 너무 광범위 (High)

### 1.2 에러 처리

- [x] 모든 에러에 적절한 HTTP 상태 코드 반환 - ✅ 정상
- [x] 내부 에러 메시지 외부 노출 방지 - ⚠️ 일부 err.Error() 직접 반환
- [x] 패닉 복구 미들웨어 확인 - ✅ gin.Recovery() 적용됨
- [x] DB 연결 실패 처리 - ✅ 에러 래핑 및 graceful shutdown

### 1.3 성능

- [x] N+1 쿼리 패턴 확인 - ❌ cleanup.go에 N+1 존재 (Critical)
- [x] DB 인덱스 적절성 - ⚠️ FK 인덱스 누락 (session_id, user_id 등)
- [x] 불필요한 데이터 로드 여부 - ⚠️ GetEventStats 4개 쿼리 분리
- [x] 커넥션 풀 설정 - ❌ 프로덕션 설정 없음 (High)

### 1.4 코드 품질

- [x] 미사용 코드/함수 제거 - ✅ 깨끗함
- [x] 중복 로직 통합 - ⚠️ extractUserID 3곳 중복, 에러핸들러 중복
- [x] 에러 반환값 무시 여부 (`_ = err`) - ⚠️ JSON unmarshal, tx.Rollback
- [x] Context 전파 확인 - ⚠️ auth/session controller에서 slog.ErrorContext 미사용

### 1.5 설정/환경

- [x] 하드코딩된 값 환경변수화 - ⚠️ CORS origins 하드코딩
- [x] 개발/프로덕션 설정 분리 - ✅ ENVIRONMENT 기반 분기 잘됨
- [x] 시크릿 노출 여부 - ❌ JWT_SECRET 기본값 약함 (Critical)

---

## 2. Extension (`apps/extension/`)

### 2.1 보안

- [x] 토큰 저장 방식 (chrome.storage.local) - ❌ 평문 저장, 암호화 없음 (Critical)
- [x] API 요청 시 토큰 전송 방식 - ✅ Bearer 헤더 사용
- [x] Content Script XSS 방지 - ⚠️ Selector 생성 시 escape 없음 (Medium)
- [x] 메시지 통신 origin 검증 - ❌ sender 검증 없음 (Critical)

### 2.2 에러 처리

- [x] API 실패 시 사용자 피드백 - ⚠️ 세션 컨트롤에서 피드백 없음 (High)
- [x] 오프라인 상태 처리 - ✅ 오프라인 감지 및 재시도 로직 있음
- [x] chrome.runtime.lastError 처리 - ❌ 전혀 구현 안됨 (Medium)
- [x] 네트워크 타임아웃 처리 - ❌ AbortController 없음 (High)

### 2.3 성능

- [x] 이벤트 배치 전송 최적화 - ✅ BATCH_SIZE=10, FLUSH_INTERVAL=30s 적용
- [x] 메모리 누수 (이벤트 리스너 정리) - ⚠️ storage.ts의 online/offline 리스너 정리 없음
- [x] Background script 수명주기 - ❌ 전역 setInterval이 SW 유휴 차단 (High)
- [x] Storage 용량 관리 - ⚠️ quota 체크 없음, TTL 없음

### 2.4 코드 품질

- [x] 타입 안전성 확인 - ✅ `as any` 사용 없음
- [x] 미사용 코드 제거 - ⚠️ storage.ts 여러 함수 미사용, ClickEvent 미사용
- [x] 중복 Chrome storage adapter 통합 - ❌ auth-store, session-store에 동일 코드 중복
- [x] 상수 하드코딩 여부 - ❌ API URL 포트 불일치 (Critical), 스토리지 키 산재

### 2.5 설정/환경

- [x] API URL 환경변수 사용 - ⚠️ background/index.ts만 8080 사용 (불일치)
- [x] Manifest 권한 최소화 - ✅ 필요한 권한만 사용
- [x] 버전 관리 - ✅ manifest.json, package.json 일치 (1.0.0)

---

## 3. Web (`apps/web/`)

### 3.1 보안

- [x] 토큰 저장 방식 (localStorage vs cookie) - ⚠️ localStorage 사용, XSS 취약 (Medium-High)
- [x] XSS 방지 (dangerouslySetInnerHTML 사용 여부) - ✅ 사용 없음
- [x] API 요청 인증 헤더 - ✅ Bearer 토큰 인터셉터 구현
- [x] 민감 데이터 클라이언트 노출 - ✅ NEXT_PUBLIC_ 만 노출, 하드코딩 없음

### 3.2 에러 처리

- [x] API 에러 사용자 피드백 (Toast) - ✅ Sonner 구현, 로그인/회원가입/삭제 시 사용
- [x] 404/500 페이지 - ❌ error.tsx, not-found.tsx 없음 (High)
- [x] 폼 검증 에러 표시 - ✅ Zod v4 + 필드별 에러 표시
- [x] 로딩 상태 처리 - ✅ Skeleton 및 버튼 로딩 상태 구현

### 3.3 성능

- [x] 불필요한 리렌더링 - ✅ 현재 규모에서 문제 없음
- [x] React Query 캐싱 전략 - ⚠️ gcTime, refetchOnWindowFocus 미설정
- [x] 이미지 최적화 - ✅ 현재 이미지 없음 (Phase 11에서 고려)
- [x] 번들 사이즈 - ✅ 최적화된 의존성, tree-shaking 적용

### 3.4 코드 품질

- [x] 미사용 import 제거 - ✅ 미사용 import 없음
- [x] 컴포넌트 분리 적절성 - ⚠️ login/signup form 중복 많음
- [x] 타입 정의 일관성 - ⚠️ statusConfig vs statusLabels 불일치
- [x] 스타일 중복 - ⚠️ 에러 메시지 스타일 5곳 중복

### 3.5 설정/환경

- [x] 환경변수 사용 (`NEXT_PUBLIC_*`) - ✅ 루트 .env 로드, fallback 있음
- [x] 빌드 설정 확인 - ✅ 표준 Next.js 16.1 설정
- [x] SEO 메타 태그 - ⚠️ 루트만 있음, 페이지별 메타 없음

---

## 4. 공통/인프라

### 4.1 API 스펙

- [x] TypeSpec ↔ 실제 구현 일치 - ✅ 파이프라인 정상 작동 (TypeSpec → OpenAPI → Go/TS)
- [x] API 버전 관리 (`/v1/`) - ✅ 모든 15개 엔드포인트 `/v1/` 접두사 사용
- [x] 에러 응답 형식 통일 - ✅ Common.ErrorResponse, Common.ValidationError 일관 사용

### 4.2 Docker/Compose

- [x] 서비스 의존성 설정 - ✅ healthcheck 설정됨 (PostgreSQL, Redis)
- [x] 볼륨 마운트 경로 - ✅ 영속 볼륨 정의됨
- [x] 환경변수 주입 - ✅ .env.example과 docker-compose 포트 일치

### 4.3 문서

- [x] CLAUDE.md 최신화 - ✅ 모든 앱별 CLAUDE.md 존재 및 최신
- [x] README 정확성 - ❌ 루트 README.md 비어있음 (High), extension README 없음
- [x] API 문서 동기화 - ✅ OpenAPI 스펙 커밋됨, 생성 코드와 동기화

---

## 진행 상황

| 영역 | 상태 | 발견된 이슈 | 해결 |
|------|------|-------------|------|
| Backend | ✅ 완료 | Critical 2, High 3, Medium 6 | - |
| Extension | ✅ 완료 | Critical 4, High 4, Medium 5 | - |
| Web | ✅ 완료 | Critical 0, High 1, Medium 5 | - |
| 공통/인프라 | ✅ 완료 | Critical 0, High 1, Medium 0 | - |

---

## 발견된 이슈 목록

### Critical (즉시 수정)

1. **[Backend] N+1 쿼리 - cleanup.go**
   - 위치: `internal/worker/handler/cleanup.go:28-49`
   - 문제: 개별 UPDATE를 루프로 실행 (1000개 세션 = 1001 쿼리)
   - 해결: `.Update()` 배치 업데이트로 변경

2. **[Backend] JWT_SECRET 기본값 약함**
   - 위치: `internal/infrastructure/config/config.go:31`
   - 문제: `"your-secret-key"` 기본값 - 프로덕션에서 위험
   - 해결: 프로덕션에서 필수 환경변수로 설정, 시작 시 검증

3. **[Extension] 토큰 평문 저장**
   - 위치: `stores/auth-store.ts:38-40`
   - 문제: chrome.storage.local에 토큰 암호화 없이 저장
   - 해결: SubtleCrypto API로 암호화 또는 chrome.storage.session 사용

4. **[Extension] 메시지 sender 검증 없음**
   - 위치: `background/index.ts:22`, `content/index.ts:19`
   - 문제: `_sender` 파라미터 완전 무시, 모든 메시지 수락
   - 해결: `sender.id === chrome.runtime.id` 검증 추가

5. **[Extension] API URL 포트 불일치**
   - 위치: `background/index.ts:4` (8080) vs `api.ts:4`, `events.ts:4` (9000)
   - 문제: background에서만 8080 사용
   - 해결: 모든 파일에서 9000으로 통일

6. **[Extension] chromeStorage 어댑터 중복**
   - 위치: `stores/auth-store.ts:14-26`, `stores/session-store.ts:33-45`
   - 문제: 동일한 코드가 두 곳에 중복
   - 해결: `lib/zustand-storage.ts`로 추출

### High (Phase 9 전 수정)

1. **[Backend] Rate Limiting 미구현**
   - 위치: `cmd/api/main.go`
   - 문제: 인증 엔드포인트에 rate limiting 없음
   - 해결: `gin-contrib/limiter` 또는 자체 미들웨어 구현

2. **[Backend] CORS chrome-extension 너무 광범위**
   - 위치: `internal/infrastructure/middleware/cors.go:11`
   - 문제: `chrome-extension://*`가 모든 확장 프로그램 허용
   - 해결: 특정 확장 프로그램 ID로 제한

3. **[Backend] DB 커넥션 풀 미설정**
   - 위치: `cmd/api/main.go:44`, `cmd/worker/main.go:43`
   - 문제: 커넥션 풀 설정 없음 (테스트 환경에만 있음)
   - 해결: `SetMaxOpenConns`, `SetMaxIdleConns` 추가

4. **[Extension] 세션 컨트롤 에러 피드백 없음**
   - 위치: `sidepanel/components/SessionControl.tsx:30-76`
   - 문제: 세션 시작/일시정지/재개/종료 실패 시 사용자 피드백 없음
   - 해결: Toast 또는 에러 메시지 UI 추가

5. **[Extension] 네트워크 타임아웃 미구현**
   - 위치: `lib/api.ts`, `background/index.ts`, `lib/events.ts`
   - 문제: fetch에 AbortController 없음, 무한 대기 가능
   - 해결: 10-30초 타임아웃 AbortController 추가

6. **[Extension] 전역 setInterval이 Service Worker 유휴 차단**
   - 위치: `background/index.ts:86-90`
   - 문제: 30초 interval이 SW 종료 방해, 세션 상태와 무관하게 실행
   - 해결: events.ts 패턴 사용 또는 세션 상태 기반 시작/중지

7. **[Extension] 하드코딩된 웹앱 URL**
   - 위치: `sidepanel/components/LoginPrompt.tsx:77`
   - 문제: `http://localhost:3000/signup` 하드코딩
   - 해결: `VITE_WEB_URL` 환경변수 사용

8. **[Web] error.tsx, not-found.tsx 없음**
   - 위치: `apps/web/src/app/`
   - 문제: 글로벌 에러 바운더리 및 404 페이지 없음
   - 해결: `error.tsx`, `not-found.tsx` 생성

9. **[공통] 루트 README.md 비어있음**
   - 위치: `/README.md`
   - 문제: 프로젝트 개요, 설치 가이드 없음
   - 해결: 프로젝트 소개, 설정 방법, 기술 스택 문서화

### Medium (추후 수정)

1. **[Backend] 이메일 로깅** - `auth_controller.go` 여러 곳에서 이메일 로깅
2. **[Backend] err.Error() 직접 반환** - session/event controller에서 내부 에러 노출 가능성
3. **[Backend] FK 인덱스 누락** - `session_id`, `user_id` 등 관계 필드 인덱스 없음
4. **[Backend] GetEventStats 쿼리 비효율** - 4개 분리 COUNT 쿼리 → 집계 쿼리로 통합 가능
5. **[Backend] extractUserID 중복** - auth/session/event controller에 중복 코드
6. **[Backend] Context 전파 불일치** - auth/session controller에서 `slog.ErrorContext` 미사용
7. **[Extension] CSS selector escape 없음** - `content/index.ts:153-157` XSS 가능성
8. **[Extension] chrome.runtime.lastError 미처리** - 메시지 전송 실패 무시
9. **[Extension] 미사용 코드** - storage.ts의 onOnline/onOffline/isOnline/getAll 미사용
10. **[Extension] 스토리지 키 산재** - "mindhit-auth", "mindhit-session" 등 상수화 필요
11. **[Extension] 토큰 파싱 로직 중복** - events.ts, background/index.ts에서 동일 코드 반복
12. **[Web] localStorage 토큰 저장** - XSS 취약, httpOnly 쿠키 권장
13. **[Web] React Query 캐싱 미설정** - gcTime, refetchOnWindowFocus 미설정
14. **[Web] login/signup form 중복** - handleChange, validation 로직 중복
15. **[Web] statusConfig 불일치** - session-card와 session-detail에서 다른 방식
16. **[Web] 에러 메시지 스타일 중복** - `text-sm text-red-500` 5곳 반복

### Low (개선 사항)

1. **[Backend] 에러 핸들러 중복** - session_controller의 handleXError 메서드들 통합 가능
2. **[Backend] CORS origins 환경변수화** - 하드코딩된 localhost:3000을 설정으로 분리
3. **[Backend] Security Headers 미구현** - X-Frame-Options, HSTS 등 추가 권장
4. **[Extension] ClickEvent 미사용** - types/index.ts에 정의만 있고 사용 안함
5. **[Extension] Storage quota 모니터링 없음** - 용량 한계 접근 시 경고 없음
6. **[Extension] 매직 넘버** - content/index.ts의 500(throttle), 10/1000(text length) 등 상수화 필요
7. **[Web] 페이지별 SEO 메타 없음** - 루트 layout만 있고 개별 페이지 metadata 없음
8. **[공통] Extension README.md 없음** - apps/extension/README.md 없음

---

## 리뷰 완료 조건

1. 모든 체크리스트 항목 검토 완료
2. Critical/High 이슈 0개
3. Medium 이슈 문서화 및 계획 수립
