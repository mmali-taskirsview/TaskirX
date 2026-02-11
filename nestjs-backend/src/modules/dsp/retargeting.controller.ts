import { Controller, Get, Query, Res } from '@nestjs/common';
import { ApiTags, ApiOperation } from '@nestjs/swagger';
import { Response } from 'express';
import { DspService } from './dsp.service';

@ApiTags('DSP Retargeting')
@Controller('dsp/pixel')
export class RetargetingController {
  constructor(private readonly dspService: DspService) {}

  @Get()
  @ApiOperation({ summary: 'Track retargeting event (1x1 pixel)' })
  async trackPixel(
    @Query('id') id: string,
    @Query('evt') evt: string,
    @Res() res: Response,
  ) {
    if (id) {
      // Fire and forget - don't block the response
      this.dspService.trackRetargetingEvent(id, evt || 'page_view').catch(console.error);
    }

    // Return 1x1 transparent GIF
    const pixel = Buffer.from(
      'R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7',
      'base64',
    );

    res.writeHead(200, {
      'Content-Type': 'image/gif',
      'Content-Length': pixel.length,
      'Cache-Control': 'no-store, no-cache, must-revalidate, proxy-revalidate',
    });
    res.end(pixel);
  }
}
