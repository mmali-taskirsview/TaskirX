import { Controller, Get, Query, UseGuards, Request, Post, Body, Param } from '@nestjs/common';
import { ApiTags, ApiOperation, ApiBearerAuth } from '@nestjs/swagger';
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
  async getDashboard(
    @Request() req,
    @Query('dateFrom') dateFrom: string,
    @Query('dateTo') dateTo: string,
  ) {
    const from = dateFrom ? new Date(dateFrom) : new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
    const to = dateTo ? new Date(dateTo) : new Date();
    
    return this.analyticsService.getDashboardStats(req.user.tenantId, from, to);
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
  async getRevenue(
    @Request() req,
    @Query('dateFrom') dateFrom: string,
    @Query('dateTo') dateTo: string,
  ) {
    const from = dateFrom ? new Date(dateFrom) : new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
    const to = dateTo ? new Date(dateTo) : new Date();
    
    return this.analyticsService.getRevenueByDate(req.user.tenantId, from, to);
  }

  @Get('top-campaigns')
  @ApiOperation({ summary: 'Get top performing campaigns' })
  async getTopCampaigns(@Request() req, @Query('limit') limit?: number) {
    return this.analyticsService.getTopPerformingCampaigns(req.user.tenantId, limit || 10);
  }

  // This endpoint is effectively public as it's called by the ad (pixel) or click redirect
  // In a real scenario, you usually disable AuthGuard for tracking pixels/links
  // or use a specialized tracking service/subdomain.
  @Post('track/impression')
  @UseGuards() // Explicitly disable AuthGuard if possible, or create a separate PublicController
  @ApiOperation({ summary: 'Track ad impression' })
  async trackImpression(@Body() data: any) {
    await this.analyticsService.trackImpression(data);
    return { success: true };
  }

  @Post('track/click')
  @ApiOperation({ summary: 'Track ad click' })
  async trackClick(@Body() data: any) {
    await this.analyticsService.trackClick(data);
    return { success: true };
  }

  @Post('track/conversion')
  @ApiOperation({ summary: 'Track conversion' })
  async trackConversion(@Body() data: any) {
    await this.analyticsService.trackConversion(data);
    return { success: true };
  }
}
