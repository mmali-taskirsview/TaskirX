import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { AnalyticsModule } from '../src/modules/analytics/analytics.module';
import { AnalyticsService } from '../src/modules/analytics/analytics.service';
import { TestAppModule } from './test-app.module';

// Mock AnalyticsService
const analyticsService = {
  trackMmpEvent: jest.fn().mockResolvedValue(undefined),
  updateRealtimeStats: jest.fn().mockResolvedValue(undefined),
};

describe('MMP Integration (e2e)', () => {
  let app: INestApplication;

  beforeAll(async () => {
    const moduleFixture: TestingModule = await Test.createTestingModule({
      imports: [TestAppModule, AnalyticsModule],
    })
    .overrideProvider(AnalyticsService)
    .useValue(analyticsService)
    .compile();

    app = moduleFixture.createNestApplication();
    await app.init();
  });

  afterAll(async () => {
    await app.close();
  });

  it('/mmp/events/track (POST) - Valid Event', async () => {
    return request(app.getHttpServer())
      .post('/mmp/events/track')
      .send({
        provider: 'appsflyer',
        eventType: 'install',
        campaignId: 'cmp-test-1',
        revenue: 1.99,
        currency: 'USD',
        timestamp: new Date().toISOString()
      })
      .expect(201)
      .expect((res) => {
        expect(res.body.success).toBe(true);
        expect(res.body.conversion_id).toBeDefined();
        // Check if service was called
        expect(analyticsService.trackMmpEvent).toHaveBeenCalledWith(expect.objectContaining({
            provider: 'appsflyer',
            eventType: 'install',
            campaignId: 'cmp-test-1',
            revenue: 1.99
        }));
      });
  });

  it('/mmp/events/track (POST) - Missing Fields', async () => {
      return request(app.getHttpServer())
        .post('/mmp/events/track')
        .send({
            // Missing provider & eventType
            campaignId: 'cmp-fail'
        })
        // Based on controller implementation, it returns { error: ... } but status is 201 by default unless thrown via Exception
        // But let's check response body
        .expect(201) 
        .expect((res) => {
           expect(res.body.error).toBeDefined();
           expect(res.body.error).toContain('Missing required fields');
        });
  });

  it('/mmp/postback (POST) - Generic Postback via Query', async () => {
      return request(app.getHttpServer())
        .post('/mmp/postback')
        .query({
            provider: 'adjust', 
            event_name: 'purchase', 
            c: 'cmp-adjust-1',
            event_revenue: '5.00'
        })
        .expect(201)
        .expect((res) => {
             expect(res.body.status).toBe('ok');
             expect(analyticsService.trackMmpEvent).toHaveBeenCalledWith(expect.objectContaining({
                 provider: 'adjust',
                 eventType: 'purchase',
                 campaignId: 'cmp-adjust-1',
                 revenue: 5.0
             }));
        });
  });
});
