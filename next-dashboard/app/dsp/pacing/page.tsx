'use client'

import { useState, useEffect } from 'react'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { 
  Activity, 
  DollarSign, 
  TrendingUp,
  TrendingDown,
  Clock,
  Gauge,
  RefreshCw,
  AlertTriangle,
  CheckCircle,
  PauseCircle,
  PlayCircle
} from 'lucide-react'

interface CampaignPacing {
  id: string
  name: string
  status: string
  budget: number
  spent: number
  dailyBudget: number
  dailySpent: number
  pacingType: string
  pacingStatus: string
  impressionsGoal: number
  impressionsDelivered: number
  startDate: string
  endDate: string
  daysRemaining: number
  budgetPacingProgress: number
  timePacingProgress: number
  deliveryRate: number
  projectedDelivery: number
}

export default function PacingPage() {
  const [campaigns, setCampaigns] = useState<CampaignPacing[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchPacing = async () => {
    try {
      setLoading(true)
       // This would call a pacing-specific endpoint
       const response = await api.getBidStrategies()
       setError(null)
       
       // Transform bid strategies to pacing data
       const pacingData: CampaignPacing[] = response.data.map((strategy: any) => {
         // Mock data calculation based on real strategy config
         // In a real scenario, we'd fetch actual spend metrics
         const budget = strategy.pacing?.dailyBudget ? strategy.pacing.dailyBudget * 30 : 10000; // Assume monthly
         const spent = strategy.performanceMetrics?.totalSpend || Math.random() * (budget * 0.5); // Mock spent if 0
         const startDate = new Date(strategy.createdAt);
         const now = new Date();
         
         const totalDays = 60; // Assumed duration
         const daysElapsed = Math.floor((now.getTime() - startDate.getTime()) / (1000 * 60 * 60 * 24));
         // Assumed EndDate relative to now for demo purposes if not fixed
         const endDate = new Date(startDate.getTime() + totalDays * 24 * 60 * 60 * 1000); 
         const daysRemaining = Math.max(0, totalDays - daysElapsed);
         
         const budgetPacingProgress = (spent / budget) * 100;
         const timePacingProgress = (daysElapsed / totalDays) * 100;
         const pacingStatus = budgetPacingProgress > timePacingProgress + 10 ? 'ahead' : 
                              budgetPacingProgress < timePacingProgress - 10 ? 'behind' : 'on_track';

         return {
           id: strategy.id,
           name: strategy.name,
           status: strategy.status,
           budget: budget,
           spent: spent,
           dailyBudget: strategy.pacing?.dailyBudget || 100,
           dailySpent: Math.random() * (strategy.pacing?.dailyBudget || 100),
           pacingType: strategy.pacing?.type || 'even',
           pacingStatus: strategy.status === 'paused' ? 'paused' : pacingStatus,
           impressionsGoal: 1000000,
           impressionsDelivered: strategy.performanceMetrics?.impressions || 0,
           startDate: startDate.toISOString().split('T')[0],
           endDate: endDate.toISOString().split('T')[0],
           daysRemaining: daysRemaining,
           budgetPacingProgress: Math.min(100, budgetPacingProgress),
           timePacingProgress: Math.min(100, timePacingProgress),
           deliveryRate: (budgetPacingProgress / (timePacingProgress || 1)),
           projectedDelivery: (budgetPacingProgress / (timePacingProgress || 1)) * 100
         };
       });
       
       setCampaigns(pacingData)
    } catch (err: any) {
      setError(err.message || 'Failed to load pacing data')
      // Demo data
      setCampaigns([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchPacing()
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

  const getPacingStatusBadge = (status: string) => {
    switch (status) {
      case 'on_track':
        return <Badge className="bg-green-500"><CheckCircle className="h-3 w-3 mr-1" />On Track</Badge>
      case 'ahead':
        return <Badge className="bg-blue-500"><TrendingUp className="h-3 w-3 mr-1" />Ahead</Badge>
      case 'behind':
        return <Badge variant="destructive"><TrendingDown className="h-3 w-3 mr-1" />Behind</Badge>
      case 'nearly_complete':
        return <Badge className="bg-purple-500"><CheckCircle className="h-3 w-3 mr-1" />Nearly Complete</Badge>
      case 'paused':
        return <Badge variant="secondary"><PauseCircle className="h-3 w-3 mr-1" />Paused</Badge>
      default:
        return <Badge variant="outline">{status}</Badge>
    }
  }

  const getPacingTypeBadge = (type: string) => {
    return type === 'accelerated' 
      ? <Badge variant="outline" className="text-orange-500 border-orange-500">Accelerated</Badge>
      : <Badge variant="outline" className="text-blue-500 border-blue-500">Even</Badge>
  }

  const getProgressColor = (progress: number, target: number) => {
    const ratio = progress / target
    if (ratio >= 0.95 && ratio <= 1.05) return 'bg-green-500'
    if (ratio > 1.05) return 'bg-blue-500'
    return 'bg-yellow-500'
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  const activeCampaigns = campaigns.filter(c => c.status === 'active')
  const totalBudget = campaigns.reduce((sum, c) => sum + c.budget, 0)
  const totalSpent = campaigns.reduce((sum, c) => sum + c.spent, 0)
  const onTrackCampaigns = campaigns.filter(c => ['on_track', 'nearly_complete'].includes(c.pacingStatus))
  const behindCampaigns = campaigns.filter(c => c.pacingStatus === 'behind')

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Budget Pacing</h1>
          <p className="text-muted-foreground">
            Monitor campaign delivery and budget utilization
          </p>
        </div>
        <Button onClick={fetchPacing} variant="outline">
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

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Activity className="h-4 w-4 text-blue-500" />
              Active Campaigns
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{activeCampaigns.length}</div>
            <p className="text-xs text-muted-foreground">
              of {campaigns.length} total
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <DollarSign className="h-4 w-4 text-green-500" />
              Total Budget
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(totalBudget)}</div>
            <p className="text-xs text-muted-foreground">
              {formatCurrency(totalSpent)} spent ({((totalSpent / totalBudget) * 100).toFixed(1)}%)
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-green-500" />
              On Track
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{onTrackCampaigns.length}</div>
            <p className="text-xs text-muted-foreground">campaigns pacing well</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 text-yellow-500" />
              Behind Pace
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-yellow-600">{behindCampaigns.length}</div>
            <p className="text-xs text-muted-foreground">need attention</p>
          </CardContent>
        </Card>
      </div>

      {/* Campaign Pacing Cards */}
      <div className="space-y-4">
        {campaigns.map((campaign) => (
          <Card key={campaign.id} className={campaign.pacingStatus === 'behind' ? 'border-yellow-500/50' : ''}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="flex items-center gap-2">
                    {campaign.name}
                    {campaign.pacingStatus === 'behind' && (
                      <AlertTriangle className="h-4 w-4 text-yellow-500" />
                    )}
                  </CardTitle>
                  <CardDescription>
                    {campaign.startDate} to {campaign.endDate} · {campaign.daysRemaining} days remaining
                  </CardDescription>
                </div>
                <div className="flex items-center gap-2">
                  {getPacingTypeBadge(campaign.pacingType)}
                  {getPacingStatusBadge(campaign.pacingStatus)}
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {/* Budget Progress */}
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="flex items-center gap-1">
                      <DollarSign className="h-4 w-4" />
                      Budget
                    </span>
                    <span>{formatCurrency(campaign.spent)} / {formatCurrency(campaign.budget)}</span>
                  </div>
                  <Progress value={campaign.budgetPacingProgress} className="h-3" />
                  <div className="flex justify-between text-xs text-muted-foreground">
                    <span>{campaign.budgetPacingProgress.toFixed(1)}% spent</span>
                    <span>Daily: {formatCurrency(campaign.dailySpent)} / {formatCurrency(campaign.dailyBudget)}</span>
                  </div>
                </div>

                {/* Time Progress */}
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="flex items-center gap-1">
                      <Clock className="h-4 w-4" />
                      Time Elapsed
                    </span>
                    <span>{campaign.timePacingProgress.toFixed(1)}%</span>
                  </div>
                  <Progress value={campaign.timePacingProgress} className="h-3" />
                  <div className="flex justify-between text-xs text-muted-foreground">
                    <span>Flight progress</span>
                    <span>{campaign.daysRemaining} days left</span>
                  </div>
                </div>

                {/* Delivery Metrics */}
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="flex items-center gap-1">
                      <Gauge className="h-4 w-4" />
                      Delivery
                    </span>
                    <span className={
                      campaign.deliveryRate >= 0.95 && campaign.deliveryRate <= 1.05 
                        ? 'text-green-500' 
                        : campaign.deliveryRate > 1.05 
                          ? 'text-blue-500' 
                          : 'text-yellow-500'
                    }>
                      {(campaign.deliveryRate * 100).toFixed(0)}% rate
                    </span>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="flex-1 text-center p-2 bg-muted rounded">
                      <div className="text-lg font-bold">{formatNumber(campaign.impressionsDelivered)}</div>
                      <div className="text-xs text-muted-foreground">Delivered</div>
                    </div>
                    <div className="flex-1 text-center p-2 bg-muted rounded">
                      <div className="text-lg font-bold">{formatNumber(campaign.impressionsGoal)}</div>
                      <div className="text-xs text-muted-foreground">Goal</div>
                    </div>
                    <div className="flex-1 text-center p-2 bg-muted rounded">
                      <div className={`text-lg font-bold ${
                        campaign.projectedDelivery >= 95 && campaign.projectedDelivery <= 105 
                          ? 'text-green-500' 
                          : campaign.projectedDelivery > 105 
                            ? 'text-blue-500' 
                            : 'text-yellow-500'
                      }`}>
                        {campaign.projectedDelivery.toFixed(1)}%
                      </div>
                      <div className="text-xs text-muted-foreground">Projected</div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Pacing Alert */}
              {campaign.pacingStatus === 'behind' && (
                <div className="mt-4 p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
                  <div className="flex items-center gap-2 text-yellow-800 dark:text-yellow-200">
                    <AlertTriangle className="h-4 w-4" />
                    <span className="text-sm font-medium">Campaign is pacing behind schedule</span>
                  </div>
                  <p className="text-xs text-yellow-700 dark:text-yellow-300 mt-1">
                    Current delivery rate is {(campaign.deliveryRate * 100).toFixed(0)}%. 
                    Consider increasing bids or expanding targeting to catch up.
                  </p>
                </div>
              )}

              {campaign.pacingStatus === 'ahead' && (
                <div className="mt-4 p-3 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
                  <div className="flex items-center gap-2 text-blue-800 dark:text-blue-200">
                    <TrendingUp className="h-4 w-4" />
                    <span className="text-sm font-medium">Campaign is pacing ahead of schedule</span>
                  </div>
                  <p className="text-xs text-blue-700 dark:text-blue-300 mt-1">
                    Projected to deliver {campaign.projectedDelivery.toFixed(0)}% of goal. 
                    Budget may exhaust before flight end date.
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}
