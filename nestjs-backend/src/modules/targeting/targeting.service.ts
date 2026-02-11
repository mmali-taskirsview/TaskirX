import { Injectable } from '@nestjs/common';
import { RedisService } from '../redis/redis.service';

@Injectable()
export class TargetingService {
  constructor(private readonly redisService: RedisService) {}

  async setUserSegments(userId: string, segments: string[]): Promise<void> {
    await this.redisService.setUserSegments(userId, segments);
  }

  async setGeoRules(countryCode: string, rules: any): Promise<void> {
    await this.redisService.setGeoRules(countryCode, rules);
  }

  async getUserSegments(userId: string): Promise<string[] | null> {
    return this.redisService.get(`user:${userId}:segments`);
  }

  async getGeoRules(countryCode: string): Promise<any | null> {
    return this.redisService.get(`geo:${countryCode}:rules`);
  }
}
