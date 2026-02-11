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

export enum PlacementStatus {
  ACTIVE = 'active',
  PAUSED = 'paused',
  ARCHIVED = 'archived',
}

export enum PlacementPosition {
  ABOVE_FOLD = 'above_fold',
  BELOW_FOLD = 'below_fold',
  SIDEBAR = 'sidebar',
  IN_CONTENT = 'in_content',
  FOOTER = 'footer',
  HEADER = 'header',
  STICKY = 'sticky',
}

@Entity('placements')
export class Placement {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  name: string;

  @Column({ nullable: true })
  description: string;

  @Column({ type: 'enum', enum: PlacementStatus, default: PlacementStatus.ACTIVE })
  status: PlacementStatus;

  @Column({ type: 'enum', enum: PlacementPosition, nullable: true })
  position: PlacementPosition;

  @Column({ nullable: true })
  domain: string;

  @Column({ nullable: true, name: 'pagetype' })
  pageType: string; // homepage, article, category, etc.

  @Column({ type: 'simple-array', nullable: true, name: 'adformats' })
  adFormats: string[]; // banner, video, native

  @Column({ type: 'simple-array', nullable: true, name: 'allowedsizes' })
  allowedSizes: string[]; // 300x250, 728x90, etc.

  @Column({ type: 'simple-array', nullable: true })
  devices: string[]; // desktop, mobile, tablet

  @Column({ type: 'decimal', precision: 10, scale: 4, nullable: true, name: 'floorprice' })
  floorPrice: number;

  @Column({ default: 'USD' })
  currency: string;

  @Column({ type: 'decimal', precision: 5, scale: 2, nullable: true, name: 'viewabilitytarget' })
  viewabilityTarget: number;

  @Column({ type: 'json', nullable: true })
  targeting: Record<string, any>;

  @Column({ type: 'json', nullable: true })
  settings: Record<string, any>;

  @Column({ type: 'bigint', default: 0 })
  impressions: number;

  @Column({ type: 'decimal', precision: 15, scale: 2, default: 0 })
  revenue: number;

  @Column({ type: 'decimal', precision: 5, scale: 2, default: 0, name: 'fillrate' })
  fillRate: number;

  @ManyToOne(() => Publisher, (publisher) => publisher.placements)
  @JoinColumn({ name: 'publisherid' })
  publisher: Publisher;

  @Column({ name: 'publisherid' })
  publisherId: string;

  @CreateDateColumn({ name: 'createdat' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updatedat' })
  updatedAt: Date;
}
