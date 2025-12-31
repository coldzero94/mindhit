# Phase 8.1: Extension 기능 보완

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | Chrome Extension UX 개선 및 누락된 기능 추가 |
| **선행 조건** | Phase 8 완료 |
| **예상 소요** | 3 Steps |
| **결과물** | 세션 목록, 설정, 웹 연동 기능이 추가된 Extension |

---

## 현재 상태 분석

### 구현 완료 (Phase 8)
- [x] Google OAuth 로그인
- [x] 세션 시작/일시정지/재개/종료
- [x] 페이지 방문 이벤트 수집
- [x] 하이라이트 수집
- [x] 오프라인 배치 저장

### 미구현 (이번 Phase에서 추가)
- [ ] 세션 목록 보기
- [ ] 웹 대시보드로 이동 버튼
- [ ] 세션 제목 입력/수정
- [ ] 네트워크 상태 표시
- [ ] 설정 페이지

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 8.1.1 | 세션 목록 및 웹 연동 | ⬜ |
| 8.1.2 | 세션 제목 입력 | ⬜ |
| 8.1.3 | 네트워크 상태 및 설정 | ⬜ |

---

## Step 8.1.1: 세션 목록 및 웹 연동

### 체크리스트

- [ ] **API 함수 추가**
  - [ ] `src/lib/api.ts` 업데이트

    ```typescript
    // 기존 api 객체에 추가

    interface SessionListResponse {
      sessions: Array<{
        id: string;
        title: string | null;
        session_status: 'recording' | 'paused' | 'processing' | 'completed' | 'failed';
        started_at: string;
        ended_at: string | null;
      }>;
      total: number;
    }

    export const api = {
      // ... 기존 메서드들

      // 세션 목록 조회
      getSessions: async (token: string, limit: number = 5): Promise<SessionListResponse> => {
        return request<SessionListResponse>(`/sessions?limit=${limit}&sort=started_at:desc`, {
          method: 'GET',
          headers: { Authorization: `Bearer ${token}` },
        });
      },

      // 세션 업데이트 (제목 변경)
      updateSession: async (
        token: string,
        sessionId: string,
        data: { title?: string; description?: string }
      ): Promise<SessionResponse> => {
        return request<SessionResponse>(`/sessions/${sessionId}`, {
          method: 'PATCH',
          headers: { Authorization: `Bearer ${token}` },
          body: JSON.stringify(data),
        });
      },
    };
    ```

- [ ] **세션 목록 컴포넌트**
  - [ ] `src/popup/components/SessionList.tsx`

    ```tsx
    import { useState, useEffect } from 'react';
    import { useAuthStore } from '@/stores/auth-store';
    import { api } from '@/lib/api';
    import { WEB_APP_URL } from '@/lib/constants';

    interface Session {
      id: string;
      title: string | null;
      session_status: string;
      started_at: string;
      ended_at: string | null;
    }

    const statusColors: Record<string, string> = {
      recording: 'bg-red-500',
      paused: 'bg-yellow-500',
      processing: 'bg-blue-500',
      completed: 'bg-green-500',
      failed: 'bg-gray-500',
    };

    const statusLabels: Record<string, string> = {
      recording: '녹화 중',
      paused: '일시정지',
      processing: '처리 중',
      completed: '완료',
      failed: '실패',
    };

    function formatDate(dateString: string): string {
      const date = new Date(dateString);
      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      const diffMins = Math.floor(diffMs / 60000);
      const diffHours = Math.floor(diffMs / 3600000);
      const diffDays = Math.floor(diffMs / 86400000);

      if (diffMins < 1) return '방금 전';
      if (diffMins < 60) return `${diffMins}분 전`;
      if (diffHours < 24) return `${diffHours}시간 전`;
      if (diffDays < 7) return `${diffDays}일 전`;

      return date.toLocaleDateString('ko-KR', {
        month: 'short',
        day: 'numeric',
      });
    }

    export function SessionList() {
      const { token } = useAuthStore();
      const [sessions, setSessions] = useState<Session[]>([]);
      const [isLoading, setIsLoading] = useState(true);
      const [error, setError] = useState<string | null>(null);

      useEffect(() => {
        if (!token) return;

        const fetchSessions = async () => {
          try {
            setIsLoading(true);
            const response = await api.getSessions(token, 5);
            setSessions(response.sessions);
          } catch (err) {
            setError('세션 목록을 불러올 수 없습니다');
            console.error('Failed to fetch sessions:', err);
          } finally {
            setIsLoading(false);
          }
        };

        fetchSessions();
      }, [token]);

      const handleSessionClick = (sessionId: string) => {
        chrome.tabs.create({
          url: `${WEB_APP_URL}/sessions/${sessionId}`,
        });
      };

      const handleViewAll = () => {
        chrome.tabs.create({
          url: `${WEB_APP_URL}/sessions`,
        });
      };

      if (isLoading) {
        return (
          <div className="bg-white rounded-xl p-4 shadow-sm">
            <div className="animate-pulse space-y-3">
              <div className="h-4 bg-gray-200 rounded w-24" />
              <div className="h-12 bg-gray-100 rounded" />
              <div className="h-12 bg-gray-100 rounded" />
            </div>
          </div>
        );
      }

      if (error) {
        return (
          <div className="bg-white rounded-xl p-4 shadow-sm">
            <p className="text-sm text-red-500">{error}</p>
          </div>
        );
      }

      if (sessions.length === 0) {
        return (
          <div className="bg-white rounded-xl p-4 shadow-sm">
            <p className="text-sm text-gray-500 text-center">
              아직 세션이 없습니다
            </p>
          </div>
        );
      }

      return (
        <div className="bg-white rounded-xl shadow-sm overflow-hidden">
          <div className="px-4 py-3 border-b border-gray-100">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-medium text-gray-900">최근 세션</h3>
              <button
                onClick={handleViewAll}
                className="text-xs text-blue-600 hover:text-blue-700"
              >
                전체 보기
              </button>
            </div>
          </div>

          <ul className="divide-y divide-gray-100">
            {sessions.map((session) => (
              <li key={session.id}>
                <button
                  onClick={() => handleSessionClick(session.id)}
                  className="w-full px-4 py-3 flex items-center gap-3 hover:bg-gray-50 transition-colors text-left"
                >
                  <div
                    className={`w-2 h-2 rounded-full ${statusColors[session.session_status]}`}
                  />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 truncate">
                      {session.title || '제목 없음'}
                    </p>
                    <p className="text-xs text-gray-500">
                      {formatDate(session.started_at)} · {statusLabels[session.session_status]}
                    </p>
                  </div>
                  <svg
                    className="w-4 h-4 text-gray-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                    />
                  </svg>
                </button>
              </li>
            ))}
          </ul>
        </div>
      );
    }
    ```

- [ ] **상수 파일 업데이트**
  - [ ] `src/lib/constants.ts`

    ```typescript
    // 기존 내용에 추가
    export const WEB_APP_URL = import.meta.env.VITE_WEB_URL || 'http://localhost:3000';
    ```

- [ ] **대시보드 바로가기 버튼**
  - [ ] `src/popup/components/DashboardLink.tsx`

    ```tsx
    import { WEB_APP_URL } from '@/lib/constants';
    import { useSessionStore } from '@/stores/session-store';

    export function DashboardLink() {
      const { sessionId, status } = useSessionStore();

      const handleClick = () => {
        // 현재 녹화 중인 세션이 있으면 해당 세션으로, 아니면 목록으로
        const url = sessionId && status !== 'idle'
          ? `${WEB_APP_URL}/sessions/${sessionId}`
          : `${WEB_APP_URL}/sessions`;

        chrome.tabs.create({ url });
      };

      return (
        <button
          onClick={handleClick}
          className="w-full flex items-center justify-center gap-2 px-4 py-2 text-sm text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
        >
          <svg
            className="w-4 h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
            />
          </svg>
          웹 대시보드에서 보기
        </button>
      );
    }
    ```

- [ ] **App.tsx 업데이트**
  - [ ] `src/popup/App.tsx`

    ```tsx
    import { useEffect, useState } from "react";
    import { useAuthStore } from "@/stores/auth-store";
    import { useSessionStore } from "@/stores/session-store";
    import { SessionControl } from "./components/SessionControl";
    import { SessionStats } from "./components/SessionStats";
    import { SessionList } from "./components/SessionList";
    import { DashboardLink } from "./components/DashboardLink";
    import { LoginPrompt } from "./components/LoginPrompt";

    export function App() {
      const { isAuthenticated, user, logout } = useAuthStore();
      const { status, updateElapsedTime, incrementPageCount, incrementHighlightCount } = useSessionStore();
      const [isHydrated, setIsHydrated] = useState(false);

      // Wait for Zustand to hydrate from chrome.storage
      useEffect(() => {
        const unsubAuth = useAuthStore.persist.onFinishHydration(() => {
          setIsHydrated(true);
        });

        if (useAuthStore.persist.hasHydrated()) {
          setIsHydrated(true);
        }

        return () => {
          unsubAuth();
        };
      }, []);

      // Update elapsed time
      useEffect(() => {
        if (status === "recording") {
          const interval = setInterval(updateElapsedTime, 1000);
          return () => clearInterval(interval);
        }
      }, [status, updateElapsedTime]);

      // Listen for page count and highlight count updates
      useEffect(() => {
        const handleMessage = (message: { type: string }) => {
          if (message.type === "INCREMENT_PAGE_COUNT") {
            incrementPageCount();
          } else if (message.type === "INCREMENT_HIGHLIGHT_COUNT") {
            incrementHighlightCount();
          }
        };

        chrome.runtime.onMessage.addListener(handleMessage);
        return () => chrome.runtime.onMessage.removeListener(handleMessage);
      }, [incrementPageCount, incrementHighlightCount]);

      if (!isHydrated) {
        return (
          <div className="p-4 flex items-center justify-center min-h-[200px]">
            <div className="text-sm text-gray-500">Loading...</div>
          </div>
        );
      }

      if (!isAuthenticated) {
        return <LoginPrompt />;
      }

      return (
        <div className="p-3 space-y-3 min-w-[320px]">
          {/* Header */}
          <header className="flex items-center justify-between">
            <h1 className="text-base font-bold text-gray-900">MindHit</h1>
            <div className="flex items-center gap-2">
              <span className="text-xs text-gray-500 truncate max-w-[120px]">{user?.email}</span>
              <button
                onClick={logout}
                className="text-xs text-gray-400 hover:text-gray-600"
              >
                Logout
              </button>
            </div>
          </header>

          {/* Session Control */}
          <SessionControl />

          {/* Session Stats */}
          {status !== "idle" && <SessionStats />}

          {/* Dashboard Link */}
          <DashboardLink />

          {/* Session List (only when idle) */}
          {status === "idle" && <SessionList />}
        </div>
      );
    }
    ```

### 검증

```bash
moonx extension:build
# Extension 리로드 후:
# 1. 로그인 상태에서 세션 목록이 표시되는지 확인
# 2. 세션 클릭 시 웹 대시보드가 열리는지 확인
# 3. "웹 대시보드에서 보기" 버튼 동작 확인
# 4. "전체 보기" 클릭 시 세션 목록 페이지로 이동 확인
```

---

## Step 8.1.2: 세션 제목 입력

### 체크리스트

- [ ] **세션 제목 입력 컴포넌트**
  - [ ] `src/popup/components/SessionTitleInput.tsx`

    ```tsx
    import { useState, useEffect, useRef } from 'react';
    import { useSessionStore } from '@/stores/session-store';
    import { useAuthStore } from '@/stores/auth-store';
    import { api } from '@/lib/api';

    export function SessionTitleInput() {
      const { sessionId, title, setTitle } = useSessionStore();
      const { token } = useAuthStore();
      const [localTitle, setLocalTitle] = useState(title || '');
      const [isEditing, setIsEditing] = useState(false);
      const [isSaving, setIsSaving] = useState(false);
      const inputRef = useRef<HTMLInputElement>(null);
      const saveTimeoutRef = useRef<ReturnType<typeof setTimeout>>();

      useEffect(() => {
        setLocalTitle(title || '');
      }, [title]);

      useEffect(() => {
        if (isEditing && inputRef.current) {
          inputRef.current.focus();
          inputRef.current.select();
        }
      }, [isEditing]);

      const handleSave = async (newTitle: string) => {
        if (!token || !sessionId) return;

        // Clear any pending save
        if (saveTimeoutRef.current) {
          clearTimeout(saveTimeoutRef.current);
        }

        // Debounce save (500ms)
        saveTimeoutRef.current = setTimeout(async () => {
          setIsSaving(true);
          try {
            await api.updateSession(token, sessionId, { title: newTitle || undefined });
            setTitle(newTitle);
          } catch (error) {
            console.error('Failed to update session title:', error);
          } finally {
            setIsSaving(false);
          }
        }, 500);
      };

      const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newTitle = e.target.value;
        setLocalTitle(newTitle);
        handleSave(newTitle);
      };

      const handleBlur = () => {
        setIsEditing(false);
      };

      const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') {
          setIsEditing(false);
          inputRef.current?.blur();
        }
        if (e.key === 'Escape') {
          setLocalTitle(title || '');
          setIsEditing(false);
        }
      };

      if (!sessionId) return null;

      return (
        <div className="flex items-center gap-2">
          {isEditing ? (
            <input
              ref={inputRef}
              type="text"
              value={localTitle}
              onChange={handleChange}
              onBlur={handleBlur}
              onKeyDown={handleKeyDown}
              placeholder="세션 제목 입력..."
              className="flex-1 px-2 py-1 text-sm border border-blue-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              maxLength={100}
            />
          ) : (
            <button
              onClick={() => setIsEditing(true)}
              className="flex-1 flex items-center gap-2 px-2 py-1 text-sm text-left hover:bg-gray-100 rounded transition-colors"
            >
              <span className="truncate text-gray-700">
                {localTitle || '제목 없음'}
              </span>
              <svg
                className="w-3 h-3 text-gray-400 flex-shrink-0"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
                />
              </svg>
            </button>
          )}
          {isSaving && (
            <span className="text-xs text-gray-400">저장 중...</span>
          )}
        </div>
      );
    }
    ```

- [ ] **세션 Store 업데이트**
  - [ ] `src/stores/session-store.ts` 에 title 필드 추가

    ```typescript
    // 기존 SessionState interface에 추가
    interface SessionState {
      sessionId: string | null;
      status: SessionStatus;
      startedAt: number | null;
      pageCount: number;
      highlightCount: number;
      elapsedSeconds: number;
      title: string | null;  // 추가

      startSession: (sessionId: string, title?: string) => void;  // 수정
      setTitle: (title: string) => void;  // 추가
      // ... 기존 메서드들
    }

    export const useSessionStore = create<SessionState>()(
      persist(
        (set, get) => ({
          // ... 기존 상태
          title: null,

          startSession: (sessionId, title) =>
            set({
              sessionId,
              status: "recording",
              startedAt: Date.now(),
              pageCount: 0,
              highlightCount: 0,
              elapsedSeconds: 0,
              title: title || null,
            }),

          setTitle: (title) => set({ title }),

          stopSession: () =>
            set({
              sessionId: null,
              status: "idle",
              startedAt: null,
              title: null,
            }),

          reset: () =>
            set({
              sessionId: null,
              status: "idle",
              startedAt: null,
              pageCount: 0,
              highlightCount: 0,
              elapsedSeconds: 0,
              title: null,
            }),

          // ... 기존 메서드들
        }),
        // ... persist 설정
      )
    );
    ```

- [ ] **세션 시작 시 제목 입력 옵션**
  - [ ] `src/popup/components/SessionControl.tsx` 수정

    ```tsx
    // handleStart 함수 수정
    const handleStart = async () => {
      if (!token) return;
      setIsLoading(true);
      try {
        // 기본 제목 생성 (날짜 기반)
        const defaultTitle = new Date().toLocaleDateString('ko-KR', {
          month: 'long',
          day: 'numeric',
          hour: '2-digit',
          minute: '2-digit',
        }) + ' 세션';

        const response = await api.startSession(token, defaultTitle);
        startSession(response.session.id, defaultTitle);

        chrome.runtime.sendMessage({
          type: 'SESSION_STARTED',
          sessionId: response.session.id,
        });
      } catch (error) {
        console.error('Failed to start session:', error);
      } finally {
        setIsLoading(false);
      }
    };
    ```

- [ ] **API 시작 함수 수정**
  - [ ] `src/lib/api.ts`

    ```typescript
    // startSession 메서드 수정
    startSession: async (token: string, title?: string): Promise<SessionResponse> => {
      return request<SessionResponse>('/sessions/start', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: JSON.stringify({ title }),
      });
    },
    ```

- [ ] **SessionControl에 제목 표시 추가**
  - [ ] `SessionControl.tsx`의 녹화 중 UI에 `SessionTitleInput` 추가

    ```tsx
    // return 부분 수정 (녹화 중일 때)
    return (
      <div className="bg-white rounded-xl p-4 shadow-sm">
        {/* 제목 입력 */}
        <SessionTitleInput />

        <div className="flex items-center justify-between my-3">
          <div className="flex items-center gap-2">
            <div className={`w-3 h-3 rounded-full ${status === 'recording' ? 'bg-red-500 animate-pulse' : 'bg-yellow-500'}`} />
            <span className="font-medium text-gray-900">
              {status === 'recording' ? '녹화 중' : '일시정지'}
            </span>
          </div>
        </div>

        {/* 버튼들 */}
        <div className="flex gap-2">
          {/* ... 기존 버튼 코드 */}
        </div>
      </div>
    );
    ```

### 검증

```bash
moonx extension:build
# Extension 리로드 후:
# 1. 세션 시작 시 자동 생성된 제목 확인
# 2. 제목 클릭하여 편집 모드 진입 확인
# 3. 제목 수정 후 자동 저장 확인
# 4. 웹 대시보드에서 제목이 반영되었는지 확인
# 5. Enter로 편집 완료, Escape로 취소 확인
```

---

## Step 8.1.3: 네트워크 상태 및 설정

### 체크리스트

- [ ] **네트워크 상태 Hook**
  - [ ] `src/lib/use-network-status.ts`

    ```typescript
    import { useState, useEffect } from 'react';

    interface NetworkStatus {
      isOnline: boolean;
      wasOffline: boolean;  // 오프라인이었다가 복구된 상태
    }

    export function useNetworkStatus(): NetworkStatus {
      const [status, setStatus] = useState<NetworkStatus>({
        isOnline: navigator.onLine,
        wasOffline: false,
      });

      useEffect(() => {
        const handleOnline = () => {
          setStatus((prev) => ({
            isOnline: true,
            wasOffline: !prev.isOnline,  // 이전에 오프라인이었으면 true
          }));

          // 3초 후 wasOffline 리셋
          setTimeout(() => {
            setStatus((prev) => ({ ...prev, wasOffline: false }));
          }, 3000);
        };

        const handleOffline = () => {
          setStatus({ isOnline: false, wasOffline: false });
        };

        window.addEventListener('online', handleOnline);
        window.addEventListener('offline', handleOffline);

        return () => {
          window.removeEventListener('online', handleOnline);
          window.removeEventListener('offline', handleOffline);
        };
      }, []);

      return status;
    }
    ```

- [ ] **네트워크 상태 배너**
  - [ ] `src/popup/components/NetworkBanner.tsx`

    ```tsx
    import { useNetworkStatus } from '@/lib/use-network-status';

    export function NetworkBanner() {
      const { isOnline, wasOffline } = useNetworkStatus();

      if (isOnline && !wasOffline) return null;

      if (!isOnline) {
        return (
          <div className="bg-red-500 text-white px-3 py-2 text-sm flex items-center gap-2">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M18.364 5.636a9 9 0 010 12.728m-3.536-3.536a4 4 0 010-5.656m-7.072 7.072a9 9 0 010-12.728m3.536 3.536a4 4 0 010 5.656"
              />
            </svg>
            <span>오프라인 상태입니다. 이벤트는 로컬에 저장됩니다.</span>
          </div>
        );
      }

      // wasOffline인 경우 (복구됨)
      return (
        <div className="bg-green-500 text-white px-3 py-2 text-sm flex items-center gap-2">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M5 13l4 4L19 7"
            />
          </svg>
          <span>네트워크 복구됨. 데이터 동기화 중...</span>
        </div>
      );
    }
    ```

- [ ] **설정 Store**
  - [ ] `src/stores/settings-store.ts`

    ```typescript
    import { create } from 'zustand';
    import { persist, createJSONStorage, StateStorage } from 'zustand/middleware';

    interface Settings {
      // 개발자 설정
      apiUrl: string;
      webUrl: string;
      // 사용자 설정
      autoStartSession: boolean;
      collectAllTabs: boolean;  // true: 모든 탭, false: 현재 탭만
    }

    interface SettingsState extends Settings {
      updateSettings: (settings: Partial<Settings>) => void;
      resetSettings: () => void;
    }

    const defaultSettings: Settings = {
      apiUrl: import.meta.env.VITE_API_URL || 'http://localhost:9000/v1',
      webUrl: import.meta.env.VITE_WEB_URL || 'http://localhost:3000',
      autoStartSession: false,
      collectAllTabs: true,
    };

    // Chrome storage adapter
    const chromeStorage: StateStorage = {
      getItem: async (name: string): Promise<string | null> => {
        const result = await chrome.storage.sync.get(name);  // sync로 변경 (기기 간 동기화)
        const value = result[name];
        return typeof value === 'string' ? value : null;
      },
      setItem: async (name: string, value: string): Promise<void> => {
        await chrome.storage.sync.set({ [name]: value });
      },
      removeItem: async (name: string): Promise<void> => {
        await chrome.storage.sync.remove(name);
      },
    };

    export const useSettingsStore = create<SettingsState>()(
      persist(
        (set) => ({
          ...defaultSettings,

          updateSettings: (newSettings) =>
            set((state) => ({ ...state, ...newSettings })),

          resetSettings: () => set(defaultSettings),
        }),
        {
          name: 'mindhit-settings',
          storage: createJSONStorage(() => chromeStorage),
        }
      )
    );
    ```

- [ ] **설정 페이지 컴포넌트**
  - [ ] `src/popup/components/Settings.tsx`

    ```tsx
    import { useState, useEffect } from 'react';
    import { useSettingsStore } from '@/stores/settings-store';

    interface SettingsProps {
      onClose: () => void;
    }

    export function Settings({ onClose }: SettingsProps) {
      const { apiUrl, webUrl, autoStartSession, collectAllTabs, updateSettings, resetSettings } =
        useSettingsStore();

      const [localApiUrl, setLocalApiUrl] = useState(apiUrl);
      const [localWebUrl, setLocalWebUrl] = useState(webUrl);
      const [showAdvanced, setShowAdvanced] = useState(false);

      useEffect(() => {
        // Store hydration
        const unsub = useSettingsStore.persist.onFinishHydration(() => {
          setLocalApiUrl(useSettingsStore.getState().apiUrl);
          setLocalWebUrl(useSettingsStore.getState().webUrl);
        });

        if (useSettingsStore.persist.hasHydrated()) {
          setLocalApiUrl(apiUrl);
          setLocalWebUrl(webUrl);
        }

        return unsub;
      }, [apiUrl, webUrl]);

      const handleSave = () => {
        updateSettings({
          apiUrl: localApiUrl,
          webUrl: localWebUrl,
        });
        onClose();
      };

      const handleReset = () => {
        if (confirm('설정을 초기화하시겠습니까?')) {
          resetSettings();
          setLocalApiUrl(import.meta.env.VITE_API_URL || 'http://localhost:9000/v1');
          setLocalWebUrl(import.meta.env.VITE_WEB_URL || 'http://localhost:3000');
        }
      };

      return (
        <div className="fixed inset-0 bg-white z-50 overflow-auto">
          {/* Header */}
          <div className="sticky top-0 bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between">
            <h2 className="text-lg font-semibold">설정</h2>
            <button
              onClick={onClose}
              className="p-1 hover:bg-gray-100 rounded-full"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div className="p-4 space-y-6">
            {/* 일반 설정 */}
            <section className="space-y-4">
              <h3 className="text-sm font-medium text-gray-900">일반</h3>

              <label className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-700">자동 세션 시작</p>
                  <p className="text-xs text-gray-500">브라우저 시작 시 자동으로 세션 시작</p>
                </div>
                <input
                  type="checkbox"
                  checked={autoStartSession}
                  onChange={(e) => updateSettings({ autoStartSession: e.target.checked })}
                  className="w-4 h-4 text-blue-600 rounded"
                />
              </label>

              <label className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-700">모든 탭 수집</p>
                  <p className="text-xs text-gray-500">비활성화 시 현재 탭만 수집</p>
                </div>
                <input
                  type="checkbox"
                  checked={collectAllTabs}
                  onChange={(e) => updateSettings({ collectAllTabs: e.target.checked })}
                  className="w-4 h-4 text-blue-600 rounded"
                />
              </label>
            </section>

            {/* 고급 설정 */}
            <section className="space-y-4">
              <button
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="flex items-center gap-2 text-sm font-medium text-gray-900"
              >
                <svg
                  className={`w-4 h-4 transition-transform ${showAdvanced ? 'rotate-90' : ''}`}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
                고급 설정 (개발자용)
              </button>

              {showAdvanced && (
                <div className="space-y-4 pl-6">
                  <div>
                    <label className="block text-sm text-gray-700 mb-1">API URL</label>
                    <input
                      type="text"
                      value={localApiUrl}
                      onChange={(e) => setLocalApiUrl(e.target.value)}
                      className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg"
                      placeholder="http://localhost:9000/v1"
                    />
                  </div>

                  <div>
                    <label className="block text-sm text-gray-700 mb-1">Web App URL</label>
                    <input
                      type="text"
                      value={localWebUrl}
                      onChange={(e) => setLocalWebUrl(e.target.value)}
                      className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg"
                      placeholder="http://localhost:3000"
                    />
                  </div>
                </div>
              )}
            </section>

            {/* 버튼 */}
            <div className="flex gap-2 pt-4 border-t">
              <button
                onClick={handleReset}
                className="flex-1 px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                초기화
              </button>
              <button
                onClick={handleSave}
                className="flex-1 px-4 py-2 text-sm text-white bg-blue-600 hover:bg-blue-700 rounded-lg"
              >
                저장
              </button>
            </div>

            {/* 버전 정보 */}
            <div className="text-center text-xs text-gray-400 pt-4">
              MindHit Extension v{chrome.runtime.getManifest().version}
            </div>
          </div>
        </div>
      );
    }
    ```

- [ ] **설정 버튼 추가**
  - [ ] `src/popup/App.tsx` 헤더에 설정 아이콘 추가

    ```tsx
    // App.tsx 수정
    import { useState } from 'react';
    import { Settings } from './components/Settings';
    import { NetworkBanner } from './components/NetworkBanner';

    export function App() {
      // ... 기존 코드
      const [showSettings, setShowSettings] = useState(false);

      // ... 기존 useEffect들

      if (!isHydrated) {
        // ... 로딩 UI
      }

      if (!isAuthenticated) {
        return <LoginPrompt />;
      }

      return (
        <div className="min-w-[320px]">
          {/* Network Banner */}
          <NetworkBanner />

          {/* Settings Modal */}
          {showSettings && <Settings onClose={() => setShowSettings(false)} />}

          <div className="p-3 space-y-3">
            {/* Header */}
            <header className="flex items-center justify-between">
              <h1 className="text-base font-bold text-gray-900">MindHit</h1>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setShowSettings(true)}
                  className="p-1 hover:bg-gray-100 rounded-full"
                  title="설정"
                >
                  <svg className="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                    />
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                    />
                  </svg>
                </button>
                <span className="text-xs text-gray-500 truncate max-w-[100px]">{user?.email}</span>
                <button
                  onClick={logout}
                  className="text-xs text-gray-400 hover:text-gray-600"
                >
                  Logout
                </button>
              </div>
            </header>

            {/* ... 나머지 컴포넌트들 */}
          </div>
        </div>
      );
    }
    ```

- [ ] **Background에서 오프라인 복구 시 재전송**
  - [ ] `src/background/index.ts` 수정

    ```typescript
    // 기존 코드에 추가

    // 네트워크 복구 시 대기 중인 이벤트 전송
    self.addEventListener('online', async () => {
      console.log('Network restored, retrying pending events...');
      await retryPendingEvents();
    });

    async function retryPendingEvents() {
      const storage = await chrome.storage.local.get(null);
      const pendingKeys = Object.keys(storage).filter((k) =>
        k.startsWith('mindhit-pending-events-')
      );

      if (pendingKeys.length === 0) return;

      const authData = await chrome.storage.local.get('mindhit-auth');
      const rawValue = authData['mindhit-auth'];
      const parsed = typeof rawValue === 'string' ? JSON.parse(rawValue) : null;
      const token = parsed?.state?.token;

      if (!token) return;

      for (const key of pendingKeys) {
        const events = storage[key];
        if (events && events.length > 0 && state.sessionId) {
          try {
            const response = await fetch(`${API_BASE_URL}/events/batch`, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${token}`,
              },
              body: JSON.stringify({
                session_id: state.sessionId,
                events,
              }),
            });

            if (response.ok) {
              await chrome.storage.local.remove(key);
              console.log(`Retried ${events.length} pending events from ${key}`);
            }
          } catch (error) {
            console.error('Failed to retry pending events:', error);
          }
        }
      }
    }
    ```

### 검증

```bash
moonx extension:build
# Extension 리로드 후:
# 1. 네트워크 끊기 (DevTools > Network > Offline)
# 2. 오프라인 배너 표시 확인
# 3. 네트워크 복구 시 녹색 배너 표시 확인
# 4. 설정 아이콘 클릭 → 설정 페이지 열림 확인
# 5. 설정 변경 후 저장 확인
# 6. Extension 재시작 후 설정 유지 확인
```

---

## Phase 8.1 완료 확인

### 전체 검증 체크리스트

- [ ] 세션 목록 표시
- [ ] 세션 클릭 시 웹 대시보드 이동
- [ ] "웹 대시보드에서 보기" 버튼 동작
- [ ] 세션 시작 시 자동 제목 생성
- [ ] 세션 제목 인라인 수정
- [ ] 제목 변경 자동 저장 (debounce)
- [ ] 네트워크 오프라인 배너
- [ ] 네트워크 복구 배너 및 자동 동기화
- [ ] 설정 페이지 열기/닫기
- [ ] 일반 설정 (자동 시작, 탭 수집 범위)
- [ ] 고급 설정 (API URL, Web URL)
- [ ] 설정 저장 및 복원

### 테스트

```bash
moonx extension:test
moonx extension:typecheck
moonx extension:lint
```

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| 세션 목록 | `src/popup/components/SessionList.tsx` |
| 대시보드 링크 | `src/popup/components/DashboardLink.tsx` |
| 제목 입력 | `src/popup/components/SessionTitleInput.tsx` |
| 네트워크 배너 | `src/popup/components/NetworkBanner.tsx` |
| 네트워크 Hook | `src/lib/use-network-status.ts` |
| 설정 페이지 | `src/popup/components/Settings.tsx` |
| 설정 Store | `src/stores/settings-store.ts` |

---

## 다음 Phase

Phase 8.1 완료 후 [Phase 11.1: React Three Fiber 설정](./phase-11.1-threejs-setup.md)으로 진행하세요.
