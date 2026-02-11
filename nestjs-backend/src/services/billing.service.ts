import { Injectable, Logger, BadRequestException } from '@nestjs/common';

/**
 * Billing & Subscription Service
 * Manages subscription tiers, invoicing, and usage tracking
 */

export enum SubscriptionTier {
  STARTER = 'starter',
  PROFESSIONAL = 'professional',
  ENTERPRISE = 'enterprise',
}

export interface SubscriptionPlan {
  tier: SubscriptionTier;
  monthlyPrice: number;
  features: string[];
  limits: {
    campaigns: number;
    monthlyBudget: number;
    apiCalls: number;
    customMetrics: number;
    webhooks: number;
    integrations: number;
  };
}

export interface Subscription {
  id: string;
  tenantId: string;
  tier: SubscriptionTier;
  status: 'active' | 'paused' | 'cancelled' | 'past_due';
  startDate: Date;
  renewalDate: Date;
  cancelledDate?: Date;
  monthlyPrice: number;
  paymentMethodId: string;
}

export interface Invoice {
  id: string;
  tenantId: string;
  subscriptionId: string;
  amount: number;
  currency: string;
  status: 'draft' | 'sent' | 'paid' | 'failed' | 'refunded';
  dueDate: Date;
  paidDate?: Date;
  items: InvoiceItem[];
  createdAt: Date;
}

export interface InvoiceItem {
  description: string;
  quantity: number;
  unitPrice: number;
  taxRate: number;
  total: number;
}

export interface UsageMetrics {
  tenantId: string;
  period: Date;
  campaignsCreated: number;
  campaignsActive: number;
  totalSpent: number;
  apiCallsUsed: number;
  customMetricsUsed: number;
  webhooksUsed: number;
  integrationsUsed: number;
}

@Injectable()
export class BillingService {
  private readonly logger = new Logger(BillingService.name);

  private subscriptionPlans: Map<SubscriptionTier, SubscriptionPlan> = new Map([
    [
      SubscriptionTier.STARTER,
      {
        tier: SubscriptionTier.STARTER,
        monthlyPrice: 99,
        features: ['Basic campaigns', 'Analytics', 'Email support'],
        limits: {
          campaigns: 10,
          monthlyBudget: 1000,
          apiCalls: 10000,
          customMetrics: 5,
          webhooks: 0,
          integrations: 0,
        },
      },
    ],
    [
      SubscriptionTier.PROFESSIONAL,
      {
        tier: SubscriptionTier.PROFESSIONAL,
        monthlyPrice: 499,
        features: [
          'Unlimited campaigns',
          'Advanced analytics',
          'Webhooks',
          'API access',
          'Priority support',
        ],
        limits: {
          campaigns: 100,
          monthlyBudget: 50000,
          apiCalls: 1000000,
          customMetrics: 50,
          webhooks: 10,
          integrations: 5,
        },
      },
    ],
    [
      SubscriptionTier.ENTERPRISE,
      {
        tier: SubscriptionTier.ENTERPRISE,
        monthlyPrice: 2499,
        features: [
          'Unlimited campaigns',
          'Premium analytics',
          'Webhooks',
          'Full API access',
          'Custom integrations',
          'Dedicated support',
          'SLA guarantee',
          'White-label options',
        ],
        limits: {
          campaigns: 1000,
          monthlyBudget: 500000,
          apiCalls: 10000000,
          customMetrics: 500,
          webhooks: 100,
          integrations: 50,
        },
      },
    ],
  ]);

  private subscriptions: Map<string, Subscription> = new Map();
  private invoices: Map<string, Invoice> = new Map();
  private usageMetrics: Map<string, UsageMetrics> = new Map();

  constructor() {
    this.initializeService();
  }

  private initializeService(): void {
    this.logger.log('Billing service initialized');
  }

  /**
   * Create a new subscription for a tenant
   */
  async createSubscription(
    tenantId: string,
    tier: SubscriptionTier,
    paymentMethodId: string,
  ): Promise<Subscription> {
    try {
      const plan = this.subscriptionPlans.get(tier);
      if (!plan) {
        throw new BadRequestException(`Invalid subscription tier: ${tier}`);
      }

      const subscription: Subscription = {
        id: `sub_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        tenantId,
        tier,
        status: 'active',
        startDate: new Date(),
        renewalDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000), // 30 days
        monthlyPrice: plan.monthlyPrice,
        paymentMethodId,
      };

      this.subscriptions.set(subscription.id, subscription);

      // Create initial invoice
      await this.createInvoice(subscription);

      this.logger.log(`Subscription created: ${subscription.id} for tenant ${tenantId}`);

      return subscription;
    } catch (error) {
      this.logger.error(`Error creating subscription: ${error.message}`);
      throw error;
    }
  }

  /**
   * Upgrade or downgrade subscription tier
   */
  async upgradeSubscription(
    subscriptionId: string,
    newTier: SubscriptionTier,
  ): Promise<Subscription> {
    try {
      const subscription = this.subscriptions.get(subscriptionId);
      if (!subscription) {
        throw new BadRequestException(`Subscription not found: ${subscriptionId}`);
      }

      const oldPlan = this.subscriptionPlans.get(subscription.tier);
      const newPlan = this.subscriptionPlans.get(newTier);

      if (!oldPlan || !newPlan) {
        throw new BadRequestException('Invalid subscription tier');
      }

      // Calculate pro-rata credit
      const daysRemaining = Math.ceil(
        (subscription.renewalDate.getTime() - Date.now()) / (24 * 60 * 60 * 1000),
      );
      const dailyRateOld = oldPlan.monthlyPrice / 30;
      const creditAmount = dailyRateOld * daysRemaining;

      // Update subscription
      subscription.tier = newTier;
      subscription.monthlyPrice = newPlan.monthlyPrice;
      subscription.renewalDate = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000);

      this.subscriptions.set(subscriptionId, subscription);

      // Create upgrade invoice with credit
      await this.createUpgradeInvoice(subscription, creditAmount);

      this.logger.log(`Subscription upgraded: ${subscriptionId} from ${subscription.tier} to ${newTier}`);

      return subscription;
    } catch (error) {
      this.logger.error(`Error upgrading subscription: ${error.message}`);
      throw error;
    }
  }

  /**
   * Cancel subscription
   */
  async cancelSubscription(subscriptionId: string): Promise<Subscription> {
    try {
      const subscription = this.subscriptions.get(subscriptionId);
      if (!subscription) {
        throw new BadRequestException(`Subscription not found: ${subscriptionId}`);
      }

      subscription.status = 'cancelled';
      subscription.cancelledDate = new Date();

      this.subscriptions.set(subscriptionId, subscription);

      this.logger.log(`Subscription cancelled: ${subscriptionId}`);

      return subscription;
    } catch (error) {
      this.logger.error(`Error cancelling subscription: ${error.message}`);
      throw error;
    }
  }

  /**
   * Get subscription details
   */
  getSubscription(subscriptionId: string): Subscription {
    const subscription = this.subscriptions.get(subscriptionId);
    if (!subscription) {
      throw new BadRequestException(`Subscription not found: ${subscriptionId}`);
    }
    return subscription;
  }

  /**
   * Get subscription plan details
   */
  getSubscriptionPlan(tier: SubscriptionTier): SubscriptionPlan {
    const plan = this.subscriptionPlans.get(tier);
    if (!plan) {
      throw new BadRequestException(`Invalid subscription tier: ${tier}`);
    }
    return plan;
  }

  /**
   * Track usage for a tenant
   */
  async trackUsage(
    tenantId: string,
    metrics: Partial<UsageMetrics>,
  ): Promise<UsageMetrics> {
    try {
      const period = new Date();
      period.setHours(0, 0, 0, 0);

      const key = `${tenantId}:${period.toISOString()}`;
      const existing = this.usageMetrics.get(key) || {
        tenantId,
        period,
        campaignsCreated: 0,
        campaignsActive: 0,
        totalSpent: 0,
        apiCallsUsed: 0,
        customMetricsUsed: 0,
        webhooksUsed: 0,
        integrationsUsed: 0,
      };

      const updated: UsageMetrics = {
        ...existing,
        ...metrics,
        tenantId,
        period,
      };

      this.usageMetrics.set(key, updated);

      return updated;
    } catch (error) {
      this.logger.error(`Error tracking usage: ${error.message}`);
      throw error;
    }
  }

  /**
   * Get usage metrics for a tenant and period
   */
  getUsageMetrics(tenantId: string, period?: Date): UsageMetrics {
    const date = period || new Date();
    date.setHours(0, 0, 0, 0);

    const key = `${tenantId}:${date.toISOString()}`;
    return this.usageMetrics.get(key) || {
      tenantId,
      period: date,
      campaignsCreated: 0,
      campaignsActive: 0,
      totalSpent: 0,
      apiCallsUsed: 0,
      customMetricsUsed: 0,
      webhooksUsed: 0,
      integrationsUsed: 0,
    };
  }

  /**
   * Check if tenant has exceeded usage limits
   */
  checkUsageLimits(tenantId: string): {
    withinLimits: boolean;
    violations: string[];
  } {
    try {
      const subscription = Array.from(this.subscriptions.values()).find(
        s => s.tenantId === tenantId && s.status === 'active',
      );

      if (!subscription) {
        return { withinLimits: false, violations: ['No active subscription'] };
      }

      const plan = this.subscriptionPlans.get(subscription.tier);
      const usage = this.getUsageMetrics(tenantId);

      const violations: string[] = [];

      if (usage.campaignsActive > plan.limits.campaigns) {
        violations.push(`Campaign limit exceeded: ${usage.campaignsActive}/${plan.limits.campaigns}`);
      }

      if (usage.apiCallsUsed > plan.limits.apiCalls) {
        violations.push(`API call limit exceeded: ${usage.apiCallsUsed}/${plan.limits.apiCalls}`);
      }

      if (usage.customMetricsUsed > plan.limits.customMetrics) {
        violations.push(`Custom metrics limit exceeded: ${usage.customMetricsUsed}/${plan.limits.customMetrics}`);
      }

      if (usage.webhooksUsed > plan.limits.webhooks) {
        violations.push(`Webhooks limit exceeded: ${usage.webhooksUsed}/${plan.limits.webhooks}`);
      }

      return {
        withinLimits: violations.length === 0,
        violations,
      };
    } catch (error) {
      this.logger.error(`Error checking usage limits: ${error.message}`);
      throw error;
    }
  }

  /**
   * Create invoice for subscription
   */
  private async createInvoice(subscription: Subscription): Promise<Invoice> {
    const plan = this.subscriptionPlans.get(subscription.tier);

    const invoice: Invoice = {
      id: `inv_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      tenantId: subscription.tenantId,
      subscriptionId: subscription.id,
      amount: plan.monthlyPrice,
      currency: 'USD',
      status: 'sent',
      dueDate: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000), // 14 days
      items: [
        {
          description: `${subscription.tier} subscription - 1 month`,
          quantity: 1,
          unitPrice: plan.monthlyPrice,
          taxRate: 0.1, // 10% tax
          total: plan.monthlyPrice * 1.1,
        },
      ],
      createdAt: new Date(),
    };

    this.invoices.set(invoice.id, invoice);

    this.logger.log(`Invoice created: ${invoice.id} for subscription ${subscription.id}`);

    return invoice;
  }

  /**
   * Create upgrade invoice with pro-rata credit
   */
  private async createUpgradeInvoice(
    subscription: Subscription,
    creditAmount: number,
  ): Promise<Invoice> {
    const plan = this.subscriptionPlans.get(subscription.tier);
    const amount = plan.monthlyPrice - creditAmount;

    const invoice: Invoice = {
      id: `inv_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      tenantId: subscription.tenantId,
      subscriptionId: subscription.id,
      amount: Math.max(0, amount),
      currency: 'USD',
      status: 'sent',
      dueDate: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000),
      items: [
        {
          description: `Upgrade to ${subscription.tier} - 1 month`,
          quantity: 1,
          unitPrice: plan.monthlyPrice,
          taxRate: 0.1,
          total: plan.monthlyPrice * 1.1,
        },
        {
          description: `Pro-rata credit for previous tier`,
          quantity: 1,
          unitPrice: -creditAmount,
          taxRate: 0,
          total: -creditAmount,
        },
      ],
      createdAt: new Date(),
    };

    this.invoices.set(invoice.id, invoice);

    this.logger.log(`Upgrade invoice created: ${invoice.id}`);

    return invoice;
  }

  /**
   * Get invoices for a tenant
   */
  getInvoices(tenantId: string): Invoice[] {
    return Array.from(this.invoices.values()).filter(inv => inv.tenantId === tenantId);
  }

  /**
   * Mark invoice as paid
   */
  async markInvoicePaid(invoiceId: string): Promise<Invoice> {
    const invoice = this.invoices.get(invoiceId);
    if (!invoice) {
      throw new BadRequestException(`Invoice not found: ${invoiceId}`);
    }

    invoice.status = 'paid';
    invoice.paidDate = new Date();

    this.invoices.set(invoiceId, invoice);

    this.logger.log(`Invoice marked as paid: ${invoiceId}`);

    return invoice;
  }
}
