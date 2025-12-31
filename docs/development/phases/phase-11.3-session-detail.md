# Phase 11.3: 세션 상세 페이지 개선

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | 세션 상세 페이지에 마인드맵 뷰어 및 타임라인 추가 |
| **선행 조건** | Phase 11.2 완료, Phase 10.2 완료 |
| **예상 소요** | 2 Steps |
| **결과물** | 마인드맵 + 타임라인 탭이 있는 세션 상세 페이지 |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 11.3.1 | 마인드맵 API 연동 | ⬜ |
| 11.3.2 | 탭 UI 및 세션 수정 | ⬜ |

---

## 현재 상태 분석

### 구현 완료 (기존)
- [x] 세션 기본 정보 표시
- [x] 방문한 페이지 목록
- [x] 하이라이트 목록
- [x] 통계 카드 (페이지 수, 하이라이트 수 등)
- [x] 세션 삭제

### 미구현 (이번 Phase에서 추가)
- [ ] 마인드맵 API 연동
- [ ] 마인드맵 탭
- [ ] 타임라인 뷰 개선
- [ ] 노드 상세 패널
- [ ] 세션 제목 수정

---

## Step 11.3.1: 마인드맵 API 연동

### 체크리스트

- [ ] **마인드맵 API 래퍼**
  - [ ] `src/lib/api/mindmap.ts`

    ```typescript
    import { apiClient } from './client';
    import type { MindmapResponse } from '@/api/generated';

    export async function getMindmap(sessionId: string): Promise<MindmapResponse> {
      const response = await apiClient.get<MindmapResponse>(
        `/v1/sessions/${sessionId}/mindmap`
      );
      return response.data;
    }

    export async function generateMindmap(sessionId: string): Promise<MindmapResponse> {
      const response = await apiClient.post<MindmapResponse>(
        `/v1/sessions/${sessionId}/mindmap/generate`
      );
      return response.data;
    }
    ```

- [ ] **React Query Hook**
  - [ ] `src/lib/hooks/use-mindmap.ts`

    ```typescript
    import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
    import { getMindmap, generateMindmap } from '@/lib/api/mindmap';
    import type { MindmapResponse } from '@/api/generated';

    export const mindmapKeys = {
      all: ['mindmaps'] as const,
      detail: (sessionId: string) => [...mindmapKeys.all, sessionId] as const,
    };

    export function useMindmap(sessionId: string) {
      return useQuery({
        queryKey: mindmapKeys.detail(sessionId),
        queryFn: () => getMindmap(sessionId),
        staleTime: 5 * 60 * 1000, // 5분
        retry: 1,
      });
    }

    export function useGenerateMindmap(sessionId: string) {
      const queryClient = useQueryClient();

      return useMutation({
        mutationFn: () => generateMindmap(sessionId),
        onSuccess: (data) => {
          queryClient.setQueryData<MindmapResponse>(
            mindmapKeys.detail(sessionId),
            data
          );
        },
      });
    }
    ```

- [ ] **마인드맵 데이터 변환 유틸**
  - [ ] `src/lib/utils/mindmap-transform.ts`

    ```typescript
    import type { MindmapResponse } from '@/api/generated';
    import type { MindmapData, MindmapNode, MindmapEdge } from '@/types/mindmap';

    export function transformMindmapResponse(response: MindmapResponse): MindmapData | null {
      if (!response.mindmap) {
        return null;
      }

      const { nodes, edges, layout } = response.mindmap;

      return {
        nodes: nodes.map((node): MindmapNode => ({
          id: node.id,
          label: node.label,
          type: node.type as MindmapNode['type'],
          size: node.size,
          color: node.color,
          position: node.position,
          data: {
            description: node.data?.description,
            urls: node.data?.urls,
            visitCount: node.data?.visit_count,
            totalDuration: node.data?.total_duration,
          },
        })),
        edges: edges.map((edge): MindmapEdge => ({
          source: edge.source,
          target: edge.target,
          weight: edge.weight,
        })),
        layout: {
          type: layout.type as MindmapData['layout']['type'],
          params: layout.params || {},
        },
      };
    }
    ```

- [ ] **노드 상세 패널**
  - [ ] `src/components/mindmap/NodeDetailPanel.tsx`

    ```tsx
    'use client';

    import { useEffect, useState } from 'react';
    import { X, ExternalLink, Clock, Eye, FileText } from 'lucide-react';
    import type { MindmapNode } from '@/types/mindmap';
    import { cn } from '@/lib/utils';

    interface NodeDetailPanelProps {
      node: MindmapNode | null;
      onClose: () => void;
    }

    export function NodeDetailPanel({ node, onClose }: NodeDetailPanelProps) {
      const [isVisible, setIsVisible] = useState(false);

      useEffect(() => {
        if (node) {
          // 약간의 딜레이 후 슬라이드 인
          const timer = setTimeout(() => setIsVisible(true), 50);
          return () => clearTimeout(timer);
        } else {
          setIsVisible(false);
        }
      }, [node]);

      if (!node) return null;

      const formatDuration = (seconds?: number) => {
        if (!seconds) return '-';
        if (seconds < 60) return `${seconds}초`;
        const minutes = Math.floor(seconds / 60);
        const secs = seconds % 60;
        return `${minutes}분 ${secs}초`;
      };

      const nodeTypeLabels: Record<MindmapNode['type'], string> = {
        core: '핵심 주제',
        topic: '주제',
        subtopic: '하위 주제',
        page: '페이지',
      };

      return (
        <div
          className={cn(
            'absolute top-4 right-4 w-80 bg-gray-900/95 backdrop-blur-sm',
            'rounded-xl border border-gray-700 shadow-2xl',
            'transform transition-all duration-300 ease-out z-10',
            isVisible ? 'translate-x-0 opacity-100' : 'translate-x-4 opacity-0'
          )}
        >
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b border-gray-700">
            <div className="flex items-center gap-2">
              <div
                className="w-3 h-3 rounded-full"
                style={{ backgroundColor: node.color }}
              />
              <span className="text-xs text-gray-400">
                {nodeTypeLabels[node.type]}
              </span>
            </div>
            <button
              onClick={onClose}
              className="p-1 hover:bg-gray-800 rounded-lg transition-colors"
            >
              <X className="w-4 h-4 text-gray-400" />
            </button>
          </div>

          {/* Content */}
          <div className="p-4 space-y-4">
            {/* Label */}
            <h3 className="text-lg font-semibold text-white">{node.label}</h3>

            {/* Description */}
            {node.data.description && (
              <p className="text-sm text-gray-400">{node.data.description}</p>
            )}

            {/* Stats */}
            <div className="grid grid-cols-2 gap-3">
              {node.data.visitCount !== undefined && (
                <div className="flex items-center gap-2 text-sm">
                  <Eye className="w-4 h-4 text-gray-500" />
                  <span className="text-gray-300">
                    {node.data.visitCount}회 방문
                  </span>
                </div>
              )}
              {node.data.totalDuration !== undefined && (
                <div className="flex items-center gap-2 text-sm">
                  <Clock className="w-4 h-4 text-gray-500" />
                  <span className="text-gray-300">
                    {formatDuration(node.data.totalDuration)}
                  </span>
                </div>
              )}
            </div>

            {/* URLs */}
            {node.data.urls && node.data.urls.length > 0 && (
              <div className="space-y-2">
                <div className="flex items-center gap-2 text-sm text-gray-400">
                  <FileText className="w-4 h-4" />
                  <span>관련 페이지 ({node.data.urls.length})</span>
                </div>
                <div className="max-h-32 overflow-y-auto space-y-1">
                  {node.data.urls.map((url, index) => {
                    const hostname = new URL(url).hostname;
                    return (
                      <a
                        key={index}
                        href={url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center gap-2 p-2 text-xs text-gray-300
                                   bg-gray-800/50 rounded-lg hover:bg-gray-800
                                   transition-colors group"
                      >
                        <span className="truncate flex-1">{hostname}</span>
                        <ExternalLink className="w-3 h-3 text-gray-500
                                                  group-hover:text-gray-300" />
                      </a>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        </div>
      );
    }
    ```

- [ ] **마인드맵 뷰어 컴포넌트**
  - [ ] `src/components/mindmap/MindmapViewer.tsx`

    ```tsx
    'use client';

    import { useState, useCallback } from 'react';
    import { MindmapCanvas } from './MindmapCanvas';
    import { Galaxy } from './Galaxy';
    import { NodeDetailPanel } from './NodeDetailPanel';
    import { useMindmap, useGenerateMindmap } from '@/lib/hooks/use-mindmap';
    import { transformMindmapResponse } from '@/lib/utils/mindmap-transform';
    import type { MindmapNode } from '@/types/mindmap';
    import { Loader2, Sparkles, AlertCircle } from 'lucide-react';

    interface MindmapViewerProps {
      sessionId: string;
    }

    export function MindmapViewer({ sessionId }: MindmapViewerProps) {
      const [selectedNode, setSelectedNode] = useState<MindmapNode | null>(null);

      const { data, isLoading, error } = useMindmap(sessionId);
      const generateMutation = useGenerateMindmap(sessionId);

      const mindmapData = data ? transformMindmapResponse(data) : null;

      const handleNodeClick = useCallback((node: MindmapNode) => {
        setSelectedNode(node);
      }, []);

      const handleClosePanel = useCallback(() => {
        setSelectedNode(null);
      }, []);

      const handleGenerate = () => {
        generateMutation.mutate();
      };

      // 로딩 상태
      if (isLoading) {
        return (
          <div className="flex items-center justify-center h-[500px] bg-gray-900 rounded-xl">
            <div className="flex flex-col items-center gap-3">
              <Loader2 className="w-8 h-8 text-blue-500 animate-spin" />
              <p className="text-gray-400">마인드맵 로딩 중...</p>
            </div>
          </div>
        );
      }

      // 에러 상태
      if (error) {
        return (
          <div className="flex items-center justify-center h-[500px] bg-gray-900 rounded-xl">
            <div className="flex flex-col items-center gap-3 text-center">
              <AlertCircle className="w-8 h-8 text-red-500" />
              <p className="text-gray-400">마인드맵을 불러올 수 없습니다</p>
              <button
                onClick={handleGenerate}
                disabled={generateMutation.isPending}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg
                           hover:bg-blue-700 transition-colors disabled:opacity-50"
              >
                {generateMutation.isPending ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  '다시 시도'
                )}
              </button>
            </div>
          </div>
        );
      }

      // 마인드맵 없음 - 생성 필요
      if (!mindmapData) {
        return (
          <div className="flex items-center justify-center h-[500px] bg-gray-900 rounded-xl">
            <div className="flex flex-col items-center gap-4 text-center">
              <Sparkles className="w-12 h-12 text-gray-600" />
              <div>
                <h3 className="text-lg font-medium text-white mb-1">
                  마인드맵이 없습니다
                </h3>
                <p className="text-sm text-gray-400">
                  AI가 브라우징 데이터를 분석하여 마인드맵을 생성합니다
                </p>
              </div>
              <button
                onClick={handleGenerate}
                disabled={generateMutation.isPending}
                className="flex items-center gap-2 px-6 py-3 bg-gradient-to-r
                           from-blue-600 to-purple-600 text-white rounded-xl
                           hover:from-blue-700 hover:to-purple-700
                           transition-all disabled:opacity-50"
              >
                {generateMutation.isPending ? (
                  <>
                    <Loader2 className="w-5 h-5 animate-spin" />
                    생성 중...
                  </>
                ) : (
                  <>
                    <Sparkles className="w-5 h-5" />
                    마인드맵 생성
                  </>
                )}
              </button>
            </div>
          </div>
        );
      }

      // 마인드맵 렌더링
      return (
        <div className="relative">
          <MindmapCanvas className="h-[500px]">
            <Galaxy
              nodes={mindmapData.nodes}
              edges={mindmapData.edges}
              onNodeClick={handleNodeClick}
              selectedNodeId={selectedNode?.id}
            />
          </MindmapCanvas>

          <NodeDetailPanel node={selectedNode} onClose={handleClosePanel} />
        </div>
      );
    }
    ```

### 검증

```bash
pnpm dev
# 세션 상세 페이지에서:
# 1. 마인드맵 데이터 로드 확인 (Network 탭)
# 2. 마인드맵 없는 경우 생성 버튼 확인
# 3. 노드 클릭 시 상세 패널 표시 확인
```

---

## Step 11.3.2: 탭 UI 및 세션 수정

### 체크리스트

- [ ] **Tabs 컴포넌트 (없는 경우)**
  - [ ] `src/components/ui/Tabs.tsx`

    ```tsx
    'use client';

    import { createContext, useContext, useState, ReactNode } from 'react';
    import { cn } from '@/lib/utils';

    interface TabsContextValue {
      activeTab: string;
      setActiveTab: (tab: string) => void;
    }

    const TabsContext = createContext<TabsContextValue | null>(null);

    function useTabsContext() {
      const context = useContext(TabsContext);
      if (!context) {
        throw new Error('Tabs components must be used within a Tabs provider');
      }
      return context;
    }

    interface TabsProps {
      defaultValue: string;
      children: ReactNode;
      className?: string;
    }

    export function Tabs({ defaultValue, children, className }: TabsProps) {
      const [activeTab, setActiveTab] = useState(defaultValue);

      return (
        <TabsContext.Provider value={{ activeTab, setActiveTab }}>
          <div className={className}>{children}</div>
        </TabsContext.Provider>
      );
    }

    interface TabsListProps {
      children: ReactNode;
      className?: string;
    }

    export function TabsList({ children, className }: TabsListProps) {
      return (
        <div
          className={cn(
            'inline-flex items-center gap-1 p-1 bg-gray-100 rounded-lg',
            className
          )}
        >
          {children}
        </div>
      );
    }

    interface TabsTriggerProps {
      value: string;
      children: ReactNode;
      className?: string;
    }

    export function TabsTrigger({ value, children, className }: TabsTriggerProps) {
      const { activeTab, setActiveTab } = useTabsContext();
      const isActive = activeTab === value;

      return (
        <button
          onClick={() => setActiveTab(value)}
          className={cn(
            'px-4 py-2 text-sm font-medium rounded-md transition-all',
            isActive
              ? 'bg-white text-gray-900 shadow-sm'
              : 'text-gray-600 hover:text-gray-900',
            className
          )}
        >
          {children}
        </button>
      );
    }

    interface TabsContentProps {
      value: string;
      children: ReactNode;
      className?: string;
    }

    export function TabsContent({ value, children, className }: TabsContentProps) {
      const { activeTab } = useTabsContext();

      if (activeTab !== value) return null;

      return <div className={className}>{children}</div>;
    }
    ```

- [ ] **타임라인 컴포넌트**
  - [ ] `src/components/sessions/SessionTimeline.tsx`

    ```tsx
    'use client';

    import { ExternalLink, Clock, ScrollText } from 'lucide-react';
    import type { Event } from '@/api/generated';

    interface SessionTimelineProps {
      events: Event[];
    }

    export function SessionTimeline({ events }: SessionTimelineProps) {
      // page_visit 이벤트만 필터링하고 시간순 정렬
      const pageVisits = events
        .filter((e) => e.event_type === 'page_visit')
        .sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());

      if (pageVisits.length === 0) {
        return (
          <div className="flex flex-col items-center justify-center py-12 text-gray-400">
            <ScrollText className="w-12 h-12 mb-3" />
            <p>방문 기록이 없습니다</p>
          </div>
        );
      }

      const formatTime = (timestamp: string) => {
        return new Date(timestamp).toLocaleTimeString('ko-KR', {
          hour: '2-digit',
          minute: '2-digit',
        });
      };

      const formatDuration = (seconds?: number) => {
        if (!seconds) return '-';
        if (seconds < 60) return `${seconds}초`;
        const minutes = Math.floor(seconds / 60);
        return `${minutes}분`;
      };

      return (
        <div className="relative">
          {/* Timeline line */}
          <div className="absolute left-6 top-0 bottom-0 w-px bg-gray-200" />

          <div className="space-y-4">
            {pageVisits.map((event, index) => {
              const data = event.data as {
                url?: string;
                title?: string;
                duration?: number;
                scroll_depth?: number;
              };

              return (
                <div key={event.id} className="relative flex gap-4">
                  {/* Timeline dot */}
                  <div className="relative z-10 flex items-center justify-center w-12">
                    <div
                      className={`w-3 h-3 rounded-full ${
                        index === 0 ? 'bg-blue-500' : 'bg-gray-300'
                      }`}
                    />
                  </div>

                  {/* Content card */}
                  <div className="flex-1 p-4 bg-white rounded-lg border border-gray-200
                                  hover:border-gray-300 transition-colors">
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex-1 min-w-0">
                        <h4 className="font-medium text-gray-900 truncate">
                          {data.title || '제목 없음'}
                        </h4>
                        <p className="text-sm text-gray-500 truncate mt-0.5">
                          {data.url}
                        </p>
                      </div>

                      <a
                        href={data.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="p-2 text-gray-400 hover:text-gray-600
                                   hover:bg-gray-100 rounded-lg transition-colors"
                      >
                        <ExternalLink className="w-4 h-4" />
                      </a>
                    </div>

                    <div className="flex items-center gap-4 mt-3 text-xs text-gray-500">
                      <span className="flex items-center gap-1">
                        <Clock className="w-3 h-3" />
                        {formatTime(event.timestamp)}
                      </span>
                      {data.duration && (
                        <span>체류 {formatDuration(data.duration)}</span>
                      )}
                      {data.scroll_depth !== undefined && (
                        <span>스크롤 {data.scroll_depth}%</span>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      );
    }
    ```

- [ ] **세션 통계 컴포넌트**
  - [ ] `src/components/sessions/SessionStats.tsx`

    ```tsx
    'use client';

    import { Clock, FileText, Highlighter, BarChart3 } from 'lucide-react';
    import type { Session, Event } from '@/api/generated';

    interface SessionStatsProps {
      session: Session;
      events: Event[];
    }

    export function SessionStats({ session, events }: SessionStatsProps) {
      // 통계 계산
      const pageVisits = events.filter((e) => e.event_type === 'page_visit');
      const highlights = events.filter((e) => e.event_type === 'highlight');

      // 총 체류 시간 계산
      const totalDuration = pageVisits.reduce((sum, event) => {
        const data = event.data as { duration?: number };
        return sum + (data.duration || 0);
      }, 0);

      // 평균 스크롤 깊이
      const avgScrollDepth = pageVisits.reduce((sum, event) => {
        const data = event.data as { scroll_depth?: number };
        return sum + (data.scroll_depth || 0);
      }, 0) / (pageVisits.length || 1);

      const formatDuration = (seconds: number) => {
        if (seconds < 60) return `${seconds}초`;
        const minutes = Math.floor(seconds / 60);
        const secs = seconds % 60;
        if (minutes < 60) return `${minutes}분 ${secs}초`;
        const hours = Math.floor(minutes / 60);
        const mins = minutes % 60;
        return `${hours}시간 ${mins}분`;
      };

      const stats = [
        {
          label: '총 체류 시간',
          value: formatDuration(totalDuration),
          icon: Clock,
          color: 'bg-blue-100 text-blue-600',
        },
        {
          label: '방문 페이지',
          value: `${pageVisits.length}개`,
          icon: FileText,
          color: 'bg-green-100 text-green-600',
        },
        {
          label: '하이라이트',
          value: `${highlights.length}개`,
          icon: Highlighter,
          color: 'bg-yellow-100 text-yellow-600',
        },
        {
          label: '평균 스크롤',
          value: `${Math.round(avgScrollDepth)}%`,
          icon: BarChart3,
          color: 'bg-purple-100 text-purple-600',
        },
      ];

      return (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {stats.map((stat) => (
            <div
              key={stat.label}
              className="p-4 bg-white rounded-xl border border-gray-200"
            >
              <div className="flex items-center gap-3">
                <div className={`p-2 rounded-lg ${stat.color}`}>
                  <stat.icon className="w-5 h-5" />
                </div>
                <div>
                  <p className="text-xs text-gray-500">{stat.label}</p>
                  <p className="text-lg font-semibold text-gray-900">
                    {stat.value}
                  </p>
                </div>
              </div>
            </div>
          ))}
        </div>
      );
    }
    ```

- [ ] **세션 제목 수정 컴포넌트**
  - [ ] `src/components/sessions/SessionTitleEdit.tsx`

    ```tsx
    'use client';

    import { useState, useRef, useEffect } from 'react';
    import { Check, X, Pencil } from 'lucide-react';

    interface SessionTitleEditProps {
      title: string;
      onSave: (newTitle: string) => Promise<void>;
      className?: string;
    }

    export function SessionTitleEdit({
      title,
      onSave,
      className,
    }: SessionTitleEditProps) {
      const [isEditing, setIsEditing] = useState(false);
      const [editValue, setEditValue] = useState(title);
      const [isSaving, setIsSaving] = useState(false);
      const inputRef = useRef<HTMLInputElement>(null);

      useEffect(() => {
        if (isEditing && inputRef.current) {
          inputRef.current.focus();
          inputRef.current.select();
        }
      }, [isEditing]);

      const handleSave = async () => {
        if (editValue.trim() === title || editValue.trim() === '') {
          setIsEditing(false);
          setEditValue(title);
          return;
        }

        setIsSaving(true);
        try {
          await onSave(editValue.trim());
          setIsEditing(false);
        } catch {
          setEditValue(title);
        } finally {
          setIsSaving(false);
        }
      };

      const handleCancel = () => {
        setIsEditing(false);
        setEditValue(title);
      };

      const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') {
          handleSave();
        } else if (e.key === 'Escape') {
          handleCancel();
        }
      };

      if (isEditing) {
        return (
          <div className={`flex items-center gap-2 ${className}`}>
            <input
              ref={inputRef}
              type="text"
              value={editValue}
              onChange={(e) => setEditValue(e.target.value)}
              onKeyDown={handleKeyDown}
              className="flex-1 px-3 py-2 text-2xl font-bold border border-gray-300
                         rounded-lg focus:outline-none focus:ring-2
                         focus:ring-blue-500 focus:border-transparent"
              disabled={isSaving}
            />
            <button
              onClick={handleSave}
              disabled={isSaving}
              className="p-2 text-green-600 hover:bg-green-100 rounded-lg
                         transition-colors disabled:opacity-50"
            >
              <Check className="w-5 h-5" />
            </button>
            <button
              onClick={handleCancel}
              disabled={isSaving}
              className="p-2 text-gray-400 hover:bg-gray-100 rounded-lg
                         transition-colors disabled:opacity-50"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        );
      }

      return (
        <button
          onClick={() => setIsEditing(true)}
          className={`group flex items-center gap-2 text-left ${className}`}
        >
          <h1 className="text-2xl font-bold text-gray-900">{title}</h1>
          <Pencil
            className="w-4 h-4 text-gray-400 opacity-0 group-hover:opacity-100
                       transition-opacity"
          />
        </button>
      );
    }
    ```

- [ ] **세션 업데이트 mutation 추가**
  - [ ] `src/lib/hooks/use-sessions.ts` 업데이트

    ```typescript
    // 기존 imports 추가
    import { useMutation, useQueryClient } from '@tanstack/react-query';

    // updateSession API 함수 추가 (src/lib/api/sessions.ts에)
    export async function updateSession(
      sessionId: string,
      data: { title?: string }
    ): Promise<Session> {
      const response = await apiClient.patch<Session>(
        `/v1/sessions/${sessionId}`,
        data
      );
      return response.data;
    }

    // Hook 추가 (src/lib/hooks/use-sessions.ts에)
    export function useUpdateSession(sessionId: string) {
      const queryClient = useQueryClient();

      return useMutation({
        mutationFn: (data: { title?: string }) => updateSession(sessionId, data),
        onSuccess: (updatedSession) => {
          // 세션 상세 캐시 업데이트
          queryClient.setQueryData(['sessions', sessionId], updatedSession);
          // 세션 목록 캐시 무효화
          queryClient.invalidateQueries({ queryKey: ['sessions'] });
        },
      });
    }
    ```

- [ ] **세션 상세 페이지 업데이트**
  - [ ] `src/app/(dashboard)/sessions/[id]/page.tsx`

    ```tsx
    'use client';

    import { useParams } from 'next/navigation';
    import { Brain, List } from 'lucide-react';
    import { useSession, useUpdateSession } from '@/lib/hooks/use-sessions';
    import { useEvents } from '@/lib/hooks/use-events';
    import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/Tabs';
    import { SessionTitleEdit } from '@/components/sessions/SessionTitleEdit';
    import { SessionStats } from '@/components/sessions/SessionStats';
    import { SessionTimeline } from '@/components/sessions/SessionTimeline';
    import { MindmapViewer } from '@/components/mindmap/MindmapViewer';
    import { toast } from 'sonner';

    export default function SessionDetailPage() {
      const params = useParams();
      const sessionId = params.id as string;

      const { data: session, isLoading: sessionLoading } = useSession(sessionId);
      const { data: eventsData, isLoading: eventsLoading } = useEvents(sessionId);
      const updateMutation = useUpdateSession(sessionId);

      const handleTitleSave = async (newTitle: string) => {
        try {
          await updateMutation.mutateAsync({ title: newTitle });
          toast.success('세션 제목이 수정되었습니다');
        } catch {
          toast.error('제목 수정에 실패했습니다');
          throw new Error('Failed to update title');
        }
      };

      if (sessionLoading || eventsLoading) {
        return (
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
          </div>
        );
      }

      if (!session) {
        return (
          <div className="text-center py-12">
            <p className="text-gray-500">세션을 찾을 수 없습니다</p>
          </div>
        );
      }

      const events = eventsData?.events || [];

      return (
        <div className="space-y-6">
          {/* Header with editable title */}
          <div className="flex items-start justify-between">
            <SessionTitleEdit
              title={session.title || `세션 #${session.id.slice(0, 8)}`}
              onSave={handleTitleSave}
            />
            <span
              className={`px-3 py-1 rounded-full text-sm font-medium ${
                session.status === 'active'
                  ? 'bg-green-100 text-green-700'
                  : session.status === 'paused'
                  ? 'bg-yellow-100 text-yellow-700'
                  : 'bg-gray-100 text-gray-700'
              }`}
            >
              {session.status === 'active' ? '진행 중' :
               session.status === 'paused' ? '일시정지' : '완료'}
            </span>
          </div>

          {/* Stats */}
          <SessionStats session={session} events={events} />

          {/* Tabs */}
          <Tabs defaultValue="mindmap" className="space-y-4">
            <TabsList>
              <TabsTrigger value="mindmap">
                <Brain className="w-4 h-4 mr-2" />
                마인드맵
              </TabsTrigger>
              <TabsTrigger value="timeline">
                <List className="w-4 h-4 mr-2" />
                타임라인
              </TabsTrigger>
            </TabsList>

            <TabsContent value="mindmap">
              <MindmapViewer sessionId={sessionId} />
            </TabsContent>

            <TabsContent value="timeline">
              <div className="bg-gray-50 rounded-xl p-6">
                <SessionTimeline events={events} />
              </div>
            </TabsContent>
          </Tabs>
        </div>
      );
    }
    ```

### 검증

```bash
pnpm dev
# 세션 상세 페이지에서:
# 1. 마인드맵 / 타임라인 탭 전환 확인
# 2. 마인드맵 렌더링 확인
# 3. 노드 클릭 → 상세 패널 확인
# 4. 타임라인 뷰 확인
# 5. 세션 제목 클릭 → 수정 확인
```

---

## Phase 11.3 완료 확인

### 전체 검증 체크리스트

- [ ] 마인드맵 API 연동
- [ ] 마인드맵 탭 렌더링
- [ ] 노드 클릭 → 상세 패널
- [ ] 타임라인 탭 렌더링
- [ ] 세션 제목 수정
- [ ] 탭 전환 동작

### 테스트

```bash
moonx web:typecheck
moonx web:lint
moonx web:build
```

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| 마인드맵 API | `src/lib/api/mindmap.ts` |
| 마인드맵 Hook | `src/lib/hooks/use-mindmap.ts` |
| 데이터 변환 | `src/lib/utils/mindmap-transform.ts` |
| 노드 상세 패널 | `src/components/mindmap/NodeDetailPanel.tsx` |
| 마인드맵 뷰어 | `src/components/mindmap/MindmapViewer.tsx` |
| Tabs 컴포넌트 | `src/components/ui/Tabs.tsx` |
| 타임라인 | `src/components/sessions/SessionTimeline.tsx` |
| 세션 통계 | `src/components/sessions/SessionStats.tsx` |
| 제목 수정 | `src/components/sessions/SessionTitleEdit.tsx` |

---

## 다음 Phase

Phase 11.3 완료 후 [Phase 11.4: 계정 및 사용량 페이지](./phase-11.4-account-usage.md)로 진행하세요.
