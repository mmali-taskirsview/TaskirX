import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';

export enum SupplyPartnerStatus {
  ACTIVE = 'active',
  PAUSED = 'paused',
  PENDING = 'pending',
  BLOCKED = 'blocked',
}

export enum SupplyPartnerType {
  SSP = 'ssp',
  EXCHANGE = 'exchange',
  DIRECT = 'direct',
}

@Entity('dsp_supply_partners')
export class SupplyPartner {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({ name: 'name' })
  name: string;

  @Column({ name: 'code', unique: true })
  code: string;

  @Column({ name: 'type', type: 'varchar', default: SupplyPartnerType.SSP })
  type: SupplyPartnerType;

  @Column({ name: 'endpoint_url' })
  endpointUrl: string;

  @Column({ name: 'status', type: 'varchar', default: SupplyPartnerStatus.PENDING })
  status: SupplyPartnerStatus;

  @Column({ name: 'qps_limit', type: 'int', default: 1000 })
  qpsLimit: number;

  @Column({ name: 'timeout_ms', type: 'int', default: 100 })
  timeoutMs: number;

  @Column({ name: 'min_bid', type: 'decimal', precision: 10, scale: 4, default: 0.01 })
  minBid: number;

  @Column({ name: 'max_bid', type: 'decimal', precision: 10, scale: 4, default: 50.00 })
  maxBid: number;

  @Column({ name: 'daily_budget', type: 'decimal', precision: 12, scale: 2, nullable: true })
  dailyBudget: number;

  @Column({ name: 'daily_spend', type: 'decimal', precision: 12, scale: 2, default: 0 })
  dailySpend: number;

  @Column({ name: 'total_requests', type: 'bigint', default: 0 })
  totalRequests: number;

  @Column({ name: 'total_bids', type: 'bigint', default: 0 })
  totalBids: number;

  @Column({ name: 'total_wins', type: 'bigint', default: 0 })
  totalWins: number;

  @Column({ name: 'total_spend', type: 'decimal', precision: 14, scale: 2, default: 0 })
  totalSpend: number;

  @Column({ name: 'supported_formats', type: 'simple-array', nullable: true })
  supportedFormats: string[];

  @Column({ name: 'geo_targets', type: 'simple-array', nullable: true })
  geoTargets: string[];

  @Column({ name: 'config', type: 'jsonb', nullable: true })
  config: Record<string, any>;

  @CreateDateColumn({ name: 'created_at' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updated_at' })
  updatedAt: Date;
}
