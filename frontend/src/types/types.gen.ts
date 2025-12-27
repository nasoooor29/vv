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
  SessionSecret: string;
  FRONTEND_DASH: string;
  BaseUrlWithPort: string;
}
export const COOKIE_NAME = "token";
export const BYPASS_RBAC_HEADER = "X-Bypass-RBAC";
export interface StorageService {
  Dispatcher?: any /* utils.Dispatcher */;
  Logger?: any /* slog.Logger */;
}
export interface UsersService {
  Dispatcher?: any /* utils.Dispatcher */;
  Logger?: any /* slog.Logger */;
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
export interface VirtualMachine {
  id: number /* int32 */;
  name: string;
  uuid: string;
}
export interface VirtualMachineInfo {
  state: number /* uint8 */;
  max_mem_kb: number /* uint64 */;
  memory_kb: number /* uint64 */;
  vcpus: number /* uint16 */;
  cpu_time_ns: number /* uint64 */;
}
export interface VirtualMachineWithInfo {
  id: number /* int32 */;
  name: string;
  uuid: string;
  state: number /* uint8 */;
  max_mem_kb: number /* uint64 */;
  memory_kb: number /* uint64 */;
  vcpus: number /* uint16 */;
  cpu_time_ns: number /* uint64 */;
}
export interface CreateVMRequest {
  name: string;
  memory: number /* int64 */;
  vcpus: number /* int32 */;
  disk_size: number /* int64 */;
  os_image?: string;
  autostart?: boolean;
}
export interface VMActionResponse {
  success: boolean;
  message: string;
}
