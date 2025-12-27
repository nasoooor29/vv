import { format } from "date-fns";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Trash2, Edit, Shield } from "lucide-react";
import type { T } from "@/types";
import { roleStringToArray } from "@/lib/rbac";
import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";

interface UsersTableProps {
  // users: T.User[] | undefined;
  // isLoading: boolean;
  // isError: boolean;
  onEdit: (user: T.User) => void;
  onManageRoles: (user: T.User) => void;
  onDelete: (user: T.User) => void;
  // isUpdatePending: boolean;
  // isDeletePending: boolean;
}

const AVAILABLE_ROLES = [
  { value: "user", label: "User" },
  { value: "user_admin", label: "User Admin" },
  { value: "docker_read", label: "Docker Read" },
  { value: "docker_write", label: "Docker Write" },
  { value: "docker_update", label: "Docker Update" },
  { value: "docker_delete", label: "Docker Delete" },
  { value: "qemu_read", label: "QEMU Read" },
  { value: "qemu_write", label: "QEMU Write" },
  { value: "qemu_update", label: "QEMU Update" },
  { value: "qemu_delete", label: "QEMU Delete" },
  { value: "event_viewer", label: "Event Viewer" },
  { value: "event_manager", label: "Event Manager" },
  { value: "settings_manager", label: "Settings Manager" },
  { value: "audit_log_viewer", label: "Audit Log Viewer" },
  { value: "health_checker", label: "Health Checker" },
];

function getRoleLabel(roleValue: string): string {
  return AVAILABLE_ROLES.find((r) => r.value === roleValue)?.label || roleValue;
}

export function UsersTable({
  onEdit,
  onManageRoles,
  onDelete,
  // isUpdatePending,
  // isDeletePending,
}: UsersTableProps) {
  const {
    data: users,
    isLoading,
    isError,
  } = useQuery(orpc.users.listUsers.queryOptions());

  if (isLoading) {
    return (
      <div className="space-y-2">
        {[1, 2, 3].map((i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <div className="text-center py-8 text-destructive">
        Failed to load users. Please try again.
      </div>
    );
  }

  if (!users || users.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No users found.
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>ID</TableHead>
            <TableHead>Username</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Roles</TableHead>
            <TableHead>Created At</TableHead>
            <TableHead>Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {users.map((user) => {
            const userRoles = roleStringToArray(user.role);
            return (
              <TableRow key={user.id}>
                <TableCell className="font-medium">{user.id}</TableCell>
                <TableCell>{user.username}</TableCell>
                <TableCell>{user.email}</TableCell>
                <TableCell>
                  <div className="flex flex-wrap gap-1">
                    {userRoles.length > 0 ? (
                      userRoles.map((role) => (
                        <Badge
                          key={role}
                          variant="secondary"
                          className="text-xs"
                        >
                          {getRoleLabel(role)}
                        </Badge>
                      ))
                    ) : (
                      <Badge variant="outline" className="text-xs">
                        No roles
                      </Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell className="text-sm text-muted-foreground">
                  {format(new Date(user.created_at), "PPpp")}
                </TableCell>
                <TableCell className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onEdit(user)}
                    // disabled={isUpdatePending}
                  >
                    <Edit className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onManageRoles(user)}
                    // disabled={isUpdatePending}
                  >
                    <Shield className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() => onDelete(user)}
                    // disabled={isDeletePending}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}
