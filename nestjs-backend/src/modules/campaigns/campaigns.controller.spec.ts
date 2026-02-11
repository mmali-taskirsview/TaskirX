import { Test, TestingModule } from '@nestjs/testing';
import { CampaignsController } from './campaigns.controller';
import { CampaignsService } from './campaigns.service';
import { Campaign, CampaignStatus } from './campaign.entity';

describe('CampaignsController', () => {
  let controller: CampaignsController;
  let service: CampaignsService;

  const mockCampaign: any = {
    id: 'uuid-1',
    name: 'Test Campaign',
    budget: 1000,
    status: CampaignStatus.ACTIVE,
    advertiserId: 'adv-1',
    dailyBudget: 100,
    startTime: new Date(),
    endTime: new Date(),
    spent: 0,
    targeting: {
      countries: ['US'],
      devices: ['mobile'],
      categories: ['tech'],
    },
    creativeUrl: 'http://example.com/ad.jpg',
    adDomain: ['example.com'],
    createdAt: new Date(),
    updatedAt: new Date(),
  };

  const mockService = {
    create: jest.fn().mockResolvedValue(mockCampaign),
    findByUser: jest.fn().mockResolvedValue([mockCampaign]),
    findOne: jest.fn().mockResolvedValue(mockCampaign),
    findAll: jest.fn().mockResolvedValue([mockCampaign]), // Added missing mock
    update: jest.fn().mockResolvedValue({ ...mockCampaign, name: 'Updated' }), // ...existing code...
    remove: jest.fn().mockResolvedValue(undefined),
    pause: jest.fn().mockResolvedValue({ ...mockCampaign, status: CampaignStatus.PAUSED }),
    resume: jest.fn().mockResolvedValue(mockCampaign),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [CampaignsController],
      providers: [
        {
          provide: CampaignsService,
          useValue: mockService,
        },
      ],
    }).compile();

    controller = module.get<CampaignsController>(CampaignsController);
    service = module.get<CampaignsService>(CampaignsService);
  });

  it('should be defined', () => {
    expect(controller).toBeDefined();
  });

  describe('findAll', () => {
    it('should return an array of campaigns', async () => {
      const mockReq = { user: { id: 'user-1', tenantId: 'tenant-1' } };
      const result = await controller.findAll(mockReq); // findAll takes only req in controller code shown
      expect(result).toEqual([mockCampaign]);
      expect(mockService.findAll).toHaveBeenCalledWith('tenant-1');
    });
  });

  describe('findOne', () => {
    it('should return a single campaign', async () => {
      const mockReq = { user: { id: 'user-1', tenantId: 'tenant-1' } };
      const result = await controller.findOne('uuid-1', mockReq);
      expect(result).toEqual(mockCampaign);
      expect(mockService.findOne).toHaveBeenCalledWith('uuid-1', 'tenant-1');
    });
  });
});
