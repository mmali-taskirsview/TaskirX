import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';

export enum AudienceSegmentStatus {
  ACTIVE = 'active',
  BUILDING = 'building',
  PAUSED = 'paused',
  ARCHIVED = 'archived',
}

export enum AudienceSegmentType {
  FIRST_PARTY = 'first_party',
  THIRD_PARTY = 'third_party',
  LOOKALIKE = 'lookalike',
  CONTEXTUAL = 'contextual',
  RETARGETING = 'retargeting',
  
  // New Types
  DEMOGRAPHIC = 'demographic',
  PSYCHOGRAPHIC = 'psychographic',
  BEHAVIORAL = 'behavioral',
  B2B = 'b2b',
  INTENT = 'intent'
}

@Entity('dsp_audience_segments')
export class AudienceSegment {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({ name: 'name' })
  name: string;

  @Column({ name: 'description', nullable: true })
  description: string;

  @Column({ name: 'type', type: 'varchar', default: AudienceSegmentType.FIRST_PARTY })
  type: AudienceSegmentType;

  @Column({ name: 'status', type: 'varchar', default: AudienceSegmentStatus.BUILDING })
  status: AudienceSegmentStatus;

  @Column({ name: 'advertiser_id', nullable: true })
  advertiserId: string;

  @Column({ name: 'size', type: 'bigint', default: 0 })
  size: number;

  @Column({ name: 'match_rate', type: 'decimal', precision: 5, scale: 2, default: 0 })
  matchRate: number;

  @Column({ name: 'cpm_modifier', type: 'decimal', precision: 5, scale: 2, default: 1.0 })
  cpmModifier: number;

  @Column({ name: 'rules', type: 'jsonb', nullable: true })
  rules: {
    include?: Array<{ field: string; operator: string; value: any }>;
    exclude?: Array<{ field: string; operator: string; value: any }>;
  };

  @Column({ name: 'demographics', type: 'jsonb', nullable: true })
  demographics: {
    ageRanges?: string[];
    genders?: string[];
    incomeRanges?: string[];
    interests?: string[];
  };

  @Column({ name: 'lookback_days', type: 'int', default: 30 })
  lookbackDays: number;

  @Column({ name: 'refresh_frequency', type: 'varchar', default: 'daily' })
  refreshFrequency: string;

  @Column({ name: 'last_refresh', type: 'timestamp', nullable: true })
  lastRefresh: Date;

  @Column({ name: 'campaigns_using', type: 'int', default: 0 })
  campaignsUsing: number;

  @CreateDateColumn({ name: 'created_at' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updated_at' })
  updatedAt: Date;
}
