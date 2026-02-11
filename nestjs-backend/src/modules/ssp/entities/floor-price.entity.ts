import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';

export enum FloorPriceRuleType {
  GLOBAL = 'global',
  GEO = 'geo',
  DEVICE = 'device',
  TIME = 'time',
  FORMAT = 'format',
  CUSTOM = 'custom',
}

export enum FloorPriceAction {
  SET_FLOOR = 'set_floor',
  MULTIPLY = 'multiply',
  ADD = 'add',
}

@Entity('floor_prices')
export class FloorPrice {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  name: string;

  @Column({ nullable: true })
  description: string;

  @Column({ name: 'ruletype', type: 'enum', enum: FloorPriceRuleType, default: FloorPriceRuleType.GLOBAL })
  ruleType: FloorPriceRuleType;

  @Column({ type: 'decimal', precision: 10, scale: 4 })
  price: number;

  @Column({ default: 'USD' })
  currency: string;

  @Column({ type: 'enum', enum: FloorPriceAction, default: FloorPriceAction.SET_FLOOR })
  action: FloorPriceAction;

  @Column({ type: 'json', nullable: true })
  conditions: Record<string, any>;

  @Column({ name: 'publisherid', nullable: true })
  publisherId: string;

  @Column({ name: 'adunitid', nullable: true })
  adUnitId: string;

  @Column({ name: 'placementid', nullable: true })
  placementId: string;

  @Column({ default: 0 })
  priority: number;

  @Column({ name: 'isactive', default: true })
  isActive: boolean;

  @Column({ name: 'startdate', nullable: true })
  startDate: Date;

  @Column({ name: 'enddate', nullable: true })
  endDate: Date;

  @CreateDateColumn({ name: 'createdat' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updatedat' })
  updatedAt: Date;
}
