"use client";

/**
 * MindmapVisual - Connected nodes in a 3D-like circular pattern
 * Represents the final 3D mindmap visualization
 */
export function MindmapVisual() {
  const nodes = [
    { id: 1, x: 128, y: 64, size: 20, color: "from-yellow-400 to-yellow-600", label: "Core" },
    { id: 2, x: 80, y: 128, size: 16, color: "from-blue-400 to-blue-600", label: "A" },
    { id: 3, x: 176, y: 128, size: 16, color: "from-green-400 to-green-600", label: "B" },
    { id: 4, x: 128, y: 192, size: 16, color: "from-purple-400 to-purple-600", label: "C" },
    { id: 5, x: 48, y: 176, size: 12, color: "from-pink-400 to-pink-600", label: "1" },
    { id: 6, x: 208, y: 176, size: 12, color: "from-indigo-400 to-indigo-600", label: "2" },
  ];

  const edges = [
    { from: 1, to: 2 },
    { from: 1, to: 3 },
    { from: 1, to: 4 },
    { from: 2, to: 5 },
    { from: 3, to: 6 },
  ];

  return (
    <div className="relative w-64 h-64">
      <svg className="absolute inset-0" viewBox="0 0 256 256">
        {/* Connection Lines */}
        {edges.map((edge, i) => {
          const fromNode = nodes.find((n) => n.id === edge.from)!;
          const toNode = nodes.find((n) => n.id === edge.to)!;
          return (
            <line
              key={i}
              x1={fromNode.x}
              y1={fromNode.y}
              x2={toNode.x}
              y2={toNode.y}
              stroke="oklch(0.7 0 0)"
              strokeWidth="2"
              strokeOpacity="0.3"
              className="animate-pulse"
              style={{ animationDelay: `${i * 0.2}s` }}
            />
          );
        })}
      </svg>

      {/* Nodes */}
      {nodes.map((node, i) => (
        <div
          key={node.id}
          className="absolute"
          style={{
            left: `${node.x}px`,
            top: `${node.y}px`,
            transform: "translate(-50%, -50%)",
            animation: `nodeFloat ${3 + i * 0.3}s ease-in-out infinite`,
            animationDelay: `${i * 0.1}s`,
          }}
        >
          <div
            className={`rounded-full bg-gradient-to-br ${node.color} shadow-xl flex items-center justify-center text-white text-xs font-bold`}
            style={{
              width: `${node.size * 2}px`,
              height: `${node.size * 2}px`,
            }}
          >
            {node.label}
          </div>
        </div>
      ))}

      {/* Animation Keyframes */}
      <style jsx>{`
        @keyframes nodeFloat {
          0%,
          100% {
            transform: translate(-50%, -50%) translateY(0px);
          }
          50% {
            transform: translate(-50%, -50%) translateY(-10px);
          }
        }
      `}</style>
    </div>
  );
}
