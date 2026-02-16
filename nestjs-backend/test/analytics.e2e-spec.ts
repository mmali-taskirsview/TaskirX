import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import request from 'supertest';
import { DataSource } from 'typeorm';
import { TestAppModule } from './test-app.module';
import { AnalyticsModule } from '../src/modules/analytics/analytics.module';
import { AnalyticsService } from '../src/modules/analytics/analytics.service';
import { JwtAuthGuard } from '../src/modules/auth/guards/jwt-auth.guard';

const analyticsService = {
  getDashboardStats: jest.fn().mockResolvedValue({
    totalImpressions: 100,
    totalClicks: 5,
    totalConversions: 1,
    totalSpend: '25.00',
    activeCampaigns: 1,
    avgCtr: '5.00',
    avgCpc: '5.00',
  }),
  getCampaignStats: jest.fn().mockResolvedValue({
    campaignId: 'camp_1',
    impressions: 10,
    clicks: 1,
    conversions: 0,
    spend: 5,
    ctr: '10.00',
    cpc: '5.00',
    cpa: 0,
  }),
  getRevenueByDate: jest.fn().mockResolvedValue([
    { date: '2026-02-13', revenue: '10.00', impressions: 10, clicks: 1 },
  ]),
  getTopPerformingCampaigns: jest.fn().mockResolvedValue([
    { campaignId: 'camp_1', impressions: 10, clicks: 1 },
  ]),
  trackImpression: jest.fn().mockResolvedValue(undefined),
  trackClick: jest.fn().mockResolvedValue(undefined),
  trackConversion: jest.fn().mockResolvedValue(undefined),
};

const authGuard = {
  canActivate: (context) => {
    const req = context.switchToHttp().getRequest();
    req.user = { tenantId: 'tenant_1' };
    return true;
  },
};

describe('AnalyticsController (e2e)', () => {
  let app: INestApplication;

  beforeAll(async () => {
    const moduleFixture: TestingModule = await Test.createTestingModule({
      imports: [TestAppModule, AnalyticsModule],
    })
      .overrideProvider(AnalyticsService)
      .useValue(analyticsService)
      .overrideGuard(JwtAuthGuard)
      .useValue(authGuard)
      .compile();

    app = moduleFixture.createNestApplication();
    await app.init();
  });

  afterAll(async () => {
    if (app) {
      const dataSource = app.get(DataSource);
      await app.close();
      if (dataSource && dataSource.isInitialized) {
        await dataSource.destroy();
      }
    }
  });

  it('/analytics/dashboard (GET)', () => {
    return request(app.getHttpServer())
      .get('/analytics/dashboard')
      .expect(200)
      .expect((res) => {
        expect(res.body.totalImpressions).toBe(100);
      });
  });

  it('/analytics/campaign/:id (GET)', () => {
    return request(app.getHttpServer())
      .get('/analytics/campaign/camp_1')
      .expect(200)
      .expect((res) => {
        expect(res.body.campaignId).toBe('camp_1');
      });
  });

  it('/analytics/revenue (GET)', () => {
    return request(app.getHttpServer())
      .get('/analytics/revenue')
      .expect(200)
      .expect((res) => {
        expect(res.body).toHaveLength(1);
        expect(res.body[0].revenue).toBe('10.00');
      });
  });

  it('/analytics/top-campaigns (GET)', () => {
    return request(app.getHttpServer())
      .get('/analytics/top-campaigns')
      .expect(200)
      .expect((res) => {
        expect(res.body[0].campaignId).toBe('camp_1');
      });
  });

  it('/analytics/track/impression (POST)', () => {
    return request(app.getHttpServer())
      .post('/analytics/track/impression')
      .send({ campaignId: 'camp_1', timestamp: new Date().toISOString() })
      .expect(201)
      .expect((res) => {
        expect(res.body.success).toBe(true);
      });
  });

  it('/analytics/track/click (POST)', () => {
    return request(app.getHttpServer())
      .post('/analytics/track/click')
      .send({ clickId: 'click_1', campaignId: 'camp_1', timestamp: new Date().toISOString() })
      .expect(201)
      .expect((res) => {
        expect(res.body.success).toBe(true);
      });
  });

  it('/analytics/track/conversion (POST)', () => {
    return request(app.getHttpServer())
      .post('/analytics/track/conversion')
      .send({ conversionId: 'conv_1', campaignId: 'camp_1', value: 25, timestamp: new Date().toISOString() })
      .expect(201)
      .expect((res) => {
        expect(res.body.success).toBe(true);
      });
  });
});
