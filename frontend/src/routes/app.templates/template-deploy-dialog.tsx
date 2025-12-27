import { useQuery, useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2, Rocket, Package } from "lucide-react";
import { useState, useEffect } from "react";
import { toast } from "sonner";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { Badge } from "@/components/ui/badge";
import { useDialog } from "@/hooks";

interface UseTemplateDeployDialogProps {
  templateId: number | null;
  clientId: string | null;
  onSuccess?: () => void;
}

// Simple markdown stripper for display purposes
function stripMarkdown(text: string): string {
  return text
    .replace(/\*\*(.+?)\*\*/g, "$1") // bold
    .replace(/\*(.+?)\*/g, "$1") // italic
    .replace(/__(.+?)__/g, "$1") // bold
    .replace(/_(.+?)_/g, "$1") // italic
    .replace(/\[([^\]]+)\]\([^)]+\)/g, "$1") // links
    .replace(/`(.+?)`/g, "$1") // inline code
    .replace(/#{1,6}\s?/g, "") // headers
    .trim();
}

export function useTemplateDeployDialog({
  templateId,
  clientId,
  onSuccess,
}: UseTemplateDeployDialogProps) {
  const [containerName, setContainerName] = useState("");
  const [envOverrides, setEnvOverrides] = useState<Record<string, string>>({});

  const dialog = useDialog({
    className: "max-w-2xl max-h-[80vh] overflow-y-auto",
  });

  // Fetch template details when dialog is open
  const templateQuery = useQuery({
    ...orpc.templates.get.queryOptions({
      input: { params: { id: templateId?.toString() ?? "0" } },
    }),
    enabled: dialog.isOpen && templateId !== null,
  });

  const template = templateQuery.data;

  // Initialize env overrides from template defaults
  useEffect(() => {
    if (template?.env) {
      const defaults: Record<string, string> = {};
      template.env.forEach((env) => {
        if (env.default) {
          defaults[env.name] = env.default;
        }
      });
      setEnvOverrides(defaults);
    }
  }, [template]);

  // Reset form when dialog closes
  useEffect(() => {
    if (!dialog.isOpen) {
      setContainerName("");
      setEnvOverrides({});
    }
  }, [dialog.isOpen]);

  // Deploy mutation
  const deployMutation = useMutation(
    orpc.templates.deploy.mutationOptions({
      onSuccess(data) {
        toast.success(data.message);
        dialog.close();
        onSuccess?.();
      },
      onError() {
        toast.error("Failed to deploy template");
      },
    }),
  );

  const handleDeploy = () => {
    if (!clientId || templateId === null) return;
    deployMutation.mutate({
      params: { clientId, id: templateId.toString() },
      body: {
        name: containerName || undefined,
        env: Object.keys(envOverrides).length > 0 ? envOverrides : undefined,
      },
    });
  };

  const handleEnvChange = (name: string, value: string) => {
    setEnvOverrides((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const content = (close: () => void) => {
    if (templateQuery.isLoading) {
      return (
        <div className="flex items-center justify-center py-8">
          <Loader2 className="h-8 w-8 animate-spin" />
        </div>
      );
    }

    if (!template) {
      return (
        <div className="py-8 text-center text-muted-foreground">
          Template not found
        </div>
      );
    }

    return (
      <div className="space-y-4">
        {/* Header with logo */}
        <div className="flex items-center gap-4">
          {template.logo ? (
            <img
              src={template.logo}
              alt={template.title}
              className="h-12 w-12 rounded object-contain"
            />
          ) : (
            <div className="flex h-12 w-12 items-center justify-center rounded bg-muted">
              <Package className="h-6 w-6 text-muted-foreground" />
            </div>
          )}
          <div>
            <h3 className="font-semibold">{template.title}</h3>
            <p className="text-sm text-muted-foreground">{template.description ? stripMarkdown(template.description) : ""}</p>
          </div>
        </div>

        {/* Image info */}
        <div className="rounded-md bg-muted p-3">
          <p className="text-xs text-muted-foreground">Image</p>
          <p className="font-mono text-sm">{template.image}</p>
        </div>

        {/* Container name */}
        <div className="space-y-2">
          <Label htmlFor="container-name">Container Name (optional)</Label>
          <Input
            id="container-name"
            placeholder={template.name || template.title.toLowerCase().replace(/\s+/g, "-")}
            value={containerName}
            onChange={(e) => setContainerName(e.target.value)}
          />
          <p className="text-xs text-muted-foreground">
            Leave empty to use the default name
          </p>
        </div>

        {/* Environment Variables */}
        {template.env && template.env.length > 0 && (
          <Accordion type="single" collapsible defaultValue="env">
            <AccordionItem value="env">
              <AccordionTrigger>
                Environment Variables ({template.env.length})
              </AccordionTrigger>
              <AccordionContent>
                <div className="space-y-4">
                  {template.env.map((env) => (
                    <div key={env.name} className="space-y-2">
                      <div className="flex items-center gap-2">
                        <Label htmlFor={`env-${env.name}`} className="font-mono text-xs">
                          {env.name}
                        </Label>
                        {env.preset && (
                          <Badge variant="secondary" className="text-xs">
                            preset
                          </Badge>
                        )}
                      </div>
                      {env.label && (
                        <p className="text-xs text-muted-foreground">{stripMarkdown(env.label)}</p>
                      )}
                      {env.description && (
                        <p className="text-xs text-muted-foreground">{stripMarkdown(env.description)}</p>
                      )}
                      {env.select && env.select.length > 0 ? (
                        <select
                          id={`env-${env.name}`}
                          className="w-full rounded-md border bg-background px-3 py-2 text-sm"
                          value={envOverrides[env.name] || env.default || ""}
                          onChange={(e) => handleEnvChange(env.name, e.target.value)}
                        >
                          {env.select.map((option) => (
                            <option key={option.value} value={option.value}>
                              {option.text}
                            </option>
                          ))}
                        </select>
                      ) : (
                        <Input
                          id={`env-${env.name}`}
                          placeholder={env.default || ""}
                          value={envOverrides[env.name] || ""}
                          onChange={(e) => handleEnvChange(env.name, e.target.value)}
                        />
                      )}
                    </div>
                  ))}
                </div>
              </AccordionContent>
            </AccordionItem>
          </Accordion>
        )}

        {/* Ports info */}
        {template.ports && template.ports.length > 0 && (
          <div className="space-y-2">
            <Label>Exposed Ports</Label>
            <div className="flex flex-wrap gap-2">
              {template.ports.map((port, idx) => (
                <Badge key={idx} variant="outline" className="font-mono">
                  {port}
                </Badge>
              ))}
            </div>
          </div>
        )}

        {/* Volumes info */}
        {template.volumes && template.volumes.length > 0 && (
          <div className="space-y-2">
            <Label>Volumes</Label>
            <div className="space-y-1">
              {template.volumes.map((vol, idx) => (
                <div key={idx} className="rounded bg-muted px-2 py-1 font-mono text-xs">
                  {vol.bind || "(auto)"} : {vol.container}
                  {vol.readonly && " (ro)"}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Note */}
        {template.note && (
          <div className="rounded-md border border-yellow-500/20 bg-yellow-500/10 p-3">
            <p className="text-xs text-yellow-600 dark:text-yellow-400">
              {stripMarkdown(template.note)}
            </p>
          </div>
        )}

        {/* Actions */}
        <div className="flex justify-end gap-2 pt-4">
          <Button variant="outline" onClick={close}>
            Cancel
          </Button>
          <Button onClick={handleDeploy} disabled={deployMutation.isPending}>
            {deployMutation.isPending ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Rocket className="mr-2 h-4 w-4" />
            )}
            Deploy Container
          </Button>
        </div>
      </div>
    );
  };

  return {
    dialog,
    content,
  };
}
