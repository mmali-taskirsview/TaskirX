import { Module } from '@nestjs/common';
import { MetricsGateway } from './metrics.gateway';
import { AlertsGateway } from './alerts.gateway';
import { CampaignsGateway } from './campaigns.gateway';
import { AnalyticsService } from '../services/analytics.service';
import { RedisModule } from '../modules/redis/redis.module';

/**
 * WebSocket Gateway Module
 * 
 * Provides real-time communication via WebSocket/Socket.IO
 * 
 * Components:
 * - MetricsGateway: Real-time metrics and analytics
 * - AlertsGateway: System and campaign alerts
 * - CampaignsGateway: Campaign updates and collaboration
 * 
 * Features:
 * - Multi-instance support via Redis
 * - Connection pooling
 * - Namespace isolation
 * - Automatic reconnection
 * - Message queuing
 */
@Module({
  imports: [RedisModule],
  providers: [
    MetricsGateway,
    AlertsGateway,
    CampaignsGateway,
    AnalyticsService,
  ],
})
export class GatewayModule {}
