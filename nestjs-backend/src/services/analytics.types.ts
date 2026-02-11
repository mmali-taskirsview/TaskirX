/**
 * Analytics Service Type Definitions
 * Exported types for use across the application
 */

export interface AnalyticsEvent {
  eventId: string;
  campaignId: string;
  userId: string;
  eventType: 'impression' | 'click' | 'conversion' | 'error';
  metadata: Record<string, any>;
  timestamp: Date;
  value?: number;
}

export interface AggregatedMetrics {
  impressions: number;
  clicks: number;
  conversions: number;
  errors: number;
  ctr: number;
  cr: number;
  averageValue: number;
}

export interface ConversionFunnel {
  step: string;
  count: number;
  dropoff: number;
  conversionRate: number;
}
