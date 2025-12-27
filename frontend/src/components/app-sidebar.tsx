import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import {
  LayoutDashboard,
  Server,
  Container,
  Activity,
  Network,
  Database,
  Users,
  Settings,
  FileText,
  Cloud,
} from "lucide-react";
import { Button } from "./ui/button";
import { Link, useNavigate } from "react-router";
import { useMutation, useQuery } from "@tanstack/react-query";
import { orpc, queryClient } from "@/lib/orpc";
import { toast } from "sonner";
import { useSession } from "@/stores/user";
import { useEffect, useState } from "react";
import { usePermission } from "./protected-content";
import {
  RBAC_QEMU_READ,
  RBAC_DOCKER_READ,
  RBAC_SETTINGS_MANAGER,
  RBAC_USER_ADMIN,
  RBAC_AUDIT_LOG_VIEWER,
} from "@/types/types.gen";

import { formatTimeDifference } from "@/lib/utils";
import { CONSTANTS } from "@/lib";

// Component for reactive time display
function RelativeTimeDisplay({
  timestamp,
}: {
  timestamp: number | Date | string | null;
}) {
  const [displayTime, setDisplayTime] = useState(() =>
    timestamp ? formatTimeDifference(timestamp) : "",
  );

  useEffect(() => {
    if (!timestamp) return;

    // Initial update
    setDisplayTime(formatTimeDifference(timestamp));

    // Update every second for more responsive display
    const interval = setInterval(() => {
      setDisplayTime(formatTimeDifference(timestamp));
    }, CONSTANTS.POLLING_INTERVAL_MS);

    return () => clearInterval(interval);
  }, [timestamp]);

  return <>{displayTime}</>;
}

// interface AppSidebarProps {
//   activeSection: string;
//   setActiveSection: (section: string) => void;
// }

export function AppSidebar() {
  const nav = useNavigate();
  const clearSession = useSession((s) => s.clearSession);
  const { hasPermission } = usePermission();
  const logoutMutation = useMutation(
    orpc.auth.logout.mutationOptions({
      onSuccess() {
        toast.success("Logged out successfully");
        clearSession();
        queryClient.clear(); // Clear all query cache
        nav("/auth/login");
      },
      onError() {
        // Even if logout fails on backend, clear local session
        clearSession();
        queryClient.clear();
        nav("/auth/login");
      },
    }),
  );
  const health = useQuery(
    orpc.health.check.queryOptions({
      staleTime: CONSTANTS.POLLING_INTERVAL_MS,
    }),
  );

  // Main menu items with their required permissions
  const menuItems = [
    {
      id: "dashboard",
      label: "Dashboard",
      icon: LayoutDashboard,
      requiredPermission: undefined, // Dashboard accessible to all authenticated users
    },
    {
      id: "vms",
      label: "Virtual Machines",
      icon: Server,
      requiredPermission: RBAC_QEMU_READ,
    },
    {
      id: "docker",
      label: "Containers",
      icon: Container,
      requiredPermission: RBAC_DOCKER_READ,
    },
    {
      id: "monitor",
      label: "Monitoring",
      icon: Activity,
      requiredPermission: RBAC_AUDIT_LOG_VIEWER,
    },
  ];

  // System items with their required permissions
  const systemItems = [
    {
      id: "networking",
      label: "Networking",
      icon: Network,
      requiredPermission: RBAC_SETTINGS_MANAGER,
    },
    {
      id: "storage",
      label: "Storage",
      icon: Database,
      requiredPermission: RBAC_SETTINGS_MANAGER,
    },
    {
      id: "users",
      label: "Users",
      icon: Users,
      requiredPermission: RBAC_USER_ADMIN,
    },
    {
      id: "logs",
      label: "Audit Logs",
      icon: FileText,
      requiredPermission: RBAC_AUDIT_LOG_VIEWER,
    },
    {
      id: "settings",
      label: "Settings",
      icon: Settings,
      requiredPermission: RBAC_SETTINGS_MANAGER,
    },
  ];

  // Filter menu items based on permissions
  const visibleMenuItems = menuItems.filter((item) =>
    item.requiredPermission ? hasPermission(item.requiredPermission) : true,
  );

  const visibleSystemItems = systemItems.filter((item) =>
    item.requiredPermission ? hasPermission(item.requiredPermission) : true,
  );

  return (
    <Sidebar>
      <SidebarHeader>
        <div className="px-2 py-2 flex flex-col items-start gap-1 text-lg">
          <div className="flex items-center gap-2">
            <Cloud className="w-5 h-5 text-primary" />
            Visory
          </div>
          <p className="text-xs text-muted-foreground">
            v{health.data?.app_version}
          </p>
        </div>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Main</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {visibleMenuItems.map((item) => (
                <SidebarMenuItem key={item.label}>
                  <SidebarMenuButton asChild>
                    <Link to={`/app/${item.id}`}>
                      <item.icon />
                      <span>{item.label}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        {visibleSystemItems.length > 0 && (
          <SidebarGroup>
            <SidebarGroupLabel>System</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {visibleSystemItems.map((item) => {
                  const Icon = item.icon;
                  return (
                    <SidebarMenuItem key={item.id}>
                      <SidebarMenuButton asChild>
                        <Link to={`/app/sys/${item.id}`}>
                          <Icon className="h-4 w-4" />
                          <span>{item.label}</span>
                        </Link>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  );
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        )}
      </SidebarContent>
      <SidebarFooter>
        <Button
          onClick={() => {
            logoutMutation.mutate({});
          }}
        >
          Logout
        </Button>
        <div className="px-2 py-2 text-xs text-muted-foreground">
          <div>System Status: {health.data?.message}</div>
          <div>
            Last Updated:{" "}
            <RelativeTimeDisplay timestamp={health.dataUpdatedAt} />
          </div>
        </div>
      </SidebarFooter>
    </Sidebar>
  );
}
