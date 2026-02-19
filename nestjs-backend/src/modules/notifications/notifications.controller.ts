import { Controller, Get, Post, Param, UseGuards, Request } from '@nestjs/common';
import { ApiTags, ApiOperation, ApiBearerAuth } from '@nestjs/swagger';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { NotificationsService } from './notifications.service';

@ApiTags('notifications')
@Controller('notifications')
@UseGuards(JwtAuthGuard)
@ApiBearerAuth()
export class NotificationsController {
  constructor(private readonly notificationsService: NotificationsService) {}

  @Get()
  @ApiOperation({ summary: 'Get my notifications' })
  async getMyNotifications(@Request() req: any) {
    return this.notificationsService.findAllForUser(req.user.id);
  }

  @Get('unread-count')
  @ApiOperation({ summary: 'Get unread count' })
  async getUnreadCount(@Request() req: any) {
    const count = await this.notificationsService.findUnreadCount(req.user.id);
    return { count };
  }

  @Post(':id/read')
  @ApiOperation({ summary: 'Mark notification as read' })
  async markAsRead(@Param('id') id: string, @Request() req: any) {
    await this.notificationsService.markAsRead(id, req.user.id);
    return { success: true };
  }
}
