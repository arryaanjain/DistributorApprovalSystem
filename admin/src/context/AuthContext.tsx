import React, { createContext, useContext, useState, useEffect } from 'react';
import { getAuthToken, setAuthToken, clearAuthToken } from '../services/api';

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
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(!!getAuthToken());
  const [user, setUser] = useState<AuthUser | null>(() => {
    const saved = localStorage.getItem('kresconet_admin_user');
    return saved ? JSON.parse(saved) : { id: 'EMP-001', name: 'Risk Admin', role: 'credit_manager' };
  });

  useEffect(() => {
    const token = getAuthToken();
    if (!token) {
      // Set default demo token for seamless dev admin experience
      const demoToken = 'DEMO-ADMIN-JWT-TOKEN';
      setAuthToken(demoToken);
      setIsAuthenticated(true);
    }
  }, []);

  const login = (token: string, userData: AuthUser) => {
    setAuthToken(token);
    setUser(userData);
    localStorage.setItem('kresconet_admin_user', JSON.stringify(userData));
    setIsAuthenticated(true);
  };

  const logout = () => {
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
