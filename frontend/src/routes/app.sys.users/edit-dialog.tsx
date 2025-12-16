import { useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { toast } from "sonner";
import { useFormGenerator, useDialog } from "@/hooks";
import { Z, T } from "@/types";
import { Button } from "@/components/ui/button";

export function EditUserDialog(user: T.User | null) {
  const updateUserMutation = useMutation(
    orpc.users.updateUser.mutationOptions({
      onSuccess() {
        toast.success("User updated successfully");
      },
      onError() {
        toast.error("Failed to update user");
      },
    }),
  );

  const EditForm = useFormGenerator(Z.updateUserParamsSchema, {
    defaultValues: user
      ? {
          username: user.username,
          email: user.email,
          role: user.role,
          id: user.id,
        }
      : undefined,
    onSubmit(data) {
      if (user) {
        updateUserMutation.mutate(
          {
            params: { id: String(user.id) },
            body: {
              username: data.username,
              email: data.email,
              role: data.role,
            },
          },
          {
            onSuccess: () => {
              dialog.close();
            },
          },
        );
      }
    },
  });

  const dialog = useDialog({
    title: "Edit User",
    description: "Update user information below",
    children: user ? (
      <EditForm.parts.wrapper>
        <EditForm.parts.errors />
        {EditForm.parts.fields}
        <div className="flex gap-2 pt-4">
          <Button variant="secondary" onClick={() => dialog.close()}>
            Cancel
          </Button>
          <EditForm.parts.submitButton disabled={updateUserMutation.isPending}>
            {updateUserMutation.isPending ? "Saving..." : "Save"}
          </EditForm.parts.submitButton>
        </div>
      </EditForm.parts.wrapper>
    ) : null,
  });

  return {
    dialog: dialog,
    form: EditForm.form,
    mutation: updateUserMutation,
  };
}
