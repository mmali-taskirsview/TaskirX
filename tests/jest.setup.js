/**
 * Jest Setup File
 * Runs before all tests to set up environment
 */

// Set test environment variables
process.env.NODE_ENV = 'test';
process.env.JWT_SECRET = 'test-jwt-secret-key-for-testing-purposes-only-do-not-use-in-production';
process.env.MONGODB_URI = 'mongodb://localhost:27017/taskirx_test';
process.env.REDIS_ENABLED = 'false';
process.env.PORT = '3001';

// Suppress console logs during tests (optional)
// global.console = {
//   ...console,
//   log: jest.fn(),
//   debug: jest.fn(),
//   info: jest.fn(),
//   warn: jest.fn(),
//   error: jest.fn(),
// };
