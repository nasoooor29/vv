import * as zodTypes from "./zod-types.gen";
import { z } from "zod";
import type * as GeneratedTypes from "./types.gen";

// Re-export everything from types.gen except VirtualMachineWithInfo (which we override)
export type {
  VirtualMachine,
  VirtualMachineInfo,
  CreateVMRequest,
  VMActionResponse,
  Login,
  RBACPolicy,
  HTTPError,
  LogResponse,
  LogRequestData,
  GetLogsResponse,
  LogStatsResponse,
  ClearOldLogsResponse,
  MetricsPeriod,
  ErrorRateByService,
  LogCountByHour,
  LogLevelStats,
  ServiceStats,
  MetricsResponse,
  ServiceMetricsResponse,
  ServiceHealth,
  HealthMetricsResponse,
  StorageDevice,
  MountPoint,
  EnvVars,
  PortainerTemplate,
  TemplateVolume,
  TemplateEnv,
  EnvSelect,
  TemplateLabel,
  TemplatesResponse,
  DeployTemplateRequest,
  TemplateListItem,
  DeployResponse,
  FirewallRule,
  FirewallStatus,
  CreateRuleRequest,
  ReorderRulesRequest,
  FirewallService,
  QemuService,
  StorageService,
  DocsService,
  UsersService,
  ClientInfo,
  DockerService,
  AuthService,
  MetricsService,
  GetMetricsRequest,
  TemplatesService,
  LogsService,
  GetLogsRequest,
  CountLogsByLevelParams,
  CountLogsByServiceGroupParams,
  CountLogsByServiceGroupAndLevelParams,
  CreateLogParams,
  GetAverageLogCountByHourRow,
  GetErrorRateByServiceRow,
  GetLogLevelDistributionParams,
  GetLogLevelDistributionRow,
  GetLogsByLevelParams,
  GetLogsByServiceGroupParams,
  GetLogsByServiceGroupAndLevelParams,
  GetLogsPaginatedParams,
  GetServiceGroupDistributionParams,
  GetServiceGroupDistributionRow,
  UpsertSessionParams,
  Service,
  Health,
  HealthStats,
  CreateUserParams,
  GetByEmailOrUsernameParams,
  GetUserAndSessionByTokenRow,
  UpdateUserParams,
  UpdateUserPasswordParams,
  UpdateUserRoleParams,
  UpsertUserParams,
  Log,
  Notification,
  User,
  UserSession,
} from "./types.gen";

// Re-export constants
export {
  VIR_DOMAIN_NOSTATE,
  VIR_DOMAIN_RUNNING,
  VIR_DOMAIN_BLOCKED,
  VIR_DOMAIN_PAUSED,
  VIR_DOMAIN_SHUTDOWN,
  VIR_DOMAIN_SHUTOFF,
  VIR_DOMAIN_CRASHED,
  VIR_DOMAIN_PMSUSPENDED,
  RBAC_DOCKER_READ,
  RBAC_DOCKER_WRITE,
  RBAC_DOCKER_UPDATE,
  RBAC_DOCKER_DELETE,
  RBAC_QEMU_READ,
  RBAC_QEMU_WRITE,
  RBAC_QEMU_UPDATE,
  RBAC_QEMU_DELETE,
  RBAC_EVENT_VIEWER,
  RBAC_EVENT_MANAGER,
  RBAC_USER_ADMIN,
  RBAC_SETTINGS_MANAGER,
  RBAC_AUDIT_LOG_VIEWER,
  RBAC_HEALTH_CHECKER,
  RBAC_FIREWALL_READ,
  RBAC_FIREWALL_WRITE,
  RBAC_FIREWALL_UPDATE,
  RBAC_FIREWALL_DELETE,
  COOKIE_NAME,
  BYPASS_RBAC_HEADER,
} from "./types.gen";

// Custom VirtualMachineWithInfo type - flat structure matching what API returns
// The Go backend has embedded struct which JSON marshals to flat structure
export interface VirtualMachineWithInfo {
  id: number;
  name: string;
  uuid: string;
  state: number;
  max_mem_kb: number;
  memory_kb: number;
  vcpus: number;
  cpu_time_ns: number;
}

// T namespace that overrides VirtualMachineWithInfo with flat structure
export namespace T {
  // Re-export all types from generated file
  export type VirtualMachine = GeneratedTypes.VirtualMachine;
  export type VirtualMachineInfo = GeneratedTypes.VirtualMachineInfo;
  // Use our custom flat VirtualMachineWithInfo instead of the generated nested one
  export type VirtualMachineWithInfo = {
    id: number;
    name: string;
    uuid: string;
    state: number;
    max_mem_kb: number;
    memory_kb: number;
    vcpus: number;
    cpu_time_ns: number;
  };
  export type CreateVMRequest = GeneratedTypes.CreateVMRequest;
  export type VMActionResponse = GeneratedTypes.VMActionResponse;
  export type Login = GeneratedTypes.Login;
  export type RBACPolicy = GeneratedTypes.RBACPolicy;
  export type HTTPError = GeneratedTypes.HTTPError;
  export type LogResponse = GeneratedTypes.LogResponse;
  export type LogRequestData = GeneratedTypes.LogRequestData;
  export type GetLogsResponse = GeneratedTypes.GetLogsResponse;
  export type LogStatsResponse = GeneratedTypes.LogStatsResponse;
  export type ClearOldLogsResponse = GeneratedTypes.ClearOldLogsResponse;
  export type MetricsPeriod = GeneratedTypes.MetricsPeriod;
  export type ErrorRateByService = GeneratedTypes.ErrorRateByService;
  export type LogCountByHour = GeneratedTypes.LogCountByHour;
  export type LogLevelStats = GeneratedTypes.LogLevelStats;
  export type ServiceStats = GeneratedTypes.ServiceStats;
  export type MetricsResponse = GeneratedTypes.MetricsResponse;
  export type ServiceMetricsResponse = GeneratedTypes.ServiceMetricsResponse;
  export type ServiceHealth = GeneratedTypes.ServiceHealth;
  export type HealthMetricsResponse = GeneratedTypes.HealthMetricsResponse;
  export type StorageDevice = GeneratedTypes.StorageDevice;
  export type MountPoint = GeneratedTypes.MountPoint;
  export type EnvVars = GeneratedTypes.EnvVars;
  export type PortainerTemplate = GeneratedTypes.PortainerTemplate;
  export type TemplateVolume = GeneratedTypes.TemplateVolume;
  export type TemplateEnv = GeneratedTypes.TemplateEnv;
  export type EnvSelect = GeneratedTypes.EnvSelect;
  export type TemplateLabel = GeneratedTypes.TemplateLabel;
  export type TemplatesResponse = GeneratedTypes.TemplatesResponse;
  export type DeployTemplateRequest = GeneratedTypes.DeployTemplateRequest;
  export type TemplateListItem = GeneratedTypes.TemplateListItem;
  export type DeployResponse = GeneratedTypes.DeployResponse;
  export type FirewallRule = GeneratedTypes.FirewallRule;
  export type FirewallStatus = GeneratedTypes.FirewallStatus;
  export type CreateRuleRequest = GeneratedTypes.CreateRuleRequest;
  export type ReorderRulesRequest = GeneratedTypes.ReorderRulesRequest;
  export type FirewallService = GeneratedTypes.FirewallService;
  export type QemuService = GeneratedTypes.QemuService;
  export type StorageService = GeneratedTypes.StorageService;
  export type DocsService = GeneratedTypes.DocsService;
  export type UsersService = GeneratedTypes.UsersService;
  export type ClientInfo = GeneratedTypes.ClientInfo;
  export type DockerService = GeneratedTypes.DockerService;
  export type AuthService = GeneratedTypes.AuthService;
  export type MetricsService = GeneratedTypes.MetricsService;
  export type GetMetricsRequest = GeneratedTypes.GetMetricsRequest;
  export type TemplatesService = GeneratedTypes.TemplatesService;
  export type LogsService = GeneratedTypes.LogsService;
  export type GetLogsRequest = GeneratedTypes.GetLogsRequest;
  export type CountLogsByLevelParams = GeneratedTypes.CountLogsByLevelParams;
  export type CountLogsByServiceGroupParams = GeneratedTypes.CountLogsByServiceGroupParams;
  export type CountLogsByServiceGroupAndLevelParams = GeneratedTypes.CountLogsByServiceGroupAndLevelParams;
  export type CreateLogParams = GeneratedTypes.CreateLogParams;
  export type GetAverageLogCountByHourRow = GeneratedTypes.GetAverageLogCountByHourRow;
  export type GetErrorRateByServiceRow = GeneratedTypes.GetErrorRateByServiceRow;
  export type GetLogLevelDistributionParams = GeneratedTypes.GetLogLevelDistributionParams;
  export type GetLogLevelDistributionRow = GeneratedTypes.GetLogLevelDistributionRow;
  export type GetLogsByLevelParams = GeneratedTypes.GetLogsByLevelParams;
  export type GetLogsByServiceGroupParams = GeneratedTypes.GetLogsByServiceGroupParams;
  export type GetLogsByServiceGroupAndLevelParams = GeneratedTypes.GetLogsByServiceGroupAndLevelParams;
  export type GetLogsPaginatedParams = GeneratedTypes.GetLogsPaginatedParams;
  export type GetServiceGroupDistributionParams = GeneratedTypes.GetServiceGroupDistributionParams;
  export type GetServiceGroupDistributionRow = GeneratedTypes.GetServiceGroupDistributionRow;
  export type UpsertSessionParams = GeneratedTypes.UpsertSessionParams;
  export type Service = GeneratedTypes.Service;
  export type Health = GeneratedTypes.Health;
  export type HealthStats = GeneratedTypes.HealthStats;
  export type CreateUserParams = GeneratedTypes.CreateUserParams;
  export type GetByEmailOrUsernameParams = GeneratedTypes.GetByEmailOrUsernameParams;
  export type GetUserAndSessionByTokenRow = GeneratedTypes.GetUserAndSessionByTokenRow;
  export type UpdateUserParams = GeneratedTypes.UpdateUserParams;
  export type UpdateUserPasswordParams = GeneratedTypes.UpdateUserPasswordParams;
  export type UpdateUserRoleParams = GeneratedTypes.UpdateUserRoleParams;
  export type UpsertUserParams = GeneratedTypes.UpsertUserParams;
  export type Log = GeneratedTypes.Log;
  export type Notification = GeneratedTypes.Notification;
  export type User = GeneratedTypes.User;
  export type UserSession = GeneratedTypes.UserSession;
}

// Docker SDK Types (external types from docker/docker library - not auto-generated)
const dockerClientInfoSchema = z.object({
  id: z.number(),
  status: z.string(),
});

const dockerPortSchema = z.object({
  IP: z.string().optional(),
  PrivatePort: z.number(),
  PublicPort: z.number().optional(),
  Type: z.string(),
});

const dockerLabelsSchema = z.record(z.string(), z.string());

const dockerHostConfigSchema = z.object({
  NetworkMode: z.string(),
});

const dockerNetworkSchema = z.object({
  IPAMConfig: z.any().optional(),
  Links: z.any().optional(),
  Aliases: z.any().optional(),
  MacAddress: z.string().optional(),
  DriverOpts: z.any().optional(),
  GwPriority: z.number().optional(),
  NetworkID: z.string().optional(),
  EndpointID: z.string().optional(),
  Gateway: z.string().optional(),
  IPAddress: z.string().optional(),
  IPPrefixLen: z.number().optional(),
  IPv6Gateway: z.string().optional(),
  GlobalIPv6Address: z.string().optional(),
  GlobalIPv6PrefixLen: z.number().optional(),
  DNSNames: z.any().optional(),
});

const dockerNetworkSettingsSchema = z.object({
  Networks: z.record(z.string(), dockerNetworkSchema),
});

const dockerMountSchema = z.object({
  Type: z.string(),
  Name: z.string().optional(),
  Source: z.string(),
  Destination: z.string(),
  Driver: z.string().optional(),
  Mode: z.string(),
  RW: z.boolean(),
  Propagation: z.string(),
});

const dockerContainerSchema = z.object({
  Id: z.string(),
  Names: z.array(z.string()),
  Image: z.string(),
  ImageID: z.string(),
  Command: z.string(),
  Created: z.number(),
  Ports: z.array(dockerPortSchema),
  Labels: dockerLabelsSchema,
  State: z.string(),
  Status: z.string(),
  HostConfig: dockerHostConfigSchema,
  NetworkSettings: dockerNetworkSettingsSchema,
  Mounts: z.array(dockerMountSchema),
});

const dockerImageSchema = z.object({
  Containers: z.number(),
  Created: z.number(),
  Id: z.string(),
  Labels: dockerLabelsSchema.optional(),
  ParentId: z.string(),
  RepoDigests: z.array(z.string()).nullable(),
  RepoTags: z.array(z.string()).nullable(),
  SharedSize: z.number(),
  Size: z.number(),
});

const containerCreateResponseSchema = z.object({
  Id: z.string(),
  Warnings: z.array(z.string()).nullable(),
});

// Override virtualMachineWithInfoSchema to flatten the nested VirtualMachineInfo
// The backend returns flat structure, but the Go type has embedded struct
const virtualMachineWithInfoSchema = z.object({
  id: z.number(),
  name: z.string(),
  uuid: z.string(),
  state: z.number(),
  max_mem_kb: z.number(),
  memory_kb: z.number(),
  vcpus: z.number(),
  cpu_time_ns: z.number(),
});

export const Z = {
  ...zodTypes,
  // Docker SDK types (external, not auto-generated)
  dockerClientInfoSchema,
  dockerPortSchema,
  dockerLabelsSchema,
  dockerHostConfigSchema,
  dockerNetworkSchema,
  dockerNetworkSettingsSchema,
  dockerMountSchema,
  dockerContainerSchema,
  dockerImageSchema,
  containerCreateResponseSchema,
  // Override for flattened VM type
  virtualMachineWithInfoSchema,
  isHttpError: (obj: any) => {
    return (
      obj &&
      typeof obj === "object" &&
      "message" in obj &&
      typeof obj.message === "string"
    );
  },
};
