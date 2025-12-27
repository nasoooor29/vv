import { useQuery, useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Loader2, Server, Box, Plus } from "lucide-react";
import { useState } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useConfirmation, useDialog } from "@/hooks";
import { toast } from "sonner";
import { ContainersGrid } from "./containers-grid";
import { ImagesCard } from "./images-card";
import { usePermission } from "@/components/protected-content";
import { RBAC_DOCKER_READ } from "@/types/types.gen";
import type { z } from "zod";
import type { Z } from "@/types";
import { ContainerDetailsContent } from "./container-details";
import { CreateContainerContent } from "./create-container-dialog";

type DockerContainer = z.infer<typeof Z.dockerContainerSchema>;

export default function DockerPage() {
  const { hasPermission } = usePermission();
  const { confirm, ConfirmationDialog } = useConfirmation();
  const [selectedClient, setSelectedClient] = useState<string | null>(null);
  const [selectedContainer, setSelectedContainer] =
    useState<DockerContainer | null>(null);

  // Fetch available docker clients
  const clientsQuery = useQuery(orpc.docker.clients.queryOptions({}));

  const deleteContainerMutation = useMutation(
    orpc.docker.deleteContainer.mutationOptions({
      onSuccess() {
        toast.success("Container deleted successfully");
      },
      onError() {
        toast.error("Failed to delete container");
      },
    }),
  );

  // Auto-select first client when loaded
  const clients = clientsQuery.data ?? [];
  const activeClientId = selectedClient ?? clients[0]?.id?.toString() ?? null;

  // Dialog hooks at page level - stable across re-renders
  const containerDetailsDialog = useDialog({
    className: "max-w-4xl",
  });

  const createContainerDialog = useDialog({
    title: "Create New Container",
    description: "Configure and create a new Docker container",
    className: "sm:max-w-lg",
  });

  // Check permission
  if (!hasPermission(RBAC_DOCKER_READ)) {
    return (
      <div className="flex items-center justify-center p-8">
        <p className="text-muted-foreground">
          You don't have permission to access Docker management.
        </p>
      </div>
    );
  }

  if (clientsQuery.isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  if (clientsQuery.error) {
    return (
      <div className="p-4 text-destructive">
        Error loading Docker clients. Make sure Docker is running.
      </div>
    );
  }

  if (clients.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>No Docker Clients Available</CardTitle>
        </CardHeader>
        <CardContent className="text-muted-foreground">
          No Docker daemon connections are available. Make sure Docker is
          running and properly configured.
        </CardContent>
      </Card>
    );
  }

  const handleViewDetails = (container: DockerContainer) => {
    setSelectedContainer(container);
    containerDetailsDialog.open();
  };

  const handleDelete = (container: DockerContainer) => {
    confirm({
      title: "Delete Container",
      description: `Are you sure you want to delete container "${container.Names[0]?.slice(1)}"? This action cannot be undone.`,
      isDestructive: true,
      onConfirm: async () => {
        await deleteContainerMutation.mutateAsync({
          params: { clientId: activeClientId!, id: container.Id },
        });
      },
    });
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Box className="h-6 w-6" />
          <h1 className="text-3xl font-bold">Container Management</h1>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Server className="h-4 w-4 text-muted-foreground" />
            <Select
              value={activeClientId ?? undefined}
              onValueChange={setSelectedClient}
            >
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Select client" />
              </SelectTrigger>
              <SelectContent>
                {clients.map((client) => (
                  <SelectItem key={client.id} value={client.id.toString()}>
                    <div className="flex items-center gap-2">
                      <span>Client {client.id}</span>
                      <Badge
                        variant={
                          client.status === "connected"
                            ? "default"
                            : "destructive"
                        }
                        className="text-xs"
                      >
                        {client.status}
                      </Badge>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <Button onClick={createContainerDialog.open}>
            <Plus className="mr-2 h-4 w-4" />
            Create Container
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Box className="h-5 w-5" />
            Containers
          </CardTitle>
        </CardHeader>
        <CardContent>
          {activeClientId && (
            <ContainersGrid
              clientId={activeClientId}
              onViewDetails={handleViewDetails}
              onDelete={handleDelete}
            />
          )}
        </CardContent>
      </Card>

      {activeClientId && <ImagesCard clientId={activeClientId} />}

      {/* Container Details Dialog */}
      <containerDetailsDialog.Component
        title={selectedContainer?.Names[0]?.slice(1) || "Container Details"}
        description={selectedContainer?.Image}
      >
        <ContainerDetailsContent
          clientId={activeClientId}
          container={selectedContainer}
        />
      </containerDetailsDialog.Component>

      {/* Create Container Dialog */}
      <createContainerDialog.Component>
        {(close) => (
          <CreateContainerContent
            clientId={activeClientId}
            onClose={close}
          />
        )}
      </createContainerDialog.Component>

      <ConfirmationDialog />
    </div>
  );
}
