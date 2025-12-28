import "../common/errors.tsp";

using TypeSpec.Http;
using TypeSpec.Rest;

namespace MindHit.Usage;

// ============ Models ============

@doc("사용량 현황")
model UsageSummary {
  @encodedName("application/json", "period_start")
  periodStart: utcDateTime;
  @encodedName("application/json", "period_end")
  periodEnd: utcDateTime;
  @encodedName("application/json", "tokens_used")
  tokensUsed: int32;
  @encodedName("application/json", "token_limit")
  tokenLimit: int32;
  @encodedName("application/json", "percent_used")
  percentUsed: float64;
  @encodedName("application/json", "is_unlimited")
  isUnlimited: boolean;
  @encodedName("application/json", "can_use_ai")
  canUseAi: boolean;
  @encodedName("application/json", "by_operation")
  byOperation: Record<int32>;
}

@doc("사용량 응답")
model UsageResponse {
  usage: UsageSummary;
}

@doc("사용량 히스토리 응답")
model UsageHistoryResponse {
  history: UsageSummary[];
}

// ============ Routes ============

@route("/v1/usage")
namespace UsageRoutes {
  @get
  @doc("현재 사용량 조회")
  op getUsage(
    @header authorization: string
  ): {
    @statusCode statusCode: 200;
    @body body: UsageResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  };

  @get
  @route("/history")
  @doc("사용량 히스토리 조회")
  op getUsageHistory(
    @header authorization: string,
    @query months?: int32 = 6
  ): {
    @statusCode statusCode: 200;
    @body body: UsageHistoryResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  };
}
