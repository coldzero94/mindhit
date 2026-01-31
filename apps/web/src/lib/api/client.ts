import axios, { InternalAxiosRequestConfig } from "axios";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:9000";
const API_KEY = process.env.NEXT_PUBLIC_API_KEY || "";

export const apiClient = axios.create({
  baseURL: `${API_BASE_URL}/v1`,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor - API key authentication
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    if (API_KEY) {
      config.headers.Authorization = `Bearer ${API_KEY}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor - error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    // Log errors for debugging
    console.error("API Error:", error.response?.status, error.response?.data);
    return Promise.reject(error);
  }
);

export interface ApiError {
  error: {
    message: string;
    code?: string;
  };
}
