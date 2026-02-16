import {
  Controller,
  Get,
  Post,
  Put,
  Body,
  Param,
  BadRequestException,
} from '@nestjs/common';
import { ApiTags, ApiOperation, ApiBearerAuth } from '@nestjs/swagger';
import { AnalyticsService } from '../services/analytics.service';
import { BiddingOptimizationService } from '../services/bidding-optimization.service';
import { BillingService, SubscriptionTier } from '../services/billing.service';
import { AggregatedMetrics, ConversionFunnel } from '../services/analytics.types';
import { BidPrediction, ModelPerformance } from '../services/bidding.types';

@ApiTags('advanced')
@ApiBearerAuth()
@Controller('advanced')
export class AdvancedController {
  constructor(
    private readonly analyticsService: AnalyticsService,
    private readonly biddingService: BiddingOptimizationService,
    private readonly billingService: BillingService,
  ) {}

  // ========== ANALYTICS ENDPOINTS ==========

  @Post('analytics/events')
  @ApiOperation({ summary: 'Track analytics event' })
  async trackEvent(
    @Body()
    eventData: {
      campaignId: string;
      eventType: 'impression' | 'click' | 'conversion' | 'error';
      userId?: string;
      value?: number;
      metadata?: Record<string, any>;
    },
  ) {
    return this.analyticsService.trackEvent({
      eventId: `evt_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      campaignId: eventData.campaignId,
      userId: eventData.userId || 'anonymous',
      eventType: eventData.eventType,
      value: eventData.value,
      metadata: eventData.metadata || {},
      timestamp: new Date(),
    });
  }

  @Get('analytics/campaigns/:campaignId/metrics')
  @ApiOperation({ summary: 'Get campaign metrics' })
  async getCampaignMetrics(
    @Param('campaignId') campaignId: string,
    @Body()
    query?: {
      startDate?: string;
      endDate?: string;
    },
  ): Promise<AggregatedMetrics> {
    return this.analyticsService.getCampaignMetrics(
      campaignId,
      query?.startDate ? new Date(query.startDate) : undefined,
      query?.endDate ? new Date(query.endDate) : undefined,
    );
  }

  @Get('analytics/campaigns/:campaignId/realtime')
  @ApiOperation({ summary: 'Get real-time metrics' })
  async getRealtimeMetrics(@Param('campaignId') campaignId: string): Promise<AggregatedMetrics> {
    return this.analyticsService.getRealtimeMetrics(campaignId);
  }

  @Get('analytics/campaigns/:campaignId/funnel')
  @ApiOperation({ summary: 'Get conversion funnel' })
  async getConversionFunnel(@Param('campaignId') campaignId: string): Promise<ConversionFunnel[]> {
    return this.analyticsService.getConversionFunnel(campaignId);
  }

  @Get('analytics/top-campaigns')
  @ApiOperation({ summary: 'Get top performing campaigns' })
  async getTopCampaigns(
    @Body()
    query?: {
      limit?: number;
      metric?: 'conversions' | 'revenue' | 'ctr' | 'roi';
    },
  ) {
    return this.analyticsService.getTopCampaigns(
      query?.limit || 10,
      query?.metric || 'conversions',
    );
  }

  @Post('analytics/reports')
  @ApiOperation({ summary: 'Generate custom report' })
  async generateReport(
    @Body()
    data: {
      campaignId: string;
      startDate: string;
      endDate: string;
      groupBy?: 'hour' | 'day' | 'week';
      includeBreakdown?: boolean;
      metrics?: string[];
    },
  ) {
    return this.analyticsService.generateReport(
      data.campaignId,
      new Date(data.startDate),
      new Date(data.endDate),
      {
        groupBy: data.groupBy,
        includeBreakdown: data.includeBreakdown,
        metrics: data.metrics,
      },
    );
  }

  // ========== BIDDING OPTIMIZATION ENDPOINTS ==========

  @Post('bidding/predict')
  @ApiOperation({ summary: 'Predict optimal bid' })
  async predictBid(
    @Body()
    context: {
      campaignId: string;
      adSpaceId: string;
      userId: string;
      deviceType: string;
      location: string;
      dayOfWeek: number;
      hourOfDay: number;
      historicalCTR: number;
      historicalCR: number;
      budget: number;
    },
  ): Promise<BidPrediction> {
    return this.biddingService.predictOptimalBid(context);
  }

  @Post('bidding/adaptive')
  @ApiOperation({ summary: 'Get adaptive bid' })
  async getAdaptiveBid(
    @Body()
    context: {
      campaignId: string;
      adSpaceId: string;
      userId: string;
      deviceType: string;
      location: string;
      dayOfWeek: number;
      hourOfDay: number;
      historicalCTR: number;
      historicalCR: number;
      budget: number;
    },
  ) {
    return this.biddingService.getAdaptiveBid(context);
  }

  @Post('bidding/ab-test')
  @ApiOperation({ summary: 'Run A/B test on bidding strategies' })
  async runABTest(
    @Body()
    data: {
      campaignId: string;
      strategyA: (context: any) => number;
      strategyB: (context: any) => number;
      testSize?: number;
    },
  ) {
    // Create wrapper functions that can be serialized
    const strategyA = (ctx: any) => ctx.budget * 0.1;
    const strategyB = (ctx: any) => ctx.budget * 0.15;

    return this.biddingService.runABTest(
      data.campaignId,
      strategyA,
      strategyB,
      data.testSize || 1000,
    );
  }

  @Post('bidding/train')
  @ApiOperation({ summary: 'Train bidding models' })
  async trainBiddingModels(
    @Body()
    data: {
      trainingData: Array<{
        context: any;
        outcome: number;
      }>;
    },
  ): Promise<ModelPerformance> {
    if (!data.trainingData || data.trainingData.length === 0) {
      throw new BadRequestException('Training data required');
    }
    return this.biddingService.trainModels(data.trainingData);
  }

  @Get('bidding/performance')
  @ApiOperation({ summary: 'Get model performance' })
  async getModelPerformance(
    @Body()
    query?: {
      modelName?: string;
    },
  ): Promise<ModelPerformance> {
    return this.biddingService.getModelPerformance(query?.modelName);
  }

  // ========== BILLING ENDPOINTS ==========

  @Post('billing/subscriptions')
  @ApiOperation({ summary: 'Create subscription' })
  async createSubscription(
    @Body()
    data: {
      tenantId: string;
      tier: SubscriptionTier;
      paymentMethodId: string;
    },
  ) {
    return this.billingService.createSubscription(
      data.tenantId,
      data.tier,
      data.paymentMethodId,
    );
  }

  @Put('billing/subscriptions/:subscriptionId/upgrade')
  @ApiOperation({ summary: 'Upgrade subscription' })
  async upgradeSubscription(
    @Param('subscriptionId') subscriptionId: string,
    @Body()
    data: {
      tier: SubscriptionTier;
    },
  ) {
    return this.billingService.upgradeSubscription(subscriptionId, data.tier);
  }

  @Put('billing/subscriptions/:subscriptionId/cancel')
  @ApiOperation({ summary: 'Cancel subscription' })
  async cancelSubscription(@Param('subscriptionId') subscriptionId: string) {
    return this.billingService.cancelSubscription(subscriptionId);
  }

  @Get('billing/subscriptions/:subscriptionId')
  @ApiOperation({ summary: 'Get subscription details' })
  async getSubscription(@Param('subscriptionId') subscriptionId: string) {
    return this.billingService.getSubscription(subscriptionId);
  }

  @Get('billing/plans/:tier')
  @ApiOperation({ summary: 'Get subscription plan details' })
  async getPlan(@Param('tier') tier: SubscriptionTier) {
    return this.billingService.getSubscriptionPlan(tier);
  }

  @Post('billing/usage')
  @ApiOperation({ summary: 'Track usage' })
  async trackUsage(
    @Body()
    data: {
      tenantId: string;
      campaignsCreated?: number;
      campaignsActive?: number;
      totalSpent?: number;
      apiCallsUsed?: number;
      customMetricsUsed?: number;
      webhooksUsed?: number;
      integrationsUsed?: number;
    },
  ) {
    return this.billingService.trackUsage(data.tenantId, {
      campaignsCreated: data.campaignsCreated || 0,
      campaignsActive: data.campaignsActive || 0,
      totalSpent: data.totalSpent || 0,
      apiCallsUsed: data.apiCallsUsed || 0,
      customMetricsUsed: data.customMetricsUsed || 0,
      webhooksUsed: data.webhooksUsed || 0,
      integrationsUsed: data.integrationsUsed || 0,
    });
  }

  @Get('billing/usage/:tenantId')
  @ApiOperation({ summary: 'Get usage metrics' })
  async getUsage(@Param('tenantId') tenantId: string) {
    return this.billingService.getUsageMetrics(tenantId);
  }

  @Get('billing/usage/:tenantId/limits')
  @ApiOperation({ summary: 'Check usage limits' })
  async checkLimits(@Param('tenantId') tenantId: string) {
    return this.billingService.checkUsageLimits(tenantId);
  }

  @Get('billing/invoices/:tenantId')
  @ApiOperation({ summary: 'Get invoices' })
  async getInvoices(@Param('tenantId') tenantId: string) {
    return this.billingService.getInvoices(tenantId);
  }

  @Put('billing/invoices/:invoiceId/paid')
  @ApiOperation({ summary: 'Mark invoice as paid' })
  async markInvoicePaid(@Param('invoiceId') invoiceId: string) {
    return this.billingService.markInvoicePaid(invoiceId);
  }

  // ========== HEALTH & STATUS ==========

  @Get('health')
  @ApiOperation({ summary: 'Advanced services health check' })
  async health() {
    return {
      status: 'healthy',
      services: {
        analytics: 'running',
        bidding: 'running',
        billing: 'running',
      },
      timestamp: new Date(),
    };
  }
}
