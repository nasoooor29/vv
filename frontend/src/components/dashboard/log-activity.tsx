import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Skeleton } from "@/components/ui/skeleton";
import { ScrollText, AlertCircle, Info, AlertTriangle } from "lucide-react";
import { formatTimeDifference } from "@/lib/utils";

function LogActivity() {
  const { data: logs, isLoading } = useQuery(
    orpc.logs.getLogs.queryOptions({
      input: { page: 1, page_size: 5, days: 1 },
      staleTime: 5000,
    })
  );

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ScrollText className="h-5 w-5" />
            Recent Activity
          </CardTitle>
          <CardDescription>Latest system events</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {[...Array(5)].map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </CardContent>
      </Card>
    );
  }

  const getLevelIcon = (level: string) => {
    switch (level.toUpperCase()) {
      case "ERROR":
        return <AlertCircle className="h-4 w-4 text-destructive" />;
      case "WARN":
      case "WARNING":
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      default:
        return <Info className="h-4 w-4 text-primary" />;
    }
  };

  const getLevelBadge = (level: string) => {
    switch (level.toUpperCase()) {
      case "ERROR":
        return <Badge variant="destructive">ERROR</Badge>;
      case "WARN":
      case "WARNING":
        return <Badge className="bg-yellow-500 text-foreground">WARN</Badge>;
      default:
        return <Badge variant="secondary">{level}</Badge>;
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <ScrollText className="h-5 w-5" />
          Recent Activity
        </CardTitle>
        <CardDescription>
          {logs?.total ?? 0} events in the last 24 hours
        </CardDescription>
      </CardHeader>
      <CardContent>
        {logs?.logs && logs.logs.length > 0 ? (
          <div className="space-y-3">
            {logs.logs.map((log) => (
              <div
                key={log.id}
                className="flex items-start justify-between gap-2 rounded-lg border p-3 overflow-hidden"
              >
                <div className="flex items-start gap-2 min-w-0 flex-1">
                  <div className="shrink-0">
                    {getLevelIcon(log.level)}
                  </div>
                  <div className="space-y-1 min-w-0 flex-1">
                    <p className="text-sm font-medium truncate">{log.action}</p>
                    {log.details && (
                      <p className="text-xs text-muted-foreground truncate">
                        {log.details}
                      </p>
                    )}
                    <div className="flex items-center gap-2 text-xs text-muted-foreground flex-wrap">
                      <Badge variant="outline" className="text-xs shrink-0">
                        {log.service_group}
                      </Badge>
                      <span className="shrink-0">{formatTimeDifference(log.created_at)}</span>
                    </div>
                  </div>
                </div>
                <div className="shrink-0">
                  {getLevelBadge(log.level)}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">No recent activity</p>
        )}
      </CardContent>
    </Card>
  );
}

export default LogActivity;
