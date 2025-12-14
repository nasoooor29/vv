import type { JsonifiedClient } from "@orpc/openapi-client";
import type { ContractRouterClient } from "@orpc/contract";
import { createORPCClient, ORPCError } from "@orpc/client";
import { OpenAPILink } from "@orpc/openapi-client/fetch";
import { createTanstackQueryUtils } from "@orpc/tanstack-query";
import { QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { healthRouter } from "./routers/health";
import { authRouter } from "./routers/auth";
import { usersRouter } from "./routers/users";
import { Z } from "@/types";

export const queryClient = new QueryClient({
  defaultOptions: {
    mutations: {
      onError(error, _variables, _onMutateResult, _context) {
        if (error instanceof ORPCError) {
          const back = Z.httpErrorSchema.safeParse(error?.data?.body);
          if (back.success) {
            toast.error(`${back.data.message}`);
          } else {
            toast.error(`Unexpected error format`);
            console.log("Unexpected error format:", error.data);
          }
          return;
        } else {
          toast.error("wtf");
        }
      },
    },
  },
  // // bring it back if you want global query error handling
  // queryCache: new QueryCache({
  //   onError: (error) => {
  //     toast.error(`Error: ${error.message}`, {
  //       action: {
  //         label: "retry",
  //         onClick: () => {
  //           queryClient.invalidateQueries();
  //         },
  //       },
  //     });
  //   },
  // }),
});

export const contract = {
  health: healthRouter,
  auth: authRouter,
  users: usersRouter,
};

const link = new OpenAPILink(contract, {
  url: "http://localhost:9997/api",
  fetch: (request, init) => {
    // when not 200-299, it will throw an ORPCError
    return globalThis.fetch(request, {
      ...init,
      credentials: "include", // Include cookies for cross-origin requests
    });
  },
  interceptors: [],
});

export const client: JsonifiedClient<ContractRouterClient<typeof contract>> =
  createORPCClient(link);

export const orpc = createTanstackQueryUtils(client);
