import { useMutation } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { AlertCircle, Play, RotateCw, Power } from "lucide-react";
import { orpc, queryClient } from "@/lib/orpc";
import { toast } from "sonner";
import type { T } from "@/types";

interface VMActionsDialogContentProps {
  vm: T.VirtualMachineWithInfo | null;
  action: "start" | "reboot" | "shutdown" | null;
  onClose: () => void;
}

export const ACTION_CONFIG = {
  start: {
    title: "Start Virtual Machine",
    description: "Are you sure you want to start this virtual machine?",
    icon: Play,
    action: "Start",
  },
  reboot: {
    title: "Reboot Virtual Machine",
    description: "Are you sure you want to reboot this virtual machine?",
    icon: RotateCw,
    action: "Reboot",
  },
  shutdown: {
    title: "Shutdown Virtual Machine",
    description:
      "Are you sure you want to shutdown this virtual machine? This will initiate a graceful shutdown.",
    icon: Power,
    action: "Shutdown",
  },
} as const;
export const getActionInfo = (action: string) => {
  if (action in ACTION_CONFIG) {
    return ACTION_CONFIG[action as keyof typeof ACTION_CONFIG];
  }
  return {
    title: "",
    description: "",
    icon: AlertCircle,
    action: "",
  };
};

export function VMActionsDialogContent({
  vm,
  action,
  onClose,
}: VMActionsDialogContentProps) {
  const startVMMutation = useMutation(
    orpc.qemu.startVirtualMachine.mutationOptions({
      onSuccess: () => {
        toast.success("Virtual machine started successfully");
        queryClient.invalidateQueries();
        onClose();
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
        onClose();
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
        onClose();
      },
      onError: () => {
        toast.error("Failed to shutdown virtual machine");
      },
    }),
  );

  const handleConfirmAction = () => {
    if (!vm || !action) return;

    if (action === "start") {
      startVMMutation.mutate({ params: { uuid: vm.uuid } });
    } else if (action === "reboot") {
      rebootVMMutation.mutate({ params: { uuid: vm.uuid } });
    } else if (action === "shutdown") {
      shutdownVMMutation.mutate({ params: { uuid: vm.uuid } });
    }
  };

  const isLoading =
    startVMMutation.isPending ||
    rebootVMMutation.isPending ||
    shutdownVMMutation.isPending;

  if (!vm || !action) return null;

  const config = ACTION_CONFIG[action];
  const Icon = config.icon;

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <AlertCircle className="h-5 w-5 text-amber-600" />
        <span className="font-semibold">{config.description}</span>
      </div>
      <div className="bg-muted rounded p-2 text-sm">
        <p className="font-semibold">{vm.name}</p>
        <p className="text-xs text-muted-foreground font-mono">{vm.uuid}</p>
      </div>
      <div className="flex gap-2 justify-end">
        <Button variant="outline" onClick={onClose} disabled={isLoading}>
          Cancel
        </Button>
        <Button
          onClick={handleConfirmAction}
          disabled={isLoading}
          className="gap-2"
        >
          <Icon className="h-4 w-4" />
          {config.action}
        </Button>
      </div>
    </div>
  );
}
