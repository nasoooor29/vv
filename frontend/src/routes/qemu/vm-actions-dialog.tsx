import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { AlertCircle, Play, RotateCw, Power } from "lucide-react";
import type { T } from "@/types";

interface VMActionsDialogProps {
  vm: T.VirtualMachineWithInfo | null;
  action: "start" | "reboot" | "shutdown" | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isLoading?: boolean;
}

const ACTION_CONFIG = {
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

export default function VMActionsDialog({
  vm,
  action,
  open,
  onOpenChange,
  onConfirm,
  isLoading = false,
}: VMActionsDialogProps) {
  if (!vm || !action) return null;

  const config = ACTION_CONFIG[action];
  const Icon = config.icon;

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <div className="flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-amber-600" />
            <AlertDialogTitle>{config.title}</AlertDialogTitle>
          </div>
          <AlertDialogDescription className="space-y-2">
            <div>{config.description}</div>
            <div className="bg-muted rounded p-2 text-sm">
              <p className="font-semibold">{vm.name}</p>
              <p className="text-xs text-muted-foreground font-mono">
                {vm.uuid}
              </p>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <div className="flex gap-2 justify-end">
          <AlertDialogCancel disabled={isLoading}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            disabled={isLoading}
            className="gap-2"
          >
            <Icon className="h-4 w-4" />
            {config.action}
          </AlertDialogAction>
        </div>
      </AlertDialogContent>
    </AlertDialog>
  );
}
