import { orpc } from "@/lib/orpc";
import { exportData, type ExportFormat } from "@/lib/export";
import { usePermission } from "@/components/protected-content";
import { RBAC_AUDIT_LOG_VIEWER, type LogResponse } from "@/types/types.gen";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { format } from "date-fns";
import { Download } from "lucide-react";

function exportLogs(logs: LogResponse[], exportFormat: ExportFormat): void {
  exportData(logs as unknown as Record<string, unknown>[], exportFormat, {
    filename: "logs",
    headers: ["id", "created_at", "service_group", "level", "action", "details", "user_id"],
    rowMapper: (log) => [
      log.id as number,
      log.created_at as string,
      log.service_group as string,
      log.level as string,
      log.action as string,
      (log.details as string) || "",
      (log.user_id as number) || "",
    ],
  });
}

interface LogFilters {
  service_group?: string;
  level?: string;
  page: number;
  page_size: number;
  days: number;
}

export default function LogsPage() {
  const { hasPermission: checkPermission } = usePermission();
  const [filters, setFilters] = useState<LogFilters>({
    page: 1,
    page_size: 20,
    days: 7,
  });

  // Fetch logs
  const logsQuery = useQuery(
    orpc.logs.getLogs.queryOptions({
      input: {
        ServiceGroup: filters.service_group || "",
        Level: filters.level || "",
        Page: filters.page,
        PageSize: filters.page_size,
        Days: filters.days,
      },
    }),
  );

  // Fetch log stats for filter options
  const statsQuery = useQuery(
    orpc.logs.getLogStats.queryOptions({
      input: {
        days: filters.days,
      },
    }),
  );

  // Check permission
  if (!checkPermission(RBAC_AUDIT_LOG_VIEWER)) {
    return (
      <div className="flex items-center justify-center p-8">
        <p className="text-muted-foreground">
          You don't have permission to access this page.
        </p>
      </div>
    );
  }

  const logs = logsQuery.data?.logs || [];
  const total = logsQuery.data?.total || 0;
  const totalPages = logsQuery.data?.total_pages || 0;
  const stats = statsQuery.data;

  const levelColors: Record<string, string> = {
    DEBUG: "bg-blue-100 text-blue-800",
    INFO: "bg-green-100 text-green-800",
    WARN: "bg-yellow-100 text-yellow-800",
    WARNING: "bg-yellow-100 text-yellow-800",
    ERROR: "bg-red-100 text-red-800",
  };

  const handleFilterChange = (
    key: keyof LogFilters,
    value: string | number,
  ) => {
    setFilters((prev) => ({
      ...prev,
      [key]: value,
      page: 1, // Reset to first page when filters change
    }));
  };

  const handlePageChange = (newPage: number) => {
    setFilters((prev) => ({
      ...prev,
      page: newPage,
    }));
  };

  const levelColor = (level: string) =>
    levelColors[level] || "bg-gray-100 text-gray-800";

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Audit Logs</h1>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" disabled={logs.length === 0}>
              <Download className="mr-2 h-4 w-4" />
              Export
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>Export Format</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => exportLogs(logs, "csv")}>
              CSV (.csv)
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => exportLogs(logs, "json")}>
              JSON (.json)
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => exportLogs(logs, "xml")}>
              XML (.xml)
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-4 gap-4">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">Total Logs</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total}</div>
              <p className="text-xs text-muted-foreground">
                Last {stats.days} days
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">Services</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {stats.service_groups?.length || 0}
              </div>
              <p className="text-xs text-muted-foreground">Active services</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">Log Levels</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {stats.levels?.length || 0}
              </div>
              <p className="text-xs text-muted-foreground">Unique levels</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">Shown</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{logs.length}</div>
              <p className="text-xs text-muted-foreground">of {total} total</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle>Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-4 gap-4">
            <div>
              <label className="text-sm font-medium">Service Group</label>
               <Select
                 value={filters.service_group === "" || !filters.service_group ? "all" : filters.service_group}
                 onValueChange={(value) =>
                   handleFilterChange("service_group", value === "all" ? "" : value)
                 }
               >
                <SelectTrigger className="mt-2">
                  <SelectValue placeholder="All Services" />
                </SelectTrigger>
                <SelectContent>
                   <SelectItem value="all">All Services</SelectItem>
                   {stats?.service_groups?.map((service) => (
                     <SelectItem key={service} value={service}>
                       {service}
                     </SelectItem>
                   ))}
                 </SelectContent>
              </Select>
            </div>

            <div>
              <label className="text-sm font-medium">Log Level</label>
               <Select
                 value={filters.level === "" || !filters.level ? "all" : filters.level}
                 onValueChange={(value) => handleFilterChange("level", value === "all" ? "" : value)}
               >
                <SelectTrigger className="mt-2">
                  <SelectValue placeholder="All Levels" />
                </SelectTrigger>
                <SelectContent>
                   <SelectItem value="all">All Levels</SelectItem>
                   {stats?.levels?.map((level) => (
                     <SelectItem key={level} value={level}>
                       {level}
                     </SelectItem>
                   ))}
                 </SelectContent>
              </Select>
            </div>

            <div>
              <label className="text-sm font-medium">Time Range (Days)</label>
              <Select
                value={String(filters.days)}
                onValueChange={(value) =>
                  handleFilterChange("days", parseInt(value))
                }
              >
                <SelectTrigger className="mt-2">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">Last 1 day</SelectItem>
                  <SelectItem value="7">Last 7 days</SelectItem>
                  <SelectItem value="30">Last 30 days</SelectItem>
                  <SelectItem value="90">Last 90 days</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <label className="text-sm font-medium">Page Size</label>
              <Select
                value={String(filters.page_size)}
                onValueChange={(value) =>
                  handleFilterChange("page_size", parseInt(value))
                }
              >
                <SelectTrigger className="mt-2">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="10">10</SelectItem>
                  <SelectItem value="20">20</SelectItem>
                  <SelectItem value="50">50</SelectItem>
                  <SelectItem value="100">100</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Logs Table */}
      <Card>
        <CardHeader>
          <CardTitle>Logs</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Timestamp</TableHead>
                  <TableHead>Service</TableHead>
                  <TableHead>Level</TableHead>
                  <TableHead>Action</TableHead>
                  <TableHead>Details</TableHead>
                  <TableHead>User ID</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {logsQuery.isLoading ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-4">
                      Loading logs...
                    </TableCell>
                  </TableRow>
                ) : logs.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-4">
                      No logs found
                    </TableCell>
                  </TableRow>
                ) : (
                  logs.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell className="whitespace-nowrap text-sm">
                        {format(new Date(log.created_at), "MMM dd, HH:mm:ss")}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{log.service_group}</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge className={levelColor(log.level)}>
                          {log.level}
                        </Badge>
                      </TableCell>
                      <TableCell className="font-medium">
                        {log.action}
                      </TableCell>
                      <TableCell className="max-w-xs truncate text-sm">
                        {log.details || "-"}
                      </TableCell>
                      <TableCell className="text-sm">
                        {log.user_id || "-"}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="mt-4 flex justify-center">
              <Pagination>
                <PaginationContent>
                  <PaginationItem>
                    <PaginationPrevious
                      onClick={() => {
                        if (filters.page > 1) {
                          handlePageChange(filters.page - 1);
                        }
                      }}
                      className={
                        filters.page === 1
                          ? "pointer-events-none opacity-50"
                          : ""
                      }
                    />
                  </PaginationItem>

                  {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
                    const pageNum = i + 1;
                    return (
                      <PaginationItem key={pageNum}>
                        <PaginationLink
                          onClick={() => handlePageChange(pageNum)}
                          isActive={pageNum === filters.page}
                        >
                          {pageNum}
                        </PaginationLink>
                      </PaginationItem>
                    );
                  })}

                  <PaginationItem>
                    <PaginationNext
                      onClick={() => {
                        if (filters.page < totalPages) {
                          handlePageChange(filters.page + 1);
                        }
                      }}
                      className={
                        filters.page === totalPages
                          ? "pointer-events-none opacity-50"
                          : ""
                      }
                    />
                  </PaginationItem>
                </PaginationContent>
              </Pagination>
            </div>
          )}

          <p className="text-sm text-muted-foreground mt-4">
            Showing {logs.length} of {total} logs (Page {filters.page} of{" "}
            {totalPages})
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
