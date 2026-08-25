import React, { createContext, useContext, useState, useEffect } from 'react';
import { api, getAuthToken, setAuthToken, clearAuthToken } from '../services/api';

interface AuthUser {
  id: string;
  name: string;
  role: string;
}

interface AuthContextType {
  isAuthenticated: boolean;
  user: AuthUser | null;
  login: (token: string, user: AuthUser) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(() => !!getAuthToken());
  const [user, setUser] = useState<AuthUser | null>(() => {
    const saved = localStorage.getItem('kresconet_admin_user');
    return saved ? JSON.parse(saved) : null;
  });

  useEffect(() => {
    const token = getAuthToken();
    if (token) {
      setIsAuthenticated(true);
    } else {
      const apiBase = import.meta.env.VITE_API_BASE || 'http://localhost:8080/api/v1';
      fetch(`${apiBase}/auth/employee/refresh`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      })
        .then((res) => res.json())
        .then((data) => {
          const newAccess = data.access_token || data.token;
          if (newAccess) {
            setAuthToken(newAccess);
            if (data.user) {
              setUser(data.user);
              localStorage.setItem('kresconet_admin_user', JSON.stringify(data.user));
            }
            setIsAuthenticated(true);
          } else {
            setIsAuthenticated(false);
          }
        })
        .catch(() => setIsAuthenticated(false));
    }
  }, []);

  const login = (token: string, userData: AuthUser) => {
    setAuthToken(token);
    setUser(userData);
    localStorage.setItem('kresconet_admin_user', JSON.stringify(userData));
    setIsAuthenticated(true);
  };

  const logout = () => {
    api.logout().catch(() => {});
    clearAuthToken();
    setUser(null);
    localStorage.removeItem('kresconet_admin_user');
    setIsAuthenticated(false);
  };

  return (
    <AuthContext.Provider value={{ isAuthenticated, user, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
