import * as zodTypes from "./zod-types.gen";
export * as T from "./types.gen";

export const Z = {
  ...zodTypes,
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
