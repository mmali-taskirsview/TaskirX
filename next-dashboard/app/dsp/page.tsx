'use client'

import { useState, useEffect } from 'react'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { 
  TrendingUp, 
  TrendingDown, 
  DollarSign, 
  Zap, 
  Target, 
  Users, 
  BarChart3,
  Activity,
  Clock,
  Award,
  RefreshCw
} from 'lucide-react'

interface DashboardStats {
  totalRequests: number
  totalBids: number
  totalWins: number
  totalSpend: number
  winRate: number
  avgCpm: number
  activeSupplyPartners: number
  activeDeals: number
  activeAudiences: number
  activeBidStrategies: number
  // Frontend/UI specific (optional/derived)
  todayBids?: number
  todayWins?: number
  todaySpend?: number
}

export default function DSPDashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchDashboard = async () => {
    try {
      setLoading(true)
      const response = await api.getDSPDashboard()
      setStats({
        ...response.data,
        // Default missing fields for UI compatibility
        todayBids: 0,
        todayWins: 0,
        todaySpend: 0,
        // avgBidPrice: derived or mock
      })
      setError(null)
    } catch (err: any) {
      setError(err.message || 'Failed to load dashboard')
      // Set mock data for demo
      setStats({
        totalRequests: 25000000,
        totalBids: 15847293,
        totalWins: 2847584,
        totalSpend: 145892.47,
        winRate: 17.97,
        avgCpm: 2.85,
        activeSupplyPartners: 12,
        activeDeals: 28,
        activeBidStrategies: 15,
        activeAudiences: 42,
        todayBids: 1247583,
        todayWins: 224584,
        todaySpend: 12458.92
      })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchDashboard()
  }, [])

  const formatNumber = (num: number) => {
    if (num >= 1000000) return (num / 1000000).toFixed(2) + 'M'
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
    return num.toLocaleString()
  }

  const formatCurrency = (num: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2
    }).format(num)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">DSP Dashboard</h1>
          <p className="text-muted-foreground">
            Demand-Side Platform - Real-Time Bidding Overview
          </p>
        </div>
        <Button onClick={fetchDashboard} variant="outline">
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </Button>
      </div>

      {error && (
        <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
          <p className="text-yellow-800 dark:text-yellow-200 text-sm">
            Using demo data - {error}
          </p>
        </div>
      )}

      {/* Main Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Total Bids</CardTitle>
            <Zap className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatNumber(stats?.totalBids || 0)}</div>
            <p className="text-xs text-muted-foreground">
              Today: {formatNumber(stats?.todayBids || 0)}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Total Wins</CardTitle>
            <Award className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatNumber(stats?.totalWins || 0)}</div>
            <p className="text-xs text-muted-foreground">
              Today: {formatNumber(stats?.todayWins || 0)}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Win Rate</CardTitle>
            <Target className="h-4 w-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.winRate?.toFixed(2)}%</div>
            <div className="flex items-center text-xs text-green-500">
              <TrendingUp className="h-3 w-3 mr-1" />
              +2.3% from last week
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Total Spend</CardTitle>
            <DollarSign className="h-4 w-4 text-yellow-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(stats?.totalSpend || 0)}</div>
            <p className="text-xs text-muted-foreground">
              Today: {formatCurrency(stats?.todaySpend || 0)}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Secondary Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Avg CPM</CardTitle>
            <BarChart3 className="h-4 w-4 text-indigo-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${stats?.avgCpm?.toFixed(2)}</div>
            <p className="text-xs text-muted-foreground">Cost per 1000 impressions</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Supply Partners</CardTitle>
            <Activity className="h-4 w-4 text-cyan-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.activeSupplyPartners}</div>
            <Badge variant="secondary" className="mt-1">Active</Badge>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Active Deals</CardTitle>
            <Clock className="h-4 w-4 text-orange-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.activeDeals}</div>
            <p className="text-xs text-muted-foreground">PMP & Preferred</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Audiences</CardTitle>
            <Users className="h-4 w-4 text-pink-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.activeAudiences}</div>
            <p className="text-xs text-muted-foreground">Targeting segments</p>
          </CardContent>
        </Card>
      </div>

      {/* Quick Links */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card className="hover:shadow-lg transition-shadow cursor-pointer" onClick={() => window.location.href = '/dsp/supply-partners'}>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5 text-cyan-500" />
              Supply Partners
            </CardTitle>
            <CardDescription>Manage SSP connections and QPS limits</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className="text-2xl font-bold">{stats?.activeSupplyPartners}</span>
              <Badge>Connected</Badge>
            </div>
          </CardContent>
        </Card>

        <Card className="hover:shadow-lg transition-shadow cursor-pointer" onClick={() => window.location.href = '/dsp/bid-management'}>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Target className="h-5 w-5 text-purple-500" />
              Bid Strategies
            </CardTitle>
            <CardDescription>Configure bidding algorithms and pacing</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className="text-2xl font-bold">{stats?.activeBidStrategies}</span>
              <Badge variant="secondary">Active</Badge>
            </div>
          </CardContent>
        </Card>

        <Card className="hover:shadow-lg transition-shadow cursor-pointer" onClick={() => window.location.href = '/dsp/deals'}>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <DollarSign className="h-5 w-5 text-green-500" />
              Deals
            </CardTitle>
            <CardDescription>PMP, Preferred, and Guaranteed deals</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className="text-2xl font-bold">{stats?.activeDeals}</span>
              <Badge variant="outline">Negotiated</Badge>
            </div>
          </CardContent>
        </Card>

        <Card className="hover:shadow-lg transition-shadow cursor-pointer" onClick={() => window.location.href = '/dsp/supply-path-optimization'}>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5 text-orange-500" />
              Supply Path Optimization
            </CardTitle>
            <CardDescription>Monitor bid flow, latency, and costs</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className="text-2xl font-bold">SPO</span>
              <Badge variant="secondary">Analytics</Badge>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* RTB Performance */}
      <Card>
        <CardHeader>
          <CardTitle>Real-Time Bidding Performance</CardTitle>
          <CardDescription>Live bidding metrics and trends</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
            <div className="text-center p-4 bg-muted rounded-lg">
              <div className="text-3xl font-bold text-blue-500">
                {((stats?.totalBids || 0) / 24 / 3600).toFixed(0)}
              </div>
              <div className="text-sm text-muted-foreground">Avg QPS</div>
            </div>
            <div className="text-center p-4 bg-muted rounded-lg">
              <div className="text-3xl font-bold text-green-500">42ms</div>
              <div className="text-sm text-muted-foreground">Avg Response Time</div>
            </div>
            <div className="text-center p-4 bg-muted rounded-lg">
              <div className="text-3xl font-bold text-purple-500">99.7%</div>
              <div className="text-sm text-muted-foreground">Bid Rate</div>
            </div>
            <div className="text-center p-4 bg-muted rounded-lg">
              <div className="text-3xl font-bold text-yellow-500">0.3%</div>
              <div className="text-sm text-muted-foreground">Timeout Rate</div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
