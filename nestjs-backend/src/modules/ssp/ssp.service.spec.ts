import { Test, TestingModule } from '@nestjs/testing';
import { getRepositoryToken } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { SspService, BidResponse } from './ssp.service';
import { Publisher } from './entities/publisher.entity';
import { AdUnit } from './entities/ad-unit.entity';
import { Placement } from './entities/placement.entity';
import { FloorPrice } from './entities/floor-price.entity';
import { DemandPartner } from './entities/demand-partner.entity';

const mockRepository = () => ({
  findOne: jest.fn(),
  find: jest.fn(),
  save: jest.fn(),
});

describe('SspService', () => {
  let service: SspService;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        SspService,
        { provide: getRepositoryToken(Publisher), useFactory: mockRepository },
        { provide: getRepositoryToken(AdUnit), useFactory: mockRepository },
        { provide: getRepositoryToken(Placement), useFactory: mockRepository },
        { provide: getRepositoryToken(FloorPrice), useFactory: mockRepository },
        { provide: getRepositoryToken(DemandPartner), useFactory: mockRepository },
      ],
    }).compile();

    service = module.get<SspService>(SspService);
  });

  it('should be defined', () => {
    expect(service).toBeDefined();
  });

  describe('selectWinner (Second Price Auction)', () => {
    const floorPrice = 1.0;

    it('should return null if no bids provided', () => {
      const result = service.selectWinner([], floorPrice);
      expect(result).toBeNull();
    });

    it('should set price to floor + 0.01 if only one bid exists', () => {
      const bids: BidResponse[] = [
        { bidId: '1', partnerId: 'p1', partnerName: 'P1', price: 5.0, creative: {} as any },
      ];

      const winner = service.selectWinner(bids, floorPrice);
      expect(winner).toBeDefined();
      expect(winner.partnerId).toBe('p1');
      expect(winner.price).toBeCloseTo(floorPrice + 0.01, 2);
    });

    it('should set price to second bid + 0.01 if multiple bids exist', () => {
      const bids: BidResponse[] = [
        { bidId: '1', partnerId: 'p1', partnerName: 'P1', price: 5.0, creative: {} as any },
        { bidId: '2', partnerId: 'p2', partnerName: 'P2', price: 3.0, creative: {} as any },
      ];

      const winner = service.selectWinner(bids, floorPrice);
      expect(winner).toBeDefined();
      expect(winner.partnerId).toBe('p1'); // Highest bidder wins
      expect(winner.price).toBeCloseTo(3.01, 2); // Pays second price + 0.01
    });

    it('should use floor price if second bid is lower than floor', () => {
      // This is edge case. Logic says winner > floor (filtered before), but if 2nd bid < floor?
      // Nest logic: Math.max(sorted[1].price, floorPrice)
      const bids: BidResponse[] = [
        { bidId: '1', partnerId: 'p1', partnerName: 'P1', price: 5.0, creative: {} as any },
        { bidId: '2', partnerId: 'p2', partnerName: 'P2', price: 0.5, creative: {} as any },
      ];

      const winner = service.selectWinner(bids, floorPrice);
      expect(winner.price).toBeCloseTo(1.01, 2); // Floor (1.0) is higher than 0.5
    });
  });
});
