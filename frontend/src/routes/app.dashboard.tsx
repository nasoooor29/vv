import {
  StorageSummary,
  ContainerSummary,
  VMSummary,
  SystemHealth,
  LogActivity,
  RecentContainers,
  RecentVMs,
  SystemResources,
} from "@/components/dashboard";
import { usePermission } from "@/components/protected-content";
import {
  RBAC_DOCKER_READ,
  RBAC_QEMU_READ,
  RBAC_SETTINGS_MANAGER,
  RBAC_AUDIT_LOG_VIEWER,
  RBAC_HEALTH_CHECKER,
} from "@/types/types.gen";

function Dashboard() {
  const { hasPermission } = usePermission();

  return (
    <div className="space-y-6">
      {/* Overview Cards */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {hasPermission(RBAC_DOCKER_READ) && <ContainerSummary />}
        {hasPermission(RBAC_QEMU_READ) && <VMSummary />}
        {hasPermission(RBAC_SETTINGS_MANAGER) && <StorageSummary />}
      </div>

      {/* Recent Activity */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {hasPermission(RBAC_DOCKER_READ) && <RecentContainers />}
        {hasPermission(RBAC_QEMU_READ) && <RecentVMs />}
      </div>

      {/* System Health & Log Activity */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {hasPermission(RBAC_HEALTH_CHECKER) && <SystemHealth />}
        {hasPermission(RBAC_AUDIT_LOG_VIEWER) && <LogActivity />}
      </div>

      {/* System Resources */}
      <SystemResources />
    </div>
  );
}

export default Dashboard;
