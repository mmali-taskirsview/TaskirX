'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Plus,
  Search,
  MoreVertical,
  Edit,
  Trash2,
  Play,
  Pause,
  CheckCircle,
  XCircle,
  DollarSign,
  Target,
  TrendingUp,
  Eye,
  MousePointer,
  Users
} from 'lucide-react';
import { api } from '@/lib/api';
import { formatCurrency, formatNumber, formatPercentage } from '@/lib/utils';
import { CreateCampaignModal, EditCampaignModal } from '@/components/campaigns/CreateCampaignModal';

type Campaign = {
  id: string;
  name: string;
  description?: string;
  status: 'draft' | 'active' | 'paused' | 'completed';
  type: 'cpm' | 'cpc' | 'cpa';
  budget: number;
  spent: number;
  bidPrice: number;
  vertical?: string;
  createdAt: string;
  updatedAt: string;
  // Real-time data
  impressions?: number;
  clicks?: number;
  conversions?: number;
  ctr?: number;
  cpa?: number;
  roas?: number;
};

export default function CampaignsPage() {
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [editingCampaignId, setEditingCampaignId] = useState<string>('');
  const [selectedCampaigns, setSelectedCampaigns] = useState<Set<string>>(new Set());

  useEffect(() => {
    loadCampaigns();
  }, []);

  const loadCampaigns = async () => {
    try {
      setLoading(true);
      const response = await api.getCampaigns();
      setCampaigns(response.data);
    } catch (error) {
      console.error('Failed to load campaigns:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleStatusChange = async (campaignId: string, newStatus: string) => {
    try {
      await api.updateCampaign(campaignId, { status: newStatus });
      await loadCampaigns(); // Refresh the list
    } catch (error) {
      console.error('Failed to update campaign status:', error);
    }
  };

  const handleDeleteCampaign = async (campaignId: string) => {
    if (!confirm('Are you sure you want to delete this campaign?')) return;

    try {
      await api.deleteCampaign(campaignId);
      await loadCampaigns(); // Refresh the list
    } catch (error) {
      console.error('Failed to delete campaign:', error);
    }
  };

  const handleEditCampaign = (campaignId: string) => {
    setEditingCampaignId(campaignId);
    setEditModalOpen(true);
  };

  const handleSelectCampaign = (campaignId: string, selected: boolean) => {
    const newSelected = new Set(selectedCampaigns);
    if (selected) {
      newSelected.add(campaignId);
    } else {
      newSelected.delete(campaignId);
    }
    setSelectedCampaigns(newSelected);
  };

  const handleSelectAll = (selected: boolean) => {
    if (selected) {
      setSelectedCampaigns(new Set(filteredCampaigns.map(c => c.id)));
    } else {
      setSelectedCampaigns(new Set());
    }
  };

  const handleBulkStatusChange = async (newStatus: string) => {
    if (selectedCampaigns.size === 0) return;

    const action = newStatus === 'active' ? 'activate' : newStatus === 'paused' ? 'pause' : 'complete';
    if (!confirm(`Are you sure you want to ${action} ${selectedCampaigns.size} campaign(s)?`)) return;

    try {
      const promises = Array.from(selectedCampaigns).map(campaignId =>
        api.updateCampaign(campaignId, { status: newStatus })
      );
      await Promise.all(promises);
      setSelectedCampaigns(new Set());
      await loadCampaigns();
    } catch (error) {
      console.error('Failed to update campaign statuses:', error);
      alert('Failed to update some campaigns. Please try again.');
    }
  };

  const handleBulkDelete = async () => {
    if (selectedCampaigns.size === 0) return;

    if (!confirm(`Are you sure you want to delete ${selectedCampaigns.size} campaign(s)? This action cannot be undone.`)) return;

    try {
      const promises = Array.from(selectedCampaigns).map(campaignId =>
        api.deleteCampaign(campaignId)
      );
      await Promise.all(promises);
      setSelectedCampaigns(new Set());
      await loadCampaigns();
    } catch (error) {
      console.error('Failed to delete campaigns:', error);
      alert('Failed to delete some campaigns. Please try again.');
    }
  };

  const filteredCampaigns = campaigns.filter(campaign => {
    const matchesSearch = campaign.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         (campaign.description?.toLowerCase().includes(searchTerm.toLowerCase()));
    const matchesStatus = statusFilter === 'all' || campaign.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'text-green-600 bg-green-100';
      case 'paused': return 'text-yellow-600 bg-yellow-100';
      case 'completed': return 'text-blue-600 bg-blue-100';
      case 'draft': return 'text-gray-600 bg-gray-100';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'cpm': return 'CPM';
      case 'cpc': return 'CPC';
      case 'cpa': return 'CPA';
      default: return type.toUpperCase();
    }
  };

  if (loading) {
    return (
      <div className="p-6 max-w-7xl mx-auto">
        <div className="flex justify-center items-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Campaigns</h1>
          <p className="text-gray-600">Manage your advertising campaigns</p>
        </div>
        <Button className="flex items-center gap-2" onClick={() => setCreateModalOpen(true)}>
          <Plus className="h-4 w-4" />
          Create Campaign
        </Button>
      </div>

      {/* Filters */}
      <div className="flex gap-4 mb-6">
        <div className="flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            <input
              type="text"
              placeholder="Search campaigns..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
        </div>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        >
          <option value="all">All Status</option>
          <option value="active">Active</option>
          <option value="paused">Paused</option>
          <option value="draft">Draft</option>
          <option value="completed">Completed</option>
        </select>
      </div>

      {/* Bulk Actions */}
      {selectedCampaigns.size > 0 && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <span className="text-sm font-medium text-blue-900">
                {selectedCampaigns.size} campaign{selectedCampaigns.size !== 1 ? 's' : ''} selected
              </span>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleBulkStatusChange('active')}
                  className="text-green-600 border-green-300 hover:bg-green-50"
                >
                  <Play className="h-4 w-4 mr-1" />
                  Activate
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleBulkStatusChange('paused')}
                  className="text-yellow-600 border-yellow-300 hover:bg-yellow-50"
                >
                  <Pause className="h-4 w-4 mr-1" />
                  Pause
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleBulkDelete}
                  className="text-red-600 border-red-300 hover:bg-red-50"
                >
                  <Trash2 className="h-4 w-4 mr-1" />
                  Delete
                </Button>
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setSelectedCampaigns(new Set())}
              className="text-blue-600 hover:text-blue-800"
            >
              Clear Selection
            </Button>
          </div>
        </div>
      )}

      {/* Campaigns Table */}
      <Card>
        <CardHeader>
          <CardTitle>Your Campaigns</CardTitle>
          <CardDescription>
            {filteredCampaigns.length} campaign{filteredCampaigns.length !== 1 ? 's' : ''} found
          </CardDescription>
        </CardHeader>
        <CardContent>
          {filteredCampaigns.length === 0 ? (
            <div className="text-center py-12">
              <Target className="mx-auto h-12 w-12 text-gray-400" />
              <h3 className="mt-2 text-sm font-medium text-gray-900">No campaigns found</h3>
              <p className="mt-1 text-sm text-gray-500">
                {searchTerm || statusFilter !== 'all'
                  ? 'Try adjusting your search or filters.'
                  : 'Get started by creating your first campaign.'}
              </p>
              <div className="mt-6">
                <Button>
                  <Plus className="h-4 w-4 mr-2" />
                  Create Campaign
                </Button>
              </div>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      <input
                        type="checkbox"
                        checked={selectedCampaigns.size === filteredCampaigns.length && filteredCampaigns.length > 0}
                        onChange={(e) => handleSelectAll(e.target.checked)}
                        className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                      />
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Campaign
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Type
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Budget
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Performance
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {filteredCampaigns.map((campaign) => (
                    <tr key={campaign.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <input
                          type="checkbox"
                          checked={selectedCampaigns.has(campaign.id)}
                          onChange={(e) => handleSelectCampaign(campaign.id, e.target.checked)}
                          className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                        />
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div>
                          <div className="text-sm font-medium text-gray-900">{campaign.name}</div>
                          <div className="text-sm text-gray-500">{campaign.description}</div>
                          <div className="text-xs text-gray-400">{campaign.vertical}</div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(campaign.status)}`}>
                          {campaign.status.charAt(0).toUpperCase() + campaign.status.slice(1)}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {getTypeLabel(campaign.type)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-gray-900">{formatCurrency(campaign.budget)}</div>
                        <div className="text-xs text-gray-500">
                          Spent: {formatCurrency(campaign.spent || 0)}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center space-x-4 text-sm">
                          {campaign.impressions !== undefined && (
                            <div className="flex items-center">
                              <Eye className="h-4 w-4 text-gray-400 mr-1" />
                              {formatNumber(campaign.impressions)}
                            </div>
                          )}
                          {campaign.clicks !== undefined && (
                            <div className="flex items-center">
                              <MousePointer className="h-4 w-4 text-gray-400 mr-1" />
                              {formatNumber(campaign.clicks)}
                            </div>
                          )}
                          {campaign.ctr !== undefined && (
                            <div className="text-gray-500">
                              {formatPercentage(campaign.ctr)}
                            </div>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                        <div className="flex items-center space-x-2">
                          {campaign.status === 'active' ? (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleStatusChange(campaign.id, 'paused')}
                              className="text-yellow-600 border-yellow-300 hover:bg-yellow-50"
                            >
                              <Pause className="h-4 w-4" />
                            </Button>
                          ) : campaign.status === 'paused' || campaign.status === 'draft' ? (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleStatusChange(campaign.id, 'active')}
                              className="text-green-600 border-green-300 hover:bg-green-50"
                            >
                              <Play className="h-4 w-4" />
                            </Button>
                          ) : null}

                          <Button variant="outline" size="sm" onClick={() => handleEditCampaign(campaign.id)}>
                            <Edit className="h-4 w-4" />
                          </Button>

                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleDeleteCampaign(campaign.id)}
                            className="text-red-600 border-red-300 hover:bg-red-50"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      <CreateCampaignModal
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
        onSuccess={loadCampaigns}
      />

      <EditCampaignModal
        open={editModalOpen}
        onOpenChange={setEditModalOpen}
        onSuccess={loadCampaigns}
        campaignId={editingCampaignId}
      />
    </div>
  );
}
