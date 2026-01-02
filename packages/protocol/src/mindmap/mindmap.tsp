import "../common/errors.tsp";

using TypeSpec.Http;
using TypeSpec.Rest;

namespace MindHit.Mindmap;

// ============ Enums ============

@doc("마인드맵 노드 타입")
enum NodeType {
  core: "core",
  topic: "topic",
  subtopic: "subtopic",
  page: "page",
}

@doc("마인드맵 레이아웃 타입")
enum LayoutType {
  galaxy: "galaxy",
  tree: "tree",
  radial: "radial",
}

@doc("마인드맵 상태")
enum MindmapStatus {
  pending: "pending",
  generating: "generating",
  completed: "completed",
  failed: "failed",
}

// ============ Models ============

@doc("3D 좌표")
model Position {
  x: float64;
  y: float64;
  z: float64;
}

@doc("마인드맵 노드")
model MindmapNode {
  id: string;
  label: string;
  @encodedName("application/json", "type")
  nodeType: string;
  size: float64;
  color: string;
  position?: Position;
  data?: Record<unknown>;
}

@doc("마인드맵 엣지")
model MindmapEdge {
  source: string;
  target: string;
  weight: float64;
  label?: string;
}

@doc("마인드맵 레이아웃 설정")
model MindmapLayout {
  @encodedName("application/json", "type")
  layoutType: string;
  params?: Record<unknown>;
}

@doc("마인드맵 데이터")
model MindmapData {
  nodes: MindmapNode[];
  edges: MindmapEdge[];
  layout: MindmapLayout;
}

@doc("마인드맵 정보")
model Mindmap {
  id: string;
  @encodedName("application/json", "session_id")
  sessionId: string;
  status: MindmapStatus;
  data?: MindmapData;
  @encodedName("application/json", "error_message")
  errorMessage?: string;
  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
  @encodedName("application/json", "updated_at")
  updatedAt: utcDateTime;
}

@doc("마인드맵 응답")
model MindmapResponse {
  mindmap: Mindmap;
}

@doc("마인드맵 생성 요청")
model GenerateMindmapRequest {
  @doc("강제 재생성 여부")
  force?: boolean = false;
}

// ============ Routes ============

@route("/v1/sessions/{id}/mindmap")
namespace MindmapRoutes {
  @get
  @doc("세션의 마인드맵 조회")
  op getMindmap(
    @header authorization: string,
    @path id: string
  ): {
    @statusCode statusCode: 200;
    @body body: MindmapResponse;
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
  @route("/generate")
  @doc("마인드맵 생성 요청")
  op generateMindmap(
    @header authorization: string,
    @path id: string,
    @body body: GenerateMindmapRequest
  ): {
    @statusCode statusCode: 202;
    @body body: MindmapResponse;
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
}
