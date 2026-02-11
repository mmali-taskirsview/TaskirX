'use client';

import React, { useState, useEffect } from 'react';
import {
  Zap,
  Plus,
  Search,
  CheckCircle,
  XCircle,
  Clock,
  TrendingUp,
  DollarSign,
  Settings,
  ExternalLink,
  RefreshCw,
  BarChart3,
  Globe,
  Layers,
  Loader2,
  AlertCircle
} from 'lucide-react';
import { api } from '@/lib/api';

// Types
interface DemandPartner {
  id: string;
  name: string;
  code: string;
  type: string;
  status: string;
  publisherId?: string;
  endpoint: string;
  credentials?: any;
  settings?: any;
  revenueShare: number;
  bidTimeout: number;
  isGlobal: boolean;
  totalImpressions: number;
  totalRevenue: number;
  winRate: number;
  avgBidPrice: number;
  createdAt: string;
  updatedAt: string;
}

const headerBiddingConfig = {
  enabled: true,
  timeout: 1000,
  priceGranularity: 'dense',
  currency: 'USD',
  debug: false
};

export default function DemandPartnersPage() {
  const [demandPartners, setDemandPartners] = useState<DemandPartner[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('partners');
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState('all');
  const [showAddPartnerModal, setShowAddPartnerModal] = useState(false);

  // Fetch demand partners from API
  useEffect(() => {
    fetchDemandPartners();
  }, []);

  const fetchDemandPartners = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.getDemandPartners();
      setDemandPartners(response.data);
    } catch (err: any) {
      console.error('Failed to fetch demand partners:', err);
      setError(err.message || 'Failed to load demand partners');
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return (
          <span className="flex items-center gap-1 px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded-full">
            <CheckCircle className="w-3 h-3" />
            Active
          </span>
        );
      case 'paused':
        return (
          <span className="flex items-center gap-1 px-2 py-1 text-xs font-medium bg-gray-100 text-gray-600 rounded-full">
            <XCircle className="w-3 h-3" />
            Paused
          </span>
        );
      case 'testing':
        return (
          <span className="flex items-center gap-1 px-2 py-1 text-xs font-medium bg-yellow-100 text-yellow-700 rounded-full">
            <Clock className="w-3 h-3" />
            Testing
          </span>
        );
      default:
        return null;
    }
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'exchange': return 'bg-blue-100 text-blue-800';
      case 'ssp': return 'bg-purple-100 text-purple-800';
      case 'dsp': return 'bg-green-100 text-green-800';
      case 'header_bidding': return 'bg-amber-100 text-amber-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const filteredPartners = demandPartners.filter(partner => {
    const matchesSearch = partner.name.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = filterStatus === 'all' || partner.status === filterStatus;
    return matchesSearch && matchesStatus;
  });

  const totalRevenue = demandPartners.reduce((sum, p) => sum + parseFloat(String(p.totalRevenue || 0)), 0);
  const activePartners = demandPartners.filter(p => p.status === 'active').length;
  const avgWinRate = demandPartners.length > 0 
    ? demandPartners.reduce((sum, p) => sum + parseFloat(String(p.winRate || 0)), 0) / demandPartners.length 
    : 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Demand Partners</h1>
          <p className="text-gray-600 mt-1">Manage your programmatic demand sources and header bidding</p>
        </div>
        <button
          onClick={() => setShowAddPartnerModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
        >
          <Plus className="w-4 h-4" />
          Add Partner
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-emerald-100 rounded-lg">
              <Zap className="w-5 h-5 text-emerald-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Active Partners</p>
              <p className="text-xl font-bold text-gray-900">{activePartners}</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-100 rounded-lg">
              <DollarSign className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Partner Revenue</p>
              <p className="text-xl font-bold text-gray-900">${(totalRevenue / 1000).toFixed(1)}K</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-purple-100 rounded-lg">
              <TrendingUp className="w-5 h-5 text-purple-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Avg Bid Rate</p>
              <p className="text-xl font-bold text-gray-900">77.6%</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-orange-100 rounded-lg">
              <Clock className="w-5 h-5 text-orange-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Avg Latency</p>
              <p className="text-xl font-bold text-gray-900">53ms</p>
            </div>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="flex gap-8">
          {['partners', 'header-bidding', 'performance'].map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`pb-4 px-1 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab
                  ? 'border-emerald-500 text-emerald-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              {tab.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')}
            </button>
          ))}
        </nav>
      </div>

      {/* Partners Tab */}
      {activeTab === 'partners' && (
        <>
          {/* Filters */}
          <div className="bg-white rounded-xl p-4 border border-gray-200">
            <div className="flex flex-wrap gap-4 items-center">
              <div className="relative flex-1 min-w-[200px]">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                <input
                  type="text"
                  placeholder="Search partners..."
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
                <option value="testing">Testing</option>
              </select>
            </div>
          </div>

          {/* Partners Table */}
          <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Partner</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Bid Rate</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Win Rate</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Avg Bid</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Revenue</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Latency</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {loading ? (
                    <tr>
                      <td colSpan={9} className="px-4 py-8 text-center">
                        <div className="flex items-center justify-center">
                          <Loader2 className="w-6 h-6 animate-spin text-emerald-600 mr-2" />
                          Loading partners...
                        </div>
                      </td>
                    </tr>
                  ) : error ? (
                    <tr>
                      <td colSpan={9} className="px-4 py-8 text-center text-red-500">
                        <AlertCircle className="w-6 h-6 inline mr-2" />
                        {error}
                      </td>
                    </tr>
                  ) : filteredPartners.length === 0 ? (
                    <tr>
                      <td colSpan={9} className="px-4 py-8 text-center text-gray-500">
                        No demand partners found
                      </td>
                    </tr>
                  ) : (
                  filteredPartners.map((partner) => (
                    <tr key={partner.id} className="hover:bg-gray-50">
                      <td className="px-4 py-4">
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 bg-gray-100 rounded-lg flex items-center justify-center">
                            <Globe className="w-4 h-4 text-gray-600" />
                          </div>
                          <div>
                            <span className="font-medium text-gray-900">{partner.name}</span>
                            <p className="text-xs text-gray-500">{partner.code}</p>
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-4">
                        <span className={`px-2 py-1 text-xs font-medium rounded-full capitalize ${getTypeColor(partner.type)}`}>
                          {partner.type.replace('_', ' ')}
                        </span>
                      </td>
                      <td className="px-4 py-4">{getStatusBadge(partner.status)}</td>
                      <td className="px-4 py-4 text-right text-sm text-gray-600">
                        {partner.revenueShare > 0 ? `${partner.revenueShare}%` : '-'}
                      </td>
                      <td className="px-4 py-4 text-right text-sm text-gray-600">
                        {parseFloat(String(partner.winRate)) > 0 ? `${parseFloat(String(partner.winRate)).toFixed(1)}%` : '-'}
                      </td>
                      <td className="px-4 py-4 text-right text-sm text-gray-600">
                        {parseFloat(String(partner.avgBidPrice)) > 0 ? `$${parseFloat(String(partner.avgBidPrice)).toFixed(2)}` : '-'}
                      </td>
                      <td className="px-4 py-4 text-right text-sm font-medium text-emerald-600">
                        {parseFloat(String(partner.totalRevenue)) > 0 ? `$${parseFloat(String(partner.totalRevenue)).toLocaleString()}` : '-'}
                      </td>
                      <td className="px-4 py-4 text-right text-sm text-gray-600">
                        {partner.bidTimeout > 0 ? `${partner.bidTimeout}ms` : '-'}
                      </td>
                      <td className="px-4 py-4 text-right">
                        <div className="flex items-center justify-end gap-2">
                          <button className="p-1 text-gray-400 hover:text-gray-600">
                            <BarChart3 className="w-4 h-4" />
                          </button>
                          <button className="p-1 text-gray-400 hover:text-gray-600">
                            <Settings className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}

      {/* Header Bidding Tab */}
      {activeTab === 'header-bidding' && (
        <div className="space-y-6">
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h3 className="font-semibold text-gray-900">Header Bidding Configuration</h3>
                <p className="text-sm text-gray-500 mt-1">Prebid.js settings for client-side auctions</p>
              </div>
              <div className="flex items-center gap-2">
                <span className={`px-3 py-1 text-sm font-medium rounded-full ${headerBiddingConfig.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'}`}>
                  {headerBiddingConfig.enabled ? 'Enabled' : 'Disabled'}
                </span>
              </div>
            </div>
            
            <div className="grid grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Auction Timeout (ms)</label>
                <input
                  type="number"
                  defaultValue={headerBiddingConfig.timeout}
                  className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Price Granularity</label>
                <select
                  defaultValue={headerBiddingConfig.priceGranularity}
                  className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                >
                  <option value="low">Low ($0.50 increments)</option>
                  <option value="medium">Medium ($0.10 increments)</option>
                  <option value="high">High ($0.01 increments)</option>
                  <option value="dense">Dense (Variable)</option>
                  <option value="auto">Auto</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Currency</label>
                <select
                  defaultValue={headerBiddingConfig.currency}
                  className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                >
                  <option value="USD">USD</option>
                  <option value="EUR">EUR</option>
                  <option value="GBP">GBP</option>
                </select>
              </div>
              <div className="flex items-center gap-4">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" defaultChecked={headerBiddingConfig.debug} className="rounded" />
                  <span className="text-sm text-gray-700">Enable Debug Mode</span>
                </label>
              </div>
            </div>

            <div className="mt-6 pt-6 border-t border-gray-100">
              <button className="px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700">
                Save Configuration
              </button>
            </div>
          </div>

          {/* Prebid.js Code */}
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold text-gray-900">Prebid.js Integration Code</h3>
              <button className="flex items-center gap-2 text-sm text-emerald-600 hover:text-emerald-700">
                <ExternalLink className="w-4 h-4" />
                View Documentation
              </button>
            </div>
            <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm">
              <code>{`<script async src="https://cdn.taskirx.com/prebid/v8.js"></script>
<script>
  var pbjs = pbjs || {};
  pbjs.que = pbjs.que || [];
  
  pbjs.que.push(function() {
    pbjs.setConfig({
      priceGranularity: 'dense',
      currency: { adServerCurrency: 'USD' },
      bidderTimeout: 1000
    });
    
    pbjs.addAdUnits(adUnits);
    pbjs.requestBids({ bidsBackHandler: sendAdServerRequest });
  });
</script>`}</code>
            </pre>
          </div>
        </div>
      )}

      {/* Performance Tab */}
      {activeTab === 'performance' && (
        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <h3 className="font-semibold text-gray-900 mb-4">Partner Performance Comparison</h3>
          <div className="space-y-4">
            {demandPartners.filter(p => p.status === 'active').map((partner) => {
              const maxRevenue = Math.max(...demandPartners.map(p => parseFloat(String(p.totalRevenue || 0))), 1);
              const partnerRevenue = parseFloat(String(partner.totalRevenue || 0));
              return (
              <div key={partner.id} className="flex items-center gap-4">
                <div className="w-32 text-sm font-medium text-gray-900">{partner.name}</div>
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <div className="flex-1 bg-gray-200 rounded-full h-4 overflow-hidden">
                      <div
                        className="bg-emerald-500 h-4 rounded-full"
                        style={{ width: `${(partnerRevenue / maxRevenue) * 100}%` }}
                      />
                    </div>
                    <span className="text-sm text-gray-600 w-20 text-right">${partnerRevenue.toLocaleString()}</span>
                  </div>
                </div>
              </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Add Partner Modal */}
      {showAddPartnerModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-lg">
            <h2 className="text-xl font-bold text-gray-900 mb-4">Add Demand Partner</h2>
            <form className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Partner Name</label>
                <input
                  type="text"
                  placeholder="e.g., Google AdX"
                  className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Partner Type</label>
                <select className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500">
                  <option value="exchange">Ad Exchange</option>
                  <option value="ssp">SSP</option>
                  <option value="dsp">DSP</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">API Endpoint</label>
                <input
                  type="text"
                  placeholder="https://..."
                  className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                />
              </div>
              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowAddPartnerModal(false)}
                  className="flex-1 px-4 py-2 border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700"
                >
                  Add Partner
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
