import React, { createContext, useContext, useState, useEffect } from 'react';
import { api, getAuthToken, setAuthToken, clearAuthToken, refreshAdminSession } from '../services/api';

export type AuthStatus = 'INITIALIZING' | 'AUTHENTICATED' | 'UNAUTHENTICATED';

interface AuthUser {
  id: string;
  name: string;
  role: string;
}

interface AuthContextType {
  authStatus: AuthStatus;
  isAuthenticated: boolean;
  user: AuthUser | null;
  error: string | null;
  login: (token: string, user: AuthUser, refreshToken?: string) => void;
  logout: () => void;
  retrySessionRecovery: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [authStatus, setAuthStatus] = useState<AuthStatus>('INITIALIZING');
  const [user, setUser] = useState<AuthUser | null>(() => {
    const saved = localStorage.getItem('kresconet_admin_user');
    return saved ? JSON.parse(saved) : null;
  });
  const [error, setError] = useState<string | null>(null);

  const attemptRefresh = async () => {
    setError(null);
    const minDelay = new Promise((resolve) => setTimeout(resolve, 600));
    try {
      const session = await refreshAdminSession();
      if (session.token) {
        const savedUser = localStorage.getItem('kresconet_admin_user');
        let userData = session.user || (savedUser ? JSON.parse(savedUser) : null);

        if (!userData && session.token) {
          try {
            const payload = JSON.parse(atob(session.token.split('.')[1]));
            userData = {
              id: payload.user_id || 'EMP-ADMIN',
              name: payload.email ? payload.email.split('@')[0] : 'Admin User',
              email: payload.email || 'admin@kresconet.com',
              role: payload.role || 'super_admin',
            };
            localStorage.setItem('kresconet_admin_user', JSON.stringify(userData));
          } catch {}
        }

        if (userData) {
          setUser(userData);
        }
        setAuthStatus('AUTHENTICATED');
      } else {
        setAuthStatus('UNAUTHENTICATED');
      }
    } catch (err: any) {
      await minDelay;
      clearAuthToken();
      setUser(null);
      setAuthStatus('UNAUTHENTICATED');
      if (err?.message && !err.message.includes('expired') && !err.message.includes('missing')) {
        setError(err.message);
      }
    }
  };

  useEffect(() => {
    const checkTokenExpiry = () => {
      const token = getAuthToken();
      if (token) {
        try {
          const payload = JSON.parse(atob(token.split('.')[1]));
          const exp = payload.exp * 1000;
          if (Date.now() >= exp - 120000) {
            attemptRefresh();
          } else {
            setAuthStatus('AUTHENTICATED');
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

  const login = (token: string, userData: AuthUser, refreshToken?: string) => {
    setAuthToken(token, refreshToken);
    setUser(userData);
    localStorage.setItem('kresconet_admin_user', JSON.stringify(userData));
    setAuthStatus('AUTHENTICATED');
  };

  const logout = () => {
    api.logout().catch(() => {});
    clearAuthToken();
    setUser(null);
    localStorage.removeItem('kresconet_admin_user');
    setAuthStatus('UNAUTHENTICATED');
  };

  const retrySessionRecovery = async () => {
    setAuthStatus('INITIALIZING');
    await attemptRefresh();
  };

  return (
    <AuthContext.Provider
      value={{
        authStatus,
        isAuthenticated: authStatus === 'AUTHENTICATED',
        user,
        error,
        login,
        logout,
        retrySessionRecovery,
      }}
    >
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
