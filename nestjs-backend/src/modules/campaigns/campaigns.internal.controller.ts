import { Controller, Get } from '@nestjs/common';
import { ApiTags, ApiOperation } from '@nestjs/swagger';
import { CampaignsService } from './campaigns.service';

@ApiTags('internal')
@Controller('internal/campaigns')
export class CampaignsInternalController {
  constructor(private readonly campaignsService: CampaignsService) {}

  @Get('active')
  @ApiOperation({ summary: 'Get active campaigns for bidding engine (Internal)' })
  async getActiveCampaigns() {
    return this.campaignsService.getActiveCampaigns();
  }
}
