import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';

export enum BrandSafetyRuleType {
  BLOCKLIST = 'blocklist',
  ALLOWLIST = 'allowlist',
  CATEGORY_BLOCK = 'category_block',
  KEYWORD_BLOCK = 'keyword_block',
}

export enum BrandSafetyTarget {
  ADVERTISER = 'advertiser',
  DOMAIN = 'domain',
  CATEGORY = 'category',
  KEYWORD = 'keyword',
}

@Entity('brand_safety_rules')
export class BrandSafetyRule {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  name: string;

  @Column({ nullable: true })
  description: string;

  @Column({ name: 'ruletype', type: 'enum', enum: BrandSafetyRuleType })
  ruleType: BrandSafetyRuleType;

  @Column({ type: 'enum', enum: BrandSafetyTarget })
  target: BrandSafetyTarget;

  @Column({ type: 'simple-array' })
  values: string[]; // list of domains, categories, keywords, etc.

  @Column({ name: 'publisherid' })
  publisherId: string;

  @Column({ name: 'adunitid', nullable: true })
  adUnitId: string;

  @Column({ name: 'isactive', default: true })
  isActive: boolean;

  @Column({ default: 0 })
  priority: number;

  @Column({ name: 'matchcount', type: 'bigint', default: 0 })
  matchCount: number;

  @CreateDateColumn({ name: 'createdat' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updatedat' })
  updatedAt: Date;
}
