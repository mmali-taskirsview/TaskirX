import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { AnalyticsService } from './analytics.service';
import { AnalyticsController } from './analytics.controller';
import { TrackingController } from './tracking.controller';
import { MmpController } from './mmp.controller';
import { Campaign } from '../campaigns/campaign.entity';
import { UsersModule } from '../users/users.module';
import { NotificationsModule } from '../notifications/notifications.module';
import { CampaignsModule } from '../campaigns/campaigns.module';

@Module({
  imports: [
    TypeOrmModule.forFeature([Campaign]),
    UsersModule, // For Admin fraud alerts
    NotificationsModule,
    CampaignsModule,
  ],
  providers: [AnalyticsService],
  controllers: [AnalyticsController, TrackingController, MmpController],
  exports: [AnalyticsService],
})
export class AnalyticsModule {}
