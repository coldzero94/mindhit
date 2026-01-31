"use client";

import { User, Settings, LogOut } from "lucide-react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";
import { SubscriptionCard } from "@/components/account/SubscriptionCard";
import { UsageCard } from "@/components/account/UsageCard";
import { UsageHistory } from "@/components/account/UsageHistory";
import { Button } from "@/components/ui/button";
import { PageTransition, FadeIn } from "@/components/ui/animations";
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
    <PageTransition>
      <div className="space-y-6">
        {/* Header */}
        <FadeIn>
          <div>
            <h1 className="text-2xl font-bold text-foreground">계정</h1>
            <p className="text-muted-foreground mt-1">계정 정보와 사용량을 확인하세요</p>
          </div>
        </FadeIn>

        {/* User Info */}
        <FadeIn delay={0.1}>
          <div className="p-6 bg-card rounded-xl border border-border">
            <div className="flex items-center gap-4">
              <div
                className="w-16 h-16 bg-gradient-to-br from-primary to-primary/60
                          rounded-full flex items-center justify-center"
              >
                <User className="w-8 h-8 text-primary-foreground" />
              </div>
              <div className="flex-1">
                <h2 className="text-xl font-semibold text-foreground">
                  {user?.email?.split("@")[0] || "사용자"}
                </h2>
                <p className="text-muted-foreground">{user?.email}</p>
              </div>
              <div className="flex items-center gap-2">
                <Button variant="ghost" size="sm" className="text-muted-foreground">
                  <Settings className="w-4 h-4 mr-2" />
                  설정
                </Button>
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button variant="ghost" size="sm" className="text-status-error">
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
        </FadeIn>

        {/* Subscription & Usage */}
        <FadeIn delay={0.2}>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <SubscriptionCard />
            <UsageCard />
          </div>
        </FadeIn>

        {/* Usage History */}
        <FadeIn delay={0.3}>
          <UsageHistory />
        </FadeIn>
      </div>
    </PageTransition>
  );
}
