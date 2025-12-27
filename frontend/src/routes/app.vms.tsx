import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { orpc, queryClient } from "@/lib/orpc";
import { usePollingVMs } from "@/hooks";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { AlertCircle, Server, Play, RotateCw, Power, Info, Plus } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { usePermission } from "@/components/protected-content";
import {
  RBAC_QEMU_READ,
  RBAC_QEMU_WRITE,
  RBAC_QEMU_UPDATE,
} from "@/types/types.gen";
import { toast } from "sonner";
import VMDetailDialog from "./qemu/vm-detail-dialog";
import VMActionsDialog from "./qemu/vm-actions-dialog";
import { CreateVMDialog } from "./qemu/create-vm-dialog";
import type { T } from "@/types";

export default function QemuPage() {
  const { hasPermission } = usePermission();
  const [selectedVM, setSelectedVM] = useState<T.VirtualMachineWithInfo | null>(
    null,
  );
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);
  const [actionDialogOpen, setActionDialogOpen] = useState(false);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [selectedAction, setSelectedAction] = useState<
    "start" | "reboot" | "shutdown" | null
  >(null);

  const vmsQuery = usePollingVMs({ enabled: true, interval: 5000 });

  const startVMMutation = useMutation(
    orpc.qemu.startVirtualMachine.mutationOptions({
      onSuccess: () => {
        toast.success("Virtual machine started successfully");
        queryClient.invalidateQueries();
        setActionDialogOpen(false);
      },
      onError: () => {
        toast.error("Failed to start virtual machine");
      },
    }),
  );

  const rebootVMMutation = useMutation(
    orpc.qemu.rebootVirtualMachine.mutationOptions({
      onSuccess: () => {
        toast.success("Virtual machine rebooted successfully");
        queryClient.invalidateQueries();
        setActionDialogOpen(false);
      },
      onError: () => {
        toast.error("Failed to reboot virtual machine");
      },
    }),
  );

  const shutdownVMMutation = useMutation(
    orpc.qemu.shutdownVirtualMachine.mutationOptions({
      onSuccess: () => {
        toast.success("Virtual machine shutdown initiated");
        queryClient.invalidateQueries();
        setActionDialogOpen(false);
      },
      onError: () => {
        toast.error("Failed to shutdown virtual machine");
      },
    }),
  );

  const handleStartVM = (vm: T.VirtualMachineWithInfo) => {
    setSelectedVM(vm);
    setSelectedAction("start");
    setActionDialogOpen(true);
  };

  const handleRebootVM = (vm: T.VirtualMachineWithInfo) => {
    setSelectedVM(vm);
    setSelectedAction("reboot");
    setActionDialogOpen(true);
  };

  const handleShutdownVM = (vm: T.VirtualMachineWithInfo) => {
    setSelectedVM(vm);
    setSelectedAction("shutdown");
    setActionDialogOpen(true);
  };

  const handleViewDetails = (vm: T.VirtualMachineWithInfo) => {
    setSelectedVM(vm);
    setDetailDialogOpen(true);
  };

  const handleConfirmAction = () => {
    if (!selectedVM || !selectedAction) return;

    if (selectedAction === "start") {
      startVMMutation.mutate({ params: { uuid: selectedVM.uuid } });
    } else if (selectedAction === "reboot") {
      rebootVMMutation.mutate({ params: { uuid: selectedVM.uuid } });
    } else if (selectedAction === "shutdown") {
      shutdownVMMutation.mutate({ params: { uuid: selectedVM.uuid } });
    }
  };

  const getVMStatus = (state: number) => {
    const states: Record<number, { label: string; variant: any }> = {
      0: { label: "No State", variant: "secondary" },
      1: { label: "Running", variant: "success" },
      2: { label: "Blocked", variant: "warning" },
      3: { label: "Paused", variant: "warning" },
      4: { label: "Shutting Down", variant: "warning" },
      5: { label: "Shut Off", variant: "destructive" },
      6: { label: "Crashed", variant: "destructive" },
      7: { label: "Suspended", variant: "secondary" },
    };
    return states[state] || { label: "Unknown", variant: "secondary" };
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
  };

  if (!hasPermission(RBAC_QEMU_READ)) {
    return (
      <Alert className="border-destructive bg-destructive/10">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          You don't have permission to access QEMU virtual machines.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Server className="h-6 w-6" />
          <h1 className="text-3xl font-bold">Virtual Machines</h1>
        </div>
        {hasPermission(RBAC_QEMU_WRITE) && (
          <Button onClick={() => setCreateDialogOpen(true)} className="gap-2">
            <Plus className="h-4 w-4" />
            Create VM
          </Button>
        )}
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Server className="h-5 w-5" />
            QEMU Virtual Machines
          </CardTitle>
          <div className="text-sm text-muted-foreground">
            {vmsQuery.data?.length || 0} VMs
          </div>
        </CardHeader>
        <CardContent>
          {vmsQuery.isLoading ? (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-32 w-full" />
              ))}
            </div>
          ) : vmsQuery.isError ? (
            <Alert className="border-destructive bg-destructive/10">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Failed to load virtual machines. Please try again later.
              </AlertDescription>
            </Alert>
          ) : !vmsQuery.data || vmsQuery.data.length === 0 ? (
            <div className="py-12 text-center text-muted-foreground">
              <Server className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No virtual machines found</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              {vmsQuery.data.map((vm) => {
                const status = getVMStatus(vm.state);
                const isRunning = vm.state === 1;

                return (
                  <div
                    key={vm.uuid}
                    className="border border-border rounded-lg p-4 space-y-3 hover:bg-accent/50 transition-colors"
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <h3 className="font-semibold text-lg">{vm.name}</h3>
                        <p className="text-xs text-muted-foreground font-mono">
                          {vm.uuid}
                        </p>
                      </div>
                      <Badge variant={status.variant as any}>
                        {status.label}
                      </Badge>
                    </div>

                    <div className="grid grid-cols-3 gap-2 text-sm">
                      <div className="space-y-1">
                        <p className="text-xs text-muted-foreground">Memory</p>
                        <p className="font-semibold">
                          {formatBytes(vm.memory_kb * 1024)}
                        </p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-xs text-muted-foreground">Max Memory</p>
                        <p className="font-semibold">
                          {formatBytes(vm.max_mem_kb * 1024)}
                        </p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-xs text-muted-foreground">vCPUs</p>
                        <p className="font-semibold">{vm.vcpus}</p>
                      </div>
                    </div>

                    <div className="space-y-1 text-xs">
                      <p className="text-muted-foreground">
                        CPU Time: {(vm.cpu_time_ns / 1e9).toFixed(2)}s
                      </p>
                    </div>

                    <div className="flex gap-2 pt-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleViewDetails(vm)}
                        className="gap-2 flex-1"
                      >
                        <Info className="h-4 w-4" />
                        Details
                      </Button>

                      {hasPermission(RBAC_QEMU_WRITE) && !isRunning && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleStartVM(vm)}
                          className="gap-2"
                        >
                          <Play className="h-4 w-4" />
                          Start
                        </Button>
                      )}

                      {hasPermission(RBAC_QEMU_UPDATE) && isRunning && (
                        <>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleRebootVM(vm)}
                            className="gap-2"
                          >
                            <RotateCw className="h-4 w-4" />
                            Reboot
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleShutdownVM(vm)}
                            className="gap-2"
                          >
                            <Power className="h-4 w-4" />
                            Shutdown
                          </Button>
                        </>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      <VMDetailDialog
        vm={selectedVM}
        open={detailDialogOpen}
        onOpenChange={setDetailDialogOpen}
      />

       <VMActionsDialog
        vm={selectedVM}
        action={selectedAction}
        open={actionDialogOpen}
        onOpenChange={setActionDialogOpen}
        onConfirm={handleConfirmAction}
        isLoading={
          startVMMutation.isPending ||
          rebootVMMutation.isPending ||
          shutdownVMMutation.isPending
        }
      />

      <CreateVMDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onSuccess={() => {
          queryClient.invalidateQueries();
        }}
      />
    </div>
  );
}
