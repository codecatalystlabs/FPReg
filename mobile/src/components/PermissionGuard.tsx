import React from 'react';
import { useAuthStore } from '../store/authStore';
import type { Role } from '../types';

interface Props {
  allowedRoles: Role[];
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

export function PermissionGuard({ allowedRoles, children, fallback = null }: Props) {
  const user = useAuthStore((s) => s.user);
  if (!user || !allowedRoles.includes(user.role)) {
    return <>{fallback}</>;
  }
  return <>{children}</>;
}
