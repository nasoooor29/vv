import { useState } from "react";
import { orpc, queryClient } from "@/lib/orpc";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import {
  AlertCircle,
  Disc3,
  Plus,
  Download,
  Trash2,
  Upload,
} from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { usePermission } from "@/components/protected-content";
import { RBAC_QEMU_READ, RBAC_QEMU_WRITE, RBAC_QEMU_DELETE } from "@/types/types.gen";
import { useMutation, useQuery } from "@tanstack/react-query";
import { CONSTANTS } from "@/lib";
import { toast } from "sonner";
import { formatBytes } from "@/lib/utils";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { format } from "date-fns";

export default function ISOPage() {
  const { hasPermission } = usePermission();
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [fileToDelete, setFileToDelete] = useState<string | null>(null);
  const [isUploadDialogOpen, setIsUploadDialogOpen] = useState(false);

  // Queries
  const isoListQuery = useQuery(
    orpc.iso.listISOs.queryOptions({
      staleTime: CONSTANTS.POLLING_INTERVAL_MS,
    }),
  );

  // Mutations
  const uploadMutation = useMutation(
    orpc.iso.uploadISO.mutationOptions({
      onSuccess() {
        toast.success("ISO file uploaded successfully");
        setSelectedFile(null);
        setIsUploadDialogOpen(false);
        queryClient.invalidateQueries();
      },
      onError(error) {
        console.error("Upload error:", error);
        toast.error("Failed to upload ISO file");
      },
    }),
  );

  const deleteMutation = useMutation(
    orpc.iso.deleteISO.mutationOptions({
      onSuccess() {
        toast.success("ISO file deleted successfully");
        setFileToDelete(null);
        queryClient.invalidateQueries();
      },
      onError(error) {
        console.error("Delete error:", error);
        toast.error("Failed to delete ISO file");
      },
    }),
  );

  const handleUpload = async () => {
    if (!selectedFile) {
      toast.error("Please select a file to upload");
      return;
    }

    if (selectedFile.size > 5 * 1024 * 1024 * 1024) {
      toast.error("File size exceeds 5GB limit");
      return;
    }

    uploadMutation.mutate({
      file: selectedFile,
    });
  };

  const handleDelete = async (filename: string) => {
    deleteMutation.mutate({
      params: { filename },
    });
  };

  const handleDownload = async (filename: string) => {
    try {
      // Get the download URL and fetch the file
      const response = await fetch(
        `http://localhost:9999/api/iso/${encodeURIComponent(filename)}/download`,
        {
          credentials: "include",
        }
      );

      if (!response.ok) {
        toast.error("Failed to download ISO file");
        return;
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      toast.success(`Downloading ${filename}`);
    } catch (error) {
      console.error("Download error:", error);
      toast.error("Failed to download ISO file");
    }
  };

  if (!hasPermission(RBAC_QEMU_READ)) {
    return (
      <Alert className="border-destructive bg-destructive/10">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          You don't have permission to access ISO files.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Disc3 className="h-6 w-6" />
          <h1 className="text-3xl font-bold">ISO Templates</h1>
        </div>
        {hasPermission(RBAC_QEMU_WRITE) && (
          <Button
            onClick={() => setIsUploadDialogOpen(true)}
            className="gap-2"
          >
            <Plus className="h-4 w-4" />
            Upload ISO
          </Button>
        )}
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Disc3 className="h-5 w-5" />
            Available ISO Files
          </CardTitle>
          <div className="text-sm text-muted-foreground">
            {isoListQuery.data?.length || 0} files
          </div>
        </CardHeader>
        <CardContent>
          {isoListQuery.isLoading ? (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-20 w-full" />
              ))}
            </div>
          ) : isoListQuery.isError ? (
            <Alert className="border-destructive bg-destructive/10">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Failed to load ISO files. Please try again later.
              </AlertDescription>
            </Alert>
          ) : !isoListQuery.data || isoListQuery.data.length === 0 ? (
            <div className="py-12 text-center text-muted-foreground">
              <Disc3 className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No ISO files found</p>
              {hasPermission(RBAC_QEMU_WRITE) && (
                <p className="text-sm mt-2">
                  Start by uploading an ISO file to use as a template
                </p>
              )}
            </div>
          ) : (
            <div className="space-y-3">
              {isoListQuery.data.map((iso) => {
                const modified = new Date(iso.modified);
                return (
                  <div
                    key={iso.name}
                    className="border border-border rounded-lg p-4 flex items-center justify-between hover:bg-accent/50 transition-colors"
                  >
                    <div className="flex-1 min-w-0">
                      <h3 className="font-semibold text-sm truncate">
                        {iso.name}
                      </h3>
                      <div className="flex gap-4 text-xs text-muted-foreground mt-1">
                        <span>Size: {formatBytes(iso.size)}</span>
                        <span>Modified: {format(modified, "MMM dd, yyyy HH:mm")}</span>
                      </div>
                    </div>

                    <div className="flex gap-2 ml-4">
                      {hasPermission(RBAC_QEMU_READ) && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleDownload(iso.name)}
                          className="gap-2"
                        >
                          <Download className="h-4 w-4" />
                          <span className="hidden sm:inline">Download</span>
                        </Button>
                      )}

                      {hasPermission(RBAC_QEMU_DELETE) && (
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => setFileToDelete(iso.name)}
                          className="gap-2"
                        >
                          <Trash2 className="h-4 w-4" />
                          <span className="hidden sm:inline">Delete</span>
                        </Button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Upload Dialog */}
      <Dialog open={isUploadDialogOpen} onOpenChange={setIsUploadDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Upload className="h-5 w-5" />
              Upload ISO File
            </DialogTitle>
            <DialogDescription>
              Select an ISO file to upload as a template. Maximum size: 5GB
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="border-2 border-dashed border-border rounded-lg p-8 text-center hover:border-primary/50 transition-colors">
              <Input
                type="file"
                accept=".iso"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) {
                    if (file.size > 5 * 1024 * 1024 * 1024) {
                      toast.error("File size exceeds 5GB limit");
                      return;
                    }
                    setSelectedFile(file);
                  }
                }}
                className="hidden"
                id="file-input"
              />
              <label
                htmlFor="file-input"
                className="cursor-pointer block space-y-2"
              >
                <Disc3 className="h-8 w-8 mx-auto opacity-50" />
                <p className="text-sm font-medium">
                  {selectedFile ? selectedFile.name : "Click to select an ISO file"}
                </p>
                {selectedFile && (
                  <p className="text-xs text-muted-foreground">
                    Size: {formatBytes(selectedFile.size)}
                  </p>
                )}
              </label>
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setIsUploadDialogOpen(false);
                setSelectedFile(null);
              }}
            >
              Cancel
            </Button>
            <Button
              onClick={handleUpload}
              disabled={!selectedFile || uploadMutation.isPending}
            >
              {uploadMutation.isPending ? "Uploading..." : "Upload"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog
        open={fileToDelete !== null}
        onOpenChange={(open) => !open && setFileToDelete(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete ISO File</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete <strong>{fileToDelete}</strong>?
              This action cannot be undone.
            </DialogDescription>
          </DialogHeader>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setFileToDelete(null)}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => fileToDelete && handleDelete(fileToDelete)}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
