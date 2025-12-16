import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertCircle, Disc3 } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Progress } from "@/components/ui/progress";
import { formatBytes } from "@/lib/utils";

export default function MountPointsPage() {
  const mountPointsQuery = useQuery(
    orpc.storage.mountPoints.queryOptions({
      staleTime: 5 * 1000,
    }),
  );

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="flex items-center gap-2">
          <Disc3 className="h-5 w-5" />
          Mount Points
        </CardTitle>
      </CardHeader>
      <CardContent>
        {mountPointsQuery.isLoading ? (
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-20 w-full" />
            ))}
          </div>
        ) : mountPointsQuery.isError ? (
          <Alert className="border-destructive bg-destructive/10">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Failed to load mount points. Please try again later.
            </AlertDescription>
          </Alert>
        ) : !mountPointsQuery.data || mountPointsQuery.data.length === 0 ? (
          <div className="py-8 text-center text-muted-foreground">
            No mount points found.
          </div>
        ) : (
          <div className="space-y-4">
            {mountPointsQuery.data.map((mp, index) => (
              <div
                key={`${mp.path}-${index}`}
                className="border border-border rounded-lg p-4 space-y-3"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <p className="font-semibold font-mono">{mp.path}</p>
                    <p className="text-sm text-muted-foreground">
                      Device: {mp.device} ({mp.fs_type})
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="text-lg font-semibold">
                      {formatBytes(mp.used)} / {formatBytes(mp.total)}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {mp.use_percent}% used
                    </p>
                  </div>
                </div>

                <div className="space-y-1">
                  <div className="flex justify-between text-xs text-muted-foreground">
                    <span>Usage</span>
                    <span>{mp.use_percent}%</span>
                  </div>
                  <Progress
                    value={mp.use_percent}
                    className="h-2"
                    style={{
                      backgroundColor: "hsl(var(--muted))",
                    }}
                  />
                </div>

                <div className="grid grid-cols-3 gap-2 text-sm">
                  <div>
                    <p className="text-muted-foreground">Used</p>
                    <p className="font-semibold">{formatBytes(mp.used)}</p>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Available</p>
                    <p className="font-semibold">{formatBytes(mp.available)}</p>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Total</p>
                    <p className="font-semibold">{formatBytes(mp.total)}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
