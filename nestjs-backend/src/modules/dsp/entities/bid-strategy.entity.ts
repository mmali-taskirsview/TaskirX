import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';

export enum BidStrategyType {
  MANUAL = 'manual',
  AUTO_OPTIMIZE = 'auto_optimize',
  TARGET_CPA = 'target_cpa',
  TARGET_ROAS = 'target_roas',
  MAXIMIZE_CONVERSIONS = 'maximize_conversions',
  MAXIMIZE_CLICKS = 'maximize_clicks',
}

export enum BidStrategyStatus {
  ACTIVE = 'active',
  PAUSED = 'paused',
  LEARNING = 'learning',
}

@Entity('dsp_bid_strategies')
export class BidStrategy {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({ name: 'name' })
  name: string;

  @Column({ name: 'type', type: 'varchar', default: BidStrategyType.MANUAL })
  type: BidStrategyType;

  @Column({ name: 'status', type: 'varchar', default: BidStrategyStatus.ACTIVE })
  status: BidStrategyStatus;

  @Column({ name: 'advertiser_id', nullable: true })
  advertiserId: string;

  @Column({ name: 'campaign_id', nullable: true })
  campaignId: string;

  @Column({ name: 'base_bid', type: 'decimal', precision: 10, scale: 4, default: 1.00 })
  baseBid: number;

  @Column({ name: 'max_bid', type: 'decimal', precision: 10, scale: 4, default: 10.00 })
  maxBid: number;

  @Column({ name: 'min_bid', type: 'decimal', precision: 10, scale: 4, default: 0.10 })
  minBid: number;

  @Column({ name: 'target_cpa', type: 'decimal', precision: 10, scale: 2, nullable: true })
  targetCpa: number;

  @Column({ name: 'target_roas', type: 'decimal', precision: 5, scale: 2, nullable: true })
  targetRoas: number;

  @Column({ name: 'bid_adjustments', type: 'jsonb', nullable: true })
  bidAdjustments: {
    device?: { mobile: number; desktop: number; tablet: number };
    geo?: Record<string, number>;
    time?: Record<string, number>;
    audience?: Record<string, number>;
  };

  @Column({ name: 'frequency_cap', type: 'jsonb', nullable: true })
  frequencyCap: {
    impressions: number;
    period: 'hour' | 'day' | 'week' | 'month';
    perUser: boolean;
  };

  @Column({ name: 'pacing', type: 'jsonb', nullable: true })
  pacing: {
    type: 'even' | 'accelerated' | 'front_loaded';
    dailyBudget?: number;
    hourlyBudget?: number;
  };

  @Column({ name: 'performance_metrics', type: 'jsonb', nullable: true })
  performanceMetrics: {
    avgCpm: number;
    avgCpc: number;
    avgCpa: number;
    winRate: number;
    conversions: number;
  };

  @Column({ name: 'learning_status', type: 'varchar', nullable: true })
  learningStatus: string;

  @Column({ name: 'campaigns_count', type: 'int', default: 0 })
  campaignsCount: number;

  @CreateDateColumn({ name: 'created_at' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updated_at' })
  updatedAt: Date;
}
