'use client';

import React from 'react';
import {
  DollarSign,
  TrendingUp,
  Eye,
  Zap,
  ArrowUpRight,
  ArrowDownRight,
  Globe,
  Monitor,
  Smartphone,
  Tablet,
  BarChart3,
  PieChart,
  Activity
} from 'lucide-react';

// Mock data for publisher dashboard
const stats = [
  {
    name: 'Today\'s Revenue',
    value: '$12,847',
    change: '+18.2%',
    trend: 'up',
    icon: DollarSign,
    color: 'emerald'
  },
  {
    name: 'Ad Impressions',
    value: '2.4M',
    change: '+12.5%',
    trend: 'up',
    icon: Eye,
    color: 'blue'
  },
  {
    name: 'Fill Rate',
    value: '94.2%',
    change: '+2.1%',
    trend: 'up',
    icon: Zap,
    color: 'purple'
  },
  {
    name: 'eCPM',
    value: '$4.82',
    change: '-0.3%',
    trend: 'down',
    icon: TrendingUp,
    color: 'amber'
  },
];

const revenueByFormat = [
  { format: 'Display Banner', revenue: 45200, impressions: '1.2M', ecpm: '$3.77', fill: '96%' },
  { format: 'Video (Instream)', revenue: 32100, impressions: '420K', ecpm: '$7.64', fill: '89%' },
  { format: 'Native Ads', revenue: 18500, impressions: '580K', ecpm: '$3.19', fill: '92%' },
  { format: 'Interstitial', revenue: 12800, impressions: '160K', ecpm: '$8.00', fill: '78%' },
];

const topDemandPartners = [
  { name: 'Google AdX', spend: '$42,300', share: '35%', fillRate: '98%' },
  { name: 'Amazon TAM', spend: '$28,400', share: '24%', fillRate: '94%' },
  { name: 'Magnite', spend: '$18,200', share: '15%', fillRate: '91%' },
  { name: 'PubMatic', spend: '$15,600', share: '13%', fillRate: '89%' },
  { name: 'OpenX', spend: '$9,500', share: '8%', fillRate: '86%' },
];

const deviceBreakdown = [
  { device: 'Mobile', icon: Smartphone, percentage: 58, revenue: '$7,451' },
  { device: 'Desktop', icon: Monitor, percentage: 32, revenue: '$4,111' },
  { device: 'Tablet', icon: Tablet, percentage: 10, revenue: '$1,285' },
];

const recentActivity = [
  { time: '2 min ago', event: 'Floor price adjusted', detail: 'Leaderboard 728x90 → $2.50' },
  { time: '15 min ago', event: 'New demand partner', detail: 'Index Exchange connected' },
  { time: '1 hour ago', event: 'Ad unit created', detail: 'Homepage Sticky Footer' },
  { time: '3 hours ago', event: 'Payout processed', detail: '$8,420 sent to bank' },
];

export default function PublisherDashboard() {
  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Publisher Dashboard</h1>
          <p className="text-gray-500 mt-1">Monitor your inventory performance and revenue</p>
        </div>
        <div className="flex items-center gap-3">
          <select className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-emerald-500">
            <option>Today</option>
            <option>Last 7 days</option>
            <option>Last 30 days</option>
            <option>This month</option>
          </select>
          <button className="px-4 py-2 bg-emerald-600 text-white rounded-lg text-sm font-medium hover:bg-emerald-700">
            Download Report
          </button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {stats.map((stat) => {
          const Icon = stat.icon;
          return (
            <div key={stat.name} className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <div className="flex items-center justify-between">
                <div className={`p-2 rounded-lg bg-${stat.color}-100`}>
                  <Icon className={`w-5 h-5 text-${stat.color}-600`} />
                </div>
                <span className={`flex items-center text-sm font-medium ${
                  stat.trend === 'up' ? 'text-green-600' : 'text-red-600'
                }`}>
                  {stat.change}
                  {stat.trend === 'up' ? (
                    <ArrowUpRight className="w-4 h-4 ml-1" />
                  ) : (
                    <ArrowDownRight className="w-4 h-4 ml-1" />
                  )}
                </span>
              </div>
              <div className="mt-4">
                <p className="text-2xl font-bold text-gray-900">{stat.value}</p>
                <p className="text-sm text-gray-500 mt-1">{stat.name}</p>
              </div>
            </div>
          );
        })}
      </div>

      {/* Revenue by Format & Top Demand Partners */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Revenue by Format */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-gray-900">Revenue by Ad Format</h2>
            <BarChart3 className="w-5 h-5 text-gray-400" />
          </div>
          <div className="space-y-4">
            {revenueByFormat.map((item) => (
              <div key={item.format} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div>
                  <p className="font-medium text-gray-900">{item.format}</p>
                  <p className="text-sm text-gray-500">{item.impressions} impressions</p>
                </div>
                <div className="text-right">
                  <p className="font-semibold text-gray-900">${item.revenue.toLocaleString()}</p>
                  <div className="flex items-center gap-2 text-sm">
                    <span className="text-gray-500">eCPM: {item.ecpm}</span>
                    <span className="text-emerald-600">Fill: {item.fill}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Top Demand Partners */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-gray-900">Top Demand Partners</h2>
            <PieChart className="w-5 h-5 text-gray-400" />
          </div>
          <div className="space-y-3">
            {topDemandPartners.map((partner, index) => (
              <div key={partner.name} className="flex items-center gap-4">
                <div className="w-8 h-8 rounded-full bg-emerald-100 flex items-center justify-center text-sm font-bold text-emerald-700">
                  {index + 1}
                </div>
                <div className="flex-1">
                  <div className="flex items-center justify-between">
                    <p className="font-medium text-gray-900">{partner.name}</p>
                    <p className="font-semibold text-gray-900">{partner.spend}</p>
                  </div>
                  <div className="flex items-center justify-between text-sm text-gray-500">
                    <span>Market share: {partner.share}</span>
                    <span>Fill: {partner.fillRate}</span>
                  </div>
                  <div className="mt-1 h-1.5 bg-gray-100 rounded-full overflow-hidden">
                    <div 
                      className="h-full bg-emerald-500 rounded-full"
                      style={{ width: partner.share }}
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Device Breakdown & Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Device Breakdown */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-6">Device Breakdown</h2>
          <div className="space-y-4">
            {deviceBreakdown.map((item) => {
              const Icon = item.icon;
              return (
                <div key={item.device} className="flex items-center gap-4">
                  <div className="p-2 bg-gray-100 rounded-lg">
                    <Icon className="w-5 h-5 text-gray-600" />
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center justify-between mb-1">
                      <span className="font-medium text-gray-900">{item.device}</span>
                      <span className="text-sm text-gray-600">{item.revenue}</span>
                    </div>
                    <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                      <div 
                        className="h-full bg-emerald-500 rounded-full"
                        style={{ width: `${item.percentage}%` }}
                      />
                    </div>
                    <p className="text-xs text-gray-500 mt-1">{item.percentage}% of traffic</p>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Recent Activity */}
        <div className="lg:col-span-2 bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-gray-900">Recent Activity</h2>
            <Activity className="w-5 h-5 text-gray-400" />
          </div>
          <div className="space-y-4">
            {recentActivity.map((activity, index) => (
              <div key={index} className="flex items-start gap-4 pb-4 border-b border-gray-100 last:border-0">
                <div className="w-2 h-2 mt-2 rounded-full bg-emerald-500" />
                <div className="flex-1">
                  <div className="flex items-center justify-between">
                    <p className="font-medium text-gray-900">{activity.event}</p>
                    <span className="text-sm text-gray-500">{activity.time}</span>
                  </div>
                  <p className="text-sm text-gray-600 mt-1">{activity.detail}</p>
                </div>
              </div>
            ))}
          </div>
          <button className="mt-4 w-full py-2 text-sm text-emerald-600 hover:text-emerald-700 font-medium">
            View All Activity →
          </button>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-gradient-to-r from-emerald-600 to-teal-600 rounded-xl p-6 text-white">
        <h2 className="text-lg font-semibold mb-4">Quick Actions</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <button className="p-4 bg-white/10 rounded-lg hover:bg-white/20 transition-colors text-left">
            <Globe className="w-6 h-6 mb-2" />
            <p className="font-medium">Add Ad Unit</p>
            <p className="text-sm text-white/70">Create new inventory</p>
          </button>
          <button className="p-4 bg-white/10 rounded-lg hover:bg-white/20 transition-colors text-left">
            <TrendingUp className="w-6 h-6 mb-2" />
            <p className="font-medium">Adjust Floors</p>
            <p className="text-sm text-white/70">Optimize pricing</p>
          </button>
          <button className="p-4 bg-white/10 rounded-lg hover:bg-white/20 transition-colors text-left">
            <Zap className="w-6 h-6 mb-2" />
            <p className="font-medium">Connect DSP</p>
            <p className="text-sm text-white/70">Add demand partner</p>
          </button>
          <button className="p-4 bg-white/10 rounded-lg hover:bg-white/20 transition-colors text-left">
            <BarChart3 className="w-6 h-6 mb-2" />
            <p className="font-medium">View Reports</p>
            <p className="text-sm text-white/70">Detailed analytics</p>
          </button>
        </div>
      </div>
    </div>
  );
}
