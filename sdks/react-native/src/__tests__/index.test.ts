// Tests for TaskirX React Native SDK

import { TaskirXClient } from '../src';
import {
  AuthService,
  CampaignService,
  AnalyticsService,
  BiddingService,
  AdService,
  WebhookService,
} from '../src/services/Services';
import { RequestManager } from '../src/network/RequestManager';
import type { ClientConfig, User, Campaign, Bid, Analytics } from '../src/types';

describe('TaskirXClient', () => {
  let client: TaskirXClient;

  beforeEach(() => {
    client = TaskirXClient.create('http://localhost:3000', 'test-key', true);
  });

  describe('Client Creation', () => {
    it('should create client instance', () => {
      expect(client).toBeDefined();
    });

    it('should have all services', () => {
      expect(client.auth).toBeInstanceOf(AuthService);
      expect(client.campaigns).toBeInstanceOf(CampaignService);
      expect(client.analytics).toBeInstanceOf(AnalyticsService);
      expect(client.bidding).toBeInstanceOf(BiddingService);
      expect(client.ads).toBeInstanceOf(AdService);
      expect(client.webhooks).toBeInstanceOf(WebhookService);
    });

    it('should enable debug mode', () => {
      const consoleSpy = jest.spyOn(console, 'log');
      client.enableDebug(true);
      expect(consoleSpy).toHaveBeenCalledWith('[TaskirX] 🐛 Debug mode enabled');
      consoleSpy.mockRestore();
    });
  });

  describe('Models', () => {
    it('should create user model', () => {
      const user: User = {
        id: 'user-1',
        email: 'test@example.com',
        name: 'Test User',
        company: 'Test Co',
        role: 'admin',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-02T00:00:00Z',
      };

      expect(user.id).toBe('user-1');
      expect(user.email).toBe('test@example.com');
    });

    it('should create campaign model', () => {
      const campaign: Campaign = {
        id: 'camp-1',
        name: 'Test Campaign',
        budget: 1000.0,
        startDate: '2024-01-01',
        endDate: '2024-01-31',
        status: 'active',
        targetAudience: {},
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      };

      expect(campaign.id).toBe('camp-1');
      expect(campaign.budget).toBe(1000.0);
    });

    it('should create bid model', () => {
      const bid: Bid = {
        id: 'bid-1',
        campaignId: 'camp-1',
        adSlotId: 'slot-1',
        amount: 5.0,
        currency: 'USD',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      };

      expect(bid.id).toBe('bid-1');
      expect(bid.amount).toBe(5.0);
    });

    it('should create analytics model', () => {
      const analytics: Analytics = {
        impressions: 1000,
        clicks: 50,
        conversions: 10,
        spend: 100.0,
        revenue: 200.0,
        ctr: 0.05,
        conversionRate: 0.1,
        roi: 1.0,
        timestamp: '2024-01-01T00:00:00Z',
      };

      expect(analytics.impressions).toBe(1000);
      expect(analytics.ctr).toBe(0.05);
    });
  });

  describe('Result Type', () => {
    it('should handle success result', () => {
      const result = { success: true as const, data: 'test' };
      expect(result.success).toBe(true);
      expect(result.data).toBe('test');
    });

    it('should handle failure result', () => {
      const error = {
        type: 'NETWORK_ERROR' as const,
        message: 'Network error',
      };
      const result = { success: false as const, error };
      expect(result.success).toBe(false);
      expect(result.error.type).toBe('NETWORK_ERROR');
    });
  });

  describe('Configuration', () => {
    it('should create client with custom config', () => {
      const customClient = TaskirXClient.create(
        'https://api.example.com',
        'custom-key',
        false
      );
      expect(customClient).toBeDefined();
    });

    it('should initialize RequestManager with config', () => {
      const config: ClientConfig = {
        apiUrl: 'http://localhost:3000',
        apiKey: 'test-key',
        debug: true,
        timeout: 30000,
        retryAttempts: 3,
      };

      const rm = new RequestManager(config);
      expect(rm).toBeDefined();
    });
  });
});

describe('AuthService', () => {
  let requestManager: RequestManager;
  let authService: AuthService;

  beforeEach(() => {
    const config: ClientConfig = {
      apiUrl: 'http://localhost:3000',
      apiKey: 'test-key',
    };
    requestManager = new RequestManager(config);
    authService = new AuthService(requestManager);
  });

  it('should create LoginRequest', () => {
    const request = {
      email: 'test@example.com',
      password: 'password123',
    };

    expect(request.email).toBe('test@example.com');
    expect(request.password).toBe('password123');
  });

  it('should create RegisterRequest', () => {
    const request = {
      email: 'newuser@example.com',
      password: 'password123',
      name: 'New User',
      company: 'New Co',
    };

    expect(request.email).toBe('newuser@example.com');
    expect(request.name).toBe('New User');
  });
});

describe('CampaignService', () => {
  let requestManager: RequestManager;
  let campaignService: CampaignService;

  beforeEach(() => {
    const config: ClientConfig = {
      apiUrl: 'http://localhost:3000',
      apiKey: 'test-key',
    };
    requestManager = new RequestManager(config);
    campaignService = new CampaignService(requestManager);
  });

  it('should initialize CampaignService', () => {
    expect(campaignService).toBeDefined();
  });

  it('should create CampaignCreateRequest', () => {
    const request = {
      name: 'Test Campaign',
      budget: 1000.0,
      startDate: '2024-01-01',
      endDate: '2024-01-31',
      targetAudience: { age: '18-35' },
    };

    expect(request.name).toBe('Test Campaign');
    expect(request.budget).toBe(1000.0);
  });
});

describe('BiddingService', () => {
  let requestManager: RequestManager;
  let biddingService: BiddingService;

  beforeEach(() => {
    const config: ClientConfig = {
      apiUrl: 'http://localhost:3000',
      apiKey: 'test-key',
    };
    requestManager = new RequestManager(config);
    biddingService = new BiddingService(requestManager);
  });

  it('should initialize BiddingService', () => {
    expect(biddingService).toBeDefined();
  });

  it('should create BidSubmitRequest', () => {
    const request = {
      campaignId: 'camp-1',
      adSlotId: 'slot-1',
      amount: 5.0,
      currency: 'USD',
    };

    expect(request.campaignId).toBe('camp-1');
    expect(request.amount).toBe(5.0);
  });
});

describe('WebhookService', () => {
  let requestManager: RequestManager;
  let webhookService: WebhookService;

  beforeEach(() => {
    const config: ClientConfig = {
      apiUrl: 'http://localhost:3000',
      apiKey: 'test-key',
    };
    requestManager = new RequestManager(config);
    webhookService = new WebhookService(requestManager);
  });

  it('should initialize WebhookService', () => {
    expect(webhookService).toBeDefined();
  });

  it('should handle webhook events', () => {
    const mockHandler = jest.fn();
    webhookService.onEvent('test.event', mockHandler);

    const event = {
      id: 'event-1',
      type: 'test.event',
      data: { test: 'data' },
      timestamp: '2024-01-01T00:00:00Z',
    };

    webhookService.handleEvent(event);
    expect(mockHandler).toHaveBeenCalledWith(event);
  });

  it('should register and unregister event handlers', () => {
    const handler1 = jest.fn();
    const handler2 = jest.fn();

    webhookService.onEvent('event.type', handler1);
    webhookService.onEvent('event.type', handler2);

    webhookService.offEvent('event.type');

    const event = {
      id: 'event-1',
      type: 'event.type',
      data: {},
      timestamp: '2024-01-01T00:00:00Z',
    };

    webhookService.handleEvent(event);
    expect(handler1).not.toHaveBeenCalled();
    expect(handler2).not.toHaveBeenCalled();
  });
});

describe('Integration Tests', () => {
  let client: TaskirXClient;

  beforeEach(() => {
    client = TaskirXClient.create('http://localhost:3000', 'test-key');
  });

  it('should integrate all services', () => {
    expect(client.auth).toBeDefined();
    expect(client.campaigns).toBeDefined();
    expect(client.analytics).toBeDefined();
    expect(client.bidding).toBeDefined();
    expect(client.ads).toBeDefined();
    expect(client.webhooks).toBeDefined();
  });

  it('should create multiple client instances', () => {
    const client1 = TaskirXClient.create('http://localhost:3000', 'key-1');
    const client2 = TaskirXClient.create('http://localhost:3001', 'key-2');

    expect(client1).toBeDefined();
    expect(client2).toBeDefined();
    expect(client1).not.toBe(client2);
  });

  it('should handle webhook event subscriptions', () => {
    const mockHandler = jest.fn();
    client.onWebhookEvent('test.event', mockHandler);

    const event = {
      id: 'event-1',
      type: 'test.event',
      data: { test: 'value' },
      timestamp: '2024-01-01T00:00:00Z',
    };

    client.webhooks.handleEvent(event);
    expect(mockHandler).toHaveBeenCalledWith(event);
  });

  it('should handle result types correctly', () => {
    const successResult = { success: true as const, data: { test: 'data' } };
    const failureResult = {
      success: false as const,
      error: {
        type: 'NETWORK_ERROR' as const,
        message: 'Network failed',
      },
    };

    expect(successResult.success).toBe(true);
    expect(failureResult.success).toBe(false);
  });
});
