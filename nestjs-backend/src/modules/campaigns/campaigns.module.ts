import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { Campaign } from './campaign.entity';
import { Creative } from '../creatives/creative.entity';
import { CampaignsService } from './campaigns.service';
import { CampaignsController } from './campaigns.controller';
import { CampaignsInternalController } from './campaigns.internal.controller';
import { RedisModule } from '../redis/redis.module';

@Module({
  imports: [
    TypeOrmModule.forFeature([Campaign, Creative]),
    RedisModule,
  ],
  providers: [CampaignsService],
  controllers: [CampaignsController, CampaignsInternalController],
  exports: [CampaignsService],
})
export class CampaignsModule {}
