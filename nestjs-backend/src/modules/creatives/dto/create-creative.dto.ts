import { Type } from 'class-transformer';
import { IsString, IsEnum, IsOptional, IsNumber, IsObject, IsArray, IsUrl, Min, IsUUID } from 'class-validator';
import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import { CreativeType, CreativeStatus, CreativeFormat } from '../creative.entity';

export class CreateCreativeDto {
  @ApiProperty({ description: 'Creative name' })
  @IsString()
  name: string;

  @ApiPropertyOptional({ description: 'Creative description' })
  @IsOptional()
  @IsString()
  description?: string;

  @ApiProperty({ enum: CreativeType })
  @IsEnum(CreativeType)
  type: CreativeType;

  @ApiProperty({ enum: CreativeFormat })
  @IsEnum(CreativeFormat)
  format: CreativeFormat;

  @ApiProperty({ description: 'Creative asset URL' })
  @IsUrl()
  url: string;

  @ApiPropertyOptional({ description: 'Thumbnail URL' })
  @IsOptional()
  @IsUrl()
  thumbnailUrl?: string;

  @ApiPropertyOptional({ description: 'Creative dimensions' })
  @IsOptional()
  @IsObject()
  dimensions?: {
    width: number;
    height: number;
  };

  @ApiPropertyOptional({ description: 'File size in MB' })
  @IsOptional()
  @IsNumber()
  @Min(0)
  fileSize?: number;

  @ApiPropertyOptional({ enum: CreativeStatus, default: CreativeStatus.DRAFT })
  @IsOptional()
  @IsEnum(CreativeStatus)
  status?: CreativeStatus;

  @ApiPropertyOptional({ description: 'Creative metadata' })
  @IsOptional()
  @IsObject()
  metadata?: {
    duration?: number;
    clickUrl?: string;
    impressionTrackingUrls?: string[];
    clickTrackingUrls?: string[];
    thirdPartyTags?: string[];
  };

  @ApiPropertyOptional({ description: 'Targeting rules' })
  @IsOptional()
  @IsObject()
  targeting?: {
    categories?: string[];
    ageGroups?: string[];
    genders?: string[];
    devices?: string[];
  };

  @ApiPropertyOptional({ description: 'Tags for organization' })
  @IsOptional()
  @IsArray()
  @IsString({ each: true })
  tags?: string[];
}

export class UpdateCreativeDto {
  @ApiPropertyOptional({ description: 'Creative name' })
  @IsOptional()
  @IsString()
  name?: string;

  @ApiPropertyOptional({ description: 'Creative description' })
  @IsOptional()
  @IsString()
  description?: string;

  @ApiPropertyOptional({ enum: CreativeType })
  @IsOptional()
  @IsEnum(CreativeType)
  type?: CreativeType;

  @ApiPropertyOptional({ enum: CreativeFormat })
  @IsOptional()
  @IsEnum(CreativeFormat)
  format?: CreativeFormat;

  @ApiPropertyOptional({ description: 'Creative asset URL' })
  @IsOptional()
  @IsUrl()
  url?: string;

  @ApiPropertyOptional({ description: 'Thumbnail URL' })
  @IsOptional()
  @IsUrl()
  thumbnailUrl?: string;

  @ApiPropertyOptional({ description: 'Creative dimensions' })
  @IsOptional()
  @IsObject()
  dimensions?: {
    width: number;
    height: number;
  };

  @ApiPropertyOptional({ description: 'File size in MB' })
  @IsOptional()
  @IsNumber()
  @Min(0)
  fileSize?: number;

  @ApiPropertyOptional({ enum: CreativeStatus })
  @IsOptional()
  @IsEnum(CreativeStatus)
  status?: CreativeStatus;

  @ApiPropertyOptional({ description: 'Creative metadata' })
  @IsOptional()
  @IsObject()
  metadata?: {
    duration?: number;
    clickUrl?: string;
    impressionTrackingUrls?: string[];
    clickTrackingUrls?: string[];
    thirdPartyTags?: string[];
  };

  @ApiPropertyOptional({ description: 'Targeting rules' })
  @IsOptional()
  @IsObject()
  targeting?: {
    categories?: string[];
    ageGroups?: string[];
    genders?: string[];
    devices?: string[];
  };

  @ApiPropertyOptional({ description: 'Tags for organization' })
  @IsOptional()
  @IsArray()
  @IsString({ each: true })
  tags?: string[];
}

export class UpdateCreativeStatsDto {
  @ApiPropertyOptional({ description: 'Number of impressions' })
  @IsOptional()
  @IsNumber()
  @Min(0)
  impressions?: number;

  @ApiPropertyOptional({ description: 'Number of clicks' })
  @IsOptional()
  @IsNumber()
  @Min(0)
  clicks?: number;

  @ApiPropertyOptional({ description: 'Number of conversions' })
  @IsOptional()
  @IsNumber()
  @Min(0)
  conversions?: number;
}