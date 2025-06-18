import { api } from './client';
import type { ApiResponse } from './client';

// Types for authentication
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
  user: User;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface User {
  id: number;
  email: string;
  name: string;
  nickname?: string;
  avatar?: string;
  role: 'admin' | 'user';
  is_disabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface UpdateProfileRequest {
  name: string;
  nickname?: string;
  avatar?: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
  confirm_password: string;
}

// Authentication API endpoints
export const authApi = {
  // Login with email and password
  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await api.post<LoginResponse>('/api/v1/auth/login', data);
    return response.data.data!;
  },

  // Refresh access token
  async refreshToken(data: RefreshTokenRequest): Promise<LoginResponse> {
    const response = await api.post<LoginResponse>('/api/v1/auth/refresh', data);
    return response.data.data!;
  },

  // Logout current user
  async logout(): Promise<void> {
    await api.post('/api/v1/auth/logout');
  },

  // Get current user profile
  async getProfile(): Promise<User> {
    const response = await api.get<User>('/api/v1/auth/profile');
    return response.data.data!;
  },

  // Update user profile
  async updateProfile(data: UpdateProfileRequest): Promise<User> {
    const response = await api.put<User>('/api/v1/auth/profile', data);
    return response.data.data!;
  },

  // Change password
  async changePassword(data: ChangePasswordRequest): Promise<void> {
    await api.post('/api/v1/auth/change-password', data);
  },

  // Validate current token
  async validateToken(): Promise<{ user: User; valid: boolean }> {
    const response = await api.post<{ user: User; valid: boolean }>('/api/v1/auth/validate');
    return response.data.data!;
  },
};

export default authApi;
