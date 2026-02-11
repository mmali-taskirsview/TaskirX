'use client'

import { useState } from 'react'
import {
  DollarSign,
  TrendingUp,
  TrendingDown,
  BarChart3,
  PieChart,
  Download,
  Calendar,
  Filter,
  ArrowUpRight,
} from 'lucide-react'

const revenueStats = [
  { name: 'Total Revenue', value: '$2.42M', change: '+18.2%', trend: 'up' },
  { name: 'Platform Fee', value: '$363K', change: '+15.4%', trend: 'up' },
  { name: 'Net Payout', value: '$2.06M', change: '+19.1%', trend: 'up' },
  { name: 'Avg. eCPM', value: '$4.82', change: '+8.3%', trend: 'up' },
]

const revenueByFormat = [
  { format: 'Display Banner', revenue: 680000, share: 28, change: '+12%' },
  { format: 'Rewarded Video', revenue: 580000, share: 24, change: '+22%' },
  { format: 'Native Ads', revenue: 460000, share: 19, change: '+8%' },
  { format: 'CTV/OTT', revenue: 380000, share: 16, change: '+45%' },
  { format: 'Interstitial', revenue: 220000, share: 9, change: '+5%' },
  { format: 'Playable', revenue: 100000, share: 4, change: '+18%' },
]

const revenueByVertical = [
  { vertical: 'Gaming', revenue: 720000, clients: 45, avgEcpm: 5.20 },
  { vertical: 'E-commerce', revenue: 580000, clients: 38, avgEcpm: 4.50 },
  { vertical: 'Finance', revenue: 420000, clients: 22, avgEcpm: 8.40 },
  { vertical: 'Entertainment', revenue: 340000, clients: 28, avgEcpm: 3.80 },
  { vertical: 'Travel', revenue: 220000, clients: 15, avgEcpm: 4.20 },
  { vertical: 'Others', revenue: 140000, clients: 12, avgEcpm: 3.50 },
]

const dailyRevenue = [
  { date: 'Mon', revenue: 320000, impressions: 680000000 },
  { date: 'Tue', revenue: 380000, impressions: 780000000 },
  { date: 'Wed', revenue: 420000, impressions: 850000000 },
  { date: 'Thu', revenue: 390000, impressions: 810000000 },
  { date: 'Fri', revenue: 450000, impressions: 920000000 },
  { date: 'Sat', revenue: 280000, impressions: 580000000 },
  { date: 'Sun', revenue: 260000, impressions: 540000000 },
]

export default function AdminRevenue() {
  const [dateRange, setDateRange] = useState('mtd')
  const maxRevenue = Math.max(...dailyRevenue.map(d => d.revenue))

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Revenue Dashboard</h1>
          <p className="text-gray-400">Platform revenue breakdown and analytics</p>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={dateRange}
            onChange={(e) => setDateRange(e.target.value)}
            className="rounded-lg border border-gray-700 bg-gray-800 px-4 py-2 text-sm text-white focus:border-purple-500 focus:outline-none"
          >
            <option value="today">Today</option>
            <option value="7d">Last 7 days</option>
            <option value="mtd">Month to Date</option>
            <option value="qtd">Quarter to Date</option>
            <option value="ytd">Year to Date</option>
          </select>
          <button className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-800 px-4 py-2 text-sm text-white hover:bg-gray-700">
            <Download className="h-4 w-4" /> Export
          </button>
        </div>
      </div>

      {/* Revenue Stats */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {revenueStats.map((stat) => (
          <div key={stat.name} className="rounded-xl bg-gray-800 p-6">
            <div className="flex items-center justify-between">
              <p className="text-sm text-gray-400">{stat.name}</p>
              <span className={`flex items-center text-sm font-medium ${
                stat.trend === 'up' ? 'text-green-400' : 'text-red-400'
              }`}>
                {stat.change}
                {stat.trend === 'up' ? <ArrowUpRight className="ml-1 h-4 w-4" /> : <TrendingDown className="ml-1 h-4 w-4" />}
              </span>
            </div>
            <p className="mt-2 text-3xl font-bold text-white">{stat.value}</p>
          </div>
        ))}
      </div>

      {/* Charts Row */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Daily Revenue Chart */}
        <div className="rounded-xl bg-gray-800 p-6">
          <div className="mb-6 flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-white">Daily Revenue</h2>
              <p className="text-sm text-gray-400">Last 7 days</p>
            </div>
            <div className="flex items-center gap-4 text-sm">
              <div className="flex items-center gap-2">
                <div className="h-3 w-3 rounded-full bg-purple-500" />
                <span className="text-gray-400">Revenue</span>
              </div>
            </div>
          </div>
          <div className="flex h-64 items-end justify-between gap-2">
            {dailyRevenue.map((day) => (
              <div key={day.date} className="flex flex-1 flex-col items-center">
                <div
                  className="w-full rounded-t-lg bg-gradient-to-t from-purple-600 to-purple-400 transition-all hover:from-purple-500 hover:to-purple-300"
                  style={{ height: `${(day.revenue / maxRevenue) * 200}px` }}
                  title={`$${(day.revenue / 1000).toFixed(0)}K`}
                />
                <span className="mt-2 text-xs text-gray-500">{day.date}</span>
                <span className="text-xs font-medium text-gray-400">${(day.revenue / 1000).toFixed(0)}K</span>
              </div>
            ))}
          </div>
        </div>

        {/* Revenue by Format */}
        <div className="rounded-xl bg-gray-800 p-6">
          <div className="mb-6 flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-white">Revenue by Format</h2>
              <p className="text-sm text-gray-400">Distribution across ad formats</p>
            </div>
            <PieChart className="h-5 w-5 text-gray-500" />
          </div>
          <div className="space-y-4">
            {revenueByFormat.map((format, idx) => (
              <div key={format.format} className="flex items-center justify-between">
                <div className="flex items-center gap-3 flex-1">
                  <div className={`h-3 w-3 rounded-full ${
                    ['bg-purple-500', 'bg-blue-500', 'bg-green-500', 'bg-yellow-500', 'bg-orange-500', 'bg-pink-500'][idx]
                  }`} />
                  <span className="text-sm text-gray-300">{format.format}</span>
                </div>
                <div className="flex items-center gap-4">
                  <div className="w-24 h-2 rounded-full bg-gray-700">
                    <div 
                      className={`h-2 rounded-full ${
                        ['bg-purple-500', 'bg-blue-500', 'bg-green-500', 'bg-yellow-500', 'bg-orange-500', 'bg-pink-500'][idx]
                      }`}
                      style={{ width: `${format.share}%` }}
                    />
                  </div>
                  <span className="text-sm text-gray-400 w-12 text-right">{format.share}%</span>
                  <span className="text-sm font-medium text-white w-20 text-right">${(format.revenue / 1000).toFixed(0)}K</span>
                  <span className={`text-xs w-12 text-right ${format.change.startsWith('+') ? 'text-green-400' : 'text-red-400'}`}>
                    {format.change}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Revenue by Vertical */}
      <div className="rounded-xl bg-gray-800 p-6">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-white">Revenue by Vertical</h2>
            <p className="text-sm text-gray-400">Performance across industry verticals</p>
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-700">
                <th className="pb-4 text-left text-xs font-medium uppercase text-gray-500">Vertical</th>
                <th className="pb-4 text-left text-xs font-medium uppercase text-gray-500">Revenue</th>
                <th className="pb-4 text-left text-xs font-medium uppercase text-gray-500">Share</th>
                <th className="pb-4 text-left text-xs font-medium uppercase text-gray-500">Clients</th>
                <th className="pb-4 text-left text-xs font-medium uppercase text-gray-500">Avg eCPM</th>
                <th className="pb-4 text-left text-xs font-medium uppercase text-gray-500">Trend</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-700">
              {revenueByVertical.map((vertical) => {
                const totalRevenue = revenueByVertical.reduce((sum, v) => sum + v.revenue, 0)
                const share = (vertical.revenue / totalRevenue * 100).toFixed(1)
                return (
                  <tr key={vertical.vertical} className="hover:bg-gray-700/50">
                    <td className="py-4">
                      <span className="font-medium text-white">{vertical.vertical}</span>
                    </td>
                    <td className="py-4">
                      <span className="font-medium text-white">${(vertical.revenue / 1000).toFixed(0)}K</span>
                    </td>
                    <td className="py-4">
                      <div className="flex items-center gap-2">
                        <div className="h-2 w-20 rounded-full bg-gray-700">
                          <div 
                            className="h-2 rounded-full bg-purple-500"
                            style={{ width: `${Number(share)}%` }}
                          />
                        </div>
                        <span className="text-sm text-gray-400">{share}%</span>
                      </div>
                    </td>
                    <td className="py-4 text-gray-400">{vertical.clients}</td>
                    <td className="py-4">
                      <span className={`font-medium ${vertical.avgEcpm >= 5 ? 'text-green-400' : 'text-white'}`}>
                        ${vertical.avgEcpm.toFixed(2)}
                      </span>
                    </td>
                    <td className="py-4">
                      <span className="flex items-center gap-1 text-green-400">
                        <TrendingUp className="h-4 w-4" />
                        <span className="text-sm">+{Math.floor(Math.random() * 20 + 5)}%</span>
                      </span>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
