import { Controller, Get, Post, Body, Param, UseGuards, Request } from '@nestjs/common';
import { ApiTags, ApiOperation, ApiBearerAuth } from '@nestjs/swagger';
import { AiAgentsService } from './ai-agents.service';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { RolesGuard } from '../auth/guards/roles.guard';
import { Roles } from '../auth/decorators/roles.decorator';
import { UserRole } from '../users/user.entity';

@ApiTags('ai-agents')
@Controller('ai')
@UseGuards(JwtAuthGuard, RolesGuard)
@ApiBearerAuth()
export class AiAgentsController {
  constructor(private readonly aiAgentsService: AiAgentsService) {}

  @Post('fraud/detect')
  @ApiOperation({ summary: 'Detect fraudulent activity' })
  @Roles(UserRole.ADMIN, UserRole.ADVERTISER)
  async detectFraud(@Body() data: any) {
    return this.aiAgentsService.detectFraud(data);
  }

  @Post('match')
  @ApiOperation({ summary: 'Match ads to user/context' })
  async matchAd(@Body() request: any) {
    return this.aiAgentsService.matchAd(request);
  }

  @Post('bid/optimize')
  @ApiOperation({ summary: 'Optimize bid price' })
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async optimizeBid(@Body() data: any) {
    return this.aiAgentsService.optimizeBid(data);
  }

  @Get('predict/:campaignId')
  @ApiOperation({ summary: 'Predict campaign performance' })
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async predictPerformance(@Param('campaignId') campaignId: string) {
    return this.aiAgentsService.predictPerformance(campaignId);
  }

  @Get('anomalies')
  @ApiOperation({ summary: 'Get detected anomalies' })
  @Roles(UserRole.ADMIN)
  async getAnomalies(@Request() req) {
    return this.aiAgentsService.getAnomalies(req.user.tenantId);
  }
}
