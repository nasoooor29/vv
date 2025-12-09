import { Z } from "@/types";
import { base } from "./general";

export const authRouter = {
  login: base
    .route({
      method: "POST",
      path: "/login",
    })
    .input(Z.getByEmailOrUsernameParamsSchema)
    .output(Z.withHttpError(Z.userSchema)),

  logout: base.route({
    method: "POST",
    path: "/logout",
  }),
  register: base
    .route({
      method: "POST",
      path: "/register",
    })
    .input(Z.upsertUserParamsSchema)
    .output(Z.withHttpError(Z.userSchema)),
};
