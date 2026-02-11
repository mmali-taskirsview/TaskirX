import { IsString, IsArray, ArrayMinSize } from 'class-validator';
import { ApiProperty } from '@nestjs/swagger';

export class SetUserSegmentsDto {
  @ApiProperty({ example: 'user-123', description: 'The unique identifier of the user' })
  @IsString()
  userId: string;

  @ApiProperty({ example: ['vip', 'sports'], description: 'List of segment IDs' })
  @IsArray()
  @IsString({ each: true })
  @ArrayMinSize(1)
  segments: string[];
}
