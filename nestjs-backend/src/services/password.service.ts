/**
 * Password Hashing & Validation Service
 * Implements bcryptjs password hashing with configurable salt rounds
 */

import { Injectable } from '@nestjs/common';
import * as bcrypt from 'bcryptjs';

@Injectable()
export class PasswordService {
  private readonly saltRounds = 10; // OWASP recommendation

  /**
   * Hash password
   * Uses bcryptjs with 10 salt rounds (CPU-intensive, protects against brute-force)
   */
  async hashPassword(password: string): Promise<string> {
    if (!password || password.length < 8) {
      throw new Error('Password must be at least 8 characters');
    }

    return bcrypt.hash(password, this.saltRounds);
  }

  /**
   * Validate password against hash
   */
  async validatePassword(password: string, hash: string): Promise<boolean> {
    return bcrypt.compare(password, hash);
  }

  /**
   * Check password strength
   * Returns score 0-4 (0=weak, 4=strong)
   */
  checkPasswordStrength(password: string): {
    score: number;
    feedback: string[];
  } {
    const feedback: string[] = [];
    let score = 0;

    if (!password) {
      return { score: 0, feedback: ['Password is required'] };
    }

    // Length check (minimum 8, bonus for 12+, 16+)
    if (password.length >= 8) {
      score++;
      feedback.push('✓ Minimum length (8+) met');
    } else {
      feedback.push('✗ Password must be at least 8 characters');
      return { score: 0, feedback };
    }

    if (password.length >= 12) {
      score++;
      feedback.push('✓ Good length (12+)');
    }

    if (password.length >= 16) {
      score++;
      feedback.push('✓ Excellent length (16+)');
    }

    // Complexity checks
    if (/[a-z]/.test(password)) {
      feedback.push('✓ Contains lowercase letters');
    } else {
      feedback.push('✗ Add lowercase letters');
    }

    if (/[A-Z]/.test(password)) {
      feedback.push('✓ Contains uppercase letters');
    } else {
      feedback.push('✗ Add uppercase letters');
    }

    if (/[0-9]/.test(password)) {
      feedback.push('✓ Contains numbers');
    } else {
      feedback.push('✗ Add numbers');
    }

    if (/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) {
      score++;
      feedback.push('✓ Contains special characters');
    } else {
      feedback.push('✗ Add special characters for more strength');
    }

    // Entropy check (avoid common patterns)
    const commonPatterns = [
      /^[a-z]{8,}$/, // only lowercase
      /^[0-9]{8,}$/, // only numbers
      /^[a-z0-9]{8,}$/, // simple alphanumeric
      /^(.)\1{7,}$/, // repeated character
      /^(123|234|345|456|567|678|789|890|abc|bcd|cde)/, // sequential
    ];

    const hasCommonPattern = commonPatterns.some(pattern => pattern.test(password));
    if (!hasCommonPattern) {
      score++;
      feedback.push('✓ Does not match common patterns');
    } else {
      feedback.push('✗ Avoid sequential or repeated characters');
    }

    return { score: Math.min(score, 4), feedback };
  }

  /**
   * Generate temporary password for admin reset
   */
  generateTemporaryPassword(): string {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*';
    let password = '';
    for (let i = 0; i < 16; i++) {
      password += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return password;
  }
}

/**
 * Environment Variable Validator
 * Ensures all critical secrets are properly configured
 */
export class EnvValidator {
  /**
   * Validate all required environment variables
   */
  static validateAll(): void {
    const required = [
      { key: 'NODE_ENV', validation: (v) => /^(development|production|staging)$/.test(v), msg: 'Must be development, production, or staging' },
      { key: 'DATABASE_HOST', validation: (v) => v && v.length > 0, msg: 'Cannot be empty' },
      { key: 'DATABASE_PORT', validation: (v) => !isNaN(parseInt(v)), msg: 'Must be a valid port number' },
      { key: 'DATABASE_USER', validation: (v) => v && v.length > 0, msg: 'Cannot be empty' },
      { key: 'DATABASE_PASSWORD', validation: (v) => v && v.length >= 12, msg: 'Must be at least 12 characters (min security)' },
      { key: 'DATABASE_NAME', validation: (v) => v && v.length > 0, msg: 'Cannot be empty' },
      { key: 'JWT_SECRET', validation: (v) => v && v.length >= 32, msg: 'Must be at least 32 characters for HS256' },
      { key: 'REDIS_HOST', validation: (v) => v && v.length > 0, msg: 'Cannot be empty' },
      { key: 'REDIS_PORT', validation: (v) => !isNaN(parseInt(v)), msg: 'Must be a valid port number' },
      { key: 'REDIS_PASSWORD', validation: (v) => v && v.length >= 12, msg: 'Must be at least 12 characters (min security)' },
      { key: 'CORS_ORIGIN', validation: (v) => v && v.length > 0, msg: 'Cannot be empty' },
    ];

    const errors: string[] = [];

    for (const { key, validation, msg } of required) {
      const value = process.env[key];
      if (!value || !validation(value)) {
        errors.push(`[${key}] ${msg}`);
      }
    }

    // Check for hardcoded secrets in common variable names
    this.checkForHardcodedSecrets();

    if (errors.length > 0) {
      throw new Error(`Environment validation failed:\n${errors.join('\n')}`);
    }

    console.log('✓ All environment variables validated');
  }

  /**
   * Check for hardcoded/default secrets
   */
  private static checkForHardcodedSecrets(): void {
    const defaults = [
      { key: 'JWT_SECRET', patterns: ['secret', 'test', 'default', '12345678'] },
      { key: 'DATABASE_PASSWORD', patterns: ['password', 'admin', 'test', '12345678'] },
      { key: 'REDIS_PASSWORD', patterns: ['password', 'admin', 'test', '12345678'] },
    ];

    if (process.env.NODE_ENV === 'production') {
      for (const { key, patterns } of defaults) {
        const value = process.env[key]?.toLowerCase() || '';
        if (patterns.some(pattern => value === pattern)) {
          throw new Error(`[SECURITY] ${key} appears to be using a default value in PRODUCTION. This is not allowed.`);
        }
      }
    }
  }

  /**
   * Log environment configuration (without revealing secrets)
   */
  static logConfiguration(): void {
    console.log('\n╔════════════════════════════════════════╗');
    console.log('║  Environment Configuration Summary    ║');
    console.log('╚════════════════════════════════════════╝');

    const config = [
      { key: 'NODE_ENV', value: process.env.NODE_ENV },
      { key: 'DATABASE_HOST', value: process.env.DATABASE_HOST },
      { key: 'DATABASE_PORT', value: process.env.DATABASE_PORT },
      { key: 'REDIS_HOST', value: process.env.REDIS_HOST },
      { key: 'REDIS_PORT', value: process.env.REDIS_PORT },
      { key: 'CORS_ORIGIN', value: process.env.CORS_ORIGIN },
    ];

    for (const { key, value } of config) {
      console.log(`  ${key}: ${value}`);
    }

    const secrets = ['DATABASE_PASSWORD', 'REDIS_PASSWORD', 'JWT_SECRET'];
    for (const secret of secrets) {
      const value = process.env[secret];
      const masked = value ? `${value.substring(0, 4)}...${value.substring(value.length - 4)}` : '❌ NOT SET';
      console.log(`  ${secret}: ${masked}`);
    }

    console.log('\n');
  }
}

/**
 * Credential Rotation Scheduler
 * Tracks and reminds about credential rotation every 90 days
 */
export class CredentialRotationScheduler {
  private static readonly ROTATION_INTERVAL_DAYS = 90;
  private lastRotationDates: Map<string, Date> = new Map();

  /**
   * Record credential rotation
   */
  recordRotation(credentialType: string): void {
    this.lastRotationDates.set(credentialType, new Date());
    console.log(`✓ Recorded rotation for: ${credentialType}`);
  }

  /**
   * Check if rotation is due
   */
  isRotationDue(credentialType: string): boolean {
    const lastRotation = this.lastRotationDates.get(credentialType);
    if (!lastRotation) return true; // If never rotated, it's due

    const daysSinceRotation = Math.floor(
      (Date.now() - lastRotation.getTime()) / (1000 * 60 * 60 * 24),
    );

    return daysSinceRotation >= CredentialRotationScheduler.ROTATION_INTERVAL_DAYS;
  }

  /**
   * Get next rotation date
   */
  getNextRotationDate(credentialType: string): Date {
    const lastRotation = this.lastRotationDates.get(credentialType) || new Date();
    const nextRotation = new Date(lastRotation);
    nextRotation.setDate(nextRotation.getDate() + CredentialRotationScheduler.ROTATION_INTERVAL_DAYS);
    return nextRotation;
  }

  /**
   * Get rotation status for all credentials
   */
  getRotationStatus(): { [key: string]: { isDue: boolean; nextRotation: Date } } {
    const credentials = ['DATABASE_PASSWORD', 'REDIS_PASSWORD', 'JWT_SECRET', 'API_KEY'];
    const status: { [key: string]: { isDue: boolean; nextRotation: Date } } = {};

    for (const credential of credentials) {
      status[credential] = {
        isDue: this.isRotationDue(credential),
        nextRotation: this.getNextRotationDate(credential),
      };
    }

    return status;
  }
}
