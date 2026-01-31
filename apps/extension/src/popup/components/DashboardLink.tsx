import { WEB_APP_URL } from "@/lib/constants";

export function DashboardLink() {
  const openDashboard = () => {
    chrome.tabs.create({ url: WEB_APP_URL });
  };

  return (
    <button
      onClick={openDashboard}
      className="flex items-center gap-1.5 px-2 py-1 text-xs text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded transition-colors"
      title="Open Dashboard"
    >
      <svg
        className="w-3.5 h-3.5"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
        />
      </svg>
      <span>Dashboard</span>
    </button>
  );
}
