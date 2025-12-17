import { Z } from "@/types";
import { COMMON_ORPC_ERROR_DEFS } from "@orpc/client";
import { oc } from "@orpc/contract";
import type { ErrorMap } from "@orpc/contract";
import z from "zod";

const e = {
  data: Z.httpErrorSchema,
} satisfies ErrorMap[string];

const errors = Object.fromEntries(
  Object.keys(COMMON_ORPC_ERROR_DEFS).map((key) => [key, e]),
) satisfies ErrorMap;

export const base = oc
  .$route({
    method: "GET",
  })
  .errors(errors);

interface DetailedArgs {
  params?: Record<string, z.ZodTypeAny>;
  query?: Record<string, z.ZodTypeAny>;
  body?: Record<string, z.ZodTypeAny>;
  headers?: Record<string, z.ZodTypeAny>;
}

export const detailed = ({
  params = {},
  query = {},
  body = {},
  headers = {},
}: DetailedArgs) =>
  z.object({
    params: z.object(params).optional(),
    query: z.object(query).optional(),
    body: z.object(body).optional(),
    headers: z.object(headers).optional(),
  });
