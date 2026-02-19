import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Creative, CreativeStatus, CreativeType, CreativeFormat } from './creative.entity';
import { CreateCreativeDto, UpdateCreativeDto, UpdateCreativeStatsDto } from './dto/create-creative.dto';
import { Express } from 'express';
import * as path from 'path';
import * as fs from 'fs';

@Injectable()
export class CreativesService {
  constructor(
    @InjectRepository(Creative)
    private creativesRepository: Repository<Creative>,
  ) {}

  async create(createCreativeDto: CreateCreativeDto, userId: string, tenantId: string): Promise<Creative> {
    const creative = this.creativesRepository.create({
      ...createCreativeDto,
      userId,
      tenantId,
      status: createCreativeDto.status || CreativeStatus.DRAFT,
    });

    return this.creativesRepository.save(creative);
  }

  async findAll(tenantId: string): Promise<Creative[]> {
    return this.creativesRepository.find({
      where: { tenantId },
      order: { createdAt: 'DESC' },
    });
  }

  async findByUser(userId: string, tenantId: string): Promise<Creative[]> {
    return this.creativesRepository.find({
      where: { userId, tenantId },
      order: { createdAt: 'DESC' },
    });
  }

  async findOne(id: string, tenantId: string): Promise<Creative> {
    const creative = await this.creativesRepository.findOne({
      where: { id, tenantId },
    });

    if (!creative) {
      throw new NotFoundException('Creative not found');
    }

    return creative;
  }

  async update(id: string, updateCreativeDto: UpdateCreativeDto, tenantId: string): Promise<Creative> {
    await this.findOne(id, tenantId); // Check if exists

    await this.creativesRepository.update(id, updateCreativeDto);
    return this.findOne(id, tenantId);
  }

  async updateStats(id: string, stats: UpdateCreativeStatsDto, tenantId: string): Promise<Creative> {
    const creative = await this.findOne(id, tenantId);

    // Calculate CTR and CVR
    const totalImpressions = (creative.impressions || 0) + (stats.impressions || 0);
    const totalClicks = (creative.clicks || 0) + (stats.clicks || 0);
    const totalConversions = (creative.conversions || 0) + (stats.conversions || 0);

    const ctr = totalImpressions > 0 ? totalClicks / totalImpressions : 0;
    const cvr = totalClicks > 0 ? totalConversions / totalClicks : 0;

    await this.creativesRepository.update(id, {
      impressions: totalImpressions,
      clicks: totalClicks,
      conversions: totalConversions,
      ctr,
      cvr,
    });

    return this.findOne(id, tenantId);
  }

  async remove(id: string, tenantId: string): Promise<void> {
    const creative = await this.findOne(id, tenantId);
    await this.creativesRepository.remove(creative);
  }

  async findByTags(tags: string[], tenantId: string): Promise<Creative[]> {
    return this.creativesRepository
      .createQueryBuilder('creative')
      .where('creative.tenantId = :tenantId', { tenantId })
      .andWhere('creative.tags && :tags', { tags })
      .orderBy('creative.createdAt', 'DESC')
      .getMany();
  }

  async findByFormat(format: string, tenantId: string): Promise<Creative[]> {
    return this.creativesRepository.find({
      where: { format: format as any, tenantId },
      order: { createdAt: 'DESC' },
    });
  }

  async findByStatus(status: CreativeStatus, tenantId: string): Promise<Creative[]> {
    return this.creativesRepository.find({
      where: { status, tenantId },
      order: { createdAt: 'DESC' },
    });
  }

  async getTopPerforming(limit: number = 10, tenantId: string): Promise<Creative[]> {
    return this.creativesRepository.find({
      where: { tenantId },
      order: { ctr: 'DESC', impressions: 'DESC' },
      take: limit,
    });
  }

  async uploadFiles(
    files: Express.Multer.File[],
    body: any,
    userId: string,
    tenantId: string
  ): Promise<Creative[]> {
    const creatives: Creative[] = [];

    for (const file of files) {
      // Determine creative type and format from file
      let type: CreativeType;
      let format: CreativeFormat;
      let dimensions: { width: number; height: number } | undefined;

      if (file.mimetype.startsWith('image/')) {
        type = CreativeType.IMAGE;
        const formatResult = this.determineImageFormat(file);
        format = formatResult.format;
        dimensions = formatResult.dimensions;
      } else if (file.mimetype.startsWith('video/')) {
        type = CreativeType.VIDEO;
        format = CreativeFormat.REWARDED_VIDEO;
      } else if (file.originalname.endsWith('.html') || file.originalname.endsWith('.zip')) {
        type = CreativeType.HTML5;
        format = CreativeFormat.RICH_MEDIA;
      } else {
        type = CreativeType.PLAYABLE;
        format = CreativeFormat.PLAYABLE;
      }

      // File is already saved by multer, use the path
      const relativePath = path.relative(process.cwd(), file.path);

      // Parse tags
      const tags = body.tags ? body.tags.split(',').map((tag: string) => tag.trim()) : [];

      // Calculate file size in MB
      const fileSizeMB = file.size / (1024 * 1024);

      // Create creative record
      const creative = this.creativesRepository.create({
        name: body.name || file.originalname.split('.')[0],
        description: `Uploaded creative: ${file.originalname}`,
        type,
        format,
        url: `/uploads/creatives/${path.basename(file.path)}`,
        dimensions,
        fileSize: fileSizeMB,
        status: CreativeStatus.PENDING,
        tags,
        metadata: {
          clickUrl: body.destinationUrl || '',
        },
        userId,
        tenantId,
        // Performance metrics will be initialized to 0 by default
      });

      const savedCreative = await this.creativesRepository.save(creative);
      creatives.push(savedCreative);
    }

    return creatives;
  }

  private determineImageFormat(file: Express.Multer.File): { format: CreativeFormat; dimensions?: { width: number; height: number } } {
    // This is a simple implementation - in production you'd want more sophisticated detection
    const fileName = file.originalname.toLowerCase();
    
    if (fileName.includes('banner') || fileName.includes('300x250')) {
      return { 
        format: CreativeFormat.BANNER, 
        dimensions: { width: 300, height: 250 } 
      };
    } else if (fileName.includes('320x50')) {
      return { 
        format: CreativeFormat.BANNER, 
        dimensions: { width: 320, height: 50 } 
      };
    } else if (fileName.includes('728x90') || fileName.includes('leaderboard')) {
      return { 
        format: CreativeFormat.BANNER, 
        dimensions: { width: 728, height: 90 } 
      };
    }
    
    return { format: CreativeFormat.BANNER };
  }
}