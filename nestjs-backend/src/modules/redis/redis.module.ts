import { Module, Global } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import Redis from 'ioredis';
import { RedisService } from './redis.service';

@Global()
@Module({
  imports: [ConfigModule],
  providers: [
    {
      provide: 'REDIS_CLIENT',
      useFactory: (configService: ConfigService) => {
        const host = configService.get('REDIS_HOST');
        const port = configService.get('REDIS_PORT');
        const password = configService.get('REDIS_PASSWORD');
        
        console.log(`Current Redis Config: Host=${host}, Port=${port}, Password=${password ? '******' : 'MISSING'}`);
        
        return new Redis({
          host,
          port,
          password,
        });
      },
      inject: [ConfigService],
    },
    RedisService,
  ],
  exports: ['REDIS_CLIENT', RedisService],
})
export class RedisModule {}
