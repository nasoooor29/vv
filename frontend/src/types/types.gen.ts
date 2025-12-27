/**
 * VirtualMachine represents a QEMU virtual machine
 */
export interface VirtualMachine {
  id: number /* int32 */;
  name: string;
  uuid: string;
}
/**
 * VirtualMachineInfo contains detailed information about a virtual machine
 */
export interface VirtualMachineInfo {
  state: number /* uint8 */;
  max_mem_kb: number /* uint64 */;
  memory_kb: number /* uint64 */;
  vcpus: number /* uint16 */;
  cpu_time_ns: number /* uint64 */;
}
/**
 * VirtualMachineWithInfo combines VM details with runtime information
 */
export interface VirtualMachineWithInfo {
  id: number /* int32 */;
  name: string;
  uuid: string;
  VirtualMachineInfo: VirtualMachineInfo;
}
/**
 * CreateVMRequest represents a request to create a new virtual machine
 */
export interface CreateVMRequest {
  name: string;
  memory: number /* int64 */;
  vcpus: number /* int32 */;
  disk_size: number /* int64 */;
  os_image: string;
  autostart: boolean;
}
/**
 * VMActionResponse represents the response from VM control operations
 */
export interface VMActionResponse {
  success: boolean;
  message: string;
}
/**
 * QEMU VM States (libvirt domain states)
 */
export const VIR_DOMAIN_NOSTATE = 0; // No state
/**
 * QEMU VM States (libvirt domain states)
 */
export const VIR_DOMAIN_RUNNING = 1; // The domain is running
/**
 * QEMU VM States (libvirt domain states)
 */
export const VIR_DOMAIN_BLOCKED = 2; // The domain is blocked on resource
/**
 * QEMU VM States (libvirt domain states)
 */
export const VIR_DOMAIN_PAUSED = 3; // The domain is paused by user
/**
 * QEMU VM States (libvirt domain states)
 */
export const VIR_DOMAIN_SHUTDOWN = 4; // The domain is being shut down
/**
 * QEMU VM States (libvirt domain states)
 */
export const VIR_DOMAIN_SHUTOFF = 5; // The domain is shut off
/**
 * QEMU VM States (libvirt domain states)
 */
export const VIR_DOMAIN_CRASHED = 6; // The domain is crashed
/**
 * QEMU VM States (libvirt domain states)
 */
export const VIR_DOMAIN_PMSUSPENDED = 7; // The domain is suspended by guest power management
export interface Login {
  username: string;
  password: string;
}
export type RBACPolicy = string;
export const RBAC_DOCKER_READ: RBACPolicy = "docker_read";
export const RBAC_DOCKER_WRITE: RBACPolicy = "docker_write";
export const RBAC_DOCKER_UPDATE: RBACPolicy = "docker_update";
export const RBAC_DOCKER_DELETE: RBACPolicy = "docker_delete";
export const RBAC_QEMU_READ: RBACPolicy = "qemu_read";
export const RBAC_QEMU_WRITE: RBACPolicy = "qemu_write";
export const RBAC_QEMU_UPDATE: RBACPolicy = "qemu_update";
export const RBAC_QEMU_DELETE: RBACPolicy = "qemu_delete";
export const RBAC_EVENT_VIEWER: RBACPolicy = "event_viewer";
export const RBAC_EVENT_MANAGER: RBACPolicy = "event_manager";
export const RBAC_USER_ADMIN: RBACPolicy = "user_admin";
export const RBAC_SETTINGS_MANAGER: RBACPolicy = "settings_manager";
export const RBAC_AUDIT_LOG_VIEWER: RBACPolicy = "audit_log_viewer";
export const RBAC_HEALTH_CHECKER: RBACPolicy = "health_checker";
export const RBAC_FIREWALL_READ: RBACPolicy = "firewall_read";
export const RBAC_FIREWALL_WRITE: RBACPolicy = "firewall_write";
export const RBAC_FIREWALL_UPDATE: RBACPolicy = "firewall_update";
export const RBAC_FIREWALL_DELETE: RBACPolicy = "firewall_delete";
/**
 * this is just to mimic the echo error structure
 * eg: {"message":"Failed to list virtual-machines"}
 * reprod: echo.NewHTTPError(http.StatusInternalServerError, "Failed to list virtual-machines").SetInternal(fmt.Errorf("database connection error"))
 */
export interface HTTPError {
  message: string;
}
export interface LogResponse {
  id: number /* int64 */;
  user_id: number /* int64 */;
  action: string;
  details?: string;
  service_group: string;
  level: string;
  created_at: string;
}
export interface LogRequestData {
  Request_id: string;
  User_id: number /* int64 */;
  Method: string;
  Path: string;
  Uri: string;
  Status: number /* int */;
  Latency: any /* time.Duration */;
  Remote_ip: string;
  User_agent: string;
  Protocol: string;
  Bytes: number /* int64 */;
  Error: any;
}
export interface GetLogsResponse {
  logs: LogResponse[];
  total: number /* int64 */;
  page: number /* int */;
  page_size: number /* int */;
  total_pages: number /* int64 */;
}
export interface LogStatsResponse {
  total: number /* int64 */;
  days: number /* int */;
  service_groups: string[];
  levels: string[];
  since: string;
}
export interface ClearOldLogsResponse {
  retention_days: number /* int */;
  before: string;
  message: string;
}
export interface MetricsPeriod {
  days: number /* int */;
  since: string;
  until: string;
}
export interface ErrorRateByService {
  service_group: string;
  error_count: number /* int64 */;
  total_count: number /* int64 */;
  error_rate: number /* float64 */;
}
export interface LogCountByHour {
  hour: string;
  log_count: number /* int64 */;
}
export interface LogLevelStats {
  level: string;
  count: number /* int64 */;
  percentage: number /* float64 */;
}
export interface ServiceStats {
  service_group: string;
  count: number /* int64 */;
  percentage: number /* float64 */;
}
export interface MetricsResponse {
  error_rate_by_service: ErrorRateByService[];
  log_count_by_hour: LogCountByHour[];
  log_level_distribution: LogLevelStats[];
  service_group_distribution: ServiceStats[];
  period: MetricsPeriod;
}
export interface ServiceMetricsResponse {
  service_group: string;
  days: number /* int */;
  since: string;
  total_logs: number /* int64 */;
  error_count: number /* int64 */;
  error_rate: number /* float64 */;
  level_distribution: LogLevelStats[];
}
export interface ServiceHealth {
  service_group: string;
  error_rate: number /* float64 */;
  error_count: number /* int64 */;
  total_count: number /* int64 */;
  status: string;
}
export interface HealthMetricsResponse {
  timestamp: string;
  period: string;
  services: ServiceHealth[];
  overall_status: string;
  alerts: string[];
}
export interface StorageDevice {
  name: string;
  size: string;
  size_bytes: number /* int64 */;
  type: string;
  mount_point: string;
  usage_percent: number /* int32 */;
}
export interface MountPoint {
  path: string;
  device: string;
  fs_type: string;
  total: number /* int64 */;
  used: number /* int64 */;
  available: number /* int64 */;
  use_percent: number /* int32 */;
}
export interface EnvVars {
  Port: string;
  AppEnv: string;
  DBPath: string;
  APP_VERSION: string;
  GoogleOAuthKey: string;
  GoogleOAuthSecret: string;
  GithubOAuthKey: string;
  GithubOAuthSecret: string;
  /**
   * OAuthCallbackURL  string `envconfig:"OAUTH_CALLBACK_URL" default:"http://localhost:9999/api/auth/oauth/callback" required:"true"`
   */
  BaseUrl: string;
  Directory: string;
  SessionSecret: string;
  FRONTEND_DASH: string;
  BaseUrlWithPort: string;
}
/**
 * PortainerTemplate represents a single template from the Portainer templates JSON
 */
export interface PortainerTemplate {
  type: number /* int */;
  title: string;
  description: string;
  categories: string[];
  platform: string;
  logo: string;
  image: string;
  name: string;
  registry: string;
  command: string;
  network: string;
  privileged: boolean;
  interactive: boolean;
  hostname: string;
  note: string;
  ports: string[];
  volumes: TemplateVolume[];
  env: TemplateEnv[];
  labels: TemplateLabel[];
  restart_policy: string;
}
/**
 * TemplateVolume represents a volume mapping in a template
 */
export interface TemplateVolume {
  container: string;
  bind: string;
  readonly: boolean;
}
/**
 * TemplateEnv represents an environment variable in a template
 */
export interface TemplateEnv {
  name: string;
  label: string;
  description: string;
  default: string;
  preset: boolean;
  select: EnvSelect[];
}
/**
 * EnvSelect represents a select option for environment variables
 */
export interface EnvSelect {
  text: string;
  value: string;
  default: boolean;
}
/**
 * TemplateLabel represents a label in a template
 */
export interface TemplateLabel {
  name: string;
  value: string;
}
/**
 * TemplatesResponse represents the response from the Portainer templates API
 */
export interface TemplatesResponse {
  version: string;
  templates: PortainerTemplate[];
}
/**
 * DeployTemplateRequest represents a request to deploy a template
 */
export interface DeployTemplateRequest {
  name: string; // Container name (optional, uses template name if empty)
  env: { [key: string]: string }; // Environment variable overrides
  ports: string[]; // Port mapping overrides
  volumes: TemplateVolume[]; // Volume mapping overrides
  network: string; // Network to attach to
  restart_policy: string; // Restart policy override
}
/**
 * TemplateListItem represents a simplified template for listing
 */
export interface TemplateListItem {
  id: number /* int */;
  title: string;
  description: string;
  categories: string[];
  platform: string;
  logo: string;
  image: string;
}
/**
 * DeployResponse represents the response after deploying a template
 */
export interface DeployResponse {
  container_id: string;
  name: string;
  message: string;
}
export const COOKIE_NAME = "token";
export const BYPASS_RBAC_HEADER = "X-Bypass-RBAC";
/**
 * FirewallRule represents a single nftables rule
 */
export interface FirewallRule {
  handle: number /* uint64 */; // nftables rule handle (unique ID)
  chain: string; // "input", "forward", or "output"
  protocol: string; // "tcp", "udp", or "" for any
  port: number /* uint16 */; // destination port (0 = any)
  source_ip: string; // source IP/CIDR or "" for any
  action: string; // "accept" or "drop"
  comment: string; // optional user comment
}
/**
 * FirewallStatus represents the current firewall state
 */
export interface FirewallStatus {
  enabled: boolean;
  rule_count: number /* int */;
  table_name: string;
}
/**
 * CreateRuleRequest is the request body for creating a new firewall rule
 */
export interface CreateRuleRequest {
  chain: string;
  protocol: string;
  port: number /* uint16 */;
  source_ip: string;
  action: string;
  comment: string;
}
/**
 * ReorderRulesRequest is the request body for reordering firewall rules
 */
export interface ReorderRulesRequest {
  chain: string;
  handles: number /* uint64 */[]; // Rule handles in desired order
}
export interface FirewallService {
  Dispatcher?: any /* utils.Dispatcher */;
  Logger?: any /* slog.Logger */;
}
export interface QemuService {
  Dispatcher?: any /* utils.Dispatcher */;
  Logger?: any /* slog.Logger */;
  LibVirt?: any /* libvirt.Libvirt */;
}
export interface StorageService {
  Dispatcher?: any /* utils.Dispatcher */;
  Logger?: any /* slog.Logger */;
}
export interface DocsService {
  Dispatcher?: any /* utils.Dispatcher */;
  Logger?: any /* slog.Logger */;
  Spec: string;
}
export interface UsersService {
  Dispatcher?: any /* utils.Dispatcher */;
  Logger?: any /* slog.Logger */;
}
export interface ClientInfo {
  id: number /* int */;
  status: string;
}
export interface DockerService {
  Dispatcher?: any /* utils.Dispatcher */;
  Logger?: any /* slog.Logger */;
  ClientManager?: any /* clientmanager.Docker */;
}
export interface AuthService {
  Dispatcher?: any /* utils.Dispatcher */;
  Logger?: any /* slog.Logger */;
  OAuthProviders: { [key: string]: any /* goth.Provider */ };
}
export interface MetricsService {
  Dispatcher?: any /* utils.Dispatcher */;
  Logger?: any /* slog.Logger */;
}
/**
 * GetMetricsRequest represents metrics query parameters
 */
export interface GetMetricsRequest {
  Days: number /* int */;
}
export interface TemplatesService {
  Dispatcher?: any /* utils.Dispatcher */;
  Logger?: any /* slog.Logger */;
}
export interface LogsService {
  Dispatcher?: any /* utils.Dispatcher */;
  Logger?: any /* slog.Logger */;
}
/**
 * GetLogsRequest represents query parameters for log filtering
 */
export interface GetLogsRequest {
  ServiceGroup: string;
  Level: string;
  Page: number /* int */;
  PageSize: number /* int */;
  Days: number /* int */; // Filter logs from last N days
}
export interface CountLogsByLevelParams {
  level: string;
  created_at: string;
}
export interface CountLogsByServiceGroupParams {
  service_group: string;
  created_at: string;
}
export interface CountLogsByServiceGroupAndLevelParams {
  service_group: string;
  level: string;
  created_at: string;
}
export interface CreateLogParams {
  user_id: number /* int64 */;
  action: string;
  details?: string;
  service_group: string;
  level: string;
}
export interface GetAverageLogCountByHourRow {
  hour: any;
  log_count: number /* int64 */;
}
export interface GetErrorRateByServiceRow {
  service_group: string;
  error_count: number /* int64 */;
  total_count: number /* int64 */;
  error_rate: number /* float64 */;
}
export interface GetLogLevelDistributionParams {
  created_at: string;
  created_at_2: string;
}
export interface GetLogLevelDistributionRow {
  level: string;
  count: number /* int64 */;
  percentage: number /* float64 */;
}
export interface GetLogsByLevelParams {
  level: string;
  created_at: string;
  limit: number /* int64 */;
  offset: number /* int64 */;
}
export interface GetLogsByServiceGroupParams {
  service_group: string;
  created_at: string;
  limit: number /* int64 */;
  offset: number /* int64 */;
}
export interface GetLogsByServiceGroupAndLevelParams {
  service_group: string;
  level: string;
  created_at: string;
  limit: number /* int64 */;
  offset: number /* int64 */;
}
export interface GetLogsPaginatedParams {
  created_at: string;
  limit: number /* int64 */;
  offset: number /* int64 */;
}
export interface GetServiceGroupDistributionParams {
  created_at: string;
  created_at_2: string;
}
export interface GetServiceGroupDistributionRow {
  service_group: string;
  count: number /* int64 */;
  percentage: number /* float64 */;
}
export interface UpsertSessionParams {
  user_id: number /* int64 */;
  session_token: string;
}
export interface Service {
  User?: any /* user.Queries */; // Assuming user.Queries is a struct generated by sqlc for user-related queries
  Session?: any /* sessions.Queries */;
  Log?: any /* logs.Queries */;
  Notification?: any /* notifications.Queries */;
}
export interface Health {
  status: string;
  message: string;
  app_version: string;
  base_url: string;
  error?: string;
  stats: HealthStats;
}
export interface HealthStats {
  open_connections: number /* int */;
  in_use: number /* int */;
  idle: number /* int */;
  wait_count: number /* int64 */;
  wait_duration: string;
  max_idle_closed: number /* int64 */;
  max_lifetime_closed: number /* int64 */;
}
export interface CreateUserParams {
  username: string;
  email: string;
  password: string;
  role: string;
}
export interface GetByEmailOrUsernameParams {
  email: string;
  username: string;
}
export interface GetUserAndSessionByTokenRow {
  user: User;
  user_session: UserSession;
}
export interface UpdateUserParams {
  username: string;
  email: string;
  role: string;
  id: number /* int64 */;
}
export interface UpdateUserPasswordParams {
  password: string;
  id: number /* int64 */;
}
export interface UpdateUserRoleParams {
  role: string;
  id: number /* int64 */;
}
export interface UpsertUserParams {
  username: string;
  email: string;
  password: string;
  role: string;
}
export interface Log {
  id: number /* int64 */;
  user_id: number /* int64 */;
  action: string;
  details?: string;
  created_at: string;
  service_group: string;
  level: string;
}
export interface Notification {
  id: number /* int64 */;
  user_id: number /* int64 */;
  message: string;
  read?: boolean;
  created_at: string;
  updated_at: string;
}
export interface User {
  id: number /* int64 */;
  username: string;
  email: string;
  password: string;
  role: string;
  created_at: string;
  updated_at: string;
}
export interface UserSession {
  id: number /* int64 */;
  user_id: number /* int64 */;
  session_token: string;
  created_at: string;
  updated_at: string;
}
