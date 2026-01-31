"use client";

import { useState } from "react";
import { SessionList } from "@/components/sessions/session-list";
import { PageTransition, FadeIn } from "@/components/ui/animations";

export default function SessionsPage() {
  const [page, setPage] = useState(1);

  return (
    <PageTransition>
      <div className="space-y-6">
        <FadeIn>
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-bold text-foreground">내 세션</h1>
          </div>
        </FadeIn>
        <SessionList page={page} onPageChange={setPage} />
      </div>
    </PageTransition>
  );
}
