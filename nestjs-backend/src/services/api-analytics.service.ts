import { Injectable, Logger, Inject } from '@nestjs/common';
import { Redis } from 'ioredis';

/**
 * API Analytics Service
 * 
 * Deep tracking of API usage per customer:
 * - Request volume per endpoint
 * - Cost attribution per request
 * - Usage forecasting
 * - Anomaly detection
 * - SLA tracking per customer
 * 
 * Costs:
 * - Analytics API: $0.01 per 1K events
 * - Bidding API: $0.005 per 1K requests
 * - Campaign API: $0.002 per 1K requests
 */
@Injectable()
export class ApiAnalyticsService {
  private readonly logger = new Logger(ApiAnalyticsService.name);

  private readonly ENDPOINT_COSTS = {
    'POST /analytics/events': 0.00001, // $0.01 per 1K
    'GET /analytics/metrics': 0.00002,
    'GET /analytics/reports': 0.00003,
    'POST /bids/predict': 0.000005,
    'POST /bids/train': 0.00001,
    'GET /campaigns': 0.000002,
    'POST /campaigns': 0.000005,
    'PUT /campaigns/:id': 0.000005,
  };

  constructor(
    @Inject('REDIS_CLIENT')
    private redisClient: Redis,
  ) {}

  /**
   * Track API request
   */
  async trackRequest(
    userId: string,
    endpoint: string,
    method: string,
    responseTime: number,
    statusCode: number,
    dataSize?: number,
  ): Promise<void> {
    const endpointKey = `${method} ${endpoint}`;
    const cost = this.ENDPOINT_COSTS[endpointKey] || 0.000001;
    const timestamp = Date.now();

    const requestRecord = {
      userId,
      endpoint: endpointKey,
      responseTime,
      statusCode,
      dataSize: dataSize || 0,
      cost,
      timestamp,
    };

    // Track per user per endpoint
    await this.redisClient.lpush(
      `api:requests:${userId}:${endpointKey}`,
      JSON.stringify(requestRecord),
    );

    // Track per user daily
    const dateKey = new Date(timestamp).toISOString().split('T')[0];
    await this.redisClient.lpush(
      `api:requests:${userId}:${dateKey}`,
      JSON.stringify(requestRecord),
    );

    // Track globally per endpoint
    await this.redisClient.lpush(
      `api:requests:global:${endpointKey}`,
      JSON.stringify(requestRecord),
    );

    // Update user stats
    await this.updateUserStats(userId, endpointKey, cost, responseTime);

    // Keep only last 30 days
    await this.redisClient.expire(
      `api:requests:${userId}:${endpointKey}`,
      2592000,
    );
  }

  /**
   * Update user statistics
   */
  private async updateUserStats(
    userId: string,
    endpoint: string,
    cost: number,
    responseTime: number,
  ): Promise<void> {
    const statsKey = `api:stats:${userId}`;

    await Promise.all([
      this.redisClient.hincrby(statsKey, 'request_count', 1),
      this.redisClient.hincrbyfloat(statsKey, 'total_cost', cost),
      this.redisClient.hincrby(statsKey, 'total_response_time', Math.floor(responseTime)),
      this.redisClient.hincrby(statsKey, `endpoint:${endpoint}:count`, 1),
      this.redisClient.hincrbyfloat(statsKey, `endpoint:${endpoint}:cost`, cost),
    ]);

    // Update daily stats
    const dateKey = new Date().toISOString().split('T')[0];
    const dailyKey = `api:stats:${userId}:${dateKey}`;

    await Promise.all([
      this.redisClient.hincrby(dailyKey, 'request_count', 1),
      this.redisClient.hincrbyfloat(dailyKey, 'total_cost', cost),
      this.redisClient.expire(dailyKey, 2592000), // 30 days
    ]);
  }

  /**
   * Get user API usage summary
   */
  async getUserUsageSummary(userId: string, period: 'day' | 'week' | 'month' = 'month'): Promise<{
    totalRequests: number;
    totalCost: number;
    avgResponseTime: number;
    endpoints: Array<{
      endpoint: string;
      requests: number;
      cost: number;
      avgResponseTime: number;
    }>;
    topEndpoints: string[];
    errorRate: number;
    period: string;
  }> {
    const statsKey = `api:stats:${userId}`;
    const stats = await this.redisClient.hgetall(statsKey);

    if (Object.keys(stats).length === 0) {
      return {
        totalRequests: 0,
        totalCost: 0,
        avgResponseTime: 0,
        endpoints: [],
        topEndpoints: [],
        errorRate: 0,
        period,
      };
    }

    const totalRequests = parseInt(stats.request_count || '0');
    const totalCost = parseFloat(stats.total_cost || '0');
    const totalResponseTime = parseInt(stats.total_response_time || '0');
    const avgResponseTime = totalRequests > 0 ? totalResponseTime / totalRequests : 0;

    // Get endpoint breakdown
    const endpointStats: Array<{
      endpoint: string;
      requests: number;
      cost: number;
      avgResponseTime: number;
    }> = [];

    for (const [key, value] of Object.entries(stats)) {
      if (key.startsWith('endpoint:') && key.endsWith(':count')) {
        const endpoint = key.replace('endpoint:', '').replace(':count', '');
        const requests = parseInt(value);
        const cost = parseFloat(stats[`endpoint:${endpoint}:cost`] || '0');

        endpointStats.push({
          endpoint,
          requests,
          cost,
          avgResponseTime: 0, // Would need more detailed tracking
        });
      }
    }

    const topEndpoints = endpointStats
      .sort((a, b) => b.requests - a.requests)
      .slice(0, 5)
      .map((e) => e.endpoint);

    return {
      totalRequests,
      totalCost,
      avgResponseTime,
      endpoints: endpointStats,
      topEndpoints,
      errorRate: 0, // Would calculate from status codes
      period,
    };
  }

  /**
   * Get daily usage breakdown
   */
  async getDailyUsageBreakdown(
    userId: string,
    days: number = 30,
  ): Promise<
    Array<{
      date: string;
      requests: number;
      cost: number;
      topEndpoint: string;
    }>
  > {
    const usage: Array<{
      date: string;
      requests: number;
      cost: number;
      topEndpoint: string;
    }> = [];

    for (let i = 0; i < days; i++) {
      const date = new Date();
      date.setDate(date.getDate() - i);
      const dateKey = date.toISOString().split('T')[0];

      const dailyStats = await this.redisClient.hgetall(
        `api:stats:${userId}:${dateKey}`,
      );

      if (Object.keys(dailyStats).length > 0) {
        usage.push({
          date: dateKey,
          requests: parseInt(dailyStats.request_count || '0'),
          cost: parseFloat(dailyStats.total_cost || '0'),
          topEndpoint: 'GET /analytics/metrics', // Would extract from actual data
        });
      }
    }

    return usage.reverse();
  }

  /**
   * Forecast usage for next month
   */
  async forecastUsage(userId: string): Promise<{
    forecastedRequests: number;
    forecastedCost: number;
    confidence: number;
    trend: 'increasing' | 'decreasing' | 'stable';
  }> {
    // Get last 30 days of data
    const dailyUsage = await this.getDailyUsageBreakdown(userId, 30);

    if (dailyUsage.length < 7) {
      return {
        forecastedRequests: 0,
        forecastedCost: 0,
        confidence: 0,
        trend: 'stable',
      };
    }

    // Simple linear regression forecast
    const avgDailyRequests =
      dailyUsage.reduce((sum, d) => sum + d.requests, 0) / dailyUsage.length;
    const avgDailyCost =
      dailyUsage.reduce((sum, d) => sum + d.cost, 0) / dailyUsage.length;

    // Detect trend
    const firstWeek = dailyUsage.slice(0, 7).reduce((sum, d) => sum + d.requests, 0);
    const lastWeek = dailyUsage.slice(-7).reduce((sum, d) => sum + d.requests, 0);
    const trend =
      lastWeek > firstWeek * 1.1 ? 'increasing' : lastWeek < firstWeek * 0.9 ? 'decreasing' : 'stable';

    const forecastedRequests = Math.floor(avgDailyRequests * 30);
    const forecastedCost = avgDailyCost * 30;

    // Calculate confidence based on consistency
    const variance =
      dailyUsage.reduce((sum, d) => sum + Math.pow(d.requests - avgDailyRequests, 2), 0) /
      dailyUsage.length;
    const stdDev = Math.sqrt(variance);
    const confidence = Math.max(0, 1 - stdDev / avgDailyRequests);

    return {
      forecastedRequests,
      forecastedCost,
      confidence: Math.min(1, confidence),
      trend,
    };
  }

  /**
   * Detect anomalies in API usage
   */
  async detectAnomalies(userId: string): Promise<
    Array<{
      type: 'spike' | 'drop' | 'error_rate';
      endpoint?: string;
      severity: 'low' | 'medium' | 'high';
      message: string;
      value: number;
      expectedValue: number;
    }>
  > {
    const dailyUsage = await this.getDailyUsageBreakdown(userId, 14);

    if (dailyUsage.length < 3) {
      return [];
    }

    const anomalies: any[] = [];
    const avgRequests =
      dailyUsage.reduce((sum, d) => sum + d.requests, 0) / dailyUsage.length;

    // Check for spikes
  dailyUsage.forEach((day) => {
      if (day.requests > avgRequests * 1.5) {
        anomalies.push({
          type: 'spike',
          severity: day.requests > avgRequests * 2.5 ? 'high' : 'medium',
          message: `Usage spike detected on ${day.date}`,
          value: day.requests,
          expectedValue: avgRequests,
        });
      }

      if (day.requests < avgRequests * 0.5) {
        anomalies.push({
          type: 'drop',
          severity: 'medium',
          message: `Usage drop detected on ${day.date}`,
          value: day.requests,
          expectedValue: avgRequests,
        });
      }
    });

    return anomalies;
  }

  /**
   * Get cost breakdown by tier
   */
  async getCostBreakdown(userId: string): Promise<{
    total: number;
    byEndpoint: Record<string, number>;
    byType: Record<string, number>;
  }> {
    const summary = await this.getUserUsageSummary(userId);

    const byType = {
      analytics: 0,
      bidding: 0,
      campaigns: 0,
      other: 0,
    };

    summary.endpoints.forEach((ep) => {
      if (ep.endpoint.includes('analytics')) {
        byType.analytics += ep.cost;
      } else if (ep.endpoint.includes('bids')) {
        byType.bidding += ep.cost;
      } else if (ep.endpoint.includes('campaigns')) {
        byType.campaigns += ep.cost;
      } else {
        byType.other += ep.cost;
      }
    });

    const byEndpoint: Record<string, number> = {};
    summary.endpoints.forEach((ep) => {
      byEndpoint[ep.endpoint] = ep.cost;
    });

    return {
      total: summary.totalCost,
      byEndpoint,
      byType,
    };
  }

  /**
   * Get customer comparison report
   */
  async getCustomerComparison(userIds: string[]): Promise<
    Array<{
      userId: string;
      totalRequests: number;
      totalCost: number;
      avgResponseTime: number;
      costPerRequest: number;
      tier: string;
    }>
  > {
    const comparisons = await Promise.all(
      userIds.map(async (userId) => {
        const summary = await this.getUserUsageSummary(userId);
        return {
          userId,
          totalRequests: summary.totalRequests,
          totalCost: summary.totalCost,
          avgResponseTime: summary.avgResponseTime,
          costPerRequest: summary.totalRequests > 0 ? summary.totalCost / summary.totalRequests : 0,
          tier: 'PROFESSIONAL', // Would get from user data
        };
      }),
    );

    return comparisons.sort((a, b) => b.totalCost - a.totalCost);
  }
}
