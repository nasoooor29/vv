import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Monitor } from "lucide-react";
import type { T } from "@/types";

interface VMDetailDialogProps {
  vm: T.VirtualMachineWithInfo | null;
  onOpenConsole?: (uuid: string) => void;
}

const VM_STATES = {
  0: { label: "No State", variant: "secondary" },
  1: { label: "Running", variant: "success" },
  2: { label: "Blocked", variant: "warning" },
  3: { label: "Paused", variant: "warning" },
  4: { label: "Shutting Down", variant: "warning" },
  5: { label: "Shut Off", variant: "destructive" },
  6: { label: "Crashed", variant: "destructive" },
  7: { label: "Suspended", variant: "secondary" },
} as const;

function formatBytes(bytes: number) {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
}

export function VMDetailDialogContent({ vm, onOpenConsole }: VMDetailDialogProps) {
  if (!vm) return null;

  const status = (VM_STATES[vm.state as keyof typeof VM_STATES] || { label: "Unknown", variant: "secondary" }) as any;

  return (
    <div className="space-y-6">
      {/* Console Button */}
      {vm.vnc_ip && vm.vnc_port && (
        <Button
          onClick={() => onOpenConsole?.(vm.uuid)}
          className="w-full gap-2"
          variant="default"
        >
          <Monitor className="h-4 w-4" />
          Open Console
        </Button>
      )}

      {/* Status Section */}
      <div className="space-y-2">
        <h3 className="font-semibold text-sm text-muted-foreground">
          STATUS
        </h3>
        <Badge variant={status.variant}>{status.label}</Badge>
      </div>

      {/* Memory Section */}
      <div className="space-y-2">
        <h3 className="font-semibold text-sm text-muted-foreground">
          MEMORY
        </h3>
        <div className="space-y-2 bg-accent/30 rounded-lg p-3">
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">Current:</span>
            <span className="font-semibold">
              {formatBytes(vm.memory_kb * 1024)}
            </span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">Maximum:</span>
            <span className="font-semibold">
              {formatBytes(vm.max_mem_kb * 1024)}
            </span>
          </div>
          <div className="w-full bg-secondary rounded-full h-2 overflow-hidden">
            <div
              className="bg-primary h-full transition-all"
              style={{
                width: `${(vm.memory_kb / vm.max_mem_kb) * 100}%`,
              }}
            />
          </div>
        </div>
      </div>

      {/* CPU Section */}
      <div className="space-y-2">
        <h3 className="font-semibold text-sm text-muted-foreground">CPU</h3>
        <div className="space-y-2 bg-accent/30 rounded-lg p-3">
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">vCPUs:</span>
            <span className="font-semibold">{vm.vcpus}</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">CPU Time:</span>
            <span className="font-semibold">
              {(vm.cpu_time_ns / 1e9).toFixed(2)}s
            </span>
          </div>
        </div>
      </div>

      {/* Identifiers Section */}
      <div className="space-y-2">
        <h3 className="font-semibold text-sm text-muted-foreground">
          IDENTIFIERS
        </h3>
        <div className="space-y-2 bg-accent/30 rounded-lg p-3">
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground">UUID</p>
            <p className="font-mono text-xs break-all">{vm.uuid}</p>
          </div>
          <div className="space-y-1 pt-2 border-t border-border">
            <p className="text-xs text-muted-foreground">Domain ID</p>
            <p className="font-mono text-xs">{vm.id}</p>
          </div>
        </div>
      </div>

      {/* Performance Section */}
      <div className="space-y-2">
        <h3 className="font-semibold text-sm text-muted-foreground">
          PERFORMANCE
        </h3>
        <div className="space-y-2 bg-accent/30 rounded-lg p-3">
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">
              Memory Usage:
            </span>
            <span className="font-semibold">
              {((vm.memory_kb / vm.max_mem_kb) * 100).toFixed(1)}%
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
