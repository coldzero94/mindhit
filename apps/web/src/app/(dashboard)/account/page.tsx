"use client";

import { User, Settings, LogOut } from "lucide-react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";
import { SubscriptionCard } from "@/components/account/SubscriptionCard";
import { UsageCard } from "@/components/account/UsageCard";
import { UsageHistory } from "@/components/account/UsageHistory";
import { Button } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";

export default function AccountPage() {
  const router = useRouter();
  const { user, logout } = useAuthStore();

  const handleLogout = () => {
    logout();
    router.push("/login");
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">계정</h1>
        <p className="text-gray-500 mt-1">계정 정보와 사용량을 확인하세요</p>
      </div>

      {/* User Info */}
      <div className="p-6 bg-white rounded-xl border border-gray-200">
        <div className="flex items-center gap-4">
          <div
            className="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600
                        rounded-full flex items-center justify-center"
          >
            <User className="w-8 h-8 text-white" />
          </div>
          <div className="flex-1">
            <h2 className="text-xl font-semibold text-gray-900">
              {user?.email?.split("@")[0] || "사용자"}
            </h2>
            <p className="text-gray-500">{user?.email}</p>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" className="text-gray-600">
              <Settings className="w-4 h-4 mr-2" />
              설정
            </Button>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="ghost" size="sm" className="text-red-600">
                  <LogOut className="w-4 h-4 mr-2" />
                  로그아웃
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>로그아웃 하시겠습니까?</AlertDialogTitle>
                  <AlertDialogDescription>
                    로그아웃하면 다시 로그인해야 합니다.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>취소</AlertDialogCancel>
                  <AlertDialogAction onClick={handleLogout}>
                    로그아웃
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </div>

      {/* Subscription & Usage */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <SubscriptionCard />
        <UsageCard />
      </div>

      {/* Usage History */}
      <UsageHistory />
    </div>
  );
}
