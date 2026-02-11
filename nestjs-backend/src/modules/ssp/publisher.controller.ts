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
import { PublisherService, CreatePublisherDto, UpdatePublisherDto } from './publisher.service';
import { PublisherStatus } from './entities/publisher.entity';

@Controller('ssp/publishers')
export class PublisherController {
  constructor(private readonly publisherService: PublisherService) {}

  @Get()
  async findAll(@Query('status') status?: PublisherStatus) {
    return this.publisherService.findAll(status);
  }

  @Get(':id')
  async findOne(@Param('id') id: string) {
    return this.publisherService.findOne(id);
  }

  @Get(':id/stats')
  async getStatistics(@Param('id') id: string) {
    return this.publisherService.getStatistics(id);
  }

  @Post()
  async create(@Body() createDto: CreatePublisherDto) {
    return this.publisherService.create(createDto);
  }

  @Put(':id')
  async update(
    @Param('id') id: string,
    @Body() updateDto: UpdatePublisherDto,
  ) {
    return this.publisherService.update(id, updateDto);
  }

  @Delete(':id')
  @HttpCode(HttpStatus.NO_CONTENT)
  async delete(@Param('id') id: string) {
    return this.publisherService.delete(id);
  }

  @Post(':id/approve')
  @HttpCode(HttpStatus.OK)
  async approve(@Param('id') id: string) {
    return this.publisherService.approve(id);
  }

  @Post(':id/suspend')
  @HttpCode(HttpStatus.OK)
  async suspend(@Param('id') id: string) {
    return this.publisherService.suspend(id);
  }

  @Post(':id/reject')
  @HttpCode(HttpStatus.OK)
  async reject(@Param('id') id: string) {
    return this.publisherService.reject(id);
  }

  @Post(':id/regenerate-api-key')
  @HttpCode(HttpStatus.OK)
  async regenerateApiKey(@Param('id') id: string) {
    return this.publisherService.regenerateApiKey(id);
  }

  @Post('validate-api-key')
  @HttpCode(HttpStatus.OK)
  async validateApiKey(@Body('apiKey') apiKey: string) {
    const publisher = await this.publisherService.validateApiKey(apiKey);
    return {
      valid: !!publisher,
      publisher: publisher ? {
        id: publisher.id,
        name: publisher.name,
        tier: publisher.tier,
      } : null,
    };
  }
}
