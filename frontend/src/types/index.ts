import * as zodTypes from "./zod-types.gen";
export * as T from "./types.gen";

import { z } from "zod";

export const Z = {
  ...zodTypes,
  withHttpError: (ok: z.ZodObject) =>
    z.discriminatedUnion("message", [ok, zodTypes.httpErrorSchema]),
};
