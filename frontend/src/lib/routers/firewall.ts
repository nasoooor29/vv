import { base, detailed } from "./general";
import z from "zod";

// Firewall rule schema
const firewallRuleSchema = z.object({
  handle: z.number(),
  chain: z.string(),
  protocol: z.string(),
  port: z.number(),
  source_ip: z.string(),
  action: z.string(),
  comment: z.string(),
});

// Firewall status schema
const firewallStatusSchema = z.object({
  enabled: z.boolean(),
  rule_count: z.number(),
  table_name: z.string(),
});

export const firewallRouter = {
  // Get firewall status
  status: base
    .route({
      method: "GET",
      path: "/firewall/status",
    })
    .output(firewallStatusSchema),

  // List all firewall rules
  rules: base
    .route({
      method: "GET",
      path: "/firewall/rules",
    })
    .output(firewallRuleSchema.array()),

  // Add a new firewall rule
  addRule: base
    .route({
      method: "POST",
      path: "/firewall/rules",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        body: {
          chain: z.enum(["input", "forward", "output"]),
          protocol: z.enum(["tcp", "udp", ""]).optional(),
          port: z.number().optional(),
          source_ip: z.string().optional(),
          action: z.enum(["accept", "drop"]),
          comment: z.string().optional(),
        },
      }),
    )
    .output(firewallRuleSchema),

  // Delete a firewall rule
  deleteRule: base
    .route({
      method: "DELETE",
      path: "/firewall/rules/{handle}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { handle: z.string() },
      }),
    )
    .output(z.void()),

  // Reorder firewall rules
  reorderRules: base
    .route({
      method: "POST",
      path: "/firewall/rules/reorder",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        body: {
          chain: z.enum(["input", "forward", "output"]),
          handles: z.array(z.number()),
        },
      }),
    )
    .output(firewallRuleSchema.array()),
};
