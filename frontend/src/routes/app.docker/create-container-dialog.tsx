import { useMutation, useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { toast } from "sonner";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Plus, X, Loader2 } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";

interface CreateContainerContentProps {
  clientId: string | null;
  onSuccess?: () => void;
  onClose?: () => void;
}

export function CreateContainerContent({
  clientId,
  onSuccess,
  onClose,
}: CreateContainerContentProps) {
  const [image, setImage] = useState("");
  const [containerName, setContainerName] = useState("");
  const [command, setCommand] = useState("");
  const [envVars, setEnvVars] = useState<{ key: string; value: string }[]>([]);
  const [ports, setPorts] = useState<{ container: string; host: string }[]>([]);

  // Fetch available images for the dropdown
  const imagesQuery = useQuery(
    orpc.docker.images.queryOptions({
      input: { params: { clientId: clientId! } },
      queryOptions: {
        enabled: !!clientId,
      },
    }),
  );

  const createContainerMutation = useMutation(
    orpc.docker.createContainer.mutationOptions({
      onSuccess() {
        toast.success("Container created successfully");
        resetForm();
        onSuccess?.();
        onClose?.();
      },
      onError(error) {
        toast.error(`Failed to create container: ${error.message}`);
      },
    }),
  );

  const resetForm = () => {
    setImage("");
    setContainerName("");
    setCommand("");
    setEnvVars([]);
    setPorts([]);
  };

  const addEnvVar = () => {
    setEnvVars([...envVars, { key: "", value: "" }]);
  };

  const removeEnvVar = (index: number) => {
    setEnvVars(envVars.filter((_, i) => i !== index));
  };

  const updateEnvVar = (
    index: number,
    field: "key" | "value",
    value: string,
  ) => {
    const updated = [...envVars];
    updated[index][field] = value;
    setEnvVars(updated);
  };

  const addPort = () => {
    setPorts([...ports, { container: "", host: "" }]);
  };

  const removePort = (index: number) => {
    setPorts(ports.filter((_, i) => i !== index));
  };

  const updatePort = (
    index: number,
    field: "container" | "host",
    value: string,
  ) => {
    const updated = [...ports];
    updated[index][field] = value;
    setPorts(updated);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!image || !clientId) {
      toast.error("Please select an image");
      return;
    }

    // Build the container config
    const envArray = envVars
      .filter((e) => e.key.trim())
      .map((e) => `${e.key}=${e.value}`);

    const exposedPorts: Record<string, object> = {};
    ports
      .filter((p) => p.container.trim())
      .forEach((p) => {
        exposedPorts[`${p.container}/tcp`] = {};
      });

    const cmdArray = command.trim() ? command.split(" ") : undefined;

    createContainerMutation.mutate({
      params: { clientId },
      query: { name: containerName.trim() || undefined },
      body: {
        Image: image,
        Cmd: cmdArray,
        Env: envArray.length > 0 ? envArray : undefined,
        ExposedPorts: Object.keys(exposedPorts).length > 0 ? exposedPorts : undefined,
      },
    });
  };

  const images = imagesQuery.data ?? [];
  const imageOptions = images
    .filter((img) => img.RepoTags && img.RepoTags.length > 0)
    .flatMap((img) => img.RepoTags || [])
    .filter((tag) => tag !== "<none>:<none>");

  return (
    <form onSubmit={handleSubmit}>
      <ScrollArea className="max-h-[60vh] pr-4">
        <div className="space-y-4">
          {/* Image Selection */}
          <div className="space-y-2">
            <Label htmlFor="image">Image *</Label>
            {imagesQuery.isLoading ? (
              <div className="flex items-center gap-2 text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" />
                Loading images...
              </div>
            ) : (
              <Select value={image} onValueChange={setImage}>
                <SelectTrigger>
                  <SelectValue placeholder="Select an image" />
                </SelectTrigger>
                <SelectContent>
                  {imageOptions.map((tag) => (
                    <SelectItem key={tag} value={tag}>
                      {tag}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>

          {/* Container Name */}
          <div className="space-y-2">
            <Label htmlFor="name">Container Name (optional)</Label>
            <Input
              id="name"
              placeholder="my-container"
              value={containerName}
              onChange={(e) => setContainerName(e.target.value)}
            />
          </div>

          {/* Command */}
          <div className="space-y-2">
            <Label htmlFor="command">Command (optional)</Label>
            <Input
              id="command"
              placeholder="e.g., /bin/bash -c 'echo hello'"
              value={command}
              onChange={(e) => setCommand(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Space-separated command and arguments
            </p>
          </div>

          {/* Environment Variables */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label>Environment Variables</Label>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addEnvVar}
              >
                <Plus className="mr-1 h-3 w-3" />
                Add
              </Button>
            </div>
            {envVars.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                No environment variables
              </p>
            ) : (
              <div className="space-y-2">
                {envVars.map((env, index) => (
                  <div key={index} className="flex items-center gap-2">
                    <Input
                      placeholder="KEY"
                      value={env.key}
                      onChange={(e) =>
                        updateEnvVar(index, "key", e.target.value)
                      }
                      className="flex-1"
                    />
                    <span className="text-muted-foreground">=</span>
                    <Input
                      placeholder="value"
                      value={env.value}
                      onChange={(e) =>
                        updateEnvVar(index, "value", e.target.value)
                      }
                      className="flex-1"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      onClick={() => removeEnvVar(index)}
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Port Mappings */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label>Port Mappings</Label>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addPort}
              >
                <Plus className="mr-1 h-3 w-3" />
                Add
              </Button>
            </div>
            {ports.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                No port mappings
              </p>
            ) : (
              <div className="space-y-2">
                {ports.map((port, index) => (
                  <div key={index} className="flex items-center gap-2">
                    <Input
                      placeholder="Host port"
                      value={port.host}
                      onChange={(e) =>
                        updatePort(index, "host", e.target.value)
                      }
                      className="flex-1"
                    />
                    <span className="text-muted-foreground">:</span>
                    <Input
                      placeholder="Container port"
                      value={port.container}
                      onChange={(e) =>
                        updatePort(index, "container", e.target.value)
                      }
                      className="flex-1"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      onClick={() => removePort(index)}
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
            <p className="text-xs text-muted-foreground">
              Note: Port bindings require host config (coming soon)
            </p>
          </div>
        </div>
      </ScrollArea>

      <div className="mt-4 flex justify-end gap-2">
        <Button
          type="button"
          variant="secondary"
          onClick={onClose}
        >
          Cancel
        </Button>
        <Button
          type="submit"
          disabled={!image || createContainerMutation.isPending}
        >
          {createContainerMutation.isPending ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Creating...
            </>
          ) : (
            "Create Container"
          )}
        </Button>
      </div>
    </form>
  );
}
