# 개발 마일스톤

> **Note**: 이 문서는 더 이상 사용되지 않습니다.
> 상세 개발 체크리스트는 [phases/](./phases/) 폴더를 참조하세요.

---

## 전체 로드맵

```mermaid
flowchart LR
    subgraph Foundation["Phase 0-1: Foundation"]
        P0[Phase 0<br/>개발 환경]
        P1[Phase 1<br/>프로젝트 초기화]
        P1_5[Phase 1.5<br/>API 스펙]
    end

    subgraph Auth["Phase 2-4: 백엔드 핵심"]
        P2[Phase 2<br/>인증]
        P3[Phase 3<br/>세션 API]
        P4[Phase 4<br/>이벤트 API]
    end

    subgraph Infra["Phase 5-6: 인프라"]
        P5[Phase 5<br/>모니터링]
        P6[Phase 6<br/>Worker]
    end

    subgraph Frontend["Phase 7-8: 프론트엔드"]
        P7[Phase 7<br/>웹앱]
        P8[Phase 8<br/>Extension]
    end

    subgraph AI["Phase 9-11: AI & 대시보드"]
        P9[Phase 9<br/>플랜/사용량]
        P10[Phase 10<br/>AI 마인드맵]
        P11[Phase 11<br/>대시보드]
    end

    subgraph Launch["Phase 12-14: 출시"]
        P12[Phase 12<br/>프로덕션 모니터링]
        P13[Phase 13<br/>배포/운영]
        P14[Phase 14<br/>결제]
    end

    P0 --> P1 --> P1_5 --> P2
    P2 --> P3 --> P4 --> P5
    P5 --> P6
    P2 --> P7
    P3 --> P7
    P4 --> P8
    P8 --> P9
    P6 --> P10
    P9 --> P10
    P7 --> P11
    P10 --> P11
    P11 --> P12 --> P13 --> P14

    style P0 fill:#e1f5fe
    style P1 fill:#e1f5fe
    style P1_5 fill:#e1f5fe
    style P2 fill:#fff3e0
    style P3 fill:#fff3e0
    style P4 fill:#fff3e0
    style P5 fill:#e8f5e9
    style P6 fill:#e8f5e9
    style P7 fill:#fce4ec
    style P8 fill:#fce4ec
    style P9 fill:#fff9c4
    style P10 fill:#f3e5f5
    style P11 fill:#f3e5f5
    style P12 fill:#efebe9
    style P13 fill:#efebe9
    style P14 fill:#fff9c4
```

---

## Phase별 문서

상세 구현 가이드는 각 Phase 문서를 참조하세요:

| Phase | 문서 | 설명 |
|-------|------|------|
| 0 | [개발 환경](./phases/phase-0-dev-environment.md) | 3단계 개발 환경 (go run, kind, EKS) |
| 1 | [프로젝트 초기화](./phases/phase-1-project-init.md) | 모노레포, Go 백엔드 기본 설정 |
| 1.5 | [API 스펙 공통화](./phases/phase-1.5-api-spec.md) | OpenAPI, 공통 응답 포맷 |
| 2 | [인증 시스템](./phases/phase-2-auth.md) | JWT, Google OAuth, 비밀번호 재설정 |
| 3 | [세션 관리 API](./phases/phase-3-sessions.md) | 세션 CRUD API |
| 4 | [이벤트 수집 API](./phases/phase-4-events.md) | 이벤트 배치 수집 |
| 5 | [모니터링 및 인프라](./phases/phase-5-infra.md) | Prometheus, Grafana |
| 6 | [Worker 및 Job Queue](./phases/phase-6-worker.md) | Asynq + Redis 비동기 작업 |
| 7 | [Next.js 웹앱](./phases/phase-7-webapp.md) | 웹 프론트엔드 |
| 8 | [Chrome Extension](./phases/phase-8-extension.md) | 브라우저 확장 프로그램 |
| 9 | [플랜 및 사용량](./phases/phase-9-plan-usage.md) | 구독 플랜, 토큰 사용량 |
| 10 | [AI 마인드맵](./phases/phase-10-ai.md) | 다중 AI Provider, Worker 연동 |
| 11 | [웹앱 대시보드](./phases/phase-11-dashboard.md) | 마인드맵 시각화 |
| 12 | [프로덕션 모니터링](./phases/phase-12-monitoring.md) | 알림, 로깅 |
| 13 | [배포 및 운영](./phases/phase-13-deployment.md) | Terraform, EKS, CI/CD |
| 14 | [Stripe 결제](./phases/phase-14-billing.md) | 결제 연동 |

---

## 진행 상황 관리

전체 진행 상황은 [phases/README.md](./phases/README.md)에서 관리합니다.
