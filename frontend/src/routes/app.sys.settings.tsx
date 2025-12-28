import { useState } from "react";
import {
  AlertCircle,
  Settings,
  Bell,
  Loader2,
  Save,
  TestTube,
  Trash2,
} from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { usePermission } from "@/components/protected-content";
import { RBAC_SETTINGS_MANAGER } from "@/types/types.gen";
import { orpc } from "@/lib/orpc";
import { useQuery, useMutation } from "@tanstack/react-query";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { toast } from "sonner";

export default function SettingsPage() {
  const { hasPermission } = usePermission();

  if (!hasPermission(RBAC_SETTINGS_MANAGER)) {
    return (
      <Alert className="border-destructive bg-destructive/10">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          You don't have permission to access settings.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Settings className="h-6 w-6" />
        <h1 className="text-3xl font-bold">Settings</h1>
      </div>

      <div className="space-y-6">
        <NotificationSettings />
      </div>
    </div>
  );
}

function NotificationSettings() {
  const {
    data: settings,
    isLoading,
    error,
  } = useQuery(orpc.settings.getNotificationSettings.queryOptions({}));

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
          Failed to load notification settings: {error.message}
        </AlertDescription>
      </Alert>
    );
  }

  // Find Discord setting or use defaults
  const discordSetting = settings?.find((s) => s.provider === "discord");

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Bell className="h-5 w-5" />
          <CardTitle className="text-lg">Notification Settings</CardTitle>
        </div>
        <CardDescription>
          Configure notification providers to receive alerts about system
          events.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <DiscordSettingsForm initialSettings={discordSetting} />
      </CardContent>
    </Card>
  );
}

interface DiscordSettings {
  id: number;
  provider: string;
  enabled?: boolean;
  webhook_url?: string;
  notify_on_error?: boolean;
  notify_on_warn?: boolean;
  notify_on_info?: boolean;
  config?: string;
  created_at: string;
  updated_at: string;
}

interface DiscordSettingsFormProps {
  initialSettings?: DiscordSettings;
}

function DiscordSettingsForm({ initialSettings }: DiscordSettingsFormProps) {
  const [enabled, setEnabled] = useState(initialSettings?.enabled ?? false);
  const [webhookUrl, setWebhookUrl] = useState(
    initialSettings?.webhook_url ?? "",
  );
  const [notifyOnError, setNotifyOnError] = useState(
    initialSettings?.notify_on_error ?? true,
  );
  const [notifyOnWarn, setNotifyOnWarn] = useState(
    initialSettings?.notify_on_warn ?? true,
  );
  const [notifyOnInfo, setNotifyOnInfo] = useState(
    initialSettings?.notify_on_info ?? false,
  );

  const upsertMutation = useMutation({
    ...orpc.settings.upsertNotificationSetting.mutationOptions(),
    onSuccess: () => {
      toast.success("Discord settings saved successfully");
    },
  });

  const deleteMutation = useMutation({
    ...orpc.settings.deleteNotificationSetting.mutationOptions(),
    onSuccess: () => {
      toast.success("Discord settings deleted");
      // Reset form
      setEnabled(false);
      setWebhookUrl("");
      setNotifyOnError(true);
      setNotifyOnWarn(true);
      setNotifyOnInfo(false);
    },
  });

  const testMutation = useMutation({
    ...orpc.settings.testNotification.mutationOptions(),
    onSuccess: (data) => {
      toast.success(data.message);
    },
  });

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    upsertMutation.mutate({
      body: {
        provider: "discord",
        enabled,
        webhook_url: webhookUrl,
        notify_on_error: notifyOnError,
        notify_on_warn: notifyOnWarn,
        notify_on_info: notifyOnInfo,
      },
    });
  };

  const handleTest = () => {
    testMutation.mutate({
      params: { provider: "discord" },
    });
  };

  const handleDelete = () => {
    deleteMutation.mutate({
      params: { provider: "discord" },
    });
  };

  const isValidWebhookUrl = webhookUrl.includes("discordapp.com/api/webhooks/");

  return (
    <form onSubmit={handleSave} className="space-y-6">
      <div className="flex items-center justify-between rounded-lg border p-4">
        <div className="space-y-0.5">
          <Label className="text-base font-medium">Discord Webhooks</Label>
          <p className="text-sm text-muted-foreground">
            Receive notifications via Discord webhook
          </p>
        </div>
        <Switch checked={enabled} onCheckedChange={setEnabled} />
      </div>

      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="webhookUrl">Webhook URL</Label>
          <Input
            id="webhookUrl"
            type="url"
            value={webhookUrl}
            onChange={(e) => setWebhookUrl(e.target.value)}
            placeholder="https://discord.com/api/webhooks/..."
            disabled={!enabled}
          />
          {webhookUrl && !isValidWebhookUrl && (
            <p className="text-sm text-destructive">
              Please enter a valid Discord webhook URL
            </p>
          )}
        </div>

        <div className="space-y-3">
          <Label className="text-sm font-medium">Notification Levels</Label>

          <div className="flex items-center justify-between rounded-md border p-3">
            <div className="space-y-0.5">
              <span className="text-sm font-medium">Errors</span>
              <p className="text-xs text-muted-foreground">
                Critical issues that require attention
              </p>
            </div>
            <Switch
              checked={notifyOnError}
              onCheckedChange={setNotifyOnError}
              disabled={!enabled}
            />
          </div>

          <div className="flex items-center justify-between rounded-md border p-3">
            <div className="space-y-0.5">
              <span className="text-sm font-medium">Warnings</span>
              <p className="text-xs text-muted-foreground">
                Potential issues that may need review
              </p>
            </div>
            <Switch
              checked={notifyOnWarn}
              onCheckedChange={setNotifyOnWarn}
              disabled={!enabled}
            />
          </div>

          <div className="flex items-center justify-between rounded-md border p-3">
            <div className="space-y-0.5">
              <span className="text-sm font-medium">Info</span>
              <p className="text-xs text-muted-foreground">
                General information and updates
              </p>
            </div>
            <Switch
              checked={notifyOnInfo}
              onCheckedChange={setNotifyOnInfo}
              disabled={!enabled}
            />
          </div>
        </div>
      </div>

      <div className="flex items-center gap-2">
        <Button
          type="submit"
          disabled={upsertMutation.isPending || (enabled && !isValidWebhookUrl)}
        >
          {upsertMutation.isPending ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Saving...
            </>
          ) : (
            <>
              <Save className="mr-2 h-4 w-4" />
              Save Settings
            </>
          )}
        </Button>

        <Button
          type="button"
          variant="secondary"
          onClick={handleTest}
          disabled={testMutation.isPending || !enabled || !isValidWebhookUrl}
        >
          {testMutation.isPending ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Testing...
            </>
          ) : (
            <>
              <TestTube className="mr-2 h-4 w-4" />
              Test
            </>
          )}
        </Button>

        {initialSettings && (
          <Button
            type="button"
            variant="destructive"
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
          >
            {deleteMutation.isPending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Trash2 className="h-4 w-4" />
            )}
          </Button>
        )}
      </div>
    </form>
  );
}
