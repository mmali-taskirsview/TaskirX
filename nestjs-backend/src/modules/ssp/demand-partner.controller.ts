import {
  Controller,
  Get,
  Post,
  Put,
  Delete,
  Body,
  Param,
  Query,
  HttpCode,
  HttpStatus,
} from '@nestjs/common';
import { DemandPartnerService, CreateDemandPartnerDto, UpdateDemandPartnerDto } from './demand-partner.service';

@Controller('ssp/demand-partners')
export class DemandPartnerController {
  constructor(private readonly demandPartnerService: DemandPartnerService) {}

  @Get()
  async findAll(@Query('publisherId') publisherId?: string) {
    return this.demandPartnerService.findAll(publisherId);
  }

  @Get('active')
  async findActive(@Query('publisherId') publisherId?: string) {
    return this.demandPartnerService.findActive(publisherId);
  }

  @Get('templates')
  async getTemplates() {
    return this.demandPartnerService.getAvailableTemplates();
  }

  @Get(':id')
  async findOne(@Param('id') id: string) {
    return this.demandPartnerService.findOne(id);
  }

  @Get(':id/stats')
  async getStats(@Param('id') id: string) {
    return this.demandPartnerService.getPartnerStats(id);
  }

  @Get(':id/bid-config')
  async getBidConfig(@Param('id') id: string) {
    return this.demandPartnerService.getPartnerBidConfig(id);
  }

  @Post()
  async create(@Body() dto: CreateDemandPartnerDto) {
    return this.demandPartnerService.create(dto);
  }

  @Post('from-template/:templateCode')
  async createFromTemplate(
    @Param('templateCode') templateCode: string,
    @Query('publisherId') publisherId?: string,
  ) {
    return this.demandPartnerService.createFromTemplate(templateCode, publisherId);
  }

  @Post('initialize-defaults')
  @HttpCode(HttpStatus.OK)
  async initializeDefaults() {
    return this.demandPartnerService.initializeDefaultPartners();
  }

  @Put(':id')
  async update(@Param('id') id: string, @Body() dto: UpdateDemandPartnerDto) {
    return this.demandPartnerService.update(id, dto);
  }

  @Delete(':id')
  @HttpCode(HttpStatus.NO_CONTENT)
  async delete(@Param('id') id: string) {
    return this.demandPartnerService.delete(id);
  }

  @Post(':id/activate')
  @HttpCode(HttpStatus.OK)
  async activate(@Param('id') id: string) {
    return this.demandPartnerService.activate(id);
  }

  @Post(':id/pause')
  @HttpCode(HttpStatus.OK)
  async pause(@Param('id') id: string) {
    return this.demandPartnerService.pause(id);
  }

  @Post(':id/disconnect')
  @HttpCode(HttpStatus.OK)
  async disconnect(@Param('id') id: string) {
    return this.demandPartnerService.disconnect(id);
  }
}
