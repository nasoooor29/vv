import { useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { toast } from "sonner";
import { useFormGenerator } from "@/hooks";
import { Z } from "@/types";
import type { T } from "@/types";
import { Button } from "@/components/ui/button";
import { useEffect } from "react";
import { allPolicies } from "@/lib";
import z from "zod";

interface EditUserDialogContentProps {
  user: T.User | null;
  onClose: () => void;
}

// Schema defined outside component to maintain stable reference
const editUserSchema = Z.updateUserParamsSchema.omit({ role: true }).extend({
  roles: z.enum(allPolicies).array(),
});

export function EditUserDialogContent({
  user,
  onClose,
}: EditUserDialogContentProps) {
  const updateUserMutation = useMutation(
    orpc.users.updateUser.mutationOptions({
      onSuccess() {
        toast.success("User updated successfully");
        onClose();
      },
      onError() {
        toast.error("Failed to update user");
      },
    }),
  );

  const EditForm = useFormGenerator(editUserSchema, {
    defaultValues: user
      ? {
          username: user.username,
          email: user.email,
          roles: user.role.split(","),
          id: user.id,
        }
      : undefined,
    onSubmit(data) {
      if (user) {
        updateUserMutation.mutate({
          params: { id: String(user.id) },
          body: {
            username: data.username,
            email: data.email,
            role: data.roles.join(","),
          },
        });
      }
    },
  });

  // Update form when user changes
  useEffect(() => {
    if (user) {
      EditForm.form.reset({
        username: user.username,
        email: user.email,
        role: user.role,
        id: user.id,
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  if (!user) return null;

  return (
    <EditForm.parts.wrapper>
      <EditForm.parts.errors />
      {EditForm.parts.fieldsList.map((field) => (
        <div key={field} className={`mb-4${field === "id" ? " hidden" : ""}`}>
          <EditForm.parts.field name={field} />
        </div>
      ))}
      <div className="flex gap-2 pt-4">
        <Button variant="secondary" onClick={onClose}>
          Cancel
        </Button>
        <EditForm.parts.submitButton disabled={updateUserMutation.isPending}>
          {updateUserMutation.isPending ? "Saving..." : "Save"}
        </EditForm.parts.submitButton>
      </div>
    </EditForm.parts.wrapper>
  );
}
