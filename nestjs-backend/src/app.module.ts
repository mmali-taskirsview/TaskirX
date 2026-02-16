import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ThrottlerModule } from '@nestjs/throttler';
import { AuthModule } from './modules/auth/auth.module';
import { UsersModule } from './modules/users/users.module';
import { CampaignsModule } from './modules/campaigns/campaigns.module';
import { BillingModule } from './modules/billing/billing.module';
import { AnalyticsModule } from './modules/analytics/analytics.module';
import { AiAgentsModule } from './modules/ai-agents/ai-agents.module';
import { HealthModule } from './modules/health/health.module';
import { SspModule } from './modules/ssp/ssp.module';
import { DspModule } from './modules/dsp/dsp.module';
import { IntegrationsModule } from './modules/integrations/integrations.module';
import { RedisModule } from './modules/redis/redis.module';
import { TargetingModule } from './modules/targeting/targeting.module';
import { PrometheusModule } from '@willsoto/nestjs-prometheus';
import * as Joi from 'joi';
import { AppController } from './app.controller';
import { AdvancedController } from './modules/advanced.controller';
import { PaymentsController } from './modules/payments.controller';
import { OnboardingController } from './modules/onboarding.controller';
import { AnalyticsService } from './services/analytics.service';
import { BiddingOptimizationService } from './services/bidding-optimization.service';
import { BillingService } from './services/billing.service';
import { OnboardingService } from './services/onboarding.service';
import { StripeService } from './services/stripe.service';

@Module({
  imports: [
    // Configuration
    ConfigModule.forRoot({
      isGlobal: true,
      validationSchema: Joi.object({
        NODE_ENV: Joi.string()
          .valid('development', 'production', 'test')
          .default('development'),
        PORT: Joi.number().default(4000),
        DATABASE_HOST: Joi.string().required(),
        DATABASE_PORT: Joi.number().default(5432),
        DATABASE_USERNAME: Joi.string().required(),
        DATABASE_PASSWORD: Joi.string().required(),
        DATABASE_NAME: Joi.string().required(),
        REDIS_HOST: Joi.string().default('localhost'),
        REDIS_PORT: Joi.number().default(6379),
        REDIS_PASSWORD: Joi.string().optional(),
        JWT_SECRET: Joi.string().min(32).required(),
        JWT_EXPIRATION: Joi.string().default('24h'),
      }),
    }),

    // Database
    TypeOrmModule.forRootAsync({
      imports: [ConfigModule],
      useFactory: (configService: ConfigService) => ({
        type: 'postgres',
        host: configService.get('DATABASE_HOST'),
        port: configService.get('DATABASE_PORT'),
        username: configService.get('DATABASE_USERNAME'),
        password: configService.get('DATABASE_PASSWORD'),
        database: configService.get('DATABASE_NAME'),
        entities: [__dirname + '/**/*.entity{.ts,.js}'],
        synchronize: configService.get('NODE_ENV') === 'development',
        logging: configService.get('NODE_ENV') === 'development',
      }),
      inject: [ConfigService],
    }),

    // Rate Limiting
    ThrottlerModule.forRoot([{
      ttl: 60000, // 1 minute
      limit: 1000, // 1000 requests per minute
    }]),

    // Monitoring
    PrometheusModule.register({
      path: '/metrics',
      defaultMetrics: {
        enabled: true,
      },
    }),

    // Feature Modules
    AuthModule,
    UsersModule,
    CampaignsModule,
    BillingModule,
    AnalyticsModule,
    AiAgentsModule,
    HealthModule,
    SspModule,
    DspModule,
    IntegrationsModule,
    RedisModule,
    TargetingModule,
  ],
  controllers: [AppController, AdvancedController, PaymentsController, OnboardingController],
  providers: [
    AnalyticsService,
    BiddingOptimizationService,
    BillingService,
    OnboardingService,
    StripeService,
  ],
})
export class AppModule {}
