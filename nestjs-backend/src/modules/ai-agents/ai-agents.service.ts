import { Injectable } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { AiCoreService } from './ai-core.service';

@Injectable()
export class AiAgentsService {
  
  constructor(
    private readonly _configService: ConfigService,
    private aiCoreService: AiCoreService
  ) {}

  private stringToHash(str: string): number {
    let hash = 0;
    if (str.length === 0) return hash;
    for (let i = 0; i < str.length; i++) {
        const char = str.charCodeAt(i);
        hash = ((hash << 5) - hash) + char;
        hash = hash & hash;
    }
    return Math.abs(hash) / 2147483647; // Normalize to 0-1
  }

  async detectFraud(data: {
    userId: string;
    ipAddress: string;
    userAgent: string;
    clickPattern: any;
  }): Promise<{ isFraud: boolean; confidence: number; reason?: string }> {
    try {
      // Feature Engineering: Convert raw data into 15 numerical indicators
      const indicators = new Array(15).fill(0);
      indicators[0] = this.stringToHash(data.ipAddress);
      indicators[1] = this.stringToHash(data.userAgent);
      indicators[2] = data.clickPattern?.frequency || 0;
      indicators[3] = data.clickPattern?.duration || 0;
      // ... fill others with derived metrics
      
      const result = await this.aiCoreService.detectFraud(indicators);

      return {
        isFraud: result.isFraud,
        confidence: result.score
      };
    } catch (error) {
      console.error('Fraud detection failed:', error.message);
      return { isFraud: false, confidence: 0 };
    }
  }

  async matchAd(_request: {
    publisherId: string;
    adSlot: string;
    userProfile: any;
    context: any;
  }): Promise<{ campaignIds: string[]; scores: number[] }> {
    // Simplified matching logic (could be another model)
    // For now, we return empty or mock, as the core requirement was specific models
    return { campaignIds: [], scores: [] };
  }

  async optimizeBid(_data: {
    campaignId: string;
    historicalPerformance: any;
    competitorBids: number[];
    targetMetric: 'ctr' | 'cpc' | 'cpa';
  }): Promise<{ recommendedBid: number; confidence: number }> {
    try {
      // 10 Features: [Ctr, Cvr, Spend, Budget, DaysLeft, CompetitorAvg, ...Context]
      const features = new Array(10).fill(0);
      // Fill with dummy/extracted data
      features[0] = Math.random(); 

      const result = await this.aiCoreService.predictOptimalBid(features);
      
      // Scale prediction (0-1) to actual currency
      const baseBid = 1.0; 
      const recommendedBid = result.bidMultiplier * baseBid * 2; // e.g. up to $2.00

      return { 
        recommendedBid, 
        confidence: result.confidence 
      };
    } catch (error) {
      console.error('Bid optimization failed:', error.message);
      return { recommendedBid: 0, confidence: 0 };
    }
  }

  async predictPerformance(_campaignId: string): Promise<{
    expectedCtr: number;
    expectedConversions: number;
    expectedRoi: number;
  }> {
    try {
      // Mock history: 7 days x 5 features
      const history = Array(7).fill(0).map(() => Array(5).fill(Math.random()));
      
      const prediction = await this.aiCoreService.predictNextDayPerformance(history);
      
      return { 
        expectedCtr: prediction[0], 
        expectedConversions: prediction[1] * 100, 
        expectedRoi: prediction[2] * 200 
      };
    } catch (error) {
      console.error('Performance prediction failed:', error.message);
      return { expectedCtr: 0, expectedConversions: 0, expectedRoi: 0 };
    }
  }

  async getAnomalies(_tenantId: string): Promise<any[]> {
    return [];
  }
}
