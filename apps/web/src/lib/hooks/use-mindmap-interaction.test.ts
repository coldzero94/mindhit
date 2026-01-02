import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useMindmapInteraction } from "./use-mindmap-interaction";
import type { MindmapNode } from "@/types/mindmap";

// Mock nodes for testing
const mockCoreNode: MindmapNode = {
  id: "node-core",
  label: "Core",
  type: "core",
  size: 20,
  color: "#4F46E5",
  position: { x: 0, y: 0, z: 0 },
  data: {},
};

const mockTopicNode: MindmapNode = {
  id: "node-topic-1",
  label: "Topic 1",
  type: "topic",
  size: 15,
  color: "#10B981",
  position: { x: 100, y: 50, z: 0 },
  data: {},
};

const mockSubtopicNode: MindmapNode = {
  id: "node-subtopic-1",
  label: "Subtopic 1",
  type: "subtopic",
  size: 10,
  color: "#F59E0B",
  position: { x: 150, y: 75, z: 0 },
  data: {},
};

describe("useMindmapInteraction", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe("initial state", () => {
    it("should have no selected node initially", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      expect(result.current.selectedNode).toBeNull();
    });

    it("should have no hovered node initially", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      expect(result.current.hoveredNode).toBeNull();
    });

    it("should be idle initially", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      expect(result.current.isIdle).toBe(true);
    });
  });

  describe("handleNodeClick", () => {
    it("should select a node when clicked", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });

      expect(result.current.selectedNode).toEqual(mockCoreNode);
    });

    it("should deselect when clicking the same node", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      // Select node
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });
      expect(result.current.selectedNode).toEqual(mockCoreNode);

      // Click same node to deselect
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });
      expect(result.current.selectedNode).toBeNull();
    });

    it("should switch to a different node when clicking another", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      // Select first node
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });
      expect(result.current.selectedNode).toEqual(mockCoreNode);

      // Select different node
      act(() => {
        result.current.handleNodeClick(mockTopicNode);
      });
      expect(result.current.selectedNode).toEqual(mockTopicNode);
    });

    it("should call onNodeSelect callback when node is selected", () => {
      const onNodeSelect = vi.fn();
      const { result } = renderHook(() =>
        useMindmapInteraction({ onNodeSelect })
      );

      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });

      expect(onNodeSelect).toHaveBeenCalledWith(mockCoreNode);
    });

    it("should call onNodeSelect with null when node is deselected", () => {
      const onNodeSelect = vi.fn();
      const { result } = renderHook(() =>
        useMindmapInteraction({ onNodeSelect })
      );

      // Select node
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });

      // Deselect by clicking same node
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });

      expect(onNodeSelect).toHaveBeenLastCalledWith(null);
    });

    it("should set isIdle to false when node is clicked", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      expect(result.current.isIdle).toBe(true);

      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });

      expect(result.current.isIdle).toBe(false);
    });
  });

  describe("handleNodeHover", () => {
    it("should set hovered node", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      act(() => {
        result.current.handleNodeHover(mockTopicNode);
      });

      expect(result.current.hoveredNode).toEqual(mockTopicNode);
    });

    it("should clear hovered node when passing null", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      // Hover node
      act(() => {
        result.current.handleNodeHover(mockTopicNode);
      });
      expect(result.current.hoveredNode).toEqual(mockTopicNode);

      // Clear hover
      act(() => {
        result.current.handleNodeHover(null);
      });
      expect(result.current.hoveredNode).toBeNull();
    });

    it("should set isIdle to false when hovering a node", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      expect(result.current.isIdle).toBe(true);

      act(() => {
        result.current.handleNodeHover(mockTopicNode);
      });

      expect(result.current.isIdle).toBe(false);
    });

    it("should not set isIdle to false when hover is cleared", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      // Hover then clear
      act(() => {
        result.current.handleNodeHover(mockTopicNode);
      });
      act(() => {
        result.current.handleNodeHover(null);
      });

      // isIdle should still be false until timer expires
      expect(result.current.isIdle).toBe(false);
    });
  });

  describe("handleBackgroundClick", () => {
    it("should deselect node when background is clicked", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      // Select a node
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });
      expect(result.current.selectedNode).toEqual(mockCoreNode);

      // Click background
      act(() => {
        result.current.handleBackgroundClick();
      });
      expect(result.current.selectedNode).toBeNull();
    });

    it("should call onNodeSelect with null when background is clicked", () => {
      const onNodeSelect = vi.fn();
      const { result } = renderHook(() =>
        useMindmapInteraction({ onNodeSelect })
      );

      // Select a node
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });

      // Click background
      act(() => {
        result.current.handleBackgroundClick();
      });

      expect(onNodeSelect).toHaveBeenLastCalledWith(null);
    });
  });

  describe("idle detection", () => {
    it("should become idle after autoRotateDelay (default 5000ms)", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      // Click a node (sets isIdle to false)
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });
      expect(result.current.isIdle).toBe(false);

      // Deselect
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });

      // Advance timer to just before delay
      act(() => {
        vi.advanceTimersByTime(4999);
      });
      expect(result.current.isIdle).toBe(false);

      // Advance timer past delay
      act(() => {
        vi.advanceTimersByTime(2);
      });
      expect(result.current.isIdle).toBe(true);
    });

    it("should use custom autoRotateDelay", () => {
      const customDelay = 2000;
      const { result } = renderHook(() =>
        useMindmapInteraction({ autoRotateDelay: customDelay })
      );

      // Click and deselect
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });

      // Advance timer
      act(() => {
        vi.advanceTimersByTime(customDelay + 1);
      });

      expect(result.current.isIdle).toBe(true);
    });

    it("should not become idle while node is selected", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      // Select a node
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });

      // Advance timer way past delay
      act(() => {
        vi.advanceTimersByTime(10000);
      });

      // Should still not be idle because node is selected
      expect(result.current.isIdle).toBe(false);
    });

    it("should not become idle while node is hovered", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      // Hover a node
      act(() => {
        result.current.handleNodeHover(mockTopicNode);
      });

      // Advance timer way past delay
      act(() => {
        vi.advanceTimersByTime(10000);
      });

      // Should still not be idle because node is hovered
      expect(result.current.isIdle).toBe(false);
    });

    it("should reset idle timer on new interaction", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      // Start timer by deselecting
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });

      // Advance halfway
      act(() => {
        vi.advanceTimersByTime(3000);
      });

      // New interaction
      act(() => {
        result.current.handleNodeClick(mockTopicNode);
      });
      act(() => {
        result.current.handleNodeClick(mockTopicNode);
      });

      // Advance timer (should have reset)
      act(() => {
        vi.advanceTimersByTime(3000);
      });

      // Should not be idle yet
      expect(result.current.isIdle).toBe(false);

      // Complete the timer
      act(() => {
        vi.advanceTimersByTime(2001);
      });

      expect(result.current.isIdle).toBe(true);
    });
  });

  describe("multiple nodes interaction", () => {
    it("should handle rapid node switching", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      // Rapidly switch between nodes
      act(() => {
        result.current.handleNodeClick(mockCoreNode);
      });
      act(() => {
        result.current.handleNodeClick(mockTopicNode);
      });
      act(() => {
        result.current.handleNodeClick(mockSubtopicNode);
      });

      expect(result.current.selectedNode).toEqual(mockSubtopicNode);
    });

    it("should handle hover and click interactions", () => {
      const { result } = renderHook(() => useMindmapInteraction());

      // Hover one node
      act(() => {
        result.current.handleNodeHover(mockCoreNode);
      });

      // Click another node
      act(() => {
        result.current.handleNodeClick(mockTopicNode);
      });

      // Both should work independently
      expect(result.current.hoveredNode).toEqual(mockCoreNode);
      expect(result.current.selectedNode).toEqual(mockTopicNode);
    });
  });
});
