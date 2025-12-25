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
