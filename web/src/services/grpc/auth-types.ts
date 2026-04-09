// TypeScript types for gRPC Auth service
// Based on the proto definitions

export interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  createdAt: number;
  updatedAt: number;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  success: boolean;
  token: string;
  user?: User;
  message: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
  email: string;
  role?: string;
}

export interface RegisterResponse {
  success: boolean;
  message: string;
  user?: User;
}

export interface GetCurrentUserRequest {
  token: string;
}

export interface GetCurrentUserResponse {
  user?: User;
  authenticated: boolean;
}

export interface RefreshTokenRequest {
  refreshToken: string;
}

export interface RefreshTokenResponse {
  accessToken: string;
  refreshToken?: string;
  success: boolean;
  message: string;
}

// gRPC Service interface
export interface AuthServiceClient {
  login(
    request: LoginRequest,
    callback: (error: Error | null, response: LoginResponse) => void
  ): void;
  
  register(
    request: RegisterRequest,
    callback: (error: Error | null, response: RegisterResponse) => void
  ): void;
  
  getCurrentUser(
    request: GetCurrentUserRequest,
    callback: (error: Error | null, response: GetCurrentUserResponse) => void
  ): void;
  
  refreshToken(
    request: RefreshTokenRequest,
    callback: (error: Error | null, response: RefreshTokenResponse) => void
  ): void;
}
