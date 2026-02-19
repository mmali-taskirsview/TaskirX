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
  UploadedFiles,
  UseInterceptors,
} from '@nestjs/common';
import { FilesInterceptor } from '@nestjs/platform-express';
import { ApiTags, ApiOperation, ApiBearerAuth, ApiQuery, ApiConsumes, ApiBody } from '@nestjs/swagger';
import { CreativesService } from './creatives.service';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { RolesGuard } from '../auth/guards/roles.guard';
import { Roles } from '../auth/decorators/roles.decorator';
import { UserRole } from '../users/user.entity';
import { CreateCreativeDto, UpdateCreativeDto, UpdateCreativeStatsDto } from './dto/create-creative.dto';
import { Express } from 'express';

@ApiTags('creatives')
@Controller('creatives')
@UseGuards(JwtAuthGuard, RolesGuard)
@ApiBearerAuth()
export class CreativesController {
  constructor(private readonly creativesService: CreativesService) {}

  @Post()
  @ApiOperation({ summary: 'Create new creative' })
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async create(@Body() createCreativeDto: CreateCreativeDto, @Request() req) {
    return this.creativesService.create(createCreativeDto, req.user.id, req.user.tenantId);
  }

  @Post('upload')
  @ApiOperation({ summary: 'Upload creative files' })
  @ApiConsumes('multipart/form-data')
  @ApiBody({
    description: 'Creative upload data',
    schema: {
      type: 'object',
      properties: {
        files: {
          type: 'array',
          items: { type: 'string', format: 'binary' },
          description: 'Creative files to upload',
        },
        name: { type: 'string', description: 'Creative name' },
        format: { type: 'string', description: 'Creative format' },
        tags: { type: 'string', description: 'Comma-separated tags' },
        destinationUrl: { type: 'string', description: 'Destination URL' },
        autoOptimize: { type: 'string', description: 'Auto-optimize flag' },
      },
    },
  })
  @UseInterceptors(FilesInterceptor('files'))
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async uploadFiles(
    @UploadedFiles() files: Express.Multer.File[],
    @Body() body: any,
    @Request() req
  ) {
    return this.creativesService.uploadFiles(files, body, req.user.id, req.user.tenantId);
  }

  @Get()
  @ApiOperation({ summary: 'Get all creatives for tenant' })
  @ApiQuery({ name: 'format', required: false })
  @ApiQuery({ name: 'status', required: false })
  @ApiQuery({ name: 'tags', required: false, isArray: true })
  @Roles(UserRole.ADMIN)
  async findAll(
    @Request() req,
    @Query('format') format?: string,
    @Query('status') status?: string,
    @Query('tags') tags?: string[]
  ) {
    if (format) {
      return this.creativesService.findByFormat(format, req.user.tenantId);
    }
    if (status) {
      return this.creativesService.findByStatus(status as any, req.user.tenantId);
    }
    if (tags && tags.length > 0) {
      return this.creativesService.findByTags(tags, req.user.tenantId);
    }
    return this.creativesService.findAll(req.user.tenantId);
  }

  @Get('my')
  @ApiOperation({ summary: 'Get my creatives' })
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async findMy(@Request() req) {
    return this.creativesService.findByUser(req.user.id, req.user.tenantId);
  }

  @Get('top-performing')
  @ApiOperation({ summary: 'Get top performing creatives' })
  @ApiQuery({ name: 'limit', required: false, type: Number })
  async getTopPerforming(@Request() req, @Query('limit') limit?: number) {
    return this.creativesService.getTopPerforming(limit || 10, req.user.tenantId);
  }

  @Get(':id')
  @ApiOperation({ summary: 'Get creative by ID' })
  async findOne(@Param('id') id: string, @Request() req) {
    return this.creativesService.findOne(id, req.user.tenantId);
  }

  @Put(':id')
  @ApiOperation({ summary: 'Update creative' })
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async update(@Param('id') id: string, @Body() updateCreativeDto: UpdateCreativeDto, @Request() req) {
    return this.creativesService.update(id, updateCreativeDto, req.user.tenantId);
  }

  @Put(':id/stats')
  @ApiOperation({ summary: 'Update creative performance stats' })
  async updateStats(@Param('id') id: string, @Body() stats: UpdateCreativeStatsDto, @Request() req) {
    return this.creativesService.updateStats(id, stats, req.user.tenantId);
  }

  @Delete(':id')
  @ApiOperation({ summary: 'Delete creative' })
  @Roles(UserRole.ADVERTISER, UserRole.ADMIN)
  async remove(@Param('id') id: string, @Request() req) {
    await this.creativesService.remove(id, req.user.tenantId);
    return { message: 'Creative deleted successfully' };
  }
}