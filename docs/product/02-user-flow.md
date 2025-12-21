# 사용자 플로우 (User Flow)

## 전체 여정 개요

```mermaid
flowchart LR
    A[온보딩<br/>최초 1회] --> B[기록 세션<br/>핵심 루프]
    B --> C[결과 확인<br/>웹앱]
    C --> D[활용/공유]
    D -.-> B
```

---

## 1. 온보딩 플로우 (최초 1회)

### 1.1 확장 프로그램 설치

```mermaid
flowchart TD
    A[Chrome 웹스토어] --> B["MindHit" 검색]
    B --> C[설치]
    C --> D[확장 프로그램 아이콘 클릭]
    D --> E["시작하기" 버튼 클릭]
    E --> F[웹앱 회원가입 페이지로 이동]
```

### 1.2 회원가입/로그인

```mermaid
flowchart TD
    A[웹앱 mindhit.app] --> B[회원가입]
    B --> C[이메일 입력]
    C --> D[비밀번호 설정]
    D --> E{소셜 로그인?}
    E -->|Yes| F[Google 로그인]
    E -->|No| G[이메일 인증]
    F --> H[로그인 완료]
    G --> H
    H --> I[확장 프로그램과 자동 연동]
```

### 1.3 확장 프로그램 연동 확인

```mermaid
flowchart TD
    A[확장 프로그램 팝업] --> B["✓ 로그인됨: user@email.com"]
    B --> C["[Start] 버튼 활성화"]
```

---

## 2. 핵심 루프: 기록 세션

### 2.1 세션 시작

```mermaid
flowchart TD
    A[사용자: 리서치 시작] --> B[확장 프로그램 아이콘 클릭]
    B --> C["팝업: Start Session 버튼"]
    C --> D["[Start Session] 클릭"]
    D --> E["팝업: ● Recording...<br/>타이머 시작"]
```

### 2.2 브라우징 중 (백그라운드 기록)

자동으로 수집되는 데이터:
- 방문한 URL
- 페이지 제목
- 각 페이지 체류 시간
- 사용자가 하이라이팅한 텍스트 (선택적)
- 탭 전환 패턴

```mermaid
flowchart LR
    subgraph 사용자_행동
        A["AI 트렌드" 검색]
        B[첫 번째 결과 클릭]
        C[3분 동안 읽음]
        D[텍스트 드래그 하이라이팅]
        E[다른 탭으로 이동]
    end

    subgraph 기록되는_데이터
        A1[검색어, 검색 결과 페이지]
        B1[URL, 제목, 시작 시간]
        C1[체류 시간 3:00]
        D1[하이라이팅된 텍스트]
        E1[탭 전환, 이전 페이지 체류 종료]
    end

    A --> A1
    B --> B1
    C --> C1
    D --> D1
    E --> E1
```

### 2.3 세션 종료

```mermaid
flowchart TD
    A[사용자: 리서치 끝남] --> B[확장 프로그램 아이콘 클릭]
    B --> C["팝업: Recording 01:32:45<br/>12 pages visited"]
    C --> D["[Stop Session] 클릭"]
    D --> E["팝업: ✓ Session saved!<br/>• 1시간 32분<br/>• 12 페이지 방문<br/>• 마인드맵 생성 중..."]
    E --> F["[웹에서 자세히 보기] 클릭"]
```

---

## 3. 결과 확인 (웹앱)

### 3.1 대시보드 진입

```mermaid
flowchart TD
    A[웹앱 접속<br/>mindhit.app/dashboard] --> B[Dashboard]
    B --> C[최근 세션 목록]
    C --> D["세션 1: AI 트렌드 리서치<br/>12 pages • 1h 32m"]
    C --> E["세션 2: 경쟁사 분석<br/>8 pages • 1h 30m"]
    D --> F["[타임라인] [마인드맵]"]
```

### 3.2 타임라인 뷰

```mermaid
flowchart TD
    A[세션 클릭] --> B[타임라인 탭]
    B --> C["14:30 - Google 검색: AI 트렌드 2024"]
    C --> D["14:32 - TechCrunch: 2024 AI Predictions<br/>⏱ 5분 12초<br/>✨ LLM이 가장 큰 영향을..."]
    D --> E["14:37 - MIT Review: AI in Enterprise<br/>⏱ 8분 45초"]
    E --> F["14:46 - ..."]
```

> ✨ = 하이라이팅한 텍스트 | ⏱ = 체류 시간

### 3.3 마인드맵 뷰

```mermaid
flowchart TD
    A[세션 클릭] --> B[마인드맵 탭]
    B --> C[마인드맵 시각화]

    subgraph 마인드맵_구조
        ROOT((AI 트렌드))
        ROOT --> LLM((LLM))
        ROOT --> ENTERPRISE((Enterprise))
        ROOT --> ETHICS((Ethics))
        LLM --> GPT4[GPT-4]
        LLM --> CLAUDE[Claude]
        ENTERPRISE --> AUTO[자동화]
        ENTERPRISE --> COST[비용절감]
        ETHICS --> REG[규제]
        ETHICS --> BIAS[편향]
    end

    C --> 마인드맵_구조
```

> 노드 클릭 → 관련 페이지 목록 표시

---

## 4. 이메일 리포트

### 4.1 세션 종료 후 자동 발송

```mermaid
flowchart TD
    A["세션 종료 (Stop 클릭)"] --> B[서버에서 AI 마인드맵 생성 완료]
    B --> C[이메일 발송]
    C --> D["📧 MindHit 세션 리포트"]
    D --> E["📊 세션 요약<br/>• 시간: 14:30 - 16:02<br/>• 방문 페이지: 12개<br/>• 주요 키워드: AI, LLM, Enterprise"]
    E --> F["🗺 마인드맵 미리보기"]
    F --> G["[웹에서 자세히 보기]"]
```

---

## 5. 예외 플로우

### 5.1 로그인 안 된 상태에서 시작 시도

```mermaid
flowchart TD
    A[확장 프로그램 클릭] --> B["팝업: 로그인이 필요합니다"]
    B --> C["[로그인하기] 버튼"]
```

### 5.2 세션 중 브라우저 종료

```mermaid
flowchart TD
    A[세션 진행 중] --> B[브라우저 강제 종료 / 크래시]
    B --> C[다음 브라우저 실행 시]
    C --> D["팝업: ⚠️ 이전 세션이 비정상 종료되었습니다"]
    D --> E{"선택"}
    E -->|저장| F["[저장하고 종료]"]
    E -->|삭제| G["[삭제]"]
```

### 5.3 네트워크 오프라인

```mermaid
flowchart TD
    A[세션 진행 중 + 네트워크 끊김] --> B[로컬에 임시 저장]
    B --> C[네트워크 복구 시]
    C --> D[자동 동기화]
```
