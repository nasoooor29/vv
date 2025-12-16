import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertCircle, Disc3 } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { formatBytes } from "@/lib/utils";

export default function StorageDevicesPage() {
  const devicesQuery = useQuery(
    orpc.storage.devices.queryOptions({
      staleTime: 5 * 1000,
    }),
  );

  return (
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
        ) : !devicesQuery.data || devicesQuery.data.length === 0 ? (
          <div className="py-8 text-center text-muted-foreground">
            No storage devices found.
          </div>
        ) : (
          <div className="space-y-4">
            {devicesQuery.data.map((device, index) => (
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
                    <span className="text-muted-foreground">Mount Point:</span>{" "}
                    <span className="font-mono">{device.mount_point}</span>
                  </p>
                )}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
