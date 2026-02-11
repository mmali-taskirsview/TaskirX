import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { IntegrationsController } from './integrations.controller';
import { IntegrationsWellKnownController } from './integrations.wellknown.controller';
import { IntegrationsService } from './integrations.service';
import { IntegrationConfig } from './entities/integration-config.entity';

@Module({
  imports: [TypeOrmModule.forFeature([IntegrationConfig])],
  controllers: [IntegrationsController, IntegrationsWellKnownController],
  providers: [IntegrationsService],
  exports: [IntegrationsService],
})
export class IntegrationsModule {}
