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
