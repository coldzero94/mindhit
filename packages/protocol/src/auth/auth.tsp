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

@doc("비밀번호 재설정 요청")
model ForgotPasswordRequest {
  @minLength(1)
  @doc("이메일 주소")
  email: string;
}

@doc("비밀번호 재설정 완료 요청")
model ResetPasswordRequest {
  @minLength(1)
  @doc("재설정 토큰")
  token: string;

  @minLength(8)
  @encodedName("application/json", "new_password")
  @doc("새 비밀번호 (최소 8자)")
  newPassword: string;
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

  @get
  @route("/me")
  @doc("현재 사용자 정보 조회")
  op me(
    @header authorization: string
  ): {
    @statusCode statusCode: 200;
    @body body: { user: User };
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  };

  @post
  @route("/logout")
  @doc("로그아웃")
  op logout(
    @header authorization: string
  ): {
    @statusCode statusCode: 200;
    @body body: { message: string };
  } | {
    @statusCode statusCode: 401;
    @body body: Common.ErrorResponse;
  };

  @post
  @route("/forgot-password")
  @doc("비밀번호 재설정 이메일 요청")
  op forgotPassword(
    @body body: ForgotPasswordRequest
  ): {
    @statusCode statusCode: 200;
    @body body: { message: string };
  } | {
    @statusCode statusCode: 400;
    @body body: Common.ValidationError;
  };

  @post
  @route("/reset-password")
  @doc("비밀번호 재설정 완료")
  op resetPassword(
    @body body: ResetPasswordRequest
  ): {
    @statusCode statusCode: 200;
    @body body: { message: string };
  } | {
    @statusCode statusCode: 400;
    @body body: Common.ErrorResponse;
  };
}
