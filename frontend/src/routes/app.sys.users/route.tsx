import { useState } from "react";
import { orpc } from "@/lib/orpc";
import { usePermission } from "@/components/protected-content";
import { RBAC_USER_ADMIN } from "@/types/types.gen";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { useConfirmation, useDialog } from "@/hooks";
import { T } from "@/types";
import { UsersTable } from "./table";
import { CreateUserDialogContent } from "./create-dialog";
import { EditUserDialogContent } from "./edit-dialog";
import { ManageRolesDialogContent } from "./roles-dialog";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";

export default function UsersPage() {
  const { hasPermission: checkPermission } = usePermission();
  const { confirm, ConfirmationDialog } = useConfirmation();
  const [selectedUser, setSelectedUser] = useState<T.User | null>(null);

  const deleteUserMutation = useMutation(
    orpc.users.deleteUser.mutationOptions({
      onSuccess() {
        toast.success("User deleted successfully");
      },
      onError() {
        toast.error("Failed to delete user");
      },
    }),
  );

  const createDialog = useDialog({
    title: "Create New User",
    description: "Add a new user to the system",
  });

  const editDialog = useDialog({
    title: "Edit User",
    description: "Update user information below",
  });

  const rolesDialog = useDialog();

  // Check permission
  if (!checkPermission(RBAC_USER_ADMIN)) {
    return (
      <div className="flex items-center justify-center p-8">
        <p className="text-muted-foreground">
          You don't have permission to access this page.
        </p>
      </div>
    );
  }

  const handleEdit = (user: T.User) => {
    setSelectedUser(user);
    editDialog.open();
  };

  const handleManageRoles = (user: T.User) => {
    setSelectedUser(user);
    rolesDialog.open();
  };

  const handleDelete = (user: T.User) => {
    confirm({
      title: "Delete User",
      description: `Are you sure you want to delete ${user.username}? This action cannot be undone.`,
      isDestructive: true,
      onConfirm: async () => {
        await deleteUserMutation.mutateAsync({
          params: { id: String(user.id) },
        });
      },
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Users Management</h1>
        <Button onClick={createDialog.open}>
          <Plus className="h-4 w-4 mr-2" />
          Create User
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Users</CardTitle>
        </CardHeader>
        <CardContent>
          <UsersTable
            onEdit={handleEdit}
            onManageRoles={handleManageRoles}
            onDelete={handleDelete}
          />
        </CardContent>
      </Card>

      <createDialog.Component>
        {(close) => <CreateUserDialogContent onClose={close} />}
      </createDialog.Component>

      <editDialog.Component>
        {(close) => <EditUserDialogContent user={selectedUser} onClose={close} />}
      </editDialog.Component>

      <rolesDialog.Component
        title="Update User Roles"
        description={`Select roles for ${selectedUser?.username || ""}`}
      >
        {(close) => <ManageRolesDialogContent user={selectedUser} onClose={close} />}
      </rolesDialog.Component>

      <ConfirmationDialog />
    </div>
  );
}
