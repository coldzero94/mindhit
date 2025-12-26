import "../common/errors.tsp";

using TypeSpec.Http;
using TypeSpec.Rest;

namespace MindHit.Events;

// ============ Enums ============

@doc("이벤트 타입")
enum EventType {
  page_visit: "page_visit",
  highlight: "highlight",
  scroll: "scroll",
  click: "click",
}

// ============ Models ============

@doc("이벤트 데이터")
model EventData {
  @doc("이벤트 타입")
  type: string;

  @doc("이벤트 발생 시간 (Unix timestamp ms)")
  timestamp: int64;

  @doc("페이지 URL")
  url?: string;

  @doc("페이지 제목")
  title?: string;

  @doc("하이라이트 텍스트")
  text?: string;

  @doc("CSS 선택자")
  selector?: string;

  @doc("하이라이트 색상")
  color?: string;

  @doc("사용자 노트")
  note?: string;

  @doc("추가 메타데이터 (JSON)")
  metadata?: string;
}

@doc("이벤트 배치 요청")
model BatchEventsRequest {
  @minItems(1)
  @maxItems(100)
  @doc("이벤트 배열 (1-100개)")
  events: EventData[];
}

@doc("이벤트 배치 응답")
model BatchEventsResponse {
  @doc("처리된 이벤트 수")
  processed: int32;

  @doc("총 이벤트 수")
  total: int32;
}

@doc("페이지 방문 정보")
model PageVisit {
  id: string;
  url: string;
  title?: string;
  @encodedName("application/json", "visited_at")
  visitedAt: utcDateTime;
  @encodedName("application/json", "duration_ms")
  durationMs?: int32;
}

@doc("하이라이트 정보")
model Highlight {
  id: string;
  text: string;
  selector?: string;
  color: string;
  note?: string;
  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
}

@doc("이벤트 목록 응답")
model EventListResponse {
  @encodedName("application/json", "page_visits")
  pageVisits: PageVisit[];
  highlights: Highlight[];
  total: int32;
}

@doc("이벤트 통계 응답")
model EventStatsResponse {
  @encodedName("application/json", "total_events")
  totalEvents: int32;

  @encodedName("application/json", "page_visits")
  pageVisits: int32;

  highlights: int32;

  @encodedName("application/json", "unique_urls")
  uniqueUrls: int32;
}

// ============ Routes ============

@route("/v1/sessions/{sessionId}/events")
namespace Routes {
  @post
  @doc("이벤트 배치 전송")
  op batchEvents(
    @header authorization: string,
    @path sessionId: string,
    @body body: BatchEventsRequest
  ): {
    @statusCode statusCode: 200;
    @body body: BatchEventsResponse;
  } | {
    @statusCode statusCode: 400;
    @body body: Common.ErrorResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  } | {
    @statusCode statusCode: 403;
    @body body: Common.ErrorResponse;
  } | {
    @statusCode statusCode: 404;
    @body body: Common.ErrorResponse;
  };

  @get
  @doc("이벤트 목록 조회")
  op listEvents(
    @header authorization: string,
    @path sessionId: string,
    @query type?: string,
    @query limit?: int32 = 50,
    @query offset?: int32 = 0
  ): {
    @statusCode statusCode: 200;
    @body body: EventListResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  } | {
    @statusCode statusCode: 403;
    @body body: Common.ErrorResponse;
  } | {
    @statusCode statusCode: 404;
    @body body: Common.ErrorResponse;
  };

  @get
  @route("/stats")
  @doc("이벤트 통계 조회")
  op getEventStats(
    @header authorization: string,
    @path sessionId: string
  ): {
    @statusCode statusCode: 200;
    @body body: EventStatsResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  } | {
    @statusCode statusCode: 403;
    @body body: Common.ErrorResponse;
  } | {
    @statusCode statusCode: 404;
    @body body: Common.ErrorResponse;
  };
}
