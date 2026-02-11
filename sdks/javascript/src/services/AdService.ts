/**
 * Ad Service
 * Handles ad placement and management operations
 */

import { RequestManager } from './RequestManager';
import { Logger } from '../utils/Logger';

export interface Ad {
  id: string;
  campaignId: string;
  placement: string;
  creativeUrl: string;
  clickUrl: string;
  width: number;
  height: number;
  status: 'active' | 'paused' | 'archived';
  createdAt: Date;
  updatedAt: Date;
}

export class AdService {
  constructor(private requestManager: RequestManager, private logger: Logger) {}

  async create(ad: Omit<Ad, 'id' | 'createdAt' | 'updatedAt'>): Promise<Ad> {
    this.logger.debug('Creating ad placement');
    return this.requestManager.post<Ad>('/api/ads', ad);
  }

  async list(campaignId?: string): Promise<Ad[]> {
    this.logger.debug('Listing ads', campaignId ? `for campaign: ${campaignId}` : '');
    const endpoint = campaignId ? `/api/ads/campaigns/${campaignId}` : '/api/ads';
    return this.requestManager.get<Ad[]>(endpoint);
  }

  async get(adId: string): Promise<Ad> {
    this.logger.debug('Getting ad:', adId);
    return this.requestManager.get<Ad>(`/api/ads/${adId}`);
  }

  async update(adId: string, updates: Partial<Ad>): Promise<Ad> {
    this.logger.debug('Updating ad:', adId);
    return this.requestManager.put<Ad>(`/api/ads/${adId}`, updates);
  }

  async delete(adId: string): Promise<void> {
    this.logger.debug('Deleting ad:', adId);
    await this.requestManager.delete(`/api/ads/${adId}`);
  }
}
