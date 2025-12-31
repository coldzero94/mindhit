'use client';

import { useState, useCallback, useMemo } from 'react';
import { useThree } from '@react-three/fiber';
import { Node } from './Node';
import { Edge } from './Edge';
import type { MindmapData, MindmapNode } from '@/types/mindmap';

interface GalaxyProps {
  data: MindmapData;
  onNodeSelect: (node: MindmapNode | null) => void;
}

export function Galaxy({ data, onNodeSelect }: GalaxyProps) {
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [hoveredNodeId, setHoveredNodeId] = useState<string | null>(null);
  const { controls } = useThree();

  // Create node map for edge lookup
  const nodeMap = useMemo(() => {
    return new Map(data.nodes.map((node) => [node.id, node]));
  }, [data.nodes]);

  // Get connected node IDs for highlighting
  const connectedNodeIds = useMemo(() => {
    const targetId = selectedNodeId || hoveredNodeId;
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
  }, [data.edges, selectedNodeId, hoveredNodeId]);

  // Handle node click
  const handleNodeClick = useCallback(
    (node: MindmapNode) => {
      if (selectedNodeId === node.id) {
        // Deselect
        setSelectedNodeId(null);
        onNodeSelect(null);
      } else {
        // Select
        setSelectedNodeId(node.id);
        onNodeSelect(node);

        // Animate camera to node
        if (node.position && controls) {
          const orbitControls = controls as { target?: { set: (x: number, y: number, z: number) => void } };
          if (orbitControls.target) {
            // Smoothly move camera target to node position
            orbitControls.target.set(
              node.position.x,
              node.position.y,
              node.position.z
            );
          }
        }
      }
    },
    [selectedNodeId, onNodeSelect, controls]
  );

  // Handle node hover
  const handleNodeHover = useCallback((nodeId: string, hovered: boolean) => {
    setHoveredNodeId(hovered ? nodeId : null);
  }, []);

  // Handle background click (deselect)
  const handleBackgroundClick = useCallback(() => {
    if (selectedNodeId) {
      setSelectedNodeId(null);
      onNodeSelect(null);
    }
  }, [selectedNodeId, onNodeSelect]);

  return (
    <group onClick={handleBackgroundClick}>
      {/* Invisible background plane for click detection */}
      <mesh position={[0, 0, -500]} visible={false}>
        <planeGeometry args={[2000, 2000]} />
        <meshBasicMaterial transparent opacity={0} />
      </mesh>

      {/* Render edges first (behind nodes) */}
      {data.edges.map((edge) => {
        const sourceNode = nodeMap.get(edge.source);
        const targetNode = nodeMap.get(edge.target);

        if (!sourceNode || !targetNode) return null;

        const isHighlighted =
          edge.source === selectedNodeId ||
          edge.target === selectedNodeId ||
          edge.source === hoveredNodeId ||
          edge.target === hoveredNodeId;

        return (
          <Edge
            key={`${edge.source}-${edge.target}`}
            edge={edge}
            sourceNode={sourceNode}
            targetNode={targetNode}
            isHighlighted={isHighlighted}
          />
        );
      })}

      {/* Render nodes */}
      {data.nodes.map((node) => (
        <Node
          key={node.id}
          node={node}
          isSelected={node.id === selectedNodeId}
          isHovered={
            node.id === hoveredNodeId ||
            connectedNodeIds.has(node.id)
          }
          onClick={() => handleNodeClick(node)}
          onHover={(hovered) => handleNodeHover(node.id, hovered)}
        />
      ))}
    </group>
  );
}
