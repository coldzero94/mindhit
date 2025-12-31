# Phase 11.4: 계정 및 사용량 페이지

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | 계정 정보, 구독, 토큰 사용량 페이지 구현 |
| **선행 조건** | Phase 7 완료, Phase 9 완료 |
| **예상 소요** | 2 Steps |
| **결과물** | 계정 페이지 (구독 정보 + 사용량 대시보드) |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 11.4.1 | API 연동 | ⬜ |
| 11.4.2 | 계정 페이지 UI | ⬜ |

---

## Step 11.4.1: API 연동

### 체크리스트

- [ ] **구독 API 래퍼**
  - [ ] `src/lib/api/subscription.ts`

    ```typescript
    import { apiClient } from './client';
    import type { Subscription, Plan } from '@/api/generated';

    export async function getSubscription(): Promise<Subscription> {
      const response = await apiClient.get<Subscription>('/v1/subscription');
      return response.data;
    }

    export async function getPlans(): Promise<Plan[]> {
      const response = await apiClient.get<{ plans: Plan[] }>('/v1/plans');
      return response.data.plans;
    }

    export async function updateSubscription(planId: string): Promise<Subscription> {
      const response = await apiClient.post<Subscription>('/v1/subscription', {
        plan_id: planId,
      });
      return response.data;
    }
    ```

- [ ] **사용량 API 래퍼**
  - [ ] `src/lib/api/usage.ts`

    ```typescript
    import { apiClient } from './client';

    export interface UsageData {
      tokens_used: number;
      tokens_limit: number;
      period_start: string;
      period_end: string;
      percentage: number;
    }

    export interface UsageHistoryItem {
      month: string;
      tokens_used: number;
      tokens_limit: number;
    }

    export async function getUsage(): Promise<UsageData> {
      const response = await apiClient.get<UsageData>('/v1/usage');
      return response.data;
    }

    export async function getUsageHistory(months: number = 6): Promise<UsageHistoryItem[]> {
      const response = await apiClient.get<{ history: UsageHistoryItem[] }>(
        `/v1/usage/history?months=${months}`
      );
      return response.data.history;
    }
    ```

- [ ] **구독 React Query Hook**
  - [ ] `src/lib/hooks/use-subscription.ts`

    ```typescript
    import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
    import { getSubscription, getPlans, updateSubscription } from '@/lib/api/subscription';

    export const subscriptionKeys = {
      all: ['subscription'] as const,
      plans: ['plans'] as const,
    };

    export function useSubscription() {
      return useQuery({
        queryKey: subscriptionKeys.all,
        queryFn: getSubscription,
        staleTime: 5 * 60 * 1000, // 5분
      });
    }

    export function usePlans() {
      return useQuery({
        queryKey: subscriptionKeys.plans,
        queryFn: getPlans,
        staleTime: 30 * 60 * 1000, // 30분 (플랜은 자주 변하지 않음)
      });
    }

    export function useUpdateSubscription() {
      const queryClient = useQueryClient();

      return useMutation({
        mutationFn: (planId: string) => updateSubscription(planId),
        onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: subscriptionKeys.all });
        },
      });
    }
    ```

- [ ] **사용량 React Query Hook**
  - [ ] `src/lib/hooks/use-usage.ts`

    ```typescript
    import { useQuery } from '@tanstack/react-query';
    import { getUsage, getUsageHistory } from '@/lib/api/usage';

    export const usageKeys = {
      current: ['usage', 'current'] as const,
      history: (months: number) => ['usage', 'history', months] as const,
    };

    export function useUsage() {
      return useQuery({
        queryKey: usageKeys.current,
        queryFn: getUsage,
        staleTime: 60 * 1000, // 1분
        refetchInterval: 5 * 60 * 1000, // 5분마다 자동 갱신
      });
    }

    export function useUsageHistory(months: number = 6) {
      return useQuery({
        queryKey: usageKeys.history(months),
        queryFn: () => getUsageHistory(months),
        staleTime: 5 * 60 * 1000, // 5분
      });
    }
    ```

### 검증

```bash
pnpm dev
# 개발자 도구 Network 탭에서:
# 1. /v1/subscription API 호출 확인
# 2. /v1/usage API 호출 확인
# 3. /v1/plans API 호출 확인
```

---

## Step 11.4.2: 계정 페이지 UI

### 체크리스트

- [ ] **Progress 컴포넌트 (없는 경우)**
  - [ ] `src/components/ui/Progress.tsx`

    ```tsx
    import { cn } from '@/lib/utils';

    interface ProgressProps {
      value: number;
      max?: number;
      className?: string;
      indicatorClassName?: string;
    }

    export function Progress({
      value,
      max = 100,
      className,
      indicatorClassName,
    }: ProgressProps) {
      const percentage = Math.min(Math.max((value / max) * 100, 0), 100);

      return (
        <div
          className={cn(
            'relative h-2 w-full overflow-hidden rounded-full bg-gray-200',
            className
          )}
        >
          <div
            className={cn(
              'h-full transition-all duration-500 ease-out rounded-full',
              percentage >= 90
                ? 'bg-red-500'
                : percentage >= 80
                ? 'bg-yellow-500'
                : 'bg-blue-500',
              indicatorClassName
            )}
            style={{ width: `${percentage}%` }}
          />
        </div>
      );
    }
    ```

- [ ] **구독 정보 카드 컴포넌트**
  - [ ] `src/components/account/SubscriptionCard.tsx`

    ```tsx
    'use client';

    import { Crown, Calendar, ArrowUpRight } from 'lucide-react';
    import { useSubscription, usePlans } from '@/lib/hooks/use-subscription';
    import type { Subscription, Plan } from '@/api/generated';

    interface SubscriptionCardProps {
      onUpgrade?: () => void;
    }

    export function SubscriptionCard({ onUpgrade }: SubscriptionCardProps) {
      const { data: subscription, isLoading } = useSubscription();
      const { data: plans } = usePlans();

      if (isLoading) {
        return (
          <div className="p-6 bg-white rounded-xl border border-gray-200 animate-pulse">
            <div className="h-6 bg-gray-200 rounded w-1/3 mb-4" />
            <div className="h-4 bg-gray-200 rounded w-1/2" />
          </div>
        );
      }

      if (!subscription) return null;

      const currentPlan = plans?.find((p) => p.id === subscription.plan_id);
      const isFree = currentPlan?.name?.toLowerCase() === 'free';

      const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString('ko-KR', {
          year: 'numeric',
          month: 'long',
          day: 'numeric',
        });
      };

      const statusLabels: Record<string, { label: string; color: string }> = {
        active: { label: '활성', color: 'bg-green-100 text-green-700' },
        canceled: { label: '취소됨', color: 'bg-red-100 text-red-700' },
        past_due: { label: '연체', color: 'bg-yellow-100 text-yellow-700' },
      };

      const status = statusLabels[subscription.status] || statusLabels.active;

      return (
        <div className="p-6 bg-white rounded-xl border border-gray-200">
          <div className="flex items-start justify-between mb-6">
            <div className="flex items-center gap-3">
              <div className={`p-3 rounded-xl ${isFree ? 'bg-gray-100' : 'bg-gradient-to-br from-yellow-400 to-orange-500'}`}>
                <Crown className={`w-6 h-6 ${isFree ? 'text-gray-600' : 'text-white'}`} />
              </div>
              <div>
                <h3 className="text-lg font-semibold text-gray-900">
                  {currentPlan?.name || 'Free'} 플랜
                </h3>
                <span className={`inline-block px-2 py-0.5 text-xs font-medium rounded-full ${status.color}`}>
                  {status.label}
                </span>
              </div>
            </div>

            {isFree && onUpgrade && (
              <button
                onClick={onUpgrade}
                className="flex items-center gap-1 px-4 py-2 text-sm font-medium
                           bg-gradient-to-r from-blue-600 to-purple-600 text-white
                           rounded-lg hover:from-blue-700 hover:to-purple-700
                           transition-all"
              >
                업그레이드
                <ArrowUpRight className="w-4 h-4" />
              </button>
            )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="flex items-center gap-2 text-sm text-gray-600">
              <Calendar className="w-4 h-4 text-gray-400" />
              <span>시작일: {formatDate(subscription.current_period_start)}</span>
            </div>
            <div className="flex items-center gap-2 text-sm text-gray-600">
              <Calendar className="w-4 h-4 text-gray-400" />
              <span>종료일: {formatDate(subscription.current_period_end)}</span>
            </div>
          </div>

          {currentPlan && (
            <div className="mt-4 pt-4 border-t border-gray-100">
              <p className="text-sm text-gray-500">
                월 {currentPlan.token_limit?.toLocaleString() || '무제한'} 토큰 제공
              </p>
            </div>
          )}
        </div>
      );
    }
    ```

- [ ] **사용량 카드 컴포넌트**
  - [ ] `src/components/account/UsageCard.tsx`

    ```tsx
    'use client';

    import { Zap, AlertTriangle } from 'lucide-react';
    import { useUsage } from '@/lib/hooks/use-usage';
    import { Progress } from '@/components/ui/Progress';

    export function UsageCard() {
      const { data: usage, isLoading } = useUsage();

      if (isLoading) {
        return (
          <div className="p-6 bg-white rounded-xl border border-gray-200 animate-pulse">
            <div className="h-6 bg-gray-200 rounded w-1/3 mb-4" />
            <div className="h-4 bg-gray-200 rounded w-full mb-2" />
            <div className="h-2 bg-gray-200 rounded w-full" />
          </div>
        );
      }

      if (!usage) return null;

      const isNearLimit = usage.percentage >= 80;
      const isOverLimit = usage.percentage >= 100;

      const formatNumber = (num: number) => {
        if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
        if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
        return num.toString();
      };

      return (
        <div className="p-6 bg-white rounded-xl border border-gray-200">
          <div className="flex items-start justify-between mb-4">
            <div className="flex items-center gap-3">
              <div className={`p-3 rounded-xl ${isNearLimit ? 'bg-yellow-100' : 'bg-blue-100'}`}>
                <Zap className={`w-6 h-6 ${isNearLimit ? 'text-yellow-600' : 'text-blue-600'}`} />
              </div>
              <div>
                <h3 className="text-lg font-semibold text-gray-900">토큰 사용량</h3>
                <p className="text-sm text-gray-500">이번 달</p>
              </div>
            </div>

            {isNearLimit && !isOverLimit && (
              <div className="flex items-center gap-1 px-3 py-1 bg-yellow-100 text-yellow-700 rounded-full text-sm">
                <AlertTriangle className="w-4 h-4" />
                <span>80% 도달</span>
              </div>
            )}
            {isOverLimit && (
              <div className="flex items-center gap-1 px-3 py-1 bg-red-100 text-red-700 rounded-full text-sm">
                <AlertTriangle className="w-4 h-4" />
                <span>한도 초과</span>
              </div>
            )}
          </div>

          <div className="space-y-3">
            <div className="flex items-end justify-between">
              <span className="text-3xl font-bold text-gray-900">
                {formatNumber(usage.tokens_used)}
              </span>
              <span className="text-sm text-gray-500">
                / {formatNumber(usage.tokens_limit)} 토큰
              </span>
            </div>

            <Progress value={usage.percentage} />

            <p className="text-xs text-gray-400">
              {usage.percentage.toFixed(1)}% 사용 중
            </p>
          </div>
        </div>
      );
    }
    ```

- [ ] **사용량 히스토리 컴포넌트**
  - [ ] `src/components/account/UsageHistory.tsx`

    ```tsx
    'use client';

    import { useUsageHistory } from '@/lib/hooks/use-usage';
    import { BarChart3 } from 'lucide-react';

    export function UsageHistory() {
      const { data: history, isLoading } = useUsageHistory(6);

      if (isLoading) {
        return (
          <div className="p-6 bg-white rounded-xl border border-gray-200 animate-pulse">
            <div className="h-6 bg-gray-200 rounded w-1/4 mb-4" />
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div key={i} className="h-12 bg-gray-200 rounded" />
              ))}
            </div>
          </div>
        );
      }

      if (!history || history.length === 0) {
        return (
          <div className="p-6 bg-white rounded-xl border border-gray-200">
            <div className="flex items-center gap-2 mb-4">
              <BarChart3 className="w-5 h-5 text-gray-400" />
              <h3 className="text-lg font-semibold text-gray-900">사용량 히스토리</h3>
            </div>
            <p className="text-sm text-gray-500">아직 사용 기록이 없습니다.</p>
          </div>
        );
      }

      const maxUsage = Math.max(...history.map((h) => h.tokens_used));

      const formatMonth = (monthStr: string) => {
        const date = new Date(monthStr + '-01');
        return date.toLocaleDateString('ko-KR', { year: 'numeric', month: 'short' });
      };

      const formatNumber = (num: number) => {
        if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
        if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
        return num.toString();
      };

      return (
        <div className="p-6 bg-white rounded-xl border border-gray-200">
          <div className="flex items-center gap-2 mb-6">
            <BarChart3 className="w-5 h-5 text-gray-400" />
            <h3 className="text-lg font-semibold text-gray-900">사용량 히스토리</h3>
          </div>

          <div className="space-y-4">
            {history.map((item) => {
              const percentage = maxUsage > 0 ? (item.tokens_used / maxUsage) * 100 : 0;
              const usagePercentage = (item.tokens_used / item.tokens_limit) * 100;

              return (
                <div key={item.month} className="space-y-1">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-600">{formatMonth(item.month)}</span>
                    <span className="text-gray-900 font-medium">
                      {formatNumber(item.tokens_used)} / {formatNumber(item.tokens_limit)}
                    </span>
                  </div>
                  <div className="relative h-6 bg-gray-100 rounded-lg overflow-hidden">
                    <div
                      className={`absolute left-0 top-0 h-full rounded-lg transition-all duration-500 ${
                        usagePercentage >= 90
                          ? 'bg-red-500'
                          : usagePercentage >= 80
                          ? 'bg-yellow-500'
                          : 'bg-blue-500'
                      }`}
                      style={{ width: `${percentage}%` }}
                    />
                    <span className="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-500">
                      {usagePercentage.toFixed(0)}%
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      );
    }
    ```

- [ ] **계정 페이지**
  - [ ] `src/app/(dashboard)/account/page.tsx`

    ```tsx
    'use client';

    import { useState } from 'react';
    import { User, Settings } from 'lucide-react';
    import { useAuth } from '@/lib/hooks/use-auth';
    import { SubscriptionCard } from '@/components/account/SubscriptionCard';
    import { UsageCard } from '@/components/account/UsageCard';
    import { UsageHistory } from '@/components/account/UsageHistory';

    export default function AccountPage() {
      const { user } = useAuth();
      const [showUpgradeModal, setShowUpgradeModal] = useState(false);

      return (
        <div className="space-y-6">
          {/* Header */}
          <div>
            <h1 className="text-2xl font-bold text-gray-900">계정</h1>
            <p className="text-gray-500 mt-1">계정 정보와 사용량을 확인하세요</p>
          </div>

          {/* User Info */}
          <div className="p-6 bg-white rounded-xl border border-gray-200">
            <div className="flex items-center gap-4">
              <div className="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600
                              rounded-full flex items-center justify-center">
                <User className="w-8 h-8 text-white" />
              </div>
              <div className="flex-1">
                <h2 className="text-xl font-semibold text-gray-900">
                  {user?.name || user?.email?.split('@')[0] || '사용자'}
                </h2>
                <p className="text-gray-500">{user?.email}</p>
              </div>
              <button className="flex items-center gap-2 px-4 py-2 text-gray-600
                                 hover:bg-gray-100 rounded-lg transition-colors">
                <Settings className="w-4 h-4" />
                <span className="text-sm">설정</span>
              </button>
            </div>
          </div>

          {/* Subscription & Usage */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <SubscriptionCard onUpgrade={() => setShowUpgradeModal(true)} />
            <UsageCard />
          </div>

          {/* Usage History */}
          <UsageHistory />

          {/* Upgrade Modal (placeholder) */}
          {showUpgradeModal && (
            <UpgradeModal onClose={() => setShowUpgradeModal(false)} />
          )}
        </div>
      );
    }

    // 간단한 업그레이드 모달 (Phase 14에서 Stripe 연동)
    function UpgradeModal({ onClose }: { onClose: () => void }) {
      return (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 max-w-md w-full mx-4">
            <h2 className="text-xl font-bold text-gray-900 mb-4">플랜 업그레이드</h2>
            <p className="text-gray-600 mb-6">
              플랜 업그레이드 기능은 준비 중입니다.
              더 많은 토큰과 기능을 원하시면 문의해 주세요.
            </p>
            <div className="flex justify-end">
              <button
                onClick={onClose}
                className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg
                           hover:bg-gray-200 transition-colors"
              >
                닫기
              </button>
            </div>
          </div>
        </div>
      );
    }
    ```

- [ ] **사이드바 업데이트**
  - [ ] `src/components/layout/Sidebar.tsx` (계정 메뉴 추가)

    ```tsx
    // 기존 Sidebar.tsx에 메뉴 항목 추가

    import { User } from 'lucide-react';

    // navigation 배열에 추가
    const navigation = [
      // ... 기존 메뉴들
      { name: '계정', href: '/account', icon: User },
    ];
    ```

- [ ] **헤더 사용량 배지 (선택사항)**
  - [ ] `src/components/layout/HeaderUsageBadge.tsx`

    ```tsx
    'use client';

    import { AlertTriangle, Zap } from 'lucide-react';
    import Link from 'next/link';
    import { useUsage } from '@/lib/hooks/use-usage';

    export function HeaderUsageBadge() {
      const { data: usage } = useUsage();

      if (!usage || usage.percentage < 80) return null;

      const isOver = usage.percentage >= 100;

      return (
        <Link
          href="/account"
          className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium transition-colors ${
            isOver
              ? 'bg-red-100 text-red-700 hover:bg-red-200'
              : 'bg-yellow-100 text-yellow-700 hover:bg-yellow-200'
          }`}
        >
          {isOver ? (
            <AlertTriangle className="w-3.5 h-3.5" />
          ) : (
            <Zap className="w-3.5 h-3.5" />
          )}
          <span>
            {isOver ? '토큰 한도 초과' : `${usage.percentage.toFixed(0)}% 사용`}
          </span>
        </Link>
      );
    }
    ```

    ```tsx
    // Header.tsx에 추가
    import { HeaderUsageBadge } from './HeaderUsageBadge';

    // Header 컴포넌트 내부
    <div className="flex items-center gap-4">
      <HeaderUsageBadge />
      {/* ... 기존 요소들 */}
    </div>
    ```

### 검증

```bash
pnpm dev
# http://localhost:3000/account 접속
# 1. 사용자 정보 표시 확인
# 2. 구독 정보 카드 확인
# 3. 사용량 Progress 바 확인
# 4. 사용량 히스토리 확인
# 5. 사이드바에서 계정 메뉴 클릭
# 6. 80% 이상 사용 시 헤더 배지 확인
```

---

## Phase 11.4 완료 확인

### 전체 검증 체크리스트

- [ ] 구독 API 연동
- [ ] 사용량 API 연동
- [ ] 계정 페이지 렌더링
- [ ] 구독 정보 표시
- [ ] 사용량 Progress 바
- [ ] 사용량 히스토리
- [ ] 사이드바 메뉴
- [ ] 헤더 사용량 배지 (80% 이상)

### 테스트

```bash
moonx web:typecheck
moonx web:lint
moonx web:build
```

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| 구독 API | `src/lib/api/subscription.ts` |
| 사용량 API | `src/lib/api/usage.ts` |
| 구독 Hook | `src/lib/hooks/use-subscription.ts` |
| 사용량 Hook | `src/lib/hooks/use-usage.ts` |
| Progress 컴포넌트 | `src/components/ui/Progress.tsx` |
| 구독 카드 | `src/components/account/SubscriptionCard.tsx` |
| 사용량 카드 | `src/components/account/UsageCard.tsx` |
| 사용량 히스토리 | `src/components/account/UsageHistory.tsx` |
| 계정 페이지 | `src/app/(dashboard)/account/page.tsx` |
| 헤더 배지 | `src/components/layout/HeaderUsageBadge.tsx` |

---

## 다음 Phase

Phase 11.4 완료 후 [Phase 11.5: 애니메이션 및 인터랙션](./phase-11.5-animation.md)으로 진행하세요.
