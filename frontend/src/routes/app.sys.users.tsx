import { useQuery, useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { usePermission } from "@/components/protected-content";
import { RBAC_USER_ADMIN } from "@/types/types.gen";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { format } from "date-fns";
import { toast } from "sonner";
import { useState } from "react";
import MultipleSelector from "@/components/ui/multi-select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Trash2, Edit, Shield } from "lucide-react";

// Available roles to choose from
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

// Helper function to split role string into array
function parseRoles(roleString: string): typeof AVAILABLE_ROLES {
  if (!roleString) return [];
  return roleString
    .split(",")
    .map((role) => role.trim())
    .filter((role) => role.length > 0)
    .map((role) => ({
      value: role,
      label: AVAILABLE_ROLES.find((r) => r.value === role)?.label || role,
    }));
}

// Helper function to merge roles array back to string
function mergeRoles(roles: typeof AVAILABLE_ROLES): string {
  return roles.map((r) => r.value).join(", ");
}

export default function UsersPage() {
  const { hasPermission } = usePermission();
  const [editingUser, setEditingUser] = useState<any>(null);
  const [deletingUserId, setDeletingUserId] = useState<number | null>(null);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [rolePromptUser, setRolePromptUser] = useState<any>(null);
  const [selectedRoles, setSelectedRoles] = useState<typeof AVAILABLE_ROLES>(
    [],
  );

  const usersQuery = useQuery(orpc.users.listUsers.queryOptions());

  const updateUserMutation = useMutation(
    orpc.users.updateUser.mutationOptions({
      onSuccess() {
        toast.success("User updated successfully");
        usersQuery.refetch();
        setShowEditDialog(false);
        setEditingUser(null);
      },
      onError() {
        toast.error("Failed to update user");
      },
    }),
  );

  const deleteUserMutation = useMutation(
    orpc.users.deleteUser.mutationOptions({
      onSuccess() {
        toast.success("User deleted successfully");
        usersQuery.refetch();
        setShowDeleteDialog(false);
        setDeletingUserId(null);
      },
    }),
  );

  const updateRoleMutation = useMutation(
    orpc.users.updateUserRole.mutationOptions({
      onSuccess() {
        toast.success("User role updated successfully");
        usersQuery.refetch();
        setRolePromptUser(null);
        setSelectedRoles([]);
      },
    }),
  );

  // Check permission
  if (!hasPermission(RBAC_USER_ADMIN)) {
    return (
      <div className="flex items-center justify-center p-8">
        <p className="text-muted-foreground">
          You don't have permission to access this page.
        </p>
      </div>
    );
  }

  const handleEdit = (user: any) => {
    setEditingUser({ ...user });
    setShowEditDialog(true);
  };

  const handleSaveEdit = () => {
    if (editingUser) {
      updateUserMutation.mutate({
        id: editingUser.id,
        username: editingUser.username,
        email: editingUser.email,
        role: editingUser.role || "user",
      });
    }
  };

  const handleDelete = (userId: number) => {
    setDeletingUserId(userId);
    setShowDeleteDialog(true);
  };

  const handleConfirmDelete = () => {
    if (deletingUserId) {
      deleteUserMutation.mutate({ id: deletingUserId });
    }
  };

  const handlePromoteClick = (user: any) => {
    setRolePromptUser(user);
    const parsedRoles = parseRoles(user.role || "");
    setSelectedRoles(parsedRoles);
  };

  const handleSaveRole = () => {
    if (rolePromptUser && selectedRoles.length > 0) {
      const rolesString = mergeRoles(selectedRoles);
      updateRoleMutation.mutate({
        id: rolePromptUser.id,
        role: rolesString,
      });
    } else if (rolePromptUser) {
      // If no roles selected, set to "user"
      updateRoleMutation.mutate({
        id: rolePromptUser.id,
        role: "user",
      });
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Users Management</h1>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Users</CardTitle>
        </CardHeader>
        <CardContent>
          {usersQuery.isLoading ? (
            <div className="space-y-2">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : usersQuery.error ? (
            <div className="text-center py-8 text-destructive">
              Failed to load users. Please try again.
            </div>
          ) : !usersQuery.data || usersQuery.data.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No users found.
            </div>
          ) : (
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
                  {usersQuery.data.map((user) => {
                    const userRoles = parseRoles(user.role || "");
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
                                  key={role.value}
                                  variant="secondary"
                                  className="text-xs"
                                >
                                  {role.label}
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
                            onClick={() => handleEdit(user)}
                            disabled={updateUserMutation.isPending}
                          >
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handlePromoteClick(user)}
                            disabled={updateRoleMutation.isPending}
                          >
                            <Shield className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => handleDelete(user.id)}
                            disabled={deleteUserMutation.isPending}
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
          )}
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit User</DialogTitle>
            <DialogDescription>Update user information below</DialogDescription>
          </DialogHeader>
          {editingUser && (
            <div className="space-y-4">
              <div>
                <Label>Username</Label>
                <Input
                  value={editingUser.username}
                  onChange={(e) =>
                    setEditingUser({
                      ...editingUser,
                      username: e.target.value,
                    })
                  }
                />
              </div>
              <div>
                <Label>Email</Label>
                <Input
                  value={editingUser.email}
                  onChange={(e) =>
                    setEditingUser({ ...editingUser, email: e.target.value })
                  }
                />
              </div>
              <div>
                <Label>Roles</Label>
                <MultipleSelector
                  value={parseRoles(editingUser.role || "")}
                  onChange={(roles) => {
                    setEditingUser({
                      ...editingUser,
                      role: mergeRoles(roles),
                    });
                  }}
                  defaultOptions={AVAILABLE_ROLES}
                  placeholder="Select roles..."
                  emptyIndicator={
                    <p className="text-center text-sm">No roles found</p>
                  }
                />
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowEditDialog(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleSaveEdit}
              disabled={updateUserMutation.isPending}
            >
              {updateUserMutation.isPending ? "Saving..." : "Save"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete User</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this user? This action cannot be
              undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmDelete}
              disabled={deleteUserMutation.isPending}
            >
              {deleteUserMutation.isPending ? "Deleting..." : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Role Selection Dialog */}
      <Dialog
        open={!!rolePromptUser}
        onOpenChange={(open) => !open && setRolePromptUser(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Update User Roles</DialogTitle>
            <DialogDescription>
              Select roles for {rolePromptUser?.username}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Roles</Label>
              <MultipleSelector
                value={selectedRoles}
                onChange={setSelectedRoles}
                defaultOptions={AVAILABLE_ROLES}
                placeholder="Select roles..."
                emptyIndicator={
                  <p className="text-center text-sm">No roles found</p>
                }
              />
              <p className="text-xs text-muted-foreground mt-2">
                Select at least one role. If no role is selected, the user will
                be assigned the 'User' role.
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRolePromptUser(null)}>
              Cancel
            </Button>
            <Button
              onClick={handleSaveRole}
              disabled={updateRoleMutation.isPending}
            >
              {updateRoleMutation.isPending ? "Updating..." : "Update"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
