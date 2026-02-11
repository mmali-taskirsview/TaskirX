import { Controller, Post, Body, Get, Param, UseGuards } from '@nestjs/common';
import { ApiTags, ApiOperation, ApiBearerAuth } from '@nestjs/swagger';
import { TargetingService } from './targeting.service';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { RolesGuard } from '../auth/guards/roles.guard';
import { Roles } from '../auth/decorators/roles.decorator';
import { UserRole } from '../users/user.entity';
import { SetUserSegmentsDto } from './dto/set-user-segments.dto';
import { SetGeoRulesDto } from './dto/set-geo-rules.dto';

@ApiTags('targeting')
@Controller('targeting')
@UseGuards(JwtAuthGuard, RolesGuard)
@ApiBearerAuth()
export class TargetingController {
  constructor(private readonly targetingService: TargetingService) {}

  @Post('user-segments')
  @ApiOperation({ summary: 'Set user segments for targeting' })
  @Roles(UserRole.ADMIN)
  async setUserSegments(@Body() dto: SetUserSegmentsDto) {
    await this.targetingService.setUserSegments(dto.userId, dto.segments);
    return { message: 'User segments updated successfully' };
  }

  @Get('user-segments/:userId')
  @ApiOperation({ summary: 'Get user segments' })
  @Roles(UserRole.ADMIN)
  async getUserSegments(@Param('userId') userId: string) {
    const segments = await this.targetingService.getUserSegments(userId);
    return { userId, segments: segments || [] };
  }

  @Post('geo-rules')
  @ApiOperation({ summary: 'Set geo-targeting rules' })
  @Roles(UserRole.ADMIN)
  async setGeoRules(@Body() dto: SetGeoRulesDto) {
    await this.targetingService.setGeoRules(dto.country, dto.rules);
    return { message: 'Geo rules updated successfully' };
  }

  @Get('geo-rules/:country')
  @ApiOperation({ summary: 'Get geo-targeting rules' })
  @Roles(UserRole.ADMIN)
  async getGeoRules(@Param('country') country: string) {
    const rules = await this.targetingService.getGeoRules(country);
    return { country, rules: rules || {} };
  }
}
