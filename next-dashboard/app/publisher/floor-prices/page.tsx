'use client';

import React, { useState, useEffect } from 'react';
import {
  TrendingUp,
  DollarSign,
  Sliders,
  Plus,
  Search,
  AlertCircle,
  CheckCircle,
  Clock,
  ArrowUpRight,
  ArrowDownRight,
  BarChart3,
  Zap,
  Loader2,
  Trash2,
  Edit
} from 'lucide-react';
import { api } from '@/lib/api';

// Types
interface FloorPrice {
  id: string;
  name: string;
  description?: string;
  ruleType: string;
  price: number;
  currency: string;
  action: string;
  conditions?: Record<string, any>;
  publisherId?: string;
  adUnitId?: string;
  placementId?: string;
  priority: number;
  isActive: boolean;
  startDate?: string;
  endDate?: string;
  createdAt: string;
  updatedAt: string;
}

// Suggested optimizations
const suggestions = [
  {
    type: 'increase',
    rule: 'Premium Desktop',
    suggestion: 'Increase floor from $3.50 to $4.00',
    reason: 'eCPM consistently 38% above floor',
    impact: '+$420/day estimated',
  },
  {
    type: 'decrease',
    rule: 'Mobile Banner - US',
    suggestion: 'Decrease floor from $2.00 to $1.50',
    reason: 'Fill rate below target (71%)',
    impact: '+15% fill rate expected',
  },
];

export default function FloorPricesPage() {
  const [floorPriceRules, setFloorPriceRules] = useState<FloorPrice[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    ruleType: 'global',
    price: '1.00',
    action: 'set_floor',
  });

  // Fetch floor prices from API
  useEffect(() => {
    fetchFloorPrices();
  }, []);

  const fetchFloorPrices = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.getFloorPrices();
      setFloorPriceRules(response.data);
    } catch (err: any) {
      console.error('Failed to fetch floor prices:', err);
      setError(err.message || 'Failed to load floor prices');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateRule = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.createFloorPrice({
        name: formData.name,
        ruleType: formData.ruleType,
        price: parseFloat(formData.price),
        action: formData.action,
        currency: 'USD',
        isActive: true,
        priority: 1,
      });
      setShowCreateModal(false);
      setFormData({ name: '', ruleType: 'global', price: '1.00', action: 'set_floor' });
      fetchFloorPrices();
    } catch (err: any) {
      console.error('Failed to create floor price rule:', err);
      alert('Failed to create rule: ' + (err.message || 'Unknown error'));
    }
  };

  const handleDeleteRule = async (id: string) => {
    if (!confirm('Are you sure you want to delete this floor price rule?')) return;
    try {
      await api.deleteFloorPrice(id);
      fetchFloorPrices();
    } catch (err: any) {
      console.error('Failed to delete rule:', err);
      alert('Failed to delete rule');
    }
  };

  // Calculate stats
  const avgFloorPrice = floorPriceRules.length > 0
    ? floorPriceRules.reduce((sum, r) => sum + parseFloat(String(r.price)), 0) / floorPriceRules.length
    : 0;
  const activeRules = floorPriceRules.filter(r => r.isActive).length;

  const filteredRules = floorPriceRules.filter(rule =>
    rule.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Floor Price Management</h1>
          <p className="text-gray-500 mt-1">Set minimum bid prices to maximize revenue</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg text-sm font-medium hover:bg-emerald-700"
        >
          <Plus className="w-4 h-4" />
          Create Rule
        </button>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Avg Floor Price</p>
              <p className="text-2xl font-bold text-gray-900">${avgFloorPrice.toFixed(2)}</p>
            </div>
            <div className="p-2 bg-emerald-100 rounded-lg">
              <DollarSign className="w-5 h-5 text-emerald-600" />
            </div>
          </div>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Total Rules</p>
              <p className="text-2xl font-bold text-gray-900">{floorPriceRules.length}</p>
            </div>
            <div className="p-2 bg-blue-100 rounded-lg">
              <TrendingUp className="w-5 h-5 text-blue-600" />
            </div>
          </div>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Active Rules</p>
              <p className="text-2xl font-bold text-gray-900">{activeRules}</p>
            </div>
            <div className="p-2 bg-purple-100 rounded-lg">
              <Zap className="w-5 h-5 text-purple-600" />
            </div>
          </div>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Rule Types</p>
              <p className="text-2xl font-bold text-gray-900">
                {new Set(floorPriceRules.map(r => r.ruleType)).size}
              </p>
            </div>
            <div className="p-2 bg-amber-100 rounded-lg">
              <Sliders className="w-5 h-5 text-amber-600" />
            </div>
          </div>
        </div>
      </div>

      {/* AI Suggestions */}
      <div className="bg-gradient-to-r from-emerald-50 to-teal-50 border border-emerald-200 rounded-xl p-6">
        <div className="flex items-center gap-2 mb-4">
          <Zap className="w-5 h-5 text-emerald-600" />
          <h2 className="text-lg font-semibold text-gray-900">AI Optimization Suggestions</h2>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {suggestions.map((suggestion, index) => (
            <div key={index} className="bg-white rounded-lg p-4 shadow-sm">
              <div className="flex items-start gap-3">
                <div className={`p-2 rounded-lg ${
                  suggestion.type === 'increase' ? 'bg-green-100' : 'bg-amber-100'
                }`}>
                  {suggestion.type === 'increase' ? (
                    <ArrowUpRight className="w-4 h-4 text-green-600" />
                  ) : (
                    <ArrowDownRight className="w-4 h-4 text-amber-600" />
                  )}
                </div>
                <div className="flex-1">
                  <p className="font-medium text-gray-900">{suggestion.rule}</p>
                  <p className="text-sm text-gray-600 mt-1">{suggestion.suggestion}</p>
                  <p className="text-xs text-gray-500 mt-1">{suggestion.reason}</p>
                  <div className="flex items-center justify-between mt-3">
                    <span className="text-sm font-medium text-emerald-600">{suggestion.impact}</span>
                    <button className="px-3 py-1 text-sm bg-emerald-600 text-white rounded hover:bg-emerald-700">
                      Apply
                    </button>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Search */}
      <div className="flex items-center gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search floor price rules..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
          />
        </div>
      </div>

      {/* Floor Price Rules Table */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-8 h-8 animate-spin text-emerald-600" />
            <span className="ml-2 text-gray-500">Loading floor prices...</span>
          </div>
        ) : error ? (
          <div className="flex items-center justify-center py-12 text-red-500">
            <AlertCircle className="w-6 h-6 mr-2" />
            {error}
          </div>
        ) : (
        <table className="w-full">
          <thead className="bg-gray-50 border-b border-gray-200">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Rule Name</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
              <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase">Floor Price</th>
              <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase">Action</th>
              <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase">Priority</th>
              <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {filteredRules.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-6 py-8 text-center text-gray-500">
                  No floor price rules found. Create your first rule to get started.
                </td>
              </tr>
            ) : (
            filteredRules.map((rule) => (
              <tr key={rule.id} className="hover:bg-gray-50">
                <td className="px-6 py-4">
                  <div>
                    <p className="font-medium text-gray-900">{rule.name}</p>
                    <p className="text-sm text-gray-500">{rule.description || rule.id.slice(0, 8)}...</p>
                  </div>
                </td>
                <td className="px-6 py-4">
                  <span className={`inline-flex px-2 py-1 text-xs font-medium rounded-full capitalize ${
                    rule.ruleType === 'global' ? 'bg-blue-100 text-blue-700' :
                    rule.ruleType === 'geo' ? 'bg-green-100 text-green-700' :
                    rule.ruleType === 'device' ? 'bg-purple-100 text-purple-700' :
                    'bg-gray-100 text-gray-700'
                  }`}>
                    {rule.ruleType}
                  </span>
                </td>
                <td className="px-6 py-4 text-center">
                  <span className="text-lg font-semibold text-gray-900">
                    ${parseFloat(String(rule.price)).toFixed(2)}
                  </span>
                  <span className="text-xs text-gray-500 ml-1">{rule.currency}</span>
                </td>
                <td className="px-6 py-4 text-center">
                  <span className={`inline-flex px-2 py-1 text-xs font-medium rounded capitalize ${
                    rule.action === 'set_floor' ? 'bg-emerald-100 text-emerald-700' :
                    rule.action === 'multiply' ? 'bg-amber-100 text-amber-700' :
                    'bg-blue-100 text-blue-700'
                  }`}>
                    {rule.action.replace('_', ' ')}
                  </span>
                </td>
                <td className="px-6 py-4 text-center">
                  <span className="font-medium text-gray-700">{rule.priority}</span>
                </td>
                <td className="px-6 py-4 text-center">
                  {rule.isActive ? (
                    <span className="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded-full">
                      <CheckCircle className="w-3 h-3" />
                      Active
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium bg-gray-100 text-gray-600 rounded-full">
                      <Clock className="w-3 h-3" />
                      Inactive
                    </span>
                  )}
                </td>
                <td className="px-6 py-4 text-right">
                  <div className="flex items-center justify-end gap-2">
                    <button className="p-1 text-gray-400 hover:text-gray-600" title="Edit">
                      <Edit className="w-4 h-4" />
                    </button>
                    <button 
                      onClick={() => handleDeleteRule(rule.id)}
                      className="p-1 text-gray-400 hover:text-red-600" 
                      title="Delete"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            ))
            )}
          </tbody>
        </table>
        )}
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-lg p-6">
            <h2 className="text-xl font-bold text-gray-900 mb-4">Create Floor Price Rule</h2>
            <form className="space-y-4" onSubmit={handleCreateRule}>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Rule Name</label>
                <input
                  type="text"
                  placeholder="e.g., Premium US Desktop"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                  required
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Rule Type</label>
                  <select 
                    value={formData.ruleType}
                    onChange={(e) => setFormData({ ...formData, ruleType: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                  >
                    <option value="global">Global</option>
                    <option value="geo">Geographic</option>
                    <option value="device">Device</option>
                    <option value="time">Time-based</option>
                    <option value="format">Format</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Action</label>
                  <select 
                    value={formData.action}
                    onChange={(e) => setFormData({ ...formData, action: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                  >
                    <option value="set_floor">Set Floor</option>
                    <option value="multiply">Multiply</option>
                    <option value="add">Add</option>
                  </select>
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {formData.action === 'multiply' ? 'Multiplier' : 'Floor Price (CPM)'}
                </label>
                <div className="relative">
                  {formData.action !== 'multiply' && (
                    <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">$</span>
                  )}
                  <input
                    type="number"
                    step="0.01"
                    placeholder={formData.action === 'multiply' ? '1.5' : '2.00'}
                    value={formData.price}
                    onChange={(e) => setFormData({ ...formData, price: e.target.value })}
                    className={`w-full ${formData.action !== 'multiply' ? 'pl-8' : 'pl-4'} pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500`}
                    required
                  />
                </div>
              </div>
              <div className="flex justify-end gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700"
                >
                  Create Rule
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
