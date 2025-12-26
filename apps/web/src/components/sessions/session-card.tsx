"use client";

import Link from "next/link";
import { formatDistanceToNow } from "date-fns";
import { ko } from "date-fns/locale";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { SessionSession, SessionSessionStatus } from "@/api/generated/types.gen";
import { Circle, Pause, Loader2, CheckCircle, XCircle } from "lucide-react";

const statusConfig: Record<
  SessionSessionStatus,
  { label: string; variant: "default" | "secondary" | "destructive" | "outline"; icon: React.ReactNode }
> = {
  recording: {
    label: "녹화 중",
    variant: "default",
    icon: <Circle className="h-3 w-3 fill-current animate-pulse" />,
  },
  paused: {
    label: "일시정지",
    variant: "secondary",
    icon: <Pause className="h-3 w-3" />,
  },
  processing: {
    label: "처리 중",
    variant: "outline",
    icon: <Loader2 className="h-3 w-3 animate-spin" />,
  },
  completed: {
    label: "완료",
    variant: "default",
    icon: <CheckCircle className="h-3 w-3" />,
  },
  failed: {
    label: "실패",
    variant: "destructive",
    icon: <XCircle className="h-3 w-3" />,
  },
};

interface SessionCardProps {
  session: SessionSession;
}

export function SessionCard({ session }: SessionCardProps) {
  const status = statusConfig[session.session_status];
  const timeAgo = formatDistanceToNow(new Date(session.started_at), {
    addSuffix: true,
    locale: ko,
  });

  return (
    <Link href={`/sessions/${session.id}`}>
      <Card className="hover:shadow-md transition-shadow cursor-pointer h-full">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between gap-2">
            <CardTitle className="text-lg truncate flex-1">
              {session.title || "제목 없음"}
            </CardTitle>
            <Badge variant={status.variant} className="shrink-0">
              {status.icon}
              <span className="ml-1">{status.label}</span>
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          {session.description && (
            <p className="text-sm text-gray-600 mb-2 line-clamp-2">
              {session.description}
            </p>
          )}
          <p className="text-sm text-gray-500">{timeAgo}</p>
        </CardContent>
      </Card>
    </Link>
  );
}
