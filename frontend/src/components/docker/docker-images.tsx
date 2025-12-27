import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Loader2 } from "lucide-react";
import DockerImageCard from "./docker-image-card";
import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";

interface DockerImagesProps {
  clientId: string;
}

function DockerImages({ clientId }: DockerImagesProps) {
  const { data, error, isFetching } = useQuery(
    orpc.docker.images.queryOptions({
      input: { params: { clientId } },
    }),
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>Container Images</CardTitle>
        <CardDescription>Manage Docker images on this host</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-3 text-center">
          {isFetching && (
            <div className="flex justify-center">
              <Loader2 className="h-6 w-6 animate-spin" />
            </div>
          )}
          {data &&
            data.map((image, index) => (
              <DockerImageCard image={image} key={index} />
            ))}
          {data && data.length === 0 && (
            <div className="text-muted-foreground">No images</div>
          )}
          {error && (
            <div className="text-destructive">Could not fetch images</div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

export default DockerImages;
