import { useSessionStore } from "@/stores/session-store";

function formatTime(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
  }
  return `${minutes}:${secs.toString().padStart(2, "0")}`;
}

export function SessionStats() {
  const { pageCount, highlightCount, elapsedSeconds } = useSessionStore();

  return (
    <div className="bg-white rounded-lg p-3 shadow-sm">
      <div className="grid grid-cols-3 gap-2 text-center">
        <div>
          <p className="text-lg font-bold text-gray-900">
            {formatTime(elapsedSeconds)}
          </p>
          <p className="text-xs text-gray-500">Elapsed</p>
        </div>
        <div>
          <p className="text-lg font-bold text-gray-900">{pageCount}</p>
          <p className="text-xs text-gray-500">Pages</p>
        </div>
        <div>
          <p className="text-lg font-bold text-gray-900">{highlightCount}</p>
          <p className="text-xs text-gray-500">Highlights</p>
        </div>
      </div>
    </div>
  );
}
