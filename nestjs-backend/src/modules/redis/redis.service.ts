import { Injectable, Inject } from '@nestjs/common';
import Redis from 'ioredis';

@Injectable()
export class RedisService {
  constructor(@Inject('REDIS_CLIENT') private readonly redis: Redis) {}

  async set(key: string, value: any, ttl?: number): Promise<void> {
    const data = JSON.stringify(value);
    if (ttl) {
      await this.redis.set(key, data, 'EX', ttl);
    } else {
      await this.redis.set(key, data);
    }
  }

  async get<T>(key: string): Promise<T | null> {
    const data = await this.redis.get(key);
    if (!data) return null;
    return JSON.parse(data);
  }

  async del(key: string): Promise<void> {
    await this.redis.del(key);
  }

  async setUserSegments(userId: string, segments: string[]): Promise<void> {
    const key = `user:${userId}:segments`;
    await this.set(key, segments, 3600); // 1 hour TTL
  }

  async setGeoRules(countryCode: string, rules: any): Promise<void> {
    const key = `geo:${countryCode}:rules`;
    await this.set(key, rules, 86400); // 24 hours TTL
  }
}
