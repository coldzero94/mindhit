"use client";

import { useParams, useRouter } from "next/navigation";
import { formatDistanceToNow, format } from "date-fns";
import { ko } from "date-fns/locale";
import { ArrowLeft, Trash2, Clock, Globe, FileText } from "lucide-react";
import { toast } from "sonner";

import {
  useSession,
  useSessionEvents,
  useSessionStats,
  useDeleteSession,
} from "@/lib/hooks/use-sessions";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
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
} from "@/components/ui/alert-dialog";
import type { SessionSessionStatus } from "@/api/generated/types.gen";

const statusLabels: Record<SessionSessionStatus, string> = {
  recording: "녹화 중",
  paused: "일시정지",
  processing: "처리 중",
  completed: "완료",
  failed: "실패",
};

export default function SessionDetailPage() {
  const params = useParams();
  const router = useRouter();
  const sessionId = params.id as string;

  const { data: session, isLoading, error } = useSession(sessionId);
  const { data: events } = useSessionEvents(sessionId);
  const { data: stats } = useSessionStats(sessionId);
  const deleteSession = useDeleteSession();

  const handleDelete = async () => {
    try {
      await deleteSession.mutateAsync(sessionId);
      toast.success("세션이 삭제되었습니다.");
      router.push("/sessions");
    } catch {
      toast.error("세션을 삭제하는데 실패했습니다.");
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <div className="grid gap-4 md:grid-cols-3">
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
        </div>
        <Skeleton className="h-64" />
      </div>
    );
  }

  if (error || !session) {
    return (
      <div className="text-center py-12">
        <p className="text-red-500 mb-4">세션을 찾을 수 없습니다.</p>
        <Button variant="outline" onClick={() => router.push("/sessions")}>
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
          <Button
            variant="ghost"
            size="icon"
            onClick={() => router.push("/sessions")}
          >
            <ArrowLeft className="h-5 w-5" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">
              {session.title || "제목 없음"}
            </h1>
            <p className="text-sm text-gray-500">
              {format(new Date(session.started_at), "yyyy년 MM월 dd일 HH:mm", {
                locale: ko,
              })}
              {" · "}
              {formatDistanceToNow(new Date(session.started_at), {
                addSuffix: true,
                locale: ko,
              })}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Badge
            variant={
              session.session_status === "failed" ? "destructive" : "default"
            }
          >
            {statusLabels[session.session_status]}
          </Badge>
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
                  이 작업은 되돌릴 수 없습니다. 세션과 관련된 모든 데이터가
                  영구적으로 삭제됩니다.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>취소</AlertDialogCancel>
                <AlertDialogAction onClick={handleDelete}>
                  삭제
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </div>

      {/* Description */}
      {session.description && (
        <p className="text-gray-600">{session.description}</p>
      )}

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500 flex items-center gap-2">
              <Globe className="h-4 w-4" />
              방문한 페이지
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {stats?.page_visits ?? events?.page_visits.length ?? 0}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500 flex items-center gap-2">
              <FileText className="h-4 w-4" />
              하이라이트
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {stats?.highlights ?? events?.highlights.length ?? 0}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500 flex items-center gap-2">
              <Globe className="h-4 w-4" />
              고유 URL
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{stats?.unique_urls ?? "-"}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500 flex items-center gap-2">
              <Clock className="h-4 w-4" />
              총 이벤트
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {stats?.total_events ?? events?.total ?? 0}
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
          {!events?.page_visits.length ? (
            <p className="text-gray-500 text-center py-4">
              방문한 페이지가 없습니다.
            </p>
          ) : (
            <ul className="divide-y">
              {events.page_visits.map((visit) => (
                <li
                  key={visit.id}
                  className="flex items-center justify-between py-3 hover:bg-gray-50 -mx-2 px-2 rounded"
                >
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate">
                      {visit.title || visit.url}
                    </p>
                    <p className="text-sm text-gray-500 truncate">
                      {visit.url}
                    </p>
                  </div>
                  <div className="text-sm text-gray-400 ml-4 shrink-0">
                    {visit.duration_ms
                      ? `${Math.floor(visit.duration_ms / 60000)}분 ${Math.floor((visit.duration_ms % 60000) / 1000)}초`
                      : format(new Date(visit.visited_at), "HH:mm")}
                  </div>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      {/* Highlights */}
      {events?.highlights && events.highlights.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>하이라이트</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-3">
              {events.highlights.map((highlight) => (
                <li
                  key={highlight.id}
                  className="p-3 rounded-lg bg-gray-50 border-l-4"
                  style={{ borderColor: highlight.color || "#3b82f6" }}
                >
                  <p className="text-sm">{highlight.text}</p>
                  {highlight.note && (
                    <p className="text-xs text-gray-500 mt-1">
                      {highlight.note}
                    </p>
                  )}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
