import {
  Entity,
  Column,
  PrimaryGeneratedColumn,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';

export enum CampaignStatus {
  DRAFT = 'draft',
  ACTIVE = 'active',
  PAUSED = 'paused',
  COMPLETED = 'completed',
}

export enum CampaignType {
  CPM = 'cpm', // Cost per thousand impressions
  CPC = 'cpc', // Cost per click
  CPA = 'cpa', // Cost per action
}

@Entity('campaigns')
export class Campaign {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  name: string;

  @Column({ type: 'text', nullable: true })
  description: string;

  @Column({ type: 'uuid' })
  userId: string;

  @Column({ type: 'uuid' })
  tenantId: string;

  @Column({ type: 'enum', enum: CampaignStatus, default: CampaignStatus.DRAFT })
  status: CampaignStatus;

  @Column({ type: 'enum', enum: CampaignType })
  type: CampaignType;

  @Column({ type: 'decimal', precision: 10, scale: 2 })
  budget: number;

  @Column({ type: 'decimal', precision: 10, scale: 2, default: 0 })
  spent: number;

  @Column({ type: 'decimal', precision: 10, scale: 4 })
  bidPrice: number; // Price per impression/click/action

  @Column({ type: 'varchar', nullable: true })
  vertical: string; // Industry Vertical (e.g., 'GAMING', 'FINANCE')

  @Column({ type: 'jsonb', nullable: true })
  targeting: {
    // Geo
    geoMarkets?: string[]; // Tier 1, 2, 3 or specific countries
    
    // Demographics
    ageGroups?: string[];
    genders?: string[];
    incomeLevels?: string[];
    educationLevels?: string[];

    // Psychographics
    lifestyles?: string[];
    interests?: string[];
    values?: string[];

    // Behavior
    onlineBehaviors?: string[];
    purchaseBehaviors?: string[];
    
    // Tech
    devices?: string[];
    os?: string[];
    connectionTypes?: string[];

    // Legacy/Simple
    countries?: string[];
    categories?: string[];
    minAge?: number;
    maxAge?: number;
  };

  @Column({ type: 'timestamp', nullable: true })
  startDate: Date;

  @Column({ type: 'timestamp', nullable: true })
  endDate: Date;

  @Column({ type: 'bigint', default: 0 })
  impressions: number;

  @Column({ type: 'bigint', default: 0 })
  clicks: number;

  @Column({ type: 'bigint', default: 0 })
  conversions: number;

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;
}
