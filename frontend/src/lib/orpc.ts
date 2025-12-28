import type { JsonifiedClient } from "@orpc/openapi-client";
import type { ContractRouterClient } from "@orpc/contract";
import { createORPCClient, ORPCError } from "@orpc/client";
import { OpenAPILink } from "@orpc/openapi-client/fetch";
import { createTanstackQueryUtils } from "@orpc/tanstack-query";
import { MutationCache, QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { healthRouter } from "./routers/health";
import { authRouter } from "./routers/auth";
import { storageRouter } from "./routers/storage";
import { usersRouter } from "./routers/users";
import { logsRouter } from "./routers/logs";
import { metricsRouter } from "./routers/metrics";
import { qemuRouter } from "./routers/qemu";
import { isoRouter } from "./routers/iso";
import { dockerRouter } from "./routers/docker";
import { firewallRouter } from "./routers/firewall";
import { templatesRouter } from "./routers/templates";
import { settingsRouter } from "./routers/settings";
import { Z } from "@/types";

export const queryClient = new QueryClient({
  mutationCache: new MutationCache({
    onSuccess(data, variables, _onMutateResult, mutation, _context) {
      console.log("Mutation successful:", { data, variables, mutation });
      queryClient.invalidateQueries();
    },
  }),
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
          toast.error("backend is probably down");
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
  storage: storageRouter,
  users: usersRouter,
  logs: logsRouter,
  metrics: metricsRouter,
  qemu: qemuRouter,
  iso: isoRouter,
  docker: dockerRouter,
  firewall: firewallRouter,
  templates: templatesRouter,
  settings: settingsRouter,
};

const link = new OpenAPILink(contract, {
  url: "http://localhost:9999/api",
  fetch: (request, init) => {
    // when not 200-299, it will throw an ORPCError
    return globalThis.fetch(request, {
      ...init,
      credentials: "include", // Include cookies for cross-origin requests
    });
  },
  plugins: [
    // new ResponseValidationPlugin(contract),
    // new RequestValidationPlugin(contract),
  ],
  interceptors: [],
});

export const client: JsonifiedClient<ContractRouterClient<typeof contract>> =
  createORPCClient(link);

export const orpc = createTanstackQueryUtils(client);
