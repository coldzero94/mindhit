'use client';

import { useState, useCallback } from 'react';
import { Loader2, RefreshCw, AlertCircle, Sparkles, ExternalLink, Tag, Clock, Target } from 'lucide-react';
import { toast } from 'sonner';

import { useMindmap, useGenerateMindmap } from '@/lib/hooks/use-mindmap';
import { transformApiMindmap } from '@/lib/utils/mindmap-transform';
import { MindmapCanvas } from './MindmapCanvas';
import { Galaxy } from './Galaxy';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { MindmapNode } from '@/types/mindmap';

interface MindmapViewerProps {
  sessionId: string;
}

// Node type labels
const NODE_TYPE_LABELS: Record<string, string> = {
  core: 'Core Topic',
  topic: 'Topic',
  subtopic: 'Subtopic',
  page: 'Page',
};

// Node type colors
const NODE_TYPE_COLORS: Record<string, string> = {
  core: 'bg-yellow-500',
  topic: 'bg-blue-500',
  subtopic: 'bg-green-500',
  page: 'bg-gray-500',
};

function NodeDetailPanel({ node, onClose }: { node: MindmapNode; onClose: () => void }) {
  const { keywords, description, relevance, url, summary } = node.data;

  return (
    <div className="absolute top-4 right-4 z-20 w-80 bg-white/95 backdrop-blur rounded-xl shadow-xl border border-gray-200 overflow-hidden">
      {/* Header */}
      <div className="p-4 border-b bg-gradient-to-r from-gray-50 to-white">
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2 min-w-0">
            <div
              className="w-4 h-4 rounded-full flex-shrink-0 ring-2 ring-white shadow"
              style={{ backgroundColor: node.color }}
            />
            <h3 className="font-semibold text-gray-900 truncate">{node.label}</h3>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 p-1 rounded hover:bg-gray-100 flex-shrink-0"
          >
            ✕
          </button>
        </div>
        <div className="mt-2 flex items-center gap-2">
          <Badge variant="secondary" className={`${NODE_TYPE_COLORS[node.type]} text-white`}>
            {NODE_TYPE_LABELS[node.type]}
          </Badge>
          {relevance && (
            <Badge variant="outline" className="text-xs">
              <Target className="w-3 h-3 mr-1" />
              Relevance {Math.round(relevance * 100)}%
            </Badge>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="p-4 space-y-4 max-h-[400px] overflow-y-auto">
        {/* Description */}
        {description && (
          <div>
            <h4 className="text-xs font-medium text-gray-500 uppercase tracking-wider mb-2">Description</h4>
            <p className="text-sm text-gray-700 leading-relaxed">{description}</p>
          </div>
        )}

        {/* Keywords */}
        {keywords && keywords.length > 0 && (
          <div>
            <h4 className="text-xs font-medium text-gray-500 uppercase tracking-wider mb-2 flex items-center gap-1">
              <Tag className="w-3 h-3" />
              Keywords
            </h4>
            <div className="flex flex-wrap gap-1.5">
              {keywords.map((keyword: string, idx: number) => (
                <Badge key={idx} variant="outline" className="text-xs bg-gray-50">
                  {keyword}
                </Badge>
              ))}
            </div>
          </div>
        )}

        {/* Summary for page nodes */}
        {summary && (
          <div>
            <h4 className="text-xs font-medium text-gray-500 uppercase tracking-wider mb-2">
              AI Summary
            </h4>
            <p className="text-sm text-gray-700 leading-relaxed">{summary}</p>
          </div>
        )}

        {/* URL for page nodes */}
        {node.type === 'page' && url && (
          <div>
            <h4 className="text-xs font-medium text-gray-500 uppercase tracking-wider mb-2 flex items-center gap-1">
              <ExternalLink className="w-3 h-3" />
              Page Link
            </h4>
            <a
              href={url}
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-blue-600 hover:text-blue-800 hover:underline break-all flex items-center gap-1"
            >
              {url.length > 50 ? url.slice(0, 50) + '...' : url}
              <ExternalLink className="w-3 h-3 flex-shrink-0" />
            </a>
          </div>
        )}

        {/* Stats */}
        {(node.data.visitCount || node.data.totalDuration) && (
          <div className="pt-2 border-t">
            <div className="flex items-center gap-4 text-xs text-gray-500">
              {node.data.visitCount && (
                <span className="flex items-center gap-1">
                  <Clock className="w-3 h-3" />
                  {node.data.visitCount} visits
                </span>
              )}
              {node.data.totalDuration && (
                <span>
                  {Math.round(node.data.totalDuration / 1000 / 60)} min duration
                </span>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export function MindmapViewer({ sessionId }: MindmapViewerProps) {
  const [selectedNode, setSelectedNode] = useState<MindmapNode | null>(null);

  const { data: mindmap, isLoading, error, refetch } = useMindmap(sessionId);
  const generateMindmap = useGenerateMindmap();

  const handleGenerate = useCallback(async (force: boolean = false) => {
    try {
      await generateMindmap.mutateAsync({ sessionId, options: { force } });
      toast.success(force ? 'Regenerating mindmap.' : 'Starting mindmap generation.');
    } catch {
      toast.error('Failed to generate mindmap.');
    }
  }, [sessionId, generateMindmap]);

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
          <p className="text-red-500 mb-4">Failed to load mindmap.</p>
          <Button variant="outline" onClick={() => refetch()}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </CardContent>
      </Card>
    );
  }

  // No mindmap yet (404)
  if (!mindmap) {
    return (
      <Card>
        <CardContent className="text-center py-12">
          <Sparkles className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            Mindmap not yet generated
          </h3>
          <p className="text-gray-500 mb-6">
            AI will generate a mindmap based on your browsing data.
          </p>
          <Button
            onClick={() => handleGenerate(false)}
            disabled={generateMindmap.isPending}
          >
            {generateMindmap.isPending ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Requesting generation...
              </>
            ) : (
              <>
                <Sparkles className="h-4 w-4 mr-2" />
                Generate Mindmap
              </>
            )}
          </Button>
        </CardContent>
      </Card>
    );
  }

  // Pending status (generation queued, waiting for worker)
  if (mindmap.status === 'pending') {
    return (
      <Card>
        <CardContent className="text-center py-12">
          <Loader2 className="h-12 w-12 animate-spin text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            Preparing mindmap generation
          </h3>
          <p className="text-gray-500">
            AI will analyze your browsing data shortly.
          </p>
          <p className="text-xs text-gray-400 mt-2">
            Checking status automatically...
          </p>
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
            Generating mindmap
          </h3>
          <p className="text-gray-500">
            AI is analyzing your browsing data. Please wait.
          </p>
          <p className="text-xs text-gray-400 mt-2">
            Checking status automatically...
          </p>
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
            Failed to generate mindmap
          </h3>
          <p className="text-gray-500 mb-6">
            {mindmap.error_message || 'An unknown error occurred.'}
          </p>
          <Button
            onClick={() => handleGenerate(true)}
            disabled={generateMindmap.isPending}
          >
            {generateMindmap.isPending ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Regenerating...
              </>
            ) : (
              <>
                <RefreshCw className="h-4 w-4 mr-2" />
                Regenerate
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
            No mindmap data
          </h3>
          <p className="text-gray-500 mb-6">
            The session may not have enough browsing data.
          </p>
          <Button
            onClick={() => handleGenerate(true)}
            disabled={generateMindmap.isPending}
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Regenerate
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header with stats and regenerate button */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium text-gray-900">Mindmap</h3>
          <p className="text-sm text-gray-500">
            {mindmapData.nodes.length} nodes · {mindmapData.edges.length} connections
            <span className="text-gray-400 ml-2">
              | Click to view details
            </span>
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
          <span className="ml-2">Regenerate</span>
        </Button>
      </div>

      {/* 3D Canvas with Detail Panel */}
      <div className="relative">
        <MindmapCanvas className="h-[600px]">
          <Galaxy
            data={mindmapData}
            onNodeSelect={handleNodeSelect}
            enableAutoRotate={!selectedNode}
          />
        </MindmapCanvas>

        {/* Node Detail Panel */}
        {selectedNode && (
          <NodeDetailPanel
            node={selectedNode}
            onClose={() => setSelectedNode(null)}
          />
        )}
      </div>

      {/* Legend */}
      <div className="flex items-center justify-center gap-6 text-xs text-gray-500">
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-3 rounded-full bg-yellow-500" />
          <span>Core Topic</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-3 rounded-full bg-blue-500" />
          <span>Topic</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-3 rounded-full bg-gray-400" />
          <span>Page</span>
        </div>
      </div>
    </div>
  );
}
