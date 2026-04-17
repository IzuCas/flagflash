import { useMemo } from 'react';
import { useAuth } from '../contexts/AuthContext';

export type UserRole = 'owner' | 'admin' | 'member' | 'viewer';

// Role hierarchy levels
const ROLE_LEVELS: Record<UserRole, number> = {
  owner: 100,
  admin: 75,
  member: 50,
  viewer: 25,
};

export interface Permissions {
  // User Management
  canCreateUser: boolean;
  canUpdateUser: boolean;
  canDeleteUser: boolean;
  canInviteUser: boolean;
  canChangeUserRole: boolean;
  
  // Applications
  canCreateApplication: boolean;
  canUpdateApplication: boolean;
  canDeleteApplication: boolean;
  
  // Environments
  canCreateEnvironment: boolean;
  canUpdateEnvironment: boolean;
  canDeleteEnvironment: boolean;
  canCopyEnvironment: boolean;
  
  // Feature Flags
  canCreateFeatureFlag: boolean;
  canUpdateFeatureFlag: boolean;
  canDeleteFeatureFlag: boolean;
  canToggleFeatureFlag: boolean;
  canCopyFeatureFlags: boolean;
  
  // Targeting Rules
  canCreateTargetingRule: boolean;
  canUpdateTargetingRule: boolean;
  canDeleteTargetingRule: boolean;
  canReorderTargetingRules: boolean;
  
  // API Keys
  canViewAPIKeys: boolean;
  canCreateAPIKey: boolean;
  canRevokeAPIKey: boolean;
  canDeleteAPIKey: boolean;
  
  // Audit Logs
  canViewAuditLogs: boolean;
  
  // Tenant Settings
  canUpdateTenant: boolean;
  canDeleteTenant: boolean;
  
  // Metrics
  canViewMetrics: boolean;
}

/**
 * Permission matrix based on user role
 * 
 * | Resource         | Viewer | Member | Admin | Owner |
 * |------------------|--------|--------|-------|-------|
 * | View Data        |   ✅   |   ✅   |  ✅   |  ✅   |
 * | Create/Update    |   ❌   |   ✅   |  ✅   |  ✅   |
 * | Delete           |   ❌   |   ❌   |  ✅   |  ✅   |
 * | User Management  |   ❌   |   ❌   |  ✅   |  ✅   |
 * | API Keys         |   ❌   |   ❌   |  ✅   |  ✅   |
 * | Audit Logs       |   ❌   |   ❌   |  ✅   |  ✅   |
 * | Tenant Settings  |   ❌   |   ❌   |  ❌   |  ✅   |
 */

export function usePermissions(): Permissions & { 
  role: UserRole | null;
  roleLevel: number;
  isAtLeast: (minRole: UserRole) => boolean;
  canManageRole: (targetRole: UserRole) => boolean;
} {
  const { selectedTenant } = useAuth();
  const role = (selectedTenant?.role as UserRole) || null;
  const roleLevel = role ? ROLE_LEVELS[role] : 0;
  
  const permissions = useMemo(() => {
    const isOwner = role === 'owner';
    const isAdminOrOwner = role === 'admin' || role === 'owner';
    const isMemberOrHigher = role === 'member' || role === 'admin' || role === 'owner';
    const isViewer = role === 'viewer';
    
    return {
      // User Management - Admin or Owner only
      canCreateUser: isAdminOrOwner,
      canUpdateUser: isAdminOrOwner,
      canDeleteUser: isAdminOrOwner,
      canInviteUser: isAdminOrOwner,
      canChangeUserRole: isAdminOrOwner,
      
      // Applications
      canCreateApplication: isMemberOrHigher,
      canUpdateApplication: isMemberOrHigher,
      canDeleteApplication: isAdminOrOwner,
      
      // Environments
      canCreateEnvironment: isMemberOrHigher,
      canUpdateEnvironment: isMemberOrHigher,
      canDeleteEnvironment: isAdminOrOwner,
      canCopyEnvironment: isMemberOrHigher,
      
      // Feature Flags
      canCreateFeatureFlag: isMemberOrHigher,
      canUpdateFeatureFlag: isMemberOrHigher,
      canDeleteFeatureFlag: isAdminOrOwner,
      canToggleFeatureFlag: isMemberOrHigher,
      canCopyFeatureFlags: isMemberOrHigher,
      
      // Targeting Rules
      canCreateTargetingRule: isMemberOrHigher,
      canUpdateTargetingRule: isMemberOrHigher,
      canDeleteTargetingRule: isAdminOrOwner,
      canReorderTargetingRules: isMemberOrHigher,
      
      // API Keys - Admin or Owner only
      canViewAPIKeys: !isViewer, // Everyone except viewer
      canCreateAPIKey: isAdminOrOwner,
      canRevokeAPIKey: isAdminOrOwner,
      canDeleteAPIKey: isAdminOrOwner,
      
      // Audit Logs - Admin or Owner only
      canViewAuditLogs: isAdminOrOwner,
      
      // Tenant Settings - Owner only
      canUpdateTenant: isOwner,
      canDeleteTenant: isOwner,
      
      // Metrics - Everyone can view
      canViewMetrics: true,
    };
  }, [role]);
  
  // Helper to check if user has at least a certain role level
  const isAtLeast = (minRole: UserRole): boolean => {
    return roleLevel >= ROLE_LEVELS[minRole];
  };
  
  // Helper to check if user can manage another user's role
  const canManageRole = (targetRole: UserRole): boolean => {
    if (targetRole === 'owner') return false; // Nobody can manage owners
    if (role === 'owner') return true;
    if (role === 'admin') return targetRole === 'member' || targetRole === 'viewer';
    return false;
  };
  
  return {
    ...permissions,
    role,
    roleLevel,
    isAtLeast,
    canManageRole,
  };
}

/**
 * Role badge colors for UI
 */
export const ROLE_BADGES: Record<UserRole, { label: string; color: string; bgColor: string }> = {
  owner: { label: 'Owner', color: 'text-purple-400', bgColor: 'bg-purple-500/20' },
  admin: { label: 'Admin', color: 'text-blue-400', bgColor: 'bg-blue-500/20' },
  member: { label: 'Member', color: 'text-green-400', bgColor: 'bg-green-500/20' },
  viewer: { label: 'Viewer', color: 'text-gray-400', bgColor: 'bg-gray-500/20' },
};

/**
 * Get available roles that a user can assign based on their own role
 */
export function getAssignableRoles(currentRole: UserRole | null): UserRole[] {
  if (currentRole === 'owner') return ['admin', 'member', 'viewer'];
  if (currentRole === 'admin') return ['member', 'viewer'];
  return [];
}
