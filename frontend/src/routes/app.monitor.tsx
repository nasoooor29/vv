import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart";
import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { usePermission } from "@/components/protected-content";
import {
  RBAC_AUDIT_LOG_VIEWER,
  RBAC_HEALTH_CHECKER,
  RBAC_SETTINGS_MANAGER,
} from "@/types/types.gen";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  PieChart,
  Pie,
  Cell,
  LineChart,
  Line,
} from "recharts";
import {
  Activity,
  HardDrive,
  AlertTriangle,
  CheckCircle,
  XCircle,
  TrendingUp,
} from "lucide-react";
import { formatBytes } from "@/lib/utils";

const COLORS = [
  "var(--color-chart-1)",
  "var(--color-chart-2)",
  "var(--color-chart-3)",
  "var(--color-chart-4)",
  "var(--color-chart-5)",
];

function MonitorPage() {
  const { hasPermission, hasAnyPermission } = usePermission();

  if (
    !hasAnyPermission([
      RBAC_AUDIT_LOG_VIEWER,
      RBAC_HEALTH_CHECKER,
      RBAC_SETTINGS_MANAGER,
    ])
  ) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-muted-foreground">
          You don't have permission to view monitoring data.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">System Monitoring</h1>
        <p className="text-muted-foreground">
          Detailed system metrics and performance data
        </p>
      </div>

      <Tabs defaultValue="metrics" className="space-y-4">
        <TabsList>
          <TabsTrigger value="metrics">Metrics</TabsTrigger>
          <TabsTrigger value="storage">Storage</TabsTrigger>
          <TabsTrigger value="health">Health</TabsTrigger>
        </TabsList>

        <TabsContent value="metrics" className="space-y-4">
          {hasPermission(RBAC_AUDIT_LOG_VIEWER) && <MetricsSection />}
        </TabsContent>

        <TabsContent value="storage" className="space-y-4">
          {hasPermission(RBAC_SETTINGS_MANAGER) && <StorageSection />}
        </TabsContent>

        <TabsContent value="health" className="space-y-4">
          {hasPermission(RBAC_HEALTH_CHECKER) && <HealthSection />}
        </TabsContent>
      </Tabs>
    </div>
  );
}

function MetricsSection() {
  const { data: metrics, isLoading } = useQuery(
    orpc.metrics.getMetrics.queryOptions({
      input: { days: 7 },
    })
  );

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {[...Array(4)].map((_, i) => (
          <Card key={i}>
            <CardHeader>
              <Skeleton className="h-5 w-32" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-48 w-full" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  const logLevelConfig: ChartConfig = {
    INFO: { label: "Info", color: "var(--color-chart-1)" },
    WARN: { label: "Warning", color: "var(--color-chart-4)" },
    ERROR: { label: "Error", color: "var(--color-destructive)" },
  };

  const serviceConfig: ChartConfig = {};
  metrics?.service_group_distribution?.forEach((s, i) => {
    serviceConfig[s.service_group] = {
      label: s.service_group,
      color: COLORS[i % COLORS.length],
    };
  });

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
      {/* Log Level Distribution */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5" />
            Log Level Distribution
          </CardTitle>
          <CardDescription>
            Distribution of log levels over the last{" "}
            {metrics?.period?.days ?? 7} days
          </CardDescription>
        </CardHeader>
        <CardContent>
          {metrics?.log_level_distribution &&
          metrics.log_level_distribution.length > 0 ? (
            <ChartContainer config={logLevelConfig} className="h-64 w-full">
              <PieChart>
                <Pie
                  data={metrics.log_level_distribution}
                  dataKey="count"
                  nameKey="level"
                  cx="50%"
                  cy="50%"
                  outerRadius={80}
                  label={({ level, percentage }) =>
                    `${level} (${percentage.toFixed(1)}%)`
                  }
                >
                  {metrics.log_level_distribution.map((entry) => (
                    <Cell
                      key={entry.level}
                      fill={
                        entry.level === "ERROR"
                          ? "var(--color-destructive)"
                          : entry.level === "WARN"
                            ? "var(--color-chart-4)"
                            : "var(--color-chart-1)"
                      }
                    />
                  ))}
                </Pie>
                <ChartTooltip content={<ChartTooltipContent />} />
              </PieChart>
            </ChartContainer>
          ) : (
            <p className="text-sm text-muted-foreground">No data available</p>
          )}
        </CardContent>
      </Card>

      {/* Service Distribution */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingUp className="h-5 w-5" />
            Service Activity
          </CardTitle>
          <CardDescription>Log count by service group</CardDescription>
        </CardHeader>
        <CardContent>
          {metrics?.service_group_distribution &&
          metrics.service_group_distribution.length > 0 ? (
            <ChartContainer config={serviceConfig} className="h-64 w-full">
              <BarChart data={metrics.service_group_distribution}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="service_group" tick={{ fontSize: 12 }} />
                <YAxis />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Bar dataKey="count" fill="var(--color-chart-1)" radius={4} />
              </BarChart>
            </ChartContainer>
          ) : (
            <p className="text-sm text-muted-foreground">No data available</p>
          )}
        </CardContent>
      </Card>

      {/* Hourly Activity */}
      <Card className="lg:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5" />
            Hourly Activity
          </CardTitle>
          <CardDescription>
            Average log count by hour of day
          </CardDescription>
        </CardHeader>
        <CardContent>
          {metrics?.log_count_by_hour &&
          metrics.log_count_by_hour.length > 0 ? (
            <ChartContainer
              config={{ log_count: { label: "Logs", color: "var(--color-chart-1)" } }}
              className="h-64 w-full"
            >
              <LineChart data={metrics.log_count_by_hour}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="hour" tick={{ fontSize: 12 }} />
                <YAxis />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Line
                  type="monotone"
                  dataKey="log_count"
                  stroke="var(--color-chart-1)"
                  strokeWidth={2}
                  dot={false}
                />
              </LineChart>
            </ChartContainer>
          ) : (
            <p className="text-sm text-muted-foreground">No data available</p>
          )}
        </CardContent>
      </Card>

      {/* Error Rates */}
      <Card className="lg:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5" />
            Error Rates by Service
          </CardTitle>
          <CardDescription>
            Error rates for each service group
          </CardDescription>
        </CardHeader>
        <CardContent>
          {metrics?.error_rate_by_service &&
          metrics.error_rate_by_service.length > 0 ? (
            <div className="space-y-4">
              {metrics.error_rate_by_service.map((service) => (
                <div key={service.service_group} className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">
                      {service.service_group}
                    </span>
                    <span className="text-sm text-muted-foreground">
                      {service.error_count} / {service.total_count} errors (
                      {service.error_rate.toFixed(2)}%)
                    </span>
                  </div>
                  <Progress
                    value={Math.min(service.error_rate, 100)}
                    className={
                      service.error_rate > 10
                        ? "[&>[data-slot=progress-indicator]]:bg-destructive"
                        : service.error_rate > 5
                          ? "[&>[data-slot=progress-indicator]]:bg-yellow-500"
                          : ""
                    }
                  />
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">No data available</p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function StorageSection() {
  const { data: devices, isLoading: devicesLoading } = useQuery(
    orpc.storage.devices.queryOptions({})
  );

  const { data: mountPoints, isLoading: mountsLoading } = useQuery(
    orpc.storage.mountPoints.queryOptions({})
  );

  const isLoading = devicesLoading || mountsLoading;

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {[...Array(2)].map((_, i) => (
          <Card key={i}>
            <CardHeader>
              <Skeleton className="h-5 w-32" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-48 w-full" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
      {/* Storage Devices */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <HardDrive className="h-5 w-5" />
            Storage Devices
          </CardTitle>
          <CardDescription>Block devices detected on the system</CardDescription>
        </CardHeader>
        <CardContent>
          {devices && devices.length > 0 ? (
            <div className="space-y-3">
              {devices.map((device) => (
                <div
                  key={device.name}
                  className="flex items-center justify-between rounded-lg border p-3"
                >
                  <div>
                    <p className="font-medium">{device.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {device.type} - {device.size}
                    </p>
                  </div>
                  <div className="text-right">
                    {device.mount_point ? (
                      <Badge variant="outline">{device.mount_point}</Badge>
                    ) : (
                      <Badge variant="secondary">Unmounted</Badge>
                    )}
                    {device.usage_percent > 0 && (
                      <p className="text-xs text-muted-foreground mt-1">
                        {device.usage_percent}% used
                      </p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">
              No storage devices found
            </p>
          )}
        </CardContent>
      </Card>

      {/* Mount Points */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <HardDrive className="h-5 w-5" />
            Mount Points
          </CardTitle>
          <CardDescription>Filesystem mount points and usage</CardDescription>
        </CardHeader>
        <CardContent>
          {mountPoints && mountPoints.length > 0 ? (
            <div className="space-y-4">
              {mountPoints.map((mount) => (
                <div key={mount.path} className="space-y-2">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-medium">{mount.path}</p>
                      <p className="text-xs text-muted-foreground">
                        {mount.device} ({mount.fs_type})
                      </p>
                    </div>
                    <span className="text-sm font-medium">
                      {mount.use_percent}%
                    </span>
                  </div>
                  <Progress value={mount.use_percent} />
                  <p className="text-xs text-muted-foreground">
                    {formatBytes(mount.used)} / {formatBytes(mount.total)} (
                    {formatBytes(mount.available)} free)
                  </p>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">
              No mount points found
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function HealthSection() {
  const { data: health, isLoading } = useQuery(
    orpc.metrics.getHealthMetrics.queryOptions({
      input: {},
    })
  );

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-32" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-48 w-full" />
        </CardContent>
      </Card>
    );
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "healthy":
        return <CheckCircle className="h-5 w-5 text-primary" />;
      case "warning":
        return <AlertTriangle className="h-5 w-5 text-yellow-500" />;
      case "critical":
        return <XCircle className="h-5 w-5 text-destructive" />;
      default:
        return <Activity className="h-5 w-5 text-muted-foreground" />;
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

  return (
    <div className="space-y-4">
      {/* Overall Status */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              {getStatusIcon(health?.overall_status ?? "unknown")}
              <div>
                <CardTitle>System Health Status</CardTitle>
                <CardDescription>
                  Last updated: {health?.period ?? "unknown"}
                </CardDescription>
              </div>
            </div>
            {getStatusBadge(health?.overall_status ?? "unknown")}
          </div>
        </CardHeader>
        <CardContent>
          {health?.alerts && health.alerts.length > 0 && (
            <div className="mb-4 space-y-2">
              {health.alerts.map((alert, idx) => (
                <div
                  key={idx}
                  className="flex items-center gap-2 rounded-lg border border-destructive/50 bg-destructive/10 p-3 text-sm"
                >
                  <AlertTriangle className="h-4 w-4 text-destructive" />
                  {alert}
                </div>
              ))}
            </div>
          )}

          {health?.services && health.services.length > 0 ? (
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
              {health.services.map((service) => (
                <Card key={service.service_group}>
                  <CardContent className="pt-6">
                    <div className="flex items-center justify-between mb-3">
                      <span className="font-medium">
                        {service.service_group}
                      </span>
                      {getStatusBadge(service.status)}
                    </div>
                    <div className="space-y-2 text-sm text-muted-foreground">
                      <div className="flex justify-between">
                        <span>Error Rate</span>
                        <span className="font-mono">
                          {service.error_rate.toFixed(2)}%
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span>Errors</span>
                        <span>{service.error_count}</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Total</span>
                        <span>{service.total_count}</span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">
              No service health data available
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default MonitorPage;
