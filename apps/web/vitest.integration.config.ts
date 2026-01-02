import { defineConfig } from "vitest/config";
import { resolve } from "path";

/**
 * Vitest configuration for integration tests
 * - No MSW mocking
 * - Real backend API calls
 *
 * Run: pnpm test:integration
 * Requires: Backend server running (moonx backend:dev-api)
 */
export default defineConfig({
  test: {
    environment: "node", // Node environment for real HTTP calls
    globals: true,
    include: ["src/test/integration/**/*.integration.test.ts"],
    testTimeout: 30000, // Longer timeout for real API calls
    hookTimeout: 30000,
    // Run tests sequentially to avoid rate limiting
    fileParallelism: false,
    sequence: {
      shuffle: false,
    },
  },
  resolve: {
    alias: {
      "@": resolve(__dirname, "./src"),
    },
  },
});
