'use client';

import { useState, useCallback, useMemo, useEffect } from 'react';
import { ThreeEvent } from '@react-three/fiber';
import type { MindmapData, MindmapNode } from '@/types/mindmap';
import { Node } from './Node';
import { Edge } from './Edge';
import { BigBangAnimation } from './BigBangAnimation';
import { ParticleField } from './ParticleField';
import { NebulaEffect } from './NebulaEffect';
import { CameraController } from './CameraController';
import { AutoRotateCamera } from './AutoRotateCamera';
import { PostProcessing } from './PostProcessing';
import { useMindmapInteraction } from '@/lib/hooks/use-mindmap-interaction';
import { getOptimalParticleCount } from '@/lib/utils/three-performance';

interface GalaxyProps {
  data: MindmapData;
  onNodeSelect: (node: MindmapNode | null) => void;
  enableAnimation?: boolean;
  enableAutoRotate?: boolean;
  enablePostProcessing?: boolean;
}

export function Galaxy({
  data,
  onNodeSelect,
  enableAnimation = true,
  enableAutoRotate = true,
  enablePostProcessing = true,
}: GalaxyProps) {
  const [isAnimationReady, setIsAnimationReady] = useState(!enableAnimation);
  const [animationComplete, setAnimationComplete] = useState(!enableAnimation);

  const {
    selectedNode,
    hoveredNode,
    isIdle,
    handleNodeClick,
    handleNodeHover,
    handleBackgroundClick,
  } = useMindmapInteraction({
    onNodeSelect,
  });

  // Create node map for edge lookup
  const nodeMap = useMemo(() => {
    return new Map(data.nodes.map((node) => [node.id, node]));
  }, [data.nodes]);

  // Get connected node IDs for highlighting
  const connectedNodeIds = useMemo(() => {
    const targetId = selectedNode?.id || hoveredNode?.id;
    if (!targetId) return new Set<string>();

    const connected = new Set<string>();
    data.edges.forEach((edge) => {
      if (edge.source === targetId) {
        connected.add(edge.target);
      }
      if (edge.target === targetId) {
        connected.add(edge.source);
      }
    });

    return connected;
  }, [data.edges, selectedNode, hoveredNode]);

  // Optimize particle count based on node count
  const particleCount = useMemo(
    () => getOptimalParticleCount(data.nodes.length),
    [data.nodes.length]
  );

  // Handle background click
  const handleCanvasClick = useCallback(
    (e: ThreeEvent<MouseEvent>) => {
      // Only process if not clicking on a node
      if (e.object.type === 'Mesh' && e.object.userData?.isNode) return;
      handleBackgroundClick();
    },
    [handleBackgroundClick]
  );

  // Start animation after component mount
  useEffect(() => {
    if (enableAnimation) {
      const timer = setTimeout(() => setIsAnimationReady(true), 100);
      return () => clearTimeout(timer);
    }
  }, [enableAnimation]);

  return (
    <>
      {/* Camera controls */}
      <CameraController selectedNode={selectedNode} />
      {enableAutoRotate && (
        <AutoRotateCamera enabled={isIdle && animationComplete} speed={0.05} />
      )}

      {/* Background effects */}
      <ParticleField count={particleCount} radius={500} size={1} color="#ffffff" />
      <NebulaEffect count={300} radius={350} />

      {/* Click detection plane */}
      <mesh
        position={[0, 0, -500]}
        onClick={handleCanvasClick}
        visible={false}
      >
        <planeGeometry args={[2000, 2000]} />
        <meshBasicMaterial transparent opacity={0} />
      </mesh>

      {/* Main content */}
      <BigBangAnimation
        isReady={isAnimationReady}
        onComplete={() => setAnimationComplete(true)}
        duration={1500}
      >
        {/* Edges (render first, behind nodes) */}
        {data.edges.map((edge) => {
          const sourceNode = nodeMap.get(edge.source);
          const targetNode = nodeMap.get(edge.target);

          if (!sourceNode || !targetNode) return null;

          const isHighlighted =
            selectedNode &&
            (edge.source === selectedNode.id || edge.target === selectedNode.id);

          const isHoveredConnection =
            hoveredNode &&
            (edge.source === hoveredNode.id || edge.target === hoveredNode.id);

          return (
            <Edge
              key={`${edge.source}-${edge.target}`}
              edge={edge}
              sourceNode={sourceNode}
              targetNode={targetNode}
              isHighlighted={isHighlighted || isHoveredConnection || false}
            />
          );
        })}

        {/* Nodes */}
        {data.nodes.map((node) => {
          const isSelected = selectedNode?.id === node.id;
          const isHovered =
            hoveredNode?.id === node.id || connectedNodeIds.has(node.id);

          return (
            <Node
              key={node.id}
              node={node}
              isSelected={isSelected}
              isHovered={isHovered}
              onClick={() => handleNodeClick(node)}
              onHover={(hovered) => handleNodeHover(hovered ? node : null)}
            />
          );
        })}
      </BigBangAnimation>

      {/* Post-processing effects */}
      {enablePostProcessing && (
        <PostProcessing bloomIntensity={0.6} bloomThreshold={0.15} />
      )}
    </>
  );
}
