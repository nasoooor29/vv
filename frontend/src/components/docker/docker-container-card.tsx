import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Play,
  RotateCcw,
  Settings,
  Square,
  Terminal,
  Trash2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { formatTimeDifference } from "@/lib/utils";
import type { z } from "zod";
import type { Z } from "@/types";
import { useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { toast } from "sonner";
import { Link } from "react-router";
import { useConfirmation } from "@/hooks/use-confirmation";

type DockerContainer = z.infer<typeof Z.dockerContainerSchema>;

interface DockerContainerCardProps {
  container: DockerContainer;
  clientId: string;
  compact?: boolean;
}

function DockerContainerCard({
  container,
  clientId,
  compact = false,
}: DockerContainerCardProps) {
  const { confirm, ConfirmationDialog } = useConfirmation();

  const startMutation = useMutation(
    orpc.docker.startContainer.mutationOptions({
      onSuccess: () => {
        toast.success(`Container ${container.Names[0]?.slice(1)} started`);
      },
    }),
  );

  const stopMutation = useMutation(
    orpc.docker.stopContainer.mutationOptions({
      onSuccess: () => {
        toast.success(`Container ${container.Names[0]?.slice(1)} stopped`);
      },
    }),
  );

  const deleteMutation = useMutation(
    orpc.docker.deleteContainer.mutationOptions({
      onSuccess: () => {
        toast.success(`Container ${container.Names[0]?.slice(1)} deleted`);
      },
    }),
  );

  const getStatusColor = (status: string) => {
    switch (status) {
      case "running":
        return "bg-green-500";
      case "stopped":
        return "bg-secondary";
      case "exited":
        return "bg-destructive";
      case "error":
        return "bg-destructive";
      default:
        return "bg-muted-foreground";
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "running":
        return (
          <Badge className="bg-green-500 text-white hover:bg-green-600">
            Running
          </Badge>
        );
      case "stopped":
        return <Badge variant="secondary">Stopped</Badge>;
      case "exited":
        return <Badge variant="destructive">Exited</Badge>;
      case "error":
        return <Badge variant="destructive">Error</Badge>;
      default:
        return <Badge variant="secondary">{status}</Badge>;
    }
  };

  const isLoading =
    startMutation.isPending ||
    stopMutation.isPending ||
    deleteMutation.isPending;

  return (
    <>
      <ConfirmationDialog />
      <Card
        className={`${compact && "rounded-md shadow-none"} py-0 transition-shadow`}
      >
        <CardContent className={`${compact ? "p-3" : "p-6"}`}>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div
                className={`${getStatusColor(container.State)} h-3 w-3 rounded-full`}
              />
              <div>
                {!compact && (
                  <Link
                    to={`/docker/${clientId}/${container.Id}`}
                    className="text-lg font-semibold hover:underline"
                  >
                    {container.Names[0]?.slice(1)}
                  </Link>
                )}
                {compact && (
                  <p className="font-medium">{container.Names[0]?.slice(1)}</p>
                )}
                <p className="text-sm text-muted-foreground">
                  {container.Image}
                </p>
              </div>

              {!compact && getStatusBadge(container.State)}
            </div>
            <div className="flex items-center gap-2">
              <Button
                size="sm"
                variant="outline"
                disabled={container.State === "running" || isLoading}
                onClick={() =>
                  startMutation.mutate({
                    params: { clientId, id: container.Id },
                  })
                }
              >
                <Play className="h-4 w-4" />
              </Button>
              <Button
                size="sm"
                variant="outline"
                disabled={container.State !== "running" || isLoading}
                onClick={() =>
                  stopMutation.mutate({
                    params: { clientId, id: container.Id },
                  })
                }
              >
                <Square className="h-4 w-4" />
              </Button>
              <Button
                size="sm"
                variant="outline"
                disabled={container.State !== "running" || isLoading}
              >
                <RotateCcw className="h-4 w-4" />
              </Button>
              {!compact && (
                <>
                  <Button
                    size="sm"
                    variant="outline"
                    disabled={container.State !== "running"}
                  >
                    <Terminal className="h-4 w-4" />
                  </Button>
                  <Button size="sm" variant="outline" asChild>
                    <Link to={`/docker/${clientId}/${container.Id}`}>
                      <Settings className="h-4 w-4" />
                    </Link>
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    className="text-destructive hover:text-destructive"
                    onClick={() =>
                      confirm({
                        title: "Delete Container",
                        description: `Are you sure you want to delete container "${container.Names[0]?.slice(1)}"? This action cannot be undone.`,
                        confirmText: "Delete",
                        cancelText: "Cancel",
                        isDestructive: true,
                        onConfirm() {
                          deleteMutation.mutate({
                            params: { clientId, id: container.Id },
                          });
                        },
                      })
                    }
                    disabled={isLoading}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </>
              )}
            </div>
          </div>

          {!compact && (
            <>
              <div className="mt-4 grid grid-cols-2 gap-4 md:grid-cols-4">
                {container.Ports.length > 0 && (
                  <div className="space-y-1">
                    <div className="text-sm font-medium">Ports</div>
                    <div className="text-sm text-muted-foreground">
                      {container.Ports.map((port, index) => (
                        <span key={index}>
                          {port.PublicPort
                            ? `${port.PublicPort}:${port.PrivatePort}`
                            : port.PrivatePort}
                          /{port.Type}{" "}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
                <div className="space-y-1">
                  <div className="text-sm font-medium">Status</div>
                  <div className="text-sm text-muted-foreground">
                    {container.Status}
                  </div>
                </div>
              </div>
              <div className="mt-3 text-xs text-muted-foreground">
                Created{" "}
                {formatTimeDifference(new Date(container.Created * 1000))}
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </>
  );
}

export default DockerContainerCard;
