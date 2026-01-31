"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { formatDistanceToNow, format } from "date-fns";
import { enUS } from "date-fns/locale";
import { ArrowLeft, Trash2, Clock, Globe, FileText, Network } from "lucide-react";
import { toast } from "sonner";

import {
  useSession,
  useSessionEvents,
  useSessionStats,
  useDeleteSession,
  useUpdateSession,
} from "@/lib/hooks/use-sessions";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
import { MindmapViewer } from "@/components/mindmap/MindmapViewer";
import { SessionTitleEdit } from "@/components/sessions/SessionTitleEdit";
import type { SessionSessionStatus } from "@/api/generated/types.gen";

const statusLabels: Record<SessionSessionStatus, string> = {
  recording: "Recording",
  paused: "Paused",
  processing: "Processing",
  completed: "Completed",
  failed: "Failed",
};

function formatDuration(ms: number): string {
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m`;
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  }
  return `${seconds}s`;
}

export default function SessionDetailPage() {
  const params = useParams();
  const router = useRouter();
  const sessionId = params.id as string;
  const [activeTab, setActiveTab] = useState<string>("events");

  const { data: session, isLoading, error } = useSession(sessionId);
  const { data: events } = useSessionEvents(sessionId);
  const { data: stats } = useSessionStats(sessionId);
  const deleteSession = useDeleteSession();
  const updateSession = useUpdateSession();

  const handleDelete = async () => {
    try {
      await deleteSession.mutateAsync(sessionId);
      toast.success("Session deleted successfully.");
      router.push("/sessions");
    } catch {
      toast.error("Failed to delete session.");
    }
  };

  const handleTitleUpdate = async (newTitle: string) => {
    await updateSession.mutateAsync({
      id: sessionId,
      data: { title: newTitle },
    });
    toast.success("Title updated successfully.");
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
        <p className="text-red-500 mb-4">Session not found.</p>
        <Button variant="outline" onClick={() => router.push("/sessions")}>
          Back to List
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
            <SessionTitleEdit
              title={session.title || "Untitled"}
              onSave={handleTitleUpdate}
            />
            <p className="text-sm text-gray-500 ml-2">
              {format(new Date(session.started_at), "MMM dd, yyyy HH:mm", {
                locale: enUS,
              })}
              {" · "}
              {formatDistanceToNow(new Date(session.started_at), {
                addSuffix: true,
                locale: enUS,
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
                <AlertDialogTitle>Delete this session?</AlertDialogTitle>
                <AlertDialogDescription>
                  This action cannot be undone. All data related to this session
                  will be permanently deleted.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction onClick={handleDelete}>
                  Delete
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
              Pages Visited
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
              Highlights
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
              Unique URLs
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
              Total Events
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {stats?.total_events ?? events?.total ?? 0}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Tabs: Events / Mindmap */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-4">
        <TabsList>
          <TabsTrigger value="events" className="flex items-center gap-2">
            <Globe className="h-4 w-4" />
            Events
          </TabsTrigger>
          <TabsTrigger value="mindmap" className="flex items-center gap-2">
            <Network className="h-4 w-4" />
            Mindmap
          </TabsTrigger>
        </TabsList>

        {/* Events Tab */}
        <TabsContent value="events" className="space-y-6">
          {/* Page Visits */}
          <Card>
            <CardHeader>
              <CardTitle>Visited Pages</CardTitle>
            </CardHeader>
            <CardContent>
              {!events?.page_visits.length ? (
                <p className="text-gray-500 text-center py-4">
                  No pages visited.
                </p>
              ) : (
                <ul className="divide-y">
                  {events.page_visits.map((visit) => (
                    <li
                      key={visit.id}
                      className="py-4 hover:bg-gray-50 -mx-2 px-2 rounded"
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1 min-w-0">
                          <p className="font-medium truncate">
                            {visit.title || visit.url}
                          </p>
                          <p className="text-sm text-gray-500 truncate">
                            {visit.url}
                          </p>

                          {/* AI Summary */}
                          {visit.summary && (
                            <p className="text-sm text-gray-600 mt-2 line-clamp-2">
                              {visit.summary}
                            </p>
                          )}

                          {/* Keywords */}
                          {visit.keywords && visit.keywords.length > 0 && (
                            <div className="flex gap-1 mt-2 flex-wrap">
                              {visit.keywords.slice(0, 5).map((keyword) => (
                                <Badge
                                  key={keyword}
                                  variant="secondary"
                                  className="text-xs"
                                >
                                  {keyword}
                                </Badge>
                              ))}
                            </div>
                          )}
                        </div>

                        {/* Visit Stats */}
                        <div className="text-right ml-4 shrink-0">
                          {visit.visit_count && visit.visit_count > 1 && (
                            <p className="text-sm font-medium text-blue-600">
                              {visit.visit_count} visits
                            </p>
                          )}
                          <p className="text-sm text-gray-400">
                            {visit.total_duration_ms
                              ? formatDuration(visit.total_duration_ms)
                              : visit.duration_ms
                                ? formatDuration(visit.duration_ms)
                                : format(new Date(visit.visited_at), "HH:mm:ss")}
                          </p>
                        </div>
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
                <CardTitle>Highlights</CardTitle>
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
        </TabsContent>

        {/* Mindmap Tab */}
        <TabsContent value="mindmap">
          <MindmapViewer sessionId={sessionId} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
