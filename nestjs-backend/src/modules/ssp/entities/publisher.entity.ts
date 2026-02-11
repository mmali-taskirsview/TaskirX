import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  OneToMany,
} from 'typeorm';
import { AdUnit } from './ad-unit.entity';
import { Placement } from './placement.entity';

export enum PublisherStatus {
  PENDING = 'pending',
  ACTIVE = 'active',
  SUSPENDED = 'suspended',
  REJECTED = 'rejected',
}

export enum PublisherTier {
  BASIC = 'basic',
  STANDARD = 'standard',
  PREMIUM = 'premium',
  ENTERPRISE = 'enterprise',
}

@Entity('publishers')
export class Publisher {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  name: string;

  @Column({ unique: true })
  email: string;

  @Column({ nullable: true, name: 'companyname' })
  companyName: string;

  @Column({ nullable: true })
  website: string;

  @Column({ type: 'simple-array', nullable: true })
  domains: string[];

  @Column({ type: 'enum', enum: PublisherStatus, default: PublisherStatus.PENDING })
  status: PublisherStatus;

  @Column({ type: 'enum', enum: PublisherTier, default: PublisherTier.BASIC })
  tier: PublisherTier;

  @Column({ type: 'decimal', precision: 5, scale: 2, default: 80, name: 'revenueshare' })
  revenueShare: number;

  @Column({ type: 'decimal', precision: 10, scale: 2, default: 0 })
  balance: number;

  @Column({ type: 'decimal', precision: 10, scale: 2, default: 100, name: 'payoutthreshold' })
  payoutThreshold: number;

  @Column({ nullable: true, name: 'paymentmethod' })
  paymentMethod: string;

  @Column({ type: 'json', nullable: true, name: 'paymentdetails' })
  paymentDetails: Record<string, any>;

  @Column({ nullable: true, name: 'apikey' })
  apiKey: string;

  @Column({ type: 'json', nullable: true })
  settings: Record<string, any>;

  @Column({ type: 'json', nullable: true, name: 'brandsafetysettings' })
  brandSafetySettings: Record<string, any>;

  @OneToMany(() => AdUnit, (adUnit) => adUnit.publisher)
  adUnits: AdUnit[];

  @OneToMany(() => Placement, (placement) => placement.publisher)
  placements: Placement[];

  @CreateDateColumn({ name: 'createdat' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updatedat' })
  updatedAt: Date;
}
