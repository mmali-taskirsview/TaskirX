import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';

export enum DealStatus {
  ACTIVE = 'active',
  PAUSED = 'paused',
  PENDING = 'pending',
  EXPIRED = 'expired',
  REJECTED = 'rejected',
}

export enum DealType {
  PREFERRED = 'preferred',
  PRIVATE_AUCTION = 'private_auction',
  PROGRAMMATIC_GUARANTEED = 'programmatic_guaranteed',
  OPEN_AUCTION = 'open_auction',
}

@Entity('dsp_deals')
export class Deal {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({ name: 'deal_id', unique: true })
  dealId: string;

  @Column({ name: 'name' })
  name: string;

  @Column({ name: 'description', nullable: true })
  description: string;

  @Column({ name: 'type', type: 'varchar', default: DealType.PREFERRED })
  type: DealType;

  @Column({ name: 'status', type: 'varchar', default: DealStatus.PENDING })
  status: DealStatus;

  @Column({ name: 'advertiser_id', nullable: true })
  advertiserId: string;

  @Column({ name: 'publisher_id', nullable: true })
  publisherId: string;

  @Column({ name: 'publisher_name', nullable: true })
  publisherName: string;

  @Column({ name: 'floor_price', type: 'decimal', precision: 10, scale: 4, default: 0 })
  floorPrice: number;

  @Column({ name: 'fixed_price', type: 'decimal', precision: 10, scale: 4, nullable: true })
  fixedPrice: number;

  @Column({ name: 'budget', type: 'decimal', precision: 14, scale: 2, nullable: true })
  budget: number;

  @Column({ name: 'spent', type: 'decimal', precision: 14, scale: 2, default: 0 })
  spent: number;

  @Column({ name: 'impressions_goal', type: 'bigint', nullable: true })
  impressionsGoal: number;

  @Column({ name: 'impressions_delivered', type: 'bigint', default: 0 })
  impressionsDelivered: number;

  @Column({ name: 'start_date', type: 'date' })
  startDate: Date;

  @Column({ name: 'end_date', type: 'date' })
  endDate: Date;

  @Column({ name: 'inventory', type: 'jsonb', nullable: true })
  inventory: {
    adFormats?: string[];
    placements?: string[];
    geos?: string[];
    devices?: string[];
  };

  @Column({ name: 'priority', type: 'int', default: 5 })
  priority: number;

  @Column({ name: 'win_rate', type: 'decimal', precision: 5, scale: 2, default: 0 })
  winRate: number;

  @Column({ name: 'avg_cpm', type: 'decimal', precision: 10, scale: 4, default: 0 })
  avgCpm: number;

  @CreateDateColumn({ name: 'created_at' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updated_at' })
  updatedAt: Date;
}
