'use client'

import { useState } from 'react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/card'
import {
  DollarSign,
  TrendingUp,
  TrendingDown,
  PieChart,
  BarChart3,
  Calendar,
  Download,
  Filter,
  Layers,
  Gamepad2,
  ShoppingCart,
  Briefcase,
  Plane,
  Smartphone,
  Heart,
} from 'lucide-react'
import { formatCurrency, formatNumber, formatPercentage } from '@/lib/utils'

// Revenue by Ad Format
const revenueByFormat = [
  { format: 'Rewarded Video', revenue: 245000, impressions: 8500000, ecpm: 28.82, change: 12.5, color: '#22c55e' },
  { format: 'Playable Ads', revenue: 189000, impressions: 4200000, ecpm: 45.00, change: 18.2, color: '#8b5cf6' },
  { format: 'Interstitial', revenue: 156000, impressions: 7800000, ecpm: 20.00, change: 5.3, color: '#f59e0b' },
  { format: 'Native Ads', revenue: 134000, impressions: 9200000, ecpm: 14.57, change: -2.1, color: '#3b82f6' },
  { format: 'Banner', revenue: 98000, impressions: 32000000, ecpm: 3.06, change: -5.4, color: '#6b7280' },
  { format: 'CTV/OTT', revenue: 78000, impressions: 1200000, ecpm: 65.00, change: 25.8, color: '#ec4899' },
  { format: 'Offerwall', revenue: 67000, impressions: 890000, ecpm: 75.28, change: 8.9, color: '#14b8a6' },
  { format: 'Audio Ads', revenue: 45000, impressions: 2100000, ecpm: 21.43, change: 15.2, color: '#f97316' },
]

// Revenue by Vertical
const revenueByVertical = [
  { vertical: 'Gaming', icon: Gamepad2, revenue: 425000, clients: 48, avgCPA: 2.15, change: 15.3, color: '#22c55e' },
  { vertical: 'E-Commerce', icon: ShoppingCart, revenue: 312000, clients: 35, avgCPA: 12.50, change: 8.7, color: '#3b82f6' },
  { vertical: 'Finance', icon: Briefcase, revenue: 198000, clients: 22, avgCPA: 45.00, change: 22.1, color: '#8b5cf6' },
  { vertical: 'Travel', icon: Plane, revenue: 145000, clients: 18, avgCPA: 28.00, change: -3.2, color: '#f59e0b' },
  { vertical: 'Utilities', icon: Smartphone, revenue: 89000, clients: 42, avgCPA: 1.85, change: 12.5, color: '#14b8a6' },
  { vertical: 'Health & Fitness', icon: Heart, revenue: 67000, clients: 15, avgCPA: 8.50, change: 6.8, color: '#ec4899' },
]

// Monthly revenue trend
const monthlyRevenue = [
  { month: 'Sep', revenue: 890000, costs: 445000, profit: 445000 },
  { month: 'Oct', revenue: 945000, costs: 472000, profit: 473000 },
  { month: 'Nov', revenue: 1020000, costs: 510000, profit: 510000 },
  { month: 'Dec', revenue: 1180000, costs: 590000, profit: 590000 },
  { month: 'Jan', revenue: 1012000, costs: 506000, profit: 506000 },
  { month: 'Feb', revenue: 1236000, costs: 618000, profit: 618000 },
]

// Top performing clients
const topClients = [
  { name: 'GameStudio Pro', vertical: 'Gaming', revenue: 125000, spend: 85000, roas: 4.2 },
  { name: 'ShopMax Global', vertical: 'E-Commerce', revenue: 98000, spend: 72000, roas: 3.8 },
  { name: 'FinanceApp Inc', vertical: 'Finance', revenue: 87000, spend: 45000, roas: 5.2 },
  { name: 'TravelBuddy', vertical: 'Travel', revenue: 65000, spend: 48000, roas: 3.5 },
  { name: 'FitLife Apps', vertical: 'Health', revenue: 54000, spend: 38000, roas: 4.0 },
]

export default function AdminRevenuePage() {
  const [dateRange, setDateRange] = useState('30d')
  const [viewMode, setViewMode] = useState<'format' | 'vertical'>('format')

  const totalRevenue = revenueByFormat.reduce((sum, f) => sum + f.revenue, 0)
  const totalImpressions = revenueByFormat.reduce((sum, f) => sum + f.impressions, 0)
  const avgEcpm = (totalRevenue / totalImpressions) * 1000
  const totalProfit = monthlyRevenue.reduce((sum, m) => sum + m.profit, 0)

  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Revenue Analytics</h1>
          <p className="text-muted-foreground">
            Platform revenue breakdown by format and vertical
          </p>
        </div>
        <div className="flex gap-2">
          <select
            value={dateRange}
            onChange={(e) => setDateRange(e.target.value)}
            className="rounded-lg border bg-background px-3 py-2"
          >
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
            <option value="90d">Last 90 Days</option>
            <option value="ytd">Year to Date</option>
          </select>
          <button className="inline-flex items-center gap-2 rounded-lg border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors">
            <Download className="h-4 w-4" />
            Export
          </button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card className="border-green-500/50">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Revenue</p>
                <p className="text-2xl font-bold text-green-600">{formatCurrency(totalRevenue)}</p>
                <p className="text-xs text-green-600 flex items-center gap-1 mt-1">
                  <TrendingUp className="h-3 w-3" />
                  +12.5% vs last period
                </p>
              </div>
              <div className="rounded-full bg-green-100 p-3 dark:bg-green-900/30">
                <DollarSign className="h-6 w-6 text-green-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Impressions</p>
                <p className="text-2xl font-bold">{formatNumber(totalImpressions)}</p>
                <p className="text-xs text-muted-foreground mt-1">Across all formats</p>
              </div>
              <div className="rounded-full bg-blue-100 p-3 dark:bg-blue-900/30">
                <BarChart3 className="h-6 w-6 text-blue-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Avg eCPM</p>
                <p className="text-2xl font-bold">${avgEcpm.toFixed(2)}</p>
                <p className="text-xs text-green-600 flex items-center gap-1 mt-1">
                  <TrendingUp className="h-3 w-3" />
                  +8.3% vs last period
                </p>
              </div>
              <div className="rounded-full bg-purple-100 p-3 dark:bg-purple-900/30">
                <PieChart className="h-6 w-6 text-purple-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-emerald-500/50">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Net Profit (6mo)</p>
                <p className="text-2xl font-bold text-emerald-600">{formatCurrency(totalProfit)}</p>
                <p className="text-xs text-emerald-600 flex items-center gap-1 mt-1">
                  <TrendingUp className="h-3 w-3" />
                  50% margin
                </p>
              </div>
              <div className="rounded-full bg-emerald-100 p-3 dark:bg-emerald-900/30">
                <TrendingUp className="h-6 w-6 text-emerald-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* View Toggle */}
      <div className="flex gap-2">
        <button
          onClick={() => setViewMode('format')}
          className={`flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
            viewMode === 'format'
              ? 'bg-blue-600 text-white'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-300'
          }`}
        >
          <Layers className="h-4 w-4" />
          By Ad Format
        </button>
        <button
          onClick={() => setViewMode('vertical')}
          className={`flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
            viewMode === 'vertical'
              ? 'bg-blue-600 text-white'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-300'
          }`}
        >
          <PieChart className="h-4 w-4" />
          By Vertical
        </button>
      </div>

      {/* Revenue Breakdown */}
      {viewMode === 'format' ? (
        <Card>
          <CardHeader>
            <CardTitle>Revenue by Ad Format</CardTitle>
            <CardDescription>Performance breakdown across all ad formats</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {revenueByFormat.map((format, index) => {
                const percentage = (format.revenue / totalRevenue) * 100
                return (
                  <div key={index} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div 
                          className="h-3 w-3 rounded-full"
                          style={{ backgroundColor: format.color }}
                        />
                        <span className="font-medium">{format.format}</span>
                      </div>
                      <div className="flex items-center gap-6 text-sm">
                        <span className="text-muted-foreground w-24 text-right">
                          {formatNumber(format.impressions)} imp
                        </span>
                        <span className="text-muted-foreground w-20 text-right">
                          ${format.ecpm.toFixed(2)} eCPM
                        </span>
                        <span className={`w-16 text-right flex items-center justify-end gap-1 ${
                          format.change >= 0 ? 'text-green-600' : 'text-red-600'
                        }`}>
                          {format.change >= 0 ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
                          {Math.abs(format.change)}%
                        </span>
                        <span className="font-semibold w-24 text-right">
                          {formatCurrency(format.revenue)}
                        </span>
                      </div>
                    </div>
                    <div className="h-2 w-full rounded-full bg-gray-100 dark:bg-gray-800">
                      <div
                        className="h-full rounded-full transition-all"
                        style={{ 
                          width: `${percentage}%`,
                          backgroundColor: format.color
                        }}
                      />
                    </div>
                  </div>
                )
              })}
            </div>

            {/* Format Distribution Chart */}
            <div className="mt-8 pt-6 border-t">
              <h4 className="font-semibold mb-4">Revenue Distribution</h4>
              <div className="flex items-center gap-4">
                <div className="flex-1 h-8 rounded-full overflow-hidden flex">
                  {revenueByFormat.map((format, index) => {
                    const percentage = (format.revenue / totalRevenue) * 100
                    return (
                      <div
                        key={index}
                        className="h-full transition-all hover:opacity-80"
                        style={{ 
                          width: `${percentage}%`,
                          backgroundColor: format.color
                        }}
                        title={`${format.format}: ${percentage.toFixed(1)}%`}
                      />
                    )
                  })}
                </div>
              </div>
              <div className="mt-4 flex flex-wrap gap-4">
                {revenueByFormat.slice(0, 5).map((format, index) => (
                  <div key={index} className="flex items-center gap-2 text-sm">
                    <div 
                      className="h-3 w-3 rounded-full"
                      style={{ backgroundColor: format.color }}
                    />
                    <span className="text-muted-foreground">{format.format}</span>
                    <span className="font-medium">{((format.revenue / totalRevenue) * 100).toFixed(1)}%</span>
                  </div>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>Revenue by Vertical</CardTitle>
            <CardDescription>Performance breakdown across industry verticals</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {revenueByVertical.map((vertical, index) => {
                const Icon = vertical.icon
                return (
                  <div 
                    key={index}
                    className="rounded-lg border p-4 hover:shadow-md transition-shadow"
                  >
                    <div className="flex items-center gap-3 mb-3">
                      <div 
                        className="rounded-lg p-2"
                        style={{ backgroundColor: `${vertical.color}20` }}
                      >
                        <Icon className="h-5 w-5" style={{ color: vertical.color }} />
                      </div>
                      <div>
                        <h4 className="font-semibold">{vertical.vertical}</h4>
                        <p className="text-xs text-muted-foreground">{vertical.clients} clients</p>
                      </div>
                    </div>
                    <div className="space-y-2">
                      <div className="flex justify-between">
                        <span className="text-sm text-muted-foreground">Revenue</span>
                        <span className="font-semibold">{formatCurrency(vertical.revenue)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-sm text-muted-foreground">Avg CPA</span>
                        <span className="font-medium">{formatCurrency(vertical.avgCPA)}</span>
                      </div>
                      <div className="flex justify-between items-center">
                        <span className="text-sm text-muted-foreground">Change</span>
                        <span className={`flex items-center gap-1 text-sm font-medium ${
                          vertical.change >= 0 ? 'text-green-600' : 'text-red-600'
                        }`}>
                          {vertical.change >= 0 ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
                          {Math.abs(vertical.change)}%
                        </span>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Monthly Trend & Top Clients */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* Monthly Revenue Trend */}
        <Card>
          <CardHeader>
            <CardTitle>Monthly Revenue Trend</CardTitle>
            <CardDescription>Revenue, costs, and profit over time</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {monthlyRevenue.map((month, index) => (
                <div key={index} className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="font-medium w-12">{month.month}</span>
                    <div className="flex gap-4">
                      <span className="text-muted-foreground">{formatCurrency(month.revenue)}</span>
                      <span className="text-green-600 font-medium">{formatCurrency(month.profit)}</span>
                    </div>
                  </div>
                  <div className="h-4 w-full rounded-full bg-gray-100 dark:bg-gray-800 overflow-hidden flex">
                    <div
                      className="h-full bg-blue-500"
                      style={{ width: `${(month.costs / month.revenue) * 100}%` }}
                      title={`Costs: ${formatCurrency(month.costs)}`}
                    />
                    <div
                      className="h-full bg-green-500"
                      style={{ width: `${(month.profit / month.revenue) * 100}%` }}
                      title={`Profit: ${formatCurrency(month.profit)}`}
                    />
                  </div>
                </div>
              ))}
            </div>
            <div className="mt-4 pt-4 border-t flex gap-6">
              <div className="flex items-center gap-2 text-sm">
                <div className="h-3 w-3 rounded-full bg-blue-500" />
                <span className="text-muted-foreground">Costs</span>
              </div>
              <div className="flex items-center gap-2 text-sm">
                <div className="h-3 w-3 rounded-full bg-green-500" />
                <span className="text-muted-foreground">Profit</span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Top Clients */}
        <Card>
          <CardHeader>
            <CardTitle>Top Performing Clients</CardTitle>
            <CardDescription>Highest revenue generating accounts</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {topClients.map((client, index) => (
                <div 
                  key={index}
                  className="flex items-center justify-between rounded-lg border p-3 hover:bg-muted/50 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 text-sm font-bold text-white">
                      {index + 1}
                    </div>
                    <div>
                      <div className="font-medium">{client.name}</div>
                      <div className="text-xs text-muted-foreground">{client.vertical}</div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-semibold">{formatCurrency(client.revenue)}</div>
                    <div className="text-xs text-green-600">{client.roas}x ROAS</div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
