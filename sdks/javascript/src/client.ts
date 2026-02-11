/**
 * TaskirX Client
 * Main SDK client combining all services
 */

import { RequestManager } from './services/RequestManager';
import { AuthService } from './services/AuthService';
import { CampaignService } from './services/CampaignService';
import { AnalyticsService } from './services/AnalyticsService';
import { BiddingService } from './services/BiddingService';
import { AdService } from './services/AdService';
import { WebhookService } from './services/WebhookService';
import { Logger } from './utils/Logger';
import { ClientConfig } from './types';

export class TaskirXClient {
  private requestManager: RequestManager;
  private logger: Logger;
  public auth: AuthService;
  public campaigns: CampaignService;
  public analytics: AnalyticsService;
  public bidding: BiddingService;
  public ads: AdService;
  public webhooks: WebhookService;

  constructor(config: ClientConfig) {
    // Initialize logger
    this.logger = new Logger(config.debug || false);
    this.logger.info('Initializing TaskirX Client', config);

    // Initialize request manager
    this.requestManager = new RequestManager(config, this.logger);

    // Initialize services
    this.auth = new AuthService(this.requestManager, this.logger);
    this.campaigns = new CampaignService(this.requestManager, this.logger);
    this.analytics = new AnalyticsService(this.requestManager, this.logger);
    this.bidding = new BiddingService(this.requestManager, this.logger);
    this.ads = new AdService(this.requestManager, this.logger);
    this.webhooks = new WebhookService(this.requestManager, this.logger);
  }

  /**
   * Check platform health and connectivity
   */
  async getHealth(): Promise<any> {
    try {
      this.logger.debug('Checking platform health');
      const response = await this.requestManager.get('/api/health');
      this.logger.info('Platform health check passed');
      return response;
    } catch (error: any) {
      this.logger.error('Health check failed:', error);
      throw error;
    }
  }

  /**
   * Get platform status information
   */
  async getStatus(): Promise<any> {
    try {
      this.logger.debug('Fetching platform status');
      const response = await this.requestManager.get('/api/status');
      this.logger.info('Platform status retrieved');
      return response;
    } catch (error: any) {
      this.logger.error('Status check failed:', error);
      throw error;
    }
  }

  /**
   * Get API version
   */
  async getVersion(): Promise<string> {
    try {
      this.logger.debug('Fetching API version');
      const response = await this.requestManager.get<{ version: string }>('/api/version');
      return response.version;
    } catch (error: any) {
      this.logger.error('Version check failed:', error);
      throw error;
    }
  }

  /**
   * Initialize platform authentication
   */
  async initialize(): Promise<void> {
    try {
      this.logger.info('Initializing TaskirX platform');
      await this.getHealth();
      this.logger.info('TaskirX platform initialized successfully');
    } catch (error) {
      this.logger.error('Initialization failed:', error);
      throw error;
    }
  }

  /**
   * Get current user profile
   */
  async getProfile(): Promise<any> {
    return this.auth.getProfile();
  }

  /**
   * Logout and cleanup
   */
  async logout(): Promise<void> {
    try {
      this.logger.info('Logging out from TaskirX');
      await this.auth.logout();
      this.logger.info('Successfully logged out');
    } catch (error) {
      this.logger.error('Logout failed:', error);
      throw error;
    }
  }

  /**
   * Set API key for authentication
   */
  setApiKey(apiKey: string): void {
    this.logger.debug('Setting API key');
    this.requestManager.setToken(apiKey);
  }

  /**
   * Get comprehensive dashboard data
   */
  async getDashboard(): Promise<any> {
    try {
      this.logger.debug('Fetching comprehensive dashboard data');
      const [realtimeAnalytics, campaigns, webhooks] = await Promise.all([
        this.analytics.getRealtime(),
        this.campaigns.list(),
        this.webhooks.list(),
      ]);

      return {
        analytics: realtimeAnalytics,
        campaigns,
        webhooks,
        timestamp: new Date().toISOString(),
      };
    } catch (error: any) {
      this.logger.error('Dashboard fetch failed:', error);
      throw error;
    }
  }

  /**
   * Batch create campaigns
   */
  async createCampaigns(campaigns: any[]): Promise<any[]> {
    try {
      this.logger.debug('Creating batch campaigns:', campaigns.length);
      const results = await Promise.all(
        campaigns.map((campaign) => this.campaigns.create(campaign))
      );
      this.logger.info('Batch campaign creation completed');
      return results;
    } catch (error) {
      this.logger.error('Batch campaign creation failed:', error);
      throw error;
    }
  }

  /**
   * Get campaign performance summary
   */
  async getCampaignPerformance(campaignId: string): Promise<any> {
    try {
      this.logger.debug('Fetching campaign performance:', campaignId);
      const [campaign, analytics, stats] = await Promise.all([
        this.campaigns.get(campaignId),
        this.analytics.getCampaignAnalytics(campaignId),
        this.bidding.getStats(campaignId),
      ]);

      return {
        campaign,
        analytics,
        bidding: stats,
      };
    } catch (error: any) {
      this.logger.error('Campaign performance fetch failed:', error);
      throw error;
    }
  }

  /**
   * Enable debug mode for detailed logging
   */
  enableDebug(enabled: boolean = true): void {
    this.logger.setDebug(enabled);
    this.logger.info(`Debug mode ${enabled ? 'enabled' : 'disabled'}`);
  }

  /**
   * Get complete platform statistics
   */
  async getStatistics(): Promise<any> {
    try {
      this.logger.debug('Fetching platform statistics');
      const [dashboard, realtime] = await Promise.all([
        this.getDashboard(),
        this.analytics.getRealtime(),
      ]);

      return {
        dashboard,
        realtime,
        timestamp: new Date().toISOString(),
      };
    } catch (error: any) {
      this.logger.error('Statistics fetch failed:', error);
      throw error;
    }
  }
}

export default TaskirXClient;
