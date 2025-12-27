import { base, detailed } from "./general";
import { Z } from "@/types";
import { z } from "zod";

export const qemuRouter = {
  // List all virtual machines
  listVirtualMachines: base
    .route({
      method: "GET",
      path: "/qemu/virtual-machines",
    })
    .output(Z.virtualMachineSchema.array()),

  // Get detailed info for all virtual machines
  getVirtualMachinesInfo: base
    .route({
      method: "GET",
      path: "/qemu/virtual-machines/info",
    })
    .output(Z.virtualMachineWithInfoSchema.array()),

  // Get specific virtual machine
  getVirtualMachine: base
    .route({
      method: "GET",
      path: "/qemu/virtual-machines/{uuid}",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { uuid: z.string() },
      }),
    )
    .output(Z.virtualMachineSchema),

  // Get specific virtual machine detailed info
  getVirtualMachineInfo: base
    .route({
      method: "GET",
      path: "/qemu/virtual-machines/{uuid}/info",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { uuid: z.string() },
      }),
    )
    .output(Z.virtualMachineWithInfoSchema),

  // Create virtual machine
  createVirtualMachine: base
    .route({
      method: "POST",
      path: "/qemu/virtual-machines",
    })
    .input(Z.createVmRequestSchema)
    .output(Z.virtualMachineSchema),

  // Start virtual machine
  startVirtualMachine: base
    .route({
      method: "POST",
      path: "/qemu/virtual-machines/{uuid}/start",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { uuid: z.string() },
      }),
    )
    .output(Z.vmActionResponseSchema),

  // Reboot virtual machine
  rebootVirtualMachine: base
    .route({
      method: "POST",
      path: "/qemu/virtual-machines/{uuid}/reboot",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { uuid: z.string() },
      }),
    )
    .output(Z.vmActionResponseSchema),

  // Shutdown virtual machine
  shutdownVirtualMachine: base
    .route({
      method: "POST",
      path: "/qemu/virtual-machines/{uuid}/shutdown",
      inputStructure: "detailed",
    })
    .input(
      detailed({
        params: { uuid: z.string() },
      }),
    )
    .output(Z.vmActionResponseSchema),
};
