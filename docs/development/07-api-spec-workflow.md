# API 스펙 공통화 워크플로우

## 개요

모노레포의 장점을 활용하여 **API 스펙을 단일 소스(Single Source of Truth)**로 관리합니다.
TypeSpec으로 API를 정의하면 Backend(Go), Frontend(TypeScript), Extension에서 모두 사용할 수 있는 코드가 자동 생성됩니다.

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Single Source of Truth                           │
│                    packages/protocol/*.tsp                          │
└─────────────────────────────────────────────────────────────────────┘
                                │
                    ┌───────────┼───────────┐
                    ▼           ▼           ▼
              ┌─────────┐ ┌─────────┐ ┌─────────┐
              │ OpenAPI │ │ Go Types│ │   TS    │
              │  Spec   │ │ + Server│ │ Client  │
              └────┬────┘ └────┬────┘ └────┬────┘
                   │           │           │
              ┌────▼────┐ ┌────▼────┐ ┌────▼────┐
              │  Docs   │ │apps/api │ │apps/web │
              │ Swagger │ │         │ │extension│
              └─────────┘ └─────────┘ └─────────┘
```

---

## 장점

| 장점 | 설명 |
|-----|------|
| **일관성** | Frontend/Backend 간 API 타입 불일치 방지 |
| **자동화** | 스펙 변경 시 코드 자동 생성 |
| **문서화** | OpenAPI 스펙으로 Swagger UI 자동 생성 |
| **타입 안전성** | Go/TypeScript 모두 타입 체크 |
| **DX 향상** | API 변경 시 컴파일 타임에 오류 감지 |

---

## 디렉토리 구조

```
mindhit/
├── packages/
│   └── protocol/                    # API 스펙 정의
│       ├── src/
│       │   ├── main.tsp             # 엔트리 포인트
│       │   ├── common/
│       │   │   ├── errors.tsp       # 공통 에러 타입
│       │   │   └── pagination.tsp   # 페이지네이션
│       │   ├── auth/
│       │   │   └── auth.tsp         # 인증 API
│       │   ├── sessions/
│       │   │   └── sessions.tsp     # 세션 API
│       │   ├── events/
│       │   │   └── events.tsp       # 이벤트 API
│       │   └── mindmap/
│       │       └── mindmap.tsp      # 마인드맵 API
│       ├── tsp-output/              # 생성된 파일
│       │   ├── openapi/
│       │   │   └── openapi.yaml     # OpenAPI 스펙
│       │   └── go/
│       │       └── types.go         # Go 타입 (선택)
│       ├── tspconfig.yaml
│       └── package.json
│
├── apps/
│   ├── api/                         # Go Backend
│   │   └── internal/
│   │       └── generated/           # oapi-codegen 생성
│   │           ├── types.gen.go
│   │           └── server.gen.go
│   │
│   ├── web/                         # Next.js
│   │   └── src/
│   │       └── api/
│   │           └── generated/       # openapi-generator 생성
│   │               ├── api.ts
│   │               └── models/
│   │
│   └── extension/                   # Chrome Extension
│       └── src/
│           └── api/
│               └── generated/       # 동일한 클라이언트 사용
```

---

## 워크플로우

### 전체 흐름

```
┌─────────────────────────────────────────────────────────────────────┐
│  1. TypeSpec 정의 (packages/protocol/)                               │
│     └── src/**/*.tsp  ──  API 스펙 작성                              │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼  pnpm run build
┌─────────────────────────────────────────────────────────────────────┐
│  2. OpenAPI 스펙 생성                                                │
│     └── tsp-output/openapi/openapi.yaml                             │
└─────────────────────────────────────────────────────────────────────┘
                    │                               │
                    ▼                               ▼
┌───────────────────────────────────┐ ┌───────────────────────────────┐
│  3a. Go 서버 코드 생성             │ │  3b. TypeScript 클라이언트     │
│      (oapi-codegen)               │ │      (openapi-generator)      │
│      └── apps/api/internal/       │ │      └── apps/web/src/api/    │
│          generated/               │ │          generated/           │
└───────────────────────────────────┘ └───────────────────────────────┘
```

### 실행 순서

```bash
# 1. TypeSpec → OpenAPI 생성
cd packages/protocol
pnpm run build

# 2. OpenAPI → Go 서버 코드 생성
cd apps/api
pnpm run generate:api
# 또는: make generate-api

# 3. OpenAPI → TypeScript 클라이언트 생성
cd apps/web
pnpm run generate:api

# 4. Extension도 동일한 클라이언트 사용
cd apps/extension
pnpm run generate:api
```

### 한 번에 실행 (루트에서)

```bash
# 모든 코드 생성
pnpm run generate

# 또는 moon 사용
moon run :generate
```

---

## TypeSpec 설정

### packages/protocol/package.json

```json
{
  "name": "@mindhit/protocol",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "build": "tsp compile .",
    "watch": "tsp compile . --watch",
    "format": "tsp format **/*.tsp"
  },
  "dependencies": {
    "@typespec/compiler": "^0.61.0",
    "@typespec/http": "^0.61.0",
    "@typespec/openapi": "^0.61.0",
    "@typespec/openapi3": "^0.61.0",
    "@typespec/rest": "^0.61.0"
  }
}
```

### packages/protocol/tspconfig.yaml

```yaml
emit:
  - "@typespec/openapi3"

options:
  "@typespec/openapi3":
    output-file: openapi.yaml
    emitter-output-dir: "{project-root}/tsp-output/openapi"
```

---

## TypeSpec 예시

### packages/protocol/src/main.tsp

```typespec
import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi";

import "./common/errors.tsp";
import "./auth/auth.tsp";
import "./sessions/sessions.tsp";
import "./events/events.tsp";
import "./mindmap/mindmap.tsp";

using TypeSpec.Http;
using TypeSpec.Rest;

@service({
  title: "MindHit API",
  version: "1.0.0",
})
@server("http://localhost:8080", "Development server")
namespace MindHit;
```

### packages/protocol/src/common/errors.tsp

```typespec
namespace MindHit.Common;

model ErrorResponse {
  error: {
    message: string;
    code?: string;
  };
}

model ValidationError {
  error: {
    message: string;
    details?: Record<string>[];
  };
}
```

### packages/protocol/src/auth/auth.tsp

```typespec
import "../common/errors.tsp";

using TypeSpec.Http;
using TypeSpec.Rest;

namespace MindHit.Auth;

model User {
  id: string;
  email: string;
  createdAt: utcDateTime;
  updatedAt: utcDateTime;
}

model SignupRequest {
  @minLength(1)
  email: string;

  @minLength(8)
  password: string;
}

model LoginRequest {
  email: string;
  password: string;
}

model AuthResponse {
  user: User;
  token: string;
}

@route("/v1/auth")
namespace AuthRoutes {
  @post
  @route("/signup")
  op signup(@body body: SignupRequest): AuthResponse | Common.ValidationError;

  @post
  @route("/login")
  op login(@body body: LoginRequest): AuthResponse | Common.ErrorResponse;
}
```

### packages/protocol/src/sessions/sessions.tsp

```typespec
import "../common/errors.tsp";

using TypeSpec.Http;
using TypeSpec.Rest;

namespace MindHit.Sessions;

enum SessionStatus {
  recording,
  paused,
  processing,
  completed,
  failed,
}

model Session {
  id: string;
  title?: string;
  status: SessionStatus;
  startedAt: utcDateTime;
  endedAt?: utcDateTime;
  createdAt: utcDateTime;
  updatedAt: utcDateTime;
}

model SessionWithDetails extends Session {
  pageVisits: PageVisit[];
  highlights: Highlight[];
  mindmap?: MindmapGraph;
}

model PageVisit {
  id: string;
  url: string;
  title?: string;
  enteredAt: utcDateTime;
  leftAt?: utcDateTime;
  dwellTimeSeconds?: int32;
  maxScrollDepth: float32;
}

model Highlight {
  id: string;
  text: string;
  selector?: string;
  color: string;
  createdAt: utcDateTime;
}

model MindmapGraph {
  id: string;
  nodes: MindmapNode[];
  edges: MindmapEdge[];
  generatedAt: utcDateTime;
}

model MindmapNode {
  id: string;
  label: string;
  type: string;
  importance: float32;
  position: {
    x: float32;
    y: float32;
    z: float32;
  };
}

model MindmapEdge {
  id: string;
  source: string;
  target: string;
  weight: float32;
}

model CreateSessionResponse {
  session: Session;
}

model SessionListResponse {
  sessions: Session[];
}

model SessionDetailResponse {
  session: SessionWithDetails;
}

@route("/v1/sessions")
@useAuth(BearerAuth)
namespace SessionRoutes {
  @post
  @route("/start")
  op start(): CreateSessionResponse | Common.ErrorResponse;

  @get
  op list(): SessionListResponse | Common.ErrorResponse;

  @get
  @route("/{id}")
  op get(@path id: string): SessionDetailResponse | Common.ErrorResponse;

  @patch
  @route("/{id}/pause")
  op pause(@path id: string): CreateSessionResponse | Common.ErrorResponse;

  @patch
  @route("/{id}/resume")
  op resume(@path id: string): CreateSessionResponse | Common.ErrorResponse;

  @post
  @route("/{id}/stop")
  op stop(@path id: string): CreateSessionResponse | Common.ErrorResponse;
}
```

### packages/protocol/src/events/events.tsp

```typespec
import "../common/errors.tsp";

using TypeSpec.Http;
using TypeSpec.Rest;

namespace MindHit.Events;

model BatchEvent {
  type: string;
  timestamp: int64;
  url?: string;
  payload: Record<unknown>;
}

model BatchEventsRequest {
  events: BatchEvent[];
}

model BatchEventsResponse {
  processed: int32;
}

@route("/v1/sessions/{sessionId}/events")
@useAuth(BearerAuth)
namespace EventRoutes {
  @post
  op batch(
    @path sessionId: string,
    @body body: BatchEventsRequest
  ): BatchEventsResponse | Common.ErrorResponse;
}
```

---

## Go 코드 생성 (oapi-codegen)

### apps/api/oapi-codegen.yaml

```yaml
package: generated
output: internal/generated/api.gen.go
generate:
  models: true
  gin-server: true
  strict-server: true
  embedded-spec: true
```

### apps/api/Makefile

```makefile
.PHONY: generate-api

generate-api:
	oapi-codegen -config oapi-codegen.yaml ../packages/protocol/tsp-output/openapi/openapi.yaml
```

### 생성된 Go 코드 사용

```go
// internal/controller/auth_controller.go
package controller

import (
    "github.com/mindhit/api/internal/generated"
    "github.com/mindhit/api/internal/service"
)

// generated.StrictServerInterface 구현
type AuthController struct {
    authService *service.AuthService
    jwtService  *service.JWTService
}

// 인터페이스 구현 확인
var _ generated.StrictServerInterface = (*AuthController)(nil)

func (c *AuthController) Signup(
    ctx context.Context,
    request generated.SignupRequestObject,
) (generated.SignupResponseObject, error) {
    user, err := c.authService.Signup(ctx, request.Body.Email, request.Body.Password)
    if err != nil {
        return generated.Signup400JSONResponse{
            Error: generated.ErrorResponse{
                Message: err.Error(),
            },
        }, nil
    }

    token, err := c.jwtService.GenerateToken(user.ID)
    if err != nil {
        return generated.Signup500JSONResponse{}, nil
    }

    return generated.Signup200JSONResponse{
        User:  mapUserToResponse(user),
        Token: token,
    }, nil
}
```

---

## TypeScript 클라이언트 생성 (openapi-generator)

### apps/web/package.json

```json
{
  "scripts": {
    "generate:api": "openapi-generator-cli generate -i ../../packages/protocol/tsp-output/openapi/openapi.yaml -g typescript-axios -o src/api/generated --additional-properties=supportsES6=true,withSeparateModelsAndApi=true,apiPackage=api,modelPackage=models"
  },
  "devDependencies": {
    "@openapitools/openapi-generator-cli": "^2.13.0"
  }
}
```

### 생성된 TypeScript 클라이언트 사용

```typescript
// apps/web/src/lib/api.ts
import { Configuration, AuthApi, SessionsApi, EventsApi } from '@/api/generated';

const config = new Configuration({
  basePath: process.env.NEXT_PUBLIC_API_URL,
  accessToken: () => localStorage.getItem('token') || '',
});

export const authApi = new AuthApi(config);
export const sessionsApi = new SessionsApi(config);
export const eventsApi = new EventsApi(config);
```

```typescript
// apps/web/src/app/login/page.tsx
'use client';

import { authApi } from '@/lib/api';
import { SignupRequest, AuthResponse } from '@/api/generated';

export default function LoginPage() {
  const handleLogin = async (email: string, password: string) => {
    try {
      // 타입 안전한 API 호출
      const response = await authApi.login({ email, password });
      const { user, token } = response.data;
      // ...
    } catch (error) {
      // 에러 타입도 자동 생성됨
    }
  };
  // ...
}
```

```typescript
// apps/web/src/hooks/useSessions.ts
import { useQuery, useMutation } from '@tanstack/react-query';
import { sessionsApi } from '@/lib/api';
import type { Session, SessionWithDetails } from '@/api/generated';

export function useSessions() {
  return useQuery({
    queryKey: ['sessions'],
    queryFn: async () => {
      const response = await sessionsApi.list();
      return response.data.sessions;
    },
  });
}

export function useSession(id: string) {
  return useQuery({
    queryKey: ['session', id],
    queryFn: async () => {
      const response = await sessionsApi.get(id);
      return response.data.session;
    },
  });
}

export function useStartSession() {
  return useMutation({
    mutationFn: async () => {
      const response = await sessionsApi.start();
      return response.data.session;
    },
  });
}
```

---

## Extension에서 사용

```typescript
// apps/extension/src/api/client.ts
import { Configuration, AuthApi, SessionsApi, EventsApi } from './generated';

const getToken = async (): Promise<string> => {
  const result = await chrome.storage.local.get('token');
  return result.token || '';
};

const config = new Configuration({
  basePath: 'http://localhost:8080',
  accessToken: getToken,
});

export const authApi = new AuthApi(config);
export const sessionsApi = new SessionsApi(config);
export const eventsApi = new EventsApi(config);
```

```typescript
// apps/extension/src/background/service-worker.ts
import { eventsApi } from '../api/client';
import type { BatchEvent } from '../api/generated';

async function sendBatchEvents(sessionId: string, events: BatchEvent[]) {
  try {
    const response = await eventsApi.batch(sessionId, { events });
    console.log(`Processed ${response.data.processed} events`);
  } catch (error) {
    console.error('Failed to send events:', error);
  }
}
```

---

## moon.yml 설정 (루트)

```yaml
# moon.yml (루트)
workspace:
  projects:
    - 'apps/*'
    - 'packages/*'

tasks:
  generate:
    command: 'echo "Generating all..."'
    deps:
      - 'protocol:build'
      - 'api:generate-api'
      - 'web:generate-api'
      - 'extension:generate-api'
```

---

## CI/CD 통합

### .github/workflows/api-check.yml

```yaml
name: API Spec Check

on:
  pull_request:
    paths:
      - 'packages/protocol/**'
      - 'apps/api/**'
      - 'apps/web/src/api/**'

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: pnpm/action-setup@v2
        with:
          version: 9

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'pnpm'

      - name: Install dependencies
        run: pnpm install

      - name: Build TypeSpec
        run: pnpm --filter @mindhit/protocol build

      - name: Generate Go code
        run: |
          cd apps/api
          go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
          make generate-api

      - name: Generate TypeScript client
        run: pnpm --filter web generate:api

      - name: Check for uncommitted changes
        run: |
          if [[ -n $(git status --porcelain) ]]; then
            echo "Generated files are out of sync!"
            git diff
            exit 1
          fi
```

---

## 요약

| 단계 | 위치 | 명령어 |
|-----|------|-------|
| API 스펙 정의 | `packages/protocol/src/**/*.tsp` | - |
| OpenAPI 생성 | `packages/protocol/` | `pnpm run build` |
| Go 서버 코드 생성 | `apps/api/` | `make generate-api` |
| TS 클라이언트 생성 | `apps/web/` | `pnpm run generate:api` |
| Extension 클라이언트 | `apps/extension/` | `pnpm run generate:api` |
| 전체 생성 (루트) | `/` | `pnpm run generate` |

**API가 바뀌면:**
1. TypeSpec 수정 (`packages/protocol/src/`)
2. `pnpm run generate` (루트에서 한 번에 실행)
3. Frontend/Backend 타입과 클라이언트가 자동 동기화
