import { Injectable, NestMiddleware, BadRequestException, Logger } from '@nestjs/common';
import { Request, Response, NextFunction } from 'express';
import { Inject } from '@nestjs/common';
import { Redis } from 'ioredis';

/**
 * Rate Limit Quota Service
 * 
 * Token bucket algorithm for rate limiting
 * Features:
 * - Per-subscription tier limits
 * - Burst allowance
 * - Per-endpoint limits
 * - User-based tracking
 * - Redis-backed for distributed systems
 * 
 * Quotas by tier:
 * STARTER: 10K requests/month, 10 req/sec burst
 * PROFESSIONAL: 1M requests/month, 100 req/sec burst
 * ENTERPRISE: 10M requests/month, 1000 req/sec burst
 */
@Injectable()
export class QuotaService {
  private readonly logger = new Logger(QuotaService.name);

  private readonly TIER_LIMITS = {
    STARTER: {
      monthlyRequests: 10000,
      burstLimit: 10,
      burstWindow: 60, // seconds
      endpoints: {
        'campaigns': 100,
        'analytics': 200,
        'bids': 150,
      },
    },
    PROFESSIONAL: {
      monthlyRequests: 1000000,
      burstLimit: 100,
      burstWindow: 60,
      endpoints: {
        'campaigns': 1000,
        'analytics': 5000,
        'bids': 2000,
      },
    },
    ENTERPRISE: {
      monthlyRequests: 10000000,
      burstLimit: 1000,
      burstWindow: 60,
      endpoints: {
        'campaigns': 10000,
        'analytics': 50000,
        'bids': 20000,
      },
    },
  };

  constructor(
    @Inject('REDIS_CLIENT')
    private redisClient: Redis,
  ) {}

  /**
   * Check if request is within quota
   */
  async checkQuota(
    userId: string,
    tier: string,
    endpoint: string,
    weight: number = 1,
  ): Promise<{
    allowed: boolean;
    remaining: number;
    resetAt: Date;
    retryAfter?: number;
  }> {
    const tierConfig = this.TIER_LIMITS[tier] || this.TIER_LIMITS.STARTER;
    const endpointKey = `quota:${userId}:${endpoint}`;
    const monthlyKey = `quota:${userId}:monthly`;
    const burstKey = `quota:${userId}:burst:${endpoint}`;

    // Check endpoint-specific limit
    const currentEndpointUsage = await this.redisClient.get(endpointKey);
    const endpointLimit = tierConfig.endpoints[endpoint] || 100;

    if (currentEndpointUsage && parseInt(currentEndpointUsage) + weight > endpointLimit) {
      const ttl = await this.redisClient.ttl(endpointKey);
      return {
        allowed: false,
        remaining: Math.max(0, endpointLimit - parseInt(currentEndpointUsage)),
        resetAt: new Date(Date.now() + ttl * 1000),
        retryAfter: ttl,
      };
    }

    // Check burst limit (token bucket algorithm)
    const currentBurst = await this.redisClient.get(burstKey);
    const tokens = currentBurst ? parseInt(currentBurst) : tierConfig.burstLimit;

    if (tokens < weight) {
      return {
        allowed: false,
        remaining: tokens,
        resetAt: new Date(Date.now() + tierConfig.burstWindow * 1000),
        retryAfter: tierConfig.burstWindow,
      };
    }

    // Check monthly limit
    const monthlyUsage = await this.redisClient.get(monthlyKey);
    if (monthlyUsage && parseInt(monthlyUsage) + weight > tierConfig.monthlyRequests) {
      return {
        allowed: false,
        remaining: 0,
        resetAt: this.getMonthStart(new Date(1000 * 86400)),
      };
    }

    // Update quotas
    await Promise.all([
      // Endpoint quota (reset daily)
      this.redisClient.incrby(endpointKey, weight),
      this.redisClient.expire(endpointKey, 86400),
      
      // Burst quota (token bucket refill)
      this.redisClient.decrby(burstKey, weight),
      this.redisClient.expire(burstKey, tierConfig.burstWindow),
      
      // Monthly quota
      this.redisClient.incrby(monthlyKey, weight),
      this.redisClient.expire(monthlyKey, 2592000), // 30 days
    ]);

    const remaining = Math.max(0, tokens - weight);

    return {
      allowed: true,
      remaining,
      resetAt: new Date(Date.now() + tierConfig.burstWindow * 1000),
    };
  }

  /**
   * Get quota status for user
   */
  async getQuotaStatus(userId: string, tier: string) {
    const tierConfig = this.TIER_LIMITS[tier] || this.TIER_LIMITS.STARTER;

    const monthlyUsage = await this.redisClient.get(`quota:${userId}:monthly`);
    const endpointUsages = {};

    for (const [endpoint] of Object.entries(tierConfig.endpoints)) {
      const usage = await this.redisClient.get(`quota:${userId}:${endpoint}`);
      endpointUsages[endpoint] = {
        current: usage ? parseInt(usage) : 0,
        limit: tierConfig.endpoints[endpoint],
      };
    }

    return {
      tier,
      monthly: {
        current: monthlyUsage ? parseInt(monthlyUsage) : 0,
        limit: tierConfig.monthlyRequests,
        percentage: monthlyUsage
          ? (parseInt(monthlyUsage) / tierConfig.monthlyRequests) * 100
          : 0,
      },
      endpoints: endpointUsages,
      burst: {
        limit: tierConfig.burstLimit,
        window: tierConfig.burstWindow,
      },
    };
  }

  /**
   * Reset quota for user (admin operation)
   */
  async resetQuota(userId: string) {
    const pattern = `quota:${userId}:*`;
    const keys = await this.redisClient.keys(pattern);
    
    if (keys.length > 0) {
      await this.redisClient.del(...keys);
    }

    return { success: true, keysDeleted: keys.length };
  }

  /**
   * Get month start date
   */
  private getMonthStart(date: Date): Date {
    return new Date(date.getFullYear(), date.getMonth(), 1);
  }
}

/**
 * Rate Limiting Middleware
 * 
 * Enforces per-tier rate limits on incoming requests
 */
@Injectable()
export class RateLimitMiddleware implements NestMiddleware {
  private readonly logger = new Logger(RateLimitMiddleware.name);

  constructor(private quotaService: QuotaService) {}

  async use(req: Request, res: Response, next: NextFunction) {
    const userId = req.user?.id;
    const tier = req.user?.subscriptionTier || 'STARTER';
    const endpoint = this.extractEndpoint(req.path);

    if (!userId) {
      return next();
    }

    try {
      const weight = this.getRequestWeight(req);
      const quotaResult = await this.quotaService.checkQuota(
        userId,
        tier,
        endpoint,
        weight,
      );

      // Set rate limit headers
      res.setHeader('X-RateLimit-Limit', '1000');
      res.setHeader('X-RateLimit-Remaining', quotaResult.remaining);
      res.setHeader('X-RateLimit-Reset', quotaResult.resetAt.toISOString());

      if (!quotaResult.allowed) {
        res.setHeader('Retry-After', quotaResult.retryAfter || 60);
        throw new BadRequestException(
          `Rate limit exceeded. Reset at ${quotaResult.resetAt.toISOString()}`,
        );
      }

      next();
    } catch (error) {
      this.logger.error(`Rate limit check failed: ${error.message}`);
      if (error instanceof BadRequestException) {
        throw error;
      }
      next();
    }
  }

  /**
   * Extract endpoint from request path
   */
  private extractEndpoint(path: string): string {
    const segments = path.split('/').filter(Boolean);
    return segments[1] || 'default';
  }

  /**
   * Determine request weight based on operation type
   */
  private getRequestWeight(req: Request): number {
    // Write operations cost more
    if (req.method === 'POST' || req.method === 'PUT') {
      return 2;
    }
    if (req.method === 'DELETE') {
      return 3;
    }
    // Read operations cost 1
    return 1;
  }
}
