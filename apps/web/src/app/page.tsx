"use client";

import { useEffect, useSyncExternalStore } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";
import { LandingNavbar } from "@/components/landing/LandingNavbar";
import { HeroSection } from "@/components/landing/HeroSection";
import { ProblemSection } from "@/components/landing/ProblemSection";
import { SolutionSection } from "@/components/landing/SolutionSection";
import { FeaturesSection } from "@/components/landing/FeaturesSection";
import { MindmapDemoSection } from "@/components/landing/MindmapDemoSection";
import { PricingSection } from "@/components/landing/PricingSection";

export default function Home() {
  const router = useRouter();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  // Detect client-side using useSyncExternalStore
  const isClient = useSyncExternalStore(
    () => () => {}, // No-op subscribe (no external state changes)
    () => true,      // Client snapshot
    () => false      // Server snapshot
  );

  // Redirect authenticated users to dashboard
  useEffect(() => {
    if (isClient && isAuthenticated) {
      router.replace("/sessions");
    }
  }, [isClient, isAuthenticated, router]);

  // Show loading during hydration or redirect
  if (!isClient || isAuthenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="animate-pulse text-gray-500">로딩 중...</div>
      </div>
    );
  }

  // Show landing page for unauthenticated users
  return (
    <main className="relative">
      <LandingNavbar />
      <HeroSection />
      <ProblemSection />
      <SolutionSection />
      <FeaturesSection />
      <MindmapDemoSection />
      <PricingSection />
    </main>
  );
}
