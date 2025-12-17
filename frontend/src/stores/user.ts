import { client } from "@/lib/orpc";
import type { Z } from "@/types";
import { ORPCError } from "@orpc/client";
import type z from "zod";
import { create } from "zustand";
import { persist } from "zustand/middleware";

export type SessionType = z.infer<typeof Z.getUserAndSessionByTokenRowSchema>;

type Store = {
  session: SessionType | null;
  setSession: (session: SessionType | null) => void;
  clearSession: () => void;
};

export const useSession = create<Store>()(
  persist(
    (set) => ({
      clearSession: () => set({ session: null }),
      session: null,
      setSession: (session) => set({ session }),
    }),
    {
      name: "user-session", // name of the item in the storage (must be unique)
    },
  ),
);

// Poll every second to update session info AND initialize on load
const initSession = async () => {
  try {
    const pathName = window.location.pathname;
    console.log("Current path:", pathName);
    // if on auth page and no session, skip fetching
    if (pathName.startsWith("/auth") && !useSession.getState().session) {
      console.log("On auth page with no session, skipping session fetch");
      return;
    }
    const session = await client.auth.me();
    if (session) {
      useSession.getState().setSession(session);
      console.log("Session initialized:", session);
    } else {
      useSession.getState().clearSession();
    }
  } catch (error) {
    if (error instanceof ORPCError) {
      if (error.status === 401) {
        useSession.getState().clearSession();
        console.log("Session cleared due to 401 Unauthorized");
      } else {
        console.error("Error fetching session:", error);
      }
    }
  }
};

// Initialize session immediately
initSession();

// Then poll every second
setInterval(initSession, 10000);
