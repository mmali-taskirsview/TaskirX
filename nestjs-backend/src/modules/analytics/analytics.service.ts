import * as crypto from 'crypto';
import axios from 'axios';
import { Injectable, OnModuleInit, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Campaign, CampaignStatus } from '../campaigns/campaign.entity';

@Injectable()
export class AnalyticsService implements OnModuleInit {
  private readonly logger = new Logger(AnalyticsService.name);
  private clickhouseUrl: string;
  private clickhouseUser: string;
  private clickhousePassword: string;

  constructor(
    private configService: ConfigService,
    @InjectRepository(Campaign)
    private campaignRepository: Repository<Campaign>,
  ) {
    const clickhouseUrl = this.configService.get<string>('CLICKHOUSE_URL');
    const clickhouseHost = this.configService.get<string>('CLICKHOUSE_HOST', 'clickhouse');
    const clickhousePort = this.configService.get<string>('CLICKHOUSE_PORT', '8123');
    this.clickhouseUrl = clickhouseUrl ?? `http://${clickhouseHost}:${clickhousePort}`;
    this.clickhouseUser = this.configService.get<string>('CLICKHOUSE_USERNAME', 'taskir');
    this.clickhousePassword = this.configService.get<string>('CLICKHOUSE_PASSWORD', 'clickhouse_password_2026');
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
        ) ENGINE = MergeTree() ORDER BY (campaignId, timestamp)`
      ];

      for (const query of queries) {
        await this.executeQuery(query);
      }
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
    
    for (const campaign of campaigns) {
      totalSpend += Number(campaign.spent) || 0;
      if (campaign.status === 'active') {
        activeCampaigns++;
      }
    }
    
    return {
      totalImpressions,
      totalClicks,
      totalConversions,
      totalSpend: totalSpend.toFixed(2),
      activeCampaigns,
      avgCtr: totalImpressions > 0 ? (totalClicks / totalImpressions * 100).toFixed(2) : 0,
      avgCpc: totalClicks > 0 ? (totalSpend / totalClicks).toFixed(2) : 0,
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
    timestamp: Date | string;
  }): Promise<void> {
    const dateObj = new Date(data.timestamp);
    const formattedDate = dateObj.toISOString().replace('T', ' ').substring(0, 19);
    const query = `INSERT INTO analytics.impressions (id, campaignId, publisherId, deviceType, country, timestamp) VALUES ('${crypto.randomUUID()}', '${data.campaignId}', '${data.publisherId}', '${data.deviceType}', '${data.country}', '${formattedDate}')`;
    await this.executeQuery(query);
  }

  async trackClick(data: {
    impressionId: string;
    campaignId: string;
    timestamp: Date | string;
  }): Promise<void> {
    const dateObj = new Date(data.timestamp);
    const formattedDate = dateObj.toISOString().replace('T', ' ').substring(0, 19);
    const query = `INSERT INTO analytics.clicks (id, impressionId, campaignId, timestamp) VALUES ('${crypto.randomUUID()}', '${data.impressionId}', '${data.campaignId}', '${formattedDate}')`;
    await this.executeQuery(query);
  }

  async trackConversion(data: {
    clickId: string;
    campaignId: string;
    conversionValue: number;
    timestamp: Date | string;
  }): Promise<void> {
    const dateObj = new Date(data.timestamp);
    const formattedDate = dateObj.toISOString().replace('T', ' ').substring(0, 19);
    const query = `INSERT INTO analytics.conversions (id, clickId, campaignId, value, timestamp) VALUES ('${crypto.randomUUID()}', '${data.clickId}', '${data.campaignId}', ${data.conversionValue}, '${formattedDate}')`;
    await this.executeQuery(query);
  }
}
