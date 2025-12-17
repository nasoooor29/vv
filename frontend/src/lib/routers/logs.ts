import { Z } from "@/types";
import { base } from "./general";
import { z } from "zod";

export const logsRouter = {
  getLogs: base
    .route({
      method: "GET",
      path: "/logs",
    })
    .input(Z.getLogsRequestSchema)
    .output(Z.getLogsResponseSchema),

  getLogStats: base
    .route({
      method: "GET",
      path: "/logs/stats",
    })
    .input(
      z.object({
        // no need to create type in golang
        days: z.number().optional().default(7),
      }),
    )
    .output(Z.logStatsResponseSchema),

  clearOldLogs: base
    .route({
      method: "DELETE",
      path: "/logs/cleanup",
    })
    .input(
      z.object({
        // no need to create type in golang
        days: z.number().optional().default(30),
      }),
    )
    .output(Z.clearOldLogsResponseSchema),
};
