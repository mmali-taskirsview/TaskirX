import React, { useState, useEffect } from 'react';
import { Plus, TrendingUp, Download, Filter } from 'lucide-react';

const BiddingInterface = ({ onUpdate }) => {
  const [bids, setBids] = useState([]);
  const [recommendations, setRecommendations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showBidModal, setShowBidModal] = useState(false);
  const [filter, setFilter] = useState('all');
  const [formData, setFormData] = useState({
    campaignId: '',
    adSlotId: '',
    bidAmount: '',
    bidCurrency: 'USD',
    strategy: 'second-price',
    maxBidAmount: ''
  });

  useEffect(() => {
    fetchBids();
    fetchRecommendations();
  }, [filter]);

  const fetchBids = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/bids?status=${filter}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) throw new Error('Failed to fetch bids');

      const data = await response.json();
      setBids(data.bids || []);
      setError(null);
    } catch (err) {
      console.error('Error fetching bids:', err);
      setError(err.message);
      setBids([]);
    } finally {
      setLoading(false);
    }
  };

  const fetchRecommendations = async () => {
    try {
      const response = await fetch('/api/bids/recommendations', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) throw new Error('Failed to fetch recommendations');

      const data = await response.json();
      setRecommendations(data.recommendations || []);
    } catch (err) {
      console.error('Error fetching recommendations:', err);
    }
  };

  const handleSubmitBid = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch('/api/bids/submit', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({
          ...formData,
          bidAmount: parseInt(formData.bidAmount),
          maxBidAmount: parseInt(formData.maxBidAmount)
        })
      });

      if (!response.ok) throw new Error('Failed to submit bid');

      setShowBidModal(false);
      resetForm();
      fetchBids();
      fetchRecommendations();
      onUpdate?.();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleAcceptRecommendation = async (recommendation) => {
    try {
      const response = await fetch('/api/bids/submit', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({
          campaignId: recommendation.campaignId,
          adSlotId: recommendation.adSlotId,
          bidAmount: recommendation.recommendedBid,
          bidCurrency: 'USD',
          strategy: recommendation.strategy,
          maxBidAmount: recommendation.recommendedMaxBid
        })
      });

      if (!response.ok) throw new Error('Failed to submit bid');

      fetchBids();
      fetchRecommendations();
      onUpdate?.();
      alert('Bid submitted successfully!');
    } catch (err) {
      setError(err.message);
    }
  };

  const resetForm = () => {
    setFormData({
      campaignId: '',
      adSlotId: '',
      bidAmount: '',
      bidCurrency: 'USD',
      strategy: 'second-price',
      maxBidAmount: ''
    });
  };

  const getBidStatusColor = (status) => {
    const colors = {
      'pending': 'bg-blue-100 text-blue-800',
      'accepted': 'bg-green-100 text-green-800',
      'rejected': 'bg-red-100 text-red-800',
      'expired': 'bg-gray-100 text-gray-800',
      'won': 'bg-purple-100 text-purple-800'
    };
    return colors[status] || colors.pending;
  };

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow p-8 text-center">
        <div className="animate-spin inline-block">
          <svg className="h-12 w-12 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          </svg>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-gray-900">💰 Bidding Interface</h2>
        <button
          onClick={() => {
            resetForm();
            setShowBidModal(true);
          }}
          className="bg-green-600 hover:bg-green-700 text-white font-medium py-2 px-4 rounded flex items-center gap-2 transition-colors"
        >
          <Plus size={20} />
          Submit Bid
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-800">
          {error}
        </div>
      )}

      {/* Recommendations Section */}
      {recommendations.length > 0 && (
        <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg shadow p-6 border border-blue-200">
          <div className="flex items-center gap-3 mb-4">
            <TrendingUp className="text-blue-600" size={24} />
            <h3 className="text-lg font-bold text-gray-900">🤖 AI Bid Recommendations</h3>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {recommendations.map((rec, idx) => (
              <div key={idx} className="bg-white rounded-lg p-4 border border-blue-200">
                <div className="mb-3">
                  <p className="text-sm text-gray-600">Campaign: {rec.campaignId}</p>
                  <p className="font-medium text-gray-900">{rec.adSlotId}</p>
                </div>
                <div className="space-y-2 mb-4">
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Recommended Bid:</span>
                    <span className="font-bold text-blue-600">${(rec.recommendedBid / 100).toFixed(4)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Strategy:</span>
                    <span className="font-medium text-gray-900">{rec.strategy}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Win Probability:</span>
                    <span className="font-bold text-green-600">{(rec.winProbability * 100).toFixed(1)}%</span>
                  </div>
                </div>
                <button
                  onClick={() => handleAcceptRecommendation(rec)}
                  className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-3 rounded text-sm transition-colors"
                >
                  Accept Recommendation
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="bg-white rounded-lg shadow p-6">
          <p className="text-gray-600 text-sm font-medium">Total Bids</p>
          <p className="text-3xl font-bold text-gray-900 mt-2">{bids.length}</p>
        </div>
        <div className="bg-white rounded-lg shadow p-6">
          <p className="text-gray-600 text-sm font-medium">Accepted</p>
          <p className="text-3xl font-bold text-green-600 mt-2">
            {bids.filter(b => b.status === 'accepted').length}
          </p>
        </div>
        <div className="bg-white rounded-lg shadow p-6">
          <p className="text-gray-600 text-sm font-medium">Win Rate</p>
          <p className="text-3xl font-bold text-purple-600 mt-2">
            {bids.length > 0 
              ? ((bids.filter(b => b.status === 'won').length / bids.length) * 100).toFixed(1)
              : '0'
            }%
          </p>
        </div>
        <div className="bg-white rounded-lg shadow p-6">
          <p className="text-gray-600 text-sm font-medium">Avg Bid Amount</p>
          <p className="text-3xl font-bold text-orange-600 mt-2">
            ${bids.length > 0
              ? (bids.reduce((sum, b) => sum + b.bidAmount, 0) / bids.length / 100).toFixed(2)
              : '0.00'
            }
          </p>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-4 flex gap-4 flex-wrap">
        <select
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          className="px-4 py-2 border border-gray-300 rounded text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="all">All Bids</option>
          <option value="pending">Pending</option>
          <option value="accepted">Accepted</option>
          <option value="won">Won</option>
          <option value="rejected">Rejected</option>
        </select>
      </div>

      {/* Bids Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50 border-b border-gray-200">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Campaign ID</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Ad Slot</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Bid Amount</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Status</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Strategy</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Submitted</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Result</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {bids.length === 0 ? (
              <tr>
                <td colSpan="7" className="px-6 py-8 text-center text-gray-500">
                  No bids yet. Submit a bid or accept a recommendation to get started!
                </td>
              </tr>
            ) : (
              bids.map(bid => (
                <tr key={bid.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 text-gray-900 font-medium">{bid.campaignId}</td>
                  <td className="px-6 py-4 text-gray-900">{bid.adSlotId}</td>
                  <td className="px-6 py-4 text-gray-900 font-bold">${(bid.bidAmount / 100).toFixed(4)}</td>
                  <td className="px-6 py-4">
                    <span className={`inline-block px-3 py-1 rounded-full text-sm font-medium ${getBidStatusColor(bid.status)}`}>
                      {bid.status.charAt(0).toUpperCase() + bid.status.slice(1)}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-gray-900">{bid.strategy}</td>
                  <td className="px-6 py-4 text-gray-900 text-sm">
                    {new Date(bid.createdAt).toLocaleDateString()} {new Date(bid.createdAt).toLocaleTimeString()}
                  </td>
                  <td className="px-6 py-4 text-gray-900">
                    {bid.status === 'won' ? (
                      <span className="text-green-600 font-medium">✓ Won</span>
                    ) : bid.status === 'rejected' ? (
                      <span className="text-red-600 font-medium">✗ Lost</span>
                    ) : (
                      <span className="text-gray-600">-</span>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Modal */}
      {showBidModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4">
            <div className="p-6 border-b border-gray-200">
              <h3 className="text-xl font-bold text-gray-900">Submit New Bid</h3>
            </div>

            <form onSubmit={handleSubmitBid} className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Campaign ID *</label>
                  <input
                    type="text"
                    required
                    value={formData.campaignId}
                    onChange={(e) => setFormData({...formData, campaignId: e.target.value})}
                    className="w-full px-3 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="campaign_123"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Ad Slot ID *</label>
                  <input
                    type="text"
                    required
                    value={formData.adSlotId}
                    onChange={(e) => setFormData({...formData, adSlotId: e.target.value})}
                    className="w-full px-3 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="slot_456"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Bid Amount (cents) *</label>
                  <input
                    type="number"
                    required
                    min="1"
                    value={formData.bidAmount}
                    onChange={(e) => setFormData({...formData, bidAmount: e.target.value})}
                    className="w-full px-3 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Max Bid Amount (cents) *</label>
                  <input
                    type="number"
                    required
                    min="1"
                    value={formData.maxBidAmount}
                    onChange={(e) => setFormData({...formData, maxBidAmount: e.target.value})}
                    className="w-full px-3 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="1000"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Bid Strategy</label>
                  <select
                    value={formData.strategy}
                    onChange={(e) => setFormData({...formData, strategy: e.target.value})}
                    className="w-full px-3 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="second-price">Second Price</option>
                    <option value="first-price">First Price</option>
                    <option value="programmatic">Programmatic</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Currency</label>
                  <select
                    value={formData.bidCurrency}
                    onChange={(e) => setFormData({...formData, bidCurrency: e.target.value})}
                    className="w-full px-3 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="USD">USD</option>
                    <option value="EUR">EUR</option>
                    <option value="GBP">GBP</option>
                    <option value="JPY">JPY</option>
                  </select>
                </div>
              </div>

              <div className="flex justify-end gap-3 pt-6 border-t border-gray-200">
                <button
                  type="button"
                  onClick={() => {
                    setShowBidModal(false);
                    resetForm();
                  }}
                  className="px-4 py-2 text-gray-700 border border-gray-300 rounded hover:bg-gray-50 transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
                >
                  Submit Bid
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default BiddingInterface;
