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
