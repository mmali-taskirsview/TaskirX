// HTTP Request Manager for React Native

let Platform: any;
try {
  Platform = require('react-native').Platform;
} catch {
  Platform = { OS: 'web', Version: '1.0' };
}

import type { ClientConfig, TaskirXError, ApiResponse } from './types';
import { TaskirXErrorType } from './types';

export class RequestManager {
  private config: ClientConfig;
  private authToken: string | null = null;
  private tokenLock = Promise.resolve();

  constructor(config: ClientConfig) {
    this.config = {
      timeout: 30000,
      retryAttempts: 3,
      ...config,
    };
  }

  setAuthToken(token: string): void {
    this.authToken = token;
  }

  clearAuthToken(): void {
    this.authToken = null;
  }

  private createError(
    type: TaskirXErrorType,
    message: string,
    statusCode?: number,
    originalError?: Error
  ): TaskirXError {
    return {
      type,
      message,
      statusCode,
      originalError,
    };
  }

  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(`${this.config.apiUrl}${endpoint}`, 'GET');
  }

  async post<T>(endpoint: string, body?: any): Promise<T> {
    return this.request<T>(`${this.config.apiUrl}${endpoint}`, 'POST', body);
  }

  async put<T>(endpoint: string, body?: any): Promise<T> {
    return this.request<T>(`${this.config.apiUrl}${endpoint}`, 'PUT', body);
  }

  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(`${this.config.apiUrl}${endpoint}`, 'DELETE');
  }

  private async request<T>(
    url: string,
    method: string,
    body?: any
  ): Promise<T> {
    const delays = [100, 300, 900]; // milliseconds
    let lastError: TaskirXError | null = null;

    for (let attempt = 0; attempt < this.config.retryAttempts!; attempt++) {
      try {
        const headers = this.buildHeaders();
        const options: RequestInit & { timeout?: number } = {
        method,
        headers,
        timeout: this.config.timeout,
      };

        if (body && (method === 'POST' || method === 'PUT')) {
          options.body = JSON.stringify(body);
        }

        if (this.config.debug) {
          console.log(`[TaskirX] 🔵 ${method} ${url}`);
        }

        const response = await fetch(url, options);

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({}));
          const errorMessage = errorData.message || `HTTP ${response.status}`;
          throw this.createError(
            TaskirXErrorType.HTTP_ERROR,
            errorMessage,
            response.status
          );
        }

        const data = await response.json();

        if (!data.success && data.error) {
          throw this.createError(
            TaskirXErrorType.HTTP_ERROR,
            data.error.message,
            response.status
          );
        }

        if (this.config.debug) {
          console.log(`[TaskirX] ✅ ${method} ${url} - Success`);
        }

        return data.data as T;
      } catch (error) {
        if (error instanceof Error && 'type' in error) {
          lastError = error as TaskirXError;
        } else if (error instanceof TypeError) {
          lastError = this.createError(
            TaskirXErrorType.NETWORK_ERROR,
            error.message,
            undefined,
            error
          );
        } else {
          lastError = this.createError(
            TaskirXErrorType.NETWORK_ERROR,
            'Unknown error occurred',
            undefined,
            error instanceof Error ? error : new Error(String(error))
          );
        }

        if (attempt < this.config.retryAttempts! - 1) {
          const delayMs = delays[Math.min(attempt, delays.length - 1)];
          if (this.config.debug) {
            console.log(`[TaskirX] ⚠️ Retry attempt ${attempt + 1} after ${delayMs}ms`);
          }
          await this.sleep(delayMs);
        } else {
          if (this.config.debug) {
            console.log(`[TaskirX] ❌ ${method} ${url} - ${lastError.message}`);
          }
        }
      }
    }

    throw lastError || this.createError(
      TaskirXErrorType.RETRY_EXHAUSTED,
      'Max retries exhausted'
    );
  }

  private buildHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'X-API-Key': this.config.apiKey,
      'X-Request-ID': this.generateRequestId(),
      'User-Agent': this.buildUserAgent(),
    };

    if (this.authToken) {
      headers['Authorization'] = `Bearer ${this.authToken}`;
    }

    return headers;
  }

  private buildUserAgent(): string {
    const platform = Platform.OS === 'ios' ? 'iOS' : 
                    Platform.OS === 'android' ? 'Android' : 'Unknown';
    return `TaskirX-RN/1.0 (${platform}/${Platform.Version})`;
  }

  private generateRequestId(): string {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
