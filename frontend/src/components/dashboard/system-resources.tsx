import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Skeleton } from "@/components/ui/skeleton";
import { HardDrive, Server, Container, ScrollText } from "lucide-react";
import { formatBytes } from "@/lib/utils";

function SystemResources() {
  // Get storage info
  const { data: mountPoints, isLoading: storageLoading } = useQuery(
    orpc.storage.mountPoints.queryOptions({})
  );

  // Get containers
  const { data: clients } = useQuery(orpc.docker.clients.queryOptions({}));
  const clientId = clients?.[0]?.id?.toString() ?? "1";
  const { data: containers, isLoading: containersLoading } = useQuery(
    orpc.docker.containers.queryOptions({
      input: { params: { clientId } },
    })
  );

  // Get VMs
  const { data: vms, isLoading: vmsLoading } = useQuery(
    orpc.qemu.getVirtualMachinesInfo.queryOptions({})
  );

  // Get logs stats
  const { data: logStats, isLoading: logsLoading } = useQuery(
    orpc.logs.getLogStats.queryOptions({
      input: { days: 7 },
    })
  );

  const isLoading = storageLoading || containersLoading || vmsLoading || logsLoading;

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>System Resources</CardTitle>
          <CardDescription>Real-time resource utilization</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="space-y-2">
                <Skeleton className="h-4 w-20" />
                <Skeleton className="h-2 w-full" />
                <Skeleton className="h-4 w-32" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  // Calculate storage usage
  const totalStorage = mountPoints?.reduce((acc, m) => acc + m.total, 0) ?? 0;
  const usedStorage = mountPoints?.reduce((acc, m) => acc + m.used, 0) ?? 0;
  const storagePercent = totalStorage > 0 ? Math.round((usedStorage / totalStorage) * 100) : 0;

  // Calculate container stats
  const runningContainers = containers?.filter((c) => c.State === "running").length ?? 0;
  const totalContainers = containers?.length ?? 0;
  const containerPercent = totalContainers > 0 ? Math.round((runningContainers / totalContainers) * 100) : 0;

  // Calculate VM stats
  const runningVMs = vms?.filter((v) => v.state === 1).length ?? 0;
  const totalVMs = vms?.length ?? 0;
  const vmPercent = totalVMs > 0 ? Math.round((runningVMs / totalVMs) * 100) : 0;

  // Calculate total VM memory
  const totalVMMemory = vms?.reduce((acc, v) => acc + (v.state === 1 ? v.memory_kb : 0), 0) ?? 0;

  return (
    <Card>
      <CardHeader>
        <CardTitle>System Resources</CardTitle>
        <CardDescription>Real-time resource utilization</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <HardDrive className="h-4 w-4" />
              <span className="text-sm font-medium">Storage</span>
            </div>
            <Progress value={storagePercent} />
            <div className="text-sm text-muted-foreground">
              {formatBytes(usedStorage)} / {formatBytes(totalStorage)}
            </div>
          </div>
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Container className="h-4 w-4" />
              <span className="text-sm font-medium">Containers</span>
            </div>
            <Progress value={containerPercent} />
            <div className="text-sm text-muted-foreground">
              {runningContainers} / {totalContainers} running
            </div>
          </div>
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Server className="h-4 w-4" />
              <span className="text-sm font-medium">Virtual Machines</span>
            </div>
            <Progress value={vmPercent} />
            <div className="text-sm text-muted-foreground">
              {runningVMs} / {totalVMs} running
              {totalVMMemory > 0 && ` (${formatBytes(totalVMMemory * 1024)} RAM)`}
            </div>
          </div>
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <ScrollText className="h-4 w-4" />
              <span className="text-sm font-medium">Logs (7 days)</span>
            </div>
            <Progress value={100} />
            <div className="text-sm text-muted-foreground">
              {logStats?.total ?? 0} total entries
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default SystemResources;
