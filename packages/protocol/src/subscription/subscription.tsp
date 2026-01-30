import "../common/errors.tsp";

using TypeSpec.Http;
using TypeSpec.Rest;

namespace MindHit.Subscription;

// ============ Models ============

@doc("플랜 정보")
model Plan {
  id: string;
  name: string;
  @encodedName("application/json", "price_cents")
  priceCents: int32;
  @encodedName("application/json", "billing_period")
  billingPeriod: string;
  @encodedName("application/json", "token_limit")
  tokenLimit?: int32;
  @encodedName("application/json", "session_retention_days")
  sessionRetentionDays?: int32;
  @encodedName("application/json", "max_concurrent_sessions")
  maxConcurrentSessions?: int32;
  features: Record<boolean>;
}

@doc("구독 정보")
model SubscriptionInfo {
  id: string;
  status: string;
  @encodedName("application/json", "current_period_start")
  currentPeriodStart: utcDateTime;
  @encodedName("application/json", "current_period_end")
  currentPeriodEnd: utcDateTime;
  @encodedName("application/json", "cancel_at_period_end")
  cancelAtPeriodEnd: boolean;
  plan: Plan;
}

@doc("구독 응답")
model SubscriptionResponse {
  subscription: SubscriptionInfo;
}

@doc("플랜 목록 응답")
model PlanListResponse {
  plans: Plan[];
}

@doc("플랜 변경 요청")
model ChangePlanRequest {
  @encodedName("application/json", "plan_id")
  planId: string;
}

@doc("플랜 변경 응답")
model ChangePlanResponse {
  subscription: SubscriptionInfo;
  message: string;
}

@doc("구독 취소 응답")
model CancelSubscriptionResponse {
  subscription: SubscriptionInfo;
  message: string;
}

// ============ Routes ============

@route("/v1/subscription")
namespace SubscriptionRoutes {
  @get
  @doc("현재 구독 정보 조회")
  op getSubscription(
    @header authorization: string
  ): {
    @statusCode statusCode: 200;
    @body body: SubscriptionResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  };

  @get
  @route("/plans")
  @doc("사용 가능한 플랜 목록 조회")
  op listPlans(
    @header authorization: string
  ): {
    @statusCode statusCode: 200;
    @body body: PlanListResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  };

  @post
  @route("/change")
  @doc("플랜 변경")
  op changePlan(
    @header authorization: string,
    @body body: ChangePlanRequest
  ): {
    @statusCode statusCode: 200;
    @body body: ChangePlanResponse;
  } | {
    @statusCode statusCode: 400;
    @body body: Common.ErrorResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  } | {
    @statusCode statusCode: 404;
    @body body: Common.ErrorResponse;
  };

  @post
  @route("/cancel")
  @doc("구독 취소 (기간 종료 시 Free로 전환)")
  op cancelSubscription(
    @header authorization: string
  ): {
    @statusCode statusCode: 200;
    @body body: CancelSubscriptionResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  };

  @post
  @route("/reactivate")
  @doc("구독 취소 철회")
  op reactivateSubscription(
    @header authorization: string
  ): {
    @statusCode statusCode: 200;
    @body body: SubscriptionResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  };
}
