import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  ManyToOne,
  JoinColumn,
} from 'typeorm';
import { Publisher } from './publisher.entity';

export enum AdUnitStatus {
  ACTIVE = 'active',
  PAUSED = 'paused',
  PENDING = 'pending',
  REJECTED = 'rejected',
}

export enum AdUnitType {
  // Display
  BANNER = 'banner',
  RICH_MEDIA = 'rich_media',
  
  // Video
  VIDEO_INSTREAM = 'video_instream',
  VIDEO_OUTSTREAM = 'video_outstream',
  CTV = 'ctv',

  // Native
  NATIVE = 'native',
  CONTENT_RECOMMENDATION = 'content_recommendation',

  // Mobile
  INTERSTITIAL = 'interstitial',
  REWARDED = 'rewarded',
  PLAYABLE = 'playable',
  
  // Audio
  AUDIO_DIGITAL = 'audio_digital',
  AUDIO_PROGRAMMATIC = 'audio_programmatic',

  // Emerging
  DCO = 'dco',
  VR_AR = 'vr_ar',
  IN_GAME = 'in_game',

  // Performance
  PUSH = 'push',
  POPUNDER = 'popunder'
}

@Entity('ad_units')
export class AdUnit {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  name: string;

  @Column({ type: 'enum', enum: AdUnitType })
  type: AdUnitType;

  @Column()
  size: string; // e.g., '300x250', '728x90', etc.

  @Column({ type: 'enum', enum: AdUnitStatus, default: AdUnitStatus.PENDING })
  status: AdUnitStatus;

  @Column({ nullable: true })
  domain: string;

  @Column({ nullable: true, name: 'pageurl' })
  pageUrl: string;

  @Column({ type: 'decimal', precision: 10, scale: 4, nullable: true, name: 'floorprice' })
  floorPrice: number;

  @Column({ default: 'USD' })
  currency: string;

  @Column({ type: 'simple-array', nullable: true, name: 'allowedcategories' })
  allowedCategories: string[];

  @Column({ type: 'simple-array', nullable: true, name: 'blockedcategories' })
  blockedCategories: string[];

  @Column({ type: 'simple-array', nullable: true, name: 'blockedadvertisers' })
  blockedAdvertisers: string[];

  @Column({ type: 'json', nullable: true })
  targeting: Record<string, any>;

  @Column({ type: 'json', nullable: true })
  settings: Record<string, any>;

  @Column({ type: 'bigint', default: 0 })
  impressions: number;

  @Column({ type: 'bigint', default: 0 })
  requests: number;

  @Column({ type: 'decimal', precision: 15, scale: 2, default: 0 })
  revenue: number;

  @ManyToOne(() => Publisher, (publisher) => publisher.adUnits)
  @JoinColumn({ name: 'publisherid' })
  publisher: Publisher;

  @Column({ name: 'publisherid' })
  publisherId: string;

  @CreateDateColumn({ name: 'createdat' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updatedat' })
  updatedAt: Date;
}
