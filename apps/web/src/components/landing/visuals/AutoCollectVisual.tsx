"use client";

/**
 * AutoCollectVisual - Browser with animated dots flowing out
 * Represents automatic browsing history collection
 */
export function AutoCollectVisual() {
  return (
    <div className="relative w-64 h-64 flex items-center justify-center">
      {/* Browser Window */}
      <div className="relative w-48 h-32 bg-white border-2 border-border rounded-lg shadow-xl overflow-hidden">
        {/* Browser Header */}
        <div className="h-8 bg-gray-100 border-b border-border flex items-center px-3 gap-1.5">
          <div className="w-2.5 h-2.5 rounded-full bg-red-400" />
          <div className="w-2.5 h-2.5 rounded-full bg-yellow-400" />
          <div className="w-2.5 h-2.5 rounded-full bg-green-400" />
        </div>

        {/* Browser Content (Simulated) */}
        <div className="p-3 space-y-2">
          <div className="h-2 bg-gray-200 rounded w-3/4 animate-pulse" />
          <div className="h-2 bg-gray-200 rounded w-full animate-pulse" style={{ animationDelay: "0.1s" }} />
          <div className="h-2 bg-gray-200 rounded w-5/6 animate-pulse" style={{ animationDelay: "0.2s" }} />
        </div>
      </div>

      {/* Animated Dots Flowing Out */}
      {[...Array(8)].map((_, i) => (
        <div
          key={i}
          className="absolute w-3 h-3 rounded-full bg-gradient-to-br from-blue-400 to-blue-600"
          style={{
            animation: `flowOut 2s ease-out infinite`,
            animationDelay: `${i * 0.25}s`,
            top: "50%",
            left: "50%",
            opacity: 0,
          }}
        />
      ))}

      {/* Animation Keyframes */}
      <style jsx>{`
        @keyframes flowOut {
          0% {
            transform: translate(-50%, -50%) translateY(0) scale(1);
            opacity: 1;
          }
          100% {
            transform: translate(-50%, -50%) translateY(-80px) scale(0.5);
            opacity: 0;
          }
        }
      `}</style>
    </div>
  );
}
