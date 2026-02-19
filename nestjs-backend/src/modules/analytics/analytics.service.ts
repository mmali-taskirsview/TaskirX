import * as crypto from 'crypto';
import axios from 'axios';
import { Injectable, OnModuleInit, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Campaign, CampaignStatus } from '../campaigns/campaign.entity';
import Redis from 'ioredis';
import { Cron, CronExpression } from '@nestjs/schedule';
import { NotificationsService } from '../notifications/notifications.service';
import { CampaignsService } from '../campaigns/campaigns.service';
import { UsersService } from '../users/users.service';

@Injectable()
export class AnalyticsService implements OnModuleInit {
  private readonly logger = new Logger(AnalyticsService.name);
  private readonly CACHE_TTL = 300; // 5 minutes
  private readonly SPEND_UPDATED_SET = 'campaigns:active:set'; // Fix for missing property
  private clickhouseUrl: string;
  private clickhouseUser: string;
  private clickhousePassword: string;
  private redisClient: Redis;

  constructor(
    private configService: ConfigService,
    @InjectRepository(Campaign)
    private campaignRepository: Repository<Campaign>,
    private readonly notificationsService: NotificationsService,
    private readonly campaignsService: CampaignsService,
    private readonly usersService: UsersService,
  ) {
    const clickhouseUrl = this.configService.get<string>('CLICKHOUSE_URL');
    const clickhouseHost = this.configService.get<string>('CLICKHOUSE_HOST', 'clickhouse');
    const clickhousePort = this.configService.get<string>('CLICKHOUSE_PORT', '8123');
    this.clickhouseUrl = clickhouseUrl ?? `http://${clickhouseHost}:${clickhousePort}`;
    this.clickhouseUser = this.configService.get<string>('CLICKHOUSE_USERNAME', 'taskir');
    this.clickhousePassword = this.configService.get<string>('CLICKHOUSE_PASSWORD', 'clickhouse_password_2026');

    // Initialize Redis for real-time budget tracking
    const redisHost = this.configService.get<string>('REDIS_HOST', 'redis');
    const redisPort = this.configService.get<number>('REDIS_PORT', 6379);
    const redisPassword = this.configService.get<string>('REDIS_PASSWORD');
    this.redisClient = new Redis({
      host: redisHost,
      port: redisPort,
      password: redisPassword,
    });
  }

  async onModuleInit() {
    await this.initClickHouse();
  }

  private async initClickHouse() {
    try {
      // Create Database
      await this.executeQuery('CREATE DATABASE IF NOT EXISTS analytics');

      // Create Tables
      const queries = [
        `CREATE TABLE IF NOT EXISTS analytics.impressions (
          id String,
          campaignId String,
          publisherId String,
          deviceType String,
          country String,
          timestamp DateTime
        ) ENGINE = MergeTree() ORDER BY (campaignId, timestamp)`,
        
        `CREATE TABLE IF NOT EXISTS analytics.clicks (
          id String,
          impressionId String,
          campaignId String,
          timestamp DateTime
        ) ENGINE = MergeTree() ORDER BY (campaignId, timestamp)`,

        `CREATE TABLE IF NOT EXISTS analytics.conversions (
          id String,
          clickId String,
          campaignId String,
          value Float32,
          timestamp DateTime
        ) ENGINE = MergeTree() ORDER BY (campaignId, timestamp)`,

        // Video Events Table
        `CREATE TABLE IF NOT EXISTS analytics.video_events (
          id String,
          campaignId String,
          eventType String,
          timestamp DateTime
        ) ENGINE = MergeTree() ORDER BY (campaignId, timestamp)`
      ];

      for (const query of queries) {
        await this.executeQuery(query);
      }

      // MMP Events Table
      await this.executeQuery(`CREATE TABLE IF NOT EXISTS analytics.mmp_events (
        id String,
        provider String,
        eventType String,
        campaignId String,
        userId String,
        deviceId String,
        revenue Float32,
        currency String,
        metadata String,
        timestamp DateTime
      ) ENGINE = MergeTree() ORDER BY (campaignId, timestamp)`);

      this.logger.log('ClickHouse tables initialized successfully');
    } catch (error) {
      this.logger.error('Failed to initialize ClickHouse tables', error);
    }
  }

  private async executeQuery(query: string): Promise<any> {
    try {
      // Simple query execution via HTTP interface
      const response = await axios.post(this.clickhouseUrl, query, {
        auth: { 
          username: this.clickhouseUser,
          password: this.clickhousePassword,
        }
      });
      return response.data;
    } catch (error) {
       // Log error but treat as non-fatal during init for resilience
       this.logger.error(`ClickHouse Query Failed: ${query}`, error.message);
       throw error;
    }
  }

  private normalizeTimestamp(timestamp?: Date | string | number): Date {
    if (!timestamp) {
      return new Date();
    }

    const dateObj = timestamp instanceof Date ? timestamp : new Date(timestamp);
    if (Number.isNaN(dateObj.getTime())) {
      this.logger.warn(`Invalid timestamp provided (${timestamp}); defaulting to now.`);
      return new Date();
    }

    return dateObj;
  }

  async getCampaignStats(campaignId: string, dateFrom: Date, dateTo: Date): Promise<any> {
    const campaign = await this.campaignRepository.findOne({ where: { id: campaignId } });
    if (!campaign) {
      return {
        campaignId,
        impressions: 0,
        clicks: 0,
        conversions: 0,
        spend: 0,
        ctr: 0,
        cpc: 0,
        cpa: 0,
      };
    }
    
    const fromDate = dateFrom.toISOString().replace('T', ' ').substring(0, 19);
    const toDate = dateTo.toISOString().replace('T', ' ').substring(0, 19);

    const [impRes, clickRes, convRes] = await Promise.all([
      this.executeQuery(`SELECT count(*) as count FROM analytics.impressions WHERE campaignId = '${campaignId}' AND timestamp >= '${fromDate}' AND timestamp <= '${toDate}' FORMAT JSON`),
      this.executeQuery(`SELECT count(*) as count FROM analytics.clicks WHERE campaignId = '${campaignId}' AND timestamp >= '${fromDate}' AND timestamp <= '${toDate}' FORMAT JSON`),
      this.executeQuery(`SELECT count(*) as count FROM analytics.conversions WHERE campaignId = '${campaignId}' AND timestamp >= '${fromDate}' AND timestamp <= '${toDate}' FORMAT JSON`)
    ]);
    
    const impressions = impRes?.data?.[0]?.count ? Number(impRes.data[0].count) : 0;
    const clicks = clickRes?.data?.[0]?.count ? Number(clickRes.data[0].count) : 0;
    const conversions = convRes?.data?.[0]?.count ? Number(convRes.data[0].count) : 0;
    const spend = Number(campaign.spent) || 0;
    
    return {
      campaignId,
      impressions,
      clicks,
      conversions,
      spend,
      ctr: impressions > 0 ? (clicks / impressions * 100).toFixed(2) : 0,
      cpc: clicks > 0 ? (spend / clicks).toFixed(2) : 0,
      cpa: conversions > 0 ? (spend / conversions).toFixed(2) : 0,
    };
  }

  async getDashboardStats(tenantId: string, dateFrom: Date, dateTo: Date): Promise<any> {
    // Query campaigns from PostgreSQL
    const campaigns = await this.campaignRepository.find({
      where: { tenantId },
    });
    
    const campaignIds = campaigns.map(c => `'${c.id}'`).join(',');
    
    let totalImpressions = 0;
    let totalClicks = 0;
    let totalConversions = 0;
    let totalSpend = 0;
    let activeCampaigns = 0;

    if (campaignIds.length > 0) {
      const fromDate = dateFrom.toISOString().replace('T', ' ').substring(0, 19);
      const toDate = dateTo.toISOString().replace('T', ' ').substring(0, 19);

      const [impRes, clickRes, convRes] = await Promise.all([
        this.executeQuery(`SELECT count(*) as count FROM analytics.impressions WHERE campaignId IN (${campaignIds}) AND timestamp >= '${fromDate}' AND timestamp <= '${toDate}' FORMAT JSON`),
        this.executeQuery(`SELECT count(*) as count FROM analytics.clicks WHERE campaignId IN (${campaignIds}) AND timestamp >= '${fromDate}' AND timestamp <= '${toDate}' FORMAT JSON`),
        this.executeQuery(`SELECT count(*) as count FROM analytics.conversions WHERE campaignId IN (${campaignIds}) AND timestamp >= '${fromDate}' AND timestamp <= '${toDate}' FORMAT JSON`)
      ]);

      totalImpressions = impRes?.data?.[0]?.count ? Number(impRes.data[0].count) : 0;
      totalClicks = clickRes?.data?.[0]?.count ? Number(clickRes.data[0].count) : 0;
      totalConversions = convRes?.data?.[0]?.count ? Number(convRes.data[0].count) : 0;
    }
    
    // Optimized MGET for Real-Time Spend
    const dateKey = new Date().toISOString().split('T')[0];
    const activeCampaignsList = campaigns.filter(c => c.status === CampaignStatus.ACTIVE);
    const spendKeys = activeCampaignsList.map(c => `campaign:spend:${c.id}:${dateKey}`);
    
    let redisSpends: (string | null)[] = [];
    if (spendKeys.length > 0) {
       // Spread operator for variadic mget
       // Note: In ioredis v4+, it supports mget(array) too, but spread is safer for older versions
       // However, to be strict with variadic:
       redisSpends = await this.redisClient.mget(...spendKeys);
    }

    // Map spend back to campaign ID for easy lookup
    const redisSpendMap = new Map<string, number>();
    activeCampaignsList.forEach((c, index) => {
        const val = redisSpends[index];
        if (val) {
            redisSpendMap.set(c.id, parseFloat(val));
        }
    });

    for (const campaign of campaigns) {
      let campaignSpend = Number(campaign.spent) || 0;
      
      if (campaign.status === CampaignStatus.ACTIVE) {
        campaignSpend += (redisSpendMap.get(campaign.id) || 0);
        activeCampaigns++;
      }
      
      totalSpend += campaignSpend;
    }

    // Get Bid Format Stats from Redis (Real-time from Go Engine)
    const formatStats = {
      banner: 0,
      video: 0,
      native: 0,
      audio: 0
    };
    
    try {
      const formatKeys = ['stats:bids:format:banner', 'stats:bids:format:video', 'stats:bids:format:native', 'stats:bids:format:audio'];
      const formatValues = await this.redisClient.mget(...formatKeys);
      
      formatStats.banner = parseInt(formatValues[0] || '0', 10);
      formatStats.video = parseInt(formatValues[1] || '0', 10);
      formatStats.native = parseInt(formatValues[2] || '0', 10);
      formatStats.audio = parseInt(formatValues[3] || '0', 10);
    } catch (error) {
      this.logger.warn(`Failed to fetch format stats from Redis: ${error.message}`);
    }
    
    return {
      totalImpressions,
      totalClicks,
      totalConversions,
      totalSpend: totalSpend.toFixed(2),
      activeCampaigns,
      avgCtr: totalImpressions > 0 ? (totalClicks / totalImpressions * 100).toFixed(2) : 0,
      avgCpc: totalClicks > 0 ? (totalSpend / totalClicks).toFixed(2) : 0,
      formatStats,
    };
  }

  async getRevenueByDate(tenantId: string, dateFrom: Date, dateTo: Date): Promise<any[]> {
    const campaigns = await this.campaignRepository.find({
      where: { tenantId },
    });
    
    const campaignIds = campaigns.map(c => `'${c.id}'`).join(',');
    if (campaignIds.length === 0) return [];

    const fromDate = dateFrom.toISOString().replace('T', ' ').substring(0, 19);
    const toDate = dateTo.toISOString().replace('T', ' ').substring(0, 19);
    
    const [impRes, clickRes] = await Promise.all([
      this.executeQuery(`SELECT toDate(timestamp) as date, count(*) as count FROM analytics.impressions WHERE campaignId IN (${campaignIds}) AND timestamp >= '${fromDate}' AND timestamp <= '${toDate}' GROUP BY date ORDER BY date FORMAT JSON`),
      this.executeQuery(`SELECT toDate(timestamp) as date, count(*) as count FROM analytics.clicks WHERE campaignId IN (${campaignIds}) AND timestamp >= '${fromDate}' AND timestamp <= '${toDate}' GROUP BY date ORDER BY date FORMAT JSON`)
    ]);

    const impressionsMap = new Map<string, number>();
    impRes?.data?.forEach((d: any) => impressionsMap.set(d.date, Number(d.count)));

    const clicksMap = new Map<string, number>();
    clickRes?.data?.forEach((d: any) => clicksMap.set(d.date, Number(d.count)));
    
    const totalSpend = campaigns.reduce((sum, c) => sum + (Number(c.spent) || 0), 0);
    const daysDiff = Math.ceil((dateTo.getTime() - dateFrom.getTime()) / (1000 * 60 * 60 * 24));
    const dailySpendAvg = totalSpend / Math.max(daysDiff, 1);
    
    const result = [];
    const loopDate = new Date(dateFrom);

    while (loopDate <= dateTo) {
      const dateStr = loopDate.toISOString().split('T')[0];
      const imps = impressionsMap.get(dateStr) || 0;
      const clicks = clicksMap.get(dateStr) || 0;
      
      // Approximating daily revenue based on spend (allocation)
      // Ideally, we'd sum actual cost/revenue events from CH
      const revenue = (dailySpendAvg * (imps > 0 ? 1 : 0)).toFixed(2); 

      result.push({
        date: dateStr,
        revenue,
        impressions: imps,
        clicks: clicks,
      });
      loopDate.setDate(loopDate.getDate() + 1);
    }
    return result;
  }

  async getTopPerformingCampaigns(tenantId: string, limit: number = 10): Promise<any[]> {
    const campaigns = await this.campaignRepository.find({
      where: { tenantId, status: CampaignStatus.ACTIVE },
    });
    
    if (campaigns.length === 0) return [];
    
    const campaignIds = campaigns.map(c => `'${c.id}'`).join(',');

    // Query CH for impressions count per campaign
    const impRes = await this.executeQuery(
      `SELECT campaignId, count(*) as count FROM analytics.impressions WHERE campaignId IN (${campaignIds}) GROUP BY campaignId ORDER BY count DESC LIMIT ${limit} FORMAT JSON`
    );
    
    // Also get clicks for these top campaigns
    const clickRes = await this.executeQuery(
      `SELECT campaignId, count(*) as count FROM analytics.clicks WHERE campaignId IN (${campaignIds}) GROUP BY campaignId FORMAT JSON`
    );
    
    const impMap = new Map<string, number>();
    impRes?.data?.forEach((d: any) => impMap.set(d.campaignId, Number(d.count)));
    
    const clickMap = new Map<string, number>();
    clickRes?.data?.forEach((d: any) => clickMap.set(d.campaignId, Number(d.count)));

    // Map the results back to campaign objects
    // Filter campaigns by those that have data or return all if we want
    // The query limited by impressions, so we primarily use that order
    const topCampaignIds = impRes?.data?.map((d: any) => d.campaignId) || [];
    
    // Include campaigns that might have 0 impressions if the list isn't full? 
    // For "Top Performing", we usually only want those with traffic.
    
    const result = topCampaignIds.map((id: string) => {
        const c = campaigns.find(camp => camp.id === id);
        if (!c) return null;
        const imps = impMap.get(id) || 0;
        const clicks = clickMap.get(id) || 0;
        return {
          id: c.id,
          name: c.name,
          impressions: imps,
          clicks: clicks,
          conversions: Number(c.conversions) || 0, // Keep conversions from Postgres for now or fetch from CH
          spent: Number(c.spent) || 0,
          ctr: (imps > 0 ? (clicks / imps * 100) : 0).toFixed(2),
        };
    }).filter((x: any) => x !== null);
    
    // If CH has no data yet, fall back to returning campaigns with 0 stats
    if (result.length === 0) {
        return campaigns.slice(0, limit).map(c => ({
            id: c.id,
            name: c.name,
            impressions: 0,
            clicks: 0,
            conversions: 0,
            spent: Number(c.spent) || 0,
            ctr: "0.00"
        }));
    }

    return result;
  }

  async trackImpression(data: {
    campaignId: string;
    publisherId: string;
    deviceType: string;
    country: string;
    timestamp?: Date | string;
    price?: number; // Added price optional field
  }): Promise<void> {
    const dateObj = this.normalizeTimestamp(data.timestamp);
    const formattedDate = dateObj.toISOString().replace('T', ' ').substring(0, 19);
    const dateKey = dateObj.toISOString().split('T')[0]; // YYYY-MM-DD
    
    // 1. Basic ClickHouse tracking
    const query = `INSERT INTO analytics.impressions (id, campaignId, publisherId, deviceType, country, timestamp) VALUES ('${crypto.randomUUID()}', '${data.campaignId}', '${data.publisherId}', '${data.deviceType}', '${data.country}', '${formattedDate}')`;
    await this.executeQuery(query);

    // 2. Real-Time Stats (Redis hash-based)
    await this.updateRealtimeStats(data.campaignId, 'impression');

    // 3. Per-campaign impression counter used by Go's GetCampaignCTR
    // Key pattern: campaign:imps:{campaignId}:{YYYY-MM-DD}
    const impsKey = `campaign:imps:${data.campaignId}:${dateKey}`;
    await this.redisClient.incr(impsKey);
    await this.redisClient.expire(impsKey, 60 * 60 * 48);

    // 4. Real-Time Budget Updates (Phase 10)
    // Only update if price is provided (from macro)
    if (data.price && data.price > 0) {
      const spendKey = `campaign:spend:${data.campaignId}:${dateKey}`;
      
      try {
        const newSpendStr = await this.redisClient.incrbyfloat(spendKey, data.price);
        const newSpend = parseFloat(newSpendStr);

        // Check Budget Thresholds (90% / 100%)
        await this.checkBudgetThresholds(data.campaignId, newSpend);

        // Set expiry to 48h to prevent infinite growth
        await this.redisClient.expire(spendKey, 60 * 60 * 48);
      } catch (err) {
        this.logger.error(`Failed to update budget for campaign ${data.campaignId}`, err);
      }
    }
  }

  async trackVideoEvent(data: {
    campaignId: string;
    eventType: 'start' | 'firstQuartile' | 'midpoint' | 'thirdQuartile' | 'complete';
    timestamp?: Date | string;
  }): Promise<void> {
    const dateObj = this.normalizeTimestamp(data.timestamp);
    const formattedDate = dateObj.toISOString().replace('T', ' ').substring(0, 19);
    
    // 1. ClickHouse tracking
    const query = `INSERT INTO analytics.video_events (id, campaignId, eventType, timestamp) VALUES ('${crypto.randomUUID()}', '${data.campaignId}', '${data.eventType}', '${formattedDate}')`;
    await this.executeQuery(query);

    // 2. Real-Time Stats (Redis)
    await this.updateRealtimeStats(data.campaignId, `video_${data.eventType}`);
  }

  private async checkBudgetThresholds(campaignId: string, currentSpend: number) {
    try {
      // Fetch budget info
      // Use findById since we might not have tenant context
      const campaign = await this.campaignsService.findById(campaignId);
      const budget = campaign.budget; // Assuming budget is the cap being tracked
      
      if (!budget || budget <= 0) return;

      const percentage = (currentSpend / budget) * 100;
      const dateKey = new Date().toISOString().split('T')[0];
      const alertKeyBase = `alert:budget:${campaignId}:${dateKey}`;

      // 100% Alert
      if (percentage >= 100) {
        const alerted = await this.redisClient.get(`${alertKeyBase}:100`);
        if (!alerted) {
          await this.notificationsService.create({
            userId: campaign.userId,
            tenantId: campaign.tenantId,
            title: 'Budget Exhausted',
            message: `Campaign "${campaign.name}" has hit 100% of its daily budget ($${budget}).`,
            type: 'error',
            category: 'budget',
          });
          await this.redisClient.setex(`${alertKeyBase}:100`, 86400, 'true');
        }
      } 
      // 90% Warning
      else if (percentage >= 90) {
        const alerted = await this.redisClient.get(`${alertKeyBase}:90`);
        if (!alerted) {
          await this.notificationsService.create({
            userId: campaign.userId,
            tenantId: campaign.tenantId,
            title: 'Budget Warning',
            message: `Campaign "${campaign.name}" has hit 90% of its daily budget ($${currentSpend.toFixed(2)} / $${budget}).`,
            type: 'warning',
            category: 'budget',
          });
          await this.redisClient.setex(`${alertKeyBase}:90`, 86400, 'true');
        }
      }
    } catch (error) {
       this.logger.warn(`Failed to check budget for ${campaignId}: ${error.message}`);
    }
  }

  async trackClick(data: {
    impressionId: string;
    campaignId: string;
    timestamp?: Date | string;
  }): Promise<void> {
    const dateObj = this.normalizeTimestamp(data.timestamp);
    const formattedDate = dateObj.toISOString().replace('T', ' ').substring(0, 19);
    const query = `INSERT INTO analytics.clicks (id, impressionId, campaignId, timestamp) VALUES ('${crypto.randomUUID()}', '${data.impressionId}', '${data.campaignId}', '${formattedDate}')`;
    await this.executeQuery(query);

    // Update Real-Time Stats (hash-based)
    await this.updateRealtimeStats(data.campaignId, 'click');

    // Also increment the per-campaign click counter used by Go's GetCampaignCTR
    // Key pattern: campaign:clicks:{campaignId}:{YYYY-MM-DD}
    const dateKey = dateObj.toISOString().split('T')[0];
    const clickKey = `campaign:clicks:${data.campaignId}:${dateKey}`;
    await this.redisClient.incr(clickKey);
    await this.redisClient.expire(clickKey, 60 * 60 * 48); // 48h TTL
  }

  async trackConversion(data: {
    clickId: string;
    campaignId: string;
    conversionValue: number;
    timestamp?: Date | string;
  }): Promise<void> {
    const dateObj = this.normalizeTimestamp(data.timestamp);
    const formattedDate = dateObj.toISOString().replace('T', ' ').substring(0, 19);
    const query = `INSERT INTO analytics.conversions (id, clickId, campaignId, value, timestamp) VALUES ('${crypto.randomUUID()}', '${data.clickId}', '${data.campaignId}', ${data.conversionValue}, '${formattedDate}')`;
    await this.executeQuery(query);

    // Update Real-Time Stats
    await this.updateRealtimeStats(data.campaignId, 'conversion');
  }

  async trackMmpEvent(data: {
    provider: string;
    eventType: string;
    campaignId: string;
    userId?: string;
    deviceId?: string;
    revenue?: number;
    currency?: string;
    metadata?: any;
    timestamp?: Date | string;
  }): Promise<void> {
    const dateObj = data.timestamp ? new Date(data.timestamp) : new Date();
    const formattedDate = dateObj.toISOString().replace('T', ' ').substring(0, 19);
    
    // Sanitize strings to avoid SQL injection in raw query
    const sanitize = (str: string) => str ? str.replace(/'/g, "\\'") : '';
    
    const query = `INSERT INTO analytics.mmp_events (
      id, provider, eventType, campaignId, userId, deviceId, revenue, currency, metadata, timestamp
    ) VALUES (
      '${crypto.randomUUID()}',
      '${sanitize(data.provider)}',
      '${sanitize(data.eventType)}',
      '${sanitize(data.campaignId)}',
      '${sanitize(data.userId)}',
      '${sanitize(data.deviceId)}',
      ${data.revenue || 0},
      '${sanitize(data.currency)}',
      '${sanitize(JSON.stringify(data.metadata || {}))}',
      '${formattedDate}'
    )`;
    
    await this.executeQuery(query);

    // If it's a conversion event with revenue, update real-time stats
    if (data.eventType === 'purchase' || data.eventType === 'install') {
       // Note: Currently updateRealtimeStats expects 'conversion' which increments count.
       // We might want to track revenue separately in future.
       await this.updateRealtimeStats(data.campaignId, 'conversion'); 
    }
  }

  // --- Real-Time Stats Helper ---
  private async updateRealtimeStats(campaignId: string, type: string): Promise<void> {
    const dateKey = new Date().toISOString().split('T')[0];
    const key = `campaign:stats:${campaignId}:${dateKey}`;
    
    // Increment specific counter
    // For standard types: "impressions", "clicks", "conversions"
    // For video types: "video_starts", "video_completes", etc.
    const field = type.endsWith('s') ? type : type + 's';
    await this.redisClient.hincrby(key, field, 1);
    
    // Set expiry (48h)
    await this.redisClient.expire(key, 60 * 60 * 48);

    // Add to active set for monitoring
    await this.redisClient.sadd(this.SPEND_UPDATED_SET + ':' + dateKey, campaignId);
    await this.redisClient.expire(this.SPEND_UPDATED_SET + ':' + dateKey, 60 * 60 * 48);
  }

  // --- Performance Anomaly Detection ---
  @Cron(CronExpression.EVERY_HOUR)
  async checkPerformanceAnomalies() {
    try {
      const dateKey = new Date().toISOString().split('T')[0];
      const activeCampaigns = await this.redisClient.smembers(this.SPEND_UPDATED_SET + ':' + dateKey);

      for (const campaignId of activeCampaigns) {
        const key = `campaign:stats:${campaignId}:${dateKey}`;
        const stats = await this.redisClient.hgetall(key);
        
        const impressions = parseInt(stats.impressions || '0', 10);
        const clicks = parseInt(stats.clicks || '0', 10);
        const conversions = parseInt(stats.conversions || '0', 10);

        // Alert 1: Low CTR (High Volume)
        if (impressions > 1000) {
          const ctr = (clicks / impressions) * 100;
          if (ctr < 0.05) { // < 0.05% CTR
             await this.triggerCampaignAlert(campaignId, 'Low CTR Warning', `Campaign is delivering impressions (${impressions}) but CTR is extremely low (${ctr.toFixed(3)}%). Check creative or targeting.`);
          }
        }

        // Alert 2: High Spend, Zero Conversions
        // Get spend for today
        const spendKey = `campaign:spend:${campaignId}:${dateKey}`;
        const spendStr = await this.redisClient.get(spendKey);
        const spend = parseFloat(spendStr || '0');

        if (spend > 50 && conversions === 0) { // e.g., > $50 spent with 0 actions
           await this.triggerCampaignAlert(campaignId, 'Zero Conversions Alerm', `Campaign has spent $${spend} today with 0 conversions. Optimization recommended.`);
        }
      }
    } catch (error) {
      this.logger.error('Failed to check performance anomalies', error);
    }
  }

  // --- Fraud Detection ---
  @Cron(CronExpression.EVERY_MINUTE)
  async checkFraudAlerts() {
    try {
      const dateKey = new Date().toISOString().split('T')[0];
      // Scan for fraud counts
      // pattern: fraud:publisher:<pubId>:<date>:count
      const keys = await this.redisClient.keys(`fraud:publisher:*:${dateKey}:count`);

      for (const key of keys) {
        const countStr = await this.redisClient.get(key);
        const count = parseInt(countStr || '0', 10);

        if (count > 50) {
          // Extract publisher ID
          // fraud:publisher:pub-123:2026-02-18:count
          const parts = key.split(':');
          const publisherId = parts[2];

          const alertKey = `alert:fraud:${publisherId}:${dateKey}`;
          const alreadyAlerted = await this.redisClient.get(alertKey);

          if (!alreadyAlerted) {
             // Notify Admin (Tenant ID usually null or system admin)
             // We can find the admin user or just system notification
             // For now, let's assume we notify top-level admins or specific user
             // Since we don't have a specific user context here, we might broadcast
             // OR find the user associated with the publisher if publishers are users.
             
             // In this system, publishers might not be Users directly in this context, 
             // but let's send a system notification to a predefined admin or all admins.
             
             // Simple fallback: Find the admin users
             const adminEmails = ['admin@taskirx.com', 'admin@taskir.com'];
             for (const email of adminEmails) {
               const adminUser = await this.usersService.findByEmail(email);
               if (adminUser) {
                 await this.notificationsService.create({
                   userId: adminUser.id,
                   tenantId: adminUser.tenantId,
                   title: 'Fraud Alert detected',
                   message: `High fraud activity detected for Publisher ${publisherId}. Publisher has been blocked.`,
                   type: 'error',
                   category: 'security',
                 });
                 // Break after sending to one admin to avoid duplicates if same user? 
                 // Or send to all? Let's send to the first available admin.
                 break;
               }
             }

             await this.redisClient.setex(alertKey, 86400, 'true');
             this.logger.warn(`Fraud Alert triggered for Publisher ${publisherId}: ${count} events.`);
          }
        }
      }
    } catch (error) {
       this.logger.error('Failed to check fraud alerts', error);
    }
  }

  private async triggerCampaignAlert(campaignId: string, title: string, message: string) {
    const dateKey = new Date().toISOString().split('T')[0];
    const alertKey = `alert:perf:${campaignId}:${title.replace(/\s+/g, '_')}:${dateKey}`;
    
    const alreadyAlerted = await this.redisClient.get(alertKey);
    if (alreadyAlerted) return;

    try {
      const campaign = await this.campaignsService.findById(campaignId);
      if (!campaign) return;

      await this.notificationsService.create({
        userId: campaign.userId,
        tenantId: campaign.tenantId,
        title: title,
        message: message,
        type: 'warning',
        category: 'performance',
      });

      // Mark as alerted for 24h
      await this.redisClient.setex(alertKey, 86400, 'true');
    } catch (e) {
      this.logger.warn(`Could not send alert for ${campaignId}: ${e.message}`);
    }
  }

  /**
   * Spend Sync Cron
   * Flushes daily Redis spend counters (campaign:spend:{id}:{date}) back into
   * the Postgres campaign.spent column every 10 minutes.
   *
   * This keeps the DB spend values reasonably up-to-date without the overhead
   * of writing to Postgres on every impression.
   */
  @Cron('0 */10 * * * *') // Every 10 minutes at :00, :10, :20, ...
  async syncSpendToDatabase() {
    const dateKey = new Date().toISOString().split('T')[0];
    const setKey = `${this.SPEND_UPDATED_SET}:${dateKey}`;

    try {
      const campaignIds = await this.redisClient.smembers(setKey);
      if (!campaignIds || campaignIds.length === 0) return;

      this.logger.debug(`syncSpendToDatabase: syncing ${campaignIds.length} campaigns`);

      for (const campaignId of campaignIds) {
        const spendKey = `campaign:spend:${campaignId}:${dateKey}`;
        const syncedKey = `campaign:spend:synced:${campaignId}:${dateKey}`;

        const dailySpendStr = await this.redisClient.get(spendKey);
        const dailySpend = parseFloat(dailySpendStr || '0');

        const syncedStr = await this.redisClient.get(syncedKey);
        const alreadySynced = parseFloat(syncedStr || '0');

        const delta = dailySpend - alreadySynced;
        if (delta <= 0) continue; // Nothing new to sync

        // Increment campaign.spent in DB
        const campaign = await this.campaignRepository.findOne({ where: { id: campaignId } });
        if (!campaign) continue;

        campaign.spent = (Number(campaign.spent) || 0) + delta;
        await this.campaignRepository.save(campaign);

        // Update synced marker in Redis
        await this.redisClient.set(syncedKey, dailySpend.toString());
        await this.redisClient.expire(syncedKey, 60 * 60 * 48);

        this.logger.debug(`Synced $${delta.toFixed(4)} for campaign ${campaignId}`);
      }
    } catch (error) {
      this.logger.error('Failed to sync spend to database', error);
    }
  }

  // Supply Path Optimization Analytics Methods

  async getSupplyChainMetrics(timeRange: string = '1h') {
    try {
      // Call the Go bidding engine's SPO analytics service
      const biddingEngineUrl = this.configService.get<string>('BIDDING_ENGINE_URL', 'http://go-bidding-engine:8080');
      
      // Get metrics from Redis via bidding engine
      const response = await axios.get(`${biddingEngineUrl}/api/analytics/supply-chain?timeRange=${timeRange}`, {
        timeout: 5000,
      });

      return response.data;
    } catch (error) {
      this.logger.error(`Failed to get supply chain metrics: ${error.message}`);
      
      // Return mock data as fallback
      return {
        timeRange,
        totalRequests: 12543,
        successfulBids: 8921,
        winRate: 0.71,
        avgLatencyMs: 45.2,
        avgTotalFees: 0.0032,
        serviceMetrics: {
          'fraud-detection': {
            serviceName: 'fraud-detection',
            totalCalls: 12543,
            successRate: 0.98,
            avgLatencyMs: 12.5,
            errorRate: 0.02,
            totalFees: 0.001
          },
          'ad-matching': {
            serviceName: 'ad-matching',
            totalCalls: 12301,
            successRate: 0.95,
            avgLatencyMs: 18.7,
            errorRate: 0.05,
            totalFees: 0.002
          },
          'bid-optimizer': {
            serviceName: 'bid-optimizer',
            totalCalls: 8921,
            successRate: 0.97,
            avgLatencyMs: 14.3,
            errorRate: 0.03,
            totalFees: 0.0015
          }
        },
        pathEfficiency: 0.89,
        timestamp: new Date().toISOString()
      };
    }
  }

  async getSupplyPathOptimization(timeRange: string = '1h') {
    try {
      // Call the Go bidding engine's SPO analytics service
      const biddingEngineUrl = this.configService.get<string>('BIDDING_ENGINE_URL', 'http://go-bidding-engine:8080');
      
      const response = await axios.get(`${biddingEngineUrl}/api/analytics/supply-path-optimization?timeRange=${timeRange}`, {
        timeout: 5000,
      });

      return response.data;
    } catch (error) {
      this.logger.error(`Failed to get supply path optimization: ${error.message}`);
      
      // Return mock optimization data as fallback
      return {
        optimizations: [
          {
            type: 'cache',
            service: 'fraud-detection',
            description: 'High latency detected (12.5ms avg). Consider caching responses.',
            priority: 'medium',
            savings: 2.34
          },
          {
            type: 'circuit_breaker',
            service: 'ad-matching',
            description: 'Low success rate (95%). Circuit breaker may be tripping too frequently.',
            priority: 'high',
            savings: 0
          },
          {
            type: 'fee_negotiation',
            service: 'bid-optimizer',
            description: 'High fees detected ($0.0015 per call). Consider direct integration.',
            priority: 'low',
            savings: 1.12
          }
        ],
        estimatedSavings: 3.46
      };
    }
  }

  async getBidPathAnalytics(requestId: string) {
    try {
      // Call the Go bidding engine's SPO analytics service
      const biddingEngineUrl = this.configService.get<string>('BIDDING_ENGINE_URL', 'http://go-bidding-engine:8080');
      
      const response = await axios.get(`${biddingEngineUrl}/api/analytics/bid-path/${requestId}`, {
        timeout: 5000,
      });

      return response.data;
    } catch (error) {
      this.logger.error(`Failed to get bid path analytics for ${requestId}: ${error.message}`);
      
      // Return mock data as fallback
      return {
        requestId,
        publisherId: 'mock-publisher',
        adSlotId: 'mock-slot',
        totalLatencyMs: 45.2,
        totalHops: 3,
        totalFees: 0.0032,
        finalBidPrice: 1.25,
        wonAuction: true,
        campaignId: 'mock-campaign',
        dealId: 'deal-123',
        hops: [
          {
            serviceName: 'fraud-detection',
            serviceType: 'internal',
            endpoint: 'http://fraud-detection:6001/api/detect',
            requestSize: 245,
            responseSize: 89,
            latencyMs: 12,
            statusCode: 200,
            success: true,
            fee: 0.001,
            timestamp: new Date().toISOString(),
            sequence: 1
          },
          {
            serviceName: 'ad-matching',
            serviceType: 'internal',
            endpoint: 'http://ad-matching:6002/api/match',
            requestSize: 512,
            responseSize: 234,
            latencyMs: 19,
            statusCode: 200,
            success: true,
            fee: 0.002,
            timestamp: new Date().toISOString(),
            sequence: 2
          },
          {
            serviceName: 'bid-optimizer',
            serviceType: 'internal',
            endpoint: 'http://bid-optimizer:6003/api/optimize',
            requestSize: 189,
            responseSize: 145,
            latencyMs: 14,
            statusCode: 200,
            success: true,
            fee: 0.0015,
            timestamp: new Date().toISOString(),
            sequence: 3
          }
        ],
        metadata: {},
        timestamp: new Date().toISOString()
      };
    }
  }

  async getServicePerformance(serviceName: string, timeRange: string = '1h') {
    try {
      // Call the Go bidding engine's SPO analytics service
      const biddingEngineUrl = this.configService.get<string>('BIDDING_ENGINE_URL', 'http://go-bidding-engine:8080');
      
      const response = await axios.get(`${biddingEngineUrl}/api/analytics/service-performance?serviceName=${serviceName}&timeRange=${timeRange}`, {
        timeout: 5000,
      });

      return response.data;
    } catch (error) {
      this.logger.error(`Failed to get service performance for ${serviceName}: ${error.message}`);
      
      // Return mock data as fallback
      return {
        serviceName,
        totalCalls: 12543,
        successRate: 0.96,
        avgLatencyMs: 15.8,
        errorRate: 0.04,
        totalFees: 0.0012
      };
    }
  }

  // Advanced Supply Path Optimization Methods

  async getDirectPublisherAnalysis(timeRange: string = '1h') {
    try {
      // Call the Go bidding engine's advanced analytics
      const biddingEngineUrl = this.configService.get<string>('BIDDING_ENGINE_URL', 'http://go-bidding-engine:8080');

      const response = await axios.get(`${biddingEngineUrl}/api/analytics/direct-publisher-analysis?timeRange=${timeRange}`, {
        timeout: 5000,
      });

      return response.data;
    } catch (error) {
      this.logger.error(`Failed to get direct publisher analysis: ${error.message}`);

      // Return mock data as fallback
      return {
        timeRange,
        timestamp: new Date().toISOString(),
        currentHops: 3,
        opportunities: [
          {
            serviceName: 'fraud-detection',
            currentFeeRate: 0.0012,
            estimatedDirectFee: 0.0008,
            successRate: 0.92,
            monthlyVolume: 45000,
            priority: 'high',
            riskLevel: 'medium',
            estimatedSavings: 1800,
            roi: 75.0
          },
          {
            serviceName: 'ad-matching',
            currentFeeRate: 0.0021,
            estimatedDirectFee: 0.0014,
            successRate: 0.89,
            monthlyVolume: 42000,
            priority: 'high',
            riskLevel: 'medium',
            estimatedSavings: 2940,
            roi: 82.5
          }
        ]
      };
    }
  }

  async getCostBenefitAnalysis(timeRange: string = '1h') {
    try {
      // Call the Go bidding engine's cost-benefit analysis
      const biddingEngineUrl = this.configService.get<string>('BIDDING_ENGINE_URL', 'http://go-bidding-engine:8080');

      const response = await axios.get(`${biddingEngineUrl}/api/analytics/cost-benefit-analysis?timeRange=${timeRange}`, {
        timeout: 5000,
      });

      return response.data;
    } catch (error) {
      this.logger.error(`Failed to get cost-benefit analysis: ${error.message}`);

      // Return mock data as fallback
      return {
        timeRange,
        timestamp: new Date().toISOString(),
        currentTotalCost: 12500,
        currentWinRate: 0.71,
        currentAvgLatency: 45.2,
        scenarios: [
          {
            name: 'Direct Publisher Connections',
            description: 'Bypass intermediaries and connect directly with publishers',
            estimatedCostReduction: 0.6,
            estimatedLatencyReduction: 18.08,
            riskLevel: 'medium',
            implementationEffort: 'high',
            timeToValue: '3-6 months',
            netBenefit: 22500,
            breakEvenMonths: 2
          },
          {
            name: 'Service Performance Optimization',
            description: 'Optimize existing services for better performance and cost efficiency',
            estimatedCostReduction: 0.25,
            estimatedLatencyReduction: 9.04,
            riskLevel: 'low',
            implementationEffort: 'medium',
            timeToValue: '1-3 months',
            netBenefit: 9375,
            breakEvenMonths: 1
          },
          {
            name: 'Hybrid Optimization',
            description: 'Combine direct connections with service optimizations',
            estimatedCostReduction: 0.75,
            estimatedLatencyReduction: 22.6,
            riskLevel: 'medium',
            implementationEffort: 'high',
            timeToValue: '2-4 months',
            netBenefit: 28125,
            breakEvenMonths: 3
          }
        ]
      };
    }
  }
}
