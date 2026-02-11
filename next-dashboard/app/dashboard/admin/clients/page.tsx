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
  Users,
  Building2,
  DollarSign,
  TrendingUp,
  TrendingDown,
  Search,
  Filter,
  MoreVertical,
  Eye,
  Edit,
  Pause,
  Play,
  Mail,
  Phone,
  Globe,
  Calendar,
  Target,
  BarChart3,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  Plus,
  Download,
} from 'lucide-react'
import { formatCurrency, formatNumber } from '@/lib/utils'

// Client data
interface Client {
  id: string
  name: string
  company: string
  email: string
  phone: string
  status: 'active' | 'paused' | 'churned'
  tier: 'enterprise' | 'growth' | 'starter'
  vertical: string
  joinDate: string
  totalSpend: number
  monthlySpend: number
  campaigns: number
  activeCampaigns: number
  roas: number
  healthScore: 'A' | 'B' | 'C' | 'D'
  lastActivity: string
  accountManager: string
}

const clients: Client[] = [
  {
    id: '1',
    name: 'John Smith',
    company: 'GameStudio Pro',
    email: 'john@gamestudio.com',
    phone: '+1 555-0123',
    status: 'active',
    tier: 'enterprise',
    vertical: 'Gaming',
    joinDate: '2024-03-15',
    totalSpend: 1250000,
    monthlySpend: 125000,
    campaigns: 45,
    activeCampaigns: 12,
    roas: 4.2,
    healthScore: 'A',
    lastActivity: '2 hours ago',
    accountManager: 'Sarah Wilson',
  },
  {
    id: '2',
    name: 'Emily Chen',
    company: 'ShopMax Global',
    email: 'emily@shopmax.com',
    phone: '+1 555-0456',
    status: 'active',
    tier: 'enterprise',
    vertical: 'E-Commerce',
    joinDate: '2024-01-20',
    totalSpend: 890000,
    monthlySpend: 98000,
    campaigns: 32,
    activeCampaigns: 8,
    roas: 3.8,
    healthScore: 'A',
    lastActivity: '5 hours ago',
    accountManager: 'Sarah Wilson',
  },
  {
    id: '3',
    name: 'Michael Brown',
    company: 'FinanceApp Inc',
    email: 'michael@financeapp.com',
    phone: '+1 555-0789',
    status: 'active',
    tier: 'growth',
    vertical: 'Finance',
    joinDate: '2024-06-10',
    totalSpend: 456000,
    monthlySpend: 87000,
    campaigns: 18,
    activeCampaigns: 5,
    roas: 5.2,
    healthScore: 'A',
    lastActivity: '1 day ago',
    accountManager: 'Mike Johnson',
  },
  {
    id: '4',
    name: 'Sarah Davis',
    company: 'TravelBuddy',
    email: 'sarah@travelbuddy.com',
    phone: '+1 555-0321',
    status: 'active',
    tier: 'growth',
    vertical: 'Travel',
    joinDate: '2024-04-05',
    totalSpend: 320000,
    monthlySpend: 65000,
    campaigns: 24,
    activeCampaigns: 6,
    roas: 3.5,
    healthScore: 'B',
    lastActivity: '3 days ago',
    accountManager: 'Mike Johnson',
  },
  {
    id: '5',
    name: 'David Wilson',
    company: 'FitLife Apps',
    email: 'david@fitlife.com',
    phone: '+1 555-0654',
    status: 'active',
    tier: 'starter',
    vertical: 'Health',
    joinDate: '2024-08-22',
    totalSpend: 145000,
    monthlySpend: 54000,
    campaigns: 12,
    activeCampaigns: 4,
    roas: 4.0,
    healthScore: 'B',
    lastActivity: '6 hours ago',
    accountManager: 'Lisa Park',
  },
  {
    id: '6',
    name: 'Jennifer Martinez',
    company: 'QuickNews Media',
    email: 'jennifer@quicknews.com',
    phone: '+1 555-0987',
    status: 'paused',
    tier: 'growth',
    vertical: 'Media',
    joinDate: '2024-02-28',
    totalSpend: 280000,
    monthlySpend: 0,
    campaigns: 15,
    activeCampaigns: 0,
    roas: 2.8,
    healthScore: 'C',
    lastActivity: '2 weeks ago',
    accountManager: 'Lisa Park',
  },
  {
    id: '7',
    name: 'Robert Taylor',
    company: 'EduLearn Plus',
    email: 'robert@edulearn.com',
    phone: '+1 555-0147',
    status: 'active',
    tier: 'starter',
    vertical: 'Education',
    joinDate: '2024-09-15',
    totalSpend: 78000,
    monthlySpend: 32000,
    campaigns: 8,
    activeCampaigns: 3,
    roas: 3.2,
    healthScore: 'B',
    lastActivity: '12 hours ago',
    accountManager: 'Sarah Wilson',
  },
  {
    id: '8',
    name: 'Amanda Lee',
    company: 'StyleBox Fashion',
    email: 'amanda@stylebox.com',
    phone: '+1 555-0258',
    status: 'churned',
    tier: 'starter',
    vertical: 'E-Commerce',
    joinDate: '2024-05-10',
    totalSpend: 45000,
    monthlySpend: 0,
    campaigns: 6,
    activeCampaigns: 0,
    roas: 1.8,
    healthScore: 'D',
    lastActivity: '1 month ago',
    accountManager: 'Mike Johnson',
  },
]

export default function ClientPortfolioPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'paused' | 'churned'>('all')
  const [filterTier, setFilterTier] = useState<'all' | 'enterprise' | 'growth' | 'starter'>('all')
  const [selectedClient, setSelectedClient] = useState<Client | null>(null)

  const filteredClients = clients.filter(client => {
    const matchesSearch = client.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          client.company.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          client.email.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesStatus = filterStatus === 'all' || client.status === filterStatus
    const matchesTier = filterTier === 'all' || client.tier === filterTier
    return matchesSearch && matchesStatus && matchesTier
  })

  // Summary stats
  const totalClients = clients.length
  const activeClients = clients.filter(c => c.status === 'active').length
  const totalMRR = clients.filter(c => c.status === 'active').reduce((sum, c) => sum + c.monthlySpend, 0)
  const avgRoas = clients.filter(c => c.status === 'active').reduce((sum, c) => sum + c.roas, 0) / activeClients

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900/30 dark:text-green-400">
            <CheckCircle2 className="h-3 w-3" />
            Active
          </span>
        )
      case 'paused':
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-yellow-100 px-2.5 py-0.5 text-xs font-medium text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400">
            <Pause className="h-3 w-3" />
            Paused
          </span>
        )
      case 'churned':
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-red-100 px-2.5 py-0.5 text-xs font-medium text-red-700 dark:bg-red-900/30 dark:text-red-400">
            <XCircle className="h-3 w-3" />
            Churned
          </span>
        )
      default:
        return null
    }
  }

  const getTierBadge = (tier: string) => {
    switch (tier) {
      case 'enterprise':
        return <span className="rounded-full bg-purple-100 px-2 py-0.5 text-xs font-medium text-purple-700 dark:bg-purple-900/30 dark:text-purple-400">Enterprise</span>
      case 'growth':
        return <span className="rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">Growth</span>
      case 'starter':
        return <span className="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700 dark:bg-gray-800 dark:text-gray-400">Starter</span>
      default:
        return null
    }
  }

  const getHealthBadge = (score: string) => {
    const colors = {
      'A': 'bg-green-500',
      'B': 'bg-blue-500',
      'C': 'bg-yellow-500',
      'D': 'bg-red-500',
    }
    return (
      <span className={`inline-flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold text-white ${colors[score as keyof typeof colors]}`}>
        {score}
      </span>
    )
  }

  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Client Portfolio</h1>
          <p className="text-muted-foreground">
            Manage and monitor all client accounts
          </p>
        </div>
        <div className="flex gap-2">
          <button className="inline-flex items-center gap-2 rounded-lg border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors">
            <Download className="h-4 w-4" />
            Export
          </button>
          <button className="inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 px-4 py-2 text-sm font-medium text-white shadow-lg hover:opacity-90 transition-opacity">
            <Plus className="h-4 w-4" />
            Add Client
          </button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Clients</p>
                <p className="text-2xl font-bold">{totalClients}</p>
              </div>
              <div className="rounded-full bg-blue-100 p-3 dark:bg-blue-900/30">
                <Users className="h-5 w-5 text-blue-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Active Clients</p>
                <p className="text-2xl font-bold">{activeClients}</p>
                <p className="text-xs text-green-600 mt-1">{((activeClients / totalClients) * 100).toFixed(0)}% of total</p>
              </div>
              <div className="rounded-full bg-green-100 p-3 dark:bg-green-900/30">
                <CheckCircle2 className="h-5 w-5 text-green-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Monthly Revenue</p>
                <p className="text-2xl font-bold">{formatCurrency(totalMRR)}</p>
                <p className="text-xs text-green-600 flex items-center gap-1 mt-1">
                  <TrendingUp className="h-3 w-3" />
                  +8.5% vs last month
                </p>
              </div>
              <div className="rounded-full bg-purple-100 p-3 dark:bg-purple-900/30">
                <DollarSign className="h-5 w-5 text-purple-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Avg ROAS</p>
                <p className="text-2xl font-bold">{avgRoas.toFixed(1)}x</p>
              </div>
              <div className="rounded-full bg-orange-100 p-3 dark:bg-orange-900/30">
                <TrendingUp className="h-5 w-5 text-orange-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search clients..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full rounded-lg border bg-background pl-10 pr-4 py-2"
          />
        </div>
        
        <select
          value={filterStatus}
          onChange={(e) => setFilterStatus(e.target.value as typeof filterStatus)}
          className="rounded-lg border bg-background px-3 py-2"
        >
          <option value="all">All Status</option>
          <option value="active">Active</option>
          <option value="paused">Paused</option>
          <option value="churned">Churned</option>
        </select>

        <select
          value={filterTier}
          onChange={(e) => setFilterTier(e.target.value as typeof filterTier)}
          className="rounded-lg border bg-background px-3 py-2"
        >
          <option value="all">All Tiers</option>
          <option value="enterprise">Enterprise</option>
          <option value="growth">Growth</option>
          <option value="starter">Starter</option>
        </select>
      </div>

      {/* Client List */}
      <Card>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b bg-muted/30">
                  <th className="p-4 text-left font-medium">Client</th>
                  <th className="p-4 text-left font-medium">Status</th>
                  <th className="p-4 text-left font-medium">Tier</th>
                  <th className="p-4 text-left font-medium">Vertical</th>
                  <th className="p-4 text-right font-medium">Monthly Spend</th>
                  <th className="p-4 text-right font-medium">Total Spend</th>
                  <th className="p-4 text-center font-medium">Campaigns</th>
                  <th className="p-4 text-center font-medium">ROAS</th>
                  <th className="p-4 text-center font-medium">Health</th>
                  <th className="p-4 text-center font-medium">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {filteredClients.map((client) => (
                  <tr key={client.id} className="hover:bg-muted/50">
                    <td className="p-4">
                      <div className="flex items-center gap-3">
                        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-purple-600 text-sm font-bold text-white">
                          {client.name.split(' ').map(n => n[0]).join('')}
                        </div>
                        <div>
                          <div className="font-medium">{client.name}</div>
                          <div className="text-xs text-muted-foreground">{client.company}</div>
                        </div>
                      </div>
                    </td>
                    <td className="p-4">{getStatusBadge(client.status)}</td>
                    <td className="p-4">{getTierBadge(client.tier)}</td>
                    <td className="p-4">{client.vertical}</td>
                    <td className="p-4 text-right font-medium">
                      {client.monthlySpend > 0 ? formatCurrency(client.monthlySpend) : '-'}
                    </td>
                    <td className="p-4 text-right">{formatCurrency(client.totalSpend)}</td>
                    <td className="p-4 text-center">
                      <span className="text-green-600 font-medium">{client.activeCampaigns}</span>
                      <span className="text-muted-foreground">/{client.campaigns}</span>
                    </td>
                    <td className="p-4 text-center">
                      <span className={`font-medium ${client.roas >= 4 ? 'text-green-600' : client.roas >= 3 ? 'text-blue-600' : 'text-yellow-600'}`}>
                        {client.roas}x
                      </span>
                    </td>
                    <td className="p-4 text-center">{getHealthBadge(client.healthScore)}</td>
                    <td className="p-4">
                      <div className="flex items-center justify-center gap-1">
                        <button 
                          onClick={() => setSelectedClient(client)}
                          className="p-1.5 rounded hover:bg-muted"
                          title="View details"
                        >
                          <Eye className="h-4 w-4" />
                        </button>
                        <button className="p-1.5 rounded hover:bg-muted" title="Edit">
                          <Edit className="h-4 w-4" />
                        </button>
                        <button className="p-1.5 rounded hover:bg-muted" title="More">
                          <MoreVertical className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {/* Client Detail Modal */}
      {selectedClient && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-2xl rounded-lg bg-white p-6 shadow-xl dark:bg-gray-900 max-h-[90vh] overflow-y-auto">
            <div className="flex items-start justify-between mb-6">
              <div className="flex items-center gap-4">
                <div className="flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-purple-600 text-xl font-bold text-white">
                  {selectedClient.name.split(' ').map(n => n[0]).join('')}
                </div>
                <div>
                  <h2 className="text-xl font-bold">{selectedClient.name}</h2>
                  <p className="text-muted-foreground">{selectedClient.company}</p>
                  <div className="flex items-center gap-2 mt-1">
                    {getStatusBadge(selectedClient.status)}
                    {getTierBadge(selectedClient.tier)}
                  </div>
                </div>
              </div>
              <button 
                onClick={() => setSelectedClient(null)}
                className="p-2 hover:bg-muted rounded-lg"
              >
                <XCircle className="h-5 w-5" />
              </button>
            </div>

            {/* Contact Info */}
            <div className="grid gap-4 md:grid-cols-2 mb-6">
              <div className="flex items-center gap-3 p-3 rounded-lg border">
                <Mail className="h-5 w-5 text-muted-foreground" />
                <div>
                  <div className="text-xs text-muted-foreground">Email</div>
                  <div className="font-medium">{selectedClient.email}</div>
                </div>
              </div>
              <div className="flex items-center gap-3 p-3 rounded-lg border">
                <Phone className="h-5 w-5 text-muted-foreground" />
                <div>
                  <div className="text-xs text-muted-foreground">Phone</div>
                  <div className="font-medium">{selectedClient.phone}</div>
                </div>
              </div>
              <div className="flex items-center gap-3 p-3 rounded-lg border">
                <Calendar className="h-5 w-5 text-muted-foreground" />
                <div>
                  <div className="text-xs text-muted-foreground">Join Date</div>
                  <div className="font-medium">{selectedClient.joinDate}</div>
                </div>
              </div>
              <div className="flex items-center gap-3 p-3 rounded-lg border">
                <Users className="h-5 w-5 text-muted-foreground" />
                <div>
                  <div className="text-xs text-muted-foreground">Account Manager</div>
                  <div className="font-medium">{selectedClient.accountManager}</div>
                </div>
              </div>
            </div>

            {/* Performance Metrics */}
            <h3 className="font-semibold mb-3">Performance Metrics</h3>
            <div className="grid gap-4 md:grid-cols-4 mb-6">
              <div className="p-4 rounded-lg bg-muted/50 text-center">
                <div className="text-2xl font-bold">{formatCurrency(selectedClient.totalSpend)}</div>
                <div className="text-xs text-muted-foreground">Total Spend</div>
              </div>
              <div className="p-4 rounded-lg bg-muted/50 text-center">
                <div className="text-2xl font-bold">{formatCurrency(selectedClient.monthlySpend)}</div>
                <div className="text-xs text-muted-foreground">Monthly Spend</div>
              </div>
              <div className="p-4 rounded-lg bg-muted/50 text-center">
                <div className="text-2xl font-bold">{selectedClient.activeCampaigns}/{selectedClient.campaigns}</div>
                <div className="text-xs text-muted-foreground">Active/Total Campaigns</div>
              </div>
              <div className="p-4 rounded-lg bg-muted/50 text-center">
                <div className="text-2xl font-bold text-green-600">{selectedClient.roas}x</div>
                <div className="text-xs text-muted-foreground">ROAS</div>
              </div>
            </div>

            {/* Health Score */}
            <div className="flex items-center justify-between p-4 rounded-lg border mb-6">
              <div className="flex items-center gap-3">
                <div className="text-3xl">{getHealthBadge(selectedClient.healthScore)}</div>
                <div>
                  <div className="font-semibold">Account Health Score</div>
                  <div className="text-sm text-muted-foreground">Last activity: {selectedClient.lastActivity}</div>
                </div>
              </div>
              <button className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">
                View Report
              </button>
            </div>

            {/* Actions */}
            <div className="flex justify-end gap-3">
              <button className="rounded-lg border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors">
                View Campaigns
              </button>
              <button className="rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 px-4 py-2 text-sm font-medium text-white hover:opacity-90 transition-opacity">
                Contact Client
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
