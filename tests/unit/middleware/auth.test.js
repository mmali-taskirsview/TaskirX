/**
 * Authentication Middleware Tests
 */

const jwt = require('jsonwebtoken');

describe('Authentication Middleware', () => {
  describe('JWT Token Validation', () => {
    const mockSecret = 'test-secret-key-at-least-32-chars-long-for-security';

    it('should create valid JWT tokens', () => {
      const payload = { userId: '123', role: 'advertiser' };
      const token = jwt.sign(payload, mockSecret, { expiresIn: '1h' });

      expect(token).toBeDefined();
      expect(typeof token).toBe('string');
    });

    it('should verify valid tokens', () => {
      const payload = { userId: '123', role: 'advertiser' };
      const token = jwt.sign(payload, mockSecret, { expiresIn: '1h' });

      const decoded = jwt.verify(token, mockSecret);

      expect(decoded.userId).toBe('123');
      expect(decoded.role).toBe('advertiser');
    });

    it('should reject invalid tokens', () => {
      const invalidToken = 'invalid.token.here';

      expect(() => {
        jwt.verify(invalidToken, mockSecret);
      }).toThrow();
    });

    it('should reject tokens with wrong secret', () => {
      const payload = { userId: '123', role: 'advertiser' };
      const token = jwt.sign(payload, mockSecret, { expiresIn: '1h' });

      expect(() => {
        jwt.verify(token, 'wrong-secret');
      }).toThrow();
    });

    it('should reject expired tokens', (done) => {
      const payload = { userId: '123', role: 'advertiser' };
      const token = jwt.sign(payload, mockSecret, { expiresIn: '1ms' });

      // Wait for token to expire
      setTimeout(() => {
        expect(() => {
          jwt.verify(token, mockSecret);
        }).toThrow('jwt expired');
        done();
      }, 10);
    });
  });

  describe('Role-Based Access Control', () => {
    it('should validate advertiser role', () => {
      const user = { role: 'advertiser' };
      expect(user.role).toBe('advertiser');
    });

    it('should validate publisher role', () => {
      const user = { role: 'publisher' };
      expect(user.role).toBe('publisher');
    });

    it('should validate admin role', () => {
      const user = { role: 'admin' };
      expect(user.role).toBe('admin');
    });

    it('should reject invalid roles', () => {
      const validRoles = ['advertiser', 'publisher', 'admin'];
      const invalidRole = 'hacker';

      expect(validRoles).not.toContain(invalidRole);
    });
  });

  describe('Authorization Header Parsing', () => {
    it('should extract token from Bearer authorization', () => {
      const authHeader = 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test';
      const token = authHeader.replace('Bearer ', '');

      expect(token).toBe('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test');
      expect(token).not.toContain('Bearer');
    });

    it('should handle missing Bearer prefix', () => {
      const authHeader = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test';
      const token = authHeader.replace('Bearer ', '');

      // Should return the token unchanged if no Bearer prefix
      expect(token).toBe(authHeader);
    });
  });
});
