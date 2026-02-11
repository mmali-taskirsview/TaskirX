import {
  Controller,
  Get,
  Post,
  Body,
  Query,
  HttpCode,
  HttpStatus,
} from '@nestjs/common';
import { SspService, BidRequest } from './ssp.service';

@Controller('ssp')
export class SspController {
  constructor(private readonly sspService: SspService) {}

  /**
   * Main auction endpoint - receives bid requests from publishers
   * OpenRTB 2.5 compatible
   */
  @Post('auction')
  @HttpCode(HttpStatus.OK)
  async runAuction(@Body() bidRequest: BidRequest) {
    const winner = await this.sspService.runAuction(bidRequest);
    
    if (!winner) {
      return {
        status: 'no_bid',
        message: 'No valid bids received',
      };
    }

    return {
      status: 'success',
      bid: winner,
    };
  }

  /**
   * OpenRTB 2.6 compatible alias endpoint
   */
  @Post('openrtb')
  @HttpCode(HttpStatus.OK)
  async runOpenRtbAuction(@Body() bidRequest: BidRequest) {
    return this.runAuction(bidRequest);
  }

  /**
   * Lightweight bid request for client-side ad tags
   */
  @Get('ad')
  async getAd(
    @Query('pub') publisherId: string,
    @Query('unit') adUnitId: string,
    @Query('placement') placementId?: string,
  ) {
    // Build minimal bid request from query params
    const bidRequest: BidRequest = {
      id: `req_${Date.now()}`,
      publisherId,
      adUnitId,
      placementId,
      device: {
        type: 'desktop',
        os: 'unknown',
        browser: 'unknown',
      },
      geo: {
        country: 'US',
        region: 'unknown',
        city: 'unknown',
      },
      // Allow by default for lightweight tags
      // Consent enforcement applies when explicit consent flags are provided
      user: {
        consent: true,
      },
      timeout: 200,
    };

    const winner = await this.sspService.runAuction(bidRequest);

    if (!winner) {
      return {
        status: 'no_fill',
      };
    }

    // Return ad creative for rendering
    return {
      status: 'success',
      ad: {
        html: `<iframe src="${winner.creative.url}" width="${winner.creative.width}" height="${winner.creative.height}" frameborder="0" scrolling="no"></iframe>`,
        width: winner.creative.width,
        height: winner.creative.height,
        bidPrice: winner.price,
        partner: winner.partnerName,
      },
    };
  }

  /**
   * SSP dashboard metrics
   */
  @Get('dashboard')
  async getDashboard(@Query('publisherId') publisherId?: string) {
    return this.sspService.getDashboardMetrics(publisherId);
  }

  /**
   * Health check for SSP service
   */
  @Get('health')
  getHealth() {
    return {
      status: 'healthy',
      service: 'ssp',
      timestamp: new Date().toISOString(),
    };
  }

  /**
   * Get floor price for a specific bid request
   */
  @Post('floor-price')
  @HttpCode(HttpStatus.OK)
  async getFloorPrice(@Body() bidRequest: Partial<BidRequest>) {
    const fullRequest: BidRequest = {
      id: `req_${Date.now()}`,
      publisherId: bidRequest.publisherId || '',
      adUnitId: bidRequest.adUnitId || '',
      device: bidRequest.device || { type: 'desktop', os: 'unknown', browser: 'unknown' },
      geo: bidRequest.geo || { country: 'US', region: '', city: '' },
      user: bidRequest.user || {},
    };

    const floorPrice = await this.sspService.calculateFloorPrice(fullRequest);
    
    return {
      floor: floorPrice,
      currency: 'USD',
    };
  }
}
