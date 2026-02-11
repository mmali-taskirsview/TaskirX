import { IsString, IsEnum, IsOptional, IsNumber, IsObject } from 'class-validator';
import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import { AudienceSegmentType, AudienceSegmentStatus } from '../entities/audience-segment.entity';

export class CreateAudienceDto {
  @ApiProperty({ description: 'Segment Name' })
  @IsString()
  name: string;

  @ApiPropertyOptional({ description: 'Description' })
  @IsOptional()
  @IsString()
  description?: string;

  @ApiProperty({ enum: AudienceSegmentType, default: AudienceSegmentType.FIRST_PARTY })
  @IsEnum(AudienceSegmentType)
  type: AudienceSegmentType;

  @ApiPropertyOptional({ enum: AudienceSegmentStatus, default: AudienceSegmentStatus.BUILDING })
  @IsOptional()
  @IsEnum(AudienceSegmentStatus)
  status?: AudienceSegmentStatus;

  @ApiPropertyOptional({ description: 'Advertiser ID' })
  @IsOptional()
  @IsString()
  advertiserId?: string;

  @ApiPropertyOptional({ description: 'CPM Adjustment Modifier', default: 1.0 })
  @IsOptional()
  @IsNumber()
  cpmModifier?: number;

  @ApiPropertyOptional({ description: 'Lookback Window (Days)', default: 30 })
  @IsOptional()
  @IsNumber()
  lookbackDays?: number;

  @ApiPropertyOptional({ description: 'Targeting Rules' })
  @IsOptional()
  @IsObject()
  rules?: Record<string, any>;
}
