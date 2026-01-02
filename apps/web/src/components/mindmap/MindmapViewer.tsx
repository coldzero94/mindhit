'use client';

import { useState, useCallback } from 'react';
import { Loader2, RefreshCw, AlertCircle, Sparkles } from 'lucide-react';
import { toast } from 'sonner';

import { useMindmap, useGenerateMindmap } from '@/lib/hooks/use-mindmap';
import { transformApiMindmap } from '@/lib/utils/mindmap-transform';
import { MindmapCanvas } from './MindmapCanvas';
import { Galaxy } from './Galaxy';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import type { MindmapNode } from '@/types/mindmap';

interface MindmapViewerProps {
  sessionId: string;
}

export function MindmapViewer({ sessionId }: MindmapViewerProps) {
  const [selectedNode, setSelectedNode] = useState<MindmapNode | null>(null);

  const { data: mindmap, isLoading, error, refetch } = useMindmap(sessionId);
  const generateMindmap = useGenerateMindmap();

  const handleGenerate = useCallback(async (force: boolean = false) => {
    try {
      await generateMindmap.mutateAsync({ sessionId, options: { force } });
      toast.success(force ? '마인드맵을 재생성하고 있습니다.' : '마인드맵을 생성하고 있습니다.');
      // Refetch after a short delay to check status
      setTimeout(() => refetch(), 2000);
    } catch {
      toast.error('마인드맵 생성에 실패했습니다.');
    }
  }, [sessionId, generateMindmap, refetch]);

  const handleNodeSelect = useCallback((node: MindmapNode | null) => {
    setSelectedNode(node);
  }, []);

  // Loading state
  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
        </CardContent>
      </Card>
    );
  }

  // Error state (not 404)
  if (error && (error as { response?: { status?: number } })?.response?.status !== 404) {
    return (
      <Card>
        <CardContent className="text-center py-12">
          <AlertCircle className="h-12 w-12 text-red-400 mx-auto mb-4" />
          <p className="text-red-500 mb-4">마인드맵을 불러오는데 실패했습니다.</p>
          <Button variant="outline" onClick={() => refetch()}>
            <RefreshCw className="h-4 w-4 mr-2" />
            다시 시도
          </Button>
        </CardContent>
      </Card>
    );
  }

  // No mindmap yet (404) or pending status
  if (!mindmap || mindmap.status === 'pending') {
    return (
      <Card>
        <CardContent className="text-center py-12">
          <Sparkles className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            마인드맵이 아직 생성되지 않았습니다
          </h3>
          <p className="text-gray-500 mb-6">
            세션의 브라우징 데이터를 기반으로 AI가 마인드맵을 생성합니다.
          </p>
          <Button
            onClick={() => handleGenerate(false)}
            disabled={generateMindmap.isPending}
          >
            {generateMindmap.isPending ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                생성 중...
              </>
            ) : (
              <>
                <Sparkles className="h-4 w-4 mr-2" />
                마인드맵 생성
              </>
            )}
          </Button>
        </CardContent>
      </Card>
    );
  }

  // Generating status
  if (mindmap.status === 'generating') {
    return (
      <Card>
        <CardContent className="text-center py-12">
          <Loader2 className="h-12 w-12 animate-spin text-blue-500 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            마인드맵을 생성하고 있습니다
          </h3>
          <p className="text-gray-500">
            AI가 브라우징 데이터를 분석하고 있습니다. 잠시만 기다려주세요.
          </p>
          <Button
            variant="outline"
            className="mt-4"
            onClick={() => refetch()}
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            상태 확인
          </Button>
        </CardContent>
      </Card>
    );
  }

  // Failed status
  if (mindmap.status === 'failed') {
    return (
      <Card>
        <CardContent className="text-center py-12">
          <AlertCircle className="h-12 w-12 text-red-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            마인드맵 생성에 실패했습니다
          </h3>
          <p className="text-gray-500 mb-6">
            {mindmap.error_message || '알 수 없는 오류가 발생했습니다.'}
          </p>
          <Button
            onClick={() => handleGenerate(true)}
            disabled={generateMindmap.isPending}
          >
            {generateMindmap.isPending ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                재생성 중...
              </>
            ) : (
              <>
                <RefreshCw className="h-4 w-4 mr-2" />
                다시 생성
              </>
            )}
          </Button>
        </CardContent>
      </Card>
    );
  }

  // Completed - show the mindmap
  const mindmapData = transformApiMindmap(mindmap);

  if (!mindmapData) {
    return (
      <Card>
        <CardContent className="text-center py-12">
          <AlertCircle className="h-12 w-12 text-yellow-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            마인드맵 데이터가 없습니다
          </h3>
          <p className="text-gray-500 mb-6">
            세션에 충분한 브라우징 데이터가 없을 수 있습니다.
          </p>
          <Button
            onClick={() => handleGenerate(true)}
            disabled={generateMindmap.isPending}
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            다시 생성
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header with regenerate button */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium text-gray-900">마인드맵</h3>
          <p className="text-sm text-gray-500">
            {mindmapData.nodes.length}개 노드 · {mindmapData.edges.length}개 연결
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => handleGenerate(true)}
          disabled={generateMindmap.isPending}
        >
          {generateMindmap.isPending ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <RefreshCw className="h-4 w-4" />
          )}
          <span className="ml-2">재생성</span>
        </Button>
      </div>

      {/* 3D Canvas */}
      <MindmapCanvas className="h-[500px]">
        <Galaxy data={mindmapData} onNodeSelect={handleNodeSelect} />
      </MindmapCanvas>

      {/* Selected Node Info */}
      {selectedNode && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base flex items-center gap-2">
              <div
                className="w-3 h-3 rounded-full"
                style={{ backgroundColor: selectedNode.color }}
              />
              {selectedNode.label}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-gray-600 space-y-2">
              <p>
                <span className="font-medium">유형:</span>{' '}
                {selectedNode.type === 'core' && '핵심'}
                {selectedNode.type === 'topic' && '주제'}
                {selectedNode.type === 'subtopic' && '하위 주제'}
                {selectedNode.type === 'page' && '페이지'}
              </p>
              {selectedNode.data.description && (
                <p>
                  <span className="font-medium">설명:</span>{' '}
                  {selectedNode.data.description}
                </p>
              )}
              {selectedNode.data.urls && selectedNode.data.urls.length > 0 && (
                <div>
                  <span className="font-medium">관련 URL:</span>
                  <ul className="mt-1 space-y-1">
                    {selectedNode.data.urls.slice(0, 3).map((url, index) => (
                      <li key={index}>
                        <a
                          href={url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-blue-500 hover:underline truncate block"
                        >
                          {url}
                        </a>
                      </li>
                    ))}
                    {selectedNode.data.urls.length > 3 && (
                      <li className="text-gray-400">
                        +{selectedNode.data.urls.length - 3}개 더
                      </li>
                    )}
                  </ul>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
