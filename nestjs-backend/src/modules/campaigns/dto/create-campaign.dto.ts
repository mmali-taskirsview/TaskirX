import { IsString, IsNumber, IsEnum, IsOptional, IsDateString, IsObject, Min } from 'class-validator';
import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import { CampaignStatus, CampaignType } from '../campaign.entity';

export class CreateCampaignDto {
  @ApiProperty({ description: 'Campaign Name' })
  @IsString()
  name: string;

  @ApiPropertyOptional({ description: 'Description' })
  @IsOptional()
  @IsString()
  description?: string;

  @ApiProperty({ enum: CampaignStatus, default: CampaignStatus.DRAFT })
  @IsOptional()
  @IsEnum(CampaignStatus)
  status?: CampaignStatus;

  @ApiProperty({ enum: CampaignType })
  @IsEnum(CampaignType)
  type: CampaignType;

  @ApiProperty({ description: 'Total Budget' })
  @IsNumber()
  @Min(0)
  budget: number;

  @ApiProperty({ description: 'Bid Price (CPM/CPC/CPA)' })
  @IsNumber()
  @Min(0)
  bidPrice: number;

  @ApiProperty({ description: 'Industry Vertical', example: 'GAMING' })
  @IsString()
  // @IsIn(VERTICAL_KEYS) // Optional: enforce strict vertical keys
  vertical: string;

  @ApiProperty({ description: 'Targeting Rules (Geo, Device, Audience)' })
  @IsOptional()
  @IsObject()
  targeting?: Record<string, any>;

  @ApiPropertyOptional()
  @IsOptional()
  @IsDateString()
  startDate?: Date;

  @ApiPropertyOptional()
  @IsOptional()
  @IsDateString()
  endDate?: Date;
}
