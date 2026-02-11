/**
 * AuthService Tests
 */

import { AuthService } from '../src/services/AuthService';
import { RequestManager } from '../src/services/RequestManager';
import { Logger } from '../src/utils/Logger';

describe('AuthService', () => {
  let authService: AuthService;
  let requestManager: jest.Mocked<RequestManager>;
  let logger: Logger;

  beforeEach(() => {
    logger = new Logger(false);
    requestManager = {
      post: jest.fn(),
      get: jest.fn(),
      put: jest.fn(),
      delete: jest.fn(),
    } as any;

    authService = new AuthService(requestManager, logger);
  });

  describe('register', () => {
    it('should register a new user', async () => {
      const userData = {
        email: 'test@example.com',
        password: 'password123',
        name: 'Test User',
      };

      const mockResponse = {
        id: 'user-123',
        email: userData.email,
        name: userData.name,
        token: 'jwt-token-123',
      };

      requestManager.post.mockResolvedValueOnce(mockResponse);

      const result = await authService.register(userData);

      expect(requestManager.post).toHaveBeenCalledWith('/api/auth/register', userData);
      expect(result).toEqual(mockResponse);
    });

    it('should handle registration errors', async () => {
      const userData = {
        email: 'test@example.com',
        password: 'password123',
        name: 'Test User',
      };

      requestManager.post.mockRejectedValueOnce(new Error('Registration failed'));

      await expect(authService.register(userData)).rejects.toThrow('Registration failed');
    });
  });

  describe('login', () => {
    it('should login user and set token', async () => {
      const credentials = {
        email: 'test@example.com',
        password: 'password123',
      };

      const mockResponse = {
        token: 'jwt-token-123',
        user: {
          id: 'user-123',
          email: credentials.email,
        },
      };

      requestManager.post.mockResolvedValueOnce(mockResponse);

      const result = await authService.login(credentials);

      expect(requestManager.post).toHaveBeenCalledWith('/api/auth/login', credentials);
      expect(result).toEqual(mockResponse);
    });
  });

  describe('logout', () => {
    it('should logout user', async () => {
      requestManager.post.mockResolvedValueOnce({ success: true });

      await authService.logout();

      expect(requestManager.post).toHaveBeenCalledWith('/api/auth/logout', {});
    });
  });

  describe('getProfile', () => {
    it('should fetch user profile', async () => {
      const mockProfile = {
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
        createdAt: new Date(),
      };

      requestManager.get.mockResolvedValueOnce(mockProfile);

      const result = await authService.getProfile();

      expect(requestManager.get).toHaveBeenCalledWith('/api/auth/profile');
      expect(result).toEqual(mockProfile);
    });
  });

  describe('refreshToken', () => {
    it('should refresh authentication token', async () => {
      const mockResponse = {
        token: 'new-jwt-token-456',
      };

      requestManager.post.mockResolvedValueOnce(mockResponse);

      const result = await authService.refreshToken();

      expect(requestManager.post).toHaveBeenCalledWith('/api/auth/refresh', {});
      expect(result).toEqual(mockResponse);
    });
  });
});
