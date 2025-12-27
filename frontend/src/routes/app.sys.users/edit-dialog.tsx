import { useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { toast } from "sonner";
import { useFormGenerator } from "@/hooks";
import { Z, T } from "@/types";
import { Button } from "@/components/ui/button";
import { useEffect } from "react";

interface EditUserDialogContentProps {
  user: T.User | null;
  onClose: () => void;
}

export function EditUserDialogContent({ user, onClose }: EditUserDialogContentProps) {
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
        updateUserMutation.mutate({
          params: { id: String(user.id) },
          body: {
            username: data.username,
            email: data.email,
            role: data.role,
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
  }, [user, EditForm.form]);

  if (!user) return null;

  return (
    <EditForm.parts.wrapper>
      <EditForm.parts.errors />
      {EditForm.parts.fields}
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
