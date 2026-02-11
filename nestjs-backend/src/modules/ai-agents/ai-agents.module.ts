import { Module } from '@nestjs/common';
import { AiAgentsService } from './ai-agents.service';
import { AiAgentsController } from './ai-agents.controller';
import { AiCoreService } from './ai-core.service';

@Module({
  providers: [AiAgentsService, AiCoreService],
  controllers: [AiAgentsController],
  exports: [AiAgentsService],
})
export class AiAgentsModule {}
