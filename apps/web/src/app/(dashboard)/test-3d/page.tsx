'use client';

import { useState } from 'react';
import { MindmapCanvas } from '@/components/mindmap/MindmapCanvas';
import { Galaxy } from '@/components/mindmap/Galaxy';
import { mockMindmapData } from '@/lib/mock-mindmap-data';
import type { MindmapNode } from '@/types/mindmap';

export default function Test3DPage() {
  const [selectedNode, setSelectedNode] = useState<MindmapNode | null>(null);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">3D Mindmap Test</h1>
        <p className="text-gray-500 mt-1">
          Galaxy component rendering test
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 3D Canvas */}
        <div className="lg:col-span-2 bg-white rounded-xl shadow-sm p-4">
          <h2 className="text-lg font-medium mb-4">Mindmap</h2>
          <MindmapCanvas className="h-[600px]">
            <Galaxy data={mockMindmapData} onNodeSelect={setSelectedNode} />
          </MindmapCanvas>
        </div>

        {/* Node Detail Panel */}
        <div className="bg-white rounded-xl shadow-sm p-4">
          <h2 className="text-lg font-medium mb-4">Node Information</h2>

          {selectedNode ? (
            <div className="space-y-4">
              <div className="flex items-center gap-3">
                <div
                  className="w-4 h-4 rounded-full"
                  style={{ backgroundColor: selectedNode.color }}
                />
                <span className="font-medium">{selectedNode.label}</span>
              </div>

              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-500">Type</span>
                  <span className="capitalize">{selectedNode.type}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-500">Size</span>
                  <span>{selectedNode.size}</span>
                </div>
                {selectedNode.position && (
                  <div className="flex justify-between">
                    <span className="text-gray-500">Position</span>
                    <span className="text-xs">
                      ({selectedNode.position.x.toFixed(0)},{' '}
                      {selectedNode.position.y.toFixed(0)},{' '}
                      {selectedNode.position.z.toFixed(0)})
                    </span>
                  </div>
                )}
                {selectedNode.data.visitCount && (
                  <div className="flex justify-between">
                    <span className="text-gray-500">Visit Count</span>
                    <span>{selectedNode.data.visitCount}</span>
                  </div>
                )}
              </div>

              {selectedNode.data.description && (
                <div className="pt-2 border-t">
                  <p className="text-sm text-gray-600">
                    {selectedNode.data.description}
                  </p>
                </div>
              )}

              {selectedNode.data.urls && selectedNode.data.urls.length > 0 && (
                <div className="pt-2 border-t">
                  <p className="text-sm font-medium text-gray-500 mb-1">Related URLs</p>
                  <ul className="space-y-1">
                    {selectedNode.data.urls.map((url, i) => (
                      <li key={i} className="text-sm text-blue-600 truncate">
                        {url}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          ) : (
            <p className="text-gray-400 text-sm">
              Click a node to view details
            </p>
          )}
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm p-4">
        <h2 className="text-lg font-medium mb-2">Controls</h2>
        <ul className="text-sm text-gray-600 space-y-1">
          <li>• Mouse drag: Rotate</li>
          <li>• Scroll: Zoom</li>
          <li>• Click node: Select and show details</li>
          <li>• Hover node: Highlight connected nodes</li>
          <li>• Click empty space: Deselect</li>
        </ul>
      </div>
    </div>
  );
}
