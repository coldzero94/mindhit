import type { NextConfig } from "next";
import { config } from "dotenv";
import { resolve } from "path";

// Load environment variables from root .env
config({ path: resolve(__dirname, "../../.env") });

const nextConfig: NextConfig = {
  reactStrictMode: true,
  transpilePackages: ["three"],
  // Turbopack config (Next.js 16+ default bundler)
  turbopack: {
    // Monorepo: set root to workspace root for proper module resolution
    root: resolve(__dirname, "../.."),
  },
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${process.env.NEXT_PUBLIC_API_URL}/v1/:path*`,
      },
    ];
  },
};

export default nextConfig;
