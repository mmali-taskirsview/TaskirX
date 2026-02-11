import {
  Column,
  CreateDateColumn,
  Entity,
  Index,
  PrimaryGeneratedColumn,
  UpdateDateColumn,
} from 'typeorm';

@Entity('integration_configs')
@Index(['tenantId', 'integrationKey'], { unique: true })
export class IntegrationConfig {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  tenantId: string;

  @Column()
  integrationKey: string;

  @Column()
  status: string;

  @Column({ type: 'jsonb', nullable: true })
  configData: Record<string, any> | null;

  @Column({ type: 'jsonb', nullable: true })
  missingFields: string[] | null;

  @Column({ type: 'jsonb', nullable: true })
  providedFields: string[] | null;

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;
}
