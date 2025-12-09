import { Z } from "@/types";
import { base } from "./general";

export const healthRouter = {
  check: base
    .route({
      method: "GET",
      path: "/health",
    })
    .output(Z.healthSchema),
};
