import { Type } from 'class-transformer';
import { IsString, IsNumber, IsEnum, IsOptional, IsDateString, IsObject, Min, IsDate } from 'class-validator';
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

  @ApiProperty({ 
    description: 'Ad Creative Details (Banner, Video, Native, Audio, Rich Media, Playable, Pop, Push)',
    example: { 
      type: 'banner', 
      url: 'https://cdn.example.com/ad.jpg', 
      width: 300, 
      height: 250 
    } 
  })
  @IsOptional()
  @IsObject()
  creative?: {
    type: string;
    url?: string;
    width?: number;
    height?: number;
    duration?: number;
    mimeType?: string;
    bitrate?: number;
    title?: string;
    description?: string;
    iconUrl?: string;
    ctaText?: string;
    htmlSnippet?: string;
    expandable?: boolean;
  };

  @ApiPropertyOptional()
  @IsOptional()
  @Type(() => Date)
  @IsDate()
  startDate?: Date;

  @ApiPropertyOptional()
  @IsOptional()
  @Type(() => Date)
  @IsDate()
  endDate?: Date;

  @ApiPropertyOptional({ description: 'PMP Deal ID' })
  @IsOptional()
  @IsString()
  dealId?: string;
}
