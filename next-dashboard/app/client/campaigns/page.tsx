'use client'

import { useState, useEffect } from 'react'
import { api } from '@/lib/api'
import {
  Plus,
  Search,
  Filter,
  MoreVertical,
  Play,
  Pause,
  Copy,
  Trash2,
  Edit,
  Eye,
  TrendingUp,
  TrendingDown,
  Calendar,
  DollarSign,
  Target,
  BarChart3,
  Loader2,
} from 'lucide-react'
import { CreateCampaignModal } from '@/components/campaigns/CreateCampaignModal'

interface Campaign {
  id: string | number;
  name: string;
  status: string;
  type: string;
  budget: number;
  spent: number;
  impressions: number;
  clicks: number;
  conversions: number;
  ctr: number;
  cpc: number;
  roas: number;
  startDate: string;
  endDate: string;
}

const campaignTypes = ['All Types', 'Display Banner', 'Rewarded Video', 'CTV/OTT', 'Native', 'Playable', 'Interstitial']

export default function ClientCampaigns() {
  const [campaigns, setCampaigns] = useState<Campaign[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [typeFilter, setTypeFilter] = useState('All Types')
  const [showCreateModal, setShowCreateModal] = useState(false)

  const fetchCampaigns = async () => {
    setLoading(true)
    try {
      const response = await api.getCampaigns()
      const data = response.data || response || []
      
      // Transform API data to expected format
      const transformed = data.map((c: any) => ({
        id: c.id,
        name: c.name || 'Unnamed Campaign',
        status: c.status || 'draft',
        type: c.type || c.adFormat || 'Display Banner',
        budget: Number(c.budget) || 0,
        spent: Number(c.spent) || Number(c.totalSpent) || 0,
        impressions: Number(c.impressions) || 0,
        clicks: Number(c.clicks) || 0,
        conversions: Number(c.conversions) || 0,
        ctr: Number(c.ctr) || (c.impressions > 0 ? ((c.clicks / c.impressions) * 100) : 0),
        cpc: Number(c.cpc) || (c.clicks > 0 ? (c.spent / c.clicks) : 0),
        roas: Number(c.roas) || 0,
        startDate: c.startDate || c.createdAt || new Date().toISOString(),
        endDate: c.endDate || new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
      }))
      
      setCampaigns(transformed)
    } catch (error) {
      console.error('Failed to fetch campaigns:', error)
      // Fallback to demo data if API fails
      setCampaigns([
        { id: 1, name: 'Summer Sale 2026', status: 'active', type: 'Display Banner', budget: 50000, spent: 32450, impressions: 2400000, clicks: 48000, conversions: 1840, ctr: 2.0, cpc: 0.68, roas: 4.2, startDate: '2026-01-15', endDate: '2026-02-28' },
        { id: 2, name: 'App Install Campaign', status: 'active', type: 'Rewarded Video', budget: 30000, spent: 18320, impressions: 1200000, clicks: 36000, conversions: 2400, ctr: 3.0, cpc: 0.51, roas: 5.8, startDate: '2026-01-20', endDate: '2026-03-15' },
      ])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchCampaigns()
  }, [])

  const filteredCampaigns = campaigns.filter(campaign => {
    const matchesSearch = campaign.name.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesStatus = statusFilter === 'all' || campaign.status === statusFilter
    const matchesType = typeFilter === 'All Types' || campaign.type === typeFilter
    return matchesSearch && matchesStatus && matchesType
  })

  const totalBudget = campaigns.reduce((sum, c) => sum + c.budget, 0)
  const totalSpent = campaigns.reduce((sum, c) => sum + c.spent, 0)
  const totalConversions = campaigns.reduce((sum, c) => sum + c.conversions, 0)
  const avgRoas = campaigns.filter(c => c.roas > 0).length > 0 
    ? campaigns.filter(c => c.roas > 0).reduce((sum, c) => sum + c.roas, 0) / campaigns.filter(c => c.roas > 0).length
    : 0

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
        <span className="ml-2 text-gray-600">Loading campaigns...</span>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Campaigns</h1>
          <p className="text-gray-500">Manage and monitor your advertising campaigns</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
        >
          <Plus className="h-5 w-5" />
          Create Campaign
        </button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-4">
        <div className="rounded-xl bg-white p-5 shadow-sm">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-blue-100 p-2">
              <DollarSign className="h-5 w-5 text-blue-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Total Budget</p>
              <p className="text-xl font-bold text-gray-900">${totalBudget.toLocaleString()}</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-white p-5 shadow-sm">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-green-100 p-2">
              <TrendingUp className="h-5 w-5 text-green-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Total Spent</p>
              <p className="text-xl font-bold text-gray-900">${totalSpent.toLocaleString()}</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-white p-5 shadow-sm">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-purple-100 p-2">
              <Target className="h-5 w-5 text-purple-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Total Conversions</p>
              <p className="text-xl font-bold text-gray-900">{totalConversions.toLocaleString()}</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-white p-5 shadow-sm">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-orange-100 p-2">
              <BarChart3 className="h-5 w-5 text-orange-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Avg ROAS</p>
              <p className="text-xl font-bold text-gray-900">{avgRoas.toFixed(1)}x</p>
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col gap-4 rounded-xl bg-white p-4 shadow-sm sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-1 items-center gap-4">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              placeholder="Search campaigns..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full rounded-lg border border-gray-300 py-2 pl-10 pr-4 focus:border-blue-500 focus:outline-none"
            />
          </div>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
          >
            <option value="all">All Status</option>
            <option value="active">Active</option>
            <option value="paused">Paused</option>
            <option value="draft">Draft</option>
          </select>
          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            className="rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
          >
            {campaignTypes.map(type => (
              <option key={type} value={type}>{type}</option>
            ))}
          </select>
        </div>
      </div>

      {/* Campaigns Table */}
      <div className="rounded-xl bg-white shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200 bg-gray-50">
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Campaign</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Type</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Budget</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Spent</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Impressions</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">CTR</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">ROAS</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {filteredCampaigns.map((campaign) => (
                <tr key={campaign.id} className="hover:bg-gray-50">
                  <td className="whitespace-nowrap px-6 py-4">
                    <div>
                      <p className="font-medium text-gray-900">{campaign.name}</p>
                      <p className="text-xs text-gray-500">{campaign.startDate} - {campaign.endDate}</p>
                    </div>
                  </td>
                  <td className="whitespace-nowrap px-6 py-4">
                    <span className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${
                      campaign.status === 'active' ? 'bg-green-100 text-green-700' :
                      campaign.status === 'paused' ? 'bg-yellow-100 text-yellow-700' :
                      'bg-gray-100 text-gray-700'
                    }`}>
                      {campaign.status}
                    </span>
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-600">{campaign.type}</td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-600">${campaign.budget.toLocaleString()}</td>
                  <td className="whitespace-nowrap px-6 py-4">
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-gray-900">${campaign.spent.toLocaleString()}</span>
                      <div className="h-1.5 w-16 rounded-full bg-gray-200">
                        <div 
                          className="h-1.5 rounded-full bg-blue-500" 
                          style={{ width: `${(campaign.spent / campaign.budget) * 100}%` }}
                        />
                      </div>
                    </div>
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-600">
                    {campaign.impressions > 0 ? (campaign.impressions / 1000000).toFixed(1) + 'M' : '-'}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-600">
                    {campaign.ctr > 0 ? campaign.ctr.toFixed(1) + '%' : '-'}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4">
                    {campaign.roas > 0 ? (
                      <span className={`flex items-center gap-1 text-sm ${campaign.roas >= 3 ? 'text-green-600' : 'text-orange-600'}`}>
                        {campaign.roas >= 3 ? <TrendingUp className="h-4 w-4" /> : <TrendingDown className="h-4 w-4" />}
                        {campaign.roas.toFixed(1)}x
                      </span>
                    ) : '-'}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4">
                    <div className="flex items-center gap-2">
                      <button className="rounded p-1 hover:bg-gray-100" title="View">
                        <Eye className="h-4 w-4 text-gray-500" />
                      </button>
                      <button className="rounded p-1 hover:bg-gray-100" title="Edit">
                        <Edit className="h-4 w-4 text-gray-500" />
                      </button>
                      {campaign.status === 'active' ? (
                        <button className="rounded p-1 hover:bg-gray-100" title="Pause">
                          <Pause className="h-4 w-4 text-yellow-500" />
                        </button>
                      ) : (
                        <button className="rounded p-1 hover:bg-gray-100" title="Start">
                          <Play className="h-4 w-4 text-green-500" />
                        </button>
                      )}
                      <button className="rounded p-1 hover:bg-gray-100" title="Duplicate">
                        <Copy className="h-4 w-4 text-gray-500" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create Campaign Modal */}
      <CreateCampaignModal 
        open={showCreateModal} 
        onOpenChange={setShowCreateModal} 
        onSuccess={fetchCampaigns}
      />
    </div>
  )
}
