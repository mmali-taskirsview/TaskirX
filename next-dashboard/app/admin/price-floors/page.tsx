'use client'

import { useState } from 'react'
import {
  DollarSign,
  TrendingUp,
  Plus,
  Edit,
  Trash2,
  ToggleLeft,
  ToggleRight,
  AlertCircle,
  Globe,
  Smartphone,
  Monitor,
  Tv,
  Play,
  Image,
} from 'lucide-react'

const floorRules = [
  {
    id: 1,
    name: 'Premium Display SEA',
    type: 'Display Banner',
    regions: ['Indonesia', 'Thailand', 'Vietnam'],
    device: 'All',
    floor: 3.50,
    status: 'active',
    performance: { fillRate: 82.4, revenue: 145000, change: '+12%' },
  },
  {
    id: 2,
    name: 'Rewarded Video Gaming',
    type: 'Rewarded Video',
    regions: ['All'],
    device: 'Mobile',
    floor: 8.00,
    status: 'active',
    performance: { fillRate: 75.2, revenue: 234000, change: '+28%' },
  },
  {
    id: 3,
    name: 'CTV Premium Inventory',
    type: 'CTV/OTT',
    regions: ['Singapore', 'Malaysia'],
    device: 'CTV',
    floor: 25.00,
    status: 'active',
    performance: { fillRate: 68.5, revenue: 89000, change: '+45%' },
  },
  {
    id: 4,
    name: 'Native Ads Standard',
    type: 'Native',
    regions: ['All'],
    device: 'All',
    floor: 2.20,
    status: 'active',
    performance: { fillRate: 88.1, revenue: 112000, change: '+8%' },
  },
  {
    id: 5,
    name: 'Interstitial Default',
    type: 'Interstitial',
    regions: ['All'],
    device: 'Mobile',
    floor: 4.50,
    status: 'paused',
    performance: { fillRate: 71.3, revenue: 67000, change: '-3%' },
  },
]

const formatIcons: Record<string, any> = {
  'Display Banner': Image,
  'Rewarded Video': Play,
  'CTV/OTT': Tv,
  'Native': Globe,
  'Interstitial': Smartphone,
  'Playable': Monitor,
}

export default function AdminPriceFloors() {
  const [showCreate, setShowCreate] = useState(false)
  const [selectedRule, setSelectedRule] = useState<number | null>(null)

  const toggleRule = (id: number) => {
    // Toggle logic here
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Price Floors</h1>
          <p className="text-gray-400">Configure and optimize floor prices for yield management</p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-purple-600 px-4 py-2 text-white hover:bg-purple-700"
        >
          <Plus className="h-5 w-5" />
          Create Rule
        </button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-4">
        <div className="rounded-xl bg-gray-800 p-5">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-purple-900/50 p-2">
              <DollarSign className="h-5 w-5 text-purple-400" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Avg. Floor CPM</p>
              <p className="text-2xl font-bold text-white">$8.64</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-gray-800 p-5">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-green-900/50 p-2">
              <TrendingUp className="h-5 w-5 text-green-400" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Revenue Lift</p>
              <p className="text-2xl font-bold text-white">+18.4%</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-gray-800 p-5">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-blue-900/50 p-2">
              <Globe className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Active Rules</p>
              <p className="text-2xl font-bold text-white">{floorRules.filter(r => r.status === 'active').length}</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-gray-800 p-5">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-orange-900/50 p-2">
              <AlertCircle className="h-5 w-5 text-orange-400" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Avg. Fill Rate</p>
              <p className="text-2xl font-bold text-white">77.1%</p>
            </div>
          </div>
        </div>
      </div>

      {/* AI Recommendation */}
      <div className="rounded-xl bg-gradient-to-r from-purple-900/50 to-blue-900/50 border border-purple-700/50 p-6">
        <div className="flex items-start gap-4">
          <div className="rounded-lg bg-purple-500/20 p-3">
            <TrendingUp className="h-6 w-6 text-purple-400" />
          </div>
          <div>
            <h3 className="font-semibold text-white">AI Optimization Suggestion</h3>
            <p className="mt-1 text-sm text-gray-300">
              Based on current market conditions, increasing CTV floor prices by 15% in Singapore could improve revenue by ~$12K/month 
              with minimal impact on fill rate. Consider A/B testing this change.
            </p>
            <div className="mt-3 flex gap-3">
              <button className="rounded-lg bg-purple-600 px-4 py-2 text-sm font-medium text-white hover:bg-purple-700">
                Apply Suggestion
              </button>
              <button className="rounded-lg border border-gray-600 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-800">
                A/B Test First
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Floor Rules Table */}
      <div className="rounded-xl bg-gray-800">
        <div className="border-b border-gray-700 p-6">
          <h2 className="text-lg font-semibold text-white">Floor Price Rules</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-700">
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Rule</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Format</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Regions</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Device</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Floor CPM</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Fill Rate</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Revenue</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Status</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-700">
              {floorRules.map((rule) => {
                const Icon = formatIcons[rule.type] || Globe
                return (
                  <tr key={rule.id} className="hover:bg-gray-700/50">
                    <td className="px-6 py-4">
                      <span className="font-medium text-white">{rule.name}</span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <Icon className="h-4 w-4 text-gray-400" />
                        <span className="text-gray-300">{rule.type}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span className="text-sm text-gray-400">
                        {rule.regions.length > 2 ? `${rule.regions[0]} +${rule.regions.length - 1}` : rule.regions.join(', ')}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-gray-400">{rule.device}</td>
                    <td className="px-6 py-4">
                      <span className="font-semibold text-green-400">${rule.floor.toFixed(2)}</span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <div className="h-1.5 w-12 rounded-full bg-gray-700">
                          <div 
                            className={`h-1.5 rounded-full ${rule.performance.fillRate >= 80 ? 'bg-green-500' : rule.performance.fillRate >= 70 ? 'bg-yellow-500' : 'bg-red-500'}`}
                            style={{ width: `${rule.performance.fillRate}%` }}
                          />
                        </div>
                        <span className="text-sm text-gray-400">{rule.performance.fillRate}%</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div>
                        <span className="font-medium text-white">${(rule.performance.revenue / 1000).toFixed(0)}K</span>
                        <span className={`ml-2 text-xs ${rule.performance.change.startsWith('+') ? 'text-green-400' : 'text-red-400'}`}>
                          {rule.performance.change}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <button
                        onClick={() => toggleRule(rule.id)}
                        className={`rounded-full px-3 py-1 text-xs font-medium ${
                          rule.status === 'active' 
                            ? 'bg-green-900/50 text-green-400' 
                            : 'bg-gray-700 text-gray-400'
                        }`}
                      >
                        {rule.status}
                      </button>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <button className="rounded p-1 text-gray-400 hover:bg-gray-700 hover:text-white">
                          <Edit className="h-4 w-4" />
                        </button>
                        <button className="rounded p-1 text-gray-400 hover:bg-gray-700 hover:text-red-400">
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>

      {/* Floor Optimization Tips */}
      <div className="rounded-xl bg-gray-800 p-6">
        <h3 className="font-semibold text-white mb-4">Optimization Guidelines</h3>
        <div className="grid gap-4 sm:grid-cols-3">
          <div className="rounded-lg bg-gray-900 p-4">
            <h4 className="font-medium text-purple-400">Premium Inventory</h4>
            <p className="mt-2 text-sm text-gray-400">
              Set higher floors ($8-15 CPM) for premium placements like above-fold and first-view positions.
            </p>
          </div>
          <div className="rounded-lg bg-gray-900 p-4">
            <h4 className="font-medium text-blue-400">Video Formats</h4>
            <p className="mt-2 text-sm text-gray-400">
              Video ads command 3-5x higher CPMs. Set floors at $15+ for CTV and $8+ for mobile video.
            </p>
          </div>
          <div className="rounded-lg bg-gray-900 p-4">
            <h4 className="font-medium text-green-400">Regional Pricing</h4>
            <p className="mt-2 text-sm text-gray-400">
              Singapore/Malaysia can support 2x higher floors than Indonesia/Vietnam due to advertiser demand.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
