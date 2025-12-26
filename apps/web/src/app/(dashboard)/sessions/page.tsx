"use client";

import { useState } from "react";
import { SessionList } from "@/components/sessions/session-list";

export default function SessionsPage() {
  const [page, setPage] = useState(1);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">내 세션</h1>
      </div>
      <SessionList page={page} onPageChange={setPage} />
    </div>
  );
}
