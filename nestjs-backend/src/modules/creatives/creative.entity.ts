import {
  Entity,
  Column,
  PrimaryGeneratedColumn,
  CreateDateColumn,
  UpdateDateColumn,
  ManyToOne,
  JoinColumn,
  ManyToMany,
} from 'typeorm';
import { Campaign } from '../campaigns/campaign.entity';

export enum CreativeType {
  IMAGE = 'image',
  VIDEO = 'video',
  HTML5 = 'html5',
  PLAYABLE = 'playable',
}

export enum CreativeStatus {
  DRAFT = 'draft',
  PENDING = 'pending',
  ACTIVE = 'active',
  REJECTED = 'rejected',
  ARCHIVED = 'archived',
}

export enum CreativeFormat {
  BANNER = 'banner',
  INTERSTITIAL = 'interstitial',
  REWARDED_VIDEO = 'rewarded_video',
  NATIVE = 'native',
  PLAYABLE = 'playable',
  RICH_MEDIA = 'rich_media',
}

@Entity('creatives')
export class Creative {
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

  @Column({ type: 'enum', enum: CreativeType })
  type: CreativeType;

  @Column({ type: 'enum', enum: CreativeFormat })
  format: CreativeFormat;

  @Column()
  url: string;

  @Column({ nullable: true })
  thumbnailUrl: string;

  @Column({ type: 'jsonb', nullable: true })
  dimensions: {
    width: number;
    height: number;
  };

  @Column({ type: 'decimal', precision: 10, scale: 2, default: 0 })
  fileSize: number; // in MB

  @Column({ type: 'enum', enum: CreativeStatus, default: CreativeStatus.DRAFT })
  status: CreativeStatus;

  @Column({ type: 'jsonb', nullable: true })
  metadata: {
    duration?: number; // for videos in seconds
    clickUrl?: string;
    impressionTrackingUrls?: string[];
    clickTrackingUrls?: string[];
    thirdPartyTags?: string[];
  };

  @Column({ type: 'jsonb', nullable: true })
  targeting: {
    categories?: string[];
    ageGroups?: string[];
    genders?: string[];
    devices?: string[];
  };

  @Column({ type: 'simple-array', nullable: true })
  tags: string[];

  // Performance metrics
  @Column({ type: 'bigint', default: 0 })
  impressions: number;

  @Column({ type: 'bigint', default: 0 })
  clicks: number;

  @Column({ type: 'bigint', default: 0 })
  conversions: number;

  @Column({ type: 'decimal', precision: 5, scale: 4, default: 0 })
  ctr: number;

  @Column({ type: 'decimal', precision: 5, scale: 4, default: 0 })
  cvr: number;

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;

  // Relationship to campaigns
  @ManyToMany(() => Campaign)
  campaigns: Campaign[];
}