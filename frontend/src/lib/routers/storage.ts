import { z } from "zod";
import { base } from "./general";

export const storageRouter = {
  devices: base
    .route({
      method: "GET",
      path: "/storage/devices",
    })
    .output(z.object({
      devices: z.array(z.object({
        name: z.string(),
        size: z.string(),
        size_bytes: z.number(),
        type: z.string(),
        mount_point: z.string().optional(),
        usage_percent: z.number().optional(),
      })),
    })),

  mountPoints: base
    .route({
      method: "GET",
      path: "/storage/mount-points",
    })
    .output(z.object({
      mount_points: z.array(z.object({
        path: z.string(),
        device: z.string(),
        fs_type: z.string(),
        total: z.number(),
        used: z.number(),
        available: z.number(),
        use_percent: z.number(),
      })),
    })),
};
