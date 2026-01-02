import { useState, useCallback, useEffect, useRef } from 'react';
import type { MindmapNode } from '@/types/mindmap';

interface UseMindmapInteractionOptions {
  onNodeSelect?: (node: MindmapNode | null) => void;
  autoRotateDelay?: number;
}

export function useMindmapInteraction(options: UseMindmapInteractionOptions = {}) {
  const { onNodeSelect, autoRotateDelay = 5000 } = options;

  const [selectedNode, setSelectedNode] = useState<MindmapNode | null>(null);
  const [hoveredNode, setHoveredNode] = useState<MindmapNode | null>(null);
  const [isIdle, setIsIdle] = useState(true);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Node click handler
  const handleNodeClick = useCallback(
    (node: MindmapNode) => {
      const newNode = selectedNode?.id === node.id ? null : node;
      setSelectedNode(newNode);
      setIsIdle(false);
      onNodeSelect?.(newNode);
    },
    [selectedNode, onNodeSelect]
  );

  // Hover handler
  const handleNodeHover = useCallback((node: MindmapNode | null) => {
    setHoveredNode(node);
    if (node) {
      setIsIdle(false);
    }
  }, []);

  // Background click to deselect
  const handleBackgroundClick = useCallback(() => {
    setSelectedNode(null);
    onNodeSelect?.(null);
  }, [onNodeSelect]);

  // Idle detection (enable auto-rotate after delay)
  // Use separate effect for timer management
  useEffect(() => {
    // Clear any existing timer
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }

    // If user is interacting, stay not idle
    if (selectedNode || hoveredNode) {
      return;
    }

    // Start timer to set idle state
    timerRef.current = setTimeout(() => {
      setIsIdle(true);
    }, autoRotateDelay);

    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
    };
  }, [selectedNode, hoveredNode, autoRotateDelay]);

  return {
    selectedNode,
    hoveredNode,
    isIdle,
    handleNodeClick,
    handleNodeHover,
    handleBackgroundClick,
  };
}
