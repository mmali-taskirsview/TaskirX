import {
  Controller,
  Get,
  Post,
  Put,
  Delete,
  Body,
  Param,
  Query,
} from '@nestjs/common';
import { ApiTags, ApiOperation } from '@nestjs/swagger';
import { DspService } from './dsp.service';
import { SupplyPartner } from './entities/supply-partner.entity';
import { AudienceSegment } from './entities/audience-segment.entity';
import { Deal } from './entities/deal.entity';
import { BidStrategy } from './entities/bid-strategy.entity';
import { CreateAudienceDto } from './dto/create-audience.dto';

@ApiTags('DSP')
@Controller('dsp')
export class DspController {
  constructor(private readonly dspService: DspService) {}

  // ==================== DASHBOARD ====================

  @Get('dashboard')
  @ApiOperation({ summary: 'Get DSP dashboard stats' })
  async getDashboard() {
    return this.dspService.getDashboardStats();
  }

  @Get('rtb-analytics')
  @ApiOperation({ summary: 'Get RTB analytics' })
  async getRtbAnalytics() {
    return this.dspService.getRtbAnalytics();
  }

  // ==================== SUPPLY PARTNERS ====================

  @Get('supply-partners')
  @ApiOperation({ summary: 'Get all supply partners' })
  async getSupplyPartners(): Promise<SupplyPartner[]> {
    return this.dspService.getSupplyPartners();
  }

  @Get('supply-partners/:id')
  @ApiOperation({ summary: 'Get supply partner by ID' })
  async getSupplyPartner(@Param('id') id: string): Promise<SupplyPartner> {
    return this.dspService.getSupplyPartner(id);
  }

  @Post('supply-partners')
  @ApiOperation({ summary: 'Create supply partner' })
  async createSupplyPartner(@Body() data: Partial<SupplyPartner>): Promise<SupplyPartner> {
    return this.dspService.createSupplyPartner(data);
  }

  @Put('supply-partners/:id')
  @ApiOperation({ summary: 'Update supply partner' })
  async updateSupplyPartner(
    @Param('id') id: string,
    @Body() data: Partial<SupplyPartner>,
  ): Promise<SupplyPartner> {
    return this.dspService.updateSupplyPartner(id, data);
  }

  @Delete('supply-partners/:id')
  @ApiOperation({ summary: 'Delete supply partner' })
  async deleteSupplyPartner(@Param('id') id: string): Promise<void> {
    return this.dspService.deleteSupplyPartner(id);
  }

  // ==================== AUDIENCE SEGMENTS ====================

  @Get('audiences')
  @ApiOperation({ summary: 'Get all audience segments' })
  async getAudiences(@Query('advertiserId') advertiserId?: string): Promise<AudienceSegment[]> {
    return this.dspService.getAudienceSegments(advertiserId);
  }

  @Get('audiences/:id')
  @ApiOperation({ summary: 'Get audience segment by ID' })
  async getAudience(@Param('id') id: string): Promise<AudienceSegment> {
    return this.dspService.getAudienceSegment(id);
  }

  @Post('audiences')
  @ApiOperation({ summary: 'Create audience segment' })
  async createAudience(@Body() data: CreateAudienceDto): Promise<AudienceSegment> {
    return this.dspService.createAudienceSegment(data);
  }

  @Put('audiences/:id')
  @ApiOperation({ summary: 'Update audience segment' })
  async updateAudience(
    @Param('id') id: string,
    @Body() data: Partial<AudienceSegment>,
  ): Promise<AudienceSegment> {
    return this.dspService.updateAudienceSegment(id, data);
  }

  @Delete('audiences/:id')
  @ApiOperation({ summary: 'Delete audience segment' })
  async deleteAudience(@Param('id') id: string): Promise<void> {
    return this.dspService.deleteAudienceSegment(id);
  }

  // ==================== DEALS ====================

  @Get('deals')
  @ApiOperation({ summary: 'Get all deals' })
  async getDeals(@Query('advertiserId') advertiserId?: string): Promise<Deal[]> {
    return this.dspService.getDeals(advertiserId);
  }

  @Get('deals/:id')
  @ApiOperation({ summary: 'Get deal by ID' })
  async getDeal(@Param('id') id: string): Promise<Deal> {
    return this.dspService.getDeal(id);
  }

  @Post('deals')
  @ApiOperation({ summary: 'Create deal' })
  async createDeal(@Body() data: Partial<Deal>): Promise<Deal> {
    return this.dspService.createDeal(data);
  }

  @Put('deals/:id')
  @ApiOperation({ summary: 'Update deal' })
  async updateDeal(
    @Param('id') id: string,
    @Body() data: Partial<Deal>,
  ): Promise<Deal> {
    return this.dspService.updateDeal(id, data);
  }

  @Delete('deals/:id')
  @ApiOperation({ summary: 'Delete deal' })
  async deleteDeal(@Param('id') id: string): Promise<void> {
    return this.dspService.deleteDeal(id);
  }

  // ==================== BID STRATEGIES ====================

  @Get('bid-strategies')
  @ApiOperation({ summary: 'Get all bid strategies' })
  async getBidStrategies(@Query('advertiserId') advertiserId?: string): Promise<BidStrategy[]> {
    return this.dspService.getBidStrategies(advertiserId);
  }

  @Get('bid-strategies/:id')
  @ApiOperation({ summary: 'Get bid strategy by ID' })
  async getBidStrategy(@Param('id') id: string): Promise<BidStrategy> {
    return this.dspService.getBidStrategy(id);
  }

  @Post('bid-strategies')
  @ApiOperation({ summary: 'Create bid strategy' })
  async createBidStrategy(@Body() data: Partial<BidStrategy>): Promise<BidStrategy> {
    return this.dspService.createBidStrategy(data);
  }

  @Put('bid-strategies/:id')
  @ApiOperation({ summary: 'Update bid strategy' })
  async updateBidStrategy(
    @Param('id') id: string,
    @Body() data: Partial<BidStrategy>,
  ): Promise<BidStrategy> {
    return this.dspService.updateBidStrategy(id, data);
  }

  @Delete('bid-strategies/:id')
  @ApiOperation({ summary: 'Delete bid strategy' })
  async deleteBidStrategy(@Param('id') id: string): Promise<void> {
    return this.dspService.deleteBidStrategy(id);
  }

  // ==================== BIDDING ====================

  @Post('bid')
  @ApiOperation({ summary: 'Process bid request' })
  async processBid(@Body() bidRequest: any) {
    return this.dspService.processBidRequest(bidRequest);
  }

  @Post('win')
  @ApiOperation({ summary: 'Record win notification' })
  async recordWin(@Body() data: { bidId: string; supplyPartnerId: string; price: number }) {
    await this.dspService.recordWin(data.bidId, data.supplyPartnerId, data.price);
    return { success: true };
  }
}
