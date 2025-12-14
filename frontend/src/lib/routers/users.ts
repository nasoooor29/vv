import { Z } from "@/types";
import { base } from "./general";
import { z } from "zod";

export const usersRouter = {
  listUsers: base
    .route({
      method: "GET",
      path: "/users",
    })
    .output(z.array(Z.userSchema)),

  createUser: base
    .route({
      method: "POST",
      path: "/users",
    })
    .input(Z.createUserParamsSchema)
    .output(Z.userSchema),

  updateUser: base
    .route({
      method: "PUT",
      path: "/users/:id",
    })
    .input(
      z.object({
        id: z.number(),
        username: z.string(),
        email: z.string(),
        role: z.string().optional(),
      })
    )
    .output(Z.userSchema),

  deleteUser: base
    .route({
      method: "DELETE",
      path: "/users/:id",
    })
    .input(z.object({ id: z.number() })),

  updateUserRole: base
    .route({
      method: "PATCH",
      path: "/users/:id/role",
    })
    .input(
      z.object({
        id: z.number(),
        role: z.string(),
      })
    )
    .output(Z.userSchema),
};
