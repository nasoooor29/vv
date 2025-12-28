import { base, detailed } from "./general";
import { Z } from "@/types";
import z from "zod";

// Message response schema
const messageSchema = z.object({
  message: z.string(),
});

// Firewall backup content schema (for getting backup content)
const firewallBackupContentSchema = z.object({
  backup: Z.firewallBackupSchema,
  rules: Z.firewallRuleSchema.array(),
});

export const backupRouter = {
  // === Backup Jobs ===

  // List all backup jobs
  listJobs: base
    .route({
      method: "GET",
      path: "/backup/jobs",
    })
    .output(Z.backupJobSchema.array()),

  // Get a specific backup job
  getJob: base
    .route({
      method: "GET",
      path: "/backup/jobs/{id}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { id: z.string() },
      })
    )
    .output(Z.backupJobSchema),

  // Delete a backup job
  deleteJob: base
    .route({
      method: "DELETE",
      path: "/backup/jobs/{id}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { id: z.string() },
      })
    )
    .output(messageSchema),

  // === Backup Stats ===

  // Get backup statistics
  getStats: base
    .route({
      method: "GET",
      path: "/backup/stats",
    })
    .output(Z.backupStatsSchema),

  // === Backup Schedules ===

  // List all backup schedules
  listSchedules: base
    .route({
      method: "GET",
      path: "/backup/schedules",
    })
    .output(Z.backupScheduleSchema.array()),

  // Get a specific backup schedule
  getSchedule: base
    .route({
      method: "GET",
      path: "/backup/schedules/{id}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { id: z.string() },
      })
    )
    .output(Z.backupScheduleSchema),

  // Create a new backup schedule
  createSchedule: base
    .route({
      method: "POST",
      path: "/backup/schedules",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        body: {
          name: z.string(),
          type: z.enum(["vm_snapshot", "vm_export", "container_export", "container_commit", "firewall"]),
          target_type: z.enum(["vm", "container", "firewall"]),
          target_id: z.string(),
          target_name: z.string(),
          client_id: z.string().optional(),
          schedule: z.enum(["hourly", "daily", "weekly"]),
          schedule_time: z.string(), // HH:MM format
          retention_count: z.number().min(1).max(100),
        },
      })
    )
    .output(Z.backupScheduleSchema),

  // Update an existing backup schedule
  updateSchedule: base
    .route({
      method: "PUT",
      path: "/backup/schedules/{id}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { id: z.string() },
        body: {
          name: z.string().optional(),
          schedule: z.enum(["hourly", "daily", "weekly"]).optional(),
          schedule_time: z.string().optional(),
          retention_count: z.number().min(1).max(100).optional(),
          enabled: z.boolean().optional(),
        },
      })
    )
    .output(Z.backupScheduleSchema),

  // Delete a backup schedule
  deleteSchedule: base
    .route({
      method: "DELETE",
      path: "/backup/schedules/{id}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { id: z.string() },
      })
    )
    .output(messageSchema),

  // Toggle backup schedule enabled/disabled
  toggleSchedule: base
    .route({
      method: "POST",
      path: "/backup/schedules/{id}/toggle",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { id: z.string() },
      })
    )
    .output(Z.backupScheduleSchema),

  // === Firewall Backups ===

  // List all firewall backups
  listFirewallBackups: base
    .route({
      method: "GET",
      path: "/backup/firewall",
    })
    .output(Z.firewallBackupSchema.array()),

  // Get a specific firewall backup (with rules content)
  getFirewallBackup: base
    .route({
      method: "GET",
      path: "/backup/firewall/{filename}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { filename: z.string() },
      })
    )
    .output(firewallBackupContentSchema),

  // Delete a firewall backup
  deleteFirewallBackup: base
    .route({
      method: "DELETE",
      path: "/backup/firewall/{filename}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { filename: z.string() },
      })
    )
    .output(messageSchema),

  // === Container Backups ===

  // List all container backups
  listContainerBackups: base
    .route({
      method: "GET",
      path: "/backup/containers",
    })
    .output(Z.containerBackupSchema.array()),

  // Delete a container backup
  deleteContainerBackup: base
    .route({
      method: "DELETE",
      path: "/backup/containers/{filename}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { filename: z.string() },
      })
    )
    .output(messageSchema),
};
