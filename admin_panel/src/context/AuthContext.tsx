import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import axios from 'axios';

interface User {
  id: string;
  full_name: string;
  email: string | null;
  role: string;
}

interface AuthContextType {
  user: User | null;
  accessToken: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
  isLoading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    const userData = localStorage.getItem('user');
    
    if (token && userData) {
      setAccessToken(token);
      setUser(JSON.parse(userData));
    }
    setIsLoading(false);
  }, []);

  const login = async (email: string, password: string) => {
    try {
      const response = await axios.post(`${API_BASE_URL}/auth/login`, {
        email,
        password,
      });

      const { user: userData, access_token } = response.data;

      if (userData.role !== 'admin') {
        throw new Error('Access denied: Admin role required');
      }

      setAccessToken(access_token);
      setUser(userData);

      localStorage.setItem('accessToken', access_token);
      localStorage.setItem('user', JSON.stringify(userData));
    } catch (error: any) {
      console.error('Login failed:', error);
      throw new Error(error.response?.data?.error || 'Login failed');
    }
  };

  const logout = () => {
    setAccessToken(null);
    setUser(null);
    localStorage.removeItem('accessToken');
    localStorage.removeItem('user');
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        accessToken,
        login,
        logout,
        isAuthenticated: !!accessToken && !!user,
        isLoading,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}

