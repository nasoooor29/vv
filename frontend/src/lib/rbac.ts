// Import auto-generated types from backend
import {
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
  type RBACPolicy,
} from "@/types/types.gen";

/**
 * Convert role string (comma-separated) to RBAC policies
 * Matches backend logic from models.go: RoleToRBACPolicies
 * 
 * Splits the role string by comma and compares each role against all available RBAC policies
 */
export function roleToRBACPolicies(roleString: string): Set<RBACPolicy> {
  const policies = new Set<RBACPolicy>();

  if (!roleString) return policies;

  // Split by comma and trim whitespace (same as backend)
  const roles = roleString.split(",").map((r) => r.trim());

  // All available RBAC policies to compare against
  const allPolicies = [
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
  ];

  // For each role, check if it matches any of the available policies
  for (const role of roles) {
    if (!role) continue;

    // Compare role string with each available policy
    for (const policy of allPolicies) {
      if (role === policy) {
        policies.add(policy);
        break; // Found a match, move to next role
      }
    }
  }

  return policies;
}

/**
 * Check if user has required permission(s)
 * Admins (USER_ADMIN) bypass all checks
 */
export function hasPermission(
  userRoles: Set<RBACPolicy>,
  requiredPolicy: RBACPolicy | RBACPolicy[]
): boolean {
  // Check if user is admin - admins have all permissions
  if (userRoles.has(RBAC_USER_ADMIN)) return true;

  const required = Array.isArray(requiredPolicy)
    ? requiredPolicy
    : [requiredPolicy];

  // Check if user has all required permissions
  return required.every((policy) => userRoles.has(policy));
}

/**
 * Check if user has ANY of the required permissions
 * Admins (USER_ADMIN) bypass all checks
 */
export function hasAnyPermission(
  userRoles: Set<RBACPolicy>,
  requiredPolicies: RBACPolicy[]
): boolean {
  // Check if user is admin
  if (userRoles.has(RBAC_USER_ADMIN)) return true;

  return requiredPolicies.some((policy) => userRoles.has(policy));
}

/**
 * Convert comma-separated role string to array of role strings
 */
export function roleStringToArray(roleString: string): string[] {
  if (!roleString) return [];
  return roleString
    .split(",")
    .map((role) => role.trim())
    .filter((role) => role.length > 0);
}

/**
 * Convert array of roles to comma-separated string
 */
export function roleArrayToString(roles: string[]): string {
  return roles.join(", ");
}
