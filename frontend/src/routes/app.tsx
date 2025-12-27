import { AppSidebar } from "@/components/app-sidebar";
import { MobileBottomNav } from "@/components/mobile-bottom-nav";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { CONSTANTS } from "@/lib";
import { useIsMobile } from "@/hooks/use-mobile";
import { orpc } from "@/lib/orpc";
import { ORPCError } from "@orpc/client";
import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";
import { Outlet, useNavigate } from "react-router";

function AppLayout() {
  const isMobile = useIsMobile();
  const data = useQuery(
    orpc.auth.me.queryOptions({
      staleTime: CONSTANTS.POLLING_INTERVAL_MS, // 1 second
      retry: 1,
    }),
  );
  const navigate = useNavigate();

  useEffect(() => {
    if (!data.isLoading && !data.data) {
      navigate("/auth/login", { replace: true });
    }
  }, [data.data, data.isLoading, navigate]);

  if (data.error && data.error instanceof ORPCError) {
    if (data.error.status !== 401) {
      // User is not authenticated, stay on auth layout
      console.error("Error fetching session:", data.error);
    }
  }

  // Mobile layout: no sidebar, use bottom navigation
  if (isMobile) {
    return (
      <div className="flex min-h-svh flex-col bg-background">
        <header className="sticky top-0 z-40 flex h-12 shrink-0 items-center gap-2 border-b bg-background px-4">
          <h1 className="text-lg font-semibold">Visory</h1>
        </header>
        <main className="flex flex-1 flex-col gap-4 p-4 pb-20">
          <Outlet />
        </main>
        <MobileBottomNav />
      </div>
    );
  }

  // Desktop layout: sidebar
  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-12 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger />
          <div className="flex w-full items-center justify-between">
            <div className="flex gap-2">
              {/* <DeployStack /> */}
              {/* <RunContainer /> */}
              {/* <AddVm /> */}
            </div>
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4">
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

export default AppLayout;
