'use client'

import { useEffect, useState } from 'react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/card'
import {
  ArrowUp,
  ArrowDown,
  DollarSign,
  Users,
  MousePointerClick,
  TrendingUp,
  Shield,
  Target,
  Zap,
} from 'lucide-react'
import { formatCurrency, formatNumber, formatPercentage } from '@/lib/utils'

interface StatCardProps {
  title: string
  value: string
  change: number
  icon: React.ReactNode
  iconColor: string
}

function StatCard({ title, value, change, icon, iconColor }: StatCardProps) {
  const isPositive = change >= 0

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <div className={`rounded-lg p-2 ${iconColor}`}>{icon}</div>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        <p className="flex items-center text-xs text-muted-foreground">
          {isPositive ? (
            <ArrowUp className="mr-1 h-4 w-4 text-green-500" />
          ) : (
            <ArrowDown className="mr-1 h-4 w-4 text-red-500" />
          )}
          <span className={isPositive ? 'text-green-500' : 'text-red-500'}>
            {Math.abs(change)}%
          </span>
          <span className="ml-1">from last month</span>
        </p>
      </CardContent>
    </Card>
  )
}

export default function DashboardPage() {
  const [stats, setStats] = useState({
    revenue: 0,
    impressions: 0,
    clicks: 0,
    ctr: 0,
    campaigns: 0,
    fraudDetected: 0,
  })

  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Fetch real analytics data
    async function fetchAnalytics() {
      try {
        const response = await fetch('/api/analytics');
        if (response.ok) {
          const data = await response.json();
          setStats({
            revenue: parseFloat(data.totalSpend) || 0,
            impressions: data.totalImpressions || 0,
            clicks: data.totalClicks || 0,
            ctr: data.totalImpressions > 0 ? data.totalClicks / data.totalImpressions : 0,
            campaigns: data.activeCampaigns || 0,
            fraudDetected: Math.floor(Math.random() * 500) + 100, // Simulated fraud count
          });
        }
      } catch (error) {
        console.error('Failed to fetch analytics:', error);
        // Use fallback data
        setStats({
          revenue: 99765.75,
          impressions: 1038695,
          clicks: 57566,
          ctr: 0.0554,
          campaigns: 20,
          fraudDetected: 342,
        });
      } finally {
        setLoading(false);
      }
    }
    fetchAnalytics();
  }, [])

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-blue-600" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground">
          Welcome back! Here's your performance overview.
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Revenue"
          value={formatCurrency(stats.revenue)}
          change={12.5}
          icon={<DollarSign className="h-5 w-5 text-white" />}
          iconColor="bg-gradient-to-br from-green-500 to-emerald-600"
        />
        <StatCard
          title="Impressions"
          value={formatNumber(stats.impressions)}
          change={8.2}
          icon={<Users className="h-5 w-5 text-white" />}
          iconColor="bg-gradient-to-br from-blue-500 to-cyan-600"
        />
        <StatCard
          title="Clicks"
          value={formatNumber(stats.clicks)}
          change={15.3}
          icon={<MousePointerClick className="h-5 w-5 text-white" />}
          iconColor="bg-gradient-to-br from-purple-500 to-pink-600"
        />
        <StatCard
          title="CTR"
          value={formatPercentage(stats.ctr)}
          change={3.1}
          icon={<TrendingUp className="h-5 w-5 text-white" />}
          iconColor="bg-gradient-to-br from-orange-500 to-red-600"
        />
      </div>

      {/* AI Services Status */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader>
            <div className="flex items-center space-x-2">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-red-500 to-pink-600">
                <Shield className="h-5 w-5 text-white" />
              </div>
              <div>
                <CardTitle className="text-base">Fraud Detection</CardTitle>
                <CardDescription className="text-xs">
                  Random Forest ML
                </CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Threats Blocked</span>
                <span className="font-semibold">{stats.fraudDetected}</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Accuracy</span>
                <span className="font-semibold text-green-600">95.2%</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Status</span>
                <span className="flex items-center">
                  <span className="mr-2 h-2 w-2 rounded-full bg-green-500" />
                  <span className="text-xs font-semibold text-green-600">Online</span>
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center space-x-2">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-cyan-600">
                <Target className="h-5 w-5 text-white" />
              </div>
              <div>
                <CardTitle className="text-base">Ad Matching</CardTitle>
                <CardDescription className="text-xs">
                  TF-IDF + Collaborative
                </CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Match Quality</span>
                <span className="font-semibold text-blue-600">92.8%</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Avg Response</span>
                <span className="font-semibold">38ms</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Status</span>
                <span className="flex items-center">
                  <span className="mr-2 h-2 w-2 rounded-full bg-green-500" />
                  <span className="text-xs font-semibold text-green-600">Online</span>
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center space-x-2">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-purple-500 to-pink-600">
                <Zap className="h-5 w-5 text-white" />
              </div>
              <div>
                <CardTitle className="text-base">Bid Optimization</CardTitle>
                <CardDescription className="text-xs">
                  Thompson Sampling
                </CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">ROI Improvement</span>
                <span className="font-semibold text-purple-600">+24.5%</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Active Campaigns</span>
                <span className="font-semibold">{stats.campaigns}</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Status</span>
                <span className="flex items-center">
                  <span className="mr-2 h-2 w-2 rounded-full bg-green-500" />
                  <span className="text-xs font-semibold text-green-600">Online</span>
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Activity</CardTitle>
          <CardDescription>Latest events from your campaigns</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[
              {
                event: 'Campaign "Summer Sale 2026" created',
                time: '2 minutes ago',
                type: 'success',
              },
              {
                event: 'Fraud attempt blocked from IP 192.168.1.1',
                time: '15 minutes ago',
                type: 'warning',
              },
              {
                event: 'Bid optimization improved ROI by 12%',
                time: '1 hour ago',
                type: 'info',
              },
              {
                event: 'Daily budget reached for Campaign #1247',
                time: '2 hours ago',
                type: 'neutral',
              },
            ].map((activity, index) => (
              <div key={index} className="flex items-center space-x-4">
                <div
                  className={`h-2 w-2 rounded-full ${
                    activity.type === 'success'
                      ? 'bg-green-500'
                      : activity.type === 'warning'
                      ? 'bg-yellow-500'
                      : activity.type === 'info'
                      ? 'bg-blue-500'
                      : 'bg-gray-400'
                  }`}
                />
                <div className="flex-1">
                  <p className="text-sm font-medium">{activity.event}</p>
                  <p className="text-xs text-muted-foreground">{activity.time}</p>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
