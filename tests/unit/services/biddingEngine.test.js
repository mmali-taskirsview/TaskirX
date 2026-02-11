/**
 * Bidding Engine Tests
 * Tests for auction algorithms and bid calculation
 */

const BiddingEngine = require('../../../backend/services/biddingEngine');

describe('BiddingEngine', () => {
  describe('runAuction', () => {
    it('should return null winner when no campaigns provided', async () => {
      const result = await BiddingEngine.runAuction({
        campaigns: [],
        floorPrice: 1.0,
        impressionData: {}
      });

      expect(result.winner).toBeNull();
      expect(result.winningBid).toBe(0);
    });

    it('should select highest bidder in second-price auction', async () => {
      const campaigns = [
        {
          _id: 'campaign1',
          bidding: { maxBid: 5.0, strategy: 'cpm' },
          budget: { total: 1000, spent: 0 },
          performance: { impressions: 100, clicks: 10 }
        },
        {
          _id: 'campaign2',
          bidding: { maxBid: 3.0, strategy: 'cpm' },
          budget: { total: 1000, spent: 0 },
          performance: { impressions: 100, clicks: 10 }
        },
        {
          _id: 'campaign3',
          bidding: { maxBid: 4.0, strategy: 'cpm' },
          budget: { total: 1000, spent: 0 },
          performance: { impressions: 100, clicks: 10 }
        }
      ];

      const result = await BiddingEngine.runAuction({
        campaigns,
        floorPrice: 1.0,
        impressionData: {}
      });

      expect(result.winner._id).toBe('campaign1');
      expect(result.totalBids).toBe(3);
      expect(result.secondPrice).toBe(4.0);
      // Winner pays second price + $0.01
      expect(result.winningBid).toBeCloseTo(4.01, 2);
    });

    it('should respect floor price', async () => {
      const campaigns = [
        {
          _id: 'campaign1',
          bidding: { maxBid: 0.5, strategy: 'cpm' },
          budget: { total: 1000, spent: 0 },
          performance: { impressions: 100, clicks: 10 }
        }
      ];

      const result = await BiddingEngine.runAuction({
        campaigns,
        floorPrice: 1.0,
        impressionData: {}
      });

      expect(result.winner).toBeNull();
    });

    it('should filter out campaigns with exceeded budget', async () => {
      const campaigns = [
        {
          _id: 'campaign1',
          bidding: { maxBid: 5.0, strategy: 'cpm' },
          budget: { total: 100, spent: 100 }, // Budget exhausted
          performance: { impressions: 100, clicks: 10 }
        },
        {
          _id: 'campaign2',
          bidding: { maxBid: 3.0, strategy: 'cpm' },
          budget: { total: 1000, spent: 0 },
          performance: { impressions: 100, clicks: 10 }
        }
      ];

      const result = await BiddingEngine.runAuction({
        campaigns,
        floorPrice: 1.0,
        impressionData: {}
      });

      expect(result.winner._id).toBe('campaign2');
      expect(result.totalBids).toBe(1);
    });

    it('should handle single bidder correctly', async () => {
      const campaigns = [
        {
          _id: 'campaign1',
          bidding: { maxBid: 5.0, strategy: 'cpm' },
          budget: { total: 1000, spent: 0 },
          performance: { impressions: 100, clicks: 10 }
        }
      ];

      const result = await BiddingEngine.runAuction({
        campaigns,
        floorPrice: 1.0,
        impressionData: {}
      });

      expect(result.winner._id).toBe('campaign1');
      // Single bidder pays 90% of their bid or floor price
      expect(result.winningBid).toBeGreaterThanOrEqual(1.0);
      expect(result.winningBid).toBeLessThan(5.0);
    });
  });

  describe('runFirstPriceAuction', () => {
    it('should apply bid shading in first-price auction', () => {
      const campaigns = [
        {
          _id: 'campaign1',
          bidding: { maxBid: 10.0, strategy: 'cpm' },
          budget: { total: 1000, spent: 0 },
          performance: { impressions: 100, clicks: 10 }
        }
      ];

      const result = BiddingEngine.runFirstPriceAuction({
        campaigns,
        floorPrice: 1.0,
        impressionData: {}
      });

      expect(result.winner._id).toBe('campaign1');
      expect(result.originalBid).toBe(10.0);
      // Bid shading: 85% of original bid
      expect(result.winningBid).toBeCloseTo(8.5, 2);
    });
  });

  describe('calculateBid', () => {
    it('should return 0 for campaigns with exhausted budget', () => {
      const campaign = {
        bidding: { maxBid: 5.0 },
        budget: { total: 100, spent: 100 },
        performance: { impressions: 100, clicks: 10 }
      };

      const bid = BiddingEngine.calculateBid(campaign, {}, 1.0);
      expect(bid).toBe(0);
    });
  });
});
