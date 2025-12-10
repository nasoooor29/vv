import { Z } from "@/types";
import { base } from "./general";

export const authRouter = {
  login: base
    .route({
      method: "POST",
      path: "/auth/login",
    })
    .input(Z.loginSchema)
    .output(Z.userSchema),

  logout: base.route({
    method: "POST",
    path: "/auth/logout",
  }),
  register: base
    .route({
      method: "POST",
      path: "/auth/register",
    })
    .input(Z.upsertUserParamsSchema)
    .output(Z.userSchema),
  me: base
    .route({
      method: "GET",
      path: "/auth/me",
    })
    .output(Z.getUserAndSessionByTokenRowSchema),
};
