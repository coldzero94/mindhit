# 개발 마일스톤

## 전체 로드맵

```
Phase 1          Phase 2           Phase 3          Phase 4
Foundation       Core Features     Polish           Launch
────────────     ─────────────     ──────────       ──────────
[기반 구축]       [핵심 기능]        [품질/UX]         [출시]
                                                         │
─────────────────────────────────────────────────────────┼── 🚀 Launch
                                                         │
```

---

## Phase 1: Foundation (기반 구축)

### 목표
프로젝트 기반 설정 및 인증/세션 기본 API 구현

### 1.1 프로젝트 초기화
- [ ] 모노레포 구조 설정
  ```
  mindhit/
  ├── apps/
  │   ├── extension/    # Chrome Extension
  │   ├── web/          # Next.js Web App
  │   └── server/       # Backend API
  ├── packages/
  │   └── shared/       # 공유 타입, 유틸
  └── package.json      # Workspace root
  ```
- [ ] TypeScript 설정
- [ ] ESLint, Prettier 설정
- [ ] Git hooks (husky, lint-staged)

### 1.2 Database 설정
- [ ] PostgreSQL 로컬 설정 (Docker)
- [ ] Prisma 스키마 정의
  - [ ] users, user_settings
  - [ ] sessions
  - [ ] urls
  - [ ] raw_events (partition 고려)
  - [ ] segments
  - [ ] highlights
  - [ ] mindmap_graphs
- [ ] Migration 생성 및 적용

### 1.3 Backend 기반
- [ ] Fastify 프로젝트 설정
- [ ] 에러 핸들링 미들웨어
- [ ] 로깅 설정 (pino)
- [ ] CORS 설정

### 1.4 인증 시스템
- [ ] POST /auth/signup
- [ ] POST /auth/login
- [ ] JWT 토큰 발급/검증
- [ ] Refresh token 로직
- [ ] 인증 미들웨어

### 1.5 세션 기본 API
- [ ] POST /sessions/start
- [ ] POST /sessions/:id/stop (AI 없이 상태만 변경)
- [ ] GET /sessions
- [ ] GET /sessions/:id
- [ ] DELETE /sessions/:id

### 1.6 Web App 기반
- [ ] Next.js 14 프로젝트 생성
- [ ] TailwindCSS 설정
- [ ] 로그인/회원가입 페이지
- [ ] 인증 상태 관리 (Zustand)
- [ ] API 클라이언트 설정

### 1.7 Extension 기반
- [ ] Manifest V3 설정
- [ ] 기본 팝업 UI
- [ ] chrome.storage 연동
- [ ] 웹앱 인증 토큰 동기화

### 완료 기준
- [ ] 회원가입 → 로그인 → 세션 생성 가능
- [ ] Extension에서 로그인 상태 확인 가능

---

## Phase 2: Core Features (핵심 기능)

### 목표
트래킹 파이프라인 및 마인드맵 생성 구현

### 2.1 Extension 트래킹
- [ ] 이벤트 리스너 구현
  - [ ] chrome.tabs.onActivated
  - [ ] chrome.webNavigation.onCommitted
  - [ ] chrome.windows.onFocusChanged
- [ ] IndexedDB 이벤트 큐
- [ ] Batch sender (5초/200개)
- [ ] Start/Stop UI 완성
- [ ] Recording 상태 표시

### 2.2 하이라이팅 수집
- [ ] Content Script 설정
- [ ] Selection 이벤트 감지
- [ ] 하이라이트 이벤트 발송

### 2.3 이벤트 수신 API
- [ ] POST /events/batch
- [ ] URL 정규화 및 urls 테이블 upsert
- [ ] raw_events bulk insert
- [ ] 멱등성 처리 (sessionId + seq)

### 2.4 Segment 빌더
- [ ] State machine 구현
- [ ] Raw → Segment 변환 로직
- [ ] Edge case 처리
  - [ ] 창 blur/focus
  - [ ] 탭 닫힘
  - [ ] 리다이렉트

### 2.5 Job Queue 설정
- [ ] Redis + BullMQ 설정
- [ ] 세션 종료 시 job enqueue
- [ ] Worker 프로세스 분리

### 2.6 AI 마인드맵 생성
- [ ] OpenAI API 연동
- [ ] 프롬프트 설계
  - [ ] 키워드 추출
  - [ ] 토픽 클러스터링
  - [ ] 계층 구조 생성
- [ ] 마인드맵 JSON 생성
- [ ] 노드 레이아웃 계산
- [ ] mindmap_graphs 저장

### 2.7 Web App 타임라인
- [ ] 대시보드 페이지
- [ ] 세션 카드 컴포넌트
- [ ] 세션 상세 페이지
- [ ] 타임라인 뷰 구현
- [ ] 하이라이트 표시

### 2.8 Web App 마인드맵
- [ ] React Flow 설정
- [ ] 마인드맵 렌더링
- [ ] 줌/팬 인터랙션
- [ ] 노드 클릭 → 관련 페이지 표시
- [ ] 중요도 기반 노드 스타일

### 완료 기준
- [ ] Extension에서 세션 기록 → 서버 전송 가능
- [ ] 세션 종료 후 마인드맵 자동 생성
- [ ] 웹앱에서 타임라인/마인드맵 확인 가능

---

## Phase 3: Polish (품질 개선)

### 목표
UX 개선, 이메일, 안정성

### 3.1 이메일 서비스
- [ ] SendGrid/Resend 연동
- [ ] 이메일 템플릿 (HTML)
- [ ] 마인드맵 이미지 생성 (서버사이드 렌더링)
- [ ] 세션 종료 시 이메일 발송
- [ ] 이메일 on/off 설정

### 3.2 Extension UX
- [ ] 툴바 아이콘 배지 (Recording)
- [ ] 브라우저 재시작 시 복구
- [ ] 네트워크 오프라인 대응
- [ ] 제외 도메인 설정

### 3.3 Web App UX
- [ ] 세션 제목 수정
- [ ] 삭제 확인 모달
- [ ] 로딩 스켈레톤
- [ ] 에러 상태 UI
- [ ] 토스트 알림

### 3.4 마인드맵 개선
- [ ] PNG/SVG 다운로드
- [ ] 마인드맵 재생성 버튼
- [ ] 노드 중요도 시각화 개선

### 3.5 설정 페이지
- [ ] 이메일 알림 설정
- [ ] 제외 도메인 관리
- [ ] 비밀번호 변경
- [ ] 계정 삭제

### 3.6 테스트 & 모니터링
- [ ] 단위 테스트 (services)
- [ ] E2E 테스트 (핵심 플로우)
- [ ] Sentry 설정
- [ ] 로깅 개선

### 완료 기준
- [ ] 세션 종료 시 이메일 수신
- [ ] 오프라인/재시작 시 데이터 유실 없음
- [ ] 주요 에러 상황 대응

---

## Phase 4: Launch (출시)

### 목표
베타 출시 및 피드백 수집

### 4.1 배포 환경
- [ ] Vercel 배포 (Web App)
- [ ] Railway 배포 (Backend + Worker)
- [ ] Supabase/Railway PostgreSQL
- [ ] 환경 변수 관리
- [ ] CI/CD (GitHub Actions)

### 4.2 Chrome 웹스토어
- [ ] Extension 빌드 최적화
- [ ] 스토어용 스크린샷
- [ ] 설명 작성
- [ ] 개인정보처리방침
- [ ] 웹스토어 제출

### 4.3 랜딩 페이지
- [ ] 제품 소개 페이지
- [ ] CTA → 웹스토어
- [ ] 사용 방법 안내

### 4.4 운영 준비
- [ ] 모니터링 대시보드
- [ ] 알림 설정
- [ ] 백업 전략
- [ ] 피드백 채널

### 완료 기준
- [ ] Chrome 웹스토어 등록
- [ ] 실제 사용자 피드백 수집

---

## Backlog (향후)

### 기능 확장
- [ ] Google OAuth
- [ ] 세션 검색
- [ ] 마인드맵 수동 편집
- [ ] 세션 공유 (공개 링크)
- [ ] 팀 기능
- [ ] Notion 연동
- [ ] 마크다운 내보내기

### 기술 개선
- [ ] WebSocket 실시간 업데이트
- [ ] PWA 지원
- [ ] 다국어 (i18n)

### 비즈니스
- [ ] 유료 플랜
- [ ] Stripe 결제
- [ ] 사용량 제한
