/**
 * Campaign Service
 * Handles campaign management operations
 */

import { RequestManager } from './RequestManager';
import { Logger } from '../utils/Logger';
import { Campaign } from '../types';

export class CampaignService {
  constructor(private requestManager: RequestManager, private logger: Logger) {}

  async create(data: Partial<Campaign>): Promise<Campaign> {
    this.logger.debug('Creating campaign:', data.name);
    return this.requestManager.post<Campaign>('/api/campaigns', data);
  }

  async list(filters?: any): Promise<Campaign[]> {
    this.logger.debug('Fetching campaigns');
    const response = await this.requestManager.get<any>('/api/campaigns', { params: filters });
    return response.campaigns || [];
  }

  async get(campaignId: string): Promise<Campaign> {
    this.logger.debug('Fetching campaign:', campaignId);
    return this.requestManager.get<Campaign>(`/api/campaigns/${campaignId}`);
  }

  async update(campaignId: string, data: Partial<Campaign>): Promise<Campaign> {
    this.logger.debug('Updating campaign:', campaignId);
    return this.requestManager.put<Campaign>(`/api/campaigns/${campaignId}`, data);
  }

  async delete(campaignId: string): Promise<any> {
    this.logger.debug('Deleting campaign:', campaignId);
    return this.requestManager.delete(`/api/campaigns/${campaignId}`);
  }

  async pause(campaignId: string): Promise<Campaign> {
    this.logger.debug('Pausing campaign:', campaignId);
    return this.requestManager.post<Campaign>(`/api/campaigns/${campaignId}/pause`, {});
  }

  async resume(campaignId: string): Promise<Campaign> {
    this.logger.debug('Resuming campaign:', campaignId);
    return this.requestManager.post<Campaign>(`/api/campaigns/${campaignId}/resume`, {});
  }
}
