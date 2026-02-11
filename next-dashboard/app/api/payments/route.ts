import { NextRequest, NextResponse } from 'next/server';
import { PHASE_PRODUCTION_BUILD } from 'next/constants';

export const dynamic = 'force-dynamic';

// Stripe Payment API Integration
// This creates a payment intent for processing invoice payments

const STRIPE_SECRET_KEY = process.env.STRIPE_SECRET_KEY || 'sk_test_demo_key';
const isBuild = process.env.NEXT_PHASE === PHASE_PRODUCTION_BUILD;

export async function POST(request: NextRequest) {
  if (isBuild) {
    return NextResponse.json({
      success: true,
      paymentIntentId: `pi_build_${Date.now()}`,
      status: 'succeeded',
      message: 'Build-time placeholder response',
      demo: true
    });
  }
  try {
    const body = await request.json();
    const { amount, invoiceId, description } = body;

    if (!amount || !invoiceId) {
      return NextResponse.json(
        { error: 'Amount and invoiceId are required' },
        { status: 400 }
      );
    }

    // If we have a real Stripe key, use the Stripe API
    if (STRIPE_SECRET_KEY && STRIPE_SECRET_KEY !== 'sk_test_demo_key') {
      try {
        const stripeResponse = await fetch('https://api.stripe.com/v1/payment_intents', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${STRIPE_SECRET_KEY}`,
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          body: new URLSearchParams({
            amount: String(Math.round(amount * 100)), // Convert to cents
            currency: 'usd',
            description: description || `Invoice payment - ${invoiceId}`,
            'automatic_payment_methods[enabled]': 'true',
          }).toString(),
        });

        if (!stripeResponse.ok) {
          const errorData = await stripeResponse.json();
          console.error('Stripe API error:', errorData);
          return NextResponse.json(
            { error: 'Payment processing failed', details: errorData },
            { status: 500 }
          );
        }

        const paymentIntent = await stripeResponse.json();
        
        return NextResponse.json({
          success: true,
          paymentIntentId: paymentIntent.id,
          clientSecret: paymentIntent.client_secret,
          status: paymentIntent.status,
          message: 'Payment intent created successfully'
        });
      } catch (stripeError) {
        console.error('Stripe error:', stripeError);
        return NextResponse.json(
          { error: 'Failed to connect to Stripe' },
          { status: 500 }
        );
      }
    }

    // Demo mode - simulate successful payment
    // In production, replace with real Stripe integration
    console.log(`Processing demo payment for invoice ${invoiceId}: $${amount}`);
    
    // Simulate processing delay
    await new Promise(resolve => setTimeout(resolve, 1000));

    return NextResponse.json({
      success: true,
      paymentIntentId: `pi_demo_${Date.now()}`,
      status: 'succeeded',
      message: 'Demo payment processed successfully',
      demo: true,
      invoice: {
        id: invoiceId,
        amount: amount,
        paidAt: new Date().toISOString()
      }
    });

  } catch (error) {
    console.error('Payment error:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}

// Get payment status
export async function GET(request: NextRequest) {
  if (isBuild) {
    return NextResponse.json({
      success: true,
      status: 'succeeded',
      demo: true
    });
  }
  const { searchParams } = new URL(request.url);
  const paymentIntentId = searchParams.get('paymentIntentId');

  if (!paymentIntentId) {
    return NextResponse.json(
      { error: 'paymentIntentId is required' },
      { status: 400 }
    );
  }

  // If we have a real Stripe key, check payment status
  if (STRIPE_SECRET_KEY && STRIPE_SECRET_KEY !== 'sk_test_demo_key') {
    try {
      const stripeResponse = await fetch(
        `https://api.stripe.com/v1/payment_intents/${paymentIntentId}`,
        {
          headers: {
            'Authorization': `Bearer ${STRIPE_SECRET_KEY}`,
          },
        }
      );

      if (!stripeResponse.ok) {
        return NextResponse.json(
          { error: 'Failed to retrieve payment status' },
          { status: 500 }
        );
      }

      const paymentIntent = await stripeResponse.json();
      
      return NextResponse.json({
        success: true,
        status: paymentIntent.status,
        amount: paymentIntent.amount / 100, // Convert from cents
      });
    } catch (error) {
      console.error('Stripe status check error:', error);
      return NextResponse.json(
        { error: 'Failed to check payment status' },
        { status: 500 }
      );
    }
  }

  // Demo mode
  return NextResponse.json({
    success: true,
    status: 'succeeded',
    demo: true
  });
}
