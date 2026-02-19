import { Entity, Column, PrimaryGeneratedColumn, CreateDateColumn } from 'typeorm';

@Entity('notifications')
export class Notification {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  tenantId: string;

  @Column()
  userId: string; // The user who should receive this (e.g., Advertiser)

  @Column()
  title: string;

  @Column('text')
  message: string;

  @Column({ default: false })
  isRead: boolean;

  @Column({ default: 'info' }) // info, warning, error, success
  type: string;

  @Column({ nullable: true })
  category: string; // budget, campaign, system, etc.

  @CreateDateColumn()
  createdAt: Date;
}
