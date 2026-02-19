import { Controller, Get, Query, UseGuards, Request, Param } from '@nestjs/common';
import { ApiTags, ApiOperation, ApiBearerAuth, ApiQuery } from '@nestjs/swagger';
import { AnalyticsService } from './analytics.service';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';

@ApiTags('analytics')
@Controller('analytics')
@UseGuards(JwtAuthGuard)
@ApiBearerAuth()
export class AnalyticsController {
  constructor(private readonly analyticsService: AnalyticsService) {}

  @Get('dashboard')
  @ApiOperation({ summary: 'Get dashboard statistics' })
  @ApiQuery({ name: 'tenantId', required: false, description: 'Override tenant ID (for admin/service use)' })
  async getDashboard(
    @Request() req,
    @Query('dateFrom') dateFrom: string,
    @Query('dateTo') dateTo: string,
    @Query('tenantId') tenantIdOverride?: string,
  ) {
    const tenantId = tenantIdOverride || req.user?.tenantId;
    const from = dateFrom ? new Date(dateFrom) : new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
    const to = dateTo ? new Date(dateTo) : new Date();
    
    return this.analyticsService.getDashboardStats(tenantId, from, to);
  }

  @Get('campaign/:id')
  @ApiOperation({ summary: 'Get campaign statistics' })
  async getCampaignStats(
    @Param('id') campaignId: string,
    @Query('dateFrom') dateFrom: string,
    @Query('dateTo') dateTo: string,
  ) {
    const from = dateFrom ? new Date(dateFrom) : new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
    const to = dateTo ? new Date(dateTo) : new Date();
    
    return this.analyticsService.getCampaignStats(campaignId, from, to);
  }

  @Get('revenue')
  @ApiOperation({ summary: 'Get revenue by date' })
  @ApiQuery({ name: 'tenantId', required: false, description: 'Override tenant ID (for admin/service use)' })
  async getRevenue(
    @Request() req,
    @Query('dateFrom') dateFrom: string,
    @Query('dateTo') dateTo: string,
    @Query('tenantId') tenantIdOverride?: string,
  ) {
    const tenantId = tenantIdOverride || req.user?.tenantId;
    const from = dateFrom ? new Date(dateFrom) : new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
    const to = dateTo ? new Date(dateTo) : new Date();
    
    return this.analyticsService.getRevenueByDate(tenantId, from, to);
  }

  @Get('top-campaigns')
  @ApiOperation({ summary: 'Get top performing campaigns' })
  @ApiQuery({ name: 'tenantId', required: false, description: 'Override tenant ID (for admin/service use)' })
  async getTopCampaigns(
    @Request() req,
    @Query('limit') limit?: number,
    @Query('tenantId') tenantIdOverride?: string,
  ) {
    const tenantId = tenantIdOverride || req.user?.tenantId;
    return this.analyticsService.getTopPerformingCampaigns(tenantId, limit || 10);
  }

  @Get('supply-chain')
  @ApiOperation({ summary: 'Get supply chain metrics for SPO analytics' })
  async getSupplyChainMetrics(@Query('timeRange') timeRange: string = '1h') {
    return this.analyticsService.getSupplyChainMetrics(timeRange);
  }

  @Get('supply-path-optimization')
  @ApiOperation({ summary: 'Get supply path optimization recommendations' })
  async getSupplyPathOptimization(@Query('timeRange') timeRange: string = '1h') {
    return this.analyticsService.getSupplyPathOptimization(timeRange);
  }

  @Get('bid-path/:requestId')
  @ApiOperation({ summary: 'Get detailed bid path analytics for a specific request' })
  async getBidPathAnalytics(@Param('requestId') requestId: string) {
    return this.analyticsService.getBidPathAnalytics(requestId);
  }

  @Get('service-performance')
  @ApiOperation({ summary: 'Get performance metrics for a specific service' })
  async getServicePerformance(
    @Query('serviceName') serviceName: string,
    @Query('timeRange') timeRange: string = '1h'
  ) {
    return this.analyticsService.getServicePerformance(serviceName, timeRange);
  }

  @Get('direct-publisher-analysis')
  @ApiOperation({ summary: 'Get analysis of direct publisher relationship opportunities' })
  async getDirectPublisherAnalysis(@Query('timeRange') timeRange: string = '1h') {
    return this.analyticsService.getDirectPublisherAnalysis(timeRange);
  }

  @Get('cost-benefit-analysis')
  @ApiOperation({ summary: 'Get detailed cost-benefit analysis for optimization scenarios' })
  async getCostBenefitAnalysis(@Query('timeRange') timeRange: string = '1h') {
    return this.analyticsService.getCostBenefitAnalysis(timeRange);
  }
}
