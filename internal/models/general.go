package models

import (
	"strings"
	"time"
)

const (
	COOKIE_NAME        = "token"
	BYPASS_RBAC_HEADER = "X-Bypass-RBAC"
)

type Login struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RBACPolicy string

const (
	RBAC_DOCKER_READ   RBACPolicy = "docker_read"
	RBAC_DOCKER_WRITE  RBACPolicy = "docker_write"
	RBAC_DOCKER_UPDATE RBACPolicy = "docker_update"
	RBAC_DOCKER_DELETE RBACPolicy = "docker_delete"

	RBAC_QEMU_READ   RBACPolicy = "qemu_read"
	RBAC_QEMU_WRITE  RBACPolicy = "qemu_write"
	RBAC_QEMU_UPDATE RBACPolicy = "qemu_update"
	RBAC_QEMU_DELETE RBACPolicy = "qemu_delete"

	RBAC_EVENT_VIEWER  RBACPolicy = "event_viewer"
	RBAC_EVENT_MANAGER RBACPolicy = "event_manager"

	RBAC_USER_ADMIN RBACPolicy = "user_admin"

	RBAC_SETTINGS_MANAGER RBACPolicy = "settings_manager"
	RBAC_AUDIT_LOG_VIEWER RBACPolicy = "audit_log_viewer"
	RBAC_HEALTH_CHECKER   RBACPolicy = "health_checker"
)

var AllRBACPolicies = map[string]RBACPolicy{
	string(RBAC_DOCKER_READ):   RBAC_DOCKER_READ,
	string(RBAC_DOCKER_WRITE):  RBAC_DOCKER_WRITE,
	string(RBAC_DOCKER_UPDATE): RBAC_DOCKER_UPDATE,
	string(RBAC_DOCKER_DELETE): RBAC_DOCKER_DELETE,

	string(RBAC_QEMU_READ):   RBAC_QEMU_READ,
	string(RBAC_QEMU_WRITE):  RBAC_QEMU_WRITE,
	string(RBAC_QEMU_UPDATE): RBAC_QEMU_UPDATE,
	string(RBAC_QEMU_DELETE): RBAC_QEMU_DELETE,

	string(RBAC_EVENT_VIEWER):  RBAC_EVENT_VIEWER,
	string(RBAC_EVENT_MANAGER): RBAC_EVENT_MANAGER,
	string(RBAC_USER_ADMIN):    RBAC_USER_ADMIN,

	string(RBAC_SETTINGS_MANAGER): RBAC_SETTINGS_MANAGER,
	string(RBAC_AUDIT_LOG_VIEWER): RBAC_AUDIT_LOG_VIEWER,
	string(RBAC_HEALTH_CHECKER):   RBAC_HEALTH_CHECKER,
}

// user role to roles
func RoleToRBACPolicies(role string) map[RBACPolicy]bool {
	// split by comma
	roles := strings.Split(role, ",")
	policies := make(map[RBACPolicy]bool)

	for _, r := range roles {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		policy, ok := AllRBACPolicies[r]
		if !ok {
			continue
		}
		policies[policy] = true
	}
	return policies
}

// ============ Logs Response Structs ============

type LogResponse struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Action       string    `json:"action"`
	Details      *string   `json:"details"`
	ServiceGroup string    `json:"service_group"`
	Level        string    `json:"level"`
	CreatedAt    time.Time `json:"created_at"`
}

type GetLogsResponse struct {
	Logs       []LogResponse `json:"logs"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int64         `json:"total_pages"`
}

type LogStatsResponse struct {
	Total         int64     `json:"total"`
	Days          int       `json:"days"`
	ServiceGroups []string  `json:"service_groups"`
	Levels        []string  `json:"levels"`
	Since         time.Time `json:"since"`
}

type ClearOldLogsResponse struct {
	RetentionDays int       `json:"retention_days"`
	Before        time.Time `json:"before"`
	Message       string    `json:"message"`
}

// ============ Metrics Response Structs ============

type MetricsPeriod struct {
	Days  int       `json:"days"`
	Since time.Time `json:"since"`
	Until time.Time `json:"until"`
}

type ErrorRateByService struct {
	ServiceGroup string  `json:"service_group"`
	ErrorCount   int64   `json:"error_count"`
	TotalCount   int64   `json:"total_count"`
	ErrorRate    float64 `json:"error_rate"`
}

type LogCountByHour struct {
	Hour     string `json:"hour"`
	LogCount int64  `json:"log_count"`
}

type LogLevelStats struct {
	Level      string  `json:"level"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

type ServiceStats struct {
	ServiceGroup string  `json:"service_group"`
	Count        int64   `json:"count"`
	Percentage   float64 `json:"percentage"`
}

type MetricsResponse struct {
	ErrorRateByService       []ErrorRateByService `json:"error_rate_by_service"`
	LogCountByHour           []LogCountByHour     `json:"log_count_by_hour"`
	LogLevelDistribution     []LogLevelStats      `json:"log_level_distribution"`
	ServiceGroupDistribution []ServiceStats       `json:"service_group_distribution"`
	Period                   MetricsPeriod        `json:"period"`
}

type ServiceMetricsResponse struct {
	ServiceGroup      string          `json:"service_group"`
	Days              int             `json:"days"`
	Since             time.Time       `json:"since"`
	TotalLogs         int64           `json:"total_logs"`
	ErrorCount        int64           `json:"error_count"`
	ErrorRate         float64         `json:"error_rate"`
	LevelDistribution []LogLevelStats `json:"level_distribution"`
}

type ServiceHealth struct {
	ServiceGroup string  `json:"service_group"`
	ErrorRate    float64 `json:"error_rate"`
	ErrorCount   int64   `json:"error_count"`
	TotalCount   int64   `json:"total_count"`
	Status       string  `json:"status"`
}

type HealthMetricsResponse struct {
	Timestamp     time.Time       `json:"timestamp"`
	Period        string          `json:"period"`
	Services      []ServiceHealth `json:"services"`
	OverallStatus string          `json:"overall_status"`
	Alerts        []string        `json:"alerts"`
}
