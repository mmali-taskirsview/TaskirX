import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Campaign, CampaignStatus } from './campaign.entity';
import { CreateCampaignDto } from './dto/create-campaign.dto';

@Injectable()
export class CampaignsService {
  constructor(
    @InjectRepository(Campaign)
    private campaignsRepository: Repository<Campaign>,
  ) {}

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
    return Array.isArray(saved) ? saved[0] : saved;
  }

  async findAll(tenantId: string): Promise<Campaign[]> {
    return this.campaignsRepository.find({
      where: { tenantId },
      order: { createdAt: 'DESC' },
    });
  }

  async findByUser(userId: string, tenantId: string): Promise<Campaign[]> {
    return this.campaignsRepository.find({
      where: { userId, tenantId },
      order: { createdAt: 'DESC' },
    });
  }

  async findOne(id: string, tenantId: string): Promise<Campaign> {
    const campaign = await this.campaignsRepository.findOne({
      where: { id, tenantId },
    });

    if (!campaign) {
      throw new NotFoundException('Campaign not found');
    }

    return campaign;
  }

  async update(id: string, updateData: Partial<Campaign>, tenantId: string): Promise<Campaign> {
    await this.findOne(id, tenantId);

    // Don't allow changing spent amount or stats directly
    delete updateData.spent;
    delete updateData.impressions;
    delete updateData.clicks;
    delete updateData.conversions;

    await this.campaignsRepository.update(id, updateData);
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
    if (campaign && campaign.spent >= campaign.budget) {
      await this.campaignsRepository.update(id, { status: CampaignStatus.COMPLETED });
    }
  }

  async remove(id: string, tenantId: string): Promise<void> {
    const campaign = await this.findOne(id, tenantId);
    await this.campaignsRepository.remove(campaign);
  }

  async getActiveCampaigns(): Promise<Campaign[]> {
    return this.campaignsRepository.find({
      where: { status: CampaignStatus.ACTIVE },
    });
  }
}
