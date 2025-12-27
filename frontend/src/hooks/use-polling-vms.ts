import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { orpc } from "@/lib/orpc";

interface UsePollingVMsOptions {
  enabled?: boolean;
  interval?: number; // in milliseconds
}

/**
 * Custom hook for polling virtual machine metrics in real-time
 * Uses React Query with refetchInterval for automatic polling
 */
export function usePollingVMs(options: UsePollingVMsOptions = {}) {
  const { enabled = true, interval = 5000 } = options; // Default 5 second polling

  const query = useQuery(
    orpc.qemu.getVirtualMachinesInfo.queryOptions({
      refetchInterval: enabled ? interval : false,
      staleTime: interval / 2, // Data is stale after half the polling interval
    }),
  );

  return query;
}

/**
 * Custom hook for polling a specific VM's metrics
 */
export function usePollingVM(
  uuid: string,
  options: UsePollingVMsOptions = {},
) {
  const { enabled = true, interval = 5000 } = options;

  const query = useQuery(
    orpc.qemu.getVirtualMachineInfo.queryOptions(
      { params: { uuid } },
      {
        refetchInterval: enabled ? interval : false,
        staleTime: interval / 2,
        enabled: !!uuid && enabled,
      },
    ),
  );

  return query;
}
