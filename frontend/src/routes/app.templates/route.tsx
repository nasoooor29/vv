import { useQuery, useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Loader2, Package, Search, Server, Rocket, RefreshCw } from "lucide-react";
import { useState, useMemo } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toast } from "sonner";
import { usePermission } from "@/components/protected-content";
import { RBAC_DOCKER_READ, RBAC_DOCKER_WRITE } from "@/types/types.gen";
import { useTemplateDeployDialog } from "./template-deploy-dialog";

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

export default function TemplatesPage() {
  const { hasPermission } = usePermission();
  const [selectedClient, setSelectedClient] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const [selectedTemplateId, setSelectedTemplateId] = useState<number | null>(null);

  // Fetch available docker clients
  const clientsQuery = useQuery(orpc.docker.clients.queryOptions({}));

  // Fetch templates
  const templatesQuery = useQuery(orpc.templates.list.queryOptions({}));

  // Fetch categories
  const categoriesQuery = useQuery(orpc.templates.categories.queryOptions({}));

  // Refresh cache mutation
  const refreshMutation = useMutation(
    orpc.templates.refresh.mutationOptions({
      onSuccess(data) {
        toast.success(`${data.message} (${data.count} templates)`);
      },
      onError() {
        toast.error("Failed to refresh templates cache");
      },
    }),
  );

  const clients = clientsQuery.data ?? [];
  const activeClientId = selectedClient ?? clients[0]?.id?.toString() ?? null;
  const templates = templatesQuery.data ?? [];
  const categories = categoriesQuery.data ?? [];

  // Deploy dialog
  const deployDialog = useTemplateDeployDialog({
    templateId: selectedTemplateId,
    clientId: activeClientId,
    onSuccess: () => setSelectedTemplateId(null),
  });

  // Filter templates based on search and category
  const filteredTemplates = useMemo(() => {
    return templates.filter((template) => {
      const matchesSearch =
        searchQuery === "" ||
        template.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
        template.description?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        template.image.toLowerCase().includes(searchQuery.toLowerCase());

      const matchesCategory =
        selectedCategory === "all" ||
        template.categories?.includes(selectedCategory);

      return matchesSearch && matchesCategory;
    });
  }, [templates, searchQuery, selectedCategory]);

  const handleDeploy = (templateId: number) => {
    setSelectedTemplateId(templateId);
    deployDialog.dialog.open();
  };

  // Check permission
  if (!hasPermission(RBAC_DOCKER_READ)) {
    return (
      <div className="flex items-center justify-center p-8">
        <p className="text-muted-foreground">
          You don't have permission to access Docker templates.
        </p>
      </div>
    );
  }

  if (templatesQuery.isLoading || clientsQuery.isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  if (templatesQuery.error) {
    return (
      <div className="p-4 text-destructive">
        Error loading templates. Please try again later.
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
          running and properly configured before deploying templates.
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Package className="h-6 w-6" />
          <h1 className="text-3xl font-bold">Docker Templates</h1>
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
          {hasPermission(RBAC_DOCKER_WRITE) && (
            <Button
              variant="outline"
              onClick={() => refreshMutation.mutate({})}
              disabled={refreshMutation.isPending}
            >
              {refreshMutation.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <RefreshCw className="mr-2 h-4 w-4" />
              )}
              Refresh
            </Button>
          )}
        </div>
      </div>

      {/* Search and Filter */}
      <div className="flex gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search templates..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <Select value={selectedCategory} onValueChange={setSelectedCategory}>
          <SelectTrigger className="w-[200px]">
            <SelectValue placeholder="Category" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Categories</SelectItem>
            {categories.map((category) => (
              <SelectItem key={category} value={category}>
                {category}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Results count */}
      <p className="text-sm text-muted-foreground">
        Showing {filteredTemplates.length} of {templates.length} templates
      </p>

      {/* Templates Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {filteredTemplates.map((template) => (
          <Card key={template.id} className="flex flex-col">
            <CardHeader className="flex-row items-start gap-4 space-y-0">
              {template.logo ? (
                <img
                  src={template.logo}
                  alt={template.title}
                  className="h-12 w-12 rounded object-contain"
                  onError={(e) => {
                    (e.target as HTMLImageElement).style.display = "none";
                  }}
                />
              ) : (
                <div className="flex h-12 w-12 items-center justify-center rounded bg-muted">
                  <Package className="h-6 w-6 text-muted-foreground" />
                </div>
              )}
              <div className="flex-1 space-y-1">
                <CardTitle className="text-base">{template.title}</CardTitle>
                <p className="text-xs text-muted-foreground line-clamp-2">
                  {template.description ? stripMarkdown(template.description) : "No description available"}
                </p>
              </div>
            </CardHeader>
            <CardContent className="flex flex-1 flex-col justify-between gap-4">
              <div className="space-y-2">
                <p className="text-xs font-mono text-muted-foreground truncate">
                  {template.image}
                </p>
                {template.categories && template.categories.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {template.categories.slice(0, 3).map((cat) => (
                      <Badge key={cat} variant="secondary" className="text-xs">
                        {cat}
                      </Badge>
                    ))}
                    {template.categories.length > 3 && (
                      <Badge variant="outline" className="text-xs">
                        +{template.categories.length - 3}
                      </Badge>
                    )}
                  </div>
                )}
              </div>
              {hasPermission(RBAC_DOCKER_WRITE) && activeClientId && (
                <Button
                  size="sm"
                  onClick={() => handleDeploy(template.id)}
                  className="w-full"
                >
                  <Rocket className="mr-2 h-4 w-4" />
                  Deploy
                </Button>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {filteredTemplates.length === 0 && (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <Package className="h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-lg font-medium">No templates found</p>
          <p className="text-sm text-muted-foreground">
            Try adjusting your search or filter criteria
          </p>
        </div>
      )}

      {/* Deploy Dialog */}
      <deployDialog.dialog.Component title="Deploy Template">
        {deployDialog.content}
      </deployDialog.dialog.Component>
    </div>
  );
}
