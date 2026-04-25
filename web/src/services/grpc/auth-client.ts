import type {
  LoginResponse,
  RegisterResponse,
  GetCurrentUserResponse,
  RefreshTokenResponse
} from './auth-types';

// gRPC client configuration
const GRPC_HOST = import.meta.env.VITE_GRPC_HOST || 'localhost:9090';

class GrpcClient {
  private host: string;
  
  constructor(host: string = GRPC_HOST) {
    this.host = host;
  }
  
  // Convert TypeScript object to protobuf message
  private createLoginRequest(data: { username: string; password: string }) {
    return { ...data };
  }
  
  private createRegisterRequest(data: { username: string; password: string; email: string; role?: string }) {
    return { ...data };
  }
  
  private createGetCurrentUserRequest(token: string) {
    return { token };
  }
  
  private createRefreshTokenRequest(refreshToken: string) {
    return { refreshToken };
  }
  
  // Generic fetch wrapper for gRPC
  private async grpcFetch<T>(
    service: string,
    method: string,
    request: any,
    token?: string
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    // Use REST proxy endpoint (which will be handled by gRPC gateway)
    const url = `http://${this.host}/api/v1/${service}/${method}`;
    
    const response = await fetch(url, {
      method: 'POST',
      headers,
      body: JSON.stringify(request),
    });
    
    if (!response.ok) {
      throw new Error(`gRPC call failed: ${response.statusText}`);
    }
    
    return response.json();
  }
  
  // Login method
  async login(username: string, password: string): Promise<LoginResponse> {
    try {
      return await this.grpcFetch<LoginResponse>(
        'auth',
        'login',
        this.createLoginRequest({ username, password })
      );
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    }
  }
  
  // Register method
  async register(username: string, password: string, email: string, role?: string): Promise<RegisterResponse> {
    try {
      return await this.grpcFetch<RegisterResponse>(
        'auth',
        'register',
        this.createRegisterRequest({ username, password, email, role })
      );
    } catch (error) {
      console.error('Registration failed:', error);
      throw error;
    }
  }
  
  // GetCurrentUser method
  async getCurrentUser(token: string): Promise<GetCurrentUserResponse> {
    try {
      return await this.grpcFetch<GetCurrentUserResponse>(
        'auth',
        'getCurrentUser',
        this.createGetCurrentUserRequest(token),
        token
      );
    } catch (error) {
      console.error('GetCurrentUser failed:', error);
      throw error;
    }
  }
  
  // RefreshToken method
  async refreshToken(refreshToken: string): Promise<RefreshTokenResponse> {
    try {
      return await this.grpcFetch<RefreshTokenResponse>(
        'auth',
        'refreshToken',
        this.createRefreshTokenRequest(refreshToken)
      );
    } catch (error) {
      console.error('RefreshToken failed:', error);
      throw error;
    }
  }
}

// Export singleton instance
export const grpcAuthClient = new GrpcClient();

// Export class for testing
export { GrpcClient };
