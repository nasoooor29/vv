import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Layers, Trash2, HardDrive, Clock, Tag } from "lucide-react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";
import { formatBytes, formatTimeDifference } from "@/lib/utils";
import { useConfirmation } from "@/hooks";
import { toast } from "sonner";

interface ImagesCardProps {
  clientId: string;
}

export function ImagesCard({ clientId }: ImagesCardProps) {
  const { confirm, ConfirmationDialog } = useConfirmation();

  const {
    data: images,
    isLoading,
    isError,
  } = useQuery(
    orpc.docker.images.queryOptions({
      input: { params: { clientId } },
    }),
  );

  const deleteImageMutation = useMutation(
    orpc.docker.deleteImage.mutationOptions({
      onSuccess() {
        toast.success("Image deleted successfully");
      },
      onError() {
        toast.error("Failed to delete image. It may be in use by a container.");
      },
    }),
  );

  const handleDelete = (imageId: string, imageName: string) => {
    confirm({
      title: "Delete Image",
      description: `Are you sure you want to delete "${imageName}"? This action cannot be undone.`,
      isDestructive: true,
      onConfirm: async () => {
        await deleteImageMutation.mutateAsync({
          params: { clientId, id: imageId },
        });
      },
    });
  };

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Layers className="h-5 w-5" />
              Container Images
            </CardTitle>
            <CardDescription>Manage Docker images on this host</CardDescription>
          </div>
          <div className="text-sm text-muted-foreground">
            {images?.length || 0} images
          </div>
        </CardHeader>
        <CardContent>
          {isLoading && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-32 w-full" />
              ))}
            </div>
          )}

          {isError && (
            <div className="py-4 text-center text-destructive">
              Failed to load images
            </div>
          )}

          {images && images.length === 0 && (
            <div className="py-12 text-center text-muted-foreground">
              <Layers className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No images found</p>
            </div>
          )}

          {images && images.length > 0 && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {images.map((image) => {
                const repoTag =
                  image.RepoTags && image.RepoTags.length > 0
                    ? image.RepoTags[0]
                    : "<none>:<none>";
                const [repo, tag] = repoTag.split(":");
                const shortId = image.Id.replace("sha256:", "").slice(0, 12);

                return (
                  <div
                    key={image.Id}
                    className="border border-border rounded-lg p-4 space-y-3 hover:bg-accent/50 transition-colors"
                  >
                    {/* Header */}
                    <div className="flex items-start justify-between">
                      <div className="flex-1 min-w-0">
                        <h3 className="font-semibold truncate" title={repo}>
                          {repo === "<none>" ? shortId : repo}
                        </h3>
                        <div className="flex items-center gap-1 mt-1">
                          <Tag className="h-3 w-3 text-muted-foreground" />
                          <Badge variant="outline" className="text-xs">
                            {tag}
                          </Badge>
                        </div>
                      </div>
                    </div>

                    {/* Info */}
                    <div className="grid grid-cols-2 gap-2 text-sm">
                      <div className="space-y-1">
                        <p className="text-xs text-muted-foreground flex items-center gap-1">
                          <HardDrive className="h-3 w-3" />
                          Size
                        </p>
                        <p className="font-medium">{formatBytes(image.Size)}</p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-xs text-muted-foreground flex items-center gap-1">
                          <Clock className="h-3 w-3" />
                          Created
                        </p>
                        <p className="font-medium">
                          {formatTimeDifference(new Date(image.Created * 1000))}
                        </p>
                      </div>
                    </div>

                    {/* ID */}
                    <div className="text-xs text-muted-foreground font-mono">
                      {shortId}
                    </div>

                    {/* Actions */}
                    <div className="flex gap-2 pt-2">
                      <Button
                        variant="outline"
                        size="sm"
                        className="gap-2 text-destructive hover:text-destructive flex-1"
                        disabled={deleteImageMutation.isPending}
                        onClick={() => handleDelete(image.Id, repoTag)}
                      >
                        <Trash2 className="h-4 w-4" />
                        Delete
                      </Button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>
      <ConfirmationDialog />
    </>
  );
}
