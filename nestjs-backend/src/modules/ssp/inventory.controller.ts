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
import {
  InventoryService,
  CreatePlacementDto,
  CreateFloorPriceDto,
  CreateBrandSafetyRuleDto,
} from './inventory.service';
import { CreateAdUnitDto } from './dto/create-ad-unit.dto';

@Controller('ssp/inventory')
export class InventoryController {
  constructor(private readonly inventoryService: InventoryService) {}

  // =================== Ad Units ===================

  @Get('ad-units')
  async findAllAdUnits(@Query('publisherId') publisherId?: string) {
    return this.inventoryService.findAllAdUnits(publisherId);
  }

  @Get('ad-units/:id')
  async findOneAdUnit(@Param('id') id: string) {
    return this.inventoryService.findOneAdUnit(id);
  }

  @Post('ad-units')
  async createAdUnit(@Body() dto: CreateAdUnitDto) {
    return this.inventoryService.createAdUnit(dto);
  }

  @Put('ad-units/:id')
  async updateAdUnit(@Param('id') id: string, @Body() dto: Partial<CreateAdUnitDto>) {
    return this.inventoryService.updateAdUnit(id, dto);
  }

  @Delete('ad-units/:id')
  @HttpCode(HttpStatus.NO_CONTENT)
  async deleteAdUnit(@Param('id') id: string) {
    return this.inventoryService.deleteAdUnit(id);
  }

  @Post('ad-units/:id/activate')
  @HttpCode(HttpStatus.OK)
  async activateAdUnit(@Param('id') id: string) {
    return this.inventoryService.activateAdUnit(id);
  }

  @Post('ad-units/:id/pause')
  @HttpCode(HttpStatus.OK)
  async pauseAdUnit(@Param('id') id: string) {
    return this.inventoryService.pauseAdUnit(id);
  }

  // =================== Placements ===================

  @Get('placements')
  async findAllPlacements(@Query('publisherId') publisherId?: string) {
    return this.inventoryService.findAllPlacements(publisherId);
  }

  @Get('placements/:id')
  async findOnePlacement(@Param('id') id: string) {
    return this.inventoryService.findOnePlacement(id);
  }

  @Post('placements')
  async createPlacement(@Body() dto: CreatePlacementDto) {
    return this.inventoryService.createPlacement(dto);
  }

  @Put('placements/:id')
  async updatePlacement(@Param('id') id: string, @Body() dto: Partial<CreatePlacementDto>) {
    return this.inventoryService.updatePlacement(id, dto);
  }

  @Delete('placements/:id')
  @HttpCode(HttpStatus.NO_CONTENT)
  async deletePlacement(@Param('id') id: string) {
    return this.inventoryService.deletePlacement(id);
  }

  // =================== Floor Prices ===================

  @Get('floor-prices')
  async findAllFloorPrices(@Query('publisherId') publisherId?: string) {
    return this.inventoryService.findAllFloorPrices(publisherId);
  }

  @Get('floor-prices/:id')
  async findOneFloorPrice(@Param('id') id: string) {
    return this.inventoryService.findOneFloorPrice(id);
  }

  @Post('floor-prices')
  async createFloorPrice(@Body() dto: CreateFloorPriceDto) {
    return this.inventoryService.createFloorPrice(dto);
  }

  @Put('floor-prices/:id')
  async updateFloorPrice(@Param('id') id: string, @Body() dto: Partial<CreateFloorPriceDto>) {
    return this.inventoryService.updateFloorPrice(id, dto);
  }

  @Delete('floor-prices/:id')
  @HttpCode(HttpStatus.NO_CONTENT)
  async deleteFloorPrice(@Param('id') id: string) {
    return this.inventoryService.deleteFloorPrice(id);
  }

  @Post('floor-prices/:id/toggle')
  @HttpCode(HttpStatus.OK)
  async toggleFloorPrice(@Param('id') id: string) {
    return this.inventoryService.toggleFloorPrice(id);
  }

  // =================== Brand Safety Rules ===================

  @Get('brand-safety')
  async findAllBrandSafetyRules(@Query('publisherId') publisherId?: string) {
    return this.inventoryService.findAllBrandSafetyRules(publisherId);
  }

  @Get('brand-safety/:id')
  async findOneBrandSafetyRule(@Param('id') id: string) {
    return this.inventoryService.findOneBrandSafetyRule(id);
  }

  @Post('brand-safety')
  async createBrandSafetyRule(@Body() dto: CreateBrandSafetyRuleDto) {
    return this.inventoryService.createBrandSafetyRule(dto);
  }

  @Put('brand-safety/:id')
  async updateBrandSafetyRule(@Param('id') id: string, @Body() dto: Partial<CreateBrandSafetyRuleDto>) {
    return this.inventoryService.updateBrandSafetyRule(id, dto);
  }

  @Delete('brand-safety/:id')
  @HttpCode(HttpStatus.NO_CONTENT)
  async deleteBrandSafetyRule(@Param('id') id: string) {
    return this.inventoryService.deleteBrandSafetyRule(id);
  }

  @Post('brand-safety/:id/toggle')
  @HttpCode(HttpStatus.OK)
  async toggleBrandSafetyRule(@Param('id') id: string) {
    return this.inventoryService.toggleBrandSafetyRule(id);
  }

  // =================== Statistics ===================

  @Get('stats/:publisherId')
  async getInventoryStats(@Param('publisherId') publisherId: string) {
    return this.inventoryService.getInventoryStats(publisherId);
  }
}
