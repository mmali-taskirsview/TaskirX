'use client'

import { useState } from 'react'
import {
  DollarSign,
  Users,
  TrendingUp,
  Activity,
  Server,
  AlertTriangle,
  ArrowUpRight,
  ArrowDownRight,
  Globe,
  Zap,
  Shield,
  BarChart3,
} from 'lucide-react'

const platformStats = [
  { name: 'Total Revenue (MTD)', value: '$2.4M', change: '+18.2%', trend: 'up', icon: DollarSign },
  { name: 'Active Clients', value: '847', change: '+12', trend: 'up', icon: Users },
  { name: 'Total Impressions', value: '4.2B', change: '+24.5%', trend: 'up', icon: TrendingUp },
  { name: 'Avg. Fill Rate', value: '78.4%', change: '+2.1%', trend: 'up', icon: Activity },
]

const systemHealth = [
  { name: 'Bid Engine', status: 'healthy', latency: '12ms', qps: '1.2M' },
  { name: 'Ad Server', status: 'healthy', latency: '8ms', qps: '2.4M' },
  { name: 'Fraud Detection', status: 'healthy', latency: '45ms', qps: '890K' },
  { name: 'Analytics Pipeline', status: 'healthy', latency: '120ms', qps: '450K' },
  { name: 'Database Cluster', status: 'healthy', latency: '5ms', qps: '3.2M' },
  { name: 'Redis Cache', status: 'healthy', latency: '1ms', qps: '8.5M' },
]

const topClients = [
  { name: 'TechCorp Asia', revenue: 245000, impressions: '890M', fillRate: 82.4, status: 'active' },
  { name: 'GameStudio Pro', revenue: 189000, impressions: '720M', fillRate: 79.8, status: 'active' },
  { name: 'E-Shop Global', revenue: 156000, impressions: '540M', fillRate: 75.2, status: 'active' },
  { name: 'MediaGroup SEA', revenue: 134000, impressions: '480M', fillRate: 81.5, status: 'active' },
  { name: 'AppDev Inc', revenue: 112000, impressions: '320M', fillRate: 77.9, status: 'warning' },
]

const recentAlerts = [
  { type: 'warning', message: 'High fraud rate detected for client AppDev Inc (12.4%)', time: '5 min ago' },
  { type: 'info', message: 'New client TravelMax onboarded successfully', time: '23 min ago' },
  { type: 'success', message: 'System maintenance completed successfully', time: '1 hour ago' },
  { type: 'warning', message: 'Fill rate dropped below 70% for Native format', time: '2 hours ago' },
]

const regionData = [
  { region: 'Indonesia', revenue: 580000, share: 24 },
  { region: 'Thailand', revenue: 420000, share: 17 },
  { region: 'Vietnam', revenue: 380000, share: 16 },
  { region: 'Philippines', revenue: 340000, share: 14 },
  { region: 'Malaysia', revenue: 290000, share: 12 },
  { region: 'Singapore', revenue: 240000, share: 10 },
  { region: 'Others', revenue: 170000, share: 7 },
]

export default function AdminDashboard() {
  const [timeRange, setTimeRange] = useState('today')

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Platform Overview</h1>
          <p className="text-gray-400">Real-time platform performance and health</p>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value)}
            className="rounded-lg border border-gray-700 bg-gray-800 px-4 py-2 text-sm text-white focus:border-purple-500 focus:outline-none"
          >
            <option value="today">Today</option>
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
            <option value="mtd">Month to Date</option>
          </select>
        </div>
      </div>

      {/* Platform Stats */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {platformStats.map((stat) => (
          <div key={stat.name} className="rounded-xl bg-gray-800 p-6">
            <div className="flex items-center justify-between">
              <div className="rounded-lg bg-purple-900/50 p-2">
                <stat.icon className="h-5 w-5 text-purple-400" />
              </div>
              <span className={`flex items-center text-sm font-medium ${
                stat.trend === 'up' ? 'text-green-400' : 'text-red-400'
              }`}>
                {stat.change}
                {stat.trend === 'up' ? <ArrowUpRight className="ml-1 h-4 w-4" /> : <ArrowDownRight className="ml-1 h-4 w-4" />}
              </span>
            </div>
            <div className="mt-4">
              <p className="text-sm text-gray-400">{stat.name}</p>
              <p className="mt-1 text-2xl font-bold text-white">{stat.value}</p>
            </div>
          </div>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* System Health */}
        <div className="lg:col-span-2 rounded-xl bg-gray-800 p-6">
          <div className="mb-4 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Server className="h-5 w-5 text-purple-400" />
              <h2 className="text-lg font-semibold text-white">System Health</h2>
            </div>
            <span className="flex items-center gap-2 rounded-full bg-green-900/50 px-3 py-1 text-xs font-medium text-green-400">
              <span className="h-2 w-2 rounded-full bg-green-400 animate-pulse" />
              All Systems Operational
            </span>
          </div>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {systemHealth.map((system) => (
              <div key={system.name} className="rounded-lg bg-gray-900 p-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-white">{system.name}</span>
                  <span className="h-2 w-2 rounded-full bg-green-400" />
                </div>
                <div className="mt-2 grid grid-cols-2 gap-2 text-xs">
                  <div>
                    <span className="text-gray-500">Latency</span>
                    <p className="font-medium text-gray-300">{system.latency}</p>
                  </div>
                  <div>
                    <span className="text-gray-500">QPS</span>
                    <p className="font-medium text-gray-300">{system.qps}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Recent Alerts */}
        <div className="rounded-xl bg-gray-800 p-6">
          <div className="mb-4 flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-yellow-400" />
            <h2 className="text-lg font-semibold text-white">Recent Alerts</h2>
          </div>
          <div className="space-y-3">
            {recentAlerts.map((alert, idx) => (
              <div key={idx} className="rounded-lg bg-gray-900 p-3">
                <div className="flex items-start gap-2">
                  <span className={`mt-1 h-2 w-2 rounded-full ${
                    alert.type === 'warning' ? 'bg-yellow-400' :
                    alert.type === 'success' ? 'bg-green-400' : 'bg-blue-400'
                  }`} />
                  <div className="flex-1">
                    <p className="text-sm text-gray-300">{alert.message}</p>
                    <p className="mt-1 text-xs text-gray-500">{alert.time}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Top Clients */}
        <div className="rounded-xl bg-gray-800 p-6">
          <div className="mb-4 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Users className="h-5 w-5 text-purple-400" />
              <h2 className="text-lg font-semibold text-white">Top Clients</h2>
            </div>
            <a href="/admin/clients" className="text-sm text-purple-400 hover:text-purple-300">View all →</a>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-700">
                  <th className="pb-3 text-left text-xs font-medium uppercase text-gray-500">Client</th>
                  <th className="pb-3 text-left text-xs font-medium uppercase text-gray-500">Revenue</th>
                  <th className="pb-3 text-left text-xs font-medium uppercase text-gray-500">Fill Rate</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700">
                {topClients.map((client) => (
                  <tr key={client.name}>
                    <td className="py-3">
                      <div className="flex items-center gap-2">
                        <span className={`h-2 w-2 rounded-full ${client.status === 'active' ? 'bg-green-400' : 'bg-yellow-400'}`} />
                        <span className="font-medium text-white">{client.name}</span>
                      </div>
                    </td>
                    <td className="py-3 text-sm text-gray-300">${(client.revenue / 1000).toFixed(0)}K</td>
                    <td className="py-3">
                      <div className="flex items-center gap-2">
                        <div className="h-1.5 w-16 rounded-full bg-gray-700">
                          <div 
                            className="h-1.5 rounded-full bg-purple-500"
                            style={{ width: `${client.fillRate}%` }}
                          />
                        </div>
                        <span className="text-sm text-gray-400">{client.fillRate}%</span>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Revenue by Region */}
        <div className="rounded-xl bg-gray-800 p-6">
          <div className="mb-4 flex items-center gap-2">
            <Globe className="h-5 w-5 text-purple-400" />
            <h2 className="text-lg font-semibold text-white">Revenue by Region</h2>
          </div>
          <div className="space-y-3">
            {regionData.map((region) => (
              <div key={region.region} className="flex items-center justify-between">
                <div className="flex items-center gap-3 flex-1">
                  <span className="w-24 text-sm text-gray-300">{region.region}</span>
                  <div className="flex-1 h-2 rounded-full bg-gray-700">
                    <div 
                      className="h-2 rounded-full bg-gradient-to-r from-purple-500 to-purple-400"
                      style={{ width: `${region.share}%` }}
                    />
                  </div>
                </div>
                <div className="flex items-center gap-4 ml-4">
                  <span className="text-sm text-gray-400">{region.share}%</span>
                  <span className="text-sm font-medium text-white w-20 text-right">${(region.revenue / 1000).toFixed(0)}K</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="grid gap-4 sm:grid-cols-4">
        <a href="/admin/clients/onboarding" className="flex items-center gap-3 rounded-xl bg-gradient-to-br from-purple-600 to-purple-700 p-4 text-white transition-transform hover:scale-[1.02]">
          <Users className="h-6 w-6" />
          <span className="font-medium">Onboard Client</span>
        </a>
        <a href="/admin/price-floors" className="flex items-center gap-3 rounded-xl bg-gradient-to-br from-blue-600 to-blue-700 p-4 text-white transition-transform hover:scale-[1.02]">
          <BarChart3 className="h-6 w-6" />
          <span className="font-medium">Adjust Floors</span>
        </a>
        <a href="/admin/fraud" className="flex items-center gap-3 rounded-xl bg-gradient-to-br from-orange-600 to-orange-700 p-4 text-white transition-transform hover:scale-[1.02]">
          <Shield className="h-6 w-6" />
          <span className="font-medium">Fraud Console</span>
        </a>
        <a href="/admin/system" className="flex items-center gap-3 rounded-xl bg-gradient-to-br from-green-600 to-green-700 p-4 text-white transition-transform hover:scale-[1.02]">
          <Zap className="h-6 w-6" />
          <span className="font-medium">System Status</span>
        </a>
      </div>
    </div>
  )
}
