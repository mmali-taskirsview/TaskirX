import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { SspController } from './ssp.controller';
import { SspService } from './ssp.service';
import { PublisherController } from './publisher.controller';
import { PublisherService } from './publisher.service';
import { InventoryController } from './inventory.controller';
import { InventoryService } from './inventory.service';
import { DemandPartnerController } from './demand-partner.controller';
import { DemandPartnerService } from './demand-partner.service';
import { Publisher } from './entities/publisher.entity';
import { AdUnit } from './entities/ad-unit.entity';
import { Placement } from './entities/placement.entity';
import { FloorPrice } from './entities/floor-price.entity';
import { DemandPartner } from './entities/demand-partner.entity';
import { BrandSafetyRule } from './entities/brand-safety-rule.entity';

@Module({
  imports: [
    TypeOrmModule.forFeature([
      Publisher,
      AdUnit,
      Placement,
      FloorPrice,
      DemandPartner,
      BrandSafetyRule,
    ]),
  ],
  controllers: [
    SspController,
    PublisherController,
    InventoryController,
    DemandPartnerController,
  ],
  providers: [
    SspService,
    PublisherService,
    InventoryService,
    DemandPartnerService,
  ],
  exports: [
    SspService,
    PublisherService,
    InventoryService,
    DemandPartnerService,
  ],
})
export class SspModule {}
