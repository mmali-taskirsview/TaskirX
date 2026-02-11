import {
  Controller,
  Post,
  Get,
  Put,
  Body,
  Param,
  BadRequestException,
  Req,
} from '@nestjs/common';
import { ApiOperation, ApiTags, ApiBearerAuth } from '@nestjs/swagger';
import { StripeService } from '../services/stripe.service';
import { Request } from 'express';

/**
 * Payment Controller - Stripe integration endpoints
 * 
 * Handles:
 * - Subscription creation and management
 * - Invoice retrieval
 * - Webhook processing
 * - Revenue reporting
 * - Billing history
 */
@ApiTags('Payments')
@Controller('payments')
export class PaymentsController {
  constructor(private stripeService: StripeService) {}

  /**
   * Create a new subscription
   */
  @Post('subscriptions')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Create subscription',
    description: 'Create a new subscription for the authenticated user',
  })
  async createSubscription(
    @Body()
    payload: {
      email: string;
      tier: 'STARTER' | 'PROFESSIONAL' | 'ENTERPRISE';
      paymentMethodId: string;
    },
  ) {
    const userId = 'user_123'; // Would come from auth context

    if (!payload.tier || !payload.paymentMethodId) {
      throw new BadRequestException('Missing required fields');
    }

    const subscription = await this.stripeService.createSubscription(
      userId,
      payload.email,
      payload.tier,
      payload.paymentMethodId,
    );

    return {
      success: true,
      subscription,
      message: `${payload.tier} subscription created successfully`,
    };
  }

  /**
   * Get current subscription
   */
  @Get('subscriptions/current')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Get current subscription',
    description: 'Get the active subscription for the current user',
  })
  async getCurrentSubscription() {
    const userId = 'user_123';

    const subscription = await this.stripeService.getSubscriptionData(userId);

    if (!subscription) {
      throw new BadRequestException('No active subscription found');
    }

    return {
      success: true,
      subscription,
    };
  }

  /**
   * Upgrade subscription tier
   */
  @Put('subscriptions/upgrade')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Upgrade subscription',
    description: 'Upgrade to a higher subscription tier with pro-rata credit',
  })
  async upgradeSubscription(
    @Body() payload: { tier: 'PROFESSIONAL' | 'ENTERPRISE' },
  ) {
    const userId = 'user_123';

    if (!payload.tier) {
      throw new BadRequestException('Tier is required');
    }

    const result = await this.stripeService.upgradeSubscription(
      userId,
      payload.tier,
    );

    return {
      success: true,
      upgrade: result,
      message: `Subscription upgraded to ${payload.tier}`,
    };
  }

  /**
   * Downgrade subscription tier
   */
  @Put('subscriptions/downgrade')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Downgrade subscription',
    description: 'Downgrade to a lower subscription tier',
  })
  async downgradeSubscription(
    @Body() payload: { tier: 'STARTER' | 'PROFESSIONAL' },
  ) {
    const userId = 'user_123';

    if (!payload.tier) {
      throw new BadRequestException('Tier is required');
    }

    const result = await this.stripeService.upgradeSubscription(
      userId,
      payload.tier,
    );

    return {
      success: true,
      downgrade: result,
      message: `Subscription downgraded to ${payload.tier}`,
    };
  }

  /**
   * Cancel subscription
   */
  @Post('subscriptions/cancel')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Cancel subscription',
    description: 'Cancel the active subscription and process refund if applicable',
  })
  async cancelSubscription(
    @Body() payload?: { reason?: string },
  ) {
    const userId = 'user_123';

    const result = await this.stripeService.cancelSubscription(
      userId,
      payload?.reason,
    );

    return {
      success: true,
      cancellation: result,
      message: 'Subscription canceled successfully',
    };
  }

  /**
   * Get invoices
   */
  @Get('invoices')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Get invoices',
    description: 'Get billing history and invoices for the current user',
  })
  async getInvoices() {
    const userId = 'user_123';
    const limit = 50;

    const invoices = await this.stripeService.getCustomerInvoices(
      userId,
      limit,
    );

    return {
      success: true,
      invoices,
      count: invoices.length,
    };
  }

  /**
   * Get invoice details
   */
  @Get('invoices/:invoiceId')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Get invoice details',
    description: 'Get details of a specific invoice',
  })
  async getInvoiceDetails(@Param('invoiceId') invoiceId: string) {
    // Mock implementation
    return {
      success: true,
      invoice: {
        id: invoiceId,
        amount: 49900,
        status: 'paid',
        date: new Date(),
        items: [
          {
            description: 'Professional Plan',
            quantity: 1,
            unitPrice: 49900,
            total: 49900,
          },
        ],
        tax: 0,
        total: 49900,
      },
    };
  }

  /**
   * Download invoice PDF
   */
  @Get('invoices/:invoiceId/pdf')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Download invoice PDF',
    description: 'Download invoice as PDF file',
  })
  async downloadInvoicePDF(@Param('invoiceId') invoiceId: string) {
    // Mock implementation
    return {
      success: true,
      message: 'PDF download would be generated here',
      invoiceId,
    };
  }

  /**
   * Update payment method
   */
  @Put('payment-methods')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Update payment method',
    description: 'Update the default payment method for the subscription',
  })
  async updatePaymentMethod(
    @Body() payload: { paymentMethodId: string },
  ) {
    // Mock implementation
    return {
      success: true,
      message: 'Payment method updated successfully',
      paymentMethodId: payload.paymentMethodId,
    };
  }

  /**
   * Get revenue metrics
   */
  @Get('metrics/revenue')
  @ApiOperation({
    summary: 'Get revenue metrics',
    description: 'Get platform revenue metrics (admin only)',
  })
  async getRevenueMetrics() {
    const metrics = await this.stripeService.getRevenueMetrics('month');

    return {
      success: true,
      metrics,
      period: 'month',
    };
  }

  /**
   * Stripe webhook endpoint
   */
  @Post('webhooks/stripe')
  @ApiOperation({
    summary: 'Stripe webhook',
    description: 'Receive webhook events from Stripe',
  })
  async handleStripeWebhook(@Req() req: Request) {
    try {
      // Get raw body for signature verification
  const _sig = req.headers['stripe-signature'];

      // TODO: Verify webhook signature with Stripe key
      // const event = stripe.webhooks.constructEvent(
      //   rawBody,
      //   sig,
      //   process.env.STRIPE_WEBHOOK_SECRET,
      // );

      // Parse body as JSON
      let event: any;
      if (typeof req.body === 'string') {
        event = JSON.parse(req.body);
      } else {
        event = req.body;
      }

      // Process webhook
      await this.stripeService.processWebhook(event);

      return {
        success: true,
        received: true,
      };
    } catch (error) {
      return {
        success: false,
        error: error.message,
      };
    }
  }

  /**
   * Get billing portal link
   */
  @Post('portal')
  @ApiBearerAuth()
  @ApiOperation({
    summary: 'Get billing portal link',
    description: 'Get link to Stripe billing portal for subscription management',
  })
  async getBillingPortalLink() {
    // Mock implementation
    return {
      success: true,
      portalUrl: 'https://billing.stripe.com/...',
      expiresAt: new Date(Date.now() + 3600000),
    };
  }

  /**
   * Get subscription plans
   */
  @Get('plans')
  @ApiOperation({
    summary: 'Get available plans',
    description: 'Get list of available subscription plans with pricing',
  })
  async getPlans() {
    return {
      success: true,
      plans: [
        {
          id: 'plan_starter',
          name: 'Starter',
          tier: 'STARTER',
          price: 99,
          currency: 'USD',
          interval: 'month',
          features: {
            campaigns: 10,
            apiCalls: 10000,
            monthlyBudget: 1000,
            support: 'email',
          },
        },
        {
          id: 'plan_professional',
          name: 'Professional',
          tier: 'PROFESSIONAL',
          price: 499,
          currency: 'USD',
          interval: 'month',
          features: {
            campaigns: 100,
            apiCalls: 1000000,
            monthlyBudget: 50000,
            support: 'priority',
          },
        },
        {
          id: 'plan_enterprise',
          name: 'Enterprise',
          tier: 'ENTERPRISE',
          price: 2499,
          currency: 'USD',
          interval: 'month',
          features: {
            campaigns: 1000,
            apiCalls: 10000000,
            monthlyBudget: 500000,
            support: '24/7 dedicated',
          },
        },
      ],
    };
  }

  /**
   * Validate coupon code
   */
  @Post('coupons/validate')
  @ApiOperation({
    summary: 'Validate coupon',
    description: 'Validate a coupon code and get discount details',
  })
  async validateCoupon(@Body() payload: { code: string }) {
    // Mock implementation
    if (payload.code === 'LAUNCH20') {
      return {
        success: true,
        valid: true,
        discount: 0.2,
        discountType: 'percentage',
        expiresAt: new Date('2026-06-30'),
      };
    }

    return {
      success: true,
      valid: false,
      message: 'Coupon code not found',
    };
  }
}
