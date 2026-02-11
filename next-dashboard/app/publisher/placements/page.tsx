'use client';

import React, { useState, useEffect } from 'react';
import {
  Layers,
  Plus,
  Search,
  Filter,
  MoreVertical,
  Eye,
  Edit,
  Trash2,
  Copy,
  Code,
  Monitor,
  Smartphone,
  Tablet,
  CheckCircle,
  XCircle,
  TrendingUp,
  DollarSign,
  BarChart3,
  Loader2,
  AlertCircle
} from 'lucide-react';
import { api } from '@/lib/api';

// Types
interface Placement {
  id: string;
  name: string;
  description?: string;
  status: string;
  position: string;
  domain: string;
  pageType?: string;
  adFormats: string[];
  allowedSizes: string[];
  devices: string[];
  floorPrice: number;
  currency: string;
  viewabilityTarget?: number;
  targeting?: Record<string, any>;
  settings?: Record<string, any>;
  impressions: number;
  revenue: number;
  fillRate: number;
  publisherId: string;
  createdAt: string;
  updatedAt: string;
}

export default function PlacementsPage() {
  const [placements, setPlacements] = useState<Placement[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState('all');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    position: 'above_fold',
    domain: '',
    adFormats: ['banner'],
    allowedSizes: ['300x250'],
    devices: ['desktop', 'mobile'],
    floorPrice: '1.50',
  });

  // Fetch placements from API
  useEffect(() => {
    fetchPlacements();
  }, []);

  const fetchPlacements = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.getPlacements();
      setPlacements(response.data);
    } catch (err: any) {
      console.error('Failed to fetch placements:', err);
      setError(err.message || 'Failed to load placements');
    } finally {
      setLoading(false);
    }
  };

  const handleCreatePlacement = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.createPlacement({
        name: formData.name,
        position: formData.position,
        domain: formData.domain,
        adFormats: formData.adFormats,
        allowedSizes: formData.allowedSizes,
        devices: formData.devices,
        floorPrice: parseFloat(formData.floorPrice),
        status: 'active',
        currency: 'USD',
      });
      setShowCreateModal(false);
      setFormData({
        name: '', position: 'above_fold', domain: '',
        adFormats: ['banner'], allowedSizes: ['300x250'],
        devices: ['desktop', 'mobile'], floorPrice: '1.50'
      });
      fetchPlacements();
    } catch (err: any) {
      console.error('Failed to create placement:', err);
      alert('Failed to create placement: ' + (err.message || 'Unknown error'));
    }
  };

  const handleDeletePlacement = async (id: string) => {
    if (!confirm('Are you sure you want to delete this placement?')) return;
    try {
      await api.deletePlacement(id);
      fetchPlacements();
    } catch (err: any) {
      console.error('Failed to delete placement:', err);
      alert('Failed to delete placement');
    }
  };

  const getDeviceIcon = (device: string) => {
    switch (device) {
      case 'mobile': return <Smartphone className="w-4 h-4" />;
      case 'desktop': return <Monitor className="w-4 h-4" />;
      case 'tablet': return <Tablet className="w-4 h-4" />;
      default: return <Monitor className="w-4 h-4" />;
    }
  };

  const getPositionColor = (position: string) => {
    switch (position) {
      case 'above_fold': return 'bg-green-100 text-green-800';
      case 'in_content': return 'bg-blue-100 text-blue-800';
      case 'below_fold': return 'bg-yellow-100 text-yellow-800';
      case 'sidebar': return 'bg-purple-100 text-purple-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const filteredPlacements = placements.filter(placement => {
    const matchesSearch = placement.name.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = filterStatus === 'all' || placement.status === filterStatus;
    return matchesSearch && matchesStatus;
  });

  // Calculate stats
  const totalRevenue = placements.reduce((sum, p) => sum + parseFloat(String(p.revenue || 0)), 0);
  const totalImpressions = placements.reduce((sum, p) => sum + (p.impressions || 0), 0);
  const avgFillRate = placements.length > 0
    ? placements.reduce((sum, p) => sum + parseFloat(String(p.fillRate || 0)), 0) / placements.length
    : 0;
  const activePlacements = placements.filter(p => p.status === 'active').length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Placements</h1>
          <p className="text-gray-600 mt-1">Manage where ads appear on your properties</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
        >
          <Plus className="w-4 h-4" />
          Create Placement
        </button>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-emerald-100 rounded-lg">
              <Layers className="w-5 h-5 text-emerald-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Total Placements</p>
              <p className="text-xl font-bold text-gray-900">{placements.length}</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-100 rounded-lg">
              <TrendingUp className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Active</p>
              <p className="text-xl font-bold text-gray-900">{activePlacements}</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-purple-100 rounded-lg">
              <Eye className="w-5 h-5 text-purple-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Total Impressions</p>
              <p className="text-xl font-bold text-gray-900">
                {totalImpressions > 1000000 
                  ? `${(totalImpressions / 1000000).toFixed(1)}M` 
                  : totalImpressions > 1000 
                    ? `${(totalImpressions / 1000).toFixed(1)}K` 
                    : totalImpressions}
              </p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-green-100 rounded-lg">
              <DollarSign className="w-5 h-5 text-green-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Total Revenue</p>
              <p className="text-xl font-bold text-gray-900">${totalRevenue.toLocaleString()}</p>
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
              placeholder="Search placements..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
            />
          </div>
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
          >
            <option value="all">All Status</option>
            <option value="active">Active</option>
            <option value="paused">Paused</option>
          </select>
        </div>
      </div>

      {/* Placements Table */}
      <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-8 h-8 animate-spin text-emerald-600" />
            <span className="ml-2 text-gray-500">Loading placements...</span>
          </div>
        ) : error ? (
          <div className="flex items-center justify-center py-12 text-red-500">
            <AlertCircle className="w-6 h-6 mr-2" />
            {error}
          </div>
        ) : (
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Placement
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Position
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Formats & Sizes
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Devices
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Floor Price
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Revenue
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {filteredPlacements.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-6 py-8 text-center text-gray-500">
                    No placements found. Create your first placement to get started.
                  </td>
                </tr>
              ) : (
              filteredPlacements.map((placement) => (
                <tr key={placement.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4">
                    <div>
                      <p className="font-medium text-gray-900">{placement.name}</p>
                      <p className="text-sm text-gray-500">{placement.domain}</p>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`px-2 py-1 text-xs font-medium rounded-full capitalize ${getPositionColor(placement.position)}`}>
                      {placement.position.replace('_', ' ')}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm">
                      <p className="text-gray-900">{placement.adFormats?.join(', ') || '-'}</p>
                      <p className="text-gray-500">{placement.allowedSizes?.join(', ') || '-'}</p>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1 text-gray-600">
                      {placement.devices?.map((device: string) => (
                        <span key={device} className="p-1 bg-gray-100 rounded" title={device}>
                          {getDeviceIcon(device)}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-6 py-4 font-medium text-gray-900">
                    ${parseFloat(String(placement.floorPrice || 0)).toFixed(2)}
                  </td>
                  <td className="px-6 py-4 font-medium text-emerald-600">
                    ${parseFloat(String(placement.revenue || 0)).toLocaleString()}
                  </td>
                  <td className="px-6 py-4">
                    {placement.status === 'active' ? (
                      <span className="flex items-center gap-1 text-green-600">
                        <CheckCircle className="w-4 h-4" />
                        Active
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-gray-500">
                        <XCircle className="w-4 h-4" />
                        Paused
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center justify-end gap-2">
                      <button className="p-1 text-gray-400 hover:text-gray-600" title="Get Ad Tag">
                        <Code className="w-4 h-4" />
                      </button>
                      <button className="p-1 text-gray-400 hover:text-gray-600" title="View Stats">
                        <BarChart3 className="w-4 h-4" />
                      </button>
                      <button className="p-1 text-gray-400 hover:text-gray-600" title="Edit">
                        <Edit className="w-4 h-4" />
                      </button>
                      <button 
                        onClick={() => handleDeletePlacement(placement.id)}
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
        </div>
        )}
      </div>

      {/* Create Placement Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-lg max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold text-gray-900 mb-4">Create New Placement</h2>
            <form className="space-y-4" onSubmit={handleCreatePlacement}>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Placement Name</label>
                <input
                  type="text"
                  placeholder="e.g., Homepage Hero Banner"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Domain</label>
                <input
                  type="text"
                  placeholder="e.g., example.com"
                  value={formData.domain}
                  onChange={(e) => setFormData({ ...formData, domain: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                  required
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Position</label>
                  <select 
                    value={formData.position}
                    onChange={(e) => setFormData({ ...formData, position: e.target.value })}
                    className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                  >
                    <option value="above_fold">Above Fold</option>
                    <option value="below_fold">Below Fold</option>
                    <option value="in_content">In Content</option>
                    <option value="sidebar">Sidebar</option>
                    <option value="footer">Footer</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Floor Price</label>
                  <div className="relative">
                    <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">$</span>
                    <input
                      type="number"
                      step="0.01"
                      value={formData.floorPrice}
                      onChange={(e) => setFormData({ ...formData, floorPrice: e.target.value })}
                      className="w-full pl-8 pr-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                      required
                    />
                  </div>
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Target Devices</label>
                <div className="flex gap-4">
                  <label className="flex items-center gap-2">
                    <input 
                      type="checkbox" 
                      checked={formData.devices.includes('desktop')}
                      onChange={(e) => {
                        const devices = e.target.checked 
                          ? [...formData.devices, 'desktop'] 
                          : formData.devices.filter(d => d !== 'desktop');
                        setFormData({ ...formData, devices });
                      }}
                      className="rounded" 
                    />
                    <span>Desktop</span>
                  </label>
                  <label className="flex items-center gap-2">
                    <input 
                      type="checkbox" 
                      checked={formData.devices.includes('mobile')}
                      onChange={(e) => {
                        const devices = e.target.checked 
                          ? [...formData.devices, 'mobile'] 
                          : formData.devices.filter(d => d !== 'mobile');
                        setFormData({ ...formData, devices });
                      }}
                      className="rounded" 
                    />
                    <span>Mobile</span>
                  </label>
                  <label className="flex items-center gap-2">
                    <input 
                      type="checkbox" 
                      checked={formData.devices.includes('tablet')}
                      onChange={(e) => {
                        const devices = e.target.checked 
                          ? [...formData.devices, 'tablet'] 
                          : formData.devices.filter(d => d !== 'tablet');
                        setFormData({ ...formData, devices });
                      }}
                      className="rounded" 
                    />
                    <span>Tablet</span>
                  </label>
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
                  Create Placement
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
