import axios from 'axios';
import type { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from 'axios';
import toast from 'react-hot-toast';

// API configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Create axios instance
export const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Types for API responses
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data?: T;
  timestamp: string;
}

export interface ApiErrorResponse {
  code: number;
  message: string;
  error?: string;
  details?: Record<string, any>;
  timestamp: string;
}

// Token management
class TokenManager {
  private static readonly ACCESS_TOKEN_KEY = 'access_token';
  private static readonly REFRESH_TOKEN_KEY = 'refresh_token';

  static getAccessToken(): string | null {
    return localStorage.getItem(this.ACCESS_TOKEN_KEY);
  }

  static setAccessToken(token: string): void {
    localStorage.setItem(this.ACCESS_TOKEN_KEY, token);
  }

  static getRefreshToken(): string | null {
    return localStorage.getItem(this.REFRESH_TOKEN_KEY);
  }

  static setRefreshToken(token: string): void {
    localStorage.setItem(this.REFRESH_TOKEN_KEY, token);
  }

  static clearTokens(): void {
    localStorage.removeItem(this.ACCESS_TOKEN_KEY);
    localStorage.removeItem(this.REFRESH_TOKEN_KEY);
  }

  static isAuthenticated(): boolean {
    return !!this.getAccessToken();
  }
}

// Request interceptor for adding auth token
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = TokenManager.getAccessToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for handling errors and token refresh
apiClient.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    // Handle 401 errors (unauthorized)
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = TokenManager.getRefreshToken();
        if (refreshToken) {
          // Try to refresh token
          const refreshResponse = await axios.post(`${API_BASE_URL}/api/v1/auth/refresh`, {
            refresh_token: refreshToken,
          });

          const { access_token, refresh_token: newRefreshToken } = refreshResponse.data.data;

          TokenManager.setAccessToken(access_token);
          if (newRefreshToken) {
            TokenManager.setRefreshToken(newRefreshToken);
          }

          // Retry original request
          originalRequest.headers.Authorization = `Bearer ${access_token}`;
          return apiClient(originalRequest);
        }
      } catch (refreshError) {
        // Refresh failed, redirect to login
        TokenManager.clearTokens();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    // Handle other errors
    const errorMessage = error.response?.data?.message || error.message || 'An unexpected error occurred';

    // Don't show toast for 401 errors (handled above)
    if (error.response?.status !== 401) {
      toast.error(errorMessage);
    }

    return Promise.reject(error);
  }
);

// API utility functions
export const api = {
  get: <T>(url: string, config?: InternalAxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> => {
    return apiClient.get(url, config);
  },

  post: <T>(url: string, data?: any, config?: InternalAxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> => {
    return apiClient.post(url, data, config);
  },

  put: <T>(url: string, data?: any, config?: InternalAxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> => {
    return apiClient.put(url, data, config);
  },

  delete: <T>(url: string, config?: InternalAxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> => {
    return apiClient.delete(url, config);
  },

  patch: <T>(url: string, data?: any, config?: InternalAxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> => {
    return apiClient.patch(url, data, config);
  },
};

// Export token manager for use in auth context
export { TokenManager };

// Error handling utility
export class ApiError extends Error {
  constructor(
    public status: number,
    public message: string,
    public details?: Record<string, any>
  ) {
    super(message);
    this.name = 'ApiError';
  }

  static fromAxiosError(error: any): ApiError {
    if (error.response) {
      return new ApiError(
        error.response.status,
        error.response.data?.message || error.message,
        error.response.data?.details
      );
    }
    return new ApiError(0, error.message || 'Network error');
  }
}

export default api;
