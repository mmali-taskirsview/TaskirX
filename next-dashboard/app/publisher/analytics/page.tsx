'use client';

import React, { useState, useEffect } from 'react';
import {
  BarChart3,
  Calendar,
  Download,
  Filter,
  TrendingUp,
  TrendingDown,
  DollarSign,
  Eye,
  MousePointer,
  Clock,
  Globe,
  Monitor,
  Smartphone,
  Layers,
  ArrowUpRight,
  ArrowDownRight,
  Loader2
} from 'lucide-react';
import { api } from '@/lib/api';

interface DashboardMetrics {
  totalAdUnits: number;
  activeAdUnits: number;
  totalPlacements: number;
  activeDemandPartners: number;
  metrics: {
    totalImpressions: number;
    totalRequests: number;
    totalRevenue: number;
    fillRate: string;
    ecpm: string;
  };
}

// Mock daily data for charts (would come from ClickHouse in production)
const dailyData = [
  { date: '2026-01-30', impressions: 580000, revenue: 6200, ecpm: 10.69, fillRate: 89.5 },
  { date: '2026-01-31', impressions: 620000, revenue: 7100, ecpm: 11.45, fillRate: 90.2 },
  { date: '2026-02-01', impressions: 590000, revenue: 6800, ecpm: 11.53, fillRate: 91.0 },
  { date: '2026-02-02', impressions: 540000, revenue: 5900, ecpm: 10.93, fillRate: 88.7 },
  { date: '2026-02-03', impressions: 610000, revenue: 7200, ecpm: 11.80, fillRate: 92.1 },
  { date: '2026-02-04', impressions: 650000, revenue: 7800, ecpm: 12.00, fillRate: 93.5 },
  { date: '2026-02-05', impressions: 660000, revenue: 7500, ecpm: 11.36, fillRate: 91.8 }
];

const topPlacements = [
  { name: 'Homepage Hero Banner', impressions: 850000, revenue: 8500, ecpm: 10.00, fillRate: 94.5 },
  { name: 'In-Article Native', impressions: 720000, revenue: 10800, ecpm: 15.00, fillRate: 88.7 },
  { name: 'Sidebar Rectangle', impressions: 650000, revenue: 5200, ecpm: 8.00, fillRate: 91.2 },
  { name: 'Mobile Interstitial', impressions: 480000, revenue: 9600, ecpm: 20.00, fillRate: 96.3 },
  { name: 'Footer Leaderboard', impressions: 420000, revenue: 2100, ecpm: 5.00, fillRate: 85.1 }
];

const geoBreakdown = [
  { country: 'United States', flag: '🇺🇸', impressions: 1800000, revenue: 25200, ecpm: 14.00, share: 42.4 },
  { country: 'United Kingdom', flag: '🇬🇧', impressions: 650000, revenue: 7800, ecpm: 12.00, share: 15.3 },
  { country: 'Canada', flag: '🇨🇦', impressions: 420000, revenue: 4620, ecpm: 11.00, share: 9.9 },
  { country: 'Germany', flag: '🇩🇪', impressions: 380000, revenue: 3800, ecpm: 10.00, share: 8.9 },
  { country: 'Australia', flag: '🇦🇺', impressions: 320000, revenue: 3520, ecpm: 11.00, share: 7.5 },
  { country: 'Other', flag: '🌍', impressions: 680000, revenue: 3560, ecpm: 5.24, share: 16.0 }
];

const deviceBreakdown = [
  { device: 'Desktop', icon: Monitor, impressions: 1950000, revenue: 23400, ecpm: 12.00, share: 45.9 },
  { device: 'Mobile', icon: Smartphone, impressions: 1870000, revenue: 20570, ecpm: 11.00, share: 44.0 },
  { device: 'Tablet', icon: Layers, impressions: 430000, revenue: 4530, ecpm: 10.53, share: 10.1 }
];

export default function PublisherAnalyticsPage() {
  const [dateRange, setDateRange] = useState('7d');
  const [activeTab, setActiveTab] = useState('overview');
  const [loading, setLoading] = useState(true);
  const [dashboardData, setDashboardData] = useState<DashboardMetrics | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Fetch dashboard metrics
  useEffect(() => {
    async function fetchDashboard() {
      try {
        setLoading(true);
        const response = await api.getSSPDashboard();
        setDashboardData(response.data);
        setError(null);
      } catch (err) {
        console.error('Error fetching dashboard:', err);
        setError('Failed to load dashboard data');
      } finally {
        setLoading(false);
      }
    }
    fetchDashboard();
  }, []);

  const formatNumber = (num: number) => {
    if (num >= 1000000) return (num / 1000000).toFixed(2) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
    return num.toString();
  };

  // Calculate overview stats from dashboard data
  const overviewStats = dashboardData ? {
    impressions: { value: dashboardData.metrics.totalImpressions, change: 12.5, trend: 'up' as const },
    revenue: { value: dashboardData.metrics.totalRevenue, change: 8.3, trend: 'up' as const },
    ecpm: { value: parseFloat(dashboardData.metrics.ecpm), change: -2.1, trend: 'down' as const },
    fillRate: { value: parseFloat(dashboardData.metrics.fillRate), change: 3.4, trend: 'up' as const },
    adUnits: { value: dashboardData.activeAdUnits, change: 0, trend: 'up' as const },
    partners: { value: dashboardData.activeDemandPartners, change: 1, trend: 'up' as const }
  } : null;

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <Loader2 className="w-8 h-8 animate-spin text-emerald-600" />
        <span className="ml-2 text-gray-600">Loading analytics...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Analytics</h1>
          <p className="text-gray-600 mt-1">Track your inventory performance and revenue</p>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={dateRange}
            onChange={(e) => setDateRange(e.target.value)}
            className="px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
          >
            <option value="1d">Today</option>
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
            <option value="90d">Last 90 Days</option>
          </select>
          <button className="flex items-center gap-2 px-4 py-2 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors">
            <Download className="w-4 h-4" />
            Export
          </button>
        </div>
      </div>

      {/* Overview Stats */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        {overviewStats && Object.entries(overviewStats).map(([key, stat]) => (
          <div key={key} className="bg-white rounded-xl p-4 border border-gray-200">
            <p className="text-sm text-gray-500 capitalize">{key.replace(/([A-Z])/g, ' $1').trim()}</p>
            <div className="flex items-end justify-between mt-2">
              <p className="text-2xl font-bold text-gray-900">
                {key === 'revenue' ? `$${formatNumber(stat.value)}` :
                 key === 'ecpm' ? `$${stat.value.toFixed(2)}` :
                 key === 'fillRate' ? `${stat.value.toFixed(1)}%` :
                 formatNumber(stat.value)}
              </p>
              {stat.change !== 0 && (
                <span className={`flex items-center text-sm ${stat.trend === 'up' ? 'text-green-600' : 'text-red-600'}`}>
                  {stat.trend === 'up' ? <ArrowUpRight className="w-4 h-4" /> : <ArrowDownRight className="w-4 h-4" />}
                  {Math.abs(stat.change)}%
                </span>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="flex gap-8">
          {['overview', 'placements', 'geo', 'devices'].map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`pb-4 px-1 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab
                  ? 'border-emerald-500 text-emerald-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              {tab.charAt(0).toUpperCase() + tab.slice(1)}
            </button>
          ))}
        </nav>
      </div>

      {/* Overview Tab */}
      {activeTab === 'overview' && (
        <div className="space-y-6">
          {/* Daily Performance Chart Placeholder */}
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <h3 className="font-semibold text-gray-900 mb-4">Daily Performance</h3>
            <div className="h-64 flex items-end gap-2">
              {dailyData.map((day, index) => (
                <div key={day.date} className="flex-1 flex flex-col items-center gap-2">
                  <div
                    className="w-full bg-emerald-500 rounded-t"
                    style={{ height: `${(day.revenue / 8000) * 200}px` }}
                    title={`$${day.revenue.toLocaleString()}`}
                  />
                  <span className="text-xs text-gray-500">
                    {new Date(day.date).toLocaleDateString('en-US', { weekday: 'short' })}
                  </span>
                </div>
              ))}
            </div>
            <div className="flex items-center justify-center gap-8 mt-4 pt-4 border-t border-gray-100">
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 bg-emerald-500 rounded" />
                <span className="text-sm text-gray-600">Revenue</span>
              </div>
            </div>
          </div>

          {/* Daily Table */}
          <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
            <div className="p-4 border-b border-gray-100">
              <h3 className="font-semibold text-gray-900">Daily Breakdown</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Impressions</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Revenue</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">eCPM</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Fill Rate</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {dailyData.map((day) => (
                    <tr key={day.date} className="hover:bg-gray-50">
                      <td className="px-4 py-3 text-sm text-gray-900">
                        {new Date(day.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-600 text-right">{formatNumber(day.impressions)}</td>
                      <td className="px-4 py-3 text-sm font-medium text-emerald-600 text-right">${day.revenue.toLocaleString()}</td>
                      <td className="px-4 py-3 text-sm text-gray-600 text-right">${day.ecpm.toFixed(2)}</td>
                      <td className="px-4 py-3 text-sm text-gray-600 text-right">{day.fillRate}%</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Placements Tab */}
      {activeTab === 'placements' && (
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div className="p-4 border-b border-gray-100">
            <h3 className="font-semibold text-gray-900">Top Performing Placements</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Placement</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Impressions</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Revenue</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">eCPM</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Fill Rate</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {topPlacements.map((placement) => (
                  <tr key={placement.name} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm font-medium text-gray-900">{placement.name}</td>
                    <td className="px-4 py-3 text-sm text-gray-600 text-right">{formatNumber(placement.impressions)}</td>
                    <td className="px-4 py-3 text-sm font-medium text-emerald-600 text-right">${placement.revenue.toLocaleString()}</td>
                    <td className="px-4 py-3 text-sm text-gray-600 text-right">${placement.ecpm.toFixed(2)}</td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <div className="w-16 bg-gray-200 rounded-full h-2">
                          <div className="bg-emerald-500 h-2 rounded-full" style={{ width: `${placement.fillRate}%` }} />
                        </div>
                        <span className="text-sm text-gray-600">{placement.fillRate}%</span>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Geo Tab */}
      {activeTab === 'geo' && (
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div className="p-4 border-b border-gray-100">
            <h3 className="font-semibold text-gray-900">Geographic Performance</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Country</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Share</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Impressions</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Revenue</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">eCPM</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {geoBreakdown.map((geo) => (
                  <tr key={geo.country} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm text-gray-900">
                      <span className="mr-2">{geo.flag}</span>
                      {geo.country}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <div className="w-20 bg-gray-200 rounded-full h-2">
                          <div className="bg-blue-500 h-2 rounded-full" style={{ width: `${geo.share}%` }} />
                        </div>
                        <span className="text-sm text-gray-600 w-12">{geo.share}%</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600 text-right">{formatNumber(geo.impressions)}</td>
                    <td className="px-4 py-3 text-sm font-medium text-emerald-600 text-right">${geo.revenue.toLocaleString()}</td>
                    <td className="px-4 py-3 text-sm text-gray-600 text-right">${geo.ecpm.toFixed(2)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Devices Tab */}
      {activeTab === 'devices' && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          {deviceBreakdown.map((device) => (
            <div key={device.device} className="bg-white rounded-xl border border-gray-200 p-6">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-3 bg-emerald-100 rounded-xl">
                  <device.icon className="w-6 h-6 text-emerald-600" />
                </div>
                <div>
                  <h3 className="font-semibold text-gray-900">{device.device}</h3>
                  <p className="text-sm text-gray-500">{device.share}% of traffic</p>
                </div>
              </div>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-500">Impressions</span>
                  <span className="text-sm font-medium text-gray-900">{formatNumber(device.impressions)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-500">Revenue</span>
                  <span className="text-sm font-medium text-emerald-600">${device.revenue.toLocaleString()}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-500">eCPM</span>
                  <span className="text-sm font-medium text-gray-900">${device.ecpm.toFixed(2)}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
