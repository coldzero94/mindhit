/**
 * Integration test setup - 실제 백엔드 API 호출
 * MSW를 사용하지 않고 실제 서버와 통신
 */
import axios from "axios";

// Integration 테스트용 API 클라이언트 (Zustand store 의존성 없음)
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:9000";

// Rate limiting 방지를 위한 딜레이
export const delay = (ms: number) =>
  new Promise((resolve) => setTimeout(resolve, ms));

export const testApiClient = axios.create({
  baseURL: `${API_BASE_URL}/v1`,
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 10000,
});

// 테스트용 유니크 이메일 생성
export function uniqueEmail(prefix: string): string {
  return `${prefix}_${Date.now()}_${Math.random().toString(36).slice(2)}@test.local`;
}

// 테스트용 인증 헬퍼
export async function createTestUser(prefix: string = "integration") {
  // Rate limiting 방지
  await delay(200);

  const email = uniqueEmail(prefix);
  const password = "testPassword123!";

  try {
    const response = await testApiClient.post("/auth/signup", {
      email,
      password,
    });

    return {
      email,
      password,
      token: response.data.token,
      user: response.data.user,
    };
  } catch (error: unknown) {
    if (
      error &&
      typeof error === "object" &&
      "response" in error &&
      error.response &&
      typeof error.response === "object" &&
      "status" in error.response &&
      error.response.status === 429
    ) {
      throw new Error(
        "RATE_LIMITED: Too many requests. Wait 1 minute and restart the backend server."
      );
    }
    throw error;
  }
}

// 인증된 API 클라이언트 생성
export function createAuthenticatedClient(token: string) {
  const client = axios.create({
    baseURL: `${API_BASE_URL}/v1`,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    timeout: 10000,
  });
  return client;
}

// 백엔드 서버 상태 확인
export async function checkServerHealth(): Promise<boolean> {
  try {
    // health 엔드포인트가 없으면 간단한 요청으로 확인
    await testApiClient.get("/auth/me", {
      validateStatus: () => true, // 어떤 상태 코드든 에러로 처리하지 않음
    });
    return true;
  } catch (error) {
    if (axios.isAxiosError(error) && error.code === "ECONNREFUSED") {
      return false;
    }
    return true; // 401 등 다른 에러는 서버가 실행 중인 것
  }
}

// 테스트 유저 삭제 (hard delete - 테스트 환경에서만 허용)
export async function deleteTestUser(token: string): Promise<void> {
  try {
    const authClient = createAuthenticatedClient(token);
    await authClient.delete("/auth/me", {
      params: { hard: true },
    });
  } catch (error) {
    // 이미 삭제되었거나 토큰이 만료된 경우 무시
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return;
    }
    // 다른 에러도 테스트 cleanup에서는 무시 (로그만 출력)
    console.warn("Failed to delete test user:", error);
  }
}

// 테스트에서 생성된 유저들을 추적하고 정리하기 위한 헬퍼
export class TestUserManager {
  private createdUsers: Array<{ token: string; email: string }> = [];

  async createUser(prefix: string = "integration") {
    const user = await createTestUser(prefix);
    this.createdUsers.push({ token: user.token, email: user.email });
    return user;
  }

  async cleanup() {
    for (const user of this.createdUsers) {
      await deleteTestUser(user.token);
    }
    this.createdUsers = [];
  }
}
