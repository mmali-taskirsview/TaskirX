'use client'

import { useState } from 'react'
import {
  Building2,
  Search,
  Plus,
  MoreVertical,
  TrendingUp,
  TrendingDown,
  DollarSign,
  Activity,
  AlertTriangle,
  CheckCircle,
  Clock,
  Eye,
  Edit,
  Pause,
  Play,
} from 'lucide-react'

const clients = [
  {
    id: 1,
    name: 'TechCorp Asia',
    tier: 'Enterprise',
    status: 'active',
    health: 'healthy',
    revenue: 245000,
    impressions: 890000000,
    fillRate: 82.4,
    fraudRate: 1.2,
    joinDate: '2024-03-15',
    contacts: 3,
  },
  {
    id: 2,
    name: 'GameStudio Pro',
    tier: 'Enterprise',
    status: 'active',
    health: 'healthy',
    revenue: 189000,
    impressions: 720000000,
    fillRate: 79.8,
    fraudRate: 2.1,
    joinDate: '2024-06-20',
    contacts: 2,
  },
  {
    id: 3,
    name: 'E-Shop Global',
    tier: 'Professional',
    status: 'active',
    health: 'healthy',
    revenue: 156000,
    impressions: 540000000,
    fillRate: 75.2,
    fraudRate: 3.4,
    joinDate: '2024-08-10',
    contacts: 2,
  },
  {
    id: 4,
    name: 'MediaGroup SEA',
    tier: 'Professional',
    status: 'active',
    health: 'healthy',
    revenue: 134000,
    impressions: 480000000,
    fillRate: 81.5,
    fraudRate: 1.8,
    joinDate: '2024-09-05',
    contacts: 4,
  },
  {
    id: 5,
    name: 'AppDev Inc',
    tier: 'Starter',
    status: 'active',
    health: 'warning',
    revenue: 112000,
    impressions: 320000000,
    fillRate: 77.9,
    fraudRate: 12.4,
    joinDate: '2025-01-15',
    contacts: 1,
  },
  {
    id: 6,
    name: 'TravelMax',
    tier: 'Professional',
    status: 'pending',
    health: 'new',
    revenue: 0,
    impressions: 0,
    fillRate: 0,
    fraudRate: 0,
    joinDate: '2026-02-01',
    contacts: 1,
  },
]

const tiers = ['All Tiers', 'Enterprise', 'Professional', 'Starter']
const statuses = ['All Status', 'active', 'pending', 'suspended']

export default function AdminClients() {
  const [searchQuery, setSearchQuery] = useState('')
  const [tierFilter, setTierFilter] = useState('All Tiers')
  const [statusFilter, setStatusFilter] = useState('All Status')

  const filteredClients = clients.filter(client => {
    const matchesSearch = client.name.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesTier = tierFilter === 'All Tiers' || client.tier === tierFilter
    const matchesStatus = statusFilter === 'All Status' || client.status === statusFilter
    return matchesSearch && matchesTier && matchesStatus
  })

  const totalRevenue = clients.reduce((sum, c) => sum + c.revenue, 0)
  const activeClients = clients.filter(c => c.status === 'active').length
  const avgFillRate = clients.filter(c => c.fillRate > 0).reduce((sum, c) => sum + c.fillRate, 0) / clients.filter(c => c.fillRate > 0).length

  const getHealthBadge = (health: string) => {
    switch (health) {
      case 'healthy':
        return <span className="flex items-center gap-1 rounded-full bg-green-900/50 px-2 py-0.5 text-xs font-medium text-green-400"><CheckCircle className="h-3 w-3" /> Healthy</span>
      case 'warning':
        return <span className="flex items-center gap-1 rounded-full bg-yellow-900/50 px-2 py-0.5 text-xs font-medium text-yellow-400"><AlertTriangle className="h-3 w-3" /> Warning</span>
      case 'new':
        return <span className="flex items-center gap-1 rounded-full bg-blue-900/50 px-2 py-0.5 text-xs font-medium text-blue-400"><Clock className="h-3 w-3" /> New</span>
      default:
        return null
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Client Portfolio</h1>
          <p className="text-gray-400">Manage and monitor all platform clients</p>
        </div>
        <a
          href="/admin/clients/onboarding"
          className="flex items-center gap-2 rounded-lg bg-purple-600 px-4 py-2 text-white hover:bg-purple-700"
        >
          <Plus className="h-5 w-5" />
          Onboard Client
        </a>
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-4">
        <div className="rounded-xl bg-gray-800 p-5">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-purple-900/50 p-2">
              <Building2 className="h-5 w-5 text-purple-400" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Total Clients</p>
              <p className="text-2xl font-bold text-white">{clients.length}</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-gray-800 p-5">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-green-900/50 p-2">
              <CheckCircle className="h-5 w-5 text-green-400" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Active</p>
              <p className="text-2xl font-bold text-white">{activeClients}</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-gray-800 p-5">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-blue-900/50 p-2">
              <DollarSign className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Total Revenue (MTD)</p>
              <p className="text-2xl font-bold text-white">${(totalRevenue / 1000).toFixed(0)}K</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-gray-800 p-5">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-orange-900/50 p-2">
              <Activity className="h-5 w-5 text-orange-400" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Avg Fill Rate</p>
              <p className="text-2xl font-bold text-white">{avgFillRate.toFixed(1)}%</p>
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col gap-4 rounded-xl bg-gray-800 p-4 sm:flex-row sm:items-center">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-500" />
          <input
            type="text"
            placeholder="Search clients..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full rounded-lg border border-gray-700 bg-gray-900 py-2 pl-10 pr-4 text-white placeholder-gray-500 focus:border-purple-500 focus:outline-none"
          />
        </div>
        <select
          value={tierFilter}
          onChange={(e) => setTierFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-900 px-4 py-2 text-white focus:border-purple-500 focus:outline-none"
        >
          {tiers.map(tier => <option key={tier} value={tier}>{tier}</option>)}
        </select>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-900 px-4 py-2 text-white focus:border-purple-500 focus:outline-none"
        >
          {statuses.map(status => <option key={status} value={status} className="capitalize">{status}</option>)}
        </select>
      </div>

      {/* Clients Table */}
      <div className="rounded-xl bg-gray-800">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-700">
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Client</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Tier</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Health</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Revenue</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Fill Rate</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Fraud Rate</th>
                <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-700">
              {filteredClients.map((client) => (
                <tr key={client.id} className="hover:bg-gray-700/50">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-purple-900/50 font-bold text-purple-400">
                        {client.name.charAt(0)}
                      </div>
                      <div>
                        <p className="font-medium text-white">{client.name}</p>
                        <p className="text-xs text-gray-500">Since {client.joinDate}</p>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                      client.tier === 'Enterprise' ? 'bg-purple-900/50 text-purple-400' :
                      client.tier === 'Professional' ? 'bg-blue-900/50 text-blue-400' :
                      'bg-gray-700 text-gray-400'
                    }`}>
                      {client.tier}
                    </span>
                  </td>
                  <td className="px-6 py-4">{getHealthBadge(client.health)}</td>
                  <td className="px-6 py-4">
                    <span className="font-medium text-white">
                      {client.revenue > 0 ? `$${(client.revenue / 1000).toFixed(0)}K` : '-'}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    {client.fillRate > 0 ? (
                      <div className="flex items-center gap-2">
                        <div className="h-1.5 w-16 rounded-full bg-gray-700">
                          <div 
                            className={`h-1.5 rounded-full ${client.fillRate >= 80 ? 'bg-green-500' : client.fillRate >= 70 ? 'bg-yellow-500' : 'bg-red-500'}`}
                            style={{ width: `${client.fillRate}%` }}
                          />
                        </div>
                        <span className="text-sm text-gray-400">{client.fillRate}%</span>
                      </div>
                    ) : '-'}
                  </td>
                  <td className="px-6 py-4">
                    {client.fraudRate > 0 ? (
                      <span className={`font-medium ${client.fraudRate <= 3 ? 'text-green-400' : client.fraudRate <= 8 ? 'text-yellow-400' : 'text-red-400'}`}>
                        {client.fraudRate}%
                      </span>
                    ) : '-'}
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <button className="rounded p-1 text-gray-400 hover:bg-gray-700 hover:text-white"><Eye className="h-4 w-4" /></button>
                      <button className="rounded p-1 text-gray-400 hover:bg-gray-700 hover:text-white"><Edit className="h-4 w-4" /></button>
                      {client.status === 'active' ? (
                        <button className="rounded p-1 text-gray-400 hover:bg-gray-700 hover:text-yellow-400"><Pause className="h-4 w-4" /></button>
                      ) : (
                        <button className="rounded p-1 text-gray-400 hover:bg-gray-700 hover:text-green-400"><Play className="h-4 w-4" /></button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
