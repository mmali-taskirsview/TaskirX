// Frontend Payment UI Component
// File: src/components/payment/PaymentModal.jsx

import React, { useState, useEffect } from 'react';
import { X, CreditCard, DollarSign, CheckCircle, AlertCircle } from 'lucide-react';
import { paymentAPI, stripeAPI, paypalAPI, SUBSCRIPTION_TIERS, TEST_CARDS } from '../../services/paymentAPI';
import toast from 'react-hot-toast';

const PaymentModal = ({ isOpen, onClose, tier = 'PROFESSIONAL', onSuccess }) => {
  const [paymentMethod, setPaymentMethod] = useState('stripe');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [step, setStep] = useState('method'); // 'method', 'payment', 'confirm'
  
  // Stripe state
  const [stripeEmail, setStripeEmail] = useState('');
  
  // PayPal state
  const [paypalEmail, setPaypalEmail] = useState('');

  const tierConfig = SUBSCRIPTION_TIERS[tier] || SUBSCRIPTION_TIERS.PROFESSIONAL;

  // Available payment methods
  const methods = [
    {
      id: 'stripe',
      name: 'Stripe',
      description: 'Visa, Mastercard, Amex, ACH',
      icon: '💳'
    },
    {
      id: 'paypal',
      name: 'PayPal',
      description: 'PayPal account or guest checkout',
      icon: '🅿️'
    },
    {
      id: 'card',
      name: 'Direct Card',
      description: 'Visa, Mastercard, American Express',
      icon: '💰'
    }
  ];

  // Handle subscription creation
  const handleCreateSubscription = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const email = paymentMethod === 'paypal' ? paypalEmail : stripeEmail;
      
      if (!email) {
        setError('Please enter an email address');
        setLoading(false);
        return;
      }

      const response = await subscriptionAPI.create({
        email,
        tier,
        paymentMethod
      });

      if (paymentMethod === 'paypal') {
        // Redirect to PayPal approval URL
        if (response.data.subscription?.approvalUrl) {
          window.location.href = response.data.subscription.approvalUrl;
        }
      } else if (paymentMethod === 'stripe' || paymentMethod === 'card') {
        // Show success and close modal
        toast.success('Subscription created successfully!');
        onSuccess?.(response.data.subscription);
        onClose();
      }
    } catch (err) {
      const errorMessage = err.response?.data?.error || err.message || 'Payment failed';
      setError(errorMessage);
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        {/* Header */}
        <div className="flex justify-between items-center p-6 border-b border-gray-200">
          <h2 className="text-2xl font-bold text-gray-900">
            {tierConfig.name} Plan
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition"
          >
            <X size={24} />
          </button>
        </div>

        {/* Body */}
        <div className="p-6">
          {/* Price Display */}
          <div className="mb-6 bg-gray-50 p-4 rounded-lg">
            <div className="flex items-baseline justify-center">
              <span className="text-4xl font-bold text-gray-900">
                ${tierConfig.price}
              </span>
              <span className="text-gray-600 ml-2">/month</span>
            </div>
          </div>

          {/* Error Alert */}
          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg flex gap-2">
              <AlertCircle size={20} className="text-red-600 flex-shrink-0" />
              <p className="text-red-600 text-sm">{error}</p>
            </div>
          )}

          {/* Payment Method Selection */}
          {step === 'method' && (
            <div className="space-y-3">
              <label className="block text-sm font-medium text-gray-700 mb-4">
                Select Payment Method
              </label>
              
              {methods.map((method) => (
                <label
                  key={method.id}
                  className={`flex items-center p-4 border-2 rounded-lg cursor-pointer transition ${
                    paymentMethod === method.id
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-200 bg-white hover:border-gray-300'
                  }`}
                >
                  <input
                    type="radio"
                    name="paymentMethod"
                    value={method.id}
                    checked={paymentMethod === method.id}
                    onChange={(e) => setPaymentMethod(e.target.value)}
                    className="w-4 h-4 text-blue-600"
                  />
                  <div className="ml-4 flex-1">
                    <p className="font-semibold text-gray-900">
                      {method.icon} {method.name}
                    </p>
                    <p className="text-sm text-gray-600">{method.description}</p>
                  </div>
                </label>
              ))}
            </div>
          )}

          {/* Payment Details Form */}
          {step === 'payment' && paymentMethod === 'stripe' && (
            <form onSubmit={handleCreateSubscription} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Email Address
                </label>
                <input
                  type="email"
                  value={stripeEmail}
                  onChange={(e) => setStripeEmail(e.target.value)}
                  placeholder="you@example.com"
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                />
              </div>

              <div className="bg-yellow-50 p-3 rounded-lg text-sm text-yellow-800">
                <p className="font-semibold mb-1">Test Card Information:</p>
                <p>Card: {TEST_CARDS.STRIPE.VISA.number}</p>
                <p>CVV: {TEST_CARDS.CVV}</p>
                <p>Date: {TEST_CARDS.FUTURE_DATE}</p>
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-semibold py-2 rounded-lg transition"
              >
                {loading ? 'Processing...' : 'Continue to Payment'}
              </button>
            </form>
          )}

          {/* PayPal Form */}
          {step === 'payment' && paymentMethod === 'paypal' && (
            <form onSubmit={handleCreateSubscription} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  PayPal Email
                </label>
                <input
                  type="email"
                  value={paypalEmail}
                  onChange={(e) => setPaypalEmail(e.target.value)}
                  placeholder="your-paypal@example.com"
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                />
              </div>

              <div className="bg-blue-50 p-3 rounded-lg text-sm text-blue-800">
                <p className="font-semibold mb-1">You will be redirected to PayPal</p>
                <p>to complete your subscription securely.</p>
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-semibold py-2 rounded-lg transition"
              >
                {loading ? 'Processing...' : 'Continue to PayPal'}
              </button>
            </form>
          )}

          {/* Direct Card Form */}
          {step === 'payment' && paymentMethod === 'card' && (
            <form onSubmit={handleCreateSubscription} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Email Address
                </label>
                <input
                  type="email"
                  value={stripeEmail}
                  onChange={(e) => setStripeEmail(e.target.value)}
                  placeholder="you@example.com"
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                />
              </div>

              <div className="bg-green-50 p-3 rounded-lg text-sm text-green-800">
                <p className="font-semibold mb-1">Secure Card Processing</p>
                <p>Your card details are encrypted and secure.</p>
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-semibold py-2 rounded-lg transition"
              >
                {loading ? 'Processing...' : 'Continue to Card Payment'}
              </button>
            </form>
          )}

          {/* Navigation Buttons */}
          {step === 'payment' && (
            <button
              onClick={() => setStep('method')}
              className="w-full mt-2 text-blue-600 hover:text-blue-700 font-semibold py-2"
            >
              ← Back to Methods
            </button>
          )}

          {step === 'method' && (
            <button
              onClick={() => setStep('payment')}
              className="w-full mt-4 bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 rounded-lg transition"
            >
              Next →
            </button>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 bg-gray-50 border-t border-gray-200 rounded-b-lg text-xs text-gray-600 text-center">
          <p>💳 Secure payment powered by Stripe, PayPal, and industry standards</p>
          <p className="mt-2">🔒 Your payment information is encrypted and PCI-DSS compliant</p>
        </div>
      </div>
    </div>
  );
};

export default PaymentModal;
