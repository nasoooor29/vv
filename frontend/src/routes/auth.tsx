import { orpc } from "@/lib/orpc";
import { ORPCError } from "@orpc/client";
import { useQuery } from "@tanstack/react-query";
import { Outlet, useNavigate } from "react-router";
import { useEffect } from "react";

export default function AuthLayout() {
  const data = useQuery(
    orpc.auth.me.queryOptions({
      staleTime: 1 * 1000, // 1 second
    }),
  );
  const navigate = useNavigate();

  useEffect(() => {
    if (!data.isLoading && data.data) {
      navigate("/app", { replace: true });
    }
  }, [data.data, data.isLoading, navigate]);

  if (data.error && data.error instanceof ORPCError) {
    if (data.error.status !== 401) {
      // User is not authenticated, stay on auth layout
      console.error("Error fetching session:", data.error);
    }
  }

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="w-full">
        <Outlet />
      </div>
    </div>
  );
}
