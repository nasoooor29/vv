import { Z } from "@/types";
import { base, detailed } from "./general";
import { z } from "zod";
// Example usage:
// const schema = detailed({
//   params: { id: z.string() },
//   query: { search: z.string().optional() },
//   body: { name: z.string() },
//   headers: { authorization: z.string() },
// });
export const metricsRouter = {
  getMetrics: base
    .route({
      method: "GET",
      path: "/metrics",
    })
    .input(
      z.object({
        days: z.number().optional().default(7),
      }),
    )
    .output(Z.metricsResponseSchema),

  getServiceMetrics: base
    .route({
      method: "GET",
      path: "/metrics/service/{service}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { service: z.string() },
        query: { days: z.number().optional().default(7) },
      }),
    )
    .output(Z.serviceMetricsResponseSchema),

  getHealthMetrics: base
    .route({
      method: "GET",
      path: "/metrics/health",
    })
    .input(z.object({}))
    .output(Z.healthMetricsResponseSchema),
};
