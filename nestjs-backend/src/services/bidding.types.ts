/**
 * Bidding Optimization Service Type Definitions
 * Exported types for use across the application
 */

export interface BiddingContext {
  campaignId: string;
  adSpaceId: string;
  userId: string;
  deviceType: string;
  location: string;
  dayOfWeek: number;
  hourOfDay: number;
  historicalCTR: number;
  historicalCR: number;
  budget: number;
}

export interface BidPrediction {
  recommendedBid: number;
  confidence: number;
  expectedROI: number;
  reasoning: string;
  timestamp: Date;
}

export interface ModelPerformance {
  accuracy: number;
  precision: number;
  recall: number;
  f1Score: number;
  auc: number;
  lastUpdated: Date;
}
