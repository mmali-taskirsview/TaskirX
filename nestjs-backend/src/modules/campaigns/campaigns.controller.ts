import {
  Controller,
  Get,
  Post,
  Put,
  Delete,
  Body,
  Param,
  Query,
  UseGuards,
  Request,
} from '@nestjs/common';
import { ApiTags, ApiOperation, ApiBearerAuth, ApiQuery } from '@nestjs/swagger';
import { CampaignsService } from './campaigns.service';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { RolesGuard } from '../auth/guards/roles.guard';
import { Roles } from '../auth/decorators/roles.decorator';
import { UserRole } from '../users/user.entity';
import { CreateCampaignDto } from './dto/create-campaign.dto';

@ApiTags('campaigns')
@Controller('campaigns')
@UseGuards(JwtAuthGuard, RolesGuard)
@ApiBearerAuth()
export class CampaignsController {
  constructor(private readonly campaignsService: CampaignsService) {}

  @Post()
  @ApiOperation({ summary: 'Create new campaign' })
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async create(@Body() createCampaignDto: CreateCampaignDto, @Request() req) {
    return this.campaignsService.create(createCampaignDto, req.user.id, req.user.tenantId);
  }

  @Get()
  @ApiOperation({ summary: 'Get all campaigns for tenant' })
  @ApiQuery({ name: 'includeRealTime', required: false, type: Boolean })
  @Roles(UserRole.ADMIN)
  async findAll(@Request() req, @Query('includeRealTime') includeRealTime?: string) {
    const start = process.hrtime();
    const result = await this.campaignsService.findAll(req.user.tenantId, includeRealTime === 'true');
    // const time = process.hrtime(start);
    // console.log(`Execution time (findAll): ${time[0]}s ${time[1] / 1000000}ms`);
    return result;
  }

  @Get('my')
  @ApiOperation({ summary: 'Get my campaigns' })
  @ApiQuery({ name: 'includeRealTime', required: false, type: Boolean })
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async findMy(@Request() req, @Query('includeRealTime') includeRealTime?: string) {
    // Default to true for user dashboard unless specified otherwise?
    // Let's stick to explicit param, or default true for UI convenience if frontend doesn't send it yet.
    // Given the risk of frontend not sending it, let's default to TRUE for 'my' campaigns.
    const shouldInclude = includeRealTime !== 'false'; // Default true
    return this.campaignsService.findByUser(req.user.id, req.user.tenantId, shouldInclude);
  }

  @Get(':id')
  @ApiOperation({ summary: 'Get campaign by ID' })
  @ApiQuery({ name: 'includeRealTime', required: false, type: Boolean })
  async findOne(@Param('id') id: string, @Request() req, @Query('includeRealTime') includeRealTime?: string) {
    const shouldInclude = includeRealTime !== 'false'; // Default true
    return this.campaignsService.findOne(id, req.user.tenantId, shouldInclude);
  }

  @Put(':id')
  @ApiOperation({ summary: 'Update campaign' })
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async update(@Param('id') id: string, @Body() updateData: any, @Request() req) {
    return this.campaignsService.update(id, updateData, req.user.tenantId);
  }

  @Delete(':id')
  @ApiOperation({ summary: 'Delete campaign' })
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async remove(@Param('id') id: string, @Request() req) {
    await this.campaignsService.remove(id, req.user.tenantId);
    return { message: 'Campaign deleted successfully' };
  }

  @Post(':id/creatives')
  @ApiOperation({ summary: 'Assign creatives to campaign' })
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async assignCreatives(
    @Param('id') campaignId: string,
    @Body() body: { creativeIds: string[] },
    @Request() req
  ) {
    return this.campaignsService.assignCreatives(campaignId, body.creativeIds, req.user.tenantId);
  }

  @Delete(':id/creatives/:creativeId')
  @ApiOperation({ summary: 'Remove creative from campaign' })
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async removeCreative(
    @Param('id') campaignId: string,
    @Param('creativeId') creativeId: string,
    @Request() req
  ) {
    return this.campaignsService.removeCreative(campaignId, creativeId, req.user.tenantId);
  }
}
