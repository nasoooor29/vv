import { useState } from "react";
import {
  AlertCircle,
  Shield,
  Plus,
  Trash2,
  Loader2,
  ChevronUp,
  ChevronDown,
} from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { usePermission } from "@/components/protected-content";
import {
  RBAC_FIREWALL_READ,
  RBAC_FIREWALL_WRITE,
  RBAC_FIREWALL_DELETE,
  RBAC_FIREWALL_UPDATE,
} from "@/types/types.gen";
import { orpc } from "@/lib/orpc";
import { useQuery, useMutation } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export default function NetworkingPage() {
  const { hasPermission } = usePermission();
  const [dialogOpen, setDialogOpen] = useState(false);

  if (!hasPermission(RBAC_FIREWALL_READ)) {
    return (
      <Alert className="border-destructive bg-destructive/10">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          You don't have permission to access firewall settings.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Shield className="h-6 w-6" />
        <h1 className="text-3xl font-bold">Firewall Management</h1>
      </div>

      <div className="space-y-6">
        <FirewallStatus />
        <FirewallRules
          dialogOpen={dialogOpen}
          setDialogOpen={setDialogOpen}
          hasWritePermission={hasPermission(RBAC_FIREWALL_WRITE)}
          hasDeletePermission={hasPermission(RBAC_FIREWALL_DELETE)}
          hasUpdatePermission={hasPermission(RBAC_FIREWALL_UPDATE)}
        />
      </div>
    </div>
  );
}

function FirewallStatus() {
  const { data: status, isLoading, error } = useQuery(
    orpc.firewall.status.queryOptions({})
  );

  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-6">
          <Loader2 className="h-6 w-6 animate-spin" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Alert className="border-destructive bg-destructive/10">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load firewall status: {error.message}
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Firewall Status</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-4">
          <Badge variant={status?.enabled ? "default" : "secondary"}>
            {status?.enabled ? "Active" : "Inactive"}
          </Badge>
          <span className="text-muted-foreground">
            Table: <code className="text-foreground">{status?.table_name}</code>
          </span>
          <span className="text-muted-foreground">
            Rules: <span className="text-foreground">{status?.rule_count}</span>
          </span>
        </div>
      </CardContent>
    </Card>
  );
}

interface FirewallRulesProps {
  dialogOpen: boolean;
  setDialogOpen: (open: boolean) => void;
  hasWritePermission: boolean;
  hasDeletePermission: boolean;
  hasUpdatePermission: boolean;
}

function FirewallRules({
  dialogOpen,
  setDialogOpen,
  hasWritePermission,
  hasDeletePermission,
  hasUpdatePermission,
}: FirewallRulesProps) {
  const {
    data: rules,
    isLoading,
    error,
  } = useQuery(orpc.firewall.rules.queryOptions({}));
  const deleteMutation = useMutation(orpc.firewall.deleteRule.mutationOptions());
  const reorderMutation = useMutation(orpc.firewall.reorderRules.mutationOptions());

  const handleDelete = (handle: number) => {
    deleteMutation.mutate({ params: { handle: handle.toString() } });
  };

  const handleMoveUp = (chain: string, index: number) => {
    if (!rules || index === 0) return;

    // Get rules for this chain
    const chainRules = rules.filter((r) => r.chain === chain);
    if (index === 0) return;

    // Swap with previous rule
    const handles = chainRules.map((r) => r.handle);
    [handles[index - 1], handles[index]] = [handles[index], handles[index - 1]];

    reorderMutation.mutate({
      body: { chain: chain as "input" | "forward" | "output", handles },
    });
  };

  const handleMoveDown = (chain: string, index: number) => {
    if (!rules) return;

    // Get rules for this chain
    const chainRules = rules.filter((r) => r.chain === chain);
    if (index >= chainRules.length - 1) return;

    // Swap with next rule
    const handles = chainRules.map((r) => r.handle);
    [handles[index], handles[index + 1]] = [handles[index + 1], handles[index]];

    reorderMutation.mutate({
      body: { chain: chain as "input" | "forward" | "output", handles },
    });
  };

  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-6">
          <Loader2 className="h-6 w-6 animate-spin" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Alert className="border-destructive bg-destructive/10">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load firewall rules: {error.message}
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="text-lg">Firewall Rules</CardTitle>
        {hasWritePermission && (
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button size="sm">
                <Plus className="mr-2 h-4 w-4" />
                Add Rule
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Add Firewall Rule</DialogTitle>
              </DialogHeader>
              <AddRuleForm onSuccess={() => setDialogOpen(false)} />
            </DialogContent>
          </Dialog>
        )}
      </CardHeader>
      <CardContent>
        {rules && rules.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                {hasUpdatePermission && <TableHead className="w-[80px]">Order</TableHead>}
                <TableHead>Chain</TableHead>
                <TableHead>Protocol</TableHead>
                <TableHead>Port</TableHead>
                <TableHead>Source IP</TableHead>
                <TableHead>Action</TableHead>
                {hasDeletePermission && (
                  <TableHead className="w-[80px]">Delete</TableHead>
                )}
              </TableRow>
            </TableHeader>
            <TableBody>
              {(["input", "forward", "output"] as const).map((chainName) => {
                const chainRules = rules.filter((r) => r.chain === chainName);
                return chainRules.map((rule, index) => (
                  <TableRow key={rule.handle}>
                    {hasUpdatePermission && (
                      <TableCell>
                        <div className="flex gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={() => handleMoveUp(chainName, index)}
                            disabled={index === 0 || reorderMutation.isPending}
                          >
                            <ChevronUp className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={() => handleMoveDown(chainName, index)}
                            disabled={
                              index >= chainRules.length - 1 ||
                              reorderMutation.isPending
                            }
                          >
                            <ChevronDown className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    )}
                    <TableCell>
                      <Badge variant="outline">{rule.chain}</Badge>
                    </TableCell>
                    <TableCell>{rule.protocol || "-"}</TableCell>
                    <TableCell>{rule.port || "-"}</TableCell>
                    <TableCell>{rule.source_ip || "-"}</TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          rule.action === "accept" ? "default" : "destructive"
                        }
                      >
                        {rule.action}
                      </Badge>
                    </TableCell>
                    {hasDeletePermission && (
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDelete(rule.handle)}
                          disabled={deleteMutation.isPending}
                        >
                          {deleteMutation.isPending ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <Trash2 className="h-4 w-4 text-destructive" />
                          )}
                        </Button>
                      </TableCell>
                    )}
                  </TableRow>
                ));
              })}
            </TableBody>
          </Table>
        ) : (
          <p className="py-4 text-center text-muted-foreground">
            No firewall rules configured
          </p>
        )}
      </CardContent>
    </Card>
  );
}

interface AddRuleFormProps {
  onSuccess: () => void;
}

function AddRuleForm({ onSuccess }: AddRuleFormProps) {
  const [chain, setChain] = useState<"input" | "forward" | "output">("input");
  const [protocol, setProtocol] = useState<"tcp" | "udp" | "any">("any");
  const [port, setPort] = useState("");
  const [sourceIp, setSourceIp] = useState("");
  const [action, setAction] = useState<"accept" | "drop">("accept");

  // Validation: port requires a specific protocol
  const portRequiresProtocol = port !== "" && protocol === "any";

  const addMutation = useMutation({
    ...orpc.firewall.addRule.mutationOptions(),
    onSuccess: () => {
      onSuccess();
      // Reset form
      setChain("input");
      setProtocol("any");
      setPort("");
      setSourceIp("");
      setAction("accept");
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (portRequiresProtocol) return;
    addMutation.mutate({
      body: {
        chain,
        protocol: protocol === "any" ? undefined : protocol,
        port: port ? parseInt(port, 10) : undefined,
        source_ip: sourceIp || undefined,
        action,
      },
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="chain">Chain</Label>
        <Select
          value={chain}
          onValueChange={(v) => setChain(v as "input" | "forward" | "output")}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select chain" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="input">Input</SelectItem>
            <SelectItem value="forward">Forward</SelectItem>
            <SelectItem value="output">Output</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="protocol">Protocol</Label>
        <Select
          value={protocol}
          onValueChange={(v) => setProtocol(v as "tcp" | "udp" | "any")}
        >
          <SelectTrigger>
            <SelectValue placeholder="Any protocol" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="any">Any</SelectItem>
            <SelectItem value="tcp">TCP</SelectItem>
            <SelectItem value="udp">UDP</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="port">Port (optional)</Label>
        <Input
          id="port"
          type="number"
          min="1"
          max="65535"
          value={port}
          onChange={(e) => setPort(e.target.value)}
          placeholder="e.g., 22, 80, 443"
        />
        {portRequiresProtocol && (
          <p className="text-sm text-destructive">
            Protocol (TCP or UDP) is required when specifying a port
          </p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="sourceIp">Source IP (optional)</Label>
        <Input
          id="sourceIp"
          value={sourceIp}
          onChange={(e) => setSourceIp(e.target.value)}
          placeholder="e.g., 192.168.1.100"
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="action">Action</Label>
        <Select
          value={action}
          onValueChange={(v) => setAction(v as "accept" | "drop")}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select action" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="accept">Accept</SelectItem>
            <SelectItem value="drop">Drop</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <Button
        type="submit"
        className="w-full"
        disabled={addMutation.isPending || portRequiresProtocol}
      >
        {addMutation.isPending ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            Adding...
          </>
        ) : (
          "Add Rule"
        )}
      </Button>
    </form>
  );
}
