import { useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { toast } from "sonner";
import { useFormGenerator } from "@/hooks";
import { Z } from "@/types";
import { Button } from "@/components/ui/button";

interface CreateUserDialogContentProps {
  onClose: () => void;
}

export function CreateUserDialogContent({ onClose }: CreateUserDialogContentProps) {
  const createUserMutation = useMutation(
    orpc.users.createUser.mutationOptions({
      onSuccess() {
        toast.success("User created successfully");
        onClose();
      },
      onError() {
        toast.error("Failed to create user");
      },
    }),
  );

  const CreateForm = useFormGenerator(Z.createUserParamsSchema, {
    onSubmit(data) {
      createUserMutation.mutate({
        username: data.username,
        email: data.email,
        password: data.password,
        role: data.role || "user",
      });
    },
  });

  return (
    <CreateForm.parts.wrapper>
      <CreateForm.parts.errors />
      {CreateForm.parts.fields}
      <div className="flex gap-2 pt-4">
        <Button variant="secondary" onClick={onClose}>
          Cancel
        </Button>
        <CreateForm.parts.submitButton disabled={createUserMutation.isPending}>
          {createUserMutation.isPending ? "Creating..." : "Create"}
        </CreateForm.parts.submitButton>
      </div>
    </CreateForm.parts.wrapper>
  );
}
