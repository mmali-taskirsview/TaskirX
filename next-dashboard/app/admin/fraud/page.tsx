'use client'

import { useState } from 'react'
import {
  Shield,
  AlertTriangle,
  XCircle,
  CheckCircle,
  TrendingDown,
  Eye,
  Ban,
  RefreshCw,
  Search,
  Filter,
  Download,
} from 'lucide-react'

const fraudStats = [
  { name: 'Fraud Rate', value: '2.4%', change: '-0.3%', status: 'good' },
  { name: 'Blocked Requests', value: '12.4M', change: '+2.1M', status: 'warning' },
  { name: 'IVT Detected', value: '8.2%', change: '-1.2%', status: 'good' },
  { name: 'Bot Traffic', value: '1.8%', change: '-0.5%', status: 'good' },
]

const fraudAlerts = [
  {
    id: 1,
    client: 'AppDev Inc',
    type: 'High Fraud Rate',
    severity: 'high',
    rate: 12.4,
    description: 'Fraud rate exceeds 10% threshold',
    time: '5 min ago',
    status: 'active',
  },
  {
    id: 2,
    client: 'GameStudio Pro',
    type: 'Bot Cluster',
    severity: 'medium',
    rate: 4.2,
    description: 'Unusual click pattern from IP range 103.x.x.x',
    time: '1 hour ago',
    status: 'investigating',
  },
  {
    id: 3,
    client: 'MediaGroup SEA',
    type: 'Click Injection',
    severity: 'low',
    rate: 2.1,
    description: 'Potential click injection detected from 3 apps',
    time: '3 hours ago',
    status: 'resolved',
  },
  {
    id: 4,
    client: 'E-Shop Global',
    type: 'Data Center Traffic',
    severity: 'medium',
    rate: 5.8,
    description: 'Traffic from known data center IPs',
    time: '6 hours ago',
    status: 'investigating',
  },
]

const blockedSources = [
  { type: 'IP Range', value: '103.21.0.0/16', reason: 'Bot network', blocked: '2.4M requests', date: '2026-01-15' },
  { type: 'Device ID', value: 'a1b2c3...', reason: 'Click fraud', blocked: '890K requests', date: '2026-01-20' },
  { type: 'App Bundle', value: 'com.spam.app', reason: 'Ad stacking', blocked: '1.2M requests', date: '2026-01-25' },
  { type: 'User Agent', value: 'Bot/1.0', reason: 'Known bot', blocked: '3.1M requests', date: '2026-01-28' },
]

export default function AdminFraud() {
  const [alertFilter, setAlertFilter] = useState('all')

  const filteredAlerts = fraudAlerts.filter(alert =>
    alertFilter === 'all' || alert.status === alertFilter
  )

  const getSeverityBadge = (severity: string) => {
    switch (severity) {
      case 'high':
        return <span className="rounded-full bg-red-900/50 px-2.5 py-0.5 text-xs font-medium text-red-400">High</span>
      case 'medium':
        return <span className="rounded-full bg-yellow-900/50 px-2.5 py-0.5 text-xs font-medium text-yellow-400">Medium</span>
      case 'low':
        return <span className="rounded-full bg-blue-900/50 px-2.5 py-0.5 text-xs font-medium text-blue-400">Low</span>
      default:
        return null
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <span className="flex items-center gap-1 text-xs text-red-400"><AlertTriangle className="h-3 w-3" /> Active</span>
      case 'investigating':
        return <span className="flex items-center gap-1 text-xs text-yellow-400"><Eye className="h-3 w-3" /> Investigating</span>
      case 'resolved':
        return <span className="flex items-center gap-1 text-xs text-green-400"><CheckCircle className="h-3 w-3" /> Resolved</span>
      default:
        return null
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Fraud Detection</h1>
          <p className="text-gray-400">Monitor and manage fraud prevention</p>
        </div>
        <div className="flex gap-3">
          <button className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-800 px-4 py-2 text-sm text-white hover:bg-gray-700">
            <RefreshCw className="h-4 w-4" /> Refresh
          </button>
          <button className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-800 px-4 py-2 text-sm text-white hover:bg-gray-700">
            <Download className="h-4 w-4" /> Export Report
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-4">
        {fraudStats.map((stat) => (
          <div key={stat.name} className="rounded-xl bg-gray-800 p-5">
            <p className="text-sm text-gray-400">{stat.name}</p>
            <div className="mt-2 flex items-end justify-between">
              <p className="text-2xl font-bold text-white">{stat.value}</p>
              <span className={`flex items-center text-sm ${
                stat.status === 'good' ? 'text-green-400' : 'text-yellow-400'
              }`}>
                <TrendingDown className="mr-1 h-4 w-4" />
                {stat.change}
              </span>
            </div>
          </div>
        ))}
      </div>

      {/* AI Model Status */}
      <div className="rounded-xl bg-gradient-to-r from-green-900/30 to-blue-900/30 border border-green-700/50 p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="rounded-lg bg-green-500/20 p-3">
              <Shield className="h-6 w-6 text-green-400" />
            </div>
            <div>
              <h3 className="font-semibold text-white">AI Fraud Detection Model</h3>
              <p className="text-sm text-gray-400">Random Forest + Neural Network ensemble</p>
            </div>
          </div>
          <div className="flex items-center gap-6">
            <div className="text-center">
              <p className="text-2xl font-bold text-green-400">97.8%</p>
              <p className="text-xs text-gray-400">Accuracy</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-blue-400">12ms</p>
              <p className="text-xs text-gray-400">Latency</p>
            </div>
            <span className="flex items-center gap-2 rounded-full bg-green-900/50 px-3 py-1 text-sm text-green-400">
              <span className="h-2 w-2 rounded-full bg-green-400 animate-pulse" />
              Active
            </span>
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Fraud Alerts */}
        <div className="rounded-xl bg-gray-800">
          <div className="border-b border-gray-700 p-6">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-white">Fraud Alerts</h2>
              <div className="flex gap-2">
                {['all', 'active', 'investigating', 'resolved'].map(status => (
                  <button
                    key={status}
                    onClick={() => setAlertFilter(status)}
                    className={`rounded-lg px-3 py-1 text-xs font-medium capitalize ${
                      alertFilter === status ? 'bg-purple-600 text-white' : 'bg-gray-700 text-gray-400 hover:text-white'
                    }`}
                  >
                    {status}
                  </button>
                ))}
              </div>
            </div>
          </div>
          <div className="divide-y divide-gray-700">
            {filteredAlerts.map((alert) => (
              <div key={alert.id} className="p-4 hover:bg-gray-700/50">
                <div className="flex items-start justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-white">{alert.client}</span>
                      {getSeverityBadge(alert.severity)}
                    </div>
                    <p className="mt-1 text-sm text-gray-400">{alert.type}: {alert.description}</p>
                    <div className="mt-2 flex items-center gap-4 text-xs">
                      <span className="text-gray-500">{alert.time}</span>
                      {getStatusBadge(alert.status)}
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-lg font-bold text-red-400">{alert.rate}%</p>
                    <p className="text-xs text-gray-500">fraud rate</p>
                  </div>
                </div>
                {alert.status === 'active' && (
                  <div className="mt-3 flex gap-2">
                    <button className="rounded-lg bg-yellow-900/50 px-3 py-1 text-xs font-medium text-yellow-400 hover:bg-yellow-900">
                      Investigate
                    </button>
                    <button className="rounded-lg bg-red-900/50 px-3 py-1 text-xs font-medium text-red-400 hover:bg-red-900">
                      Block Client
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Block List */}
        <div className="rounded-xl bg-gray-800">
          <div className="border-b border-gray-700 p-6">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-white">Block List</h2>
              <button className="flex items-center gap-1 text-sm text-purple-400 hover:text-purple-300">
                <Ban className="h-4 w-4" /> Add Entry
              </button>
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-700">
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">Type</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">Value</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">Blocked</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700">
                {blockedSources.map((source, idx) => (
                  <tr key={idx} className="hover:bg-gray-700/50">
                    <td className="px-4 py-3">
                      <span className="rounded bg-gray-700 px-2 py-0.5 text-xs text-gray-300">{source.type}</span>
                    </td>
                    <td className="px-4 py-3">
                      <span className="font-mono text-sm text-gray-400">{source.value}</span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-400">{source.blocked}</td>
                    <td className="px-4 py-3">
                      <button className="text-red-400 hover:text-red-300">
                        <XCircle className="h-4 w-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}
