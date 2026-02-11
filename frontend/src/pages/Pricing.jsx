// Frontend Pricing Component
// File: src/pages/Pricing.jsx

import React, { useState } from 'react';
import { Check, Zap, Crown, Rocket } from 'lucide-react';
import PaymentModal from '../components/payment/PaymentModal';
import { SUBSCRIPTION_TIERS } from '../services/paymentAPI';

const Pricing = () => {
  const [selectedTier, setSelectedTier] = useState(null);
  const [isPaymentOpen, setIsPaymentOpen] = useState(false);

  const tiers = [
    {
      key: 'STARTER',
      icon: Zap,
      highlighted: false,
      color: 'blue'
    },
    {
      key: 'PROFESSIONAL',
      icon: Rocket,
      highlighted: true,
      color: 'purple'
    },
    {
      key: 'ENTERPRISE',
      icon: Crown,
      highlighted: false,
      color: 'gold'
    }
  ];

  const handleSubscribe = (tier) => {
    setSelectedTier(tier);
    setIsPaymentOpen(true);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 py-12 px-4">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-4xl md:text-5xl font-bold text-gray-900 mb-4">
            Simple, Transparent Pricing
          </h1>
          <p className="text-xl text-gray-600">
            Choose the perfect plan for your advertising needs
          </p>
        </div>

        {/* Pricing Cards */}
        <div className="grid md:grid-cols-3 gap-8 mb-12">
          {tiers.map(({ key, icon: Icon, highlighted, color }) => {
            const tier = SUBSCRIPTION_TIERS[key];
            return (
              <div
                key={key}
                className={`relative rounded-2xl transition-all duration-300 ${
                  highlighted
                    ? 'transform scale-105 shadow-2xl bg-white border-2 border-purple-500'
                    : 'bg-white shadow-lg border border-gray-200 hover:shadow-xl hover:border-gray-300'
                }`}
              >
                {/* Highlighted Badge */}
                {highlighted && (
                  <div className="absolute top-0 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
                    <span className="bg-purple-500 text-white px-4 py-1 rounded-full text-sm font-semibold">
                      Most Popular ⭐
                    </span>
                  </div>
                )}

                <div className="p-8">
                  {/* Icon & Name */}
                  <div className="flex items-center gap-3 mb-4">
                    <div className={`p-3 rounded-lg ${
                      color === 'purple' ? 'bg-purple-100' :
                      color === 'gold' ? 'bg-yellow-100' :
                      'bg-blue-100'
                    }`}>
                      <Icon className={
                        color === 'purple' ? 'text-purple-600' :
                        color === 'gold' ? 'text-yellow-600' :
                        'text-blue-600'
                      } size={24} />
                    </div>
                    <h2 className="text-2xl font-bold text-gray-900">{tier.name}</h2>
                  </div>

                  {/* Price */}
                  <div className="mb-6">
                    <div className="flex items-baseline mb-2">
                      <span className="text-5xl font-bold text-gray-900">
                        ${tier.price}
                      </span>
                      <span className="text-gray-600 ml-2">/month</span>
                    </div>
                    <p className="text-gray-600 text-sm">
                      Cancel anytime. No setup fees.
                    </p>
                  </div>

                  {/* CTA Button */}
                  <button
                    onClick={() => handleSubscribe(key)}
                    className={`w-full py-3 rounded-lg font-semibold transition-all duration-200 mb-6 ${
                      highlighted
                        ? 'bg-purple-600 hover:bg-purple-700 text-white'
                        : 'bg-gray-100 hover:bg-gray-200 text-gray-900'
                    }`}
                  >
                    Get Started
                  </button>

                  {/* Features */}
                  <div className="space-y-4">
                    {tier.features.map((feature, idx) => (
                      <div key={idx} className="flex items-start gap-3">
                        <Check className="text-green-500 flex-shrink-0 mt-1" size={20} />
                        <span className="text-gray-700">{feature}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* FAQ Section */}
        <div className="max-w-3xl mx-auto bg-white rounded-2xl shadow-lg p-8 mb-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-8">Frequently Asked Questions</h2>
          
          <div className="space-y-6">
            <div>
              <h3 className="font-semibold text-gray-900 mb-2">Can I change my plan anytime?</h3>
              <p className="text-gray-600">
                Yes! You can upgrade or downgrade your plan at any time. Changes take effect immediately.
              </p>
            </div>

            <div>
              <h3 className="font-semibold text-gray-900 mb-2">What payment methods do you accept?</h3>
              <p className="text-gray-600">
                We accept Stripe (Visa, Mastercard, Amex, ACH), PayPal, and direct card payments. All payments are secure and encrypted.
              </p>
            </div>

            <div>
              <h3 className="font-semibold text-gray-900 mb-2">Do you offer a trial?</h3>
              <p className="text-gray-600">
                Contact our sales team for a custom trial period on Enterprise plans.
              </p>
            </div>

            <div>
              <h3 className="font-semibold text-gray-900 mb-2">Is billing monthly or yearly?</h3>
              <p className="text-gray-600">
                Billing is monthly. Annual billing with discounts is available upon request.
              </p>
            </div>

            <div>
              <h3 className="font-semibold text-gray-900 mb-2">What if I need to cancel?</h3>
              <p className="text-gray-600">
                You can cancel your subscription at any time. There are no cancellation fees or long-term contracts.
              </p>
            </div>
          </div>
        </div>

        {/* Payment Methods Section */}
        <div className="bg-white rounded-2xl shadow-lg p-8">
          <h2 className="text-2xl font-bold text-gray-900 mb-8 text-center">Secure Payment Options</h2>
          
          <div className="grid md:grid-cols-3 gap-8 text-center">
            {/* Stripe */}
            <div className="p-6 rounded-lg bg-gray-50 border border-gray-200">
              <div className="text-4xl mb-3">💳</div>
              <h3 className="font-semibold text-gray-900 mb-2">Stripe</h3>
              <p className="text-gray-600 text-sm">
                Visa, Mastercard, American Express, ACH transfers, and wire payments
              </p>
            </div>

            {/* PayPal */}
            <div className="p-6 rounded-lg bg-gray-50 border border-gray-200">
              <div className="text-4xl mb-3">🅿️</div>
              <h3 className="font-semibold text-gray-900 mb-2">PayPal</h3>
              <p className="text-gray-600 text-sm">
                PayPal account or guest checkout for easy recurring billing
              </p>
            </div>

            {/* Direct Card */}
            <div className="p-6 rounded-lg bg-gray-50 border border-gray-200">
              <div className="text-4xl mb-3">💰</div>
              <h3 className="font-semibold text-gray-900 mb-2">Direct Card</h3>
              <p className="text-gray-600 text-sm">
                Process payments directly with 3D Secure support
              </p>
            </div>
          </div>

          {/* Security Footer */}
          <div className="mt-8 pt-8 border-t border-gray-200 text-center">
            <p className="text-gray-600 text-sm">
              🔒 All payments are PCI-DSS compliant and encrypted with industry-standard security
            </p>
          </div>
        </div>
      </div>

      {/* Payment Modal */}
      <PaymentModal
        isOpen={isPaymentOpen}
        onClose={() => setIsPaymentOpen(false)}
        tier={selectedTier}
        onSuccess={(subscription) => {
          console.log('Subscription created:', subscription);
          // Redirect to dashboard or confirmation page
          window.location.href = '/dashboard';
        }}
      />
    </div>
  );
};

export default Pricing;
