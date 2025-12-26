# Phase 7: Next.js 웹앱

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | Next.js 16.1 App Router 기반 웹앱 기초 구축 |
| **선행 조건** | Phase 2 (인증), Phase 3 (세션 API), Phase 5 (인프라) 완료 |
| **예상 소요** | 4 Steps |
| **결과물** | 인증, 세션 목록 조회 가능한 웹앱 |

---

## 아키텍처

```mermaid
flowchart TB
    subgraph Next.js App
        subgraph Pages
            LOGIN[/auth/login]
            DASH[/dashboard]
            SESS[/sessions]
            MM[/mindmaps/:id]
        end

        subgraph Components
            AUTH[AuthProvider]
            QUERY[QueryProvider]
        end

        subgraph State
            STORE[Zustand Store]
            CACHE[React Query Cache]
        end
    end

    subgraph API Server
        BACKEND[Go Backend]
    end

    LOGIN --> AUTH
    DASH --> QUERY
    SESS --> QUERY
    MM --> QUERY
    AUTH --> STORE
    QUERY --> CACHE
    CACHE -->|API Calls| BACKEND
```

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 7.1 | Next.js 프로젝트 설정 | ✅ |
| 7.2 | 인증 UI 구현 | ✅ |
| 7.3 | API 클라이언트 설정 | ✅ |
| 7.4 | 세션 목록 페이지 | ✅ |

---

## Step 7.1: Next.js 프로젝트 설정

### 체크리스트

- [ ] **프로젝트 생성**

  ```bash
  cd apps
  pnpm create next-app@latest web --typescript --tailwind --eslint --app --src-dir --import-alias "@/*"
  ```

- [ ] **추가 의존성 설치**

  ```bash
  cd apps/web
  pnpm add @tanstack/react-query axios zustand
  pnpm add -D @tanstack/eslint-plugin-query
  ```

- [ ] **shadcn/ui 설정**

  ```bash
  pnpm dlx shadcn@latest init

  # 필요한 컴포넌트 설치
  pnpm dlx shadcn@latest add button input label card form toast
  ```

- [ ] **프로젝트 구조 생성**

  ```
  apps/web/
  ├── src/
  │   ├── app/
  │   │   ├── (auth)/
  │   │   │   ├── login/
  │   │   │   │   └── page.tsx
  │   │   │   ├── signup/
  │   │   │   │   └── page.tsx
  │   │   │   └── layout.tsx
  │   │   ├── (dashboard)/
  │   │   │   ├── sessions/
  │   │   │   │   ├── [id]/
  │   │   │   │   │   └── page.tsx
  │   │   │   │   └── page.tsx
  │   │   │   └── layout.tsx
  │   │   ├── layout.tsx
  │   │   ├── page.tsx
  │   │   └── providers.tsx
  │   ├── components/
  │   │   ├── ui/           # shadcn components
  │   │   ├── auth/
  │   │   ├── sessions/
  │   │   └── layout/
  │   ├── lib/
  │   │   ├── api/
  │   │   │   ├── client.ts
  │   │   │   ├── auth.ts
  │   │   │   └── sessions.ts
  │   │   ├── hooks/
  │   │   └── utils.ts
  │   ├── stores/
  │   │   └── auth-store.ts
  │   └── types/
  │       └── api.ts        # OpenAPI 생성 타입
  ├── next.config.js
  ├── tailwind.config.ts
  └── tsconfig.json
  ```

- [ ] **환경 변수 설정**
  - [ ] `.env.local`

    ```env
    NEXT_PUBLIC_API_URL=http://localhost:8080
    ```

- [ ] **next.config.js 설정**

  ```javascript
  /** @type {import('next').NextConfig} */
  const nextConfig = {
    reactStrictMode: true,
    async rewrites() {
      return [
        {
          source: '/api/:path*',
          destination: `${process.env.NEXT_PUBLIC_API_URL}/v1/:path*`,
        },
      ];
    },
  };

  module.exports = nextConfig;
  ```

- [ ] **moon.yml 설정**

  ```yaml
  language: typescript
  type: application

  tasks:
    dev:
      command: next dev
      local: true

    build:
      command: next build
      inputs:
        - "src/**/*"
        - "public/**/*"
        - "next.config.js"
        - "tailwind.config.ts"
      outputs:
        - ".next"

    start:
      command: next start
      local: true

    lint:
      command: next lint
      inputs:
        - "src/**/*"

    typecheck:
      command: tsc --noEmit
      inputs:
        - "src/**/*"
        - "tsconfig.json"
  ```

### 검증

```bash
cd apps/web
pnpm dev
# http://localhost:3000 접속 확인
```

---

## Step 7.2: 인증 UI 구현

### 체크리스트

- [ ] **Auth Store (Zustand)**
  - [ ] `src/stores/auth-store.ts`

    ```typescript
    import { create } from 'zustand';
    import { persist } from 'zustand/middleware';

    interface User {
      id: string;
      email: string;
    }

    interface AuthState {
      user: User | null;
      token: string | null;
      isAuthenticated: boolean;
      setAuth: (user: User, token: string) => void;
      logout: () => void;
    }

    export const useAuthStore = create<AuthState>()(
      persist(
        (set) => ({
          user: null,
          token: null,
          isAuthenticated: false,
          setAuth: (user, token) =>
            set({ user, token, isAuthenticated: true }),
          logout: () =>
            set({ user: null, token: null, isAuthenticated: false }),
        }),
        {
          name: 'auth-storage',
        }
      )
    );
    ```

- [ ] **로그인 폼 컴포넌트**
  - [ ] `src/components/auth/login-form.tsx`

    ```tsx
    'use client';

    import { useState } from 'react';
    import { useRouter } from 'next/navigation';
    import { useForm } from 'react-hook-form';
    import { zodResolver } from '@hookform/resolvers/zod';
    import { z } from 'zod';

    import { Button } from '@/components/ui/button';
    import { Input } from '@/components/ui/input';
    import { Label } from '@/components/ui/label';
    import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
    import { useToast } from '@/components/ui/use-toast';

    import { useAuthStore } from '@/stores/auth-store';
    import { authApi } from '@/lib/api/auth';

    const loginSchema = z.object({
      email: z.string().email('유효한 이메일을 입력하세요'),
      password: z.string().min(8, '비밀번호는 8자 이상이어야 합니다'),
    });

    type LoginFormData = z.infer<typeof loginSchema>;

    export function LoginForm() {
      const router = useRouter();
      const { toast } = useToast();
      const setAuth = useAuthStore((state) => state.setAuth);
      const [isLoading, setIsLoading] = useState(false);

      const {
        register,
        handleSubmit,
        formState: { errors },
      } = useForm<LoginFormData>({
        resolver: zodResolver(loginSchema),
      });

      const onSubmit = async (data: LoginFormData) => {
        setIsLoading(true);
        try {
          const response = await authApi.login(data);
          setAuth(response.user, response.token);
          toast({
            title: '로그인 성공',
            description: '환영합니다!',
          });
          router.push('/sessions');
        } catch (error) {
          toast({
            title: '로그인 실패',
            description: '이메일 또는 비밀번호를 확인하세요.',
            variant: 'destructive',
          });
        } finally {
          setIsLoading(false);
        }
      };

      return (
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>로그인</CardTitle>
            <CardDescription>MindHit 계정으로 로그인하세요</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">이메일</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="email@example.com"
                  {...register('email')}
                />
                {errors.email && (
                  <p className="text-sm text-red-500">{errors.email.message}</p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">비밀번호</Label>
                <Input
                  id="password"
                  type="password"
                  {...register('password')}
                />
                {errors.password && (
                  <p className="text-sm text-red-500">{errors.password.message}</p>
                )}
              </div>
              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? '로그인 중...' : '로그인'}
              </Button>
            </form>
          </CardContent>
        </Card>
      );
    }
    ```

- [ ] **회원가입 폼 컴포넌트**
  - [ ] `src/components/auth/signup-form.tsx`

    ```tsx
    'use client';

    import { useState } from 'react';
    import { useRouter } from 'next/navigation';
    import { useForm } from 'react-hook-form';
    import { zodResolver } from '@hookform/resolvers/zod';
    import { z } from 'zod';

    import { Button } from '@/components/ui/button';
    import { Input } from '@/components/ui/input';
    import { Label } from '@/components/ui/label';
    import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
    import { useToast } from '@/components/ui/use-toast';

    import { useAuthStore } from '@/stores/auth-store';
    import { authApi } from '@/lib/api/auth';

    const signupSchema = z.object({
      email: z.string().email('유효한 이메일을 입력하세요'),
      password: z.string().min(8, '비밀번호는 8자 이상이어야 합니다'),
      confirmPassword: z.string(),
    }).refine((data) => data.password === data.confirmPassword, {
      message: '비밀번호가 일치하지 않습니다',
      path: ['confirmPassword'],
    });

    type SignupFormData = z.infer<typeof signupSchema>;

    export function SignupForm() {
      const router = useRouter();
      const { toast } = useToast();
      const setAuth = useAuthStore((state) => state.setAuth);
      const [isLoading, setIsLoading] = useState(false);

      const {
        register,
        handleSubmit,
        formState: { errors },
      } = useForm<SignupFormData>({
        resolver: zodResolver(signupSchema),
      });

      const onSubmit = async (data: SignupFormData) => {
        setIsLoading(true);
        try {
          const response = await authApi.signup({
            email: data.email,
            password: data.password,
          });
          setAuth(response.user, response.token);
          toast({
            title: '회원가입 성공',
            description: 'MindHit에 오신 것을 환영합니다!',
          });
          router.push('/sessions');
        } catch (error) {
          toast({
            title: '회원가입 실패',
            description: '이미 사용 중인 이메일입니다.',
            variant: 'destructive',
          });
        } finally {
          setIsLoading(false);
        }
      };

      return (
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>회원가입</CardTitle>
            <CardDescription>MindHit 계정을 만드세요</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">이메일</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="email@example.com"
                  {...register('email')}
                />
                {errors.email && (
                  <p className="text-sm text-red-500">{errors.email.message}</p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">비밀번호</Label>
                <Input
                  id="password"
                  type="password"
                  {...register('password')}
                />
                {errors.password && (
                  <p className="text-sm text-red-500">{errors.password.message}</p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="confirmPassword">비밀번호 확인</Label>
                <Input
                  id="confirmPassword"
                  type="password"
                  {...register('confirmPassword')}
                />
                {errors.confirmPassword && (
                  <p className="text-sm text-red-500">{errors.confirmPassword.message}</p>
                )}
              </div>
              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? '가입 중...' : '회원가입'}
              </Button>
            </form>
          </CardContent>
        </Card>
      );
    }
    ```

- [ ] **로그인 페이지**
  - [ ] `src/app/(auth)/login/page.tsx`

    ```tsx
    import Link from 'next/link';
    import { LoginForm } from '@/components/auth/login-form';

    export default function LoginPage() {
      return (
        <div className="flex min-h-screen items-center justify-center">
          <div className="space-y-4">
            <LoginForm />
            <p className="text-center text-sm text-gray-600">
              계정이 없으신가요?{' '}
              <Link href="/signup" className="text-blue-600 hover:underline">
                회원가입
              </Link>
            </p>
          </div>
        </div>
      );
    }
    ```

- [ ] **회원가입 페이지**
  - [ ] `src/app/(auth)/signup/page.tsx`

    ```tsx
    import Link from 'next/link';
    import { SignupForm } from '@/components/auth/signup-form';

    export default function SignupPage() {
      return (
        <div className="flex min-h-screen items-center justify-center">
          <div className="space-y-4">
            <SignupForm />
            <p className="text-center text-sm text-gray-600">
              이미 계정이 있으신가요?{' '}
              <Link href="/login" className="text-blue-600 hover:underline">
                로그인
              </Link>
            </p>
          </div>
        </div>
      );
    }
    ```

- [ ] **Auth Layout**
  - [ ] `src/app/(auth)/layout.tsx`

    ```tsx
    export default function AuthLayout({
      children,
    }: {
      children: React.ReactNode;
    }) {
      return (
        <div className="min-h-screen bg-gray-50">
          {children}
        </div>
      );
    }
    ```

### 검증

```bash
# 로그인/회원가입 페이지 접속
open http://localhost:3000/login
open http://localhost:3000/signup
```

---

## Step 7.3: API 클라이언트 설정

> **에러 처리 가이드**: Frontend 에러 처리 패턴은
> [09-error-handling.md#10](../09-error-handling.md#10-frontend-에러-처리-nextjs)을 참조하세요.
>
> - Axios interceptor로 401/403/429 에러 처리
> - Toast 메시지 헬퍼 (`useApiError` hook)
> - Error Boundary 컴포넌트
> - React Query 에러 재시도 설정

### 체크리스트

- [ ] **OpenAPI 타입 생성 설정**

  ```bash
  # packages/protocol에서 타입 생성
  cd packages/protocol
  pnpm generate:client

  # 또는 apps/web에서 직접 사용
  cd apps/web
  pnpm add @mindhit/protocol
  ```

- [ ] **Axios 클라이언트**
  - [ ] `src/lib/api/client.ts`

    ```typescript
    import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
    import { useAuthStore } from '@/stores/auth-store';

    const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    export const apiClient = axios.create({
      baseURL: `${API_BASE_URL}/v1`,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor - 토큰 추가
    apiClient.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        const token = useAuthStore.getState().token;
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor - 에러 처리
    apiClient.interceptors.response.use(
      (response) => response,
      (error: AxiosError) => {
        if (error.response?.status === 401) {
          // 토큰 만료 시 로그아웃
          useAuthStore.getState().logout();
          if (typeof window !== 'undefined') {
            window.location.href = '/login';
          }
        }
        return Promise.reject(error);
      }
    );

    export interface ApiError {
      error: {
        message: string;
        code?: string;
      };
    }
    ```

- [ ] **Auth API**
  - [ ] `src/lib/api/auth.ts`

    ```typescript
    import { apiClient } from './client';

    export interface LoginRequest {
      email: string;
      password: string;
    }

    export interface SignupRequest {
      email: string;
      password: string;
    }

    export interface AuthResponse {
      user: {
        id: string;
        email: string;
      };
      token: string;
    }

    export const authApi = {
      login: async (data: LoginRequest): Promise<AuthResponse> => {
        const response = await apiClient.post<AuthResponse>('/auth/login', data);
        return response.data;
      },

      signup: async (data: SignupRequest): Promise<AuthResponse> => {
        const response = await apiClient.post<AuthResponse>('/auth/signup', data);
        return response.data;
      },

      me: async (): Promise<AuthResponse['user']> => {
        const response = await apiClient.get<{ user: AuthResponse['user'] }>('/auth/me');
        return response.data.user;
      },
    };
    ```

- [ ] **Sessions API**
  - [ ] `src/lib/api/sessions.ts`

    ```typescript
    import { apiClient } from './client';

    export type SessionStatus = 'recording' | 'paused' | 'processing' | 'completed' | 'failed';

    export interface Session {
      id: string;
      title: string | null;
      status: SessionStatus;
      started_at: string;
      ended_at: string | null;
      created_at: string;
      updated_at: string;
    }

    export interface SessionWithDetails extends Session {
      page_visits: PageVisit[];
      highlights: Highlight[];
      mindmap: MindmapGraph | null;
    }

    export interface PageVisit {
      id: string;
      url: {
        id: string;
        url: string;
        title: string | null;
      };
      entered_at: string;
      left_at: string | null;
      duration_ms: number | null;
      max_scroll_depth: number;
    }

    export interface Highlight {
      id: string;
      text: string;
      color: string;
      created_at: string;
    }

    export interface MindmapGraph {
      id: string;
      nodes: MindmapNode[];
      edges: MindmapEdge[];
      layout: Record<string, unknown>;
    }

    export interface MindmapNode {
      id: string;
      label: string;
      type: 'core' | 'topic' | 'subtopic' | 'page';
      size: number;
      color: string;
      data: Record<string, unknown>;
    }

    export interface MindmapEdge {
      source: string;
      target: string;
      weight: number;
    }

    export interface SessionListResponse {
      sessions: Session[];
      pagination: {
        total: number;
        page: number;
        per_page: number;
        total_pages: number;
      };
    }

    export const sessionsApi = {
      list: async (page = 1, perPage = 20): Promise<SessionListResponse> => {
        const response = await apiClient.get<SessionListResponse>('/sessions', {
          params: { page, per_page: perPage },
        });
        return response.data;
      },

      get: async (id: string): Promise<SessionWithDetails> => {
        const response = await apiClient.get<{ session: SessionWithDetails }>(`/sessions/${id}`);
        return response.data.session;
      },

      delete: async (id: string): Promise<void> => {
        await apiClient.delete(`/sessions/${id}`);
      },
    };
    ```

- [ ] **React Query 설정**
  - [ ] `src/lib/hooks/use-sessions.ts`

    ```typescript
    import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
    import { sessionsApi, Session, SessionWithDetails } from '@/lib/api/sessions';

    export const sessionKeys = {
      all: ['sessions'] as const,
      lists: () => [...sessionKeys.all, 'list'] as const,
      list: (page: number) => [...sessionKeys.lists(), page] as const,
      details: () => [...sessionKeys.all, 'detail'] as const,
      detail: (id: string) => [...sessionKeys.details(), id] as const,
    };

    export function useSessions(page = 1) {
      return useQuery({
        queryKey: sessionKeys.list(page),
        queryFn: () => sessionsApi.list(page),
      });
    }

    export function useSession(id: string) {
      return useQuery({
        queryKey: sessionKeys.detail(id),
        queryFn: () => sessionsApi.get(id),
        enabled: !!id,
      });
    }

    export function useDeleteSession() {
      const queryClient = useQueryClient();

      return useMutation({
        mutationFn: (id: string) => sessionsApi.delete(id),
        onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: sessionKeys.lists() });
        },
      });
    }
    ```

- [ ] **Providers 설정**
  - [ ] `src/app/providers.tsx`

    ```tsx
    'use client';

    import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
    import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
    import { useState } from 'react';
    import { Toaster } from '@/components/ui/toaster';

    export function Providers({ children }: { children: React.ReactNode }) {
      const [queryClient] = useState(
        () =>
          new QueryClient({
            defaultOptions: {
              queries: {
                staleTime: 60 * 1000, // 1분
                retry: 1,
              },
            },
          })
      );

      return (
        <QueryClientProvider client={queryClient}>
          {children}
          <Toaster />
          <ReactQueryDevtools initialIsOpen={false} />
        </QueryClientProvider>
      );
    }
    ```

- [ ] **Root Layout 업데이트**
  - [ ] `src/app/layout.tsx`

    ```tsx
    import type { Metadata } from 'next';
    import { Inter } from 'next/font/google';
    import './globals.css';
    import { Providers } from './providers';

    const inter = Inter({ subsets: ['latin'] });

    export const metadata: Metadata = {
      title: 'MindHit',
      description: 'Transform your browsing history into knowledge',
    };

    export default function RootLayout({
      children,
    }: {
      children: React.ReactNode;
    }) {
      return (
        <html lang="ko">
          <body className={inter.className}>
            <Providers>{children}</Providers>
          </body>
        </html>
      );
    }
    ```

### 검증

```bash
# API 연동 테스트 (백엔드 실행 필요)
# 1. 회원가입 후 로그인
# 2. 네트워크 탭에서 API 호출 확인
# 3. 토큰이 저장되는지 확인
```

---

## Step 7.4: 세션 목록 페이지

### 체크리스트

- [ ] **Dashboard Layout**
  - [ ] `src/app/(dashboard)/layout.tsx`

    ```tsx
    'use client';

    import { useEffect } from 'react';
    import { useRouter } from 'next/navigation';
    import Link from 'next/link';
    import { useAuthStore } from '@/stores/auth-store';
    import { Button } from '@/components/ui/button';

    export default function DashboardLayout({
      children,
    }: {
      children: React.ReactNode;
    }) {
      const router = useRouter();
      const { isAuthenticated, user, logout } = useAuthStore();

      useEffect(() => {
        if (!isAuthenticated) {
          router.push('/login');
        }
      }, [isAuthenticated, router]);

      if (!isAuthenticated) {
        return null;
      }

      const handleLogout = () => {
        logout();
        router.push('/login');
      };

      return (
        <div className="min-h-screen bg-gray-50">
          {/* Header */}
          <header className="bg-white shadow-sm">
            <div className="mx-auto max-w-7xl px-4 py-4 sm:px-6 lg:px-8">
              <div className="flex items-center justify-between">
                <Link href="/sessions" className="text-xl font-bold text-gray-900">
                  MindHit
                </Link>
                <div className="flex items-center gap-4">
                  <span className="text-sm text-gray-600">{user?.email}</span>
                  <Button variant="outline" size="sm" onClick={handleLogout}>
                    로그아웃
                  </Button>
                </div>
              </div>
            </div>
          </header>

          {/* Main content */}
          <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
            {children}
          </main>
        </div>
      );
    }
    ```

- [ ] **세션 카드 컴포넌트**
  - [ ] `src/components/sessions/session-card.tsx`

    ```tsx
    import Link from 'next/link';
    import { formatDistanceToNow } from 'date-fns';
    import { ko } from 'date-fns/locale';
    import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
    import { Badge } from '@/components/ui/badge';
    import { Session, SessionStatus } from '@/lib/api/sessions';

    const statusConfig: Record<SessionStatus, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
      recording: { label: '녹화 중', variant: 'default' },
      paused: { label: '일시정지', variant: 'secondary' },
      processing: { label: '처리 중', variant: 'outline' },
      completed: { label: '완료', variant: 'default' },
      failed: { label: '실패', variant: 'destructive' },
    };

    interface SessionCardProps {
      session: Session;
    }

    export function SessionCard({ session }: SessionCardProps) {
      const status = statusConfig[session.status];
      const timeAgo = formatDistanceToNow(new Date(session.started_at), {
        addSuffix: true,
        locale: ko,
      });

      return (
        <Link href={`/sessions/${session.id}`}>
          <Card className="hover:shadow-md transition-shadow cursor-pointer">
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <CardTitle className="text-lg">
                  {session.title || '제목 없음'}
                </CardTitle>
                <Badge variant={status.variant}>{status.label}</Badge>
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-gray-500">{timeAgo}</p>
            </CardContent>
          </Card>
        </Link>
      );
    }
    ```

- [ ] **세션 목록 컴포넌트**
  - [ ] `src/components/sessions/session-list.tsx`

    ```tsx
    'use client';

    import { useSessions } from '@/lib/hooks/use-sessions';
    import { SessionCard } from './session-card';
    import { Button } from '@/components/ui/button';
    import { Skeleton } from '@/components/ui/skeleton';

    interface SessionListProps {
      page: number;
      onPageChange: (page: number) => void;
    }

    export function SessionList({ page, onPageChange }: SessionListProps) {
      const { data, isLoading, error } = useSessions(page);

      if (isLoading) {
        return (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-32" />
            ))}
          </div>
        );
      }

      if (error) {
        return (
          <div className="text-center py-8">
            <p className="text-red-500">세션을 불러오는데 실패했습니다.</p>
          </div>
        );
      }

      if (!data?.sessions.length) {
        return (
          <div className="text-center py-8">
            <p className="text-gray-500">아직 녹화된 세션이 없습니다.</p>
            <p className="text-sm text-gray-400 mt-2">
              Chrome Extension을 사용하여 첫 번째 세션을 녹화해보세요.
            </p>
          </div>
        );
      }

      return (
        <div className="space-y-6">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {data.sessions.map((session) => (
              <SessionCard key={session.id} session={session} />
            ))}
          </div>

          {/* Pagination */}
          {data.pagination.total_pages > 1 && (
            <div className="flex justify-center gap-2">
              <Button
                variant="outline"
                disabled={page <= 1}
                onClick={() => onPageChange(page - 1)}
              >
                이전
              </Button>
              <span className="flex items-center px-4 text-sm">
                {page} / {data.pagination.total_pages}
              </span>
              <Button
                variant="outline"
                disabled={page >= data.pagination.total_pages}
                onClick={() => onPageChange(page + 1)}
              >
                다음
              </Button>
            </div>
          )}
        </div>
      );
    }
    ```

- [ ] **세션 목록 페이지**
  - [ ] `src/app/(dashboard)/sessions/page.tsx`

    ```tsx
    'use client';

    import { useState } from 'react';
    import { SessionList } from '@/components/sessions/session-list';

    export default function SessionsPage() {
      const [page, setPage] = useState(1);

      return (
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-bold">내 세션</h1>
          </div>
          <SessionList page={page} onPageChange={setPage} />
        </div>
      );
    }
    ```

- [ ] **세션 상세 페이지 (기본)**
  - [ ] `src/app/(dashboard)/sessions/[id]/page.tsx`

    ```tsx
    'use client';

    import { useParams, useRouter } from 'next/navigation';
    import { formatDistanceToNow, format } from 'date-fns';
    import { ko } from 'date-fns/locale';
    import { ArrowLeft, Trash2 } from 'lucide-react';

    import { useSession, useDeleteSession } from '@/lib/hooks/use-sessions';
    import { Button } from '@/components/ui/button';
    import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
    import { Badge } from '@/components/ui/badge';
    import { Skeleton } from '@/components/ui/skeleton';
    import { useToast } from '@/components/ui/use-toast';
    import {
      AlertDialog,
      AlertDialogAction,
      AlertDialogCancel,
      AlertDialogContent,
      AlertDialogDescription,
      AlertDialogFooter,
      AlertDialogHeader,
      AlertDialogTitle,
      AlertDialogTrigger,
    } from '@/components/ui/alert-dialog';

    export default function SessionDetailPage() {
      const params = useParams();
      const router = useRouter();
      const { toast } = useToast();
      const sessionId = params.id as string;

      const { data: session, isLoading, error } = useSession(sessionId);
      const deleteSession = useDeleteSession();

      const handleDelete = async () => {
        try {
          await deleteSession.mutateAsync(sessionId);
          toast({
            title: '세션 삭제됨',
            description: '세션이 성공적으로 삭제되었습니다.',
          });
          router.push('/sessions');
        } catch (error) {
          toast({
            title: '삭제 실패',
            description: '세션을 삭제하는데 실패했습니다.',
            variant: 'destructive',
          });
        }
      };

      if (isLoading) {
        return (
          <div className="space-y-4">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-64" />
          </div>
        );
      }

      if (error || !session) {
        return (
          <div className="text-center py-8">
            <p className="text-red-500">세션을 찾을 수 없습니다.</p>
            <Button variant="outline" className="mt-4" onClick={() => router.push('/sessions')}>
              목록으로 돌아가기
            </Button>
          </div>
        );
      }

      return (
        <div className="space-y-6">
          {/* Header */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Button variant="ghost" size="icon" onClick={() => router.push('/sessions')}>
                <ArrowLeft className="h-5 w-5" />
              </Button>
              <div>
                <h1 className="text-2xl font-bold">{session.title || '제목 없음'}</h1>
                <p className="text-sm text-gray-500">
                  {format(new Date(session.started_at), 'yyyy년 MM월 dd일 HH:mm', { locale: ko })}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Badge>{session.status}</Badge>
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="destructive" size="icon">
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>세션을 삭제하시겠습니까?</AlertDialogTitle>
                    <AlertDialogDescription>
                      이 작업은 되돌릴 수 없습니다. 세션과 관련된 모든 데이터가 영구적으로 삭제됩니다.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>취소</AlertDialogCancel>
                    <AlertDialogAction onClick={handleDelete}>삭제</AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </div>
          </div>

          {/* Stats */}
          <div className="grid gap-4 md:grid-cols-3">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-gray-500">
                  방문한 페이지
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">{session.page_visits.length}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-gray-500">
                  하이라이트
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">{session.highlights.length}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-gray-500">
                  마인드맵
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">
                  {session.mindmap ? '생성됨' : '없음'}
                </p>
              </CardContent>
            </Card>
          </div>

          {/* Page Visits */}
          <Card>
            <CardHeader>
              <CardTitle>방문한 페이지</CardTitle>
            </CardHeader>
            <CardContent>
              {session.page_visits.length === 0 ? (
                <p className="text-gray-500">방문한 페이지가 없습니다.</p>
              ) : (
                <ul className="space-y-2">
                  {session.page_visits.map((visit) => (
                    <li key={visit.id} className="flex items-center justify-between p-2 rounded hover:bg-gray-50">
                      <div className="flex-1 min-w-0">
                        <p className="font-medium truncate">
                          {visit.url.title || visit.url.url}
                        </p>
                        <p className="text-sm text-gray-500 truncate">{visit.url.url}</p>
                      </div>
                      <div className="text-sm text-gray-400 ml-4">
                        {visit.duration_ms
                          ? `${Math.floor(visit.duration_ms / 60000)}분 ${Math.floor((visit.duration_ms % 60000) / 1000)}초`
                          : '-'}
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </CardContent>
          </Card>

          {/* Highlights */}
          {session.highlights.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>하이라이트</CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="space-y-2">
                  {session.highlights.map((highlight) => (
                    <li
                      key={highlight.id}
                      className="p-3 rounded border-l-4"
                      style={{ borderColor: highlight.color }}
                    >
                      <p className="text-sm">{highlight.text}</p>
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          )}
        </div>
      );
    }
    ```

- [ ] **추가 의존성 설치**

  ```bash
  pnpm add date-fns lucide-react
  pnpm dlx shadcn@latest add alert-dialog skeleton badge
  ```

### 검증

```bash
# 세션 목록 페이지 접속
open http://localhost:3000/sessions

# 확인 사항:
# 1. 로그인 없이 접근 시 로그인 페이지로 리다이렉트
# 2. 로그인 후 세션 목록 표시
# 3. 세션 클릭 시 상세 페이지 이동
# 4. 삭제 버튼 동작
```

---

## Phase 7 완료 확인

### 전체 검증 체크리스트

- [ ] 회원가입 동작
- [ ] 로그인/로그아웃 동작
- [ ] 인증 상태 유지 (새로고침 후)
- [ ] 세션 목록 조회
- [ ] 세션 상세 조회
- [ ] 세션 삭제
- [ ] 페이지네이션 동작

### 테스트 요구사항

| 테스트 유형 | 대상 | 도구 |
| ----------- | ---- | ---- |
| 컴포넌트 테스트 | Auth 컴포넌트 | Vitest + React Testing Library |
| 컴포넌트 테스트 | Session 컴포넌트 | Vitest + React Testing Library |
| E2E 테스트 | 로그인 플로우 | Playwright (Phase 10 이후) |

```bash
# Phase 7 테스트 실행
moonx web:test
```

> **Note**: 웹앱 테스트는 API 서버가 실행 중이어야 합니다 (MSW mock 또는 실제 서버).

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| Next.js 프로젝트 | `apps/web/` |
| Auth Store | `src/stores/auth-store.ts` |
| API 클라이언트 | `src/lib/api/` |
| React Query Hooks | `src/lib/hooks/` |
| 인증 컴포넌트 | `src/components/auth/` |
| 세션 컴포넌트 | `src/components/sessions/` |
| 테스트 | `src/**/*.test.tsx` |

---

## 다음 Phase

Phase 7 완료 후 [Phase 8: Chrome Extension](./phase-8-extension.md)으로 진행하세요.
