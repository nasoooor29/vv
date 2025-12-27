import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { HardDrive } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Skeleton } from "@/components/ui/skeleton";
import { formatBytes } from "@/lib/utils";

function StorageSummary() {
  const { data: mountPoints, isLoading } = useQuery(
    orpc.storage.mountPoints.queryOptions({})
  );

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Storage</CardTitle>
          <HardDrive className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-8 w-20 mb-2" />
          <Skeleton className="h-2 w-full" />
        </CardContent>
      </Card>
    );
  }

  // Get the main/root mount point or first available
  const mainMount =
    mountPoints?.find((m) => m.path === "/") ?? mountPoints?.[0];
  const totalUsed = mountPoints?.reduce((acc, m) => acc + m.used, 0) ?? 0;
  const totalSize = mountPoints?.reduce((acc, m) => acc + m.total, 0) ?? 0;
  const avgUsage =
    totalSize > 0 ? Math.round((totalUsed / totalSize) * 100) : 0;

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">Storage</CardTitle>
        <HardDrive className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{avgUsage}%</div>
        <Progress value={avgUsage} className="mt-2" />
        <p className="mt-2 text-xs text-muted-foreground">
          {formatBytes(totalUsed)} / {formatBytes(totalSize)} used
          {mainMount && ` (${mountPoints?.length ?? 0} mount points)`}
        </p>
      </CardContent>
    </Card>
  );
}

export default StorageSummary;
