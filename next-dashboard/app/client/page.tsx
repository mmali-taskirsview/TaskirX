'use client'

import { useState } from 'react'
import {
  TrendingUp,
  TrendingDown,
  DollarSign,
  Eye,
  MousePointer,
  Target,
  ArrowUpRight,
  ArrowDownRight,
  Calendar,
  RefreshCw,
} from 'lucide-react'

const stats = [
  {
    name: 'Total Spend',
    value: '$45,231',
    change: '+12.5%',
    trend: 'up',
    icon: DollarSign,
    color: 'blue',
  },
  {
    name: 'Impressions',
    value: '2.4M',
    change: '+8.2%',
    trend: 'up',
    icon: Eye,
    color: 'green',
  },
  {
    name: 'Clicks',
    value: '48.2K',
    change: '+15.3%',
    trend: 'up',
    icon: MousePointer,
    color: 'purple',
  },
  {
    name: 'Conversions',
    value: '1,847',
    change: '-2.4%',
    trend: 'down',
    icon: Target,
    color: 'orange',
  },
]

const recentCampaigns = [
  { id: 1, name: 'Summer Sale Banner', status: 'active', spend: '$12,450', impressions: '890K', ctr: '2.4%', conversions: 456 },
  { id: 2, name: 'App Install Video', status: 'active', spend: '$8,320', impressions: '520K', ctr: '3.1%', conversions: 312 },
  { id: 3, name: 'Retargeting Native', status: 'paused', spend: '$5,120', impressions: '340K', ctr: '1.8%', conversions: 189 },
  { id: 4, name: 'Brand Awareness CTV', status: 'active', spend: '$15,780', impressions: '1.2M', ctr: '0.8%', conversions: 234 },
  { id: 5, name: 'Holiday Promo', status: 'draft', spend: '$0', impressions: '0', ctr: '0%', conversions: 0 },
]

const performanceData = [
  { date: 'Mon', impressions: 320000, clicks: 6400, spend: 4200 },
  { date: 'Tue', impressions: 380000, clicks: 7600, spend: 4800 },
  { date: 'Wed', impressions: 420000, clicks: 8400, spend: 5200 },
  { date: 'Thu', impressions: 390000, clicks: 7800, spend: 4900 },
  { date: 'Fri', impressions: 450000, clicks: 9000, spend: 5600 },
  { date: 'Sat', impressions: 280000, clicks: 5600, spend: 3500 },
  { date: 'Sun', impressions: 260000, clicks: 5200, spend: 3200 },
]

export default function ClientDashboard() {
  const [dateRange, setDateRange] = useState('7d')
  const maxImpressions = Math.max(...performanceData.map(d => d.impressions))

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Welcome back, Demo Client</h1>
          <p className="text-gray-500">Here's what's happening with your campaigns today.</p>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={dateRange}
            onChange={(e) => setDateRange(e.target.value)}
            className="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm focus:border-blue-500 focus:outline-none"
          >
            <option value="24h">Last 24 hours</option>
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
            <option value="90d">Last 90 days</option>
          </select>
          <button className="flex items-center gap-2 rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm hover:bg-gray-50">
            <RefreshCw className="h-4 w-4" />
            Refresh
          </button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <div key={stat.name} className="rounded-xl bg-white p-6 shadow-sm">
            <div className="flex items-center justify-between">
              <div className={`rounded-lg p-2 ${
                stat.color === 'blue' ? 'bg-blue-100' :
                stat.color === 'green' ? 'bg-green-100' :
                stat.color === 'purple' ? 'bg-purple-100' :
                'bg-orange-100'
              }`}>
                <stat.icon className={`h-5 w-5 ${
                  stat.color === 'blue' ? 'text-blue-600' :
                  stat.color === 'green' ? 'text-green-600' :
                  stat.color === 'purple' ? 'text-purple-600' :
                  'text-orange-600'
                }`} />
              </div>
              <span className={`flex items-center text-sm font-medium ${
                stat.trend === 'up' ? 'text-green-600' : 'text-red-600'
              }`}>
                {stat.change}
                {stat.trend === 'up' ? (
                  <ArrowUpRight className="ml-1 h-4 w-4" />
                ) : (
                  <ArrowDownRight className="ml-1 h-4 w-4" />
                )}
              </span>
            </div>
            <div className="mt-4">
              <h3 className="text-sm font-medium text-gray-500">{stat.name}</h3>
              <p className="mt-1 text-2xl font-bold text-gray-900">{stat.value}</p>
            </div>
          </div>
        ))}
      </div>

      {/* Performance Chart */}
      <div className="rounded-xl bg-white p-6 shadow-sm">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Performance Overview</h2>
            <p className="text-sm text-gray-500">Daily impressions and spend</p>
          </div>
          <div className="flex gap-4">
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-full bg-blue-500" />
              <span className="text-sm text-gray-600">Impressions</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-full bg-green-500" />
              <span className="text-sm text-gray-600">Spend</span>
            </div>
          </div>
        </div>
        <div className="flex h-64 items-end justify-between gap-2">
          {performanceData.map((day, index) => (
            <div key={day.date} className="flex flex-1 flex-col items-center gap-2">
              <div className="flex w-full flex-col items-center gap-1">
                <div
                  className="w-full max-w-[40px] rounded-t-lg bg-blue-500"
                  style={{ height: `${(day.impressions / maxImpressions) * 180}px` }}
                />
                <div
                  className="w-full max-w-[40px] rounded-lg bg-green-500"
                  style={{ height: `${(day.spend / 6000) * 60}px` }}
                />
              </div>
              <span className="text-xs text-gray-500">{day.date}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Recent Campaigns */}
      <div className="rounded-xl bg-white shadow-sm">
        <div className="border-b border-gray-200 p-6">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-gray-900">Recent Campaigns</h2>
              <p className="text-sm text-gray-500">Your active and recent campaigns</p>
            </div>
            <a href="/client/campaigns" className="text-sm font-medium text-blue-600 hover:text-blue-700">
              View all →
            </a>
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200 bg-gray-50">
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Campaign</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Spend</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Impressions</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">CTR</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Conversions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {recentCampaigns.map((campaign) => (
                <tr key={campaign.id} className="hover:bg-gray-50">
                  <td className="whitespace-nowrap px-6 py-4">
                    <span className="font-medium text-gray-900">{campaign.name}</span>
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
                  <td className="whitespace-nowrap px-6 py-4 text-gray-600">{campaign.spend}</td>
                  <td className="whitespace-nowrap px-6 py-4 text-gray-600">{campaign.impressions}</td>
                  <td className="whitespace-nowrap px-6 py-4 text-gray-600">{campaign.ctr}</td>
                  <td className="whitespace-nowrap px-6 py-4 text-gray-600">{campaign.conversions.toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="grid gap-4 sm:grid-cols-3">
        <a href="/client/campaigns" className="flex items-center gap-4 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 p-6 text-white shadow-sm transition-transform hover:scale-[1.02]">
          <div className="rounded-lg bg-white/20 p-3">
            <TrendingUp className="h-6 w-6" />
          </div>
          <div>
            <h3 className="font-semibold">Create Campaign</h3>
            <p className="text-sm text-blue-100">Launch a new ad campaign</p>
          </div>
        </a>
        <a href="/client/analytics" className="flex items-center gap-4 rounded-xl bg-gradient-to-br from-purple-500 to-purple-600 p-6 text-white shadow-sm transition-transform hover:scale-[1.02]">
          <div className="rounded-lg bg-white/20 p-3">
            <Eye className="h-6 w-6" />
          </div>
          <div>
            <h3 className="font-semibold">View Analytics</h3>
            <p className="text-sm text-purple-100">Deep dive into performance</p>
          </div>
        </a>
        <a href="/client/creatives" className="flex items-center gap-4 rounded-xl bg-gradient-to-br from-green-500 to-green-600 p-6 text-white shadow-sm transition-transform hover:scale-[1.02]">
          <div className="rounded-lg bg-white/20 p-3">
            <Target className="h-6 w-6" />
          </div>
          <div>
            <h3 className="font-semibold">Upload Creatives</h3>
            <p className="text-sm text-green-100">Add new ad creatives</p>
          </div>
        </a>
      </div>
    </div>
  )
}
