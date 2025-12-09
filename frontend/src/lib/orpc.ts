import type { JsonifiedClient } from "@orpc/openapi-client";
import type { ContractRouterClient } from "@orpc/contract";
import { createORPCClient, onError } from "@orpc/client";
import { OpenAPILink } from "@orpc/openapi-client/fetch";
import { createTanstackQueryUtils } from "@orpc/tanstack-query";
import { QueryCache, QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { healthRouter } from "./routers/health";
import { authRouter } from "./routers/auth";

export const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: (error) => {
      toast.error(`Error: ${error.message}`, {
        action: {
          label: "retry",
          onClick: () => {
            queryClient.invalidateQueries();
          },
        },
      });
    },
  }),
});

export const contract = {
  health: healthRouter,
  auth: authRouter,
};

const link = new OpenAPILink(contract, {
  url: "http://localhost:8080/api",
  fetch: (request, init) => {
    return globalThis.fetch(request, {
      ...init,
      credentials: "include", // Include cookies for cross-origin requests
    });
  },
  interceptors: [
    onError((error) => {
      console.error(error);
    }),
  ],
});

export const client: JsonifiedClient<ContractRouterClient<typeof contract>> =
  createORPCClient(link);

export const orpc = createTanstackQueryUtils(client);
