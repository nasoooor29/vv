import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Server } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Skeleton } from "@/components/ui/skeleton";

// VM States from libvirt
const VM_STATE_RUNNING = 1;
const VM_STATE_BLOCKED = 2;
const VM_STATE_PAUSED = 3;
const VM_STATE_SHUTDOWN = 4;
const VM_STATE_SHUTOFF = 5;
const VM_STATE_CRASHED = 6;
const VM_STATE_PMSUSPENDED = 7;

function VMSummary() {
  const { data: vms, isLoading } = useQuery(
    orpc.qemu.getVirtualMachinesInfo.queryOptions({})
  );

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Virtual Machines</CardTitle>
          <Server className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-8 w-16 mb-2" />
          <Skeleton className="h-5 w-32" />
        </CardContent>
      </Card>
    );
  }

  const stats = {
    total: vms?.length ?? 0,
    running: vms?.filter((v) => v.state === VM_STATE_RUNNING).length ?? 0,
    stopped:
      vms?.filter(
        (v) => v.state === VM_STATE_SHUTOFF || v.state === VM_STATE_SHUTDOWN
      ).length ?? 0,
    paused:
      vms?.filter(
        (v) => v.state === VM_STATE_PAUSED || v.state === VM_STATE_PMSUSPENDED
      ).length ?? 0,
    error:
      vms?.filter(
        (v) => v.state === VM_STATE_CRASHED || v.state === VM_STATE_BLOCKED
      ).length ?? 0,
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">Virtual Machines</CardTitle>
        <Server className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{stats.total}</div>
        <div className="mt-2 flex flex-wrap gap-1">
          {stats.running > 0 && (
            <Badge variant="default" className="bg-primary">
              {stats.running} Running
            </Badge>
          )}
          {stats.stopped > 0 && (
            <Badge variant="secondary">{stats.stopped} Stopped</Badge>
          )}
          {stats.paused > 0 && (
            <Badge variant="outline">{stats.paused} Paused</Badge>
          )}
          {stats.error > 0 && (
            <Badge variant="destructive">{stats.error} Error</Badge>
          )}
          {stats.total === 0 && (
            <span className="text-xs text-muted-foreground">No VMs</span>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

export default VMSummary;
