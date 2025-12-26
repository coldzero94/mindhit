import "../common/errors.tsp";

using TypeSpec.Http;
using TypeSpec.Rest;

namespace MindHit.Session;

// ============ Enums ============

@doc("세션 상태")
enum SessionStatus {
  recording: "recording",
  paused: "paused",
  processing: "processing",
  completed: "completed",
  failed: "failed",
}

// ============ Models ============

@doc("세션 정보")
model Session {
  id: string;
  title?: string;
  description?: string;
  @encodedName("application/json", "session_status")
  sessionStatus: SessionStatus;
  @encodedName("application/json", "started_at")
  startedAt: utcDateTime;
  @encodedName("application/json", "ended_at")
  endedAt?: utcDateTime;
  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
  @encodedName("application/json", "updated_at")
  updatedAt: utcDateTime;
}

@doc("세션 업데이트 요청")
model UpdateSessionRequest {
  title?: string;
  description?: string;
}

@doc("세션 응답")
model SessionResponse {
  session: Session;
}

@doc("세션 목록 응답")
model SessionListResponse {
  sessions: Session[];
}

// ============ Routes ============

@route("/v1/sessions")
namespace Routes {
  @post
  @route("/start")
  @doc("새 세션 시작")
  op start(
    @header authorization: string
  ): {
    @statusCode statusCode: 201;
    @body body: SessionResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  };

  @get
  @doc("세션 목록 조회")
  op list(
    @header authorization: string,
    @query limit?: int32 = 20,
    @query offset?: int32 = 0
  ): {
    @statusCode statusCode: 200;
    @body body: SessionListResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  };

  @get
  @route("/{id}")
  @doc("세션 상세 조회")
  op get(
    @header authorization: string,
    @path id: string
  ): {
    @statusCode statusCode: 200;
    @body body: SessionResponse;
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

  @put
  @route("/{id}")
  @doc("세션 메타데이터 업데이트")
  op update(
    @header authorization: string,
    @path id: string,
    @body body: UpdateSessionRequest
  ): {
    @statusCode statusCode: 200;
    @body body: SessionResponse;
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

  @patch
  @route("/{id}/pause")
  @doc("세션 일시정지")
  op pause(
    @header authorization: string,
    @path id: string
  ): {
    @statusCode statusCode: 200;
    @body body: SessionResponse;
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

  @patch
  @route("/{id}/resume")
  @doc("세션 재개")
  op resume(
    @header authorization: string,
    @path id: string
  ): {
    @statusCode statusCode: 200;
    @body body: SessionResponse;
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

  @post
  @route("/{id}/stop")
  @doc("세션 종료")
  op stop(
    @header authorization: string,
    @path id: string
  ): {
    @statusCode statusCode: 200;
    @body body: SessionResponse;
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

  @delete
  @route("/{id}")
  @doc("세션 삭제")
  op delete(
    @header authorization: string,
    @path id: string
  ): {
    @statusCode statusCode: 204;
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
