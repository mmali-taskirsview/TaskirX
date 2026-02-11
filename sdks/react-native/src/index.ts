// Main TaskirX Client for React Native

import { RequestManager } from './network/RequestManager';
import {
  AuthService,
  CampaignService,
  AnalyticsService,
  BiddingService,
  AdService,
  WebhookService,
} from './services/Services';
import type { ClientConfig, Result, TaskirXError } from './types';

export class TaskirXClient {
  private requestManager: RequestManager;

  public auth: AuthService;
  public campaigns: CampaignService;
  public analytics: AnalyticsService;
  public bidding: BiddingService;
  public ads: AdService;
  public webhooks: WebhookService;

  private constructor(config: ClientConfig) {
    this.requestManager = new RequestManager(config);

    this.auth = new AuthService(this.requestManager);
    this.campaigns = new CampaignService(this.requestManager);
    this.analytics = new AnalyticsService(this.requestManager);
    this.bidding = new BiddingService(this.requestManager);
    this.ads = new AdService(this.requestManager);
    this.webhooks = new WebhookService(this.requestManager);
  }

  public static create(apiUrl: string, apiKey: string, debug: boolean = false): TaskirXClient {
    const config: ClientConfig = { apiUrl, apiKey, debug };
    return new TaskirXClient(config);
  }

  // Health & Status
  public async getHealth(): Promise<Result<Record<string, any>>> {
    try {
      const health = await this.requestManager.get<Record<string, any>>('/health');
      return { success: true, data: health };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public async getStatus(): Promise<Result<Record<string, any>>> {
    try {
      const status = await this.requestManager.get<Record<string, any>>('/status');
      return { success: true, data: status };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  // Profile Operations
  public async getProfile(): Promise<Result<any>> {
    try {
      const user = await this.auth.getProfile();
      return { success: true, data: user };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public async logout(): Promise<Result<void>> {
    try {
      await this.auth.logout();
      return { success: true, data: undefined };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  // Dashboard & Analytics
  public async getDashboard(): Promise<Result<Record<string, any>>> {
    try {
      const dashboard = await this.analytics.dashboard();
      return { success: true, data: dashboard };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public async getCampaignPerformance(campaignId: string): Promise<Result<any>> {
    try {
      const analytics = await this.analytics.campaign(campaignId);
      return { success: true, data: analytics };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public async getRealtimeAnalytics(): Promise<Result<any>> {
    try {
      const analytics = await this.analytics.realtime();
      return { success: true, data: analytics };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  // Campaign Operations
  public async createCampaign(
    name: string,
    budget: number,
    startDate: string,
    endDate: string,
    targetAudience: Record<string, any>
  ): Promise<Result<any>> {
    try {
      const campaign = await this.campaigns.create(name, budget, startDate, endDate, targetAudience);
      return { success: true, data: campaign };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public async getCampaigns(limit: number = 50, offset: number = 0): Promise<Result<any[]>> {
    try {
      const campaigns = await this.campaigns.list(limit, offset);
      return { success: true, data: campaigns };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public async getCampaign(id: string): Promise<Result<any>> {
    try {
      const campaign = await this.campaigns.get(id);
      return { success: true, data: campaign };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public async pauseCampaign(id: string): Promise<Result<any>> {
    try {
      const campaign = await this.campaigns.pause(id);
      return { success: true, data: campaign };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public async resumeCampaign(id: string): Promise<Result<any>> {
    try {
      const campaign = await this.campaigns.resume(id);
      return { success: true, data: campaign };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  // Bidding Operations
  public async submitBid(
    campaignId: string,
    adSlotId: string,
    amount: number,
    currency: string = 'USD'
  ): Promise<Result<any>> {
    try {
      const bid = await this.bidding.submitBid(campaignId, adSlotId, amount, currency);
      return { success: true, data: bid };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public async getBidRecommendations(): Promise<Result<any[]>> {
    try {
      const recommendations = await this.bidding.recommendations();
      return { success: true, data: recommendations };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public async getBidStatistics(): Promise<Result<Record<string, any>>> {
    try {
      const stats = await this.bidding.stats();
      return { success: true, data: stats };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  // Ad Operations
  public async createAd(
    campaignId: string,
    placement: string,
    imageUrl: string,
    clickUrl: string,
    dimensions: string
  ): Promise<Result<any>> {
    try {
      const ad = await this.ads.create(campaignId, placement, imageUrl, clickUrl, dimensions);
      return { success: true, data: ad };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public async getAds(campaignId: string, limit: number = 50): Promise<Result<any[]>> {
    try {
      const ads = await this.ads.list(campaignId, limit);
      return { success: true, data: ads };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  // Webhook Operations
  public async subscribeWebhook(url: string, events: string[]): Promise<Result<any>> {
    try {
      const webhook = await this.webhooks.subscribe(url, events);
      return { success: true, data: webhook };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public async getWebhooks(limit: number = 50): Promise<Result<any[]>> {
    try {
      const webhooks = await this.webhooks.list(limit);
      return { success: true, data: webhooks };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public async testWebhook(id: string): Promise<Result<Record<string, any>>> {
    try {
      const result = await this.webhooks.test(id);
      return { success: true, data: result };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  public onWebhookEvent(type: string, handler: (event: any) => void): void {
    this.webhooks.onEvent(type, handler);
  }

  // Batch Operations
  public async getStatistics(): Promise<Result<Record<string, any>>> {
    try {
      const campaigns = await this.campaigns.list(1000);
      const analytics = await this.analytics.realtime();
      const bids = await this.bidding.list(1000);

      const stats = {
        campaignCount: campaigns.length,
        analytics,
        bidCount: bids.length,
      };

      return { success: true, data: stats };
    } catch (error) {
      return { success: false, error: error as TaskirXError };
    }
  }

  // Debug
  public enableDebug(enabled: boolean): void {
    if (enabled) {
      console.log('[TaskirX] 🐛 Debug mode enabled');
    }
  }
}

export * from './types';
export * from './services/Services';
export { RequestManager } from './network/RequestManager';
