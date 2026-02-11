import { Injectable, NotFoundException, ConflictException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Publisher, PublisherStatus, PublisherTier } from './entities/publisher.entity';
import { v4 as uuidv4 } from 'uuid';

export interface CreatePublisherDto {
  name: string;
  email: string;
  companyName?: string;
  website?: string;
  domains?: string[];
}

export interface UpdatePublisherDto {
  name?: string;
  companyName?: string;
  website?: string;
  domains?: string[];
  status?: PublisherStatus;
  tier?: PublisherTier;
  revenueShare?: number;
  payoutThreshold?: number;
  paymentMethod?: string;
  paymentDetails?: Record<string, any>;
  settings?: Record<string, any>;
  brandSafetySettings?: Record<string, any>;
}

@Injectable()
export class PublisherService {
  constructor(
    @InjectRepository(Publisher)
    private publisherRepository: Repository<Publisher>,
  ) {}

  async findAll(status?: PublisherStatus): Promise<Publisher[]> {
    const where = status ? { status } : {};
    return this.publisherRepository.find({
      where,
      order: { createdAt: 'DESC' },
    });
  }

  async findOne(id: string): Promise<Publisher> {
    const publisher = await this.publisherRepository.findOne({
      where: { id },
      relations: ['adUnits', 'placements'],
    });
    
    if (!publisher) {
      throw new NotFoundException(`Publisher with ID ${id} not found`);
    }
    
    return publisher;
  }

  async findByEmail(email: string): Promise<Publisher | null> {
    return this.publisherRepository.findOne({ where: { email } });
  }

  async create(createDto: CreatePublisherDto): Promise<Publisher> {
    // Check if email already exists
    const existing = await this.findByEmail(createDto.email);
    if (existing) {
      throw new ConflictException('Publisher with this email already exists');
    }

    const publisher = this.publisherRepository.create({
      ...createDto,
      apiKey: this.generateApiKey(),
      status: PublisherStatus.PENDING,
    });

    return this.publisherRepository.save(publisher);
  }

  async update(id: string, updateDto: UpdatePublisherDto): Promise<Publisher> {
    const publisher = await this.findOne(id);
    Object.assign(publisher, updateDto);
    return this.publisherRepository.save(publisher);
  }

  async delete(id: string): Promise<void> {
    const publisher = await this.findOne(id);
    await this.publisherRepository.remove(publisher);
  }

  async approve(id: string): Promise<Publisher> {
    return this.update(id, { status: PublisherStatus.ACTIVE });
  }

  async suspend(id: string): Promise<Publisher> {
    return this.update(id, { status: PublisherStatus.SUSPENDED });
  }

  async reject(id: string): Promise<Publisher> {
    return this.update(id, { status: PublisherStatus.REJECTED });
  }

  async regenerateApiKey(id: string): Promise<Publisher> {
    const publisher = await this.findOne(id);
    publisher.apiKey = this.generateApiKey();
    return this.publisherRepository.save(publisher);
  }

  async validateApiKey(apiKey: string): Promise<Publisher | null> {
    return this.publisherRepository.findOne({
      where: { apiKey, status: PublisherStatus.ACTIVE },
    });
  }

  async updateBalance(id: string, amount: number): Promise<Publisher> {
    const publisher = await this.findOne(id);
    publisher.balance = Number(publisher.balance) + amount;
    return this.publisherRepository.save(publisher);
  }

  async getStatistics(id: string): Promise<{
    totalRevenue: number;
    totalImpressions: number;
    totalRequests: number;
    fillRate: number;
    ecpm: number;
  }> {
    const publisher = await this.publisherRepository.findOne({
      where: { id },
      relations: ['adUnits'],
    });

    if (!publisher || !publisher.adUnits) {
      return {
        totalRevenue: 0,
        totalImpressions: 0,
        totalRequests: 0,
        fillRate: 0,
        ecpm: 0,
      };
    }

    const totalRevenue = publisher.adUnits.reduce((sum, u) => sum + Number(u.revenue), 0);
    const totalImpressions = publisher.adUnits.reduce((sum, u) => sum + Number(u.impressions), 0);
    const totalRequests = publisher.adUnits.reduce((sum, u) => sum + Number(u.requests), 0);
    const fillRate = totalRequests > 0 ? (totalImpressions / totalRequests) * 100 : 0;
    const ecpm = totalImpressions > 0 ? (totalRevenue / totalImpressions) * 1000 : 0;

    return {
      totalRevenue,
      totalImpressions,
      totalRequests,
      fillRate,
      ecpm,
    };
  }

  private generateApiKey(): string {
    return `pub_live_sk_${uuidv4().replace(/-/g, '')}`;
  }
}
