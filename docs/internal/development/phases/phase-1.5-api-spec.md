# Phase 1.5: API 스펙 공통화

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | TypeSpec 기반 API 스펙 정의 → Go/TypeScript 코드 자동 생성 |
| **선행 조건** | Phase 1 완료 |
| **예상 소요** | 5 Steps |
| **결과물** | OpenAPI 스펙 + 생성된 Go/TS 코드 |

> 📖 상세 워크플로우: [07-api-spec-workflow.md](../07-api-spec-workflow.md)

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 1.5.1 | TypeSpec 패키지 설정 | ✅ |
| 1.5.2 | 공통 타입 및 Auth API 스펙 작성 | ✅ |
| 1.5.3 | oapi-codegen 설정 (Go) | ✅ |
| 1.5.4 | @hey-api/openapi-ts 설정 (TypeScript) | ✅ |
| 1.5.5 | 루트 generate 스크립트 설정 | ✅ |

---

## 워크플로우 개요

```mermaid
flowchart TD
    subgraph Source["Single Source of Truth"]
        TSP[packages/protocol/*.tsp<br/>TypeSpec 정의]
    end

    TSP -->|tsp compile| OpenAPI

    subgraph Generated["OpenAPI 스펙"]
        OpenAPI[tsp-output/openapi/openapi.yaml<br/>OpenAPI 3.0 스펙]
    end

    OpenAPI -->|oapi-codegen| GoCode
    OpenAPI -->|@hey-api/openapi-ts| TSCode

    subgraph GoCode["Go 서버 코드"]
        GO_TYPES[apps/backend/internal/generated/<br/>api.gen.go]
    end

    subgraph TSCode["TypeScript (Hey API)"]
        TS_TYPES[types.gen.ts<br/>TypeScript 타입]
        TS_SDK[sdk.gen.ts<br/>API SDK]
        TS_ZOD[zod.gen.ts<br/>Zod v4 스키마]
    end

    style Source fill:#e1f5fe
    style Generated fill:#fff3e0
    style GoCode fill:#e8f5e9
    style TSCode fill:#fce4ec
```

---

## Step 1.5.1: TypeSpec 패키지 설정

### 목표

TypeSpec 기반 API 스펙 정의 환경 구성

### 체크리스트

- [x] **디렉토리 생성**

  ```bash
  mkdir -p packages/protocol/src/{common,auth,sessions,events,mindmap}
  ```

- [x] **package.json 작성**
  - [x] `packages/protocol/package.json`

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

- [x] **tspconfig.yaml 작성**
  - [x] `packages/protocol/tspconfig.yaml`

    ```yaml
    emit:
      - "@typespec/openapi3"

    options:
      "@typespec/openapi3":
        output-file: openapi.yaml
        emitter-output-dir: "{project-root}/tsp-output/openapi"
    ```

- [x] **main.tsp 작성**
  - [x] `packages/protocol/main.tsp` (루트에 위치)

    > **Note**: TypeSpec 컴파일러는 기본적으로 루트의 `main.tsp`를 찾습니다.

    ```typespec
    import "@typespec/http";
    import "@typespec/rest";
    import "@typespec/openapi";

    import "./src/common/errors.tsp";
    import "./src/auth/auth.tsp";

    using TypeSpec.Http;
    using TypeSpec.Rest;

    @service({
      title: "MindHit API",
      version: "1.0.0",
    })
    @server("http://localhost:8080", "Development server")
    namespace MindHit;
    ```

- [x] **의존성 설치**

  ```bash
  cd packages/protocol
  pnpm install
  ```

### 검증

```bash
cd packages/protocol
pnpm run build
# tsp-output/openapi/openapi.yaml 생성 확인
```

### 결과물

```
packages/protocol/
├── main.tsp              # 루트에 위치
├── src/
│   ├── common/
│   │   ├── errors.tsp
│   │   └── pagination.tsp
│   ├── auth/
│   │   └── auth.tsp
│   ├── sessions/
│   ├── events/
│   └── mindmap/
├── tsp-output/
│   └── openapi/
│       └── openapi.yaml
├── tspconfig.yaml
└── package.json
```

---

## Step 1.5.2: 공통 타입 및 Auth API 스펙 작성

### 목표

공통 에러 타입 및 인증 API TypeSpec 정의

### 체크리스트

- [x] **공통 에러 타입**
  - [x] `packages/protocol/src/common/errors.tsp`

    ```typespec
    namespace MindHit.Common;

    @doc("기본 에러 응답")
    model ErrorResponse {
      error: {
        message: string;
        code?: string;
      };
    }

    @doc("유효성 검증 에러")
    model ValidationError {
      error: {
        message: string;
        details?: ValidationDetail[];
      };
    }

    model ValidationDetail {
      field: string;
      message: string;
    }
    ```

- [x] **페이지네이션 타입**
  - [x] `packages/protocol/src/common/pagination.tsp`

    ```typespec
    namespace MindHit.Common;

    model PaginationParams {
      @query
      @doc("페이지 번호 (1부터 시작)")
      page?: int32 = 1;

      @query
      @doc("페이지당 항목 수")
      limit?: int32 = 20;
    }

    model PaginationMeta {
      page: int32;
      limit: int32;
      total: int32;
      totalPages: int32;
    }
    ```

- [x] **Auth API 스펙**
  - [x] `packages/protocol/src/auth/auth.tsp`

    ```typespec
    import "../common/errors.tsp";

    using TypeSpec.Http;
    using TypeSpec.Rest;

    namespace MindHit.Auth;

    // ============ Models ============

    @doc("사용자 정보")
    model User {
      id: string;
      email: string;
      @encodedName("application/json", "created_at")
      createdAt: utcDateTime;
      @encodedName("application/json", "updated_at")
      updatedAt: utcDateTime;
    }

    @doc("회원가입 요청")
    model SignupRequest {
      @minLength(1)
      @doc("이메일 주소")
      email: string;

      @minLength(8)
      @doc("비밀번호 (최소 8자)")
      password: string;
    }

    @doc("로그인 요청")
    model LoginRequest {
      email: string;
      password: string;
    }

    @doc("인증 응답")
    model AuthResponse {
      user: User;
      token: string;
    }

    // ============ Routes ============

    @route("/v1/auth")
    namespace Routes {
      @post
      @route("/signup")
      @doc("회원가입")
      op signup(
        @body body: SignupRequest
      ): {
        @statusCode statusCode: 201;
        @body body: AuthResponse;
      } | {
        @statusCode statusCode: 400;
        @body body: Common.ValidationError;
      } | {
        @statusCode statusCode: 409;
        @body body: Common.ErrorResponse;
      };

      @post
      @route("/login")
      @doc("로그인")
      op login(
        @body body: LoginRequest
      ): {
        @statusCode statusCode: 200;
        @body body: AuthResponse;
      } | {
        @statusCode statusCode: 401;
        @body body: Common.ErrorResponse;
      };

      @post
      @route("/refresh")
      @doc("토큰 갱신")
      op refresh(
        @header authorization: string
      ): {
        @statusCode statusCode: 200;
        @body body: { token: string };
      } | {
        @statusCode statusCode: 401;
        @body body: Common.ErrorResponse;
      };
    }
    ```

- [x] **OpenAPI 생성 확인**

  ```bash
  cd packages/protocol
  pnpm run build
  cat tsp-output/openapi/openapi.yaml
  ```

### 검증

```bash
# OpenAPI 스펙에 /v1/auth/signup, /v1/auth/login 포함 확인
grep -A 5 "/v1/auth" packages/protocol/tsp-output/openapi/openapi.yaml
```

### 결과물

- `packages/protocol/src/common/errors.tsp`
- `packages/protocol/src/common/pagination.tsp`
- `packages/protocol/src/auth/auth.tsp`
- `packages/protocol/tsp-output/openapi/openapi.yaml`

---

## Step 1.5.3: oapi-codegen 설정 (Go)

### 목표

OpenAPI 스펙에서 Go 서버 코드 자동 생성

### 체크리스트

- [x] **oapi-codegen 설치**

  > **Note**: 패키지 경로가 변경되었습니다.

  ```bash
  go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
  ```

- [x] **설정 파일 작성**
  - [x] `apps/backend/oapi-codegen.yaml`

    ```yaml
    package: generated
    output: internal/generated/api.gen.go
    generate:
      models: true
      gin-server: true
      strict-server: true
      embedded-spec: true
    ```

- [x] **generated 디렉토리 생성**

  ```bash
  mkdir -p apps/backend/internal/generated
  ```

- [x] **Makefile에 타겟 추가**
  - [x] `apps/backend/Makefile`

    ```makefile
    .PHONY: generate-api build test lint run

    OPENAPI_SPEC := ../../packages/protocol/tsp-output/openapi/openapi.yaml

    generate-api:
     oapi-codegen -config oapi-codegen.yaml $(OPENAPI_SPEC)

    build:
     go build -o ./bin/server ./cmd/server

    test:
     go test -v -race -coverprofile=coverage.out ./...

    lint:
     golangci-lint run

    run:
     go run ./cmd/server
    ```

- [x] **코드 생성 실행**

  ```bash
  cd apps/backend
  make generate-api
  ```

- [x] **생성된 코드 확인**
  - [x] `internal/generated/api.gen.go` 파일 존재
  - [x] `SignupRequest`, `LoginRequest`, `AuthResponse` 타입 확인
  - [x] `StrictServerInterface` 인터페이스 확인

- [x] **필요 의존성 추가**

  ```bash
  go get github.com/getkin/kin-openapi/openapi3
  go get github.com/oapi-codegen/runtime
  go get github.com/oapi-codegen/runtime/strictmiddleware/gin
  ```

### 검증

```bash
cd apps/backend
make generate-api
ls internal/generated/
# api.gen.go

# 타입 확인
grep "type SignupRequest" internal/generated/api.gen.go
```

### 결과물

```
apps/backend/
├── Makefile
├── oapi-codegen.yaml
└── internal/
    └── generated/
        └── api.gen.go
```

---

## Step 1.5.4: @hey-api/openapi-ts 설정 (TypeScript)

### 목표

OpenAPI 스펙에서 TypeScript 클라이언트 + Zod 스키마 자동 생성

> **Note**: [@hey-api/openapi-ts](https://heyapi.dev/openapi-ts/plugins/zod)는 타입, SDK, Zod 스키마를 한 번에 생성합니다.
> Zod v4를 지원하며, validation 규칙이 자동으로 Zod 스키마에 포함됩니다.

### 체크리스트

- [x] **apps/web 초기화** (아직 없다면)

  ```bash
  mkdir -p apps/web
  cd apps/web
  pnpm init
  ```

- [x] **@hey-api/openapi-ts 설치**

  ```bash
  cd apps/web
  pnpm add -D @hey-api/openapi-ts
  pnpm add zod axios
  ```

- [x] **설정 파일 작성**
  - [x] `apps/web/openapi-ts.config.ts`

    ```typescript
    import { defineConfig } from '@hey-api/openapi-ts';

    export default defineConfig({
      input: '../../packages/protocol/tsp-output/openapi/openapi.yaml',
      output: {
        path: 'src/api/generated',
        format: 'prettier',
      },
      plugins: [
        '@hey-api/typescript',
        '@hey-api/sdk',
        {
          name: 'zod',
          // Zod v4 is the default
        },
      ],
    });
    ```

- [x] **package.json 스크립트 추가**
  - [x] `apps/web/package.json`

    ```json
    {
      "name": "@mindhit/web",
      "version": "0.1.0",
      "private": true,
      "scripts": {
        "generate": "openapi-ts"
      },
      "dependencies": {
        "axios": "^1.6.0",
        "zod": "^4.2.1"
      },
      "devDependencies": {
        "@hey-api/openapi-ts": "^0.89.2"
      }
    }
    ```

- [x] **코드 생성 실행**

  ```bash
  cd apps/web
  pnpm run generate
  ```

- [x] **API 클라이언트 래퍼 작성**
  - [x] `apps/web/src/lib/api.ts`

    ```typescript
    import { createClient } from '../api/generated';

    export const apiClient = createClient({
      baseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
    });

    // Re-export SDK functions for convenience
    export * from '../api/generated/sdk.gen';

    // Re-export Zod schemas for validation
    export * from '../api/generated/zod.gen';

    // Re-export types
    export type * from '../api/generated/types.gen';
    ```

- [ ] **Extension용 설정** (선택)
  - [ ] `apps/extension/`에 동일한 설정 추가
  - [ ] 또는 web의 generated 코드를 symlink

### 검증

```bash
cd apps/web
pnpm run generate
ls src/api/generated/
# types.gen.ts, sdk.gen.ts, zod.gen.ts 확인
```

### 결과물

```
apps/web/
├── package.json
├── openapi-ts.config.ts     # Hey API 설정
└── src/
    ├── api/
    │   └── generated/
    │       ├── types.gen.ts    # TypeScript 타입
    │       ├── sdk.gen.ts      # API SDK 함수
    │       ├── zod.gen.ts      # Zod v4 스키마 (validation 포함)
    │       ├── client.gen.ts   # HTTP 클라이언트
    │       └── index.ts        # 통합 export
    └── lib/
        └── api.ts              # 편의 래퍼
```

### Validation 사용 예시

```typescript
import { zAuthSignupRequest } from '../api/generated/zod.gen';

// 폼 validation
const result = zAuthSignupRequest.safeParse({
  email: 'test@example.com',
  password: '1234',  // 8자 미만 → 실패
});

if (!result.success) {
  console.log(result.error.issues);
  // [{ code: 'too_small', minimum: 8, path: ['password'], ... }]
}
```

### API 호출 예시

```typescript
import { routesLogin } from '../api/generated/sdk.gen';
import { apiClient } from '../lib/api';

// SDK 함수로 API 호출
const response = await routesLogin({
  client: apiClient,
  body: { email: 'user@example.com', password: 'password123' },
});
```

---

## Step 1.5.5: 루트 generate 스크립트 설정

### 목표

한 번의 명령어로 전체 코드 생성

### 체크리스트

- [x] **루트 package.json 업데이트**
  - [x] `package.json`

    ```json
    {
      "name": "mindhit",
      "private": true,
      "scripts": {
        "dev": "moonx :dev",
        "build": "moonx :build",
        "test": "moonx :test",
        "lint": "moonx :lint",
        "ci": "moon ci",
        "generate": "pnpm run generate:protocol && pnpm run generate:api:go && pnpm run generate:api:ts",
        "generate:protocol": "pnpm --filter @mindhit/protocol build",
        "generate:api:go": "cd apps/backend && make generate-api",
        "generate:api:ts": "pnpm --filter @mindhit/web generate:api"
      }
    }
    ```

- [ ] **moon.yml에 generate 태스크 추가** (선택)
  - [ ] `.moon/tasks.yml` 또는 각 프로젝트 moon.yml

    ```yaml
    tasks:
      generate:
        command: 'echo "Generating..."'
        deps:
          - 'protocol:build'
        platform: system
    ```

- [x] **CI용 변경 감지 스크립트**
  - [x] `scripts/check-generated.sh`

    ```bash
    #!/bin/bash
    set -e

    echo "Generating all code..."
    pnpm run generate

    echo "Checking for uncommitted changes..."
    if [[ -n $(git status --porcelain) ]]; then
      echo "❌ Generated files are out of sync!"
      git diff
      exit 1
    fi

    echo "✅ All generated files are up to date"
    ```

- [x] **실행 권한 부여**

  ```bash
  chmod +x scripts/check-generated.sh
  ```

- [ ] **.gitignore 업데이트** (선택)

  ```
  # Generated files (commit these)
  # apps/backend/internal/generated/
  # apps/web/src/api/generated/

  # Or ignore if regenerating in CI
  # Uncomment below to ignore:
  # apps/backend/internal/generated/
  # apps/web/src/api/generated/
  ```

### 검증

```bash
# 루트에서 전체 생성
pnpm run generate

# 각 프로젝트에서 생성된 파일 확인
ls apps/backend/internal/generated/
ls apps/web/src/api/generated/
```

### 결과물

- `pnpm run generate` 명령어로 전체 코드 생성
- CI에서 변경 감지 가능

---

## Phase 1.5 완료 확인

### 전체 검증 체크리스트

- [x] **TypeSpec 컴파일**

  ```bash
  cd packages/protocol && pnpm run build
  cat tsp-output/openapi/openapi.yaml | head -50
  ```

- [x] **Go 코드 생성**

  ```bash
  cd apps/backend && make generate-api
  grep "StrictServerInterface" internal/generated/api.gen.go
  ```

- [x] **TypeScript 클라이언트 생성**

  ```bash
  cd apps/web && pnpm run generate:api
  ls src/api/generated/
  ```

- [x] **전체 생성 스크립트**

  ```bash
  pnpm run generate
  ```

### 테스트 요구사항

| 테스트 유형 | 대상 | 검증 방법 |
| ----------- | ---- | --------- |
| 스펙 검증 | TypeSpec 컴파일 | `pnpm run build` 성공 |
| 코드 생성 | Go 서버 코드 | `go build` 성공 |
| 코드 생성 | TS 클라이언트 | TypeScript 컴파일 성공 |
| 스키마 검증 | OpenAPI 유효성 | `spectral lint openapi.yaml` |

```bash
# Phase 1.5 검증
cd packages/protocol && pnpm run build
cd apps/backend && go build ./...
cd apps/web && pnpm run typecheck
```

> **Note**: Phase 1.5는 코드 생성이 핵심이므로 생성된 코드의 컴파일 성공이 완료 기준입니다.

### 산출물 요약

| 항목 | 위치 | 용도 |
| ---- | ---- | ---- |
| TypeSpec 소스 | `packages/protocol/src/` | API 스펙 정의 (Single Source) |
| OpenAPI 스펙 | `packages/protocol/tsp-output/openapi/openapi.yaml` | 중간 산출물 |
| Go 생성 코드 | `apps/backend/internal/generated/api.gen.go` | 서버 타입/인터페이스 |
| TS 타입 | `apps/web/src/api/generated/types.gen.ts` | TypeScript 타입 |
| TS SDK | `apps/web/src/api/generated/sdk.gen.ts` | API 호출 함수 |
| Zod 스키마 | `apps/web/src/api/generated/zod.gen.ts` | 런타임 validation (Zod v4) |

### API 변경 시 워크플로우

```
1. TypeSpec 수정
   └── packages/protocol/src/**/*.tsp

2. 전체 생성
   └── pnpm run generate

3. 타입 확인
   └── Go: 컴파일 에러 확인
   └── TS: TypeScript 에러 확인

4. 코드 수정
   └── 인터페이스 구현 업데이트

5. 커밋
   └── TypeSpec + 생성 코드 함께 커밋
```

---

## 다음 Phase

Phase 1.5 완료 후 [Phase 2: 인증 시스템](./phase-2-auth.md)으로 진행하세요.
