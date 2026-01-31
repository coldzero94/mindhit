import type { MindmapData } from "@/types/mindmap";

/**
 * Demo mindmap data for landing page 3D preview
 * Simple structure showcasing the mindmap visualization
 */
export const demoMindmapData: MindmapData = {
  nodes: [
    // Core node (center)
    {
      id: "1",
      label: "Web Development",
      type: "core",
      size: 20,
      color: "#FFD700",
      position: { x: 0, y: 0, z: 0 },
      data: {
        keywords: ["web", "development", "programming"],
        description: "Core topics of web development",
      },
    },

    // Topic nodes (main branches)
    {
      id: "2",
      label: "React",
      type: "topic",
      size: 15,
      color: "#61DAFB",
      position: { x: 80, y: 30, z: 20 },
      data: {
        keywords: ["react", "frontend", "component"],
        description: "Modern frontend library",
      },
    },
    {
      id: "3",
      label: "TypeScript",
      type: "topic",
      size: 15,
      color: "#3178C6",
      position: { x: -80, y: 30, z: -20 },
      data: {
        keywords: ["typescript", "static-typing", "javascript"],
        description: "JavaScript superset with type safety",
      },
    },
    {
      id: "4",
      label: "Node.js",
      type: "topic",
      size: 15,
      color: "#68A063",
      position: { x: 0, y: -60, z: 40 },
      data: {
        keywords: ["node", "backend", "server"],
        description: "JavaScript runtime environment",
      },
    },

    // Page nodes (leaves)
    {
      id: "5",
      label: "React Hooks",
      type: "page",
      size: 10,
      color: "#9BA3AF",
      position: { x: 120, y: 50, z: 40 },
      data: {
        keywords: ["hooks", "useState", "useEffect"],
        description: "State management in functional components",
        url: "https://react.dev/reference/react",
      },
    },
    {
      id: "6",
      label: "Next.js",
      type: "page",
      size: 10,
      color: "#9BA3AF",
      position: { x: 100, y: 0, z: -30 },
      data: {
        keywords: ["nextjs", "SSR", "framework"],
        description: "React-based fullstack framework",
        url: "https://nextjs.org",
      },
    },
    {
      id: "7",
      label: "TypeScript Basics",
      type: "page",
      size: 10,
      color: "#9BA3AF",
      position: { x: -110, y: 10, z: 30 },
      data: {
        keywords: ["types", "interface", "generics"],
        description: "TypeScript fundamentals",
        url: "https://www.typescriptlang.org/docs",
      },
    },
    {
      id: "8",
      label: "Express.js",
      type: "page",
      size: 10,
      color: "#9BA3AF",
      position: { x: -30, y: -80, z: 60 },
      data: {
        keywords: ["express", "API", "middleware"],
        description: "Node.js web framework",
        url: "https://expressjs.com",
      },
    },
  ],

  edges: [
    // Core to topics
    { source: "1", target: "2", weight: 1.0 },
    { source: "1", target: "3", weight: 1.0 },
    { source: "1", target: "4", weight: 1.0 },

    // Topics to pages
    { source: "2", target: "5", weight: 0.8 },
    { source: "2", target: "6", weight: 0.9 },
    { source: "3", target: "7", weight: 0.9 },
    { source: "4", target: "8", weight: 0.9 },

    // Cross-connections
    { source: "3", target: "2", weight: 0.5 },
    { source: "4", target: "6", weight: 0.6 },
  ],

  layout: {
    type: "galaxy",
    params: {},
  },
};
