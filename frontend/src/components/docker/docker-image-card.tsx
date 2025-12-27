import { formatBytes, formatTimeDifference } from "@/lib/utils";
import { Container, Download, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { z } from "zod";
import type { Z } from "@/types";

type DockerImage = z.infer<typeof Z.dockerImageSchema>;

interface DockerImageCardProps {
  image: DockerImage;
}

function DockerImageCard({ image }: DockerImageCardProps) {
  const imageTag =
    image.RepoTags && image.RepoTags.length > 0
      ? image.RepoTags[0]
      : image.Id.slice(7, 19);

  const imageTitle =
    image.Labels?.["org.opencontainers.image.title"] || imageTag;

  return (
    <div className="flex items-center justify-between rounded-lg border p-3">
      <div className="flex items-center gap-3">
        <Container className="h-5 w-5 text-muted-foreground" />
        <div>
          <div className="text-left font-medium">{imageTitle}</div>
          <div className="text-left text-sm text-muted-foreground">
            {formatBytes(image.Size)} - Created{" "}
            {formatTimeDifference(new Date(image.Created * 1000))}
          </div>
        </div>
      </div>
      <div className="flex gap-2">
        <Button size="sm" variant="outline">
          <Download className="h-3 w-3" />
        </Button>
        <Button
          size="sm"
          variant="outline"
          className="text-destructive hover:text-destructive"
        >
          <Trash2 className="h-3 w-3" />
        </Button>
      </div>
    </div>
  );
}

export default DockerImageCard;
