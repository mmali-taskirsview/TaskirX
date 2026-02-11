import { Injectable, Logger } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Publisher, PublisherStatus } from './entities/publisher.entity';
import { AdUnit, AdUnitStatus } from './entities/ad-unit.entity';
import { Placement } from './entities/placement.entity';
import { FloorPrice } from './entities/floor-price.entity';
import { DemandPartner, DemandPartnerStatus } from './entities/demand-partner.entity';

export interface BidRequest {
  id: string;
  publisherId: string;
  adUnitId: string;
  placementId?: string;
  device: {
    type: string;
    os: string;
    browser: string;
  };
  geo: {
    country: string;
    region: string;
    city: string;
  };
  user: {
    id?: string;
    consent?: boolean;
    tcf?: {
      tcString?: string | null;
      gdprApplies?: boolean | null;
      purposeConsents?: Record<string, string | boolean>;
      vendorConsents?: Record<string, string | boolean>;
    };
    gpp?: {
      gppString?: string | null;
      sections?: string[];
    };
    ccpa?: {
      usPrivacyString?: string | null;
      optOutSale?: boolean | null;
    };
    gdpr?: {
      applies?: boolean | null;
      consentString?: string | null;
    };
    identity?: {
      uid2?: string | null;
      id5?: string | null;
      liveramp?: string | null;
      ttd?: string | null;
      sharedId?: string | null;
      pubcid?: string | null;
      ppid?: string | null;
    };
  };
  floor?: number;
  timeout?: number;
}

export interface BidResponse {
  bidId: string;
  partnerId: string;
  partnerName: string;
  price: number;
  creative: {
    id: string;
    type: string;
    url: string;
    width: number;
    height: number;
  };
  dealId?: string;
}

@Injectable()
export class SspService {
  private readonly logger = new Logger(SspService.name);

  constructor(
    @InjectRepository(Publisher)
    private publisherRepository: Repository<Publisher>,
    @InjectRepository(AdUnit)
    private adUnitRepository: Repository<AdUnit>,
    @InjectRepository(Placement)
    private placementRepository: Repository<Placement>,
    @InjectRepository(FloorPrice)
    private floorPriceRepository: Repository<FloorPrice>,
    @InjectRepository(DemandPartner)
    private demandPartnerRepository: Repository<DemandPartner>,
  ) {}

  /**
   * Main SSP auction orchestration
   */
  async runAuction(bidRequest: BidRequest): Promise<BidResponse | null> {
    this.logger.log(`Running auction for ad unit: ${bidRequest.adUnitId}`);

    if (!this.isConsentAllowed(bidRequest)) {
      this.logger.warn(`Consent not granted for request: ${bidRequest.id}`);
      return null;
    }

    // 1. Validate publisher and ad unit
    const publisher = await this.validatePublisher(bidRequest.publisherId);
    if (!publisher) {
      this.logger.warn(`Invalid publisher: ${bidRequest.publisherId}`);
      return null;
    }

    const adUnit = await this.validateAdUnit(bidRequest.adUnitId);
    if (!adUnit) {
      this.logger.warn(`Invalid ad unit: ${bidRequest.adUnitId}`);
      return null;
    }

    // 2. Calculate effective floor price
    const floorPrice = await this.calculateFloorPrice(bidRequest);
    this.logger.log(`Floor price: $${floorPrice}`);

    // 3. Get active demand partners
    const partners = await this.getActiveDemandPartners(bidRequest.publisherId);
    this.logger.log(`Found ${partners.length} active demand partners`);

    // 4. Send bid requests to all partners in parallel
    const bids = await this.collectBids(bidRequest, partners, floorPrice);
    this.logger.log(`Collected ${bids.length} valid bids`);

    // 5. Run auction (second-price)
    const winner = this.selectWinner(bids, floorPrice);

    if (winner) {
      this.logger.log(`Winner: ${winner.partnerName} at $${winner.price}`);
      
      // 6. Update stats
      await this.updateAuctionStats(adUnit.id, winner.partnerId, winner.price);
    }

    return winner;
  }

  private isConsentAllowed(bidRequest: BidRequest): boolean {
    if (bidRequest.user?.consent === false) {
      return false;
    }

    const tcf = bidRequest.user?.tcf;
    if (tcf?.gdprApplies === true && !tcf.tcString) {
      return false;
    }

    const gpp = bidRequest.user?.gpp;
    if (gpp?.gppString === 'OPT_OUT') {
      return false;
    }

    const ccpa = bidRequest.user?.ccpa;
    if (ccpa?.optOutSale === true) {
      return false;
    }

    if (ccpa?.usPrivacyString && this.isUsPrivacyOptOut(ccpa.usPrivacyString)) {
      return false;
    }

    const gdpr = bidRequest.user?.gdpr;
    if (gdpr?.applies === true && !gdpr.consentString) {
      return false;
    }

    return true;
  }

  private isUsPrivacyOptOut(usPrivacyString: string): boolean {
    if (usPrivacyString.length < 3) {
      return false;
    }

    return usPrivacyString[2]?.toUpperCase() === 'Y';
  }

  /**
   * Validate publisher is active
   */
  async validatePublisher(publisherId: string): Promise<Publisher | null> {
    try {
      const publisher = await this.publisherRepository.findOne({
        where: { id: publisherId, status: PublisherStatus.ACTIVE },
      });
      return publisher;
    } catch (error) {
      this.logger.error(`Error validating publisher: ${error.message}`);
      return null;
    }
  }

  /**
   * Validate ad unit is active
   */
  async validateAdUnit(adUnitId: string): Promise<AdUnit | null> {
    try {
      const adUnit = await this.adUnitRepository.findOne({
        where: { id: adUnitId, status: AdUnitStatus.ACTIVE },
      });
      return adUnit;
    } catch (error) {
      this.logger.error(`Error validating ad unit: ${error.message}`);
      return null;
    }
  }

  /**
   * Calculate effective floor price based on rules
   */
  async calculateFloorPrice(bidRequest: BidRequest): Promise<number> {
    // Get floor price rules ordered by priority
    const rules = await this.floorPriceRepository.find({
      where: {
        isActive: true,
        publisherId: bidRequest.publisherId,
      },
      order: { priority: 'DESC' },
    });

    let floorPrice = bidRequest.floor || 0.01; // Default minimum

    for (const rule of rules) {
      if (this.matchesFloorPriceRule(rule, bidRequest)) {
        switch (rule.action) {
          case 'set_floor':
            floorPrice = Number(rule.price);
            break;
          case 'multiply':
            floorPrice *= Number(rule.price);
            break;
          case 'add':
            floorPrice += Number(rule.price);
            break;
        }
      }
    }

    return Math.max(floorPrice, 0.01);
  }

  /**
   * Check if bid request matches floor price rule conditions
   */
  private matchesFloorPriceRule(rule: FloorPrice, bidRequest: BidRequest): boolean {
    if (!rule.conditions) return true;

    const conditions = rule.conditions;

    // Check geo conditions
    if (conditions.countries && conditions.countries.length > 0) {
      if (!conditions.countries.includes(bidRequest.geo.country)) {
        return false;
      }
    }

    // Check device conditions
    if (conditions.devices && conditions.devices.length > 0) {
      if (!conditions.devices.includes(bidRequest.device.type)) {
        return false;
      }
    }

    // Check time conditions
    if (conditions.hours) {
      const currentHour = new Date().getUTCHours();
      if (!conditions.hours.includes(currentHour)) {
        return false;
      }
    }

    return true;
  }

  /**
   * Get active demand partners for publisher
   */
  async getActiveDemandPartners(publisherId: string): Promise<DemandPartner[]> {
    // Get global partners + publisher-specific partners
    return this.demandPartnerRepository.find({
      where: [
        { isGlobal: true, status: DemandPartnerStatus.ACTIVE },
        { publisherId, status: DemandPartnerStatus.ACTIVE },
      ],
    });
  }

  /**
   * Collect bids from all demand partners
   */
  async collectBids(
    bidRequest: BidRequest,
    partners: DemandPartner[],
    floorPrice: number,
  ): Promise<BidResponse[]> {
    const timeout = bidRequest.timeout || 100; // Default 100ms timeout
    
    const bidPromises = partners.map(async (partner) => {
      try {
        // Simulate bid request to partner
        const bid = await this.requestBidFromPartner(partner, bidRequest, timeout);
        
        // Only accept bids above floor
        if (bid && bid.price >= floorPrice) {
          return bid;
        }
        return null;
      } catch (error) {
        this.logger.warn(`Bid timeout/error from ${partner.name}: ${error.message}`);
        return null;
      }
    });

    const results = await Promise.allSettled(bidPromises);
    return results
      .filter((r): r is PromiseFulfilledResult<BidResponse> => 
        r.status === 'fulfilled' && r.value !== null)
      .map(r => r.value);
  }

  /**
   * Simulate requesting bid from a demand partner
   */
  private async requestBidFromPartner(
    partner: DemandPartner,
    bidRequest: BidRequest,
    timeout: number,
  ): Promise<BidResponse | null> {
    // Check if this is our Internal DSP (Special Mock for Live Testing)
    if (partner.code === 'INTERNAL_DSP') {
      // In a real scenario, this would call CampaignsService to match active campaigns
      // For this comprehensive test, we will simulate a "Real Match"
      return {
        bidId: `bid_${Date.now()}_internal`,
        partnerId: partner.id,
        partnerName: partner.name,
        price: 2.50, // Fixed price for testing
        creative: {
          id: `cr_internal_${Date.now()}`,
          type: 'banner',
          url: `https://cdn.taskirx.com/creatives/banner-sample.jpg`,
          width: 300,
          height: 250,
        },
      };
    }

    // External Partners (Simulated Network Latency)
    return new Promise((resolve) => {
      const delay = Math.random() * timeout;
      
      setTimeout(() => {
        // Simulate 70% bid rate
        if (Math.random() > 0.3) {
          const bid: BidResponse = {
            bidId: `bid_${Date.now()}_${partner.code}`,
            partnerId: partner.id,
            partnerName: partner.name,
            price: Math.random() * 5 + 0.5, // Random price $0.50 - $5.50
            creative: {
              id: `cr_${partner.code}_${Date.now()}`,
              type: 'banner',
              url: `https://cdn.${partner.code}.com/creative.html`,
              width: 300,
              height: 250,
            },
          };
          resolve(bid);
        } else {
          resolve(null);
        }
      }, delay);
    });
  }

  /**
   * Second-price auction: winner pays second highest bid + $0.01
   */
  selectWinner(bids: BidResponse[], floorPrice: number): BidResponse | null {
    if (bids.length === 0) return null;

    // Sort by price descending
    const sorted = [...bids].sort((a, b) => b.price - a.price);
    const winner = sorted[0];

    // Second price (or floor if only one bid)
    const secondPrice = sorted.length > 1 
      ? Math.max(sorted[1].price, floorPrice)
      : floorPrice;

    // Winner pays second price + $0.01
    winner.price = secondPrice + 0.01;

    return winner;
  }

  /**
   * Update auction statistics
   */
  async updateAuctionStats(
    adUnitId: string,
    partnerId: string,
    price: number,
  ): Promise<void> {
    // Update ad unit stats
    await this.adUnitRepository.increment(
      { id: adUnitId },
      'impressions',
      1,
    );
    await this.adUnitRepository.increment(
      { id: adUnitId },
      'revenue',
      price,
    );

    // Update demand partner stats
    await this.demandPartnerRepository.increment(
      { id: partnerId },
      'totalImpressions',
      1,
    );
    await this.demandPartnerRepository.increment(
      { id: partnerId },
      'totalRevenue',
      price,
    );
  }

  /**
   * Get SSP dashboard metrics
   */
  async getDashboardMetrics(publisherId?: string) {
    const where = publisherId ? { publisherId } : {};
    
    const [adUnits, placements, partners] = await Promise.all([
      this.adUnitRepository.find({ where }),
      this.placementRepository.find({ where }),
      this.demandPartnerRepository.find({
        where: publisherId 
          ? [{ publisherId, status: DemandPartnerStatus.ACTIVE }, { isGlobal: true, status: DemandPartnerStatus.ACTIVE }]
          : { status: DemandPartnerStatus.ACTIVE },
      }),
    ]);

    const totalImpressions = adUnits.reduce((sum, u) => sum + Number(u.impressions), 0);
    const totalRequests = adUnits.reduce((sum, u) => sum + Number(u.requests), 0);
    const totalRevenue = adUnits.reduce((sum, u) => sum + Number(u.revenue), 0);
    const fillRate = totalRequests > 0 ? (totalImpressions / totalRequests) * 100 : 0;
    const ecpm = totalImpressions > 0 ? (totalRevenue / totalImpressions) * 1000 : 0;

    return {
      totalAdUnits: adUnits.length,
      activeAdUnits: adUnits.filter(u => u.status === AdUnitStatus.ACTIVE).length,
      totalPlacements: placements.length,
      activeDemandPartners: partners.length,
      metrics: {
        totalImpressions,
        totalRequests,
        totalRevenue,
        fillRate: fillRate.toFixed(2),
        ecpm: ecpm.toFixed(4),
      },
    };
  }
}
