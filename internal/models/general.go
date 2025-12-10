package models

import (
	"strings"
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
