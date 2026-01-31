"use client";

import Link from "next/link";
import { LoginForm } from "@/components/auth/login-form";
import { PageTransition, FadeIn } from "@/components/ui/animations";

export default function LoginPage() {
  return (
    <PageTransition>
      <div className="flex min-h-screen items-center justify-center bg-background">
        <FadeIn delay={0.1}>
          <div className="space-y-4">
            <LoginForm />
            <p className="text-center text-sm text-muted-foreground">
              계정이 없으신가요?{" "}
              <Link href="/signup" className="text-primary hover:underline">
                회원가입
              </Link>
            </p>
          </div>
        </FadeIn>
      </div>
    </PageTransition>
  );
}
