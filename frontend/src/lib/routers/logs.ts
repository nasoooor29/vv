import { Z } from "@/types";
import { base } from "./general";
import { z } from "zod";

export const logsRouter = {
  getLogs: base
    .route({
      method: "GET",
      path: "/logs",
      inputStructure: "query",
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
    .output(
      z.object({
        logs: z.array(
          z.object({
            id: z.number(),
            user_id: z.number(),
            action: z.string(),
            details: z.string().nullable(),
            service_group: z.string(),
            level: z.string(),
            created_at: z.string().datetime(),
          })
        ),
        total: z.number(),
        page: z.number(),
        page_size: z.number(),
        total_pages: z.number(),
      })
    ),

  getLogStats: base
    .route({
      method: "GET",
      path: "/logs/stats",
      inputStructure: "query",
    })
    .input(
      z.object({
        days: z.number().optional().default(7),
      })
    )
    .output(
      z.object({
        total: z.number(),
        days: z.number(),
        service_groups: z.array(z.string()),
        levels: z.array(z.string()),
        since: z.string().datetime(),
      })
    ),

  clearOldLogs: base
    .route({
      method: "DELETE",
      path: "/logs/cleanup",
      inputStructure: "query",
    })
    .input(
      z.object({
        days: z.number().optional().default(30),
      })
    )
    .output(
      z.object({
        retention_days: z.number(),
        before: z.string().datetime(),
        message: z.string(),
      })
    ),
};
