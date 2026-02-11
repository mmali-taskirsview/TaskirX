// Models and Types for TaskirX React Native SDK

export interface ClientConfig {
  apiUrl: string;
  apiKey: string;
  debug?: boolean;
  timeout?: number;
  retryAttempts?: number;
}

export interface AuthResponse {
  token: string;
  refreshToken: string;
  user: User;
  expiresIn: number;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
  company?: string;
}

export interface User {
  id: string;
  email: string;
  name: string;
  company?: string;
  role: string;
  createdAt: string;
  updatedAt: string;
}

export interface Campaign {
  id: string;
  name: string;
  budget: number;
  startDate: string;
  endDate: string;
  status: string;
  targetAudience: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export interface CampaignCreateRequest {
  name: string;
  budget: number;
  startDate: string;
  endDate: string;
  targetAudience: Record<string, any>;
}

export interface Bid {
  id: string;
  campaignId: string;
  adSlotId: string;
  amount: number;
  currency: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface BidSubmitRequest {
  campaignId: string;
  adSlotId: string;
  amount: number;
  currency?: string;
}

export interface Analytics {
  impressions: number;
  clicks: number;
  conversions: number;
  spend: number;
  revenue: number;
  ctr: number;
  conversionRate: number;
  roi: number;
  timestamp: string;
}

export interface Ad {
  id: string;
  campaignId: string;
  placement: string;
  imageUrl: string;
  clickUrl: string;
  dimensions: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface AdCreateRequest {
  campaignId: string;
  placement: string;
  imageUrl: string;
  clickUrl: string;
  dimensions: string;
}

export interface Webhook {
  id: string;
  url: string;
  events: string[];
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface WebhookCreateRequest {
  url: string;
  events: string[];
}

export interface WebhookEvent {
  id: string;
  type: string;
  data: Record<string, any>;
  timestamp: string;
}

export interface ErrorResponse {
  code: string;
  message: string;
  details?: Record<string, any>;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: ErrorResponse;
}

export type Result<T> = 
  | { success: true; data: T }
  | { success: false; error: TaskirXError };

export enum TaskirXErrorType {
  NETWORK_ERROR = 'NETWORK_ERROR',
  DECODING_ERROR = 'DECODING_ERROR',
  HTTP_ERROR = 'HTTP_ERROR',
  INVALID_RESPONSE = 'INVALID_RESPONSE',
  TIMEOUT = 'TIMEOUT',
  RETRY_EXHAUSTED = 'RETRY_EXHAUSTED',
}

export interface TaskirXError {
  type: TaskirXErrorType;
  message: string;
  statusCode?: number;
  originalError?: Error;
}
