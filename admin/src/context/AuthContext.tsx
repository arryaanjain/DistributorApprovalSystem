import React, { createContext, useContext, useState, useEffect } from 'react';
import { api, getAuthToken, setAuthToken, clearAuthToken, refreshAdminSession } from '../services/api';

interface AuthUser {
  id: string;
  name: string;
  role: string;
}

interface AuthContextType {
  isAuthenticated: boolean;
  user: AuthUser | null;
  login: (token: string, user: AuthUser, refreshToken?: string) => void;
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
    const attemptRefresh = async () => {
      try {
        const newAccess = await refreshAdminSession();
        if (newAccess) {
          setIsAuthenticated(true);
          const savedUser = localStorage.getItem('kresconet_admin_user');
          if (savedUser) {
            setUser(JSON.parse(savedUser));
          }
        }
      } catch {
        clearAuthToken();
        setUser(null);
        setIsAuthenticated(false);
      }
    };

    const checkTokenExpiry = () => {
      const token = getAuthToken();
      if (token) {
        try {
          const payload = JSON.parse(atob(token.split('.')[1]));
          const exp = payload.exp * 1000;
          // Silently refresh if token is expired or expiring in less than 2 minutes
          if (Date.now() >= exp - 120000) {
            attemptRefresh();
          } else {
            setIsAuthenticated(true);
          }
        } catch {
          attemptRefresh();
        }
      } else {
        attemptRefresh();
      }
    };

    checkTokenExpiry();
    const interval = setInterval(checkTokenExpiry, 60000);
    return () => clearInterval(interval);
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
