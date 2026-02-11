'use client'

import { useState, useEffect } from 'react'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { 
  Activity, 
  Zap, 
  Clock, 
  TrendingUp,
  TrendingDown,
  Target,
  DollarSign,
  RefreshCw,
  BarChart3,
  AlertTriangle,
  CheckCircle
} from 'lucide-react'

interface RTBAnalytics {
  totalRequests: number
  totalBids: number
  totalWins: number
  totalTimeouts: number
  totalErrors: number
  bidRate: number
  winRate: number
  avgResponseTime: number
  avgBidPrice: number
  avgWinPrice: number
  qps: number
  p50Latency: number
  p95Latency: number
  p99Latency: number
  partnerMetrics: {
    partnerId: string
    partnerName: string
    requests: number
    bids: number
    wins: number
    winRate: number
    avgLatency: number
    spend: number
  }[]
  hourlyTrend: {
    hour: string
    requests: number
    bids: number
    wins: number
  }[]
}

export default function RTBAnalyticsPage() {
  const [analytics, setAnalytics] = useState<RTBAnalytics | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [refreshing, setRefreshing] = useState(false)

  const fetchAnalytics = async () => {
    try {
      setRefreshing(true)
      const response = await api.getRTBAnalytics()
      setAnalytics(response.data)
      setError(null)
    } catch (err: any) {
      setError(err.message || 'Failed to load RTB analytics')
      // Demo data
      setAnalytics({
        totalRequests: 15847293,
        totalBids: 14582847,
        totalWins: 2847584,
        totalTimeouts: 47382,
        totalErrors: 12847,
        bidRate: 92.02,
        winRate: 19.52,
        avgResponseTime: 42,
        avgBidPrice: 2.85,
        avgWinPrice: 3.42,
        qps: 4587,
        p50Latency: 35,
        p95Latency: 85,
        p99Latency: 145,
        partnerMetrics: [
          { partnerId: '1', partnerName: 'Google Ad Exchange', requests: 5847293, bids: 5482847, wins: 1247584, winRate: 22.75, avgLatency: 38, spend: 45283.47 },
          { partnerId: '2', partnerName: 'AppNexus', requests: 4847293, bids: 4382847, wins: 847584, winRate: 19.34, avgLatency: 45, spend: 32847.29 },
          { partnerId: '3', partnerName: 'Rubicon Project', requests: 3847293, bids: 3582847, wins: 547584, winRate: 15.29, avgLatency: 52, spend: 21847.18 },
          { partnerId: '4', partnerName: 'PubMatic', requests: 1305414, bids: 1134306, wins: 204832, winRate: 18.06, avgLatency: 48, spend: 8473.28 }
        ],
        hourlyTrend: [
          { hour: '00:00', requests: 450000, bids: 420000, wins: 85000 },
          { hour: '04:00', requests: 320000, bids: 295000, wins: 58000 },
          { hour: '08:00', requests: 680000, bids: 640000, wins: 125000 },
          { hour: '12:00', requests: 920000, bids: 865000, wins: 175000 },
          { hour: '16:00', requests: 850000, bids: 795000, wins: 162000 },
          { hour: '20:00', requests: 780000, bids: 725000, wins: 148000 }
        ]
      })
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    fetchAnalytics()
    const interval = setInterval(fetchAnalytics, 30000) // Auto-refresh every 30s
    return () => clearInterval(interval)
  }, [])

  const formatNumber = (num: number) => {
    if (num >= 1000000) return (num / 1000000).toFixed(2) + 'M'
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
    return num.toLocaleString()
  }

  const formatCurrency = (num: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
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
          <h1 className="text-3xl font-bold">RTB Analytics</h1>
          <p className="text-muted-foreground">
            Real-Time Bidding performance metrics and latency analysis
          </p>
        </div>
        <div className="flex gap-2 items-center">
          {refreshing && (
            <Badge variant="outline" className="animate-pulse">
              <RefreshCw className="h-3 w-3 mr-1 animate-spin" />
              Updating...
            </Badge>
          )}
          <Button onClick={fetchAnalytics} variant="outline" disabled={refreshing}>
            <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {error && (
        <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
          <p className="text-yellow-800 dark:text-yellow-200 text-sm">
            Using demo data - {error}
          </p>
        </div>
      )}

      {/* Live QPS Indicator */}
      <Card className="bg-gradient-to-r from-blue-500/10 to-purple-500/10 border-blue-500/30">
        <CardContent className="pt-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="p-3 bg-blue-500 rounded-lg">
                <Zap className="h-6 w-6 text-white" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Current QPS</p>
                <p className="text-4xl font-bold">{analytics?.qps?.toLocaleString()}</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <span className="h-3 w-3 bg-green-500 rounded-full animate-pulse"></span>
              <span className="text-sm text-green-500">Live</span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Main Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
            <Activity className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatNumber(analytics?.totalRequests || 0)}</div>
            <p className="text-xs text-muted-foreground">Bid requests received</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Bid Rate</CardTitle>
            <Target className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{analytics?.bidRate?.toFixed(2)}%</div>
            <div className="flex items-center text-xs text-green-500">
              <TrendingUp className="h-3 w-3 mr-1" />
              {formatNumber(analytics?.totalBids || 0)} bids placed
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Win Rate</CardTitle>
            <CheckCircle className="h-4 w-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{analytics?.winRate?.toFixed(2)}%</div>
            <div className="flex items-center text-xs text-purple-500">
              {formatNumber(analytics?.totalWins || 0)} wins
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Avg Response</CardTitle>
            <Clock className="h-4 w-4 text-orange-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{analytics?.avgResponseTime}ms</div>
            <p className="text-xs text-muted-foreground">Avg bid response time</p>
          </CardContent>
        </Card>
      </div>

      {/* Latency Percentiles */}
      <Card>
        <CardHeader>
          <CardTitle>Latency Distribution</CardTitle>
          <CardDescription>Response time percentiles across all bid requests</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-6">
            <div className="text-center p-6 bg-green-50 dark:bg-green-900/20 rounded-lg">
              <div className="text-4xl font-bold text-green-600">{analytics?.p50Latency}ms</div>
              <div className="text-sm text-muted-foreground">P50 Latency</div>
              <div className="text-xs text-green-600">Median response</div>
            </div>
            <div className="text-center p-6 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
              <div className="text-4xl font-bold text-yellow-600">{analytics?.p95Latency}ms</div>
              <div className="text-sm text-muted-foreground">P95 Latency</div>
              <div className="text-xs text-yellow-600">95th percentile</div>
            </div>
            <div className="text-center p-6 bg-red-50 dark:bg-red-900/20 rounded-lg">
              <div className="text-4xl font-bold text-red-600">{analytics?.p99Latency}ms</div>
              <div className="text-sm text-muted-foreground">P99 Latency</div>
              <div className="text-xs text-red-600">99th percentile</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Error Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-yellow-500" />
              Timeouts
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div>
                <div className="text-3xl font-bold">{formatNumber(analytics?.totalTimeouts || 0)}</div>
                <p className="text-sm text-muted-foreground">
                  {((analytics?.totalTimeouts || 0) / (analytics?.totalRequests || 1) * 100).toFixed(2)}% of requests
                </p>
              </div>
              <div className={`p-3 rounded-full ${
                ((analytics?.totalTimeouts || 0) / (analytics?.totalRequests || 1) * 100) < 1 
                  ? 'bg-green-100 text-green-600' 
                  : 'bg-yellow-100 text-yellow-600'
              }`}>
                <Clock className="h-6 w-6" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-red-500" />
              Errors
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div>
                <div className="text-3xl font-bold">{formatNumber(analytics?.totalErrors || 0)}</div>
                <p className="text-sm text-muted-foreground">
                  {((analytics?.totalErrors || 0) / (analytics?.totalRequests || 1) * 100).toFixed(3)}% error rate
                </p>
              </div>
              <div className={`p-3 rounded-full ${
                ((analytics?.totalErrors || 0) / (analytics?.totalRequests || 1) * 100) < 0.5 
                  ? 'bg-green-100 text-green-600' 
                  : 'bg-red-100 text-red-600'
              }`}>
                <AlertTriangle className="h-6 w-6" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Pricing Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <DollarSign className="h-5 w-5 text-green-500" />
              Average Bid Price
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-4xl font-bold">{formatCurrency(analytics?.avgBidPrice || 0)}</div>
            <p className="text-sm text-muted-foreground">CPM across all bids</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <DollarSign className="h-5 w-5 text-purple-500" />
              Average Win Price
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-4xl font-bold">{formatCurrency(analytics?.avgWinPrice || 0)}</div>
            <p className="text-sm text-muted-foreground">CPM for winning bids</p>
          </CardContent>
        </Card>
      </div>

      {/* Partner Performance */}
      <Card>
        <CardHeader>
          <CardTitle>Supply Partner Performance</CardTitle>
          <CardDescription>RTB metrics by supply partner</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {analytics?.partnerMetrics.map((partner) => (
              <div key={partner.partnerId} className="flex items-center justify-between p-4 border rounded-lg">
                <div className="flex-1">
                  <div className="font-medium">{partner.partnerName}</div>
                  <div className="text-sm text-muted-foreground">
                    {formatNumber(partner.requests)} requests
                  </div>
                </div>
                <div className="flex items-center gap-8">
                  <div className="text-center">
                    <div className="text-lg font-bold">{partner.winRate.toFixed(1)}%</div>
                    <div className="text-xs text-muted-foreground">Win Rate</div>
                  </div>
                  <div className="text-center">
                    <div className="text-lg font-bold">{partner.avgLatency}ms</div>
                    <div className="text-xs text-muted-foreground">Latency</div>
                  </div>
                  <div className="text-center">
                    <div className="text-lg font-bold text-green-600">{formatCurrency(partner.spend)}</div>
                    <div className="text-xs text-muted-foreground">Spend</div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Hourly Trend */}
      <Card>
        <CardHeader>
          <CardTitle>Hourly Volume Trend</CardTitle>
          <CardDescription>Bid request volume throughout the day</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-6 gap-4">
            {analytics?.hourlyTrend.map((hour) => (
              <div key={hour.hour} className="text-center p-4 bg-muted rounded-lg">
                <div className="text-xs text-muted-foreground mb-2">{hour.hour}</div>
                <div className="text-lg font-bold">{formatNumber(hour.requests)}</div>
                <div className="text-xs text-green-500">{formatNumber(hour.wins)} wins</div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
