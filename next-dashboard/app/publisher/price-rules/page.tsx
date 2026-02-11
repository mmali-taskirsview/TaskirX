'use client';

import React, { useState } from 'react';
import {
  DollarSign,
  Plus,
  Search,
  Edit,
  Trash2,
  Save,
  X,
  AlertCircle,
  TrendingUp,
  TrendingDown,
  Clock,
  Globe,
  Monitor,
  Smartphone,
  Layers,
  Target
} from 'lucide-react';

// Mock price rules data
const priceRules = [
  {
    id: 'rule_001',
    name: 'Premium Geo - US/UK',
    type: 'geo',
    conditions: { countries: ['US', 'UK', 'CA', 'AU'] },
    floorPrice: 8.00,
    adjustment: '+60%',
    priority: 1,
    status: 'active',
    impressions: 450000,
    revenue: 36000
  },
  {
    id: 'rule_002',
    name: 'Peak Hours Boost',
    type: 'time',
    conditions: { hours: '18:00 - 23:00', timezone: 'EST' },
    floorPrice: 6.50,
    adjustment: '+30%',
    priority: 2,
    status: 'active',
    impressions: 280000,
    revenue: 18200
  },
  {
    id: 'rule_003',
    name: 'Mobile Traffic Premium',
    type: 'device',
    conditions: { devices: ['mobile', 'tablet'] },
    floorPrice: 5.50,
    adjustment: '+10%',
    priority: 3,
    status: 'active',
    impressions: 520000,
    revenue: 28600
  },
  {
    id: 'rule_004',
    name: 'Above Fold Inventory',
    type: 'placement',
    conditions: { position: 'above_fold', viewability: '>70%' },
    floorPrice: 7.00,
    adjustment: '+40%',
    priority: 4,
    status: 'active',
    impressions: 320000,
    revenue: 22400
  },
  {
    id: 'rule_005',
    name: 'Weekend Discount',
    type: 'time',
    conditions: { days: ['Saturday', 'Sunday'] },
    floorPrice: 3.50,
    adjustment: '-30%',
    priority: 5,
    status: 'paused',
    impressions: 180000,
    revenue: 6300
  },
  {
    id: 'rule_006',
    name: 'Video Content Premium',
    type: 'format',
    conditions: { formats: ['video_preroll', 'video_midroll'] },
    floorPrice: 15.00,
    adjustment: '+200%',
    priority: 6,
    status: 'active',
    impressions: 95000,
    revenue: 14250
  },
  {
    id: 'rule_007',
    name: 'Low-Tier Geo Minimum',
    type: 'geo',
    conditions: { countries: ['IN', 'BR', 'ID', 'PH'] },
    floorPrice: 1.00,
    adjustment: '-80%',
    priority: 7,
    status: 'active',
    impressions: 850000,
    revenue: 8500
  }
];

export default function PriceRulesPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingRule, setEditingRule] = useState<string | null>(null);

  const getRuleIcon = (type: string) => {
    switch (type) {
      case 'geo': return <Globe className="w-4 h-4" />;
      case 'time': return <Clock className="w-4 h-4" />;
      case 'device': return <Smartphone className="w-4 h-4" />;
      case 'placement': return <Layers className="w-4 h-4" />;
      case 'format': return <Monitor className="w-4 h-4" />;
      default: return <Target className="w-4 h-4" />;
    }
  };

  const getRuleColor = (type: string) => {
    switch (type) {
      case 'geo': return 'bg-blue-100 text-blue-800';
      case 'time': return 'bg-purple-100 text-purple-800';
      case 'device': return 'bg-green-100 text-green-800';
      case 'placement': return 'bg-orange-100 text-orange-800';
      case 'format': return 'bg-pink-100 text-pink-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getConditionText = (rule: typeof priceRules[0]) => {
    const conditions = rule.conditions;
    if ('countries' in conditions) return `Countries: ${(conditions as any).countries.join(', ')}`;
    if ('hours' in conditions) return `Hours: ${(conditions as any).hours} (${(conditions as any).timezone})`;
    if ('devices' in conditions) return `Devices: ${(conditions as any).devices.join(', ')}`;
    if ('position' in conditions) return `Position: ${(conditions as any).position}, Viewability: ${(conditions as any).viewability}`;
    if ('formats' in conditions) return `Formats: ${(conditions as any).formats.join(', ')}`;
    if ('days' in conditions) return `Days: ${(conditions as any).days.join(', ')}`;
    return 'Custom conditions';
  };

  const filteredRules = priceRules.filter(rule => {
    const matchesSearch = rule.name.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesType = filterType === 'all' || rule.type === filterType;
    return matchesSearch && matchesType;
  });

  const totalRevenue = priceRules.reduce((sum, rule) => sum + rule.revenue, 0);
  const totalImpressions = priceRules.reduce((sum, rule) => sum + rule.impressions, 0);
  const avgFloor = (priceRules.reduce((sum, rule) => sum + rule.floorPrice, 0) / priceRules.length).toFixed(2);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Price Rules</h1>
          <p className="text-gray-600 mt-1">Dynamic floor pricing based on conditions</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
        >
          <Plus className="w-4 h-4" />
          Create Rule
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-emerald-100 rounded-lg">
              <Target className="w-5 h-5 text-emerald-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Active Rules</p>
              <p className="text-xl font-bold text-gray-900">{priceRules.filter(r => r.status === 'active').length}</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-100 rounded-lg">
              <DollarSign className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Avg Floor Price</p>
              <p className="text-xl font-bold text-gray-900">${avgFloor}</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-purple-100 rounded-lg">
              <TrendingUp className="w-5 h-5 text-purple-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Rule-Impacted Revenue</p>
              <p className="text-xl font-bold text-gray-900">${(totalRevenue / 1000).toFixed(1)}K</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-green-100 rounded-lg">
              <Layers className="w-5 h-5 text-green-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Impressions Affected</p>
              <p className="text-xl font-bold text-gray-900">{(totalImpressions / 1000000).toFixed(2)}M</p>
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-xl p-4 border border-gray-200">
        <div className="flex flex-wrap gap-4 items-center">
          <div className="relative flex-1 min-w-[200px]">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
            <input
              type="text"
              placeholder="Search rules..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
            />
          </div>
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            className="px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
          >
            <option value="all">All Types</option>
            <option value="geo">Geographic</option>
            <option value="time">Time-based</option>
            <option value="device">Device</option>
            <option value="placement">Placement</option>
            <option value="format">Format</option>
          </select>
        </div>
      </div>

      {/* Rules List */}
      <div className="space-y-4">
        {filteredRules.map((rule, index) => (
          <div
            key={rule.id}
            className={`bg-white rounded-xl border ${rule.status === 'active' ? 'border-gray-200' : 'border-gray-100 opacity-60'} overflow-hidden`}
          >
            <div className="p-4">
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-4">
                  <div className="flex items-center justify-center w-8 h-8 bg-gray-100 rounded-full text-sm font-medium text-gray-600">
                    {rule.priority}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="font-semibold text-gray-900">{rule.name}</h3>
                      <span className={`flex items-center gap-1 px-2 py-0.5 text-xs font-medium rounded-full ${getRuleColor(rule.type)}`}>
                        {getRuleIcon(rule.type)}
                        {rule.type}
                      </span>
                      {rule.status !== 'active' && (
                        <span className="px-2 py-0.5 text-xs font-medium bg-gray-100 text-gray-600 rounded-full">
                          Paused
                        </span>
                      )}
                    </div>
                    <p className="text-sm text-gray-500 mt-1">{getConditionText(rule)}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => setEditingRule(rule.id)}
                    className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
                  >
                    <Edit className="w-4 h-4" />
                  </button>
                  <button className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors">
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
              
              <div className="flex items-center gap-8 mt-4 pt-4 border-t border-gray-100">
                <div>
                  <p className="text-xs text-gray-500 uppercase">Floor Price</p>
                  <p className="text-lg font-bold text-gray-900">${rule.floorPrice.toFixed(2)}</p>
                </div>
                <div>
                  <p className="text-xs text-gray-500 uppercase">Adjustment</p>
                  <p className={`text-lg font-bold ${rule.adjustment.startsWith('+') ? 'text-green-600' : 'text-red-600'}`}>
                    {rule.adjustment.startsWith('+') ? (
                      <span className="flex items-center gap-1">
                        <TrendingUp className="w-4 h-4" />
                        {rule.adjustment}
                      </span>
                    ) : (
                      <span className="flex items-center gap-1">
                        <TrendingDown className="w-4 h-4" />
                        {rule.adjustment}
                      </span>
                    )}
                  </p>
                </div>
                <div>
                  <p className="text-xs text-gray-500 uppercase">Impressions</p>
                  <p className="text-lg font-bold text-gray-900">{(rule.impressions / 1000).toFixed(0)}K</p>
                </div>
                <div>
                  <p className="text-xs text-gray-500 uppercase">Revenue</p>
                  <p className="text-lg font-bold text-emerald-600">${(rule.revenue / 1000).toFixed(1)}K</p>
                </div>
                <div>
                  <p className="text-xs text-gray-500 uppercase">eCPM</p>
                  <p className="text-lg font-bold text-gray-900">${((rule.revenue / rule.impressions) * 1000).toFixed(2)}</p>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Tips Card */}
      <div className="bg-amber-50 border border-amber-200 rounded-xl p-4">
        <div className="flex items-start gap-3">
          <AlertCircle className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
          <div>
            <h4 className="font-medium text-amber-800">Price Rule Tips</h4>
            <ul className="mt-2 text-sm text-amber-700 space-y-1">
              <li>• Rules are evaluated in priority order (1 = highest priority)</li>
              <li>• First matching rule wins - place specific rules before general ones</li>
              <li>• Use geo targeting for premium markets (US, UK, CA typically have higher CPMs)</li>
              <li>• Consider time-based rules for peak traffic hours</li>
            </ul>
          </div>
        </div>
      </div>

      {/* Create Rule Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-lg max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold text-gray-900 mb-4">Create Price Rule</h2>
            <form className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Rule Name</label>
                <input
                  type="text"
                  placeholder="e.g., Premium US Traffic"
                  className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Rule Type</label>
                <select className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500">
                  <option value="geo">Geographic</option>
                  <option value="time">Time-based</option>
                  <option value="device">Device Type</option>
                  <option value="placement">Placement</option>
                  <option value="format">Ad Format</option>
                </select>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Floor Price ($)</label>
                  <input
                    type="number"
                    step="0.01"
                    placeholder="5.00"
                    className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Priority</label>
                  <input
                    type="number"
                    min="1"
                    placeholder="1"
                    className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                  />
                </div>
              </div>
              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="flex-1 px-4 py-2 border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700"
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
