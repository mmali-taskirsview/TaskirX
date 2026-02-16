import { Injectable, Logger, BadRequestException } from '@nestjs/common';
import { Inject } from '@nestjs/common';
import { Redis } from 'ioredis';

/**
 * Stripe Payment Integration Service
 * 
 * Handles:
 * - Subscription lifecycle management
 * - Invoice generation and tracking
 * - Payment processing and webhooks
 * - Revenue analytics
 * - Dunning management (failed payments)
 * 
 * Subscriptions:
 * - STARTER: $99/month
 * - PROFESSIONAL: $499/month
 * - ENTERPRISE: $2,499/month
 * 
 * Note: Integration requires Stripe SDK (npm install stripe)
 */
@Injectable()
export class StripeService {
  private readonly logger = new Logger(StripeService.name);
  private stripe: any; // Stripe client instance

  constructor(
    @Inject('REDIS_CLIENT')
    private redisClient: Redis,
  ) {
    // Initialize Stripe (requires API key from env)
    // const Stripe = require('stripe');
    // this.stripe = new Stripe(process.env.STRIPE_SECRET_KEY);
  }

  /**
   * Create a subscription for a customer
   */
  async createSubscription(
    userId: string,
    email: string,
    tier: 'STARTER' | 'PROFESSIONAL' | 'ENTERPRISE',
    _paymentMethodId: string,
  ): Promise<{
    subscriptionId: string;
    customerId: string;
    status: string;
    amount: number;
    nextBillingDate: Date;
  }> {
    try {
      // Mock Stripe customer creation
      // const customer = await this.stripe.customers.create({
      //   email,
      //   payment_method: paymentMethodId,
      //   invoice_settings: {
      //     default_payment_method: paymentMethodId,
      //   },
      // });

      const planPrices = {
        STARTER: 9900, // $99.00
        PROFESSIONAL: 49900, // $499.00
        ENTERPRISE: 249900, // $2,499.00
      };

      // Mock subscription creation
      const subscriptionData = {
        customerId: `cus_${userId}_${Date.now()}`,
        subscriptionId: `sub_${userId}_${Date.now()}`,
        tier,
        amount: planPrices[tier],
        status: 'active',
        createdAt: new Date(),
        nextBillingDate: this.getNextBillingDate(),
        currentPeriodStart: new Date(),
        currentPeriodEnd: this.getNextBillingDate(),
      };

      // Store in Redis
      await this.redisClient.hset(
        `stripe:subscription:${userId}`,
        'data',
        JSON.stringify(subscriptionData),
      );

      return {
        subscriptionId: subscriptionData.subscriptionId,
        customerId: subscriptionData.customerId,
        status: subscriptionData.status,
        amount: subscriptionData.amount,
        nextBillingDate: subscriptionData.nextBillingDate,
      };
    } catch (error) {
      this.logger.error(`Failed to create subscription: ${error.message}`);
      throw new BadRequestException('Subscription creation failed');
    }
  }

  /**
   * Upgrade subscription tier
   */
  async upgradeSubscription(
    userId: string,
    newTier: 'STARTER' | 'PROFESSIONAL' | 'ENTERPRISE',
  ): Promise<{
    subscriptionId: string;
    newTier: string;
    proRataCredit: number;
    newAmount: number;
    effectiveDate: Date;
  }> {
    try {
      const subscriptionData = await this.getSubscriptionData(userId);

      if (!subscriptionData) {
        throw new BadRequestException('No active subscription found');
      }

      const planPrices = {
        STARTER: 9900,
        PROFESSIONAL: 49900,
        ENTERPRISE: 249900,
      };

      // Calculate pro-rata credit for remaining billing period
      const proRataCredit = this.calculateProRataCredit(
        subscriptionData.amount,
        subscriptionData.currentPeriodEnd,
      );

      const newAmount = planPrices[newTier];
      const effectiveAmount = newAmount - proRataCredit;

      // Mock Stripe subscription update
      const updatedData = {
        ...subscriptionData,
        tier: newTier,
        amount: newAmount,
        proRataCredit,
        upgradedAt: new Date(),
      };

      await this.redisClient.hset(
        `stripe:subscription:${userId}`,
        'data',
        JSON.stringify(updatedData),
      );

      return {
        subscriptionId: subscriptionData.subscriptionId,
        newTier,
        proRataCredit,
        newAmount: effectiveAmount,
        effectiveDate: new Date(),
      };
    } catch (error) {
      this.logger.error(`Failed to upgrade subscription: ${error.message}`);
      throw new BadRequestException('Upgrade failed');
    }
  }

  /**
   * Cancel subscription
   */
  async cancelSubscription(
    userId: string,
    reason?: string,
  ): Promise<{
    subscriptionId: string;
    canceledAt: Date;
    refund?: number;
  }> {
    try {
      const subscriptionData = await this.getSubscriptionData(userId);

      if (!subscriptionData) {
        throw new BadRequestException('No active subscription found');
      }

      // Calculate refund for remaining period
      const refund = this.calculateRefund(
        subscriptionData.amount,
        subscriptionData.currentPeriodEnd,
      );

      // Mock Stripe subscription cancellation
      const canceledData = {
        ...subscriptionData,
        status: 'canceled',
        canceledAt: new Date(),
        cancelReason: reason,
        refund,
      };

      await this.redisClient.hset(
        `stripe:subscription:${userId}`,
        'data',
        JSON.stringify(canceledData),
      );

      // Log for analytics
      await this.redisClient.lpush(
        'churn:log',
        JSON.stringify({
          userId,
          tier: subscriptionData.tier,
          reason,
          refund,
          timestamp: new Date(),
        }),
      );

      return {
        subscriptionId: subscriptionData.subscriptionId,
        canceledAt: new Date(),
        refund,
      };
    } catch (error) {
      this.logger.error(`Failed to cancel subscription: ${error.message}`);
      throw new BadRequestException('Cancellation failed');
    }
  }

  /**
   * Process webhook from Stripe
   */
  async processWebhook(event: any): Promise<void> {
    try {
      switch (event.type) {
        case 'payment_intent.succeeded':
          await this.handlePaymentSucceeded(event.data.object);
          break;
        case 'payment_intent.payment_failed':
          await this.handlePaymentFailed(event.data.object);
          break;
        case 'invoice.payment_succeeded':
          await this.handleInvoicePaid(event.data.object);
          break;
        case 'invoice.payment_failed':
          await this.handleInvoiceFailure(event.data.object);
          break;
        case 'customer.subscription.updated':
          await this.handleSubscriptionUpdated(event.data.object);
          break;
        case 'customer.subscription.deleted':
          await this.handleSubscriptionDeleted(event.data.object);
          break;
        default:
          this.logger.log(`Unhandled webhook type: ${event.type}`);
      }
    } catch (error) {
      this.logger.error(`Webhook processing error: ${error.message}`);
      throw error;
    }
  }

  /**
   * Get subscription data for user
   */
  async getSubscriptionData(userId: string): Promise<any> {
    const data = await this.redisClient.hget(
      `stripe:subscription:${userId}`,
      'data',
    );
    if (!data) {
      return null;
    }

    const parsed = JSON.parse(data);
    parsed.createdAt = parsed.createdAt ? new Date(parsed.createdAt) : parsed.createdAt;
    parsed.nextBillingDate = parsed.nextBillingDate
      ? new Date(parsed.nextBillingDate)
      : parsed.nextBillingDate;
    parsed.currentPeriodStart = parsed.currentPeriodStart
      ? new Date(parsed.currentPeriodStart)
      : parsed.currentPeriodStart;
    parsed.currentPeriodEnd = parsed.currentPeriodEnd
      ? new Date(parsed.currentPeriodEnd)
      : parsed.currentPeriodEnd;
    parsed.upgradedAt = parsed.upgradedAt ? new Date(parsed.upgradedAt) : parsed.upgradedAt;
    parsed.canceledAt = parsed.canceledAt ? new Date(parsed.canceledAt) : parsed.canceledAt;

    return parsed;
  }

  /**
   * Get revenue metrics
   */
  async getRevenueMetrics(_period: 'day' | 'month' | 'year' = 'month'): Promise<{
    totalRevenue: number;
    activeSubscriptions: number;
    mrr: number;
    arr: number;
    churnRate: number;
    ltv: number;
  }> {
    // Mock data - would query Stripe API in production
    return {
      totalRevenue: 125000,
      activeSubscriptions: 150,
      mrr: 49900, // Monthly recurring revenue
      arr: 598800, // Annual recurring revenue
      churnRate: 0.05, // 5% monthly churn
      ltv: 5988, // Lifetime value (12 months)
    };
  }

  /**
   * Get customer invoices
   */
  async getCustomerInvoices(userId: string, _limit: number = 10): Promise<any[]> {
    // Mock implementation
    return [
      {
        invoiceId: `inv_${userId}_001`,
        date: new Date(),
        amount: 49900,
        status: 'paid',
        description: 'Professional Plan',
      },
    ];
  }

  /**
   * Calculate pro-rata credit
   */
  private calculateProRataCredit(
    monthlyAmount: number,
    periodEnd: Date,
  ): number {
    const daysRemaining = Math.ceil(
      (periodEnd.getTime() - Date.now()) / (1000 * 60 * 60 * 24),
    );
    const daysInMonth = 30;
    return Math.floor((monthlyAmount / daysInMonth) * daysRemaining);
  }

  /**
   * Calculate refund for canceled subscription
   */
  private calculateRefund(monthlyAmount: number, periodEnd: Date): number {
    // Refund only if canceled mid-period
    const now = Date.now();
    if (periodEnd.getTime() > now) {
      return this.calculateProRataCredit(monthlyAmount, periodEnd);
    }
    return 0;
  }

  /**
   * Get next billing date (30 days from now)
   */
  private getNextBillingDate(): Date {
    const date = new Date();
    date.setDate(date.getDate() + 30);
    return date;
  }

  /**
   * Handle payment succeeded event
   */
  private async handlePaymentSucceeded(payment: any): Promise<void> {
    this.logger.log(`Payment succeeded: ${payment.id}`);
    await this.redisClient.lpush(
      'payments:succeeded',
      JSON.stringify({
        paymentId: payment.id,
        amount: payment.amount,
        timestamp: new Date(),
      }),
    );
  }

  /**
   * Handle payment failed event
   */
  private async handlePaymentFailed(payment: any): Promise<void> {
    this.logger.warn(`Payment failed: ${payment.id}`);
    await this.redisClient.lpush(
      'payments:failed',
      JSON.stringify({
        paymentId: payment.id,
        error: payment.last_payment_error,
        timestamp: new Date(),
      }),
    );
  }

  /**
   * Handle invoice paid event
   */
  private async handleInvoicePaid(invoice: any): Promise<void> {
    this.logger.log(`Invoice paid: ${invoice.id}`);
  }

  /**
   * Handle invoice failure event
   */
  private async handleInvoiceFailure(invoice: any): Promise<void> {
    this.logger.warn(`Invoice failed: ${invoice.id}`);
  }

  /**
   * Handle subscription updated event
   */
  private async handleSubscriptionUpdated(subscription: any): Promise<void> {
    this.logger.log(`Subscription updated: ${subscription.id}`);
  }

  /**
   * Handle subscription deleted event
   */
  private async handleSubscriptionDeleted(subscription: any): Promise<void> {
    this.logger.log(`Subscription deleted: ${subscription.id}`);
  }
}
