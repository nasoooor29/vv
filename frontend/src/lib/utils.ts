import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { client } from "./orpc";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export async function redirectToOAuthProvider(provider: string) {
  const health = await client.health.check();
  const redirectUri = encodeURIComponent(
    `${window.location.origin}/app/dashboard`,
  );
  console.log("Redirecting to OAuth provider:", provider);
  console.log("Health base URL:", health);
  window.location.href = `${health.base_url}/api/auth/oauth/${provider}?redirect_to=${redirectUri}`;
}

// Utility function to format time difference
export function formatTimeDifference(timestamp: number | Date): string {
  const now = Date.now();
  const updated = new Date(timestamp).getTime();
  const diffMs = now - updated;
  const diffSec = Math.floor(diffMs / 1000);

  if (diffSec < 60) return `${diffSec} sec ago`;
  const diffMin = Math.floor(diffSec / 60);
  if (diffMin < 60) return `${diffMin} min ago`;
  const diffHr = Math.floor(diffMin / 60);
  if (diffHr < 24) return `${diffHr} hr ago`;
  const diffDay = Math.floor(diffHr / 24);
  return `${diffDay} day${diffDay > 1 ? "s" : ""} ago`;
}
