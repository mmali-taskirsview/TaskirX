import { Injectable, Logger } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Notification } from './notification.entity';
import { UsersService } from '../users/users.service';

@Injectable()
export class NotificationsService {
  private readonly logger = new Logger(NotificationsService.name);

  constructor(
    @InjectRepository(Notification)
    private readonly notificationRepository: Repository<Notification>,
    private readonly usersService: UsersService,
  ) {}

  async create(data: Partial<Notification> & { userId: string }): Promise<Notification | null> {
    try {
      // Check user preferences
      const user = await this.usersService.findById(data.userId);
      const preferences = user.preferences?.notifications || {};

      // Default budget alerts to enabled
      if (data.category === 'budget') {
        if (preferences.budgetAlerts === false) {
          this.logger.debug(`Skipping budget alert for user ${data.userId} based on preferences`);
          return null;
        }
      }

      const notification = this.notificationRepository.create(data);
      return await this.notificationRepository.save(notification);
    } catch (error) {
      this.logger.error(`Failed to create notification: ${error.message}`);
      // Don't throw, just log and return null to prevent blocking the caller
      return null;
    }
  }

  async findAllForUser(userId: string): Promise<Notification[]> {
    return this.notificationRepository.find({
      where: { userId },
      order: { createdAt: 'DESC' },
      take: 50, // Limit to last 50
    });
  }
    
  async markAsRead(id: string, userId: string): Promise<void> {
    await this.notificationRepository.update({ id, userId }, { isRead: true });
  }

  async findUnreadCount(userId: string): Promise<number> {
    return this.notificationRepository.count({
      where: { userId, isRead: false },
    });
  }
}
