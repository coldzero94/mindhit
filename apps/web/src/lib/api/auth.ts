import { apiClient } from "./client";
import type {
  AuthLoginRequest,
  AuthSignupRequest,
  AuthAuthResponse,
  AuthUser,
  AuthGoogleAuthRequest,
} from "@/api/generated/types.gen";

export const authApi = {
  login: async (data: AuthLoginRequest): Promise<AuthAuthResponse> => {
    const response = await apiClient.post<AuthAuthResponse>(
      "/auth/login",
      data
    );
    return response.data;
  },

  signup: async (data: AuthSignupRequest): Promise<AuthAuthResponse> => {
    const response = await apiClient.post<AuthAuthResponse>(
      "/auth/signup",
      data
    );
    return response.data;
  },

  googleAuth: async (
    data: AuthGoogleAuthRequest
  ): Promise<AuthAuthResponse> => {
    const response = await apiClient.post<AuthAuthResponse>(
      "/auth/google",
      data
    );
    return response.data;
  },

  me: async (): Promise<AuthUser> => {
    const response = await apiClient.get<{ user: AuthUser }>("/auth/me");
    return response.data.user;
  },

  logout: async (): Promise<void> => {
    await apiClient.post("/auth/logout");
  },

  refresh: async (): Promise<{ token: string }> => {
    const response = await apiClient.post<{ token: string }>("/auth/refresh");
    return response.data;
  },
};
