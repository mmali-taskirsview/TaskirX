/**
 * Webhook Service
 * Handles webhook subscriptions and event management
 */

import { RequestManager } from './RequestManager';
import { Logger } from '../utils/Logger';
import { Webhook, WebhookEvent } from '../types';

export interface WebhookLog {
  id: string;
  webhookId: string;
  eventType: string;
  payload: any;
  status: number;
  response: string;
  attemptCount: number;
  lastAttemptAt: Date;
  nextRetryAt?: Date;
  createdAt: Date;
}

export class WebhookService {
  private eventHandlers: Map<string, (event: WebhookEvent) => void> = new Map();

  constructor(private requestManager: RequestManager, private logger: Logger) {}

  async subscribe(webhook: Omit<Webhook, 'id' | 'createdAt' | 'active'>): Promise<Webhook> {
    this.logger.debug('Creating webhook subscription');
    return this.requestManager.post<Webhook>('/api/webhooks', webhook);
  }

  async list(): Promise<Webhook[]> {
    this.logger.debug('Listing webhooks');
    return this.requestManager.get<Webhook[]>('/api/webhooks');
  }

  async get(webhookId: string): Promise<Webhook> {
    this.logger.debug('Getting webhook:', webhookId);
    return this.requestManager.get<Webhook>(`/api/webhooks/${webhookId}`);
  }

  async update(webhookId: string, updates: Partial<Webhook>): Promise<Webhook> {
    this.logger.debug('Updating webhook:', webhookId);
    return this.requestManager.put<Webhook>(`/api/webhooks/${webhookId}`, updates);
  }

  async delete(webhookId: string): Promise<void> {
    this.logger.debug('Deleting webhook:', webhookId);
    await this.requestManager.delete(`/api/webhooks/${webhookId}`);
  }

  async test(webhookId: string): Promise<any> {
    this.logger.debug('Testing webhook:', webhookId);
    return this.requestManager.post(`/api/webhooks/${webhookId}/test`, {});
  }

  async getLogs(webhookId: string, limit?: number): Promise<WebhookLog[]> {
    this.logger.debug('Fetching webhook logs:', webhookId);
    const params = limit ? { limit } : {};
    return this.requestManager.get<WebhookLog[]>(`/api/webhooks/${webhookId}/logs`, { params });
  }

  onEvent(eventType: string, handler: (event: WebhookEvent) => void): void {
    this.logger.debug('Registering event handler for:', eventType);
    this.eventHandlers.set(eventType, handler);
  }

  offEvent(eventType: string): void {
    this.logger.debug('Removing event handler for:', eventType);
    this.eventHandlers.delete(eventType);
  }

  handleEvent(event: WebhookEvent): void {
    const handler = this.eventHandlers.get(event.type);
    if (handler) {
      this.logger.debug('Handling event:', event.type);
      handler(event);
    }
  }
}
