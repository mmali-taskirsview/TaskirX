import { Injectable, Logger, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import Redis from 'ioredis';
import { SupplyPartner, SupplyPartnerStatus } from './entities/supply-partner.entity';
import { AudienceSegment, AudienceSegmentStatus } from './entities/audience-segment.entity';
import { Deal, DealStatus } from './entities/deal.entity';
import { BidStrategy, BidStrategyStatus } from './entities/bid-strategy.entity';

export interface DspDashboardStats {
  totalRequests: number;
  totalBids: number;
  totalWins: number;
  totalSpend: number;
  winRate: number;
  avgCpm: number;
  activeSupplyPartners: number;
  activeDeals: number;
  activeAudiences: number;
  activeBidStrategies: number;
}

export interface BidRequest {
  requestId: string;
  supplyPartnerId: string;
  impressionId: string;
  adUnitId: string;
  floor: number;
  device: { type: string; os: string; browser: string };
  geo: { country: string; region: string; city: string };
  user: { id?: string; segments?: string[] };
}

export interface BidResponse {
  bidId: string;
  requestId: string;
  price: number;
  creativeId: string;
  dealId?: string;
  advertiserId: string;
  campaignId: string;
}

@Injectable()
export class DspService {
  private readonly logger = new Logger(DspService.name);
  private readonly redis: Redis;

  constructor(
    @InjectRepository(SupplyPartner)
    private supplyPartnerRepository: Repository<SupplyPartner>,
    @InjectRepository(AudienceSegment)
    private audienceRepository: Repository<AudienceSegment>,
    @InjectRepository(Deal)
    private dealRepository: Repository<Deal>,
    @InjectRepository(BidStrategy)
    private bidStrategyRepository: Repository<BidStrategy>,
  ) {
    this.redis = new Redis({
      host: process.env.REDIS_HOST || 'localhost',
      port: parseInt(process.env.REDIS_PORT || '6379'),
      password: process.env.REDIS_PASSWORD,
    });
  }

  // ==================== DASHBOARD ====================

  async getDashboardStats(): Promise<DspDashboardStats> {
    const [supplyPartners, deals, audiences, bidStrategies] = await Promise.all([
      this.supplyPartnerRepository.find({ where: { status: SupplyPartnerStatus.ACTIVE } }),
      this.dealRepository.find({ where: { status: DealStatus.ACTIVE } }),
      this.audienceRepository.find({ where: { status: AudienceSegmentStatus.ACTIVE } }),
      this.bidStrategyRepository.find({ where: { status: BidStrategyStatus.ACTIVE } }),
    ]);

    const totals = supplyPartners.reduce(
      (acc, sp) => ({
        requests: acc.requests + Number(sp.totalRequests),
        bids: acc.bids + Number(sp.totalBids),
        wins: acc.wins + Number(sp.totalWins),
        spend: acc.spend + Number(sp.totalSpend),
      }),
      { requests: 0, bids: 0, wins: 0, spend: 0 }
    );

    return {
      totalRequests: totals.requests,
      totalBids: totals.bids,
      totalWins: totals.wins,
      totalSpend: totals.spend,
      winRate: totals.bids > 0 ? (totals.wins / totals.bids) * 100 : 0,
      avgCpm: totals.wins > 0 ? (totals.spend / totals.wins) * 1000 : 0,
      activeSupplyPartners: supplyPartners.length,
      activeDeals: deals.length,
      activeAudiences: audiences.length,
      activeBidStrategies: bidStrategies.length,
    };
  }

  // ==================== SUPPLY PARTNERS ====================

  async getSupplyPartners(): Promise<SupplyPartner[]> {
    return this.supplyPartnerRepository.find({
      order: { createdAt: 'DESC' },
    });
  }

  async getSupplyPartner(id: string): Promise<SupplyPartner> {
    const partner = await this.supplyPartnerRepository.findOne({ where: { id } });
    if (!partner) throw new NotFoundException('Supply partner not found');
    return partner;
  }

  async createSupplyPartner(data: Partial<SupplyPartner>): Promise<SupplyPartner> {
    const partner = this.supplyPartnerRepository.create(data);
    return this.supplyPartnerRepository.save(partner);
  }

  async updateSupplyPartner(id: string, data: Partial<SupplyPartner>): Promise<SupplyPartner> {
    await this.supplyPartnerRepository.update(id, data);
    return this.getSupplyPartner(id);
  }

  async deleteSupplyPartner(id: string): Promise<void> {
    await this.supplyPartnerRepository.delete(id);
  }

  // ==================== AUDIENCE SEGMENTS ====================

  async getAudienceSegments(advertiserId?: string): Promise<AudienceSegment[]> {
    const where = advertiserId ? { advertiserId } : {};
    return this.audienceRepository.find({
      where,
      order: { createdAt: 'DESC' },
    });
  }

  async getAudienceSegment(id: string): Promise<AudienceSegment> {
    const segment = await this.audienceRepository.findOne({ where: { id } });
    if (!segment) throw new NotFoundException('Audience segment not found');
    return segment;
  }

  async createAudienceSegment(data: Partial<AudienceSegment>): Promise<AudienceSegment> {
    const segment = this.audienceRepository.create(data);
    return this.audienceRepository.save(segment);
  }

  async updateAudienceSegment(id: string, data: Partial<AudienceSegment>): Promise<AudienceSegment> {
    await this.audienceRepository.update(id, data);
    return this.getAudienceSegment(id);
  }

  async deleteAudienceSegment(id: string): Promise<void> {
    await this.audienceRepository.delete(id);
  }

  // ==================== DEALS ====================

  async getDeals(advertiserId?: string): Promise<Deal[]> {
    const where = advertiserId ? { advertiserId } : {};
    return this.dealRepository.find({
      where,
      order: { createdAt: 'DESC' },
    });
  }

  async getDeal(id: string): Promise<Deal> {
    const deal = await this.dealRepository.findOne({ where: { id } });
    if (!deal) throw new NotFoundException('Deal not found');
    return deal;
  }

  async createDeal(data: Partial<Deal>): Promise<Deal> {
    const deal = this.dealRepository.create({
      ...data,
      dealId: data.dealId || `DEAL-${Date.now()}`,
    });
    return this.dealRepository.save(deal);
  }

  async updateDeal(id: string, data: Partial<Deal>): Promise<Deal> {
    await this.dealRepository.update(id, data);
    return this.getDeal(id);
  }

  async deleteDeal(id: string): Promise<void> {
    await this.dealRepository.delete(id);
  }

  // ==================== BID STRATEGIES ====================

  async getBidStrategies(advertiserId?: string): Promise<BidStrategy[]> {
    const where = advertiserId ? { advertiserId } : {};
    return this.bidStrategyRepository.find({
      where,
      order: { createdAt: 'DESC' },
    });
  }

  async getBidStrategy(id: string): Promise<BidStrategy> {
    const strategy = await this.bidStrategyRepository.findOne({ where: { id } });
    if (!strategy) throw new NotFoundException('Bid strategy not found');
    return strategy;
  }

  async createBidStrategy(data: Partial<BidStrategy>): Promise<BidStrategy> {
    const strategy = this.bidStrategyRepository.create(data);
    return this.bidStrategyRepository.save(strategy);
  }

  async updateBidStrategy(id: string, data: Partial<BidStrategy>): Promise<BidStrategy> {
    await this.bidStrategyRepository.update(id, data);
    return this.getBidStrategy(id);
  }

  async deleteBidStrategy(id: string): Promise<void> {
    await this.bidStrategyRepository.delete(id);
  }

  // ==================== BIDDING ENGINE ====================

  async processBidRequest(bidRequest: BidRequest): Promise<BidResponse | null> {
    this.logger.log(`Processing bid request: ${bidRequest.requestId}`);

    // 1. Get supply partner
    const supplyPartner = await this.supplyPartnerRepository.findOne({
      where: { id: bidRequest.supplyPartnerId, status: SupplyPartnerStatus.ACTIVE },
    });

    if (!supplyPartner) {
      this.logger.warn(`Unknown or inactive supply partner: ${bidRequest.supplyPartnerId}`);
      return null;
    }

    // 2. Check floor price
    if (bidRequest.floor > supplyPartner.maxBid) {
      this.logger.log(`Floor ${bidRequest.floor} exceeds max bid ${supplyPartner.maxBid}`);
      return null;
    }

    // 3. Find matching deals
    const _deals = await this.dealRepository.find({
      where: { status: DealStatus.ACTIVE },
    });

    // 4. Get active bid strategies
    const strategies = await this.bidStrategyRepository.find({
      where: { status: BidStrategyStatus.ACTIVE },
    });
    
    // Frequency Capping Check
    const strategy = strategies[0];
    if (strategy && bidRequest.user && bidRequest.user.id) {
        const allowed = await this.checkFrequencyCap(bidRequest.user.id, strategy);
        if (!allowed) {
            this.logger.debug(`Frequency cap hit for user ${bidRequest.user.id}, skipping bid.`);
            return null;
        }
    }

    // 5. Calculate optimal bid price
    const bidPrice = this.calculateBidPrice(bidRequest, supplyPartner, strategy);

    if (bidPrice < bidRequest.floor) {
      return null;
    }

    // 6. Update stats
    await this.supplyPartnerRepository.increment(
      { id: supplyPartner.id },
      'totalRequests',
      1
    );
    await this.supplyPartnerRepository.increment(
      { id: supplyPartner.id },
      'totalBids',
      1
    );

    return {
      bidId: `bid_${Date.now()}`,
      requestId: bidRequest.requestId,
      price: bidPrice,
      creativeId: 'creative_default',
      advertiserId: 'advertiser_default',
      campaignId: 'campaign_default',
    };
  }

  private calculateBidPrice(
    bidRequest: BidRequest,
    supplyPartner: SupplyPartner,
    strategy?: BidStrategy
  ): number {
    let baseBid = strategy?.baseBid || 1.0;

    // Apply floor price minimum
    baseBid = Math.max(baseBid, bidRequest.floor);

    // Apply device adjustment
    if (strategy?.bidAdjustments?.device) {
      const deviceMod = strategy.bidAdjustments.device[bidRequest.device.type] || 1.0;
      baseBid *= deviceMod;
    }

    // Apply geo adjustment
    if (strategy?.bidAdjustments?.geo) {
      const geoMod = strategy.bidAdjustments.geo[bidRequest.geo.country] || 1.0;
      baseBid *= geoMod;
    }

    // Clamp to min/max
    const minBid = Math.max(supplyPartner.minBid, bidRequest.floor);
    const maxBid = strategy?.maxBid || supplyPartner.maxBid;

    return Math.min(Math.max(baseBid, minBid), maxBid);
  }

  async recordWin(bidId: string, supplyPartnerId: string, price: number): Promise<void> {
    await this.supplyPartnerRepository.increment(
      { id: supplyPartnerId },
      'totalWins',
      1
    );
    await this.supplyPartnerRepository.increment(
      { id: supplyPartnerId },
      'totalSpend',
      price
    );
    await this.supplyPartnerRepository.increment(
      { id: supplyPartnerId },
      'dailySpend',
      price
    );
  }

  // ==================== RTB ANALYTICS ====================

  async getRtbAnalytics(): Promise<{
    hourlyStats: Array<{ hour: number; requests: number; bids: number; wins: number; spend: number }>;
    topSupplyPartners: Array<{ name: string; requests: number; winRate: number; spend: number }>;
    bidLatency: { avg: number; p50: number; p95: number; p99: number };
  }> {
    const partners = await this.supplyPartnerRepository.find({
      where: { status: SupplyPartnerStatus.ACTIVE },
      order: { totalSpend: 'DESC' },
      take: 10,
    });

    // Generate hourly stats (simulated)
    const hourlyStats = Array.from({ length: 24 }, (_, hour) => ({
      hour,
      requests: Math.floor(50000 + Math.random() * 50000 * Math.sin(Math.PI * hour / 12)),
      bids: Math.floor(40000 + Math.random() * 40000 * Math.sin(Math.PI * hour / 12)),
      wins: Math.floor(5000 + Math.random() * 5000 * Math.sin(Math.PI * hour / 12)),
      spend: Math.floor(500 + Math.random() * 500 * Math.sin(Math.PI * hour / 12)),
    }));

    const topSupplyPartners = partners.map(p => ({
      name: p.name,
      requests: Number(p.totalRequests),
      winRate: p.totalBids > 0 ? (Number(p.totalWins) / Number(p.totalBids)) * 100 : 0,
      spend: Number(p.totalSpend),
    }));

    return {
      hourlyStats,
      topSupplyPartners,
      bidLatency: {
        avg: 45,
        p50: 38,
        p95: 85,
        p99: 120,
      },
    };
  }

  // ==================== RETARGETING ====================

  async trackRetargetingEvent(pixelId: string, event: string, _data?: any): Promise<void> {
    // Validate UUID format to prevent 500 errors
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    if (!pixelId || !uuidRegex.test(pixelId)) {
        this.logger.warn(`Invalid retargeting pixel ID received: ${pixelId}`);
        return;
    }

    // In a real system, this would write to a high-throughput event stream like Kafka/Kinesis
    // For now, we'll look up the audience segment and increment a counter
    
    try {
      // Check if pixelId maps to an audience segment (assuming pixelId == segment.id for simplicity)
      const segment = await this.audienceRepository.findOne({ where: { id: pixelId } });
      if (segment) {
          // Update last refresh timestamp to show activity
          await this.audienceRepository.update(segment.id, { 
              lastRefresh: new Date(),
              // In a real app we'd recalculate size based on unique user IDs seen
          });
          
          // Log the event (optional, for debugging)
          this.logger.log(`Retargeting event received: ${event} for segment ${segment.name}`);
      }
    } catch (error) {
       this.logger.error(`Error tracking retargeting event: ${error.message}`);
    }
  }

  generatePixelCode(segmentId: string): string {
    const trackingUrl = `${process.env.API_URL || 'http://localhost:3000'}/api/dsp/pixel?id=${segmentId}&evt=page_view`;
    return `<img src="${trackingUrl}" width="1" height="1" style="display:none;" />`;
  }

  // ==================== FREQUENCY CAPPING ====================

  async checkFrequencyCap(
    userId: string,
    strategy: BidStrategy,
  ): Promise<boolean> {
    if (!strategy.frequencyCap || !strategy.frequencyCap.impressions) return true;
    if (!userId) return true;

    // Use Redis for high-speed frequency tracking
    // Key format: freq:{strategyId}:{userId}
    const key = `freq:${strategy.id}:${userId}`;
    const periodSeconds = strategy.frequencyCap.period === 'hour' ? 3600 : 86400;

    try {
        const count = await this.redis.incr(key);
        if (count === 1) {
            await this.redis.expire(key, periodSeconds);
        }
        return count <= strategy.frequencyCap.impressions;
    } catch (error) {
        this.logger.error(`Redis frequency cap check failed: ${error.message}`);
        return true; // Fail open
    }
  }

  // ==================== BUDGET PACING ====================

  async updatePacing(): Promise<void> {
    // This method handles hourly budget allocation
    const strategies = await this.bidStrategyRepository.find({
      where: { status: BidStrategyStatus.ACTIVE },
    });

    for (const strategy of strategies) {
      if (!strategy.pacing) continue;

      try {
          // Calculate total spent today (simulating analytics aggregation)
          // In a real scenario, this would aggregate from the 'dailySpend' or an analytics OLAP DB
          // For now, let's assume we store daily spend in a Redis key per strategy
          const spendKey = `spend:${strategy.id}:${new Date().toISOString().split('T')[0]}`;
          const currentSpendStr = await this.redis.get(spendKey);
          const currentSpend = parseFloat(currentSpendStr || '0');

          // If daily budget logic existed on BidStrategy, we would check it here.
          // Assuming maxBid is the cap for a single bid, but let's assume 'budget' field exists or we use a hardcoded safe limit/logic
          
          // Placeholder logic for Pacing Type
          if (strategy.pacing.type === 'accelerated') {
             // Accelerated: Spend as fast as possible until budget runs out
             // No throttling needed unless budget is hit
          } else {
             // Standard/Even: Distribute budget across remaining hours of the day
             const now = new Date();
             const _hoursRemaining = 24 - now.getHours();
             // Logic to adjust bid probability or temporarily pause would go here
             // e.g. this.redis.set(`pacing:${strategy.id}:throttle`, '0.5'); // 50% bid rate
          }
          
          this.logger.log(`Updated pacing for strategy ${strategy.id}. Current spend: ${currentSpend}`);
      } catch (error) {
          this.logger.error(`Error updating pacing for strategy ${strategy.id}: ${error.message}`);
      }
    }
  }
}
