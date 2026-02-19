import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { ConfigService } from '@nestjs/config';
import { Campaign, CampaignStatus } from './campaign.entity';
import { Creative } from '../creatives/creative.entity';
import { CreateCampaignDto } from './dto/create-campaign.dto';
import Redis from 'ioredis';

@Injectable()
export class CampaignsService {
  private redisClient: Redis;

  constructor(
    @InjectRepository(Campaign)
    private campaignsRepository: Repository<Campaign>,
    @InjectRepository(Creative)
    private creativesRepository: Repository<Creative>,
    private configService: ConfigService,
  ) {
    const redisHost = this.configService.get<string>('REDIS_HOST', 'redis');
    const redisPort = this.configService.get<number>('REDIS_PORT', 6379);
    const redisPassword = this.configService.get<string>('REDIS_PASSWORD');
    this.redisClient = new Redis({
      host: redisHost,
      port: redisPort,
      password: redisPassword,
    });
  }

  /**
   * Augments campaign object with real-time spend from Redis.
   */
  private async augmentWithRealTimeSpend(campaign: Campaign): Promise<Campaign> {
    if (campaign.status !== CampaignStatus.ACTIVE) return campaign;

    const dateKey = new Date().toISOString().split('T')[0];
    const spendKey = `campaign:spend:${campaign.id}:${dateKey}`;
    try {
      const dailySpendStr = await this.redisClient.get(spendKey);
      if (dailySpendStr) {
        const dailySpend = parseFloat(dailySpendStr);
        // Add daily spend (Redis) to historical spend (Postgres)
        // We only modify the returned object, not the DB record.
        // Convert to number to handle potential string types from DB driver
        campaign.spent = Number(campaign.spent) + dailySpend;
      }
    } catch (error) {
      console.error(`Failed to fetch real-time spend for campaign ${campaign.id}`, error);
    }
    return campaign;
  }

  /**
   * Sync all ACTIVE campaigns to Redis for the Bidding Engine (Go)
   * Key: 'campaigns:active' (List of Campaign JSON strings)
   */
  async syncActiveCampaigns(): Promise<void> {
    const activeCampaigns = await this.campaignsRepository.find({
      where: { status: CampaignStatus.ACTIVE },
    });

    // The Bidding Engine expects a single JSON array of campaign objects
    // We filter sensitive/unnecessary fields if needed, but for now send full object
    // Assuming Go struct matches JSON usage
    
    // We intentionally store the whole array under one key for simplicity
    // If scale becomes an issue (10k+ campaigns), we should switch to HSET or individual keys + index set.
    // For now (<1000 campaigns), single key is efficient for bulk read.
    
    // Transform dates to strings if necessary (Go time.Time expects RFC3339)
    // TypeORM usually returns Date objects which JSON.stringify handles correctly.
    
    if (activeCampaigns.length > 0) {
      await this.redisClient.set('campaigns:active', JSON.stringify(activeCampaigns));
    } else {
      await this.redisClient.del('campaigns:active');
    }
  }

  async create(createCampaignDto: CreateCampaignDto, userId: string, tenantId: string): Promise<Campaign> {
    const campaign = this.campaignsRepository.create({
      ...createCampaignDto,
      userId,
      tenantId,
      status: createCampaignDto.status || CampaignStatus.DRAFT,
      spent: 0,
      impressions: 0,
      clicks: 0,
      conversions: 0,
    });

    const saved = await this.campaignsRepository.save(campaign);
    
    // Sync if created as active
    if (saved.status === CampaignStatus.ACTIVE) {
      await this.syncActiveCampaigns();
    }
    
    return Array.isArray(saved) ? saved[0] : saved;
  }

  async findAll(tenantId: string, includeRealTime: boolean = false): Promise<Campaign[]> {
    const campaigns = await this.campaignsRepository.find({
      where: { tenantId },
      order: { createdAt: 'DESC' },
    });
    if (includeRealTime) {
      return Promise.all(campaigns.map(c => this.augmentWithRealTimeSpend(c)));
    }
    return campaigns;
  }

  async findByUser(userId: string, tenantId: string, includeRealTime: boolean = false): Promise<Campaign[]> {
    const campaigns = await this.campaignsRepository.find({
      where: { userId, tenantId },
      order: { createdAt: 'DESC' },
    });
    if (includeRealTime) {
      return Promise.all(campaigns.map(c => this.augmentWithRealTimeSpend(c)));
    }
    return campaigns;
  }

  async findOne(id: string, tenantId: string, includeRealTime: boolean = false): Promise<Campaign> {
    const campaign = await this.campaignsRepository.findOne({
      where: { id, tenantId },
    });

    if (!campaign) {
      throw new NotFoundException('Campaign not found');
    }

    if (includeRealTime) {
      return this.augmentWithRealTimeSpend(campaign);
    }
    return campaign;
  }

  /**
   * Internal Use Only: Find campaign by ID without tenant scope.
   * Useful for background jobs (analytics, billing).
   */
  async findById(id: string): Promise<Campaign> {
    const campaign = await this.campaignsRepository.findOne({ where: { id } });
    if (!campaign) {
      throw new NotFoundException('Campaign not found');
    }
    return this.augmentWithRealTimeSpend(campaign);
  }

  async update(id: string, updateData: Partial<Campaign>, tenantId: string): Promise<Campaign> {
    await this.findOne(id, tenantId);

    // Don't allow changing spent amount or stats directly
    delete updateData.spent;
    delete updateData.impressions;
    delete updateData.clicks;
    delete updateData.conversions;

    await this.campaignsRepository.update(id, updateData);
    
    // Always sync on update as status or targeting might match
    await this.syncActiveCampaigns();
    
    return this.findOne(id, tenantId);
  }

  async updateStats(id: string, stats: {
    spent?: number;
    impressions?: number;
    clicks?: number;
    conversions?: number;
  }): Promise<void> {
    await this.campaignsRepository.increment({ id }, 'spent', stats.spent || 0);
    await this.campaignsRepository.increment({ id }, 'impressions', stats.impressions || 0);
    await this.campaignsRepository.increment({ id }, 'clicks', stats.clicks || 0);
    await this.campaignsRepository.increment({ id }, 'conversions', stats.conversions || 0);

    // Auto-pause if budget exceeded
    const campaign = await this.campaignsRepository.findOne({ where: { id } });
    if (campaign && campaign.spent >= campaign.budget && campaign.status === CampaignStatus.ACTIVE) {
      await this.campaignsRepository.update(id, { status: CampaignStatus.COMPLETED });
      // Campaign paused, so sync
      await this.syncActiveCampaigns();
    }
  }

  async remove(id: string, tenantId: string): Promise<void> {
    const campaign = await this.findOne(id, tenantId);
    await this.campaignsRepository.remove(campaign);
    await this.syncActiveCampaigns();
  }

  async getActiveCampaigns(): Promise<Campaign[]> {
    return this.campaignsRepository.find({
      where: { status: CampaignStatus.ACTIVE },
    });
  }

  async assignCreatives(campaignId: string, creativeIds: string[], tenantId: string): Promise<Campaign> {
    const campaign = await this.findOne(campaignId, tenantId);
    const creatives = await this.creativesRepository.findByIds(creativeIds.filter(id => id));

    campaign.creatives = creatives;
    return this.campaignsRepository.save(campaign);
  }

  async removeCreative(campaignId: string, creativeId: string, tenantId: string): Promise<Campaign> {
    const campaign = await this.findOne(campaignId, tenantId);
    campaign.creatives = campaign.creatives.filter(creative => creative.id !== creativeId);
    return this.campaignsRepository.save(campaign);
  }
}
