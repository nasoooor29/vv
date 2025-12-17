import { Z } from "@/types";
import { base } from "./general";
import { z } from "zod";

export const logsRouter = {
  getLogs: base
    .route({
      method: "GET",
      path: "/logs",
    })
    .input(
      z.object({
        service_group: z.string().optional(),
        level: z.string().optional(),
        page: z.number().optional().default(1),
        page_size: z.number().optional().default(20),
        days: z.number().optional().default(7),
      })
    )
    .output(Z.getLogsResponseSchema),

  getLogStats: base
    .route({
      method: "GET",
      path: "/logs/stats",
    })
    .input(
      z.object({
        days: z.number().optional().default(7),
      })
    )
    .output(Z.logStatsResponseSchema),

  clearOldLogs: base
    .route({
      method: "DELETE",
      path: "/logs/cleanup",
    })
    .input(
      z.object({
        days: z.number().optional().default(30),
      })
    )
    .output(Z.clearOldLogsResponseSchema),
};
