import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { client, orpc } from "@/lib/orpc";
import type { CreateVMRequest } from "@/types/types.gen";
import { toast } from "sonner";
import { useQuery } from "@tanstack/react-query";
import { CONSTANTS } from "@/lib";
import { AlertCircle, Loader2 } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";

interface CreateVMDialogContentProps {
  onSuccess?: () => void;
  onClose: () => void;
}

export function CreateVMDialogContent({
  onSuccess,
  onClose,
}: CreateVMDialogContentProps) {
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<CreateVMRequest>({
    name: "",
    memory: 2048,
    vcpus: 2,
    disk_size: 20,
    os_image: "",
    autostart: false,
  });

  // Fetch ISO files
  const isoQuery = useQuery(
    orpc.iso.listISOs.queryOptions({
      staleTime: CONSTANTS.POLLING_INTERVAL_MS,
    }),
  );

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>,
  ) => {
    const { name, value, type } = e.target;
    setFormData({
      ...formData,
      [name]:
        type === "checkbox"
          ? (e.target as HTMLInputElement).checked
          : type === "number"
            ? parseInt(value)
            : value,
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validation
    if (!formData.name.trim()) {
      toast.error("VM name is required");
      return;
    }

    if (formData.memory <= 0) {
      toast.error("Memory must be greater than 0 MB");
      return;
    }

    if (formData.vcpus <= 0) {
      toast.error("VCPUs must be greater than 0");
      return;
    }

    if (formData.disk_size <= 0) {
      toast.error("Disk size must be greater than 0 GB");
      return;
    }

    setLoading(true);

    try {
      const result = await client.qemu.createVirtualMachine({
        name: formData.name,
        memory: formData.memory,
        vcpus: formData.vcpus,
        disk_size: formData.disk_size,
        os_image: formData.os_image,
        autostart: formData.autostart,
      });

      toast.success(`Virtual machine '${result.name}' created successfully`);

      // Reset form
      setFormData({
        name: "",
        memory: 2048,
        vcpus: 2,
        disk_size: 20,
        os_image: "",
        autostart: false,
      });

      onClose();
      onSuccess?.();
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to create VM",
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="name">VM Name *</Label>
        <Input
          id="name"
          name="name"
          placeholder="my-vm"
          value={formData.name}
          onChange={handleInputChange}
          required
          disabled={loading}
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="memory">Memory (MB) *</Label>
          <Input
            id="memory"
            name="memory"
            type="number"
            min="256"
            max="262144"
            step="256"
            value={formData.memory}
            onChange={handleInputChange}
            required
            disabled={loading}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="vcpus">vCPUs *</Label>
          <Input
            id="vcpus"
            name="vcpus"
            type="number"
            min="1"
            max="256"
            value={formData.vcpus}
            onChange={handleInputChange}
            required
            disabled={loading}
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="disk_size">Disk Size (GB) *</Label>
        <Input
          id="disk_size"
          name="disk_size"
          type="number"
          min="1"
          max="10000"
          step="10"
          value={formData.disk_size}
          onChange={handleInputChange}
          required
          disabled={loading}
        />
      </div>

       <div className="space-y-2">
         <Label htmlFor="os_image">OS Image *</Label>
         <div className="relative">
           {isoQuery.isLoading ? (
             <div className="flex items-center justify-center h-10 border border-input rounded-md bg-muted">
               <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
             </div>
           ) : isoQuery.isError ? (
             <Alert className="border-destructive bg-destructive/10">
               <AlertCircle className="h-4 w-4" />
               <AlertDescription className="text-sm">
                 Failed to load ISO files
               </AlertDescription>
             </Alert>
           ) : (
             <select
               id="os_image"
               name="os_image"
               value={formData.os_image}
               onChange={handleInputChange}
               disabled={loading}
               className="w-full h-10 px-3 py-2 border border-input rounded-md bg-background text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
             >
               <option value="">Select an ISO file or leave empty</option>
               {isoQuery.data?.map((iso) => (
                 <option key={iso.name} value={iso.name}>
                   {iso.name}
                 </option>
               ))}
             </select>
           )}
         </div>
         {!isoQuery.isLoading && !isoQuery.isError && (!isoQuery.data || isoQuery.data.length === 0) && (
           <p className="text-xs text-muted-foreground">
             No ISO files available. Upload one in the ISO Templates section.
           </p>
         )}
       </div>

      <div className="flex items-center space-x-2">
        <input
          id="autostart"
          name="autostart"
          type="checkbox"
          checked={formData.autostart}
          onChange={handleInputChange}
          disabled={loading}
          className="h-4 w-4 rounded border border-input"
        />
        <Label htmlFor="autostart" className="text-sm font-normal">
          Autostart VM on system boot
        </Label>
      </div>

      <div className="flex gap-2 justify-end w-full">
        <Button
          className="flex-1"
          type="button"
          variant="outline"
          onClick={onClose}
          disabled={loading}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={loading} className="flex-1">
          {loading ? "Creating..." : "Create VM"}
        </Button>
      </div>
    </form>
  );
}
