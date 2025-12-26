"use client";

import { useSessions } from "@/lib/hooks/use-sessions";
import { SessionCard } from "./session-card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { ChevronLeft, ChevronRight } from "lucide-react";

interface SessionListProps {
  page: number;
  onPageChange: (page: number) => void;
  perPage?: number;
}

export function SessionList({ page, onPageChange, perPage = 12 }: SessionListProps) {
  const offset = (page - 1) * perPage;
  const { data, isLoading, error } = useSessions(perPage, offset);

  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-32 rounded-lg" />
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-red-500">세션을 불러오는데 실패했습니다.</p>
        <Button
          variant="outline"
          className="mt-4"
          onClick={() => window.location.reload()}
        >
          다시 시도
        </Button>
      </div>
    );
  }

  if (!data?.sessions.length) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-400 mb-4">
          <svg
            className="mx-auto h-12 w-12"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1}
              d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
            />
          </svg>
        </div>
        <p className="text-gray-500 text-lg">아직 녹화된 세션이 없습니다.</p>
        <p className="text-sm text-gray-400 mt-2">
          Chrome Extension을 사용하여 첫 번째 세션을 녹화해보세요.
        </p>
      </div>
    );
  }

  // Calculate total pages (simple heuristic: if we got full page, there might be more)
  const hasMore = data.sessions.length === perPage;
  const hasPrev = page > 1;

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {data.sessions.map((session) => (
          <SessionCard key={session.id} session={session} />
        ))}
      </div>

      {/* Pagination */}
      {(hasPrev || hasMore) && (
        <div className="flex justify-center items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={!hasPrev}
            onClick={() => onPageChange(page - 1)}
          >
            <ChevronLeft className="h-4 w-4 mr-1" />
            이전
          </Button>
          <span className="px-4 text-sm text-gray-600">페이지 {page}</span>
          <Button
            variant="outline"
            size="sm"
            disabled={!hasMore}
            onClick={() => onPageChange(page + 1)}
          >
            다음
            <ChevronRight className="h-4 w-4 ml-1" />
          </Button>
        </div>
      )}
    </div>
  );
}
