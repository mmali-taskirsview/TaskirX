// Services for TaskirX React Native SDK

import { RequestManager } from '../network/RequestManager';
import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  User,
  Campaign,
  CampaignCreateRequest,
  Bid,
  BidSubmitRequest,
  Analytics,
  Ad,
  AdCreateRequest,
  Webhook,
  WebhookCreateRequest,
  WebhookEvent,
} from '../types';

// Auth Service
export class AuthService {
  private requestManager: RequestManager;
  private currentUser: User | null = null;

  constructor(requestManager: RequestManager) {
    this.requestManager = requestManager;
  }

  async register(email: string, password: string, name: string, company?: string): Promise<AuthResponse> {
    const request: RegisterRequest = { email, password, name, company };
    const response = await this.requestManager.post<AuthResponse>('/auth/register', request);
    this.currentUser = response.user;
    this.requestManager.setAuthToken(response.token);
    return response;
  }

  async login(email: string, password: string): Promise<AuthResponse> {
    const request: LoginRequest = { email, password };
    const response = await this.requestManager.post<AuthResponse>('/auth/login', request);
    this.currentUser = response.user;
    this.requestManager.setAuthToken(response.token);
    return response;
  }

  async logout(): Promise<void> {
    this.requestManager.clearAuthToken();
    this.currentUser = null;
  }

  async getProfile(): Promise<User> {
    const user = await this.requestManager.get<User>('/auth/profile');
    this.currentUser = user;
    return user;
  }

  async refreshToken(refreshToken: string): Promise<AuthResponse> {
    const response = await this.requestManager.post<AuthResponse>('/auth/refresh', { refreshToken });
    this.requestManager.setAuthToken(response.token);
    return response;
  }
}

// Campaign Service
export class CampaignService {
  private requestManager: RequestManager;

  constructor(requestManager: RequestManager) {
    this.requestManager = requestManager;
  }

  async create(name: string, budget: number, startDate: string, endDate: string, targetAudience: Record<string, any>): Promise<Campaign> {
    const request: CampaignCreateRequest = { name, budget, startDate, endDate, targetAudience };
    return this.requestManager.post<Campaign>('/campaigns', request);
  }

  async list(limit: number = 50, offset: number = 0): Promise<Campaign[]> {
    return this.requestManager.get<Campaign[]>(`/campaigns?limit=${limit}&offset=${offset}`);
  }

  async get(id: string): Promise<Campaign> {
    return this.requestManager.get<Campaign>(`/campaigns/${id}`);
  }

  async update(id: string, name?: string, budget?: number): Promise<Campaign> {
    const updates: Record<string, any> = {};
    if (name) updates.name = name;
    if (budget) updates.budget = budget;
    return this.requestManager.put<Campaign>(`/campaigns/${id}`, updates);
  }

  async delete(id: string): Promise<Record<string, any>> {
    return this.requestManager.delete<Record<string, any>>(`/campaigns/${id}`);
  }

  async pause(id: string): Promise<Campaign> {
    return this.requestManager.put<Campaign>(`/campaigns/${id}/pause`);
  }

  async resume(id: string): Promise<Campaign> {
    return this.requestManager.put<Campaign>(`/campaigns/${id}/resume`);
  }
}

// Analytics Service
export class AnalyticsService {
  private requestManager: RequestManager;

  constructor(requestManager: RequestManager) {
    this.requestManager = requestManager;
  }

  async realtime(): Promise<Analytics> {
    return this.requestManager.get<Analytics>('/analytics/realtime');
  }

  async campaign(id: string): Promise<Analytics> {
    return this.requestManager.get<Analytics>(`/analytics/campaigns/${id}`);
  }

  async breakdown(type: string): Promise<Record<string, any>[]> {
    return this.requestManager.get<Record<string, any>[]>(`/analytics/breakdown?type=${type}`);
  }

  async dashboard(): Promise<Record<string, any>> {
    return this.requestManager.get<Record<string, any>>('/analytics/dashboard');
  }
}

// Bidding Service
export class BiddingService {
  private requestManager: RequestManager;

  constructor(requestManager: RequestManager) {
    this.requestManager = requestManager;
  }

  async submitBid(campaignId: string, adSlotId: string, amount: number, currency: string = 'USD'): Promise<Bid> {
    const request: BidSubmitRequest = { campaignId, adSlotId, amount, currency };
    return this.requestManager.post<Bid>('/bids', request);
  }

  async recommendations(): Promise<Record<string, any>[]> {
    return this.requestManager.get<Record<string, any>[]>('/bids/recommendations');
  }

  async list(limit: number = 50): Promise<Bid[]> {
    return this.requestManager.get<Bid[]>(`/bids?limit=${limit}`);
  }

  async get(id: string): Promise<Bid> {
    return this.requestManager.get<Bid>(`/bids/${id}`);
  }

  async stats(): Promise<Record<string, any>> {
    return this.requestManager.get<Record<string, any>>('/bids/stats');
  }
}

// Ad Service
export class AdService {
  private requestManager: RequestManager;

  constructor(requestManager: RequestManager) {
    this.requestManager = requestManager;
  }

  async create(campaignId: string, placement: string, imageUrl: string, clickUrl: string, dimensions: string): Promise<Ad> {
    const request: AdCreateRequest = { campaignId, placement, imageUrl, clickUrl, dimensions };
    return this.requestManager.post<Ad>('/ads', request);
  }

  async list(campaignId: string, limit: number = 50): Promise<Ad[]> {
    return this.requestManager.get<Ad[]>(`/ads?campaignId=${campaignId}&limit=${limit}`);
  }

  async get(id: string): Promise<Ad> {
    return this.requestManager.get<Ad>(`/ads/${id}`);
  }

  async update(id: string, placement?: string): Promise<Ad> {
    const updates: Record<string, any> = {};
    if (placement) updates.placement = placement;
    return this.requestManager.put<Ad>(`/ads/${id}`, updates);
  }

  async delete(id: string): Promise<Record<string, any>> {
    return this.requestManager.delete<Record<string, any>>(`/ads/${id}`);
  }
}

// Webhook Service
export class WebhookService {
  private requestManager: RequestManager;
  private eventHandlers: Map<string, Array<(event: WebhookEvent) => void>> = new Map();

  constructor(requestManager: RequestManager) {
    this.requestManager = requestManager;
  }

  async subscribe(url: string, events: string[]): Promise<Webhook> {
    const request: WebhookCreateRequest = { url, events };
    return this.requestManager.post<Webhook>('/webhooks', request);
  }

  async list(limit: number = 50): Promise<Webhook[]> {
    return this.requestManager.get<Webhook[]>(`/webhooks?limit=${limit}`);
  }

  async get(id: string): Promise<Webhook> {
    return this.requestManager.get<Webhook>(`/webhooks/${id}`);
  }

  async update(id: string, active?: boolean): Promise<Webhook> {
    const updates: Record<string, any> = {};
    if (active !== undefined) updates.active = active;
    return this.requestManager.put<Webhook>(`/webhooks/${id}`, updates);
  }

  async delete(id: string): Promise<Record<string, any>> {
    return this.requestManager.delete<Record<string, any>>(`/webhooks/${id}`);
  }

  async test(id: string): Promise<Record<string, any>> {
    return this.requestManager.post<Record<string, any>>(`/webhooks/${id}/test`);
  }

  async getLogs(id: string, limit: number = 50): Promise<Record<string, any>[]> {
    return this.requestManager.get<Record<string, any>[]>(`/webhooks/${id}/logs?limit=${limit}`);
  }

  onEvent(type: string, handler: (event: WebhookEvent) => void): void {
    if (!this.eventHandlers.has(type)) {
      this.eventHandlers.set(type, []);
    }
    this.eventHandlers.get(type)!.push(handler);
  }

  offEvent(type: string): void {
    this.eventHandlers.delete(type);
  }

  handleEvent(event: WebhookEvent): void {
    const handlers = this.eventHandlers.get(event.type);
    if (handlers) {
      handlers.forEach(handler => handler(event));
    }
  }
}
