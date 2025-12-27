import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Play,
  Square,
  Trash2,
  Info,
  RotateCcw,
  Box,
  Clock,
  Network,
} from "lucide-react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { formatTimeDifference } from "@/lib/utils";
import { toast } from "sonner";
import type { z } from "zod";
import type { Z } from "@/types";

type DockerContainer = z.infer<typeof Z.dockerContainerSchema>;

interface ContainersGridProps {
  clientId: string;
  onViewDetails: (container: DockerContainer) => void;
  onDelete: (container: DockerContainer) => void;
}

export function ContainersGrid({
  clientId,
  onViewDetails,
  onDelete,
}: ContainersGridProps) {
  const {
    data: containers,
    isLoading,
    isError,
  } = useQuery(
    orpc.docker.containers.queryOptions({
      input: { params: { clientId } },
    }),
  );

  const startMutation = useMutation(
    orpc.docker.startContainer.mutationOptions({
      onSuccess: () => {
        toast.success("Container started");
      },
    }),
  );

  const stopMutation = useMutation(
    orpc.docker.stopContainer.mutationOptions({
      onSuccess: () => {
        toast.success("Container stopped");
      },
    }),
  );

  const restartMutation = useMutation(
    orpc.docker.restartContainer.mutationOptions({
      onSuccess: () => {
        toast.success("Container restarted");
      },
    }),
  );

  const isActionPending =
    startMutation.isPending ||
    stopMutation.isPending ||
    restartMutation.isPending;

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {[1, 2, 3].map((i) => (
          <Skeleton key={i} className="h-40 w-full" />
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <div className="py-8 text-center text-destructive">
        Failed to load containers. Check Docker connection.
      </div>
    );
  }

  const sortedContainers = [...(containers || [])].sort((a, b) => {
    if (a.State === "running" && b.State !== "running") return -1;
    if (a.State !== "running" && b.State === "running") return 1;
    return 0;
  });

  if (!sortedContainers || sortedContainers.length === 0) {
    return (
      <div className="py-12 text-center text-muted-foreground">
        <Box className="h-12 w-12 mx-auto mb-4 opacity-50" />
        <p>No containers found</p>
        <p className="text-sm">Pull an image and create a container to get started.</p>
      </div>
    );
  }

  const getStatusBadge = (state: string) => {
    switch (state) {
      case "running":
        return (
          <Badge className="bg-green-500 text-white hover:bg-green-600">
            Running
          </Badge>
        );
      case "exited":
        return <Badge variant="destructive">Exited</Badge>;
      case "paused":
        return <Badge variant="secondary">Paused</Badge>;
      case "created":
        return <Badge variant="outline">Created</Badge>;
      default:
        return <Badge variant="outline">{state}</Badge>;
    }
  };

  const formatPorts = (ports: DockerContainer["Ports"]) => {
    if (!ports || ports.length === 0) return null;
    return ports
      .filter((p) => p.PublicPort)
      .map((p) => `${p.PublicPort}:${p.PrivatePort}`)
      .slice(0, 3)
      .join(", ");
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
      {sortedContainers.map((container) => {
        const isRunning = container.State === "running";
        const containerName = container.Names[0]?.slice(1) || "unnamed";
        const portsDisplay = formatPorts(container.Ports);

        return (
          <div
            key={container.Id}
            className="border border-border rounded-lg p-4 space-y-3 hover:bg-accent/50 transition-colors"
          >
            {/* Header */}
            <div className="flex items-start justify-between">
              <div className="flex-1 min-w-0">
                <h3 className="font-semibold text-lg truncate">{containerName}</h3>
                <p className="text-xs text-muted-foreground font-mono truncate">
                  {container.Id.slice(0, 12)}
                </p>
              </div>
              {getStatusBadge(container.State)}
            </div>

            {/* Info Grid */}
            <div className="grid grid-cols-2 gap-2 text-sm">
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground flex items-center gap-1">
                  <Box className="h-3 w-3" />
                  Image
                </p>
                <p className="font-medium truncate" title={container.Image}>
                  {container.Image.split(":")[0].split("/").pop()}
                </p>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  Created
                </p>
                <p className="font-medium">
                  {formatTimeDifference(new Date(container.Created * 1000))}
                </p>
              </div>
            </div>

            {/* Ports */}
            {portsDisplay && (
              <div className="text-xs">
                <span className="text-muted-foreground flex items-center gap-1">
                  <Network className="h-3 w-3" />
                  Ports: <span className="font-mono text-foreground">{portsDisplay}</span>
                </span>
              </div>
            )}

            {/* Status info */}
            <div className="text-xs text-muted-foreground">
              {container.Status}
            </div>

            {/* Actions */}
            <div className="flex gap-2 pt-2 flex-wrap">
              <Button
                variant="outline"
                size="sm"
                onClick={() => onViewDetails(container)}
                className="gap-2 flex-1"
              >
                <Info className="h-4 w-4" />
                Details
              </Button>

              {!isRunning && (
                <Button
                  variant="outline"
                  size="sm"
                  disabled={isActionPending}
                  onClick={() =>
                    startMutation.mutate({
                      params: { clientId, id: container.Id },
                    })
                  }
                  className="gap-2"
                >
                  <Play className="h-4 w-4" />
                  Start
                </Button>
              )}

              {isRunning && (
                <>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={isActionPending}
                    onClick={() =>
                      stopMutation.mutate({
                        params: { clientId, id: container.Id },
                      })
                    }
                    className="gap-2"
                  >
                    <Square className="h-4 w-4" />
                    Stop
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={isActionPending}
                    onClick={() =>
                      restartMutation.mutate({
                        params: { clientId, id: container.Id },
                      })
                    }
                    className="gap-2"
                  >
                    <RotateCcw className="h-4 w-4" />
                    Restart
                  </Button>
                </>
              )}

              <Button
                variant="outline"
                size="sm"
                onClick={() => onDelete(container)}
                className="gap-2 text-destructive hover:text-destructive"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </div>
        );
      })}
    </div>
  );
}
