import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';

export enum DemandPartnerType {
  DSP = 'dsp',
  AD_NETWORK = 'ad_network',
  HEADER_BIDDING = 'header_bidding',
  DIRECT = 'direct',
}

export enum DemandPartnerStatus {
  ACTIVE = 'active',
  PAUSED = 'paused',
  PENDING = 'pending',
  DISCONNECTED = 'disconnected',
}

@Entity('demand_partners')
export class DemandPartner {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  name: string;

  @Column({ unique: true })
  code: string; // google_adx, prebid, amazon_tam, etc.

  @Column({ type: 'enum', enum: DemandPartnerType })
  type: DemandPartnerType;

  @Column({ type: 'enum', enum: DemandPartnerStatus, default: DemandPartnerStatus.PENDING })
  status: DemandPartnerStatus;

  @Column({ nullable: true, name: 'publisherid' })
  publisherId: string;

  @Column({ nullable: true })
  endpoint: string;

  @Column({ type: 'json', nullable: true })
  credentials: Record<string, any>;

  @Column({ type: 'json', nullable: true })
  settings: Record<string, any>;

  @Column({ type: 'decimal', precision: 5, scale: 2, default: 0, name: 'revenueshare' })
  revenueShare: number;

  @Column({ default: 100, name: 'bidtimeout' })
  bidTimeout: number; // milliseconds

  @Column({ default: true, name: 'isglobal' })
  isGlobal: boolean; // available to all publishers

  @Column({ type: 'bigint', default: 0, name: 'totalimpressions' })
  totalImpressions: number;

  @Column({ type: 'decimal', precision: 15, scale: 2, default: 0, name: 'totalrevenue' })
  totalRevenue: number;

  @Column({ type: 'decimal', precision: 5, scale: 2, default: 0, name: 'winrate' })
  winRate: number;

  @Column({ type: 'decimal', precision: 10, scale: 4, default: 0, name: 'avgbidprice' })
  avgBidPrice: number;

  @CreateDateColumn({ name: 'createdat' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updatedat' })
  updatedAt: Date;
}
