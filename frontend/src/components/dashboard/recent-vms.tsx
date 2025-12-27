import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useQuery, useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Skeleton } from "@/components/ui/skeleton";
import { Server, Play, Power, RotateCcw } from "lucide-react";
import { formatBytes } from "@/lib/utils";
import { toast } from "sonner";

// VM States from libvirt
const VM_STATE_RUNNING = 1;

function RecentVMs() {
  const { data: vms, isLoading } = useQuery(
    orpc.qemu.getVirtualMachinesInfo.queryOptions({})
  );

  const startMutation = useMutation(
    orpc.qemu.startVirtualMachine.mutationOptions({
      onSuccess: () => toast.success("VM started"),
    })
  );

  const shutdownMutation = useMutation(
    orpc.qemu.shutdownVirtualMachine.mutationOptions({
      onSuccess: () => toast.success("VM shutdown initiated"),
    })
  );

  const rebootMutation = useMutation(
    orpc.qemu.rebootVirtualMachine.mutationOptions({
      onSuccess: () => toast.success("VM reboot initiated"),
    })
  );

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Server className="h-5 w-5" />
            Virtual Machines
          </CardTitle>
          <CardDescription>QEMU/KVM virtual machines</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {[...Array(4)].map((_, i) => (
            <Skeleton key={i} className="h-16 w-full" />
          ))}
        </CardContent>
      </Card>
    );
  }

  const getStateLabel = (state: number) => {
    switch (state) {
      case 1:
        return "Running";
      case 2:
        return "Blocked";
      case 3:
        return "Paused";
      case 4:
        return "Shutdown";
      case 5:
        return "Shutoff";
      case 6:
        return "Crashed";
      case 7:
        return "Suspended";
      default:
        return "Unknown";
    }
  };

  const getStateColor = (state: number) => {
    switch (state) {
      case 1:
        return "bg-primary";
      case 3:
      case 7:
        return "bg-yellow-500";
      case 4:
      case 5:
        return "bg-muted-foreground";
      case 2:
      case 6:
        return "bg-destructive";
      default:
        return "bg-muted-foreground";
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Server className="h-5 w-5" />
          Virtual Machines
        </CardTitle>
        <CardDescription>{vms?.length ?? 0} VMs configured</CardDescription>
      </CardHeader>
      <CardContent>
        {vms && vms.length > 0 ? (
          <div className="space-y-3">
            {vms.slice(0, 5).map((vm) => (
              <div
                key={vm.uuid}
                className="flex items-center justify-between gap-2 rounded-lg border p-3 overflow-hidden"
              >
                <div className="flex items-center gap-2 min-w-0 flex-1">
                  <div
                    className={`h-2 w-2 rounded-full shrink-0 ${getStateColor(vm.state)}`}
                  />
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium truncate">{vm.name}</p>
                    <p className="text-xs text-muted-foreground truncate">
                      {vm.vcpus} vCPU, {formatBytes(vm.memory_kb * 1024)} RAM
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <Badge variant="outline" className="text-xs hidden sm:inline-flex">
                    {getStateLabel(vm.state)}
                  </Badge>
                  <div className="flex gap-1">
                    {vm.state !== VM_STATE_RUNNING && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7"
                        onClick={() =>
                          startMutation.mutate({
                            params: { uuid: vm.uuid },
                          })
                        }
                        disabled={startMutation.isPending}
                      >
                        <Play className="h-3 w-3" />
                      </Button>
                    )}
                    {vm.state === VM_STATE_RUNNING && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7"
                        onClick={() =>
                          shutdownMutation.mutate({
                            params: { uuid: vm.uuid },
                          })
                        }
                        disabled={shutdownMutation.isPending}
                      >
                        <Power className="h-3 w-3" />
                      </Button>
                    )}
                    {vm.state === VM_STATE_RUNNING && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7"
                        onClick={() =>
                          rebootMutation.mutate({
                            params: { uuid: vm.uuid },
                          })
                        }
                        disabled={rebootMutation.isPending}
                      >
                        <RotateCcw className="h-3 w-3" />
                      </Button>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            No virtual machines found
          </p>
        )}
      </CardContent>
    </Card>
  );
}

export default RecentVMs;
