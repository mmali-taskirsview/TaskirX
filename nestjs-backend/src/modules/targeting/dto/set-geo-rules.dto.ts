import { IsString, IsObject, IsNotEmpty } from 'class-validator';
import { ApiProperty } from '@nestjs/swagger';

export class SetGeoRulesDto {
  @ApiProperty({ example: 'US', description: 'ISO 2-letter country code' })
  @IsString()
  @IsNotEmpty()
  country: string;

  @ApiProperty({ 
    example: { blocked: false, boost_multiplier: 1.5 }, 
    description: 'Rules object containing blocking status or bid multipliers' 
  })
  @IsObject()
  rules: Record<string, any>;
}
