import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { DemandPartner, DemandPartnerType, DemandPartnerStatus } from './entities/demand-partner.entity';

export interface CreateDemandPartnerDto {
  name: string;
  code: string;
  type: DemandPartnerType;
  publisherId?: string;
  endpoint?: string;
  credentials?: Record<string, any>;
  settings?: Record<string, any>;
  revenueShare?: number;
  bidTimeout?: number;
  isGlobal?: boolean;
}

export interface UpdateDemandPartnerDto {
  name?: string;
  endpoint?: string;
  credentials?: Record<string, any>;
  settings?: Record<string, any>;
  revenueShare?: number;
  bidTimeout?: number;
  status?: DemandPartnerStatus;
}

export interface DemandPartnerBidConfig {
  partnerId: string;
  endpoint: string;
  timeout: number;
  headers: Record<string, string>;
  bidRequestTransform?: (request: any) => any;
  bidResponseTransform?: (response: any) => any;
}

@Injectable()
export class DemandPartnerService {
  // Pre-configured demand partner templates
  private readonly partnerTemplates = {
    google_adx: {
      name: 'Google AdX',
      type: DemandPartnerType.DSP,
      endpoint: 'https://pubads.g.doubleclick.net/gampad/adx',
      bidTimeout: 100,
      revenueShare: 15,
    },
    prebid: {
      name: 'Prebid.js',
      type: DemandPartnerType.HEADER_BIDDING,
      endpoint: 'local',
      bidTimeout: 1000,
      revenueShare: 0,
    },
    amazon_tam: {
      name: 'Amazon TAM',
      type: DemandPartnerType.HEADER_BIDDING,
      endpoint: 'https://aax.amazon-adsystem.com/e/dtb/bid',
      bidTimeout: 300,
      revenueShare: 10,
    },
    openx: {
      name: 'OpenX',
      type: DemandPartnerType.DSP,
      endpoint: 'https://rtb.openx.net/bid',
      bidTimeout: 100,
      revenueShare: 12,
    },
    appnexus: {
      name: 'AppNexus (Xandr)',
      type: DemandPartnerType.DSP,
      endpoint: 'https://ib.adnxs.com/openrtb2',
      bidTimeout: 100,
      revenueShare: 10,
    },
    rubicon: {
      name: 'Rubicon Project',
      type: DemandPartnerType.DSP,
      endpoint: 'https://fastlane.rubiconproject.com/a/api/fastlane.json',
      bidTimeout: 150,
      revenueShare: 12,
    },
    index_exchange: {
      name: 'Index Exchange',
      type: DemandPartnerType.DSP,
      endpoint: 'https://htlb.casalemedia.com/openrtb/pbjs',
      bidTimeout: 100,
      revenueShare: 10,
    },
    pubmatic: {
      name: 'PubMatic',
      type: DemandPartnerType.DSP,
      endpoint: 'https://hbopenbid.pubmatic.com/translator',
      bidTimeout: 100,
      revenueShare: 11,
    },
  };

  constructor(
    @InjectRepository(DemandPartner)
    private demandPartnerRepository: Repository<DemandPartner>,
  ) {}

  async findAll(publisherId?: string): Promise<DemandPartner[]> {
    if (publisherId) {
      // Return global partners + publisher-specific partners
      return this.demandPartnerRepository.find({
        where: [
          { isGlobal: true },
          { publisherId },
        ],
        order: { name: 'ASC' },
      });
    }
    return this.demandPartnerRepository.find({ order: { name: 'ASC' } });
  }

  async findActive(publisherId?: string): Promise<DemandPartner[]> {
    if (publisherId) {
      return this.demandPartnerRepository.find({
        where: [
          { isGlobal: true, status: DemandPartnerStatus.ACTIVE },
          { publisherId, status: DemandPartnerStatus.ACTIVE },
        ],
        order: { name: 'ASC' },
      });
    }
    return this.demandPartnerRepository.find({
      where: { status: DemandPartnerStatus.ACTIVE },
      order: { name: 'ASC' },
    });
  }

  async findOne(id: string): Promise<DemandPartner> {
    const partner = await this.demandPartnerRepository.findOne({ where: { id } });
    if (!partner) {
      throw new NotFoundException(`Demand partner with ID ${id} not found`);
    }
    return partner;
  }

  async findByCode(code: string): Promise<DemandPartner | null> {
    return this.demandPartnerRepository.findOne({ where: { code } });
  }

  async create(dto: CreateDemandPartnerDto): Promise<DemandPartner> {
    const partner = this.demandPartnerRepository.create({
      ...dto,
      status: DemandPartnerStatus.PENDING,
    });
    return this.demandPartnerRepository.save(partner);
  }

  async createFromTemplate(templateCode: string, publisherId?: string): Promise<DemandPartner> {
    const template = this.partnerTemplates[templateCode];
    if (!template) {
      throw new NotFoundException(`Template ${templateCode} not found`);
    }

    return this.create({
      ...template,
      code: templateCode,
      publisherId,
      isGlobal: !publisherId,
    });
  }

  async update(id: string, dto: UpdateDemandPartnerDto): Promise<DemandPartner> {
    const partner = await this.findOne(id);
    Object.assign(partner, dto);
    return this.demandPartnerRepository.save(partner);
  }

  async delete(id: string): Promise<void> {
    const partner = await this.findOne(id);
    await this.demandPartnerRepository.remove(partner);
  }

  async activate(id: string): Promise<DemandPartner> {
    return this.update(id, { status: DemandPartnerStatus.ACTIVE });
  }

  async pause(id: string): Promise<DemandPartner> {
    return this.update(id, { status: DemandPartnerStatus.PAUSED });
  }

  async disconnect(id: string): Promise<DemandPartner> {
    return this.update(id, { status: DemandPartnerStatus.DISCONNECTED });
  }

  async getPartnerBidConfig(id: string): Promise<DemandPartnerBidConfig> {
    const partner = await this.findOne(id);
    
    return {
      partnerId: partner.id,
      endpoint: partner.endpoint || '',
      timeout: partner.bidTimeout,
      headers: this.buildHeaders(partner),
    };
  }

  private buildHeaders(partner: DemandPartner): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    };

    if (partner.credentials?.apiKey) {
      headers['Authorization'] = `Bearer ${partner.credentials.apiKey}`;
    }

    return headers;
  }

  async getPartnerStats(id: string) {
    const partner = await this.findOne(id);
    
    return {
      totalImpressions: Number(partner.totalImpressions),
      totalRevenue: Number(partner.totalRevenue),
      winRate: Number(partner.winRate),
      avgBidPrice: Number(partner.avgBidPrice),
      ecpm: partner.totalImpressions > 0 
        ? (Number(partner.totalRevenue) / Number(partner.totalImpressions)) * 1000 
        : 0,
    };
  }

  async getAvailableTemplates(): Promise<Array<{ code: string; name: string; type: string }>> {
    return Object.entries(this.partnerTemplates).map(([code, template]) => ({
      code,
      name: template.name,
      type: template.type,
    }));
  }

  async initializeDefaultPartners(): Promise<DemandPartner[]> {
    const existing = await this.demandPartnerRepository.find({ where: { isGlobal: true } });
    
    if (existing.length > 0) {
      return existing;
    }

    const defaultPartners = ['google_adx', 'prebid', 'amazon_tam', 'openx', 'appnexus'];
    const created: DemandPartner[] = [];

    for (const code of defaultPartners) {
      const partner = await this.createFromTemplate(code);
      created.push(partner);
    }

    return created;
  }
}
