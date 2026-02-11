import { IsString, IsEnum, IsNumber, IsOptional, IsArray, IsObject, IsUUID } from 'class-validator';
import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import { AdUnitType } from '../entities/ad-unit.entity';

export class CreateAdUnitDto {
  @ApiProperty({ description: 'Ad Unit Name' })
  @IsString()
  name: string;

  @ApiProperty({ enum: AdUnitType })
  @IsEnum(AdUnitType)
  type: AdUnitType;

  @ApiProperty({ description: 'Dimensions (e.g., 300x250)' })
  @IsString()
  size: string;

  @ApiProperty({ description: 'Publisher ID' })
  @IsUUID()
  publisherId: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  domain?: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  pageUrl?: string;

  @ApiPropertyOptional({ default: 0.1 })
  @IsOptional()
  @IsNumber()
  floorPrice?: number;

  @ApiPropertyOptional()
  @IsOptional()
  @IsArray()
  @IsString({ each: true })
  allowedCategories?: string[];

  @ApiPropertyOptional()
  @IsOptional()
  @IsArray()
  @IsString({ each: true })
  blockedCategories?: string[];

  @ApiPropertyOptional()
  @IsOptional()
  @IsArray()
  @IsString({ each: true })
  blockedAdvertisers?: string[];

  @ApiPropertyOptional()
  @IsOptional()
  @IsObject()
  targeting?: Record<string, any>;

  @ApiPropertyOptional()
  @IsOptional()
  @IsObject()
  settings?: Record<string, any>;
}
