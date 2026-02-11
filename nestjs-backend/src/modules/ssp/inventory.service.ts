import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { AdUnit, AdUnitStatus } from './entities/ad-unit.entity';
import { Placement, PlacementStatus, PlacementPosition } from './entities/placement.entity';
import { FloorPrice, FloorPriceRuleType, FloorPriceAction } from './entities/floor-price.entity';
import { BrandSafetyRule, BrandSafetyRuleType, BrandSafetyTarget } from './entities/brand-safety-rule.entity';
import { CreateAdUnitDto } from './dto/create-ad-unit.dto';

export interface CreatePlacementDto {
  name: string;
  publisherId: string;
  description?: string;
  position?: PlacementPosition;
  domain?: string;
  pageType?: string;
  adFormats?: string[];
  allowedSizes?: string[];
  devices?: string[];
  floorPrice?: number;
  viewabilityTarget?: number;
  targeting?: Record<string, any>;
  settings?: Record<string, any>;
}

export interface CreateFloorPriceDto {
  name: string;
  publisherId: string;
  description?: string;
  ruleType: FloorPriceRuleType;
  price: number;
  action: FloorPriceAction;
  conditions?: Record<string, any>;
  adUnitId?: string;
  placementId?: string;
  priority?: number;
}

export interface CreateBrandSafetyRuleDto {
  name: string;
  publisherId: string;
  description?: string;
  ruleType: BrandSafetyRuleType;
  target: BrandSafetyTarget;
  values: string[];
  adUnitId?: string;
  priority?: number;
}

@Injectable()
export class InventoryService {
  constructor(
    @InjectRepository(AdUnit)
    private adUnitRepository: Repository<AdUnit>,
    @InjectRepository(Placement)
    private placementRepository: Repository<Placement>,
    @InjectRepository(FloorPrice)
    private floorPriceRepository: Repository<FloorPrice>,
    @InjectRepository(BrandSafetyRule)
    private brandSafetyRepository: Repository<BrandSafetyRule>,
  ) {}

  // =================== Ad Units ===================

  async findAllAdUnits(publisherId?: string): Promise<AdUnit[]> {
    const where = publisherId ? { publisherId } : {};
    return this.adUnitRepository.find({
      where,
      order: { createdAt: 'DESC' },
    });
  }

  async findOneAdUnit(id: string): Promise<AdUnit> {
    const adUnit = await this.adUnitRepository.findOne({ where: { id } });
    if (!adUnit) {
      throw new NotFoundException(`Ad unit with ID ${id} not found`);
    }
    return adUnit;
  }

  async createAdUnit(dto: CreateAdUnitDto): Promise<AdUnit> {
    const adUnit = this.adUnitRepository.create({
      ...dto,
      status: AdUnitStatus.PENDING,
    });
    return this.adUnitRepository.save(adUnit);
  }

  async updateAdUnit(id: string, dto: Partial<CreateAdUnitDto>): Promise<AdUnit> {
    const adUnit = await this.findOneAdUnit(id);
    Object.assign(adUnit, dto);
    return this.adUnitRepository.save(adUnit);
  }

  async deleteAdUnit(id: string): Promise<void> {
    const adUnit = await this.findOneAdUnit(id);
    await this.adUnitRepository.remove(adUnit);
  }

  async activateAdUnit(id: string): Promise<AdUnit> {
    return this.updateAdUnit(id, { status: AdUnitStatus.ACTIVE } as any);
  }

  async pauseAdUnit(id: string): Promise<AdUnit> {
    return this.updateAdUnit(id, { status: AdUnitStatus.PAUSED } as any);
  }

  // =================== Placements ===================

  async findAllPlacements(publisherId?: string): Promise<Placement[]> {
    const where = publisherId ? { publisherId } : {};
    return this.placementRepository.find({
      where,
      order: { createdAt: 'DESC' },
    });
  }

  async findOnePlacement(id: string): Promise<Placement> {
    const placement = await this.placementRepository.findOne({ where: { id } });
    if (!placement) {
      throw new NotFoundException(`Placement with ID ${id} not found`);
    }
    return placement;
  }

  async createPlacement(dto: CreatePlacementDto): Promise<Placement> {
    const placement = this.placementRepository.create({
      ...dto,
      status: PlacementStatus.ACTIVE,
    });
    return this.placementRepository.save(placement);
  }

  async updatePlacement(id: string, dto: Partial<CreatePlacementDto>): Promise<Placement> {
    const placement = await this.findOnePlacement(id);
    Object.assign(placement, dto);
    return this.placementRepository.save(placement);
  }

  async deletePlacement(id: string): Promise<void> {
    const placement = await this.findOnePlacement(id);
    await this.placementRepository.remove(placement);
  }

  // =================== Floor Prices ===================

  async findAllFloorPrices(publisherId?: string): Promise<FloorPrice[]> {
    const where = publisherId ? { publisherId } : {};
    return this.floorPriceRepository.find({
      where,
      order: { priority: 'DESC', createdAt: 'DESC' },
    });
  }

  async findOneFloorPrice(id: string): Promise<FloorPrice> {
    const floorPrice = await this.floorPriceRepository.findOne({ where: { id } });
    if (!floorPrice) {
      throw new NotFoundException(`Floor price rule with ID ${id} not found`);
    }
    return floorPrice;
  }

  async createFloorPrice(dto: CreateFloorPriceDto): Promise<FloorPrice> {
    const floorPrice = this.floorPriceRepository.create({
      ...dto,
      isActive: true,
    });
    return this.floorPriceRepository.save(floorPrice);
  }

  async updateFloorPrice(id: string, dto: Partial<CreateFloorPriceDto>): Promise<FloorPrice> {
    const floorPrice = await this.findOneFloorPrice(id);
    Object.assign(floorPrice, dto);
    return this.floorPriceRepository.save(floorPrice);
  }

  async deleteFloorPrice(id: string): Promise<void> {
    const floorPrice = await this.findOneFloorPrice(id);
    await this.floorPriceRepository.remove(floorPrice);
  }

  async toggleFloorPrice(id: string): Promise<FloorPrice> {
    const floorPrice = await this.findOneFloorPrice(id);
    floorPrice.isActive = !floorPrice.isActive;
    return this.floorPriceRepository.save(floorPrice);
  }

  // =================== Brand Safety Rules ===================

  async findAllBrandSafetyRules(publisherId?: string): Promise<BrandSafetyRule[]> {
    const where = publisherId ? { publisherId } : {};
    return this.brandSafetyRepository.find({
      where,
      order: { priority: 'DESC', createdAt: 'DESC' },
    });
  }

  async findOneBrandSafetyRule(id: string): Promise<BrandSafetyRule> {
    const rule = await this.brandSafetyRepository.findOne({ where: { id } });
    if (!rule) {
      throw new NotFoundException(`Brand safety rule with ID ${id} not found`);
    }
    return rule;
  }

  async createBrandSafetyRule(dto: CreateBrandSafetyRuleDto): Promise<BrandSafetyRule> {
    const rule = this.brandSafetyRepository.create({
      ...dto,
      isActive: true,
    });
    return this.brandSafetyRepository.save(rule);
  }

  async updateBrandSafetyRule(id: string, dto: Partial<CreateBrandSafetyRuleDto>): Promise<BrandSafetyRule> {
    const rule = await this.findOneBrandSafetyRule(id);
    Object.assign(rule, dto);
    return this.brandSafetyRepository.save(rule);
  }

  async deleteBrandSafetyRule(id: string): Promise<void> {
    const rule = await this.findOneBrandSafetyRule(id);
    await this.brandSafetyRepository.remove(rule);
  }

  async toggleBrandSafetyRule(id: string): Promise<BrandSafetyRule> {
    const rule = await this.findOneBrandSafetyRule(id);
    rule.isActive = !rule.isActive;
    return this.brandSafetyRepository.save(rule);
  }

  // =================== Inventory Statistics ===================

  async getInventoryStats(publisherId: string) {
    const [adUnits, placements, floorPrices, brandSafetyRules] = await Promise.all([
      this.findAllAdUnits(publisherId),
      this.findAllPlacements(publisherId),
      this.findAllFloorPrices(publisherId),
      this.findAllBrandSafetyRules(publisherId),
    ]);

    return {
      adUnits: {
        total: adUnits.length,
        active: adUnits.filter(u => u.status === AdUnitStatus.ACTIVE).length,
        paused: adUnits.filter(u => u.status === AdUnitStatus.PAUSED).length,
        pending: adUnits.filter(u => u.status === AdUnitStatus.PENDING).length,
        byType: this.groupByProperty(adUnits, 'type'),
      },
      placements: {
        total: placements.length,
        active: placements.filter(p => p.status === PlacementStatus.ACTIVE).length,
        byPosition: this.groupByProperty(placements, 'position'),
      },
      floorPrices: {
        total: floorPrices.length,
        active: floorPrices.filter(f => f.isActive).length,
        byType: this.groupByProperty(floorPrices, 'ruleType'),
      },
      brandSafety: {
        total: brandSafetyRules.length,
        active: brandSafetyRules.filter(r => r.isActive).length,
        byType: this.groupByProperty(brandSafetyRules, 'ruleType'),
      },
    };
  }

  private groupByProperty<T>(items: T[], property: keyof T): Record<string, number> {
    return items.reduce((acc, item) => {
      const key = String(item[property] || 'unknown');
      acc[key] = (acc[key] || 0) + 1;
      return acc;
    }, {} as Record<string, number>);
  }
}
