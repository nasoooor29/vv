import { Link, useLocation } from "react-router";
import {
  LayoutDashboard,
  Server,
  Container,
  Activity,
  Menu,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { usePermission } from "./protected-content";
import {
  RBAC_QEMU_READ,
  RBAC_DOCKER_READ,
  RBAC_EVENT_VIEWER,
} from "@/types/types.gen";
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "./ui/drawer";
import { Button } from "./ui/button";
import { useMutation, useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { toast } from "sonner";
import { useSession } from "@/stores/user";
import { useNavigate } from "react-router";
import { Network, Database, Users, FileText, Settings } from "lucide-react";
import {
  RBAC_SETTINGS_MANAGER,
  RBAC_USER_ADMIN,
  RBAC_AUDIT_LOG_VIEWER,
} from "@/types/types.gen";
import { useState } from "react";

export function MobileBottomNav() {
  const location = useLocation();
  const { hasPermission } = usePermission();
  const [menuOpen, setMenuOpen] = useState(false);

  // Main navigation items (shown in bottom bar)
  const mainNavItems = [
    {
      id: "dashboard",
      path: "/dashboard",
      label: "Dashboard",
      icon: LayoutDashboard,
      requiredPermission: undefined,
    },
    {
      id: "vms",
      path: "/vms",
      label: "VMs",
      icon: Server,
      requiredPermission: RBAC_QEMU_READ,
    },
    {
      id: "containers",
      path: "/containers",
      label: "Containers",
      icon: Container,
      requiredPermission: RBAC_DOCKER_READ,
    },
    {
      id: "monitor",
      path: "/monitor",
      label: "Monitor",
      icon: Activity,
      requiredPermission: RBAC_EVENT_VIEWER,
    },
  ];

  // Filter items based on permissions
  const visibleNavItems = mainNavItems.filter((item) =>
    item.requiredPermission ? hasPermission(item.requiredPermission) : true,
  );

  // Take only first 4 items for bottom nav, rest goes to "More" menu
  const bottomNavItems = visibleNavItems.slice(0, 3);

  const isActive = (path: string) => {
    return location.pathname.startsWith(path);
  };

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 border-t bg-background md:hidden">
      <div className="flex h-16 items-center justify-around px-2">
        {bottomNavItems.map((item) => {
          const Icon = item.icon;
          const active = isActive(item.path);
          return (
            <Link
              key={item.id}
              to={item.path}
              className={cn(
                "flex flex-1 flex-col items-center justify-center gap-1 py-2 text-xs transition-colors",
                active
                  ? "text-primary"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              <Icon className={cn("h-5 w-5", active && "text-primary")} />
              <span className="truncate">{item.label}</span>
            </Link>
          );
        })}

        {/* More menu for system items */}
        <MobileMoreMenu open={menuOpen} onOpenChange={setMenuOpen} />
      </div>
    </nav>
  );
}

function MobileMoreMenu({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const location = useLocation();
  const nav = useNavigate();
  const { hasPermission } = usePermission();
  const clearSession = useSession((s) => s.clearSession);

  const health = useQuery(
    orpc.health.check.queryOptions({
      staleTime: 1 * 1000,
    }),
  );

  const logoutMutation = useMutation(
    orpc.auth.logout.mutationOptions({
      onSuccess() {
        toast.success("Logged out successfully");
        clearSession();
        nav("/auth/login");
      },
    }),
  );

  // System menu items
  const systemItems = [
    {
      id: "networking",
      label: "Networking",
      icon: Network,
      path: "/app/sys/networking",
      requiredPermission: RBAC_SETTINGS_MANAGER,
    },
    {
      id: "storage",
      label: "Storage",
      icon: Database,
      path: "/app/sys/storage",
      requiredPermission: RBAC_SETTINGS_MANAGER,
    },
    {
      id: "users",
      label: "Users",
      icon: Users,
      path: "/app/sys/users",
      requiredPermission: RBAC_USER_ADMIN,
    },
    {
      id: "logs",
      label: "Audit Logs",
      icon: FileText,
      path: "/app/sys/logs",
      requiredPermission: RBAC_AUDIT_LOG_VIEWER,
    },
    {
      id: "settings",
      label: "Settings",
      icon: Settings,
      path: "/app/sys/settings",
      requiredPermission: RBAC_SETTINGS_MANAGER,
    },
  ];

  const visibleSystemItems = systemItems.filter((item) =>
    item.requiredPermission ? hasPermission(item.requiredPermission) : true,
  );

  const isMenuActive = systemItems.some((item) =>
    location.pathname.startsWith(item.path),
  );

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <button
          className={cn(
            "flex flex-1 flex-col items-center justify-center gap-1 py-2 text-xs transition-colors",
            isMenuActive
              ? "text-primary"
              : "text-muted-foreground hover:text-foreground",
          )}
        >
          <Menu className={cn("h-5 w-5", isMenuActive && "text-primary")} />
          <span>More</span>
        </button>
      </DrawerTrigger>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle className="flex items-center justify-between">
            <span>Visory</span>
            <span className="text-sm font-normal text-muted-foreground">
              v{health.data?.app_version}
            </span>
          </DrawerTitle>
        </DrawerHeader>

        <div className="p-4 space-y-4">
          {visibleSystemItems.length > 0 && (
            <div className="space-y-2">
              <h3 className="text-sm font-medium text-muted-foreground">
                System
              </h3>
              <div className="grid grid-cols-3 gap-2">
                {visibleSystemItems.map((item) => {
                  const Icon = item.icon;
                  const active = location.pathname.startsWith(item.path);
                  return (
                    <Link
                      key={item.id}
                      to={item.path}
                      onClick={() => onOpenChange(false)}
                      className={cn(
                        "flex flex-col items-center gap-2 rounded-lg border p-3 transition-colors",
                        active
                          ? "border-primary bg-primary/10 text-primary"
                          : "border-border hover:bg-accent",
                      )}
                    >
                      <Icon className="h-5 w-5" />
                      <span className="text-xs">{item.label}</span>
                    </Link>
                  );
                })}
              </div>
            </div>
          )}

          <div className="space-y-2 border-t pt-4">
            <Button
              variant="destructive"
              className="w-full"
              onClick={() => {
                logoutMutation.mutate({});
              }}
            >
              Logout
            </Button>
            <div className="text-center text-xs text-muted-foreground">
              System Status: {health.data?.message}
            </div>
          </div>
        </div>
      </DrawerContent>
    </Drawer>
  );
}
