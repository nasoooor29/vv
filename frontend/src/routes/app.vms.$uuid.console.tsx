import { useEffect, useRef, useState } from "react";
import { useParams, useNavigate } from "react-router";
import { orpc } from "@/lib/orpc";
import { useQuery } from "@tanstack/react-query";
import { CONSTANTS } from "@/lib";
import { AlertCircle, ArrowLeft, Monitor } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "sonner";
import * as VNC from "novnc/core/rfb";

export default function VMConsolePage() {
  const { uuid } = useParams<{ uuid: string }>();
  const navigate = useNavigate();
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const rfbRef = useRef<VNC.RFB | null>(null);
  const [isConnecting, setIsConnecting] = useState(false);
  const [isConnected, setIsConnected] = useState(false);

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

  const vncHost = vmQuery.data?.vnc_ip;
  const vncPort = vmQuery.data?.vnc_port;
  const vmName = vmQuery.data?.name;

  useEffect(() => {
    if (!vncHost || !vncPort || !canvasRef.current || isConnecting || isConnected) {
      return;
    }

    const connectVNC = async () => {
      try {
        setIsConnecting(true);

        // Check if host is localhost/127.0.0.1 and adjust for browser access
        let host = vncHost;
        if (host === "127.0.0.1" || host === "localhost") {
          host = window.location.hostname;
        }

        const url = `ws://${host}:${vncPort}`;
        console.log(`Connecting to VNC at ${url}`);

        const rfb = new VNC.RFB(canvasRef.current!, url, {
          credentials: {
            username: "",
            password: "",
            target: "",
          },
        });

        rfb.addEventListener("connect", () => {
          console.log("VNC connected");
          setIsConnected(true);
          setIsConnecting(false);
          toast.success("Connected to VM console");
        });

        rfb.addEventListener("disconnect", () => {
          console.log("VNC disconnected");
          setIsConnected(false);
          toast.info("Disconnected from VM console");
        });

        rfb.addEventListener("error", (e: any) => {
          console.error("VNC error:", e);
          toast.error(`VNC connection error: ${e.detail}`);
          setIsConnecting(false);
        });

        rfbRef.current = rfb;
      } catch (error) {
        console.error("Failed to initialize VNC:", error);
        toast.error("Failed to initialize VNC connection");
        setIsConnecting(false);
      }
    };

    connectVNC();

    return () => {
      if (rfbRef.current) {
        rfbRef.current.disconnect();
        rfbRef.current = null;
      }
    };
  }, [vncHost, vncPort, isConnecting, isConnected]);

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

  if (!vncHost || !vncPort) {
    return (
      <Alert className="border-yellow-500 bg-yellow-500/10">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          VNC is not available for this VM. Please ensure the VM is running and has VNC configured.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => navigate("/app/vms")}
          className="gap-2"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to VMs
        </Button>
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Monitor className="h-6 w-6" />
          Console: {vmName}
        </h1>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">
            <div className="space-y-2">
              <div>
                <span className="text-muted-foreground">Host:</span> {vncHost}:{vncPort}
              </div>
              <div className="flex items-center gap-2">
                <div
                  className={`h-2 w-2 rounded-full ${
                    isConnected ? "bg-green-500" : isConnecting ? "bg-yellow-500" : "bg-red-500"
                  }`}
                />
                <span className="text-xs text-muted-foreground">
                  {isConnected
                    ? "Connected"
                    : isConnecting
                      ? "Connecting..."
                      : "Disconnected"}
                </span>
              </div>
            </div>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="bg-black rounded-lg overflow-hidden border border-border">
            <canvas
              ref={canvasRef}
              className="w-full"
              style={{
                minHeight: "400px",
              }}
            />
          </div>
          <div className="mt-4 text-xs text-muted-foreground space-y-1">
            <p>💡 Click and drag to move the cursor</p>
            <p>💡 Use your mouse wheel to scroll</p>
            <p>💡 Right-click for context menu</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
