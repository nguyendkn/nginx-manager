import { createContext, useContext, useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import { authApi } from '../services/api/auth';
import type { LoginRequest, User } from '../services/api/auth';
import { TokenManager } from '../services/api/client';
import toast from 'react-hot-toast';

// Authentication context types
interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
  updateProfile: (data: { name: string; nickname?: string; avatar?: string }) => Promise<void>;
}

// Create auth context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Auth provider props
interface AuthProviderProps {
  children: ReactNode;
}

// Auth provider component
export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Check if user is authenticated based on token
  const isAuthenticated = TokenManager.isAuthenticated() && user !== null;

  // Initialize authentication state
  useEffect(() => {
    const initializeAuth = async () => {
      if (TokenManager.isAuthenticated()) {
        try {
          await refreshUser();
        } catch (error) {
          console.error('Failed to initialize auth:', error);
          TokenManager.clearTokens();
        }
      }
      setIsLoading(false);
    };

    initializeAuth();
  }, []);

  // Login function
  const login = async (credentials: LoginRequest): Promise<void> => {
    try {
      setIsLoading(true);
      const response = await authApi.login(credentials);

      // Store tokens
      TokenManager.setAccessToken(response.access_token);
      TokenManager.setRefreshToken(response.refresh_token);

      // Set user
      setUser(response.user);

      toast.success('Login successful!');
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  // Logout function
  const logout = async (): Promise<void> => {
    try {
      setIsLoading(true);

      // Call logout API if authenticated
      if (TokenManager.isAuthenticated()) {
        try {
          await authApi.logout();
        } catch (error) {
          // Ignore logout API errors, still clear local state
          console.warn('Logout API call failed:', error);
        }
      }

      // Clear tokens and user state
      TokenManager.clearTokens();
      setUser(null);

      toast.success('Logged out successfully');
    } catch (error) {
      console.error('Logout failed:', error);
    } finally {
      setIsLoading(false);
    }
  };

  // Refresh user profile
  const refreshUser = async (): Promise<void> => {
    try {
      const userProfile = await authApi.getProfile();
      setUser(userProfile);
    } catch (error) {
      console.error('Failed to refresh user:', error);
      // Clear auth if profile fetch fails
      TokenManager.clearTokens();
      setUser(null);
      throw error;
    }
  };

  // Update user profile
  const updateProfile = async (data: { name: string; nickname?: string; avatar?: string }): Promise<void> => {
    try {
      const updatedUser = await authApi.updateProfile(data);
      setUser(updatedUser);
      toast.success('Profile updated successfully');
    } catch (error) {
      console.error('Profile update failed:', error);
      throw error;
    }
  };

  const value: AuthContextType = {
    user,
    isAuthenticated,
    isLoading,
    login,
    logout,
    refreshUser,
    updateProfile,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

// Custom hook to use auth context
export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

// Hook to check if user has admin role
export function useIsAdmin(): boolean {
  const { user } = useAuth();
  return user?.role === 'admin';
}

// Hook to require authentication
export function useRequireAuth(): AuthContextType {
  const auth = useAuth();

  useEffect(() => {
    if (!auth.isLoading && !auth.isAuthenticated) {
      // Redirect to login if not authenticated
      window.location.href = '/login';
    }
  }, [auth.isAuthenticated, auth.isLoading]);

  return auth;
}

export default AuthContext;
