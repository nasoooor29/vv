import { base } from "./general";
import { Z } from "@/types";

export const storageRouter = {
  devices: base
    .route({
      method: "GET",
      path: "/storage/devices",
    })
    .output(Z.storageDeviceSchema.array()),

  mountPoints: base
    .route({
      method: "GET",
      path: "/storage/mount-points",
    })
    .output(Z.mountPointSchema.array()),
};
