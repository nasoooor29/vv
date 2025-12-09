import { Z } from "@/types";
import { COMMON_ORPC_ERROR_DEFS } from "@orpc/client";
import { oc } from "@orpc/contract";
import type { ErrorMap } from "@orpc/contract";

const e = {
  data: Z.httpErrorSchema,
};

const errors = Object.fromEntries(
  Object.keys(COMMON_ORPC_ERROR_DEFS).map((key) => [key, e]),
) satisfies ErrorMap;

export const base = oc
  .$route({
    method: "GET",
  })
  .errors(errors);
