# Phase 13: 결제 및 구독 시스템

## 목표

사용자가 Free, Pro, Enterprise 플랜을 선택하고 결제할 수 있는 시스템을 구축합니다.
토큰 사용량을 추적하고 사용자에게 시각화하여 제공합니다.

## 선행 조건

- [ ] Phase 12 (배포 및 운영) 완료
- [ ] Stripe 계정 생성 및 API 키 확보

## 관련 문서

- [가격 정책](../../product/05-pricing.md)
- [데이터 구조](../02-data-structure.md)

---

## Step 13.1: Ent 스키마 추가

### 목표

결제 및 구독 관련 데이터베이스 스키마를 추가합니다.

### 체크리스트

- [ ] Plan 스키마 생성 (`ent/schema/plan.go`)
- [ ] Subscription 스키마 생성 (`ent/schema/subscription.go`)
- [ ] TokenUsage 스키마 생성 (`ent/schema/tokenusage.go`)
- [ ] Invoice 스키마 생성 (`ent/schema/invoice.go`)
- [ ] User 스키마에 edge 추가 (subscriptions, token_usage)
- [ ] Session 스키마에 edge 추가 (token_usage)
- [ ] `go generate ./ent` 실행

### 코드 예시

**ent/schema/plan.go:**

```go
package schema

import (
    "time"

    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
)

type Plan struct {
    ent.Schema
}

func (Plan) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").Unique().Immutable(), // 'free', 'pro', 'enterprise'
        field.String("name").NotEmpty(),
        field.Int("price_cents").Default(0),
        field.String("billing_period").Default("monthly"),
        field.Int("token_limit").Optional().Nillable(),
        field.Int("session_retention_days").Optional().Nillable(),
        field.Int("max_concurrent_sessions").Optional().Nillable(),
        field.JSON("features", map[string]bool{}).Default(map[string]bool{}),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}

func (Plan) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("subscriptions", Subscription.Type),
    }
}
```

**ent/schema/subscription.go:**

```go
package schema

import (
    "time"

    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
)

type Subscription struct {
    ent.Schema
}

func (Subscription) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.Enum("status").
            Values("active", "canceled", "past_due", "trialing").
            Default("active"),
        field.Time("current_period_start"),
        field.Time("current_period_end"),
        field.Bool("cancel_at_period_end").Default(false),
        field.String("stripe_subscription_id").Optional().Nillable(),
        field.String("stripe_customer_id").Optional().Nillable(),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (Subscription) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("user", User.Type).Ref("subscriptions").Unique().Required(),
        edge.From("plan", Plan.Type).Ref("subscriptions").Unique().Required(),
    }
}

func (Subscription) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("status").Edges("user"),
    }
}
```

**ent/schema/tokenusage.go:**

```go
package schema

import (
    "time"

    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

type TokenUsage struct {
    ent.Schema
}

func (TokenUsage) Fields() []ent.Field {
    return []ent.Field{
        field.String("operation").NotEmpty(), // 'summarize', 'mindmap', 'keywords'
        field.Int("tokens_used").Positive(),
        field.String("ai_model").Optional(),
        field.Time("period_start"),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}

func (TokenUsage) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("user", User.Type).Ref("token_usage").Unique().Required(),
        edge.From("session", Session.Type).Ref("token_usage").Unique(),
    }
}

func (TokenUsage) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("period_start").Edges("user"),
    }
}
```

**ent/schema/invoice.go:**

```go
package schema

import (
    "time"

    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "github.com/google/uuid"
)

type Invoice struct {
    ent.Schema
}

func (Invoice) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.Int("amount_cents").NonNegative(),
        field.String("currency").Default("USD"),
        field.Enum("status").
            Values("pending", "paid", "failed", "refunded").
            Default("pending"),
        field.String("stripe_invoice_id").Optional().Nillable(),
        field.Time("paid_at").Optional().Nillable(),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}

func (Invoice) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("user", User.Type).Ref("invoices").Unique().Required(),
        edge.From("subscription", Subscription.Type).Ref("invoices").Unique(),
    }
}
```

---

## Step 13.2: Migration 생성 및 적용

### 목표

새 스키마에 대한 데이터베이스 마이그레이션을 생성하고 적용합니다.

### 체크리스트

- [ ] `atlas migrate diff add_billing_tables` 실행
- [ ] 생성된 SQL 파일 검토
- [ ] 개발 DB에 `atlas migrate apply` 실행
- [ ] 기본 플랜 데이터 seed 스크립트 작성

### Seed 데이터

```sql
INSERT INTO plans (id, name, price_cents, billing_period, token_limit, session_retention_days, max_concurrent_sessions, features) VALUES
('free', 'Free', 0, 'monthly', 50000, 30, 1, '{"export_png": true}'),
('pro', 'Pro', 1200, 'monthly', 500000, NULL, 5, '{"export_png": true, "export_svg": true, "export_md": true, "export_json": true, "priority_support": true}'),
('enterprise', 'Enterprise', NULL, 'monthly', NULL, NULL, NULL, '{"export_png": true, "export_svg": true, "export_md": true, "export_json": true, "api_access": true, "team_sharing": true, "sso": true, "custom_ai": true}');
```

---

## Step 13.3: Backend 서비스 구현

### 목표

구독, 사용량 추적, Stripe 연동 서비스를 구현합니다.

### 체크리스트

- [ ] `internal/service/usage_service.go` 생성
  - [ ] RecordUsage: 토큰 사용량 기록
  - [ ] GetCurrentUsage: 현재 기간 사용량 조회
  - [ ] GetUsageHistory: 월별 히스토리 조회
  - [ ] CheckLimit: 제한 체크
- [ ] `internal/service/subscription_service.go` 생성
  - [ ] GetSubscription: 현재 구독 조회
  - [ ] CreateCheckoutSession: Stripe Checkout 생성
  - [ ] HandleWebhook: Stripe webhook 처리
  - [ ] CancelSubscription: 구독 취소
- [ ] `internal/service/stripe_service.go` 생성
  - [ ] Stripe API 연동

### 빌링 주기 (Billing Period) 핵심 로직

**모든 플랜은 가입일/구독일 기준 30일 주기**로 토큰이 리셋됩니다:

| 플랜 | 빌링 주기 기준 |
|-----|--------------|
| Free | 가입일 기준 30일 주기 (예: 12/10 가입 → 12/10~1/8, 1/9~2/7...) |
| Pro | 구독 시작일 기준 (예: 15일 구독 → 매월 15일~다음달 14일) |
| Enterprise | 계약일 기준 |

**핵심 필드:**
- `users.created_at`: Free 플랜 빌링 주기 계산에 사용
- `subscriptions.current_period_start`: Pro/Enterprise 현재 빌링 주기 시작일
- `subscriptions.current_period_end`: Pro/Enterprise 현재 빌링 주기 종료일
- `token_usage.period_start`: 토큰이 어느 빌링 주기에 속하는지

### 코드 예시

**internal/service/usage_service.go:**

```go
package service

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/mindhit/api/ent"
    "github.com/mindhit/api/ent/subscription"
    "github.com/mindhit/api/ent/tokenusage"
    "github.com/mindhit/api/ent/user"
)

type UsageService struct {
    client *ent.Client
}

type UsageRecord struct {
    UserID    uuid.UUID
    SessionID uuid.UUID
    Operation string
    Tokens    int
    AIModel   string
}

type LimitStatus struct {
    TokensUsed  int
    TokenLimit  int
    IsUnlimited bool
    PercentUsed float64
    CanUseAI    bool
}

func NewUsageService(client *ent.Client) *UsageService {
    return &UsageService{client: client}
}

func (s *UsageService) RecordUsage(ctx context.Context, record UsageRecord) error {
    periodStart := s.getCurrentPeriodStart(ctx, record.UserID)

    builder := s.client.TokenUsage.Create().
        SetUserID(record.UserID).
        SetOperation(record.Operation).
        SetTokensUsed(record.Tokens).
        SetPeriodStart(periodStart)

    if record.SessionID != uuid.Nil {
        builder.SetSessionID(record.SessionID)
    }
    if record.AIModel != "" {
        builder.SetAiModel(record.AIModel)
    }

    _, err := builder.Save(ctx)
    return err
}

func (s *UsageService) CheckLimit(ctx context.Context, userID uuid.UUID) (*LimitStatus, error) {
    sub, _ := s.client.Subscription.
        Query().
        Where(
            subscription.StatusEQ(subscription.StatusActive),
            subscription.HasUserWith(user.IDEQ(userID)),
        ).
        WithPlan().
        Only(ctx)

    periodStart := s.getCurrentPeriodStart(ctx, userID)

    var usage int
    err := s.client.TokenUsage.
        Query().
        Where(
            tokenusage.HasUserWith(user.IDEQ(userID)),
            tokenusage.PeriodStartGTE(periodStart),
        ).
        Aggregate(ent.Sum(tokenusage.FieldTokensUsed)).
        Scan(ctx, &usage)
    if err != nil {
        usage = 0
    }

    limit := 50000 // Free 기본값
    isUnlimited := false

    if sub != nil && sub.Edges.Plan != nil {
        if sub.Edges.Plan.TokenLimit != nil {
            limit = *sub.Edges.Plan.TokenLimit
        } else {
            isUnlimited = true
        }
    }

    percentUsed := 0.0
    if !isUnlimited && limit > 0 {
        percentUsed = float64(usage) / float64(limit) * 100
    }

    return &LimitStatus{
        TokensUsed:  usage,
        TokenLimit:  limit,
        IsUnlimited: isUnlimited,
        PercentUsed: percentUsed,
        CanUseAI:    isUnlimited || usage < limit,
    }, nil
}

func (s *UsageService) getCurrentPeriodStart(ctx context.Context, userID uuid.UUID) time.Time {
    sub, err := s.client.Subscription.
        Query().
        Where(
            subscription.StatusEQ(subscription.StatusActive),
            subscription.HasUserWith(user.IDEQ(userID)),
        ).
        Only(ctx)

    if err != nil || sub == nil {
        // Free 플랜: 매월 1일
        now := time.Now()
        return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
    }

    return sub.CurrentPeriodStart
}
```

---

## Step 13.4: API 엔드포인트 구현

### 목표

구독 및 사용량 관련 API 엔드포인트를 구현합니다.

### 체크리스트

- [ ] `internal/controller/subscription_controller.go` 생성
  - [ ] GET `/v1/subscription` - 현재 구독 조회
  - [ ] GET `/v1/subscription/plans` - 플랜 목록
  - [ ] POST `/v1/subscription/checkout` - Checkout 세션 생성
  - [ ] POST `/v1/subscription/portal` - Customer Portal URL
  - [ ] POST `/v1/subscription/cancel` - 구독 취소
- [ ] `internal/controller/usage_controller.go` 생성
  - [ ] GET `/v1/usage` - 현재 사용량
  - [ ] GET `/v1/usage/history` - 월별 히스토리
- [ ] `internal/controller/webhook_controller.go` 생성
  - [ ] POST `/webhook/stripe` - Stripe webhook
- [ ] 라우터에 엔드포인트 등록

### API 응답 예시

**GET /v1/usage:**

```json
{
  "usage": {
    "periodStart": "2024-12-01T00:00:00Z",
    "periodEnd": "2024-12-31T23:59:59Z",
    "tokensUsed": 234500,
    "tokenLimit": 500000,
    "percentUsed": 46.9,
    "byOperation": {
      "summarize": 145200,
      "mindmap": 78300,
      "keywords": 11000
    },
    "sessionCount": 23
  }
}
```

---

## Step 13.5: AI 서비스 수정 (토큰 측정)

### 목표

AI 서비스에서 토큰 사용량을 **정확하게** 추적하고 제한을 체크합니다.

### 토큰 측정 원리

AI API는 응답에 **실제 사용된 토큰 수**를 포함하여 반환합니다. 예상치가 아닌 정확한 값입니다.

**AI 제공업체별 응답 필드:**

| 제공업체 | 응답 필드 | 예시 값 |
|---------|----------|--------|
| OpenAI | `response.Usage.TotalTokens` | 3847 |
| Google Gemini | `response.UsageMetadata.TotalTokenCount` | 3847 |
| Anthropic Claude | `response.Usage.InputTokens + OutputTokens` | 3847 |

### 체크리스트

- [ ] AI 서비스에 UsageService 의존성 주입
- [ ] 요청 전 사용량 제한 체크
- [ ] API 호출 후 **응답에서 토큰 사용량 추출**
- [ ] token_usage 테이블에 정확한 값 기록
- [ ] 제한 초과 시 적절한 에러 반환

### 코드 예시: OpenAI 토큰 추출

```go
// internal/service/ai_service.go
package service

import (
    "context"

    "github.com/google/uuid"
    openai "github.com/sashabaranov/go-openai"
)

type AIService struct {
    openaiClient *openai.Client
    usageService *UsageService
}

func (s *AIService) GenerateMindmap(ctx context.Context, userID, sessionID uuid.UUID, content string) (*Mindmap, error) {
    // 1. 제한 체크
    status, err := s.usageService.CheckLimit(ctx, userID)
    if err != nil {
        return nil, err
    }
    if !status.CanUseAI {
        return nil, ErrTokenLimitExceeded
    }

    // 2. OpenAI API 호출
    resp, err := s.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: openai.GPT4Turbo,
        Messages: []openai.ChatCompletionMessage{
            {Role: openai.ChatMessageRoleUser, Content: content},
        },
    })
    if err != nil {
        return nil, err
    }

    // 3. 응답에서 정확한 토큰 사용량 추출
    //    ⚠️ 이 값은 예상치가 아닌 실제 사용량입니다!
    tokensUsed := resp.Usage.TotalTokens  // prompt_tokens + completion_tokens

    // 4. 사용량 기록 (빌링 주기 자동 계산)
    if err := s.usageService.RecordUsage(ctx, UsageRecord{
        UserID:    userID,
        SessionID: sessionID,
        Operation: "mindmap",
        Tokens:    tokensUsed,
        AIModel:   "gpt-4-turbo",
    }); err != nil {
        slog.Error("failed to record usage", "error", err)
        // 기록 실패해도 결과는 반환 (graceful degradation)
    }

    return parseMindmap(resp.Choices[0].Message.Content), nil
}
```

### 코드 예시: Google Gemini 토큰 추출

```go
// Google Gemini API 사용 시
func (s *AIService) callGemini(ctx context.Context, prompt string) (string, int, error) {
    resp, err := s.geminiModel.GenerateContent(ctx, genai.Text(prompt))
    if err != nil {
        return "", 0, err
    }

    // Gemini는 UsageMetadata에서 토큰 수 제공
    tokensUsed := int(resp.UsageMetadata.TotalTokenCount)

    return resp.Candidates[0].Content.Parts[0].(genai.Text), tokensUsed, nil
}
```

### 코드 예시: Anthropic Claude 토큰 추출

```go
// Anthropic Claude API 사용 시
func (s *AIService) callClaude(ctx context.Context, prompt string) (string, int, error) {
    resp, err := s.claudeClient.CreateMessage(ctx, anthropic.MessageRequest{
        Model:     "claude-3-opus-20240229",
        MaxTokens: 4096,
        Messages: []anthropic.Message{
            {Role: "user", Content: prompt},
        },
    })
    if err != nil {
        return "", 0, err
    }

    // Claude는 input_tokens + output_tokens 합산
    tokensUsed := resp.Usage.InputTokens + resp.Usage.OutputTokens

    return resp.Content[0].Text, tokensUsed, nil
}
```

### 토큰 기록 시 빌링 주기 자동 계산

```go
// internal/service/usage_service.go

func (s *UsageService) RecordUsage(ctx context.Context, record UsageRecord) error {
    // 현재 사용자의 빌링 주기 시작일 자동 계산
    periodStart := s.getCurrentPeriodStart(ctx, record.UserID)

    return s.client.TokenUsage.Create().
        SetUserID(record.UserID).
        SetSessionID(record.SessionID).
        SetOperation(record.Operation).
        SetTokensUsed(record.Tokens).
        SetAiModel(record.AIModel).
        SetPeriodStart(periodStart).  // 빌링 주기 자동 설정
        Exec(ctx)
}

// getCurrentPeriodStart: 사용자별 빌링 주기 시작일 계산
func (s *UsageService) getCurrentPeriodStart(ctx context.Context, userID uuid.UUID) time.Time {
    // 활성 구독 조회
    sub, err := s.client.Subscription.Query().
        Where(
            subscription.StatusEQ(subscription.StatusActive),
            subscription.HasUserWith(user.IDEQ(userID)),
        ).
        Only(ctx)

    if err != nil || sub == nil {
        // Free 플랜: 가입일 기준 30일 주기
        return s.calculateFreePlanPeriodStart(ctx, userID)
    }

    // Pro/Enterprise 플랜: 구독의 current_period_start 사용
    return sub.CurrentPeriodStart
}

// calculateFreePlanPeriodStart: Free 플랜 빌링 주기 계산 (가입일 기준 30일)
func (s *UsageService) calculateFreePlanPeriodStart(ctx context.Context, userID uuid.UUID) time.Time {
    // 사용자 가입일 조회
    u, err := s.client.User.Get(ctx, userID)
    if err != nil {
        // 조회 실패 시 현재 시간 기준
        return time.Now().UTC()
    }

    signupDate := u.CreatedAt.UTC()
    now := time.Now().UTC()

    // 가입일로부터 몇 번째 30일 주기인지 계산
    daysSinceSignup := int(now.Sub(signupDate).Hours() / 24)
    periodNumber := daysSinceSignup / 30

    // 현재 빌링 주기 시작일 = 가입일 + (periodNumber * 30일)
    periodStart := signupDate.AddDate(0, 0, periodNumber*30)

    return periodStart
}
```

### 토큰 측정 정확도 보장

| 항목 | 설명 |
|-----|------|
| 데이터 소스 | AI API 응답의 `usage` 필드 (예상치 아님) |
| 저장 시점 | API 호출 직후 즉시 기록 |
| 빌링 주기 | 사용자별 구독 시작일 기준 자동 계산 |
| 실패 처리 | 기록 실패 시 로그만 남기고 서비스는 계속 |

---

## Step 13.6: Web App UI 구현

### 목표

구독 관리 페이지와 사용량 시각화를 구현합니다.

### 체크리스트

- [ ] `apps/web/app/(dashboard)/settings/subscription/page.tsx` 생성
  - [ ] 현재 플랜 표시
  - [ ] 플랜 비교 카드
  - [ ] 업그레이드/다운그레이드 버튼
- [ ] `apps/web/components/usage-progress.tsx` 생성
  - [ ] 토큰 사용량 프로그레스 바
  - [ ] 사용량 상세 breakdown
- [ ] `apps/web/components/usage-chart.tsx` 생성
  - [ ] 월별 사용량 차트 (recharts)
- [ ] Stripe Checkout 리다이렉트 처리
- [ ] 결제 성공/실패 페이지

### UI 컴포넌트

```tsx
// apps/web/components/usage-progress.tsx
interface UsageProgressProps {
  used: number;
  limit: number | null;
  showBreakdown?: boolean;
}

export function UsageProgress({ used, limit, showBreakdown }: UsageProgressProps) {
  const isUnlimited = limit === null;
  const percent = isUnlimited ? 0 : (used / limit) * 100;

  return (
    <div className="space-y-2">
      <div className="flex justify-between text-sm">
        <span>{used.toLocaleString()} tokens used</span>
        <span>{isUnlimited ? 'Unlimited' : `${limit.toLocaleString()} limit`}</span>
      </div>
      <Progress value={percent} className="h-2" />
      {!isUnlimited && (
        <p className="text-xs text-muted-foreground">
          {percent.toFixed(1)}% used
        </p>
      )}
    </div>
  );
}
```

---

## Step 13.7: 문서 및 테스트

### 목표

결제 시스템 문서화 및 테스트를 완료합니다.

### 체크리스트

- [ ] API 스펙 문서 업데이트 (OpenAPI)
- [ ] 사용량 관련 E2E 테스트 작성
- [ ] Stripe webhook 테스트
- [ ] 구독 플로우 통합 테스트
- [ ] 가격 정책 문서 최종 검토

---

## 결과물

### 생성되는 파일

```
apps/api/
├── ent/schema/
│   ├── plan.go
│   ├── subscription.go
│   ├── tokenusage.go
│   └── invoice.go
├── internal/
│   ├── controller/
│   │   ├── subscription_controller.go
│   │   ├── usage_controller.go
│   │   └── webhook_controller.go
│   └── service/
│       ├── subscription_service.go
│       ├── usage_service.go
│       └── stripe_service.go

apps/web/
├── app/(dashboard)/settings/subscription/
│   └── page.tsx
└── components/
    ├── usage-progress.tsx
    └── usage-chart.tsx
```

### 환경 변수

```env
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_PRO_MONTHLY=price_...
STRIPE_PRICE_PRO_YEARLY=price_...
```

---

## 검증

### API 테스트

```bash
# 현재 구독 조회
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/v1/subscription

# 사용량 조회
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/v1/usage

# 플랜 목록
curl http://localhost:8080/v1/subscription/plans
```

### Stripe 테스트

```bash
# Stripe CLI로 webhook 테스트
stripe listen --forward-to localhost:8080/webhook/stripe

# 테스트 결제
stripe trigger checkout.session.completed
```

---

## 다음 단계

Phase 13 완료 후:
- 운영 환경에서 Stripe 라이브 모드 활성화
- 결제 분석 대시보드 구축 (선택)
- A/B 테스트로 가격 최적화 (선택)
