import type { Role } from '../types';

const ADMIN_ROLES: Role[] = ['superadmin', 'facility_admin'];
const WRITE_ROLES: Role[] = ['superadmin', 'facility_admin', 'facility_user'];

export function canCreateRegistration(role: Role): boolean {
  return WRITE_ROLES.includes(role);
}

export function canEditRegistration(role: Role): boolean {
  return WRITE_ROLES.includes(role);
}

export function canDeleteRegistration(role: Role): boolean {
  return ADMIN_ROLES.includes(role);
}

export function canManageUsers(role: Role): boolean {
  return ADMIN_ROLES.includes(role);
}

export function canViewAuditLogs(role: Role): boolean {
  return ADMIN_ROLES.includes(role);
}

export function isAdmin(role: Role): boolean {
  return ADMIN_ROLES.includes(role);
}

export function isSuperAdmin(role: Role): boolean {
  return role === 'superadmin';
}
