import { orpc } from "@/lib/orpc";
import { usePermission } from "@/components/protected-content";
import {
  RBAC_BACKUP_READ,
  RBAC_BACKUP_WRITE,
  RBAC_BACKUP_DELETE,
  type BackupJob,
  type BackupSchedule,
  type FirewallBackup,
} from "@/types/types.gen";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useQuery, useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { format } from "date-fns";
import {
  Archive,
  Clock,
  HardDrive,
  MoreVertical,
  Play,
  Shield,
  Trash2,
  Container,
  Server,
  RefreshCw,
} from "lucide-react";
import { toast } from "sonner";
import { useConfirmation } from "@/hooks/use-confirmation";
import { CONSTANTS } from "@/lib";

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
}

function getStatusColor(status: string): string {
  switch (status) {
    case "completed":
      return "bg-green-100 text-green-800";
    case "running":
      return "bg-blue-100 text-blue-800";
    case "pending":
      return "bg-yellow-100 text-yellow-800";
    case "failed":
      return "bg-red-100 text-red-800";
    default:
      return "bg-gray-100 text-gray-800";
  }
}

function getTypeIcon(type: string) {
  switch (type) {
    case "vm_snapshot":
    case "vm_export":
      return <Server className="h-4 w-4" />;
    case "container_export":
    case "container_commit":
      return <Container className="h-4 w-4" />;
    case "firewall":
      return <Shield className="h-4 w-4" />;
    default:
      return <Archive className="h-4 w-4" />;
  }
}

// Backup Jobs Tab Component
function BackupJobsTab() {
  const { hasPermission } = usePermission();
  const { confirm, ConfirmationDialog } = useConfirmation();

  const jobsQuery = useQuery(
    orpc.backup.listJobs.queryOptions({
      refetchInterval: CONSTANTS.POLLING_INTERVAL_MS,
    })
  );

  const deleteJobMutation = useMutation(
    orpc.backup.deleteJob.mutationOptions({
      onSuccess: () => {
        toast.success("Backup job deleted");
      },
    })
  );

  const jobs = jobsQuery.data || [];

  const handleDeleteJob = async (job: BackupJob) => {
    const confirmed = await confirm({
      title: "Delete Backup Job",
      description: `Are you sure you want to delete the backup job "${job.name}"?`,
      confirmText: "Delete",
      cancelText: "Cancel",
      variant: "destructive",
    });

    if (confirmed) {
      deleteJobMutation.mutate({ params: { id: String(job.id) } });
    }
  };

  return (
    <div className="space-y-4">
      <ConfirmationDialog />
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Target</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Progress</TableHead>
            <TableHead>Size</TableHead>
            <TableHead>Started</TableHead>
            <TableHead>Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {jobsQuery.isLoading ? (
            <TableRow>
              <TableCell colSpan={8} className="text-center py-4">
                Loading backup jobs...
              </TableCell>
            </TableRow>
          ) : jobs.length === 0 ? (
            <TableRow>
              <TableCell colSpan={8} className="text-center py-4">
                No backup jobs found
              </TableCell>
            </TableRow>
          ) : (
            jobs.map((job) => (
              <TableRow key={job.id}>
                <TableCell className="font-medium">
                  <div className="flex items-center gap-2">
                    {getTypeIcon(job.type)}
                    {job.name}
                  </div>
                </TableCell>
                <TableCell>
                  <Badge variant="outline">{job.type}</Badge>
                </TableCell>
                <TableCell>{job.target_name || job.target_id}</TableCell>
                <TableCell>
                  <Badge className={getStatusColor(job.status)}>
                    {job.status}
                  </Badge>
                </TableCell>
                <TableCell>
                  {job.status === "running" ? (
                    <div className="flex items-center gap-2">
                      <div className="w-20 h-2 bg-secondary rounded-full overflow-hidden">
                        <div
                          className="h-full bg-primary transition-all"
                          style={{ width: `${job.progress || 0}%` }}
                        />
                      </div>
                      <span className="text-sm">{job.progress || 0}%</span>
                    </div>
                  ) : (
                    "-"
                  )}
                </TableCell>
                <TableCell>
                  {job.size_bytes ? formatBytes(job.size_bytes) : "-"}
                </TableCell>
                <TableCell className="text-sm">
                  {job.started_at
                    ? format(new Date(job.started_at), "MMM dd, HH:mm")
                    : "-"}
                </TableCell>
                <TableCell>
                  {hasPermission(RBAC_BACKUP_DELETE) && (
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          className="text-destructive"
                          onClick={() => handleDeleteJob(job)}
                        >
                          <Trash2 className="h-4 w-4 mr-2" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  )}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}

// Backup Schedules Tab Component
function BackupSchedulesTab() {
  const { hasPermission } = usePermission();
  const { confirm, ConfirmationDialog } = useConfirmation();

  const schedulesQuery = useQuery(orpc.backup.listSchedules.queryOptions());

  const toggleScheduleMutation = useMutation(
    orpc.backup.toggleSchedule.mutationOptions({
      onSuccess: (data) => {
        toast.success(
          `Schedule ${data.enabled ? "enabled" : "disabled"}`
        );
      },
    })
  );

  const deleteScheduleMutation = useMutation(
    orpc.backup.deleteSchedule.mutationOptions({
      onSuccess: () => {
        toast.success("Schedule deleted");
      },
    })
  );

  const schedules = schedulesQuery.data || [];

  const handleToggleSchedule = (schedule: BackupSchedule) => {
    toggleScheduleMutation.mutate({ params: { id: String(schedule.id) } });
  };

  const handleDeleteSchedule = async (schedule: BackupSchedule) => {
    const confirmed = await confirm({
      title: "Delete Schedule",
      description: `Are you sure you want to delete the schedule "${schedule.name}"?`,
      confirmText: "Delete",
      cancelText: "Cancel",
      variant: "destructive",
    });

    if (confirmed) {
      deleteScheduleMutation.mutate({ params: { id: String(schedule.id) } });
    }
  };

  return (
    <div className="space-y-4">
      <ConfirmationDialog />
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Target</TableHead>
            <TableHead>Schedule</TableHead>
            <TableHead>Time</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Last Run</TableHead>
            <TableHead>Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {schedulesQuery.isLoading ? (
            <TableRow>
              <TableCell colSpan={8} className="text-center py-4">
                Loading schedules...
              </TableCell>
            </TableRow>
          ) : schedules.length === 0 ? (
            <TableRow>
              <TableCell colSpan={8} className="text-center py-4">
                No backup schedules configured
              </TableCell>
            </TableRow>
          ) : (
            schedules.map((schedule) => (
              <TableRow key={schedule.id}>
                <TableCell className="font-medium">
                  <div className="flex items-center gap-2">
                    {getTypeIcon(schedule.type)}
                    {schedule.name}
                  </div>
                </TableCell>
                <TableCell>
                  <Badge variant="outline">{schedule.type}</Badge>
                </TableCell>
                <TableCell>
                  {schedule.target_name || schedule.target_id}
                </TableCell>
                <TableCell className="capitalize">{schedule.schedule}</TableCell>
                <TableCell>{schedule.schedule_time || "-"}</TableCell>
                <TableCell>
                  <Badge
                    variant={schedule.enabled ? "default" : "secondary"}
                  >
                    {schedule.enabled ? "Enabled" : "Disabled"}
                  </Badge>
                </TableCell>
                <TableCell className="text-sm">
                  {schedule.last_run_at
                    ? format(new Date(schedule.last_run_at), "MMM dd, HH:mm")
                    : "Never"}
                </TableCell>
                <TableCell>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="icon">
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      {hasPermission(RBAC_BACKUP_WRITE) && (
                        <DropdownMenuItem
                          onClick={() => handleToggleSchedule(schedule)}
                        >
                          <Play className="h-4 w-4 mr-2" />
                          {schedule.enabled ? "Disable" : "Enable"}
                        </DropdownMenuItem>
                      )}
                      {hasPermission(RBAC_BACKUP_DELETE) && (
                        <DropdownMenuItem
                          className="text-destructive"
                          onClick={() => handleDeleteSchedule(schedule)}
                        >
                          <Trash2 className="h-4 w-4 mr-2" />
                          Delete
                        </DropdownMenuItem>
                      )}
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}

// Firewall Backups Tab Component
function FirewallBackupsTab() {
  const { hasPermission } = usePermission();
  const { confirm, ConfirmationDialog } = useConfirmation();

  const backupsQuery = useQuery(orpc.backup.listFirewallBackups.queryOptions());

  const deleteBackupMutation = useMutation(
    orpc.backup.deleteFirewallBackup.mutationOptions({
      onSuccess: () => {
        toast.success("Firewall backup deleted");
      },
    })
  );

  const backups = backupsQuery.data || [];

  const handleDeleteBackup = async (backup: FirewallBackup) => {
    const confirmed = await confirm({
      title: "Delete Firewall Backup",
      description: `Are you sure you want to delete the backup "${backup.filename}"?`,
      confirmText: "Delete",
      cancelText: "Cancel",
      variant: "destructive",
    });

    if (confirmed) {
      deleteBackupMutation.mutate({ params: { filename: backup.filename } });
    }
  };

  return (
    <div className="space-y-4">
      <ConfirmationDialog />
      <div className="text-sm text-muted-foreground mb-4">
        Firewall rules are automatically backed up whenever changes are made.
        Only the last 10 backups are retained.
      </div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Filename</TableHead>
            <TableHead>Rules</TableHead>
            <TableHead>Size</TableHead>
            <TableHead>Created</TableHead>
            <TableHead>Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {backupsQuery.isLoading ? (
            <TableRow>
              <TableCell colSpan={5} className="text-center py-4">
                Loading firewall backups...
              </TableCell>
            </TableRow>
          ) : backups.length === 0 ? (
            <TableRow>
              <TableCell colSpan={5} className="text-center py-4">
                No firewall backups found
              </TableCell>
            </TableRow>
          ) : (
            backups.map((backup) => (
              <TableRow key={backup.id}>
                <TableCell className="font-medium">
                  <div className="flex items-center gap-2">
                    <Shield className="h-4 w-4" />
                    {backup.filename}
                  </div>
                </TableCell>
                <TableCell>{backup.rule_count} rules</TableCell>
                <TableCell>{formatBytes(backup.size_bytes)}</TableCell>
                <TableCell className="text-sm">
                  {format(new Date(backup.created_at), "MMM dd, yyyy HH:mm")}
                </TableCell>
                <TableCell>
                  {hasPermission(RBAC_BACKUP_DELETE) && (
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          className="text-destructive"
                          onClick={() => handleDeleteBackup(backup)}
                        >
                          <Trash2 className="h-4 w-4 mr-2" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  )}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}

export default function BackupsPage() {
  const { hasPermission } = usePermission();
  const [activeTab, setActiveTab] = useState("jobs");

  const statsQuery = useQuery(orpc.backup.getStats.queryOptions());

  // Check permission
  if (!hasPermission(RBAC_BACKUP_READ)) {
    return (
      <div className="flex items-center justify-center p-8">
        <p className="text-muted-foreground">
          You don't have permission to access this page.
        </p>
      </div>
    );
  }

  const stats = statsQuery.data;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Backups</h1>
        <Button
          variant="outline"
          size="sm"
          onClick={() => statsQuery.refetch()}
          disabled={statsQuery.isFetching}
        >
          <RefreshCw
            className={`h-4 w-4 mr-2 ${statsQuery.isFetching ? "animate-spin" : ""}`}
          />
          Refresh
        </Button>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium flex items-center gap-2">
                <Archive className="h-4 w-4" />
                Total Backups
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total_backups}</div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium flex items-center gap-2">
                <Server className="h-4 w-4" />
                VM Snapshots
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.vm_snapshots}</div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium flex items-center gap-2">
                <Container className="h-4 w-4" />
                Container Backups
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {stats.container_backups}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium flex items-center gap-2">
                <Shield className="h-4 w-4" />
                Firewall Backups
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.firewall_backups}</div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium flex items-center gap-2">
                <HardDrive className="h-4 w-4" />
                Total Size
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {formatBytes(stats.total_size_bytes)}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium flex items-center gap-2">
                <Clock className="h-4 w-4" />
                Active Schedules
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.active_schedules}</div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Tabs */}
      <Card>
        <CardContent className="pt-6">
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="grid w-full grid-cols-3 max-w-md">
              <TabsTrigger value="jobs">Backup Jobs</TabsTrigger>
              <TabsTrigger value="schedules">Schedules</TabsTrigger>
              <TabsTrigger value="firewall">Firewall</TabsTrigger>
            </TabsList>
            <TabsContent value="jobs" className="mt-4">
              <BackupJobsTab />
            </TabsContent>
            <TabsContent value="schedules" className="mt-4">
              <BackupSchedulesTab />
            </TabsContent>
            <TabsContent value="firewall" className="mt-4">
              <FirewallBackupsTab />
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}
