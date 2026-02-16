import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import request from 'supertest';
import { getRepositoryToken } from '@nestjs/typeorm';
import { Repository, DataSource } from 'typeorm';
import { TestAppModule } from './test-app.module';
import { Publisher, PublisherStatus } from './../src/modules/ssp/entities/publisher.entity';
import { AdUnit, AdUnitStatus, AdUnitType } from './../src/modules/ssp/entities/ad-unit.entity';
import { DemandPartner, DemandPartnerStatus, DemandPartnerType } from './../src/modules/ssp/entities/demand-partner.entity';

describe('SspController (e2e)', () => {
  let app: INestApplication;
  let publisherRepository: Repository<Publisher>;
  let adUnitRepository: Repository<AdUnit>;
  let demandPartnerRepository: Repository<DemandPartner>;
  const publisherId = '11111111-1111-1111-1111-111111111111';
  const adUnitId = '22222222-2222-2222-2222-222222222222';
  const demandPartnerId = '33333333-3333-3333-3333-333333333333';

  beforeAll(async () => {
    const moduleFixture: TestingModule = await Test.createTestingModule({
      imports: [TestAppModule],
    }).compile();

    app = moduleFixture.createNestApplication();
    await app.init();

    publisherRepository = moduleFixture.get(getRepositoryToken(Publisher));
    adUnitRepository = moduleFixture.get(getRepositoryToken(AdUnit));
    demandPartnerRepository = moduleFixture.get(getRepositoryToken(DemandPartner));

    await publisherRepository.save({
      id: publisherId,
      name: 'Test Publisher',
      email: 'publisher@test.local',
      status: PublisherStatus.ACTIVE,
    });

    await adUnitRepository.save({
      id: adUnitId,
      name: 'Test Ad Unit',
      type: AdUnitType.BANNER,
      size: '300x250',
      status: AdUnitStatus.ACTIVE,
      publisherId,
      currency: 'USD',
    });

    await demandPartnerRepository.save({
      id: demandPartnerId,
      name: 'Internal DSP',
      code: 'INTERNAL_DSP',
      type: DemandPartnerType.DSP,
      status: DemandPartnerStatus.ACTIVE,
      isGlobal: true,
      bidTimeout: 50,
    });
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

  it('/ssp/auction (POST) - Valid Bid Request', () => {
    return request(app.getHttpServer())
      .post('/ssp/auction')
      .send({
        id: 'req_123',
        publisherId,
        adUnitId,
        device: { type: 'mobile', os: 'ios', browser: 'safari' },
        geo: { country: 'US', region: 'CA', city: 'San Francisco' },
        user: { consent: true },
        floor: 1.49,
      })
      .expect(200)
      .expect((res) => {
        expect(res.body.status).toBe('success');
        expect(res.body.bid.price).toBeCloseTo(1.5, 2);
      });
  });

  it('/ssp/ad (GET) - Valid Tag Request', () => {
    return request(app.getHttpServer())
      .get(`/ssp/ad?pub=${publisherId}&unit=${adUnitId}`)
      .expect(200)
      .expect((res) => {
        expect(res.body.status).toBe('success');
        expect(res.body.ad).toBeDefined();
      });
  });
});
