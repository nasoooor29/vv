import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertCircle, HardDrive, Disc3 } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Progress } from "@/components/ui/progress";
import { usePermission } from "@/components/protected-content";
import { RBAC_SETTINGS_MANAGER } from "@/types/types.gen";
import { formatBytes } from "@/lib/utils";

export default function StoragePage() {
  const { hasPermission } = usePermission();

  const devicesQuery = useQuery(
    orpc.storage.devices.queryOptions({
      staleTime: 5 * 1000,
    }),
  );

  const mountPointsQuery = useQuery(
    orpc.storage.mountPoints.queryOptions({
      staleTime: 5 * 1000,
    }),
  );

  if (!hasPermission(RBAC_SETTINGS_MANAGER)) {
    return (
      <Alert className="border-destructive bg-destructive/10">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          You don't have permission to access storage settings.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <HardDrive className="h-6 w-6" />
        <h1 className="text-3xl font-bold">Storage Management</h1>
      </div>

      {/* Storage Devices Section */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Disc3 className="h-5 w-5" />
            Storage Devices
          </CardTitle>
        </CardHeader>
        <CardContent>
          {devicesQuery.isLoading ? (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : devicesQuery.isError ? (
            <Alert className="border-destructive bg-destructive/10">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Failed to load storage devices. Please try again later.
              </AlertDescription>
            </Alert>
          ) : !devicesQuery.data?.devices ||
            devicesQuery.data.devices.length === 0 ? (
            <div className="py-8 text-center text-muted-foreground">
              No storage devices found.
            </div>
          ) : (
            <div className="space-y-4">
              {devicesQuery.data.devices.map((device: any, index: number) => (
                <div
                  key={`${device.name}-${index}`}
                  className="border border-border rounded-lg p-4 space-y-2"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <p className="font-semibold font-mono">{device.name}</p>
                      <p className="text-sm text-muted-foreground">
                        Type: {device.type}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="font-semibold">{device.size}</p>
                      <p className="text-sm text-muted-foreground">
                        {formatBytes(device.size_bytes)}
                      </p>
                    </div>
                  </div>
                  {device.mount_point && (
                    <p className="text-sm">
                      <span className="text-muted-foreground">
                        Mount Point:
                      </span>{" "}
                      <span className="font-mono">{device.mount_point}</span>
                    </p>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Mount Points Section */}
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
          ) : !mountPointsQuery.data?.mount_points ||
            mountPointsQuery.data.mount_points.length === 0 ? (
            <div className="py-8 text-center text-muted-foreground">
              No mount points found.
            </div>
          ) : (
            <div className="space-y-4">
              {mountPointsQuery.data.mount_points.map(
                (mp: any, index: number) => (
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
                        <p className="font-semibold">
                          {formatBytes(mp.available)}
                        </p>
                      </div>
                      <div>
                        <p className="text-muted-foreground">Total</p>
                        <p className="font-semibold">{formatBytes(mp.total)}</p>
                      </div>
                    </div>
                  </div>
                ),
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
