'use client';

import React, { useState, useEffect } from 'react';
import {
  Package,
  Plus,
  Search,
  Filter,
  MoreVertical,
  Eye,
  DollarSign,
  Zap,
  Edit,
  Trash2,
  Copy,
  ExternalLink,
  Globe,
  Monitor,
  Smartphone,
  Loader2,
  AlertCircle
} from 'lucide-react';
import { api } from '@/lib/api';

// Types
interface AdUnit {
  id: string;
  name: string;
  size: string;
  type: string;
  status: string;
  domain: string;
  pageUrl?: string;
  floorPrice: number;
  currency: string;
  impressions: number;
  requests: number;
  revenue: number;
  publisherId: string;
  createdAt: string;
  updatedAt: string;
}

const platformIcons: Record<string, React.ElementType> = {
  desktop: Monitor,
  mobile: Smartphone,
  tablet: Globe,
};

// Format large numbers
const formatNumber = (num: number): string => {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
  return num.toString();
};

// Calculate eCPM
const calculateEcpm = (revenue: number, impressions: number): string => {
  if (impressions === 0) return '-';
  return `$${((revenue / impressions) * 1000).toFixed(2)}`;
};

export default function InventoryPage() {
  const [adUnits, setAdUnits] = useState<AdUnit[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterType, setFilterType] = useState('all');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    type: 'banner',
    size: '728x90',
    domain: '',
    floorPrice: '2.00',
  });

  // Fetch ad units from API
  useEffect(() => {
    fetchAdUnits();
  }, []);

  const fetchAdUnits = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.getAdUnits();
      setAdUnits(response.data);
    } catch (err: any) {
      console.error('Failed to fetch ad units:', err);
      setError(err.message || 'Failed to load ad units');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateAdUnit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.createAdUnit({
        name: formData.name,
        type: formData.type,
        size: formData.size,
        domain: formData.domain,
        floorPrice: parseFloat(formData.floorPrice),
        status: 'active',
        currency: 'USD',
      });
      setShowCreateModal(false);
      setFormData({ name: '', type: 'banner', size: '728x90', domain: '', floorPrice: '2.00' });
      fetchAdUnits(); // Refresh list
    } catch (err: any) {
      console.error('Failed to create ad unit:', err);
      alert('Failed to create ad unit: ' + (err.message || 'Unknown error'));
    }
  };

  const handleDeleteAdUnit = async (id: string) => {
    if (!confirm('Are you sure you want to delete this ad unit?')) return;
    try {
      await api.deleteAdUnit(id);
      fetchAdUnits(); // Refresh list
    } catch (err: any) {
      console.error('Failed to delete ad unit:', err);
      alert('Failed to delete ad unit');
    }
  };

  const filteredUnits = adUnits.filter((unit) => {
    const matchesSearch = unit.name.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesType = filterType === 'all' || unit.type.toLowerCase() === filterType;
    return matchesSearch && matchesType;
  });

  // Calculate totals
  const totalImpressions = adUnits.reduce((sum, u) => sum + (u.impressions || 0), 0);
  const totalRevenue = adUnits.reduce((sum, u) => sum + parseFloat(String(u.revenue || 0)), 0);
  const avgFillRate = adUnits.length > 0 
    ? (adUnits.filter(u => u.status === 'active').length / adUnits.length) * 100 
    : 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Ad Inventory</h1>
          <p className="text-gray-500 mt-1">Manage your ad units and placements</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg text-sm font-medium hover:bg-emerald-700"
        >
          <Plus className="w-4 h-4" />
          Create Ad Unit
        </button>
      </div>

      {/* Stats Summary */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-emerald-100 rounded-lg">
              <Package className="w-5 h-5 text-emerald-600" />
            </div>
            <div>
              <p className="text-2xl font-bold text-gray-900">{adUnits.length}</p>
              <p className="text-sm text-gray-500">Total Ad Units</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-100 rounded-lg">
              <Eye className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-2xl font-bold text-gray-900">{formatNumber(totalImpressions)}</p>
              <p className="text-sm text-gray-500">Total Impressions</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-purple-100 rounded-lg">
              <DollarSign className="w-5 h-5 text-purple-600" />
            </div>
            <div>
              <p className="text-2xl font-bold text-gray-900">${totalRevenue.toLocaleString()}</p>
              <p className="text-sm text-gray-500">Total Revenue</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-amber-100 rounded-lg">
              <Zap className="w-5 h-5 text-amber-600" />
            </div>
            <div>
              <p className="text-2xl font-bold text-gray-900">{avgFillRate.toFixed(1)}%</p>
              <p className="text-sm text-gray-500">Active Rate</p>
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-4 bg-white rounded-lg border border-gray-200 p-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search ad units..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500"
          />
        </div>
        <select
          value={filterType}
          onChange={(e) => setFilterType(e.target.value)}
          className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
        >
          <option value="all">All Types</option>
          <option value="display">Display</option>
          <option value="video">Video</option>
          <option value="native">Native</option>
          <option value="interstitial">Interstitial</option>
        </select>
        <button className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">
          <Filter className="w-4 h-4" />
          More Filters
        </button>
      </div>

      {/* Ad Units Table */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-8 h-8 animate-spin text-emerald-600" />
            <span className="ml-2 text-gray-500">Loading ad units...</span>
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
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Ad Unit</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Size / Type</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Domain</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Floor Price</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Impressions</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Revenue</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">eCPM</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {filteredUnits.length === 0 ? (
              <tr>
                <td colSpan={9} className="px-6 py-8 text-center text-gray-500">
                  No ad units found. Create your first ad unit to get started.
                </td>
              </tr>
            ) : (
            filteredUnits.map((unit) => (
              <tr key={unit.id} className="hover:bg-gray-50">
                <td className="px-6 py-4">
                  <div>
                    <p className="font-medium text-gray-900">{unit.name}</p>
                    <p className="text-xs text-gray-500">{unit.id.slice(0, 8)}...</p>
                  </div>
                </td>
                <td className="px-6 py-4">
                  <div>
                    <p className="text-gray-900">{unit.size}</p>
                    <span className="inline-flex px-2 py-0.5 text-xs font-medium bg-gray-100 text-gray-700 rounded capitalize">
                      {unit.type}
                    </span>
                  </div>
                </td>
                <td className="px-6 py-4">
                  <div className="flex items-center gap-1">
                    <Globe className="w-4 h-4 text-gray-400" />
                    <span className="text-sm text-gray-600">{unit.domain || '-'}</span>
                  </div>
                </td>
                <td className="px-6 py-4">
                  <span className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${
                    unit.status === 'active'
                      ? 'bg-green-100 text-green-700'
                      : 'bg-gray-100 text-gray-600'
                  }`}>
                    {unit.status}
                  </span>
                </td>
                <td className="px-6 py-4 text-right text-gray-900">
                  ${parseFloat(String(unit.floorPrice)).toFixed(2)}
                </td>
                <td className="px-6 py-4 text-right text-gray-900">{formatNumber(unit.impressions || 0)}</td>
                <td className="px-6 py-4 text-right font-medium text-gray-900">
                  ${parseFloat(String(unit.revenue || 0)).toFixed(2)}
                </td>
                <td className="px-6 py-4 text-right text-gray-900">
                  {calculateEcpm(parseFloat(String(unit.revenue || 0)), unit.impressions || 0)}
                </td>
                <td className="px-6 py-4 text-right">
                  <div className="flex items-center justify-end gap-2">
                    <button className="p-1 text-gray-400 hover:text-gray-600" title="Edit">
                      <Edit className="w-4 h-4" />
                    </button>
                    <button className="p-1 text-gray-400 hover:text-gray-600" title="Copy Tag">
                      <Copy className="w-4 h-4" />
                    </button>
                    <button 
                      onClick={() => handleDeleteAdUnit(unit.id)}
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
            <h2 className="text-xl font-bold text-gray-900 mb-4">Create Ad Unit</h2>
            <form className="space-y-4" onSubmit={handleCreateAdUnit}>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Ad Unit Name</label>
                <input
                  type="text"
                  placeholder="e.g., Homepage Banner"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                  required
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Ad Type</label>
                  <select 
                    value={formData.type}
                    onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                  >
                    <option value="banner">Banner</option>
                    <option value="video">Video</option>
                    <option value="native">Native</option>
                    <option value="interstitial">Interstitial</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Size</label>
                  <select 
                    value={formData.size}
                    onChange={(e) => setFormData({ ...formData, size: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                  >
                    <option value="728x90">728x90 (Leaderboard)</option>
                    <option value="300x250">300x250 (Medium Rectangle)</option>
                    <option value="320x50">320x50 (Mobile Banner)</option>
                    <option value="160x600">160x600 (Wide Skyscraper)</option>
                    <option value="640x360">640x360 (Video)</option>
                    <option value="responsive">Responsive</option>
                  </select>
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Domain</label>
                <input
                  type="text"
                  placeholder="e.g., example.com"
                  value={formData.domain}
                  onChange={(e) => setFormData({ ...formData, domain: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Floor Price (CPM)</label>
                <div className="relative">
                  <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">$</span>
                  <input
                    type="number"
                    step="0.01"
                    placeholder="2.00"
                    value={formData.floorPrice}
                    onChange={(e) => setFormData({ ...formData, floorPrice: e.target.value })}
                    className="w-full pl-8 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
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
                  Create Ad Unit
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
