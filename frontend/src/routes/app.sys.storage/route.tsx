import StorageDevicesPage from "./devices";
import MountPointsPage from "./mount-points";
import { AlertCircle, HardDrive } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { usePermission } from "@/components/protected-content";
import { RBAC_SETTINGS_MANAGER } from "@/types/types.gen";

export default function StoragePage() {
  const { hasPermission } = usePermission();

  if (!hasPermission(RBAC_SETTINGS_MANAGER)) {
    return (
      <Alert className="border-destructive bg-destructive/10">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          You don't have permission to access storage settings.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <HardDrive className="h-6 w-6" />
        <h1 className="text-3xl font-bold">Storage Management</h1>
      </div>

      <div className="space-y-6">
        <StorageDevicesPage />
        <MountPointsPage />
      </div>
    </div>
  );
}
