import { client } from "@/lib/orpc";
import type { Z } from "@/types";
import type z from "zod";
import { create } from "zustand";

export type HealthType = z.infer<typeof Z.healthSchema>;

type Store = {
  health: HealthType | null;
  setHealth: (session: HealthType | null) => void;
};

export const useHealth = create<Store>()((set) => ({
  health: null,
  setHealth: (session) => set({ health: session }),
}));

// Poll every second to update session info
setInterval(async () => {
  try {
    const health = await client.health.check();
    useHealth.getState().setHealth(health); // i promise it's perfect dw
  } catch (error) {
    useHealth.getState().setHealth(null);
  }
}, 1000);
