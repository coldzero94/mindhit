"use client";

import { Suspense, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";

function ExtensionCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const setAuth = useAuthStore((state) => state.setAuth);

  useEffect(() => {
    const token = searchParams.get("token");
    const userParam = searchParams.get("user");
    const redirectPath = searchParams.get("redirect") || "/sessions";

    if (token && userParam) {
      try {
        // Decode user data from base64
        const user = JSON.parse(atob(userParam)) as { id: string; email: string };

        // Store auth in localStorage (Zustand persist)
        // Web app uses accessToken + refreshToken, but we only have one token from extension
        // Use the same token for both (API will handle refresh if needed)
        setAuth(user, token, token);

        // Small delay to ensure storage is written before redirect
        setTimeout(() => {
          router.replace(redirectPath);
        }, 100);
      } catch (error) {
        console.error("Failed to parse auth data:", error);
        router.replace("/login?error=invalid_token");
      }
    } else {
      // No token provided, redirect to login
      router.replace("/login");
    }
  }, [searchParams, setAuth, router]);

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto" />
        <p className="mt-4 text-gray-600">Logging in from extension...</p>
      </div>
    </div>
  );
}

/**
 * Extension callback page.
 * Receives token from Chrome extension and stores it in localStorage.
 * URL format: /auth/extension-callback?token=xxx&user=base64_encoded_user&redirect=/path
 */
export default function ExtensionCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto" />
            <p className="mt-4 text-gray-600">Loading...</p>
          </div>
        </div>
      }
    >
      <ExtensionCallbackContent />
    </Suspense>
  );
}
