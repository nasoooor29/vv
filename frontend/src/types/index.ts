import * as zodTypes from "./zod-types.gen";
import { z } from "zod";
export * as T from "./types.gen";

// Override schemas that need Zod v4 compatible z.record() with 2 arguments
const dockerLabelsSchema = z.record(z.string(), z.string());

const dockerNetworkSettingsSchema = z.object({
  Networks: z.record(z.string(), zodTypes.dockerNetworkSchema),
});

export const Z = {
  ...zodTypes,
  // Zod v4 compatible overrides
  dockerLabelsSchema,
  dockerNetworkSettingsSchema,
  // withHttpError: (ok: z.ZodObject) =>
  //   z.discriminatedUnion("message", [ok, zodTypes.httpErrorSchema]),
  isHttpError: (obj: any) => {
    return (
      obj &&
      typeof obj === "object" &&
      "message" in obj &&
      typeof obj.message === "string"
    );
  },
};
