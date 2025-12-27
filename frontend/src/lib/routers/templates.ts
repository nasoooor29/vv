import { base, detailed } from "./general";
import { Z } from "@/types";
import z from "zod";

export const templatesRouter = {
  // List all templates
  list: base
    .route({
      method: "GET",
      path: "/templates",
    })
    .output(Z.templateListItemSchema.array()),

  // Get template categories
  categories: base
    .route({
      method: "GET",
      path: "/templates/categories",
    })
    .output(z.array(z.string())),

  // Get a single template by ID
  get: base
    .route({
      method: "GET",
      path: "/templates/{id}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { id: z.string() },
      }),
    )
    .output(Z.portainerTemplateSchema),

  // Refresh templates cache
  refresh: base
    .route({
      method: "POST",
      path: "/templates/refresh",
    })
    .output(
      z.object({
        message: z.string(),
        count: z.number(),
      }),
    ),

  // Deploy a template
  deploy: base
    .route({
      method: "POST",
      path: "/docker/{clientId}/templates/{id}/deploy",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { clientId: z.string(), id: z.string() },
        body: {
          name: z.string().optional(),
          env: z.record(z.string(), z.string()).optional(),
          ports: z.array(z.string()).optional(),
          volumes: Z.templateVolumeSchema.array().optional(),
          network: z.string().optional(),
          restart_policy: z.string().optional(),
        },
      }),
    )
    .output(Z.deployResponseSchema),
};
