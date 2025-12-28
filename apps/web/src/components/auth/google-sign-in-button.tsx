"use client";

import { useEffect, useCallback, useRef } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

import { useAuthStore } from "@/stores/auth-store";
import { authApi } from "@/lib/api/auth";

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (config: GoogleInitConfig) => void;
          renderButton: (
            element: HTMLElement,
            options: GoogleButtonOptions
          ) => void;
          prompt: () => void;
        };
      };
    };
  }
}

interface GoogleInitConfig {
  client_id: string;
  callback: (response: { credential: string }) => void;
  auto_select?: boolean;
}

interface GoogleButtonOptions {
  theme?: "outline" | "filled_blue" | "filled_black";
  size?: "large" | "medium" | "small";
  text?: "signin_with" | "signup_with" | "continue_with";
  shape?: "rectangular" | "pill" | "circle" | "square";
  width?: number;
}

interface GoogleSignInButtonProps {
  text?: "signin_with" | "signup_with" | "continue_with";
}

export function GoogleSignInButton({
  text = "signin_with",
}: GoogleSignInButtonProps) {
  const router = useRouter();
  const setAuth = useAuthStore((state) => state.setAuth);
  const buttonRef = useRef<HTMLDivElement>(null);
  const isInitialized = useRef(false);

  const handleGoogleCallback = useCallback(
    async (response: { credential: string }) => {
      try {
        const result = await authApi.googleAuth({ credential: response.credential });
        setAuth(
          { id: result.user.id, email: result.user.email },
          result.token,
          result.token
        );
        toast.success("로그인 성공", { description: "Google 계정으로 로그인되었습니다." });
        router.push("/sessions");
      } catch {
        toast.error("Google 로그인 실패", {
          description: "다시 시도해주세요.",
        });
      }
    },
    [router, setAuth]
  );

  useEffect(() => {
    const clientId = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID;

    if (!clientId) {
      console.warn("NEXT_PUBLIC_GOOGLE_CLIENT_ID is not set");
      return;
    }

    // Check if script is already loaded
    const existingScript = document.querySelector(
      'script[src="https://accounts.google.com/gsi/client"]'
    );

    const initializeGoogle = () => {
      if (window.google && buttonRef.current && !isInitialized.current) {
        isInitialized.current = true;
        window.google.accounts.id.initialize({
          client_id: clientId,
          callback: handleGoogleCallback,
          auto_select: false,
        });

        window.google.accounts.id.renderButton(buttonRef.current, {
          theme: "outline",
          size: "large",
          text: text,
          shape: "rectangular",
          width: 280,
        });
      }
    };

    if (existingScript) {
      // Script already exists, just initialize
      initializeGoogle();
      return;
    }

    // Load script
    const script = document.createElement("script");
    script.src = "https://accounts.google.com/gsi/client";
    script.async = true;
    script.defer = true;
    script.onload = initializeGoogle;
    document.head.appendChild(script);

    return () => {
      isInitialized.current = false;
    };
  }, [handleGoogleCallback, text]);

  return <div ref={buttonRef} className="flex justify-center" />;
}
