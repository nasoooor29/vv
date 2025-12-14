import { useSession } from "@/stores/user";
import {
  hasPermission,
  hasAnyPermission,
  roleToRBACPolicies,
} from "@/lib/rbac";
import type { RBACPolicy } from "@/types/types.gen";
import type { ReactNode } from "react";

interface ProtectedContentProps {
  children: ReactNode;
  /** Required permissions - user must have ALL of these */
  requiredPermissions?: RBACPolicy | RBACPolicy[];
  /** Any of these permissions - user needs at least ONE */
  anyPermission?: RBACPolicy[];
  /** Fallback content to show when user lacks permissions */
  fallback?: ReactNode;
  /** Show nothing if permission denied (overrides fallback) */
  hideIfUnauthorized?: boolean;
}

/**
 * ProtectedContent Component
 *
 * Conditionally renders children based on user permissions.
 * Matches backend RBAC logic from @internal/server/routes.go
 *
 * @example
 * // User must have docker_read permission
 * <ProtectedContent requiredPermissions="docker_read">
 *   <DockerPanel />
 * </ProtectedContent>
 *
 * @example
 * // User must have both docker_read AND docker_write
 * <ProtectedContent requiredPermissions={["docker_read", "docker_write"]}>
 *   <DockerEditPanel />
 * </ProtectedContent>
 *
 * @example
 * // User needs either event_viewer OR event_manager
 * <ProtectedContent anyPermission={["event_viewer", "event_manager"]}>
 *   <EventPanel />
 * </ProtectedContent>
 *
 * @example
 * // With fallback content
 * <ProtectedContent
 *   requiredPermissions="settings_manager"
 *   fallback={<div>You don't have permission to manage settings</div>}
 * >
 *   <SettingsPanel />
 * </ProtectedContent>
 */
export function ProtectedContent({
  children,
  requiredPermissions,
  anyPermission,
  fallback,
  hideIfUnauthorized = false,
}: ProtectedContentProps) {
  const session = useSession((s) => s.session);

  console.log("ProtectedContent render - session:", session);
  console.log("ProtectedContent render - required:", requiredPermissions);

  // If no session exists, deny access
  if (!session || !session.user.role) {
    console.log("ProtectedContent - No session or role, showing fallback");
    if (hideIfUnauthorized) return null;
    return fallback ? <>{fallback}</> : null;
  }

  // Convert user roles to permissions set
  const userPermissions = roleToRBACPolicies(session.user.role);
  console.log("ProtectedContent - User role:", session.user.role);
  console.log("ProtectedContent - User permissions:", userPermissions);

  // Check permissions
  let hasAccess = true;

  // If specific permissions are required, check them
  if (requiredPermissions) {
    hasAccess = hasPermission(userPermissions, requiredPermissions);
    console.log("ProtectedContent - Permission check result:", hasAccess);
  }

  // If any permissions are specified, check if user has at least one
  if (anyPermission && anyPermission.length > 0) {
    hasAccess = hasAnyPermission(userPermissions, anyPermission);
  }

  // Render based on permission check
  if (!hasAccess) {
    console.log("ProtectedContent - Access denied");
    if (hideIfUnauthorized) return null;
    return fallback ? <>{fallback}</> : null;
  }

  console.log("ProtectedContent - Access granted, rendering children");
  return <>{children}</>;
}

/**
 * Hook to check if current user has required permissions
 *
 * @example
 * const { hasPermission } = usePermission();
 * if (hasPermission("docker_write")) {
 *   // Show edit button
 * }
 */
export function usePermission() {
  const session = useSession((s) => s.session);
  // console.log("usePermission session:", session);
  // console.log("usePermission user role:", session?.user?.role);

  const checkPermission = (required: RBACPolicy | RBACPolicy[]): boolean => {
    if (!session || !session.user.role) return false;
    const userPermissions = roleToRBACPolicies(session.user.role);
    return hasPermission(userPermissions, required);
  };

  const checkAnyPermission = (policies: RBACPolicy[]): boolean => {
    if (!session || !session.user.role) return false;
    const userPermissions = roleToRBACPolicies(session.user.role);
    return hasAnyPermission(userPermissions, policies);
  };

  const getUserPermissions = () => {
    if (!session || !session.user.role) return new Set<RBACPolicy>();
    return roleToRBACPolicies(session.user.role);
  };

  return {
    hasPermission: checkPermission,
    hasAnyPermission: checkAnyPermission,
    getUserPermissions,
    session,
  };
}
