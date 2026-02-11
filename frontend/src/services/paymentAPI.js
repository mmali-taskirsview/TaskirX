// Frontend Payment API Service Integration
// File: src/services/paymentAPI.js

import axios from 'axios';

const API_BASE_URL = '/api';

// Create axios instance for payments
const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json'
  }
});

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

/**
 * Payment API - Multi-Gateway Support
 * Supports: Stripe, PayPal, Direct Card Processing
 */

// Subscription Management
export const subscriptionAPI = {
  /**
   * Create subscription with any payment method
   * @param {Object} payload - { email, tier, paymentMethod }
   * @returns {Promise} subscription details
   */
  create: (payload) => 
    api.post('/payments/subscriptions', payload),
  
  /**
   * Get subscription details
   * @param {string} subscriptionId
   * @param {string} method - 'stripe' | 'paypal' | 'card'
   */
  get: (subscriptionId, method = 'stripe') =>
    api.get(`/payments/subscriptions/${subscriptionId}`, {
      params: { method }
    }),
  
  /**
   * Cancel subscription
   * @param {string} subscriptionId
   * @param {Object} payload - { method, reason }
   */
  cancel: (subscriptionId, payload) =>
    api.post(`/payments/subscriptions/${subscriptionId}/cancel`, payload)
};

// Payment Processing
export const paymentAPI = {
  /**
   * Get available payment methods
   */
  getAvailableMethods: () =>
    api.get('/payments/methods'),
  
  /**
   * Process one-time payment
   * @param {Object} payload - { email, amount, currency, paymentMethod, description }
   */
  processPayment: (payload) =>
    api.post('/payments/process', payload),
  
  /**
   * Get saved payment methods for user
   */
  getSavedMethods: () =>
    api.get('/payments/saved-methods'),
  
  /**
   * Get payment history
   * @param {Object} params - { limit }
   */
  getHistory: (params = {}) =>
    api.get('/payments/history', { params }),
  
  /**
   * Get billing portal links
   */
  getPortals: () =>
    api.get('/payments/portals'),
  
  /**
   * Process refund
   * @param {Object} payload - { transactionId, method, amount, reason }
   */
  refund: (payload) =>
    api.post('/payments/refunds', payload)
};

// Stripe-Specific Methods
export const stripeAPI = {
  /**
   * Create Stripe payment intent
   * @param {Object} payload - { amount, currency, description }
   */
  createIntent: (payload) =>
    api.post('/payments/stripe/intent', payload),
  
  /**
   * Confirm card payment
   * @param {Object} payload - { paymentIntentId, paymentMethodId }
   */
  confirmPayment: (payload) =>
    api.post('/payments/card/confirm', payload)
};

// PayPal-Specific Methods
export const paypalAPI = {
  /**
   * Create PayPal payment
   * @param {Object} payload - { amount, currency, email, description }
   */
  createPayment: (payload) =>
    api.post('/payments/paypal/create', payload),
  
  /**
   * Create PayPal subscription
   * @param {Object} payload - { email, tier }
   */
  createSubscription: (payload) =>
    api.post('/payments/paypal/subscription', payload),
  
  /**
   * Capture PayPal payment
   * @param {Object} payload - { paymentId, payerId }
   */
  capturePayment: (payload) =>
    api.post('/payments/paypal/capture', payload)
};

// Test Cards for Development
export const TEST_CARDS = {
  STRIPE: {
    VISA: { number: '4242424242424242', name: 'Visa' },
    MASTERCARD: { number: '5555555555554444', name: 'Mastercard' },
    AMEX: { number: '378282246310005', name: 'American Express' },
    THREE_D_SECURE: { number: '4000002500003155', name: 'Visa (3D Secure)' },
    DECLINE: { number: '4000000000000002', name: 'Decline' }
  },
  CVV: '123',
  FUTURE_DATE: '12/25'
};

// Tier Configuration
export const SUBSCRIPTION_TIERS = {
  STARTER: {
    id: 'starter',
    name: 'Starter',
    price: 9.99,
    priceInCents: 999,
    features: ['Basic analytics', 'Up to 10 campaigns', 'Community support']
  },
  PROFESSIONAL: {
    id: 'professional',
    name: 'Professional',
    price: 49.99,
    priceInCents: 4999,
    features: ['Advanced analytics', 'Unlimited campaigns', 'Email support', 'API access']
  },
  ENTERPRISE: {
    id: 'enterprise',
    name: 'Enterprise',
    price: 249.99,
    priceInCents: 24999,
    features: ['Custom analytics', 'Unlimited everything', 'Priority support', 'Dedicated account manager']
  }
};

export default api;
