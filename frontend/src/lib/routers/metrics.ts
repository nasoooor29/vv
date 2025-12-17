import { Z } from "@/types";
import { base } from "./general";
import { z } from "zod";

export const metricsRouter = {
  getMetrics: base
    .route({
      method: "GET",
      path: "/metrics",
    })
    .input(
      z.object({
        days: z.number().optional().default(7),
      })
    )
    .output(Z.metricsResponseSchema),

  getServiceMetrics: base
    .route({
      method: "GET",
      path: "/metrics/service/:service",
    })
    .input(
      z.object({
        service: z.string(),
        days: z.number().optional().default(7),
      })
    )
    .output(Z.serviceMetricsResponseSchema),

  getHealthMetrics: base
    .route({
      method: "GET",
      path: "/metrics/health",
    })
    .input(
      z.object({})
    )
    .output(Z.healthMetricsResponseSchema),
};
