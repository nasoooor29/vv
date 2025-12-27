import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Container } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Skeleton } from "@/components/ui/skeleton";

function ContainerSummary() {
  // First get available docker clients
  const { data: clients, isLoading: clientsLoading } = useQuery(
    orpc.docker.clients.queryOptions({})
  );

  // Get containers from the first available client
  const clientId = clients?.[0]?.id?.toString() ?? "1";
  const { data: containers, isLoading: containersLoading } = useQuery(
    orpc.docker.containers.queryOptions({
      input: { params: { clientId } },
    })
  );

  const isLoading = clientsLoading || containersLoading;

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Containers</CardTitle>
          <Container className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-8 w-16 mb-2" />
          <Skeleton className="h-5 w-32" />
        </CardContent>
      </Card>
    );
  }

  const stats = {
    total: containers?.length ?? 0,
    running: containers?.filter((c) => c.State === "running").length ?? 0,
    stopped:
      containers?.filter((c) => c.State === "exited" || c.State === "stopped")
        .length ?? 0,
    error:
      containers?.filter(
        (c) => c.State === "dead" || c.State === "restarting"
      ).length ?? 0,
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">Containers</CardTitle>
        <Container className="h-4 w-4 text-muted-foreground" />
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
          {stats.error > 0 && (
            <Badge variant="destructive">{stats.error} Error</Badge>
          )}
          {stats.total === 0 && (
            <span className="text-xs text-muted-foreground">No containers</span>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

export default ContainerSummary;
