/**
 * Authentication Service
 * Handles user registration, login, and token management
 */

import { RequestManager } from './RequestManager';
import { Logger } from '../utils/Logger';

export class AuthService {
  constructor(private requestManager: RequestManager, private logger: Logger) {}

  async register(email: string, password: string, companyName: string): Promise<any> {
    this.logger.debug('Registering user:', email);
    return this.requestManager.post('/api/auth/register', {
      email,
      password,
      company_name: companyName,
    });
  }

  async login(email: string, password: string): Promise<any> {
    this.logger.debug('Logging in:', email);
    const response = await this.requestManager.post('/api/auth/login', {
      email,
      password,
    });
    if (response.access_token) {
      this.requestManager.setToken(response.access_token);
    }
    return response;
  }

  async refreshToken(refreshToken: string): Promise<any> {
    this.logger.debug('Refreshing token');
    const response = await this.requestManager.post('/api/auth/refresh', {
      refresh_token: refreshToken,
    });
    if (response.access_token) {
      this.requestManager.setToken(response.access_token);
    }
    return response;
  }

  async logout(): Promise<any> {
    this.logger.debug('Logging out');
    const response = await this.requestManager.post('/api/auth/logout', {});
    this.requestManager.clearToken();
    return response;
  }

  async getProfile(): Promise<any> {
    this.logger.debug('Fetching profile');
    return this.requestManager.get('/api/auth/profile');
  }
}
