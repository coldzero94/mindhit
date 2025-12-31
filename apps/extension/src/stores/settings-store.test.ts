import { describe, it, expect, beforeEach } from "vitest";
import { useSettingsStore } from "./settings-store";

describe("SettingsStore", () => {
  beforeEach(() => {
    useSettingsStore.getState().resetSettings();
  });

  describe("initial state", () => {
    it("should have default settings initially", () => {
      const state = useSettingsStore.getState();
      expect(state.apiUrl).toBe("http://localhost:9000/v1");
      expect(state.webUrl).toBe("http://localhost:3000");
      expect(state.autoStartSession).toBe(false);
      expect(state.collectAllTabs).toBe(true);
    });
  });

  describe("updateSettings", () => {
    it("should update single setting", () => {
      useSettingsStore.getState().updateSettings({ autoStartSession: true });
      const state = useSettingsStore.getState();

      expect(state.autoStartSession).toBe(true);
      expect(state.collectAllTabs).toBe(true); // unchanged
    });

    it("should update multiple settings", () => {
      useSettingsStore.getState().updateSettings({
        autoStartSession: true,
        collectAllTabs: false,
      });
      const state = useSettingsStore.getState();

      expect(state.autoStartSession).toBe(true);
      expect(state.collectAllTabs).toBe(false);
    });

    it("should update developer settings", () => {
      useSettingsStore.getState().updateSettings({
        apiUrl: "https://api.example.com/v1",
        webUrl: "https://app.example.com",
      });
      const state = useSettingsStore.getState();

      expect(state.apiUrl).toBe("https://api.example.com/v1");
      expect(state.webUrl).toBe("https://app.example.com");
    });

    it("should preserve other settings when updating", () => {
      useSettingsStore.getState().updateSettings({
        autoStartSession: true,
        apiUrl: "https://custom.api.com",
      });
      useSettingsStore.getState().updateSettings({
        collectAllTabs: false,
      });
      const state = useSettingsStore.getState();

      expect(state.autoStartSession).toBe(true);
      expect(state.apiUrl).toBe("https://custom.api.com");
      expect(state.collectAllTabs).toBe(false);
    });
  });

  describe("resetSettings", () => {
    it("should reset all settings to defaults", () => {
      // Change all settings
      useSettingsStore.getState().updateSettings({
        apiUrl: "https://custom.api.com",
        webUrl: "https://custom.web.com",
        autoStartSession: true,
        collectAllTabs: false,
      });

      // Reset
      useSettingsStore.getState().resetSettings();
      const state = useSettingsStore.getState();

      expect(state.apiUrl).toBe("http://localhost:9000/v1");
      expect(state.webUrl).toBe("http://localhost:3000");
      expect(state.autoStartSession).toBe(false);
      expect(state.collectAllTabs).toBe(true);
    });
  });
});
