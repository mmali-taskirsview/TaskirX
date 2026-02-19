import {
  Entity,
  Column,
  PrimaryGeneratedColumn,
  CreateDateColumn,
  UpdateDateColumn,
  ManyToMany,
  JoinTable,
} from 'typeorm';
import { Creative } from '../creatives/creative.entity';

export class ColumnNumericTransformer {
  to(data: number): number {
    return data;
  }
  from(data: string): number {
    return parseFloat(data);
  }
}

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

  @Column({ type: 'decimal', precision: 10, scale: 2, transformer: new ColumnNumericTransformer() })
  budget: number;

  @Column({ type: 'decimal', precision: 10, scale: 2, default: 0, transformer: new ColumnNumericTransformer() })
  spent: number;

  @Column({ type: 'decimal', precision: 10, scale: 2, nullable: true, transformer: new ColumnNumericTransformer() })
  dailyBudget: number; // Daily spend cap (if null, defaults to budget/30)

  @Column({ type: 'varchar', length: 20, default: 'even' })
  pacingStrategy: string; // 'asap', 'even', 'front', 'back'

  @Column({ type: 'int', default: 5 })
  priority: number; // Priority level 1-10 (higher = more important, default: 5)

  // Goal-Based Pacing
  @Column({ type: 'varchar', length: 20, nullable: true })
  goalType: string; // 'impressions', 'clicks', 'conversions'

  @Column({ type: 'bigint', default: 0 })
  goalTarget: number; // Target number of events

  @Column({ type: 'bigint', default: 0 })
  goalDelivered: number; // Current progress

  @Column({ type: 'date', nullable: true })
  goalEndDate: Date; // Target delivery date (YYYY-MM-DD)

  @Column({ type: 'decimal', precision: 10, scale: 4, transformer: new ColumnNumericTransformer() })
  bidPrice: number; // Price per impression/click/action

  @Column({ type: 'varchar', nullable: true })
  vertical: string; // Industry Vertical (e.g., 'GAMING', 'FINANCE')

  @Column({ type: 'jsonb', nullable: true })
  targeting: {
    // Geo
    geoMarkets?: string[]; // Tier 1, 2, 3 or specific countries
    geoFences?: {
      lat: number;
      lon: number;
      radius: number; // in km
      name?: string;
    }[];
    
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
    
    // Dayparting
    hourSchedule?: number[]; // Hours 0-23 when campaign is active (empty = always active)
  };

  @Column({ type: 'timestamp', nullable: true })
  startDate: Date;

  @Column({ type: 'timestamp', nullable: true })
  endDate: Date;

  @Column({ type: 'jsonb', nullable: true })
  creative: {
    type: string; // 'banner', 'video', 'native', 'rich_media', 'audio', 'playable', 'pop', 'push'
    url?: string; // Main asset URL (banner image, video file, main image for native)
    width?: number;
    height?: number;
    
    // Video/Audio specific
    duration?: number; // seconds
    mimeType?: string; // 'video/mp4', 'audio/mp3'
    bitrate?: number;  // Audio bitrate
    
    // Native/Push specific
    title?: string;
    description?: string;
    iconUrl?: string; // Icon image for native ad/push icon
    ctaText?: string; // Call to Action text
    
    // Rich Media / Playable / Pop specific
    htmlSnippet?: string; // Custom HTML/JS payload
    expandable?: boolean; // Rich media expandable property
  };

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

  @Column({ type: 'varchar', nullable: true })
  dealId: string; // Private Marketplace Deal ID

  // Relationship to creatives
  @ManyToMany(() => Creative)
  @JoinTable({
    name: 'campaign_creatives',
    joinColumn: {
      name: 'campaignId',
      referencedColumnName: 'id',
    },
    inverseJoinColumn: {
      name: 'creativeId',
      referencedColumnName: 'id',
    },
  })
  creatives: Creative[];
}
