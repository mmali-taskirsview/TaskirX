/**
 * Bidding Service
 * Handles bidding operations and recommendations
 */

import { RequestManager } from './RequestManager';
import { Logger } from '../utils/Logger';
import { Bid } from '../types';

export class BiddingService {
  constructor(private requestManager: RequestManager, private logger: Logger) {}

  async submitBid(bid: Omit<Bid, 'id' | 'createdAt'>): Promise<Bid> {
    this.logger.debug('Submitting bid');
    return this.requestManager.post<Bid>('/api/bids', bid);
  }

  async getRecommendations(campaignId: string): Promise<any> {
    this.logger.debug('Fetching bid recommendations for campaign:', campaignId);
    return this.requestManager.get(`/api/bids/campaigns/${campaignId}/recommendations`);
  }

  async getBids(campaignId?: string): Promise<Bid[]> {
    this.logger.debug('Fetching bids', campaignId ? `for campaign: ${campaignId}` : '');
    const endpoint = campaignId ? `/api/bids/campaigns/${campaignId}` : '/api/bids';
    return this.requestManager.get<Bid[]>(endpoint);
  }

  async getBid(bidId: string): Promise<Bid> {
    this.logger.debug('Fetching bid:', bidId);
    return this.requestManager.get<Bid>(`/api/bids/${bidId}`);
  }

  async getStats(campaignId: string): Promise<any> {
    this.logger.debug('Fetching bid statistics for campaign:', campaignId);
    return this.requestManager.get(`/api/bids/campaigns/${campaignId}/stats`);
  }
}
