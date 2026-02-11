import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { DspController } from './dsp.controller';
import { RetargetingController } from './retargeting.controller';
import { DspService } from './dsp.service';
import { SupplyPartner } from './entities/supply-partner.entity';
import { AudienceSegment } from './entities/audience-segment.entity';
import { Deal } from './entities/deal.entity';
import { BidStrategy } from './entities/bid-strategy.entity';

@Module({
  imports: [
    TypeOrmModule.forFeature([
      SupplyPartner,
      AudienceSegment,
      Deal,
      BidStrategy,
    ]),
  ],
  controllers: [DspController, RetargetingController],
  providers: [DspService],
  exports: [DspService],
})
export class DspModule {}
