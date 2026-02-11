/**
 * Analytics Service
 * Handles analytics and reporting operations
 */

import { RequestManager } from './RequestManager';
import { Logger } from '../utils/Logger';
import { Analytics } from '../types';

export class AnalyticsService {
  constructor(private requestManager: RequestManager, private logger: Logger) {}

  async getRealtime(): Promise<Analytics> {
    this.logger.debug('Fetching real-time analytics');
    return this.requestManager.get<Analytics>('/api/analytics/realtime');
  }

  async getCampaignAnalytics(
    campaignId: string,
    dateRange?: { startDate: string; endDate: string }
  ): Promise<Analytics> {
    this.logger.debug('Fetching campaign analytics:', campaignId);
    const params = dateRange ? { start_date: dateRange.startDate, end_date: dateRange.endDate } : {};
    return this.requestManager.get<Analytics>(`/api/analytics/campaigns/${campaignId}`, { params });
  }

  async getBreakdown(
    campaignId: string,
    breakdownType: 'device' | 'geo' | 'browser' | 'os'
  ): Promise<any> {
    this.logger.debug('Fetching analytics breakdown:', breakdownType);
    return this.requestManager.get(`/api/analytics/campaigns/${campaignId}/breakdown/${breakdownType}`);
  }

  async getDashboard(): Promise<any> {
    this.logger.debug('Fetching dashboard analytics');
    return this.requestManager.get('/api/analytics/dashboard');
  }
}
