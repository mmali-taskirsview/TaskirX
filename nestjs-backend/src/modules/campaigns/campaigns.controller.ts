import {
  Controller,
  Get,
  Post,
  Put,
  Delete,
  Body,
  Param,
  UseGuards,
  Request,
} from '@nestjs/common';
import { ApiTags, ApiOperation, ApiBearerAuth } from '@nestjs/swagger';
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
  @Roles(UserRole.ADMIN)
  async findAll(@Request() req) {
    return this.campaignsService.findAll(req.user.tenantId);
  }

  @Get('my')
  @ApiOperation({ summary: 'Get my campaigns' })
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async findMy(@Request() req) {
    return this.campaignsService.findByUser(req.user.id, req.user.tenantId);
  }

  @Get(':id')
  @ApiOperation({ summary: 'Get campaign by ID' })
  async findOne(@Param('id') id: string, @Request() req) {
    return this.campaignsService.findOne(id, req.user.tenantId);
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
}
