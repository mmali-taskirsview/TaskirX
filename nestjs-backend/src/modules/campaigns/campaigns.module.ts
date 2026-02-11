import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { Campaign } from './campaign.entity';
import { CampaignsService } from './campaigns.service';
import { CampaignsController } from './campaigns.controller';
import { CampaignsInternalController } from './campaigns.internal.controller';

@Module({
  imports: [TypeOrmModule.forFeature([Campaign])],
  providers: [CampaignsService],
  controllers: [CampaignsController, CampaignsInternalController],
  exports: [CampaignsService],
})
export class CampaignsModule {}
