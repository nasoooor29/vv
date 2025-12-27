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
import { Container, Play, Square, RotateCcw } from "lucide-react";
import { formatTimeDifference } from "@/lib/utils";
import { toast } from "sonner";

function RecentContainers() {
  const { data: clients } = useQuery(orpc.docker.clients.queryOptions({}));

  const clientId = clients?.[0]?.id?.toString() ?? "1";
  const { data: containers, isLoading } = useQuery(
    orpc.docker.containers.queryOptions({
      input: { params: { clientId } },
    })
  );

  const startMutation = useMutation(
    orpc.docker.startContainer.mutationOptions({
      onSuccess: () => toast.success("Container started"),
    })
  );

  const stopMutation = useMutation(
    orpc.docker.stopContainer.mutationOptions({
      onSuccess: () => toast.success("Container stopped"),
    })
  );

  const restartMutation = useMutation(
    orpc.docker.restartContainer.mutationOptions({
      onSuccess: () => toast.success("Container restarted"),
    })
  );

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Container className="h-5 w-5" />
            Recent Containers
          </CardTitle>
          <CardDescription>Recently created containers</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {[...Array(4)].map((_, i) => (
            <Skeleton key={i} className="h-16 w-full" />
          ))}
        </CardContent>
      </Card>
    );
  }

  // Sort by created time and take top 5
  const recentContainers = [...(containers ?? [])]
    .sort((a, b) => b.Created - a.Created)
    .slice(0, 5);

  const getStateColor = (state: string) => {
    switch (state.toLowerCase()) {
      case "running":
        return "bg-primary";
      case "exited":
      case "stopped":
        return "bg-muted-foreground";
      case "paused":
        return "bg-yellow-500";
      case "dead":
      case "restarting":
        return "bg-destructive";
      default:
        return "bg-muted-foreground";
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Container className="h-5 w-5" />
          Recent Containers
        </CardTitle>
        <CardDescription>
          {containers?.length ?? 0} containers total
        </CardDescription>
      </CardHeader>
      <CardContent>
        {recentContainers.length > 0 ? (
          <div className="space-y-3">
            {recentContainers.map((container) => (
              <div
                key={container.Id}
                className="flex items-center justify-between gap-2 rounded-lg border p-3 overflow-hidden"
              >
                <div className="flex items-center gap-2 min-w-0 flex-1">
                  <div
                    className={`h-2 w-2 rounded-full shrink-0 ${getStateColor(container.State)}`}
                  />
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium truncate">
                      {container.Names[0]?.replace(/^\//, "") ?? container.Id.slice(0, 12)}
                    </p>
                    <p className="text-xs text-muted-foreground truncate">
                      {container.Image}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <Badge variant="outline" className="text-xs hidden sm:inline-flex">
                    {formatTimeDifference(container.Created * 1000)}
                  </Badge>
                  <div className="flex gap-1">
                    {container.State !== "running" && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7"
                        onClick={() =>
                          startMutation.mutate({
                            params: { clientId, id: container.Id },
                          })
                        }
                        disabled={startMutation.isPending}
                      >
                        <Play className="h-3 w-3" />
                      </Button>
                    )}
                    {container.State === "running" && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7"
                        onClick={() =>
                          stopMutation.mutate({
                            params: { clientId, id: container.Id },
                          })
                        }
                        disabled={stopMutation.isPending}
                      >
                        <Square className="h-3 w-3" />
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7"
                      onClick={() =>
                        restartMutation.mutate({
                          params: { clientId, id: container.Id },
                        })
                      }
                      disabled={restartMutation.isPending}
                    >
                      <RotateCcw className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">No containers found</p>
        )}
      </CardContent>
    </Card>
  );
}

export default RecentContainers;
