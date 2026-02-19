import { Controller, Post, Get, Body, Query, Res } from '@nestjs/common';
import { Response } from 'express';
import { ApiTags, ApiOperation, ApiResponse, ApiQuery } from '@nestjs/swagger';
import { AnalyticsService } from './analytics.service';

@ApiTags('tracking')
@Controller('analytics/track') // Keeps the same route structure
export class TrackingController {
  constructor(private readonly analyticsService: AnalyticsService) {}

  @Post('impression')
  @ApiOperation({ summary: 'Track ad impression (Public)' })
  @ApiResponse({ status: 201, description: 'Impression tracked successfully' })
  async trackImpression(@Body() data: any) {
    await this.analyticsService.trackImpression(data);
    return { success: true };
  }

  @Get('impression') // Support GET for pixels
  @ApiOperation({ summary: 'Track ad impression via Pixel (Public)' })
  async trackImpressionPixel(@Query() query: any, @Res() res: Response) {
    await this.analyticsService.trackImpression({
      ...query,
      price: query.price ? parseFloat(query.price) : undefined, // Convert price string to float
    });
    // Return 1x1 pixel
    const pixel = Buffer.from(
      'R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7',
      'base64',
    );
    res.writeHead(200, {
      'Content-Type': 'image/gif',
      'Content-Length': pixel.length,
    });
    res.end(pixel);
  }

  @Post('click')
  @ApiOperation({ summary: 'Track ad click (Public)' })
  async trackClick(@Body() data: any) {
    await this.analyticsService.trackClick(data);
    return { success: true };
  }

  /**
   * GET /analytics/track/click
   *
   * Click beacon endpoint used by the ClickURL embedded in ad markup.
   * The Go bidding engine generates URLs of the form:
   *   /api/analytics/track/click?campaign_id=X&request_id=Y
   *
   * Supports an optional `redirect` query param so the browser can be
   * forwarded to the advertiser's landing page after the click is recorded.
   */
  @Get('click')
  @ApiOperation({ summary: 'Track ad click via GET beacon / redirect (Public)' })
  @ApiQuery({ name: 'campaign_id', required: true, description: 'Campaign ID' })
  @ApiQuery({ name: 'request_id', required: false, description: 'Bid request / impression ID' })
  @ApiQuery({ name: 'redirect', required: false, description: 'Landing page URL to redirect to after tracking' })
  async trackClickBeacon(
    @Query('campaign_id') campaignId: string,
    @Query('request_id') requestId: string,
    @Query('redirect') redirectUrl: string,
    @Res() res: Response,
  ) {
    if (!campaignId) {
      res.status(400).json({ error: 'campaign_id is required' });
      return;
    }

    // Record the click — use requestId as impressionId linkage
    try {
      await this.analyticsService.trackClick({
        campaignId,
        impressionId: requestId || '',
        timestamp: new Date(),
      });
    } catch (_err) {
      // Non-fatal: tracking failure should never block the user redirect
    }

    if (redirectUrl) {
      res.redirect(302, redirectUrl);
    } else {
      // Return a 1×1 transparent GIF as fallback pixel response
      const pixel = Buffer.from(
        'R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7',
        'base64',
      );
      res.writeHead(200, {
        'Content-Type': 'image/gif',
        'Content-Length': pixel.length,
      });
      res.end(pixel);
    }
  }

  @Post('conversion')
  @ApiOperation({ summary: 'Track conversion (Public)' })
  async trackConversion(@Body() data: any) {
    await this.analyticsService.trackConversion(data);
    return { success: true };
  }

  @Get('video')
  @ApiOperation({ summary: 'Track video events (Public)' })
  async trackVideo(
    @Query('campaignId') campaignId: string,
    @Query('event') event: 'start' | 'firstQuartile' | 'midpoint' | 'thirdQuartile' | 'complete',
    @Res() res: Response
  ) {
    if (campaignId && event) {
      await this.analyticsService.trackVideoEvent({
        campaignId,
        eventType: event,
        timestamp: new Date()
      });
    }

    // Return 1x1 pixel
    const pixel = Buffer.from(
      'R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7',
      'base64',
    );
    res.writeHead(200, {
      'Content-Type': 'image/gif',
      'Content-Length': pixel.length,
    });
    res.end(pixel);
  }
}
