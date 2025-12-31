import { useState, useEffect } from "react";
import { useSettingsStore } from "@/stores/settings-store";

interface SettingsProps {
  onClose: () => void;
}

export function Settings({ onClose }: SettingsProps) {
  const {
    apiUrl,
    webUrl,
    autoStartSession,
    collectAllTabs,
    updateSettings,
    resetSettings,
  } = useSettingsStore();

  const [localApiUrl, setLocalApiUrl] = useState(apiUrl);
  const [localWebUrl, setLocalWebUrl] = useState(webUrl);
  const [showAdvanced, setShowAdvanced] = useState(false);

  useEffect(() => {
    // Store hydration
    const unsub = useSettingsStore.persist.onFinishHydration(() => {
      setLocalApiUrl(useSettingsStore.getState().apiUrl);
      setLocalWebUrl(useSettingsStore.getState().webUrl);
    });

    if (useSettingsStore.persist.hasHydrated()) {
      setLocalApiUrl(apiUrl);
      setLocalWebUrl(webUrl);
    }

    return unsub;
  }, [apiUrl, webUrl]);

  const handleSave = () => {
    updateSettings({
      apiUrl: localApiUrl,
      webUrl: localWebUrl,
    });
    onClose();
  };

  const handleReset = () => {
    if (confirm("Reset all settings to defaults?")) {
      resetSettings();
      setLocalApiUrl(
        import.meta.env.VITE_API_URL || "http://localhost:9000/v1"
      );
      setLocalWebUrl(import.meta.env.VITE_WEB_URL || "http://localhost:3000");
    }
  };

  return (
    <div className="fixed inset-0 bg-white z-50 overflow-auto">
      {/* Header */}
      <div className="sticky top-0 bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between">
        <h2 className="text-base font-semibold">Settings</h2>
        <button
          onClick={onClose}
          className="p-1 hover:bg-gray-100 rounded-full"
        >
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      <div className="p-4 space-y-6">
        {/* General Settings */}
        <section className="space-y-4">
          <h3 className="text-sm font-medium text-gray-900">General</h3>

          <label className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-700">Auto-start session</p>
              <p className="text-xs text-gray-500">
                Start session when browser opens
              </p>
            </div>
            <input
              type="checkbox"
              checked={autoStartSession}
              onChange={(e) =>
                updateSettings({ autoStartSession: e.target.checked })
              }
              className="w-4 h-4 text-blue-600 rounded"
            />
          </label>

          <label className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-700">Collect all tabs</p>
              <p className="text-xs text-gray-500">
                Disable to collect current tab only
              </p>
            </div>
            <input
              type="checkbox"
              checked={collectAllTabs}
              onChange={(e) =>
                updateSettings({ collectAllTabs: e.target.checked })
              }
              className="w-4 h-4 text-blue-600 rounded"
            />
          </label>
        </section>

        {/* Advanced Settings */}
        <section className="space-y-4">
          <button
            onClick={() => setShowAdvanced(!showAdvanced)}
            className="flex items-center gap-2 text-sm font-medium text-gray-900"
          >
            <svg
              className={`w-4 h-4 transition-transform ${showAdvanced ? "rotate-90" : ""}`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 5l7 7-7 7"
              />
            </svg>
            Advanced (Developer)
          </button>

          {showAdvanced && (
            <div className="space-y-4 pl-6">
              <div>
                <label className="block text-sm text-gray-700 mb-1">
                  API URL
                </label>
                <input
                  type="text"
                  value={localApiUrl}
                  onChange={(e) => setLocalApiUrl(e.target.value)}
                  className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg"
                  placeholder="http://localhost:9000/v1"
                />
              </div>

              <div>
                <label className="block text-sm text-gray-700 mb-1">
                  Web App URL
                </label>
                <input
                  type="text"
                  value={localWebUrl}
                  onChange={(e) => setLocalWebUrl(e.target.value)}
                  className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg"
                  placeholder="http://localhost:3000"
                />
              </div>
            </div>
          )}
        </section>

        {/* Buttons */}
        <div className="flex gap-2 pt-4 border-t">
          <button
            onClick={handleReset}
            className="flex-1 px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg"
          >
            Reset
          </button>
          <button
            onClick={handleSave}
            className="flex-1 px-4 py-2 text-sm text-white bg-blue-600 hover:bg-blue-700 rounded-lg"
          >
            Save
          </button>
        </div>

        {/* Version Info */}
        <div className="text-center text-xs text-gray-400 pt-4">
          MindHit Extension v{chrome.runtime.getManifest().version}
        </div>
      </div>
    </div>
  );
}
