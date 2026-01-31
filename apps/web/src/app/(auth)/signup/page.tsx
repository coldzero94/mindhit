"use client";

import Link from "next/link";
import { SignupForm } from "@/components/auth/signup-form";
import { PageTransition, FadeIn } from "@/components/ui/animations";

export default function SignupPage() {
  return (
    <PageTransition>
      <div className="flex min-h-screen items-center justify-center bg-background">
        <FadeIn delay={0.1}>
          <div className="space-y-4">
            <SignupForm />
            <p className="text-center text-sm text-muted-foreground">
              이미 계정이 있으신가요?{" "}
              <Link href="/login" className="text-primary hover:underline">
                로그인
              </Link>
            </p>
          </div>
        </FadeIn>
      </div>
    </PageTransition>
  );
}
