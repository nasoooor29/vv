import { useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { toast } from "sonner";
import { useFormGenerator, useDialog } from "@/hooks";
import { Z } from "@/types";
import { Button } from "@/components/ui/button";

export function CreateUserDialog() {
  const createUserMutation = useMutation(
    orpc.users.createUser.mutationOptions({
      onSuccess() {
        toast.success("User created successfully");
        dialog.close();
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
  const dialog = useDialog({
    title: "Create New User",
    description: "Add a new user to the system",
    children: (
      <CreateForm.parts.wrapper>
        <CreateForm.parts.errors />
        {CreateForm.parts.fields}
        <div className="flex gap-2 pt-4">
          <Button variant="secondary" onClick={() => dialog.close()}>
            Cancel
          </Button>
          <CreateForm.parts.submitButton
            disabled={createUserMutation.isPending}
          >
            {createUserMutation.isPending ? "Creating..." : "Create"}
          </CreateForm.parts.submitButton>
        </div>
      </CreateForm.parts.wrapper>
    ),
  });

  return {
    dialog: dialog,
    mutation: createUserMutation,
    form: CreateForm,
  };
}
