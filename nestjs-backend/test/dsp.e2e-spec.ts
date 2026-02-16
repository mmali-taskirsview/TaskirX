import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import request from 'supertest';
import { DataSource } from 'typeorm';
import { TestAppModule } from './test-app.module';
import { DspService } from '../src/modules/dsp/dsp.service';

const mockSupplyPartner = {
  id: 'sp_123',
  name: 'Test Supply Partner',
  status: 'active',
  totalRequests: 10,
  totalBids: 5,
  totalWins: 2,
  totalSpend: 12.5,
};

const mockBidResponse = {
  bidId: 'bid_123',
  requestId: 'req_123',
  price: 1.25,
  creativeId: 'creative_123',
  advertiserId: 'adv_1',
  campaignId: 'camp_1',
};

const mockAudience = {
  id: 'aud_1',
  name: 'Test Audience',
  type: 'RETARGETING',
  advertiserId: 'adv_1',
};

const mockDeal = {
  id: 'deal_1',
  name: 'Test Deal',
  type: 'PRIVATE_AUCTION',
  floorPrice: 5.0,
  supplyPartnerId: 'sp_123',
};

const mockBidStrategy = {
  id: 'strat_1',
  name: 'Test Strategy',
  type: 'MAX_REVENUE',
  status: 'active',
};

describe('DspController (e2e)', () => {
  let app: INestApplication;
  const dspService = {
    getDashboardStats: jest.fn().mockResolvedValue({
      totalRequests: 10,
      totalBids: 5,
      totalWins: 2,
      totalSpend: 12.5,
      winRate: 40,
      avgCpm: 6250,
      activeSupplyPartners: 1,
      activeDeals: 0,
      activeAudiences: 0,
      activeBidStrategies: 0,
    }),
    getRtbAnalytics: jest.fn().mockResolvedValue({
      totalBidRequests: 10,
      totalBids: 5,
      avgLatencyMs: 120,
    }),
    getSupplyPartners: jest.fn().mockResolvedValue([mockSupplyPartner]),
    createSupplyPartner: jest.fn().mockResolvedValue(mockSupplyPartner),
    processBidRequest: jest.fn().mockResolvedValue(mockBidResponse),
    recordWin: jest.fn().mockResolvedValue(undefined),
    // Audience Segments
    getAudienceSegments: jest.fn().mockResolvedValue([mockAudience]),
    getAudienceSegment: jest.fn().mockResolvedValue(mockAudience),
    createAudienceSegment: jest.fn().mockResolvedValue(mockAudience),
    updateAudienceSegment: jest.fn().mockResolvedValue(mockAudience),
    deleteAudienceSegment: jest.fn().mockResolvedValue(undefined),
    // Deals
    getDeals: jest.fn().mockResolvedValue([mockDeal]),
    getDeal: jest.fn().mockResolvedValue(mockDeal),
    createDeal: jest.fn().mockResolvedValue(mockDeal),
    updateDeal: jest.fn().mockResolvedValue(mockDeal),
    deleteDeal: jest.fn().mockResolvedValue(undefined),
    // Bid Strategies
    getBidStrategies: jest.fn().mockResolvedValue([mockBidStrategy]),
    getBidStrategy: jest.fn().mockResolvedValue(mockBidStrategy),
    createBidStrategy: jest.fn().mockResolvedValue(mockBidStrategy),
    updateBidStrategy: jest.fn().mockResolvedValue(mockBidStrategy),
    deleteBidStrategy: jest.fn().mockResolvedValue(undefined),
  };

  beforeAll(async () => {
    const moduleFixture: TestingModule = await Test.createTestingModule({
      imports: [TestAppModule],
    })
      .overrideProvider(DspService)
      .useValue(dspService)
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

  it('/dsp/dashboard (GET)', () => {
    return request(app.getHttpServer())
      .get('/dsp/dashboard')
      .expect(200)
      .expect((res) => {
        expect(res.body.totalRequests).toBe(10);
        expect(res.body.activeSupplyPartners).toBe(1);
      });
  });

  it('/dsp/rtb-analytics (GET)', () => {
    return request(app.getHttpServer())
      .get('/dsp/rtb-analytics')
      .expect(200)
      .expect((res) => {
        expect(res.body.totalBidRequests).toBe(10);
      });
  });

  it('/dsp/supply-partners (GET)', () => {
    return request(app.getHttpServer())
      .get('/dsp/supply-partners')
      .expect(200)
      .expect((res) => {
        expect(res.body).toHaveLength(1);
        expect(res.body[0].id).toBe('sp_123');
      });
  });

  it('/dsp/supply-partners (POST)', () => {
    return request(app.getHttpServer())
      .post('/dsp/supply-partners')
      .send({ name: 'Test Supply Partner' })
      .expect(201)
      .expect((res) => {
        expect(res.body.id).toBe('sp_123');
      });
  });

  it('/dsp/bid (POST)', () => {
    return request(app.getHttpServer())
      .post('/dsp/bid')
      .send({
        requestId: 'req_123',
        supplyPartnerId: 'sp_123',
        impressionId: 'imp_123',
        adUnitId: 'unit_1',
        floor: 1,
        device: { type: 'desktop', os: 'windows', browser: 'chrome' },
        geo: { country: 'US', region: 'CA', city: 'SF' },
        user: { id: 'user_1' },
      })
      .expect(201)
      .expect((res) => {
        expect(res.body.bidId).toBe('bid_123');
      });
  });

  it('/dsp/win (POST)', () => {
    return request(app.getHttpServer())
      .post('/dsp/win')
      .send({ bidId: 'bid_123', supplyPartnerId: 'sp_123', price: 1.25 })
      .expect(201)
      .expect((res) => {
        expect(res.body.success).toBe(true);
      });
  });

  it('/dsp/audiences (GET)', () => {
    return request(app.getHttpServer())
      .get('/dsp/audiences')
      .expect(200)
      .expect((res) => {
        expect(res.body).toHaveLength(1);
        expect(res.body[0].id).toBe('aud_1');
      });
  });

  it('/dsp/audiences (POST)', () => {
    return request(app.getHttpServer())
      .post('/dsp/audiences')
      .send({ name: 'Test Audience', type: 'RETARGETING' })
      .expect(201)
      .expect((res) => {
        expect(res.body.id).toBe('aud_1');
      });
  });

  it('/dsp/deals (GET)', () => {
    return request(app.getHttpServer())
      .get('/dsp/deals')
      .expect(200)
      .expect((res) => {
        expect(res.body).toHaveLength(1);
        expect(res.body[0].id).toBe('deal_1');
      });
  });

  it('/dsp/deals (POST)', () => {
    return request(app.getHttpServer())
      .post('/dsp/deals')
      .send({ name: 'Test Deal', type: 'PRIVATE_AUCTION' })
      .expect(201)
      .expect((res) => {
        expect(res.body.id).toBe('deal_1');
      });
  });

  it('/dsp/bid-strategies (GET)', () => {
    return request(app.getHttpServer())
      .get('/dsp/bid-strategies')
      .expect(200)
      .expect((res) => {
        expect(res.body).toHaveLength(1);
        expect(res.body[0].id).toBe('strat_1');
      });
  });

  it('/dsp/bid-strategies (POST)', () => {
    return request(app.getHttpServer())
      .post('/dsp/bid-strategies')
      .send({ name: 'Test Strategy', type: 'MAX_REVENUE' })
      .expect(201)
      .expect((res) => {
        expect(res.body.id).toBe('strat_1');
      });
  });
});
