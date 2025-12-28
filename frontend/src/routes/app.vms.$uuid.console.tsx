import { useEffect, useRef, useState } from "react";
import { useParams, useNavigate } from "react-router";
import { VncScreen, type VncScreenHandle } from "react-vnc";
import { orpc } from "@/lib/orpc";
import { useQuery } from "@tanstack/react-query";
import { CONSTANTS } from "@/lib";
import { AlertCircle, ArrowLeft, Monitor } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "sonner";

export default function VMConsolePage() {
  console.log("help");
  const { uuid } = useParams<{ uuid: string }>();
  const navigate = useNavigate();
  const ref = useRef<VncScreenHandle>(null);

  // Fetch VM details
  const vmQuery = useQuery(
    orpc.qemu.getVirtualMachineInfo.queryOptions({
      input: {
        params: { uuid: uuid || "" },
      },
      queryOptions: {
        staleTime: CONSTANTS.POLLING_INTERVAL_MS,
      },
    }),
  );

  const health = useQuery(
    orpc.health.check.queryOptions({
      queryOptions: {
        staleTime: CONSTANTS.POLLING_INTERVAL_MS,
      },
    }),
  );
  const baseURL = health.data?.base_url?.replace(/^https?:\/\//, "");

  const vmName = vmQuery.data?.name;
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const url = `${protocol}//${baseURL}/api/qemu/virtual-machines/${uuid}/console`;

  useEffect(() => {
    if (!uuid || !ref.current) {
      return;
    }
  }, [uuid]);

  if (!uuid) {
    return (
      <Alert className="border-destructive bg-destructive/10">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>Invalid VM UUID</AlertDescription>
      </Alert>
    );
  }

  if (vmQuery.isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-12 w-full" />
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  if (vmQuery.isError) {
    return (
      <Alert className="border-destructive bg-destructive/10">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>Failed to load VM details</AlertDescription>
      </Alert>
    );
  }

  // if (!vncIP || !vncPort) {
  //   return (
  //     <Alert className="border-yellow-500 bg-yellow-500/10">
  //       <AlertCircle className="h-4 w-4" />
  //       <AlertDescription>
  //         VNC is not available for this VM. Please ensure the VM is running and
  //         has VNC configured.
  //       </AlertDescription>
  //     </Alert>
  //   );
  // }

  return (
    <div className="flex flex-col h-[calc(100vh-2rem)] space-y-4">
      {/* Header Section */}
      <div className="flex items-center justify-between shrink-0">
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate("/app/vms")}
            className="gap-2"
          >
            <ArrowLeft className="h-4 w-4" />
            Back
          </Button>
          <h1 className="text-xl font-bold flex items-center gap-2 truncate">
            <Monitor className="h-5 w-5" />
            {vmName}
          </h1>
        </div>
        <div className="text-xs font-mono text-muted-foreground bg-secondary px-2 py-1 rounded">
          {url}
        </div>
      </div>

      {/* Main Console Card */}
      <Card className="flex-1 flex flex-col min-h-0 overflow-hidden">
        <CardContent className="flex-1 p-0 bg-black relative min-h-0">
          <div className="absolute inset-0 flex items-center justify-center">
            <VncScreen
              url={url}
              scaleViewport
              background="#000"
              style={{
                width: "100%",
                height: "100%",
                objectFit: "contain", // Keeps aspect ratio without distortion
              }}
              ref={ref}
            />
          </div>
        </CardContent>
      </Card>

      {/* Footer Info - keep it tight */}
      <div className="flex gap-4 text-[10px] text-muted-foreground shrink-0 uppercase tracking-wider">
        <span>🖱️ Drag to move</span>
        <span>🎡 Scroll to zoom</span>
        <span>🖱️ Right-click menu</span>
      </div>
    </div>
  );
}
