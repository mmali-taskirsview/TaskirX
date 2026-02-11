/**
 * Request Manager - Handles all HTTP communication
 * Includes retry logic, error handling, and request/response interceptors
 */

import { ClientConfig } from '../types';
import { Logger } from '../utils/Logger';
import { ErrorHandler } from '../utils/ErrorHandler';

export class RequestManager {
  private baseUrl: string;
  private apiKey: string;
  private token?: string;
  private timeout: number;
  private retries: number;
  private logger: Logger;
  private requestId = 0;

  constructor(config: ClientConfig, logger: Logger) {
    this.baseUrl = config.baseUrl || 'https://api.taskirx.com';
    this.apiKey = config.apiKey;
    this.timeout = config.timeout || 30000;
    this.retries = config.retries || 3;
    this.logger = logger;
  }

  /**
   * Set authentication token
   */
  setToken(token: string): void {
    this.token = token;
  }

  /**
   * Clear authentication token
   */
  clearToken(): void {
    this.token = undefined;
  }

  /**
   * Perform GET request
   */
  async get<T>(endpoint: string, options?: any): Promise<T> {
    return this.request<T>('GET', endpoint, undefined, options);
  }

  /**
   * Perform POST request
   */
  async post<T>(endpoint: string, data?: any, options?: any): Promise<T> {
    return this.request<T>('POST', endpoint, data, options);
  }

  /**
   * Perform PUT request
   */
  async put<T>(endpoint: string, data?: any, options?: any): Promise<T> {
    return this.request<T>('PUT', endpoint, data, options);
  }

  /**
   * Perform DELETE request
   */
  async delete<T>(endpoint: string, options?: any): Promise<T> {
    return this.request<T>('DELETE', endpoint, undefined, options);
  }

  /**
   * Generic request method with retry logic
   */
  private async request<T>(
    method: string,
    endpoint: string,
    data?: any,
    options?: any
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const requestId = ++this.requestId;

    this.logger.debug(`[${requestId}] ${method} ${url}`);

    for (let attempt = 1; attempt <= this.retries; attempt++) {
      try {
        const response = await this.fetchWithTimeout(url, {
          method,
          headers: this.buildHeaders(),
          body: data ? JSON.stringify(data) : undefined,
          ...options,
        });

        if (!response.ok) {
          const error = await response.json().catch(() => ({ message: response.statusText }));
          throw ErrorHandler.fromResponse(response.status, error);
        }

        const result = await response.json();
        this.logger.debug(`[${requestId}] Response: ${JSON.stringify(result).substring(0, 100)}...`);
        return result as T;
      } catch (error) {
        this.logger.warn(`[${requestId}] Attempt ${attempt} failed:`, error);

        if (attempt === this.retries) {
          throw error;
        }

        // Exponential backoff: 100ms, 300ms, 900ms
        const backoff = Math.min(100 * Math.pow(2, attempt - 1), 5000);
        await this.delay(backoff);
      }
    }

    throw new Error('Max retries exceeded');
  }

  /**
   * Fetch with timeout
   */
  private fetchWithTimeout(url: string, options: any): Promise<Response> {
    return Promise.race([
      fetch(url, options),
      new Promise<Response>((_, reject) =>
        setTimeout(() => reject(new Error('Request timeout')), this.timeout)
      ),
    ]);
  }

  /**
   * Build request headers
   */
  private buildHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      'X-API-Key': this.apiKey,
      'X-Request-ID': `req_${this.requestId}`,
      'User-Agent': 'TaskirX-SDK/1.0.0 (JavaScript)',
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    return headers;
  }

  /**
   * Delay utility for backoff
   */
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
