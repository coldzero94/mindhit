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
        keywords: ["웹", "개발", "프로그래밍"],
        description: "웹 개발의 핵심 주제",
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
        keywords: ["리액트", "프론트엔드", "컴포넌트"],
        description: "모던 프론트엔드 라이브러리",
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
        keywords: ["타입스크립트", "정적타입", "자바스크립트"],
        description: "타입 안정성을 제공하는 JavaScript 슈퍼셋",
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
        keywords: ["노드", "백엔드", "서버"],
        description: "JavaScript 런타임 환경",
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
        description: "함수형 컴포넌트에서 상태 관리",
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
        keywords: ["넥스트", "SSR", "프레임워크"],
        description: "React 기반 풀스택 프레임워크",
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
        keywords: ["타입", "인터페이스", "제네릭"],
        description: "TypeScript 기초 문법",
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
        keywords: ["익스프레스", "API", "미들웨어"],
        description: "Node.js 웹 프레임워크",
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
