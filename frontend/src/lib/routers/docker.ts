import { base, detailed } from "./general";
import { Z } from "@/types";
import z from "zod";

export const dockerRouter = {
  // Get available docker clients
  clients: base
    .route({
      method: "GET",
      path: "/docker",
    })
    .output(Z.dockerClientInfoSchema.array()),

  // List containers for a specific client
  containers: base
    .route({
      method: "GET",
      path: "/docker/{clientId}/containers",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { clientId: z.string() },
      }),
    )
    .output(Z.dockerContainerSchema.array()),

  // List images for a specific client
  images: base
    .route({
      method: "GET",
      path: "/docker/{clientId}/images",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { clientId: z.string() },
      }),
    )
    .output(Z.dockerImageSchema.array()),

  // Delete an image
  deleteImage: base
    .route({
      method: "DELETE",
      path: "/docker/{clientId}/images/{id}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { clientId: z.string(), id: z.string() },
      }),
    )
    .output(z.void()),

  // Inspect a container
  inspectContainer: base
    .route({
      method: "GET",
      path: "/docker/{clientId}/containers/{id}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { clientId: z.string(), id: z.string() },
      }),
    )
    .output(z.any()),

  // Get container stats
  containerStats: base
    .route({
      method: "GET",
      path: "/docker/{clientId}/containers/{id}/stats",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { clientId: z.string(), id: z.string() },
      }),
    )
    .output(z.any()),

  // Start a container
  startContainer: base
    .route({
      method: "POST",
      path: "/docker/{clientId}/containers/{id}/start",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { clientId: z.string(), id: z.string() },
      }),
    )
    .output(z.void()),

  // Stop a container
  stopContainer: base
    .route({
      method: "POST",
      path: "/docker/{clientId}/containers/{id}/stop",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { clientId: z.string(), id: z.string() },
      }),
    )
    .output(z.void()),

  // Restart a container
  restartContainer: base
    .route({
      method: "POST",
      path: "/docker/{clientId}/containers/{id}/restart",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { clientId: z.string(), id: z.string() },
      }),
    )
    .output(z.void()),

  // Delete a container
  deleteContainer: base
    .route({
      method: "DELETE",
      path: "/docker/{clientId}/containers/{id}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { clientId: z.string(), id: z.string() },
      }),
    )
    .output(z.void()),

  // Get container logs
  containerLogs: base
    .route({
      method: "GET",
      path: "/docker/{clientId}/containers/{id}/logs",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { clientId: z.string(), id: z.string() },
      }),
    )
    .output(z.string()),

  // Create a container
  createContainer: base
    .route({
      method: "POST",
      path: "/docker/{clientId}/containers",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { clientId: z.string() },
        body: {
          Image: z.string(),
          Cmd: z.array(z.string()).optional(),
          Env: z.array(z.string()).optional(),
          ExposedPorts: z.record(z.any()).optional(),
        },
      }),
    )
    .output(Z.containerCreateResponseSchema),
};
