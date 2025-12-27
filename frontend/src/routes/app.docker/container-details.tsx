import { useQuery, useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Play, Square, RefreshCw, Loader2, FileText } from "lucide-react";
import { formatBytes, formatTimeDifference } from "@/lib/utils";
import { toast } from "sonner";
import type { z } from "zod";
import type { Z } from "@/types";
import { useDialog } from "@/hooks";

type DockerContainer = z.infer<typeof Z.dockerContainerSchema>;

export function ContainerDetailsDialog(
  clientId: string | null,
  container: DockerContainer | null,
) {
  const containerQuery = useQuery(
    orpc.docker.inspectContainer.queryOptions({
      input: {
        params: { clientId: clientId!, id: container?.Id! },
      },
      queryOptions: {
        enabled: !!clientId && !!container?.Id,
      },
    }),
  );

  const statsQuery = useQuery(
    orpc.docker.containerStats.queryOptions({
      input: {
        params: { clientId: clientId!, id: container?.Id! },
      },
      queryOptions: {
        enabled: !!clientId && !!container?.Id,
        refetchInterval: 5000,
      },
    }),
  );

  const logsQuery = useQuery(
    orpc.docker.containerLogs.queryOptions({
      input: {
        params: { clientId: clientId!, id: container?.Id! },
      },
      queryOptions: {
        enabled: !!clientId && !!container?.Id,
      },
    }),
  );

  const startMutation = useMutation(
    orpc.docker.startContainer.mutationOptions({
      onSuccess: () => toast.success("Container started"),
    }),
  );

  const stopMutation = useMutation(
    orpc.docker.stopContainer.mutationOptions({
      onSuccess: () => toast.success("Container stopped"),
    }),
  );

  const restartMutation = useMutation(
    orpc.docker.restartContainer.mutationOptions({
      onSuccess: () => toast.success("Container restarted"),
    }),
  );

  const inspectData = containerQuery.data;
  const stats = statsQuery.data;

  const isActionPending =
    startMutation.isPending ||
    stopMutation.isPending ||
    restartMutation.isPending;

  // const getStatusBadge = (status: string) => {
  //   const isRunning = status === "running";
  //   return (
  //     <Badge
  //       className={
  //         isRunning
  //           ? "bg-green-500 text-white hover:bg-green-600"
  //           : "bg-destructive"
  //       }
  //     >
  //       {status}
  //     </Badge>
  //   );
  // };

  // Calculate CPU and memory usage from stats
  const cpuPercent = stats?.cpu_stats?.cpu_usage?.total_usage
    ? (
        ((stats.cpu_stats.cpu_usage.total_usage -
          (stats.precpu_stats?.cpu_usage?.total_usage || 0)) /
          (stats.cpu_stats.system_cpu_usage -
            (stats.precpu_stats?.system_cpu_usage || 0))) *
        100
      ).toFixed(2)
    : "N/A";

  const memUsage = stats?.memory_stats?.usage
    ? formatBytes(stats.memory_stats.usage)
    : "N/A";
  const memLimit = stats?.memory_stats?.limit
    ? formatBytes(stats.memory_stats.limit)
    : "N/A";

  const dialog = useDialog({
    title: container?.Names[0]?.slice(1) || "Container Details",
    description: container?.Image,
    className: "max-w-4xl",
    children: (
      <>
        {containerQuery.isLoading ? (
          <div className="flex h-64 items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin" />
          </div>
        ) : containerQuery.error ? (
          <div className="p-4 text-center text-destructive">
            Failed to load container details
          </div>
        ) : (
          <ScrollArea className="h-[calc(100vh-200px)]">
            <div className="space-y-4 p-4">
              {/* Actions */}
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={inspectData?.State?.Running || isActionPending}
                  onClick={() =>
                    startMutation.mutate({
                      params: { clientId: clientId!, id: container?.Id! },
                    })
                  }
                >
                  <Play className="mr-1 h-4 w-4" />
                  Start
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={!inspectData?.State?.Running || isActionPending}
                  onClick={() =>
                    stopMutation.mutate({
                      params: { clientId: clientId!, id: container?.Id! },
                    })
                  }
                >
                  <Square className="mr-1 h-4 w-4" />
                  Stop
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={!inspectData?.State?.Running || isActionPending}
                  onClick={() =>
                    restartMutation.mutate({
                      params: { clientId: clientId!, id: container?.Id! },
                    })
                  }
                >
                  <RefreshCw className="mr-1 h-4 w-4" />
                  Restart
                </Button>
              </div>

              <Tabs defaultValue="overview">
                <TabsList className="w-full">
                  <TabsTrigger value="overview">Overview</TabsTrigger>
                  <TabsTrigger value="logs">Logs</TabsTrigger>
                  <TabsTrigger value="stats">Stats</TabsTrigger>
                  <TabsTrigger value="env">Environment</TabsTrigger>
                  <TabsTrigger value="mounts">Mounts</TabsTrigger>
                  <TabsTrigger value="network">Network</TabsTrigger>
                </TabsList>

                <TabsContent value="overview" className="space-y-4">
                  <div className="space-y-2 rounded-lg border p-4">
                    <h4 className="font-semibold">Container Info</h4>
                    <div className="grid grid-cols-2 gap-2 text-sm">
                      <span className="text-muted-foreground">ID</span>
                      <span className="font-mono">
                        {inspectData?.Id?.slice(0, 12)}
                      </span>
                      <span className="text-muted-foreground">Image</span>
                      <span>{inspectData?.Config?.Image}</span>
                      <span className="text-muted-foreground">Created</span>
                      <span>
                        {inspectData?.Created
                          ? formatTimeDifference(new Date(inspectData.Created))
                          : "N/A"}
                      </span>
                      <span className="text-muted-foreground">Status</span>
                      <span>{inspectData?.State?.Status}</span>
                      <span className="text-muted-foreground">PID</span>
                      <span>{inspectData?.State?.Pid || "N/A"}</span>
                      <span className="text-muted-foreground">
                        Restart Count
                      </span>
                      <span>{inspectData?.RestartCount ?? 0}</span>
                    </div>
                  </div>

                  {inspectData?.NetworkSettings?.Ports &&
                    Object.keys(inspectData.NetworkSettings.Ports).length >
                      0 && (
                      <div className="space-y-2 rounded-lg border p-4">
                        <h4 className="font-semibold">Ports</h4>
                        <Table>
                          <TableHeader>
                            <TableRow>
                              <TableHead>Container</TableHead>
                              <TableHead>Host</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {Object.entries(
                              inspectData.NetworkSettings.Ports,
                            ).map(([port, bindings]: [string, any]) => (
                              <TableRow key={port}>
                                <TableCell>{port}</TableCell>
                                <TableCell>
                                  {bindings
                                    ? bindings
                                        .map(
                                          (b: any) =>
                                            `${b.HostIp || "0.0.0.0"}:${b.HostPort}`,
                                        )
                                        .join(", ")
                                    : "Not bound"}
                                </TableCell>
                              </TableRow>
                            ))}
                          </TableBody>
                        </Table>
                      </div>
                    )}
                </TabsContent>

                <TabsContent value="logs" className="space-y-4">
                  <div className="flex items-center justify-between">
                    <h4 className="font-semibold flex items-center gap-2">
                      <FileText className="h-4 w-4" />
                      Container Logs
                    </h4>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => logsQuery.refetch()}
                      disabled={logsQuery.isFetching}
                    >
                      <RefreshCw
                        className={`mr-1 h-4 w-4 ${logsQuery.isFetching ? "animate-spin" : ""}`}
                      />
                      Refresh
                    </Button>
                  </div>
                  <div className="rounded-lg border bg-muted/50 p-4">
                    {logsQuery.isLoading ? (
                      <div className="flex items-center justify-center py-8">
                        <Loader2 className="h-6 w-6 animate-spin" />
                      </div>
                    ) : logsQuery.error ? (
                      <p className="text-destructive">Failed to load logs</p>
                    ) : logsQuery.data ? (
                      <pre className="max-h-[400px] overflow-auto whitespace-pre-wrap break-all font-mono text-xs">
                        {logsQuery.data || "No logs available"}
                      </pre>
                    ) : (
                      <p className="text-muted-foreground">No logs available</p>
                    )}
                  </div>
                </TabsContent>

                <TabsContent value="stats" className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="rounded-lg border p-4">
                      <h4 className="text-sm font-medium text-muted-foreground">
                        CPU Usage
                      </h4>
                      <p className="text-2xl font-bold">{cpuPercent}%</p>
                    </div>
                    <div className="rounded-lg border p-4">
                      <h4 className="text-sm font-medium text-muted-foreground">
                        Memory Usage
                      </h4>
                      <p className="text-2xl font-bold">{memUsage}</p>
                      <p className="text-xs text-muted-foreground">
                        of {memLimit}
                      </p>
                    </div>
                  </div>
                </TabsContent>

                <TabsContent value="env">
                  <div className="rounded-lg border p-4">
                    {inspectData?.Config?.Env &&
                    inspectData.Config.Env.length > 0 ? (
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Variable</TableHead>
                            <TableHead>Value</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {inspectData.Config.Env.map(
                            (env: string, i: number) => {
                              const [key, ...val] = env.split("=");
                              return (
                                <TableRow key={i}>
                                  <TableCell className="font-mono text-sm">
                                    {key}
                                  </TableCell>
                                  <TableCell className="max-w-[200px] truncate font-mono text-sm">
                                    {val.join("=")}
                                  </TableCell>
                                </TableRow>
                              );
                            },
                          )}
                        </TableBody>
                      </Table>
                    ) : (
                      <p className="text-muted-foreground">
                        No environment variables
                      </p>
                    )}
                  </div>
                </TabsContent>

                <TabsContent value="mounts">
                  <div className="rounded-lg border p-4">
                    {inspectData?.Mounts && inspectData.Mounts.length > 0 ? (
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Type</TableHead>
                            <TableHead>Source</TableHead>
                            <TableHead>Destination</TableHead>
                            <TableHead>Mode</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {inspectData.Mounts.map((m: any, i: number) => (
                            <TableRow key={i}>
                              <TableCell>
                                <Badge variant="outline">{m.Type}</Badge>
                              </TableCell>
                              <TableCell className="max-w-[150px] truncate font-mono text-xs">
                                {m.Source}
                              </TableCell>
                              <TableCell className="max-w-[150px] truncate font-mono text-xs">
                                {m.Destination}
                              </TableCell>
                              <TableCell>
                                <Badge variant={m.RW ? "default" : "secondary"}>
                                  {m.RW ? "RW" : "RO"}
                                </Badge>
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    ) : (
                      <p className="text-muted-foreground">No mounts</p>
                    )}
                  </div>
                </TabsContent>

                <TabsContent value="network">
                  <div className="space-y-4">
                    {inspectData?.NetworkSettings?.Networks ? (
                      Object.entries(inspectData.NetworkSettings.Networks).map(
                        ([name, network]: [string, any]) => (
                          <div key={name} className="rounded-lg border p-4">
                            <h4 className="mb-2 font-semibold">{name}</h4>
                            <div className="grid grid-cols-2 gap-2 text-sm">
                              <span className="text-muted-foreground">
                                IP Address
                              </span>
                              <span className="font-mono">
                                {network.IPAddress || "N/A"}
                              </span>
                              <span className="text-muted-foreground">
                                Gateway
                              </span>
                              <span className="font-mono">
                                {network.Gateway || "N/A"}
                              </span>
                              <span className="text-muted-foreground">
                                MAC Address
                              </span>
                              <span className="font-mono">
                                {network.MacAddress || "N/A"}
                              </span>
                            </div>
                          </div>
                        ),
                      )
                    ) : (
                      <p className="text-muted-foreground">
                        No network information
                      </p>
                    )}
                  </div>
                </TabsContent>
              </Tabs>
            </div>
          </ScrollArea>
        )}
      </>
    ),
  });

  return dialog;
}
