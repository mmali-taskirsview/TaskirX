import { Controller, Post, Get, Body, Query, Param, Headers, UseGuards } from '@nestjs/common';
import { ApiTags, ApiOperation, ApiResponse, ApiHeader } from '@nestjs/swagger';
import { AnalyticsService } from './analytics.service';

@ApiTags('mmp')
@Controller('mmp')
export class MmpController {
  constructor(private readonly analyticsService: AnalyticsService) {}

  @Post('events/track')
  @ApiOperation({ summary: 'Track MMP Event (Install/Purchase)' })
  @ApiResponse({ status: 201, description: 'Event tracked successfully' })
  @ApiHeader({ name: 'Authorization', description: 'Bearer Token (Optional for some providers)' })
  async trackEvent(@Body() data: any, @Headers('Authorization') authHeader: string) {
    // Validate provider
    if (!data.provider || !data.eventType) {
      return { error: 'Missing required fields: provider, eventType' };
    }

    await this.analyticsService.trackMmpEvent({
        provider: data.provider,
        eventType: data.eventType,
        campaignId: data.campaignId,
        userId: data.userId,
        deviceId: data.deviceId,
        revenue: data.revenue,
        currency: data.currency,
        metadata: data.metadata,
        timestamp: data.timestamp
    });

    return { 
        success: true, 
        message: 'MMP Event Received',
        conversion_id: 'conv-' + Date.now() 
    };
  }

  @Post('postback')
  @ApiOperation({ summary: 'Receive MMP Postback (Server-to-Server)' })
  @ApiResponse({ status: 200, description: 'Postback processed' })
  async receivePostback(@Query() query: any, @Body() body: any) {
     // MMPs sometimes send data in Query Params (GET/POST) or Body
     const data = { ...query, ...body };
     
     // Normalize Adjust/AppsFlyer payload to our internal structure
     const event = {
         provider: data.provider || 'generic',
         eventType: data.event_name || data.activity_kind || 'unknown',
         campaignId: data.campaign_id || data.c,
         deviceId: data.advertising_id || data.idfa || data.gaid,
         revenue: parseFloat(data.revenue || data.event_revenue || '0'),
         currency: data.currency || data.event_revenue_currency || 'USD',
         metadata: data
     };

     if (event.campaignId) {
         await this.analyticsService.trackMmpEvent(event);
         return { status: 'ok' };
     }
     
     return { status: 'ignored', reason: 'no_campaign_id' };
  }

  @Get('events/:campaignId/stats')
  @ApiOperation({ summary: 'Get MMP Attribution Stats' })
  async getStats(@Param('campaignId') campaignId: string) {
      // Logic to fetch stats from ClickHouse (MmpEvents table)
      // For MVP, we return mocked or basic stats if not implemented in service
      // Ideally calling this.analyticsService.getMmpStats(campaignId)
      return {
          campaignId,
          installs: 0, // TODO: Implement aggregation query
          revenue: 0,
          events: []
      }
  }
}
