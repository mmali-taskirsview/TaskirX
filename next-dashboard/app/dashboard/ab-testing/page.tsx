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
  FlaskConical,
  Plus,
  Play,
  Pause,
  CheckCircle2,
  XCircle,
  TrendingUp,
  TrendingDown,
  BarChart3,
  Users,
  Target,
  Clock,
  Award,
  ArrowRight,
  RefreshCw,
} from 'lucide-react'
import { formatCurrency, formatNumber, formatPercentage } from '@/lib/utils'

// Types
interface Variant {
  id: string
  name: string
  creative: string
  impressions: number
  clicks: number
  conversions: number
  revenue: number
  ctr: number
  cvr: number
  confidence: number
  isWinner?: boolean
  isControl?: boolean
}

interface ABTest {
  id: string
  name: string
  campaign: string
  status: 'running' | 'completed' | 'paused' | 'draft'
  startDate: string
  endDate?: string
  trafficSplit: number
  variants: Variant[]
  metric: string
  minSampleSize: number
  currentSampleSize: number
}

// Sample data
const abTests: ABTest[] = [
  {
    id: '1',
    name: 'Hero Image Test - Summer Campaign',
    campaign: 'Summer Sale 2026',
    status: 'running',
    startDate: '2026-01-15',
    trafficSplit: 50,
    metric: 'Conversion Rate',
    minSampleSize: 10000,
    currentSampleSize: 7823,
    variants: [
      {
        id: 'a',
        name: 'Control - Blue Background',
        creative: 'creative_blue_hero.jpg',
        impressions: 125000,
        clicks: 4500,
        conversions: 180,
        revenue: 8640,
        ctr: 3.6,
        cvr: 4.0,
        confidence: 0,
        isControl: true,
      },
      {
        id: 'b',
        name: 'Variant B - Lifestyle Image',
        creative: 'creative_lifestyle.jpg',
        impressions: 125000,
        clicks: 5125,
        conversions: 225,
        revenue: 10800,
        ctr: 4.1,
        cvr: 4.4,
        confidence: 87,
      },
    ],
  },
  {
    id: '2',
    name: 'CTA Button Color Test',
    campaign: 'Mobile App Promotion',
    status: 'completed',
    startDate: '2026-01-01',
    endDate: '2026-01-14',
    trafficSplit: 50,
    metric: 'Click-Through Rate',
    minSampleSize: 15000,
    currentSampleSize: 15000,
    variants: [
      {
        id: 'a',
        name: 'Control - Green CTA',
        creative: 'cta_green.png',
        impressions: 180000,
        clicks: 5400,
        conversions: 270,
        revenue: 12960,
        ctr: 3.0,
        cvr: 5.0,
        confidence: 0,
        isControl: true,
      },
      {
        id: 'b',
        name: 'Variant B - Orange CTA',
        creative: 'cta_orange.png',
        impressions: 180000,
        clicks: 6300,
        conversions: 315,
        revenue: 15120,
        ctr: 3.5,
        cvr: 5.0,
        confidence: 95,
        isWinner: true,
      },
    ],
  },
  {
    id: '3',
    name: 'Video Length Test',
    campaign: 'Brand Awareness Q1',
    status: 'running',
    startDate: '2026-01-18',
    trafficSplit: 33,
    metric: 'Video Completion Rate',
    minSampleSize: 20000,
    currentSampleSize: 12450,
    variants: [
      {
        id: 'a',
        name: 'Control - 15 Second',
        creative: 'video_15s.mp4',
        impressions: 95000,
        clicks: 3800,
        conversions: 152,
        revenue: 7296,
        ctr: 4.0,
        cvr: 4.0,
        confidence: 0,
        isControl: true,
      },
      {
        id: 'b',
        name: 'Variant B - 30 Second',
        creative: 'video_30s.mp4',
        impressions: 95000,
        clicks: 3325,
        conversions: 166,
        revenue: 7968,
        ctr: 3.5,
        cvr: 5.0,
        confidence: 72,
      },
      {
        id: 'c',
        name: 'Variant C - 6 Second',
        creative: 'video_6s.mp4',
        impressions: 95000,
        clicks: 4275,
        conversions: 128,
        revenue: 6144,
        ctr: 4.5,
        cvr: 3.0,
        confidence: 45,
      },
    ],
  },
  {
    id: '4',
    name: 'Headline Copy Test',
    campaign: 'New Product Launch',
    status: 'paused',
    startDate: '2026-01-10',
    trafficSplit: 50,
    metric: 'Conversion Rate',
    minSampleSize: 10000,
    currentSampleSize: 4500,
    variants: [
      {
        id: 'a',
        name: 'Control - Feature-Focused',
        creative: 'headline_features.html',
        impressions: 45000,
        clicks: 1575,
        conversions: 63,
        revenue: 3024,
        ctr: 3.5,
        cvr: 4.0,
        confidence: 0,
        isControl: true,
      },
      {
        id: 'b',
        name: 'Variant B - Benefit-Focused',
        creative: 'headline_benefits.html',
        impressions: 45000,
        clicks: 1620,
        conversions: 68,
        revenue: 3264,
        ctr: 3.6,
        cvr: 4.2,
        confidence: 42,
      },
    ],
  },
]

export default function ABTestingPage() {
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [selectedTest, setSelectedTest] = useState<ABTest | null>(null)
  const [filter, setFilter] = useState<'all' | 'running' | 'completed' | 'paused'>('all')

  const filteredTests = filter === 'all' 
    ? abTests 
    : abTests.filter(test => test.status === filter)

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'running':
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900/30 dark:text-green-400">
            <span className="h-1.5 w-1.5 rounded-full bg-green-500 animate-pulse" />
            Running
          </span>
        )
      case 'completed':
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
            <CheckCircle2 className="h-3 w-3" />
            Completed
          </span>
        )
      case 'paused':
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-yellow-100 px-2.5 py-0.5 text-xs font-medium text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400">
            <Pause className="h-3 w-3" />
            Paused
          </span>
        )
      default:
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-700 dark:bg-gray-800 dark:text-gray-400">
            Draft
          </span>
        )
    }
  }

  const calculateLift = (control: Variant, variant: Variant, metric: string) => {
    if (metric === 'Click-Through Rate') {
      return ((variant.ctr - control.ctr) / control.ctr * 100).toFixed(1)
    }
    return ((variant.cvr - control.cvr) / control.cvr * 100).toFixed(1)
  }

  // Summary stats
  const runningTests = abTests.filter(t => t.status === 'running').length
  const completedTests = abTests.filter(t => t.status === 'completed').length
  const totalImpressions = abTests.reduce((sum, test) => 
    sum + test.variants.reduce((vSum, v) => vSum + v.impressions, 0), 0)
  const avgConfidence = completedTests > 0 
    ? abTests.filter(t => t.status === 'completed')
        .reduce((sum, t) => sum + Math.max(...t.variants.map(v => v.confidence)), 0) / completedTests
    : 0

  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">A/B Testing</h1>
          <p className="text-muted-foreground">
            Test and optimize your creative assets with statistical confidence
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 px-4 py-2 text-sm font-medium text-white shadow-lg hover:opacity-90 transition-opacity"
        >
          <Plus className="h-4 w-4" />
          Create New Test
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Running Tests</p>
                <p className="text-2xl font-bold">{runningTests}</p>
              </div>
              <div className="rounded-full bg-green-100 p-3 dark:bg-green-900/30">
                <Play className="h-5 w-5 text-green-600" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Completed Tests</p>
                <p className="text-2xl font-bold">{completedTests}</p>
              </div>
              <div className="rounded-full bg-blue-100 p-3 dark:bg-blue-900/30">
                <CheckCircle2 className="h-5 w-5 text-blue-600" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Impressions</p>
                <p className="text-2xl font-bold">{formatNumber(totalImpressions)}</p>
              </div>
              <div className="rounded-full bg-purple-100 p-3 dark:bg-purple-900/30">
                <Users className="h-5 w-5 text-purple-600" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Avg Confidence</p>
                <p className="text-2xl font-bold">{avgConfidence.toFixed(0)}%</p>
              </div>
              <div className="rounded-full bg-orange-100 p-3 dark:bg-orange-900/30">
                <Target className="h-5 w-5 text-orange-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Filter Tabs */}
      <div className="flex gap-2">
        {(['all', 'running', 'completed', 'paused'] as const).map((status) => (
          <button
            key={status}
            onClick={() => setFilter(status)}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              filter === status
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700'
            }`}
          >
            {status.charAt(0).toUpperCase() + status.slice(1)}
          </button>
        ))}
      </div>

      {/* Tests List */}
      <div className="space-y-4">
        {filteredTests.map((test) => {
          const control = test.variants.find(v => v.isControl)!
          const progress = (test.currentSampleSize / test.minSampleSize) * 100

          return (
            <Card key={test.id} className="overflow-hidden">
              <CardHeader className="border-b bg-muted/30">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="rounded-lg bg-gradient-to-br from-purple-500 to-blue-600 p-2">
                      <FlaskConical className="h-5 w-5 text-white" />
                    </div>
                    <div>
                      <CardTitle className="text-lg">{test.name}</CardTitle>
                      <CardDescription className="flex items-center gap-2">
                        Campaign: {test.campaign}
                        <span className="text-muted-foreground">•</span>
                        Testing: {test.metric}
                      </CardDescription>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    {getStatusBadge(test.status)}
                    {test.status === 'running' && (
                      <button className="rounded-lg border p-2 hover:bg-muted transition-colors">
                        <Pause className="h-4 w-4" />
                      </button>
                    )}
                    {test.status === 'paused' && (
                      <button className="rounded-lg border p-2 hover:bg-muted transition-colors">
                        <Play className="h-4 w-4" />
                      </button>
                    )}
                  </div>
                </div>
              </CardHeader>
              <CardContent className="pt-4">
                {/* Progress Bar */}
                <div className="mb-4">
                  <div className="flex items-center justify-between text-sm mb-1">
                    <span className="text-muted-foreground">Sample Size Progress</span>
                    <span className="font-medium">
                      {formatNumber(test.currentSampleSize)} / {formatNumber(test.minSampleSize)}
                    </span>
                  </div>
                  <div className="h-2 w-full rounded-full bg-gray-200 dark:bg-gray-700">
                    <div 
                      className={`h-full rounded-full transition-all ${
                        progress >= 100 ? 'bg-green-500' : 'bg-blue-500'
                      }`}
                      style={{ width: `${Math.min(progress, 100)}%` }}
                    />
                  </div>
                </div>

                {/* Variants Comparison */}
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                  {test.variants.map((variant) => {
                    const lift = !variant.isControl ? calculateLift(control, variant, test.metric) : null
                    const isPositiveLift = lift && parseFloat(lift) > 0

                    return (
                      <div 
                        key={variant.id}
                        className={`rounded-lg border p-4 ${
                          variant.isWinner 
                            ? 'border-green-500 bg-green-50 dark:bg-green-900/10' 
                            : variant.isControl 
                              ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/10'
                              : ''
                        }`}
                      >
                        <div className="flex items-center justify-between mb-3">
                          <div className="flex items-center gap-2">
                            <span className="font-medium">{variant.name}</span>
                            {variant.isControl && (
                              <span className="rounded-full bg-blue-100 px-2 py-0.5 text-xs text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
                                Control
                              </span>
                            )}
                            {variant.isWinner && (
                              <Award className="h-4 w-4 text-green-500" />
                            )}
                          </div>
                        </div>

                        <div className="space-y-2 text-sm">
                          <div className="flex justify-between">
                            <span className="text-muted-foreground">Impressions</span>
                            <span className="font-medium">{formatNumber(variant.impressions)}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-muted-foreground">Clicks</span>
                            <span className="font-medium">{formatNumber(variant.clicks)}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-muted-foreground">CTR</span>
                            <span className="font-medium">{variant.ctr}%</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-muted-foreground">Conversions</span>
                            <span className="font-medium">{formatNumber(variant.conversions)}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-muted-foreground">Revenue</span>
                            <span className="font-medium">{formatCurrency(variant.revenue)}</span>
                          </div>
                        </div>

                        {!variant.isControl && (
                          <div className="mt-3 pt-3 border-t">
                            <div className="flex items-center justify-between">
                              <span className="text-sm text-muted-foreground">Lift vs Control</span>
                              <span className={`flex items-center gap-1 font-semibold ${
                                isPositiveLift ? 'text-green-600' : 'text-red-600'
                              }`}>
                                {isPositiveLift ? (
                                  <TrendingUp className="h-4 w-4" />
                                ) : (
                                  <TrendingDown className="h-4 w-4" />
                                )}
                                {lift}%
                              </span>
                            </div>
                            <div className="flex items-center justify-between mt-1">
                              <span className="text-sm text-muted-foreground">Confidence</span>
                              <span className={`font-semibold ${
                                variant.confidence >= 95 ? 'text-green-600' :
                                variant.confidence >= 80 ? 'text-yellow-600' : 'text-gray-600'
                              }`}>
                                {variant.confidence}%
                              </span>
                            </div>
                            {variant.confidence >= 95 && (
                              <div className="mt-2 rounded-lg bg-green-100 px-3 py-2 text-xs text-green-700 dark:bg-green-900/30 dark:text-green-400">
                                <CheckCircle2 className="inline h-3 w-3 mr-1" />
                                Statistically significant
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    )
                  })}
                </div>

                {/* Test Info Footer */}
                <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
                  <div className="flex items-center gap-4">
                    <span className="flex items-center gap-1">
                      <Clock className="h-4 w-4" />
                      Started: {test.startDate}
                    </span>
                    {test.endDate && (
                      <span className="flex items-center gap-1">
                        <CheckCircle2 className="h-4 w-4" />
                        Ended: {test.endDate}
                      </span>
                    )}
                  </div>
                  <button className="flex items-center gap-1 text-blue-600 hover:underline">
                    View Details
                    <ArrowRight className="h-4 w-4" />
                  </button>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {/* Create Test Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-2xl rounded-lg bg-white p-6 shadow-xl dark:bg-gray-900">
            <h2 className="text-xl font-bold mb-4">Create New A/B Test</h2>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">Test Name</label>
                <input
                  type="text"
                  placeholder="e.g., Hero Image Test - Summer Campaign"
                  className="w-full rounded-lg border px-3 py-2"
                />
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium mb-1">Campaign</label>
                  <select className="w-full rounded-lg border px-3 py-2">
                    <option>Summer Sale 2026</option>
                    <option>New Product Launch</option>
                    <option>Brand Awareness Q1</option>
                    <option>Mobile App Promotion</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Primary Metric</label>
                  <select className="w-full rounded-lg border px-3 py-2">
                    <option>Conversion Rate</option>
                    <option>Click-Through Rate</option>
                    <option>Revenue per Impression</option>
                    <option>Video Completion Rate</option>
                  </select>
                </div>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium mb-1">Traffic Split</label>
                  <select className="w-full rounded-lg border px-3 py-2">
                    <option>50/50</option>
                    <option>70/30</option>
                    <option>80/20</option>
                    <option>33/33/34 (3 variants)</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Min Sample Size</label>
                  <input
                    type="number"
                    defaultValue={10000}
                    className="w-full rounded-lg border px-3 py-2"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Confidence Level Required</label>
                <select className="w-full rounded-lg border px-3 py-2">
                  <option>95% (Recommended)</option>
                  <option>90%</option>
                  <option>99%</option>
                </select>
              </div>

              <div className="border-t pt-4">
                <h3 className="font-medium mb-3">Variants</h3>
                <div className="space-y-3">
                  <div className="rounded-lg border p-3">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium">Control (A)</span>
                      <span className="text-xs text-blue-600">Baseline</span>
                    </div>
                    <input
                      type="text"
                      placeholder="Creative asset or description"
                      className="w-full rounded border px-3 py-2 text-sm"
                    />
                  </div>
                  <div className="rounded-lg border p-3">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium">Variant B</span>
                    </div>
                    <input
                      type="text"
                      placeholder="Creative asset or description"
                      className="w-full rounded border px-3 py-2 text-sm"
                    />
                  </div>
                </div>
                <button className="mt-2 text-sm text-blue-600 hover:underline">
                  + Add another variant
                </button>
              </div>
            </div>

            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowCreateModal(false)}
                className="rounded-lg border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() => setShowCreateModal(false)}
                className="rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 px-4 py-2 text-sm font-medium text-white hover:opacity-90 transition-opacity"
              >
                Create Test
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
