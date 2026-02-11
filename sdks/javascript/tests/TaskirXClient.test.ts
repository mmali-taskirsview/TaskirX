/**
 * TaskirX Client Tests
 */

import TaskirXClient from '../src/client';
import { ClientConfig } from '../src/types';

describe('TaskirXClient', () => {
  let client: TaskirXClient;
  const mockConfig: ClientConfig = {
    apiUrl: 'http://localhost:3000',
    apiKey: 'test-api-key-123',
    debug: true,
  };

  beforeEach(() => {
    client = new TaskirXClient(mockConfig);
  });

  describe('Initialization', () => {
    it('should initialize with valid config', () => {
      expect(client).toBeDefined();
      expect(client.auth).toBeDefined();
      expect(client.campaigns).toBeDefined();
      expect(client.analytics).toBeDefined();
      expect(client.bidding).toBeDefined();
      expect(client.ads).toBeDefined();
      expect(client.webhooks).toBeDefined();
    });

    it('should have all services available', () => {
      expect(client.auth).toBeTruthy();
      expect(client.campaigns).toBeTruthy();
      expect(client.analytics).toBeTruthy();
      expect(client.bidding).toBeTruthy();
      expect(client.ads).toBeTruthy();
      expect(client.webhooks).toBeTruthy();
    });

    it('should support debug mode toggling', () => {
      expect(() => client.enableDebug(true)).not.toThrow();
      expect(() => client.enableDebug(false)).not.toThrow();
    });

    it('should allow setting API key', () => {
      expect(() => client.setApiKey('new-api-key')).not.toThrow();
    });
  });

  describe('Service Access', () => {
    it('should access auth service', () => {
      expect(client.auth).toBeDefined();
      expect(typeof client.auth.register).toBe('function');
      expect(typeof client.auth.login).toBe('function');
      expect(typeof client.auth.logout).toBe('function');
    });

    it('should access campaigns service', () => {
      expect(client.campaigns).toBeDefined();
      expect(typeof client.campaigns.create).toBe('function');
      expect(typeof client.campaigns.list).toBe('function');
      expect(typeof client.campaigns.get).toBe('function');
      expect(typeof client.campaigns.update).toBe('function');
      expect(typeof client.campaigns.delete).toBe('function');
    });

    it('should access analytics service', () => {
      expect(client.analytics).toBeDefined();
      expect(typeof client.analytics.getRealtime).toBe('function');
      expect(typeof client.analytics.getCampaignAnalytics).toBe('function');
      expect(typeof client.analytics.getBreakdown).toBe('function');
    });

    it('should access bidding service', () => {
      expect(client.bidding).toBeDefined();
      expect(typeof client.bidding.submitBid).toBe('function');
      expect(typeof client.bidding.getRecommendations).toBe('function');
      expect(typeof client.bidding.getStats).toBe('function');
    });

    it('should access ads service', () => {
      expect(client.ads).toBeDefined();
      expect(typeof client.ads.create).toBe('function');
      expect(typeof client.ads.list).toBe('function');
      expect(typeof client.ads.delete).toBe('function');
    });

    it('should access webhooks service', () => {
      expect(client.webhooks).toBeDefined();
      expect(typeof client.webhooks.subscribe).toBe('function');
      expect(typeof client.webhooks.list).toBe('function');
      expect(typeof client.webhooks.onEvent).toBe('function');
    });
  });

  describe('Client Methods', () => {
    it('should have getProfile method', () => {
      expect(typeof client.getProfile).toBe('function');
    });

    it('should have logout method', () => {
      expect(typeof client.logout).toBe('function');
    });

    it('should have getDashboard method', () => {
      expect(typeof client.getDashboard).toBe('function');
    });

    it('should have getCampaignPerformance method', () => {
      expect(typeof client.getCampaignPerformance).toBe('function');
    });

    it('should have getStatistics method', () => {
      expect(typeof client.getStatistics).toBe('function');
    });

    it('should have createCampaigns batch method', () => {
      expect(typeof client.createCampaigns).toBe('function');
    });
  });

  describe('Configuration', () => {
    it('should use provided configuration', () => {
      const customConfig: ClientConfig = {
        apiUrl: 'https://api.example.com',
        apiKey: 'custom-key',
        debug: false,
        timeout: 60000,
        retryAttempts: 5,
      };

      const customClient = new TaskirXClient(customConfig);
      expect(customClient).toBeDefined();
    });

    it('should support minimal configuration', () => {
      const minimalConfig: ClientConfig = {
        apiUrl: 'http://localhost:3000',
        apiKey: 'test-key',
      };

      const minimalClient = new TaskirXClient(minimalConfig);
      expect(minimalClient).toBeDefined();
    });
  });
});
