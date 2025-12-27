import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Skeleton } from "@/components/ui/skeleton";
import { Activity, AlertTriangle, CheckCircle, XCircle } from "lucide-react";

function SystemHealth() {
  const { data: health, isLoading } = useQuery(
    orpc.metrics.getHealthMetrics.queryOptions({
      input: {},
    })
  );

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5" />
            System Health
          </CardTitle>
          <CardDescription>Real-time service health monitoring</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
        </CardContent>
      </Card>
    );
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "healthy":
        return <CheckCircle className="h-4 w-4 text-primary" />;
      case "warning":
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      case "critical":
        return <XCircle className="h-4 w-4 text-destructive" />;
      default:
        return <Activity className="h-4 w-4 text-muted-foreground" />;
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "healthy":
        return <Badge className="bg-primary">Healthy</Badge>;
      case "warning":
        return <Badge className="bg-yellow-500 text-foreground">Warning</Badge>;
      case "critical":
        return <Badge variant="destructive">Critical</Badge>;
      default:
        return <Badge variant="secondary">Unknown</Badge>;
    }
  };

  const overallStatus = health?.overall_status ?? "unknown";

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              System Health
            </CardTitle>
            <CardDescription>Real-time service health monitoring</CardDescription>
          </div>
          {getStatusBadge(overallStatus)}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Alerts */}
        {health?.alerts && health.alerts.length > 0 && (
          <div className="space-y-2">
            {health.alerts.map((alert, idx) => (
              <Alert key={idx} variant="destructive">
                <AlertTriangle className="h-4 w-4" />
                <AlertTitle>Alert</AlertTitle>
                <AlertDescription>{alert}</AlertDescription>
              </Alert>
            ))}
          </div>
        )}

        {/* Services */}
        {health?.services && health.services.length > 0 ? (
          <div className="space-y-2">
            {health.services.map((service) => (
              <div
                key={service.service_group}
                className="flex items-center justify-between rounded-lg border p-3"
              >
                <div className="flex items-center gap-2">
                  {getStatusIcon(service.status)}
                  <span className="font-medium">{service.service_group}</span>
                </div>
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span>
                    {service.error_count} / {service.total_count} errors
                  </span>
                  <span className="font-mono">
                    {service.error_rate.toFixed(1)}%
                  </span>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            No service health data available
          </p>
        )}

        <p className="text-xs text-muted-foreground">
          Period: {health?.period ?? "unknown"}
        </p>
      </CardContent>
    </Card>
  );
}

export default SystemHealth;
