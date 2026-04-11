import React, { createContext, useContext, useState } from 'react';

export interface TenantWithRole {
  id: string;
  name: string;
  slug: string;
  role: string;
  created_at?: string;
  updated_at?: string;
}

export interface SelectedTenant {
  id: string;
  name: string;
  slug: string;
  role?: string;
}

export interface AuthUser {
  id: string;
  email: string;
  name: string;
}

interface AuthContextType {
  token: string | null;
  user: AuthUser | null;
  tenants: TenantWithRole[];
  login: (token: string, user: AuthUser, tenants?: TenantWithRole[]) => void;
  logout: () => void;
  isAuthenticated: boolean;
  selectedTenant: SelectedTenant | null;
  selectTenant: (tenant: SelectedTenant) => void;
  clearTenant: () => void;
  setTenants: (tenants: TenantWithRole[]) => void;
  updateUser: (user: Partial<AuthUser>) => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('flagflash_token'));
  const [user, setUser] = useState<AuthUser | null>(() => {
    const stored = localStorage.getItem('auth_user');
    if (stored) {
      return JSON.parse(stored);
    }
    // Migration: check for old format
    const oldUsername = localStorage.getItem('auth_username');
    if (oldUsername) {
      const migratedUser = { id: '', email: oldUsername, name: oldUsername };
      localStorage.setItem('auth_user', JSON.stringify(migratedUser));
      localStorage.removeItem('auth_username');
      return migratedUser;
    }
    return null;
  });
  const [tenants, setTenantsState] = useState<TenantWithRole[]>(() => {
    const stored = localStorage.getItem('auth_tenants');
    return stored ? JSON.parse(stored) : [];
  });
  const [selectedTenant, setSelectedTenant] = useState<SelectedTenant | null>(() => {
    const stored = localStorage.getItem('selected_tenant');
    return stored ? JSON.parse(stored) : null;
  });

  const login = (newToken: string, newUser: AuthUser, newTenants?: TenantWithRole[]) => {
    localStorage.setItem('flagflash_token', newToken);
    localStorage.setItem('auth_user', JSON.stringify(newUser));
    if (newTenants) {
      localStorage.setItem('auth_tenants', JSON.stringify(newTenants));
      setTenantsState(newTenants);
    }
    setToken(newToken);
    setUser(newUser);
  };

  const logout = () => {
    localStorage.removeItem('flagflash_token');
    localStorage.removeItem('auth_user');
    localStorage.removeItem('auth_tenants');
    localStorage.removeItem('selected_tenant');
    setToken(null);
    setUser(null);
    setTenantsState([]);
    setSelectedTenant(null);
  };

  const selectTenant = (tenant: SelectedTenant) => {
    localStorage.setItem('selected_tenant', JSON.stringify(tenant));
    setSelectedTenant(tenant);
  };

  const clearTenant = () => {
    localStorage.removeItem('selected_tenant');
    setSelectedTenant(null);
  };

  const setTenants = (newTenants: TenantWithRole[]) => {
    localStorage.setItem('auth_tenants', JSON.stringify(newTenants));
    setTenantsState(newTenants);
  };

  const updateUser = (updates: Partial<AuthUser>) => {
    if (user) {
      const updatedUser = { ...user, ...updates };
      localStorage.setItem('auth_user', JSON.stringify(updatedUser));
      setUser(updatedUser);
    }
  };

  return (
    <AuthContext.Provider value={{ token, user, tenants, login, logout, isAuthenticated: !!token, selectedTenant, selectTenant, clearTenant, setTenants, updateUser }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
