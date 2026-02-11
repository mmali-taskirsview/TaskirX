/**
 * Type Definitions for TaskirX SDK
 */

export interface ClientConfig {
  baseUrl?: string;
  apiKey: string;
  apiSecret?: string;
  timeout?: number;
  retries?: number;
  environment?: 'production' | 'staging' | 'development';
  debug?: boolean;
  [key: string]: any;
}

export interface Campaign {
  campaign_id: string;
  name: string;
  budget: number;
  spent: number;
  status: 'draft' | 'active' | 'paused' | 'completed';
  start_date: string;
  end_date: string;
  impressions: number;
  clicks: number;
  conversions: number;
  revenue: number;
  created_at: string;
  updated_at: string;
  [key: string]: any;
}

export interface Bid {
  bid_id: string;
  campaign_id: string;
  ad_unit_id: string;
  amount: number;
  currency: string;
  status: 'active' | 'inactive';
  created_at: string;
  [key: string]: any;
}

export interface Analytics {
  impressions: number;
  clicks: number;
  conversions: number;
  revenue: number;
  ctr: number;
  conversion_rate: number;
  roi: number;
  period?: string;
  timestamp?: string;
  [key: string]: any;
}

export interface Webhook {
  webhook_id: string;
  url: string;
  events: string[];
  active: boolean;
  created_at: string;
  updated_at: string;
  last_fired?: string;
  failure_count?: number;
  [key: string]: any;
}

export interface WebhookEvent {
  event: string;
  data: any;
  timestamp: string;
  [key: string]: any;
}

export interface RequestOptions {
  timeout?: number;
  retries?: number;
  headers?: Record<string, string>;
  [key: string]: any;
}
