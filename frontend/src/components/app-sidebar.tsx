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
  Shield,
  Users,
  Settings,
} from "lucide-react";
import { Button } from "./ui/button";
import { Link, useNavigate } from "react-router";
import { useMutation, useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { toast } from "sonner";
import { useSession } from "@/stores/user";
import { useEffect, useState } from "react";
import { formatTimeDifference } from "@/lib/utils";

// Component for reactive time display
function RelativeTimeDisplay({
  timestamp,
}: {
  timestamp: number | Date | null;
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
    }, 1000);

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
  const logoutMutation = useMutation(
    orpc.auth.logout.mutationOptions({
      onSuccess() {
        toast.success("Logged out successfully");
        clearSession();
        nav("/auth/login");
      },
    }),
  );
  const health = useQuery(
    orpc.health.check.queryOptions({
      staleTime: 1 * 1000,
    }),
  );
  const menuItems = [
    { id: "dashboard", label: "Dashboard", icon: LayoutDashboard },
    { id: "vms", label: "Virtual Machines", icon: Server },
    { id: "containers", label: "Containers", icon: Container },
    { id: "monitoring", label: "Monitoring", icon: Activity },
  ];

  const systemItems = [
    { id: "networking", label: "Networking", icon: Network },
    { id: "storage", label: "Storage", icon: Database },
    { id: "security", label: "Security", icon: Shield },
    { id: "users", label: "Users", icon: Users },
    { id: "settings", label: "Settings", icon: Settings },
  ];

  return (
    <Sidebar>
      <SidebarHeader>
        <div className="px-2 py-2">
          <h2 className="text-lg font-semibold">Visory</h2>
          <p className="text-sm text-muted-foreground">
            v{health.data?.app_version}
          </p>
        </div>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Main</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {menuItems.map((item) => (
                <SidebarMenuItem key={item.label}>
                  <SidebarMenuButton asChild>
                    <Link to={`/${item.id}`}>
                      <item.icon />
                      <span>{item.label}</span>
                    </Link>
                    {/* <a href={`/${item.id}`}> */}
                    {/* </a> */}
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
              {/* {menuItems.map((item) => { */}
              {/*   const Icon = item.icon; */}
              {/*   return ( */}
              {/*     <SidebarMenuItem key={item.id}> */}
              {/*       <SidebarMenuButton */}
              {/*       // onClick={() => setActiveSection(item.id)} */}
              {/*       // isActive={activeSection === item.id} */}
              {/*       > */}
              {/*         <Icon className="h-4 w-4" /> */}
              {/*         <span>{item.label}</span> */}
              {/*       </SidebarMenuButton> */}
              {/*     </SidebarMenuItem> */}
              {/*   ); */}
              {/* })} */}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        <SidebarGroup>
          <SidebarGroupLabel>System</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {systemItems.map((item) => {
                const Icon = item.icon;
                return (
                  <SidebarMenuItem key={item.id}>
                    <SidebarMenuButton
                    // onClick={() => setActiveSection(item.id)}
                    // isActive={activeSection === item.id}
                    >
                      <Icon className="h-4 w-4" />
                      <span>{item.label}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                );
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
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
