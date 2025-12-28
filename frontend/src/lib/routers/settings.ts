import { base, detailed } from "./general";
import z from "zod";

// Notification setting schema based on the generated type
const notificationSettingSchema = z.object({
  id: z.number(),
  provider: z.string(),
  enabled: z.boolean().optional(),
  webhook_url: z.string().optional(),
  notify_on_error: z.boolean().optional(),
  notify_on_warn: z.boolean().optional(),
  notify_on_info: z.boolean().optional(),
  config: z.string().optional(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const settingsRouter = {
  // Get all notification settings
  getNotificationSettings: base
    .route({
      method: "GET",
      path: "/settings/notifications",
    })
    .output(notificationSettingSchema.array()),

  // Get notification setting by provider
  getNotificationSettingByProvider: base
    .route({
      method: "GET",
      path: "/settings/notifications/{provider}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { provider: z.string() },
      }),
    )
    .output(notificationSettingSchema),

  // Create or update notification setting
  upsertNotificationSetting: base
    .route({
      method: "POST",
      path: "/settings/notifications",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        body: {
          provider: z.string(),
          enabled: z.boolean(),
          webhook_url: z.string(),
          notify_on_error: z.boolean(),
          notify_on_warn: z.boolean(),
          notify_on_info: z.boolean(),
          config: z.string().optional(),
        },
      }),
    )
    .output(notificationSettingSchema),

  // Delete notification setting
  deleteNotificationSetting: base
    .route({
      method: "DELETE",
      path: "/settings/notifications/{provider}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { provider: z.string() },
      }),
    )
    .output(z.void()),

  // Test notification
  testNotification: base
    .route({
      method: "POST",
      path: "/settings/notifications/{provider}/test",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { provider: z.string() },
      }),
    )
    .output(z.object({ message: z.string() })),
};
