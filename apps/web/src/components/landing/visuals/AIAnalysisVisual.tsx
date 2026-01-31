"use client";

/**
 * AIAnalysisVisual - Brain icon with pulsing effect and connections
 * Represents AI analysis and keyword extraction
 */
export function AIAnalysisVisual() {
  return (
    <div className="relative w-64 h-64 flex items-center justify-center">
      {/* Central Brain Shape */}
      <div className="relative">
        {/* Brain Icon (Simplified) */}
        <div className="w-32 h-32 rounded-full bg-gradient-to-br from-purple-400 to-purple-600 flex items-center justify-center shadow-2xl animate-pulse">
          <svg
            className="w-20 h-20 text-white"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            viewBox="0 0 24 24"
          >
            <path d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
          </svg>
        </div>

        {/* Pulsing Rings */}
        <div className="absolute inset-0 rounded-full border-4 border-purple-300 animate-ping opacity-20" />
        <div
          className="absolute inset-0 rounded-full border-4 border-purple-300 animate-ping opacity-10"
          style={{ animationDelay: "0.5s" }}
        />
      </div>

      {/* Orbiting Keywords (Simulated) */}
      {[
        { label: "AI", angle: 0, color: "from-blue-400 to-blue-600" },
        { label: "Tags", angle: 120, color: "from-green-400 to-green-600" },
        { label: "Analysis", angle: 240, color: "from-yellow-400 to-yellow-600" },
      ].map((keyword, i) => (
        <div
          key={i}
          className="absolute"
          style={{
            top: "50%",
            left: "50%",
            transform: `translate(-50%, -50%) rotate(${keyword.angle}deg) translateY(-100px)`,
            animation: `orbit 6s linear infinite`,
            animationDelay: `${i * 2}s`,
          }}
        >
          <div
            className={`px-3 py-1.5 rounded-full bg-gradient-to-br ${keyword.color} text-white text-xs font-semibold shadow-lg`}
            style={{ transform: `rotate(-${keyword.angle}deg)` }}
          >
            {keyword.label}
          </div>
        </div>
      ))}

      {/* Animation Keyframes */}
      <style jsx>{`
        @keyframes orbit {
          from {
            transform: translate(-50%, -50%) rotate(0deg) translateY(-100px);
          }
          to {
            transform: translate(-50%, -50%) rotate(360deg) translateY(-100px);
          }
        }
      `}</style>
    </div>
  );
}
