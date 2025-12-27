import { useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { toast } from "sonner";
import { T } from "@/types";
import { Button } from "@/components/ui/button";
import MultipleSelector from "@/components/ui/multi-select";
import { roleStringToArray, roleArrayToString } from "@/lib/rbac";
import { useState, useEffect } from "react";

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

interface ManageRolesDialogContentProps {
  user: T.User | null;
  onClose: () => void;
}

export function ManageRolesDialogContent({ user, onClose }: ManageRolesDialogContentProps) {
  const [selectedRoles, setSelectedRoles] = useState<string[]>([]);

  // Update selected roles when user changes
  useEffect(() => {
    if (user) {
      setSelectedRoles(roleStringToArray(user.role));
    }
  }, [user]);

  const updateRoleMutation = useMutation(
    orpc.users.updateUserRole.mutationOptions({
      onSuccess() {
        toast.success("User role updated successfully");
        onClose();
      },
      onError() {
        toast.error("Failed to update user role");
      },
    }),
  );

  if (!user) return null;

  return (
    <div className="space-y-4">
      <MultipleSelector
        value={selectedRoles.map((role) => ({
          value: role,
          label: AVAILABLE_ROLES.find((r) => r.value === role)?.label || role,
        }))}
        onChange={(selected) => {
          setSelectedRoles(selected.map((s) => s.value));
        }}
        defaultOptions={AVAILABLE_ROLES}
        placeholder="Select roles..."
        emptyIndicator={<p className="text-center text-sm">No roles found</p>}
      />
      <p className="text-xs text-muted-foreground">
        Select at least one role. If no role is selected, the user will be
        assigned the 'User' role.
      </p>
      <div className="flex gap-2 pt-4">
        <Button variant="secondary" onClick={onClose}>
          Cancel
        </Button>
        <Button
          onClick={() => {
            const rolesString =
              selectedRoles.length > 0
                ? roleArrayToString(selectedRoles)
                : "user";
            updateRoleMutation.mutate({
              params: { id: String(user.id) },
              body: { role: rolesString },
            });
          }}
          disabled={updateRoleMutation.isPending}
        >
          {updateRoleMutation.isPending ? "Updating..." : "Update"}
        </Button>
      </div>
    </div>
  );
}
