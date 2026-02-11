'use client'

import { useState, useMemo } from 'react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/card'
import {
  TrendingUp,
  TrendingDown,
  DollarSign,
  Users,
  MousePointerClick,
  Target,
  Calendar,
  Download,
  ChevronDown,
  FileSpreadsheet,
  Check,
  Globe,
  MapPin,
  BarChart2,
  Clock,
} from 'lucide-react'
import { Button } from '@/components/button'
import { formatCurrency, formatNumber, formatPercentage } from '@/lib/utils'

// Data sets for different date ranges
const dataByRange = {
  '1d': {
    metrics: {
      revenue: { value: 18230, change: 5.2, trend: 'up' as const },
      impressions: { value: 178234, change: 3.1, trend: 'up' as const },
      clicks: { value: 6543, change: 8.4, trend: 'up' as const },
      ctr: { value: 0.0367, change: 1.2, trend: 'up' as const },
      conversions: { value: 334, change: 2.5, trend: 'up' as const },
      cpa: { value: 54.6, change: -3.2, trend: 'down' as const },
      roas: { value: 3.9, change: 4.1, trend: 'up' as const },
      avgBid: { value: 1.82, change: 0.8, trend: 'up' as const },
    },
    performanceData: [
      { date: 'Today 00:00', impressions: 7426, clicks: 273, revenue: 760 },
      { date: 'Today 04:00', impressions: 5234, clicks: 192, revenue: 534 },
      { date: 'Today 08:00', impressions: 23456, clicks: 862, revenue: 2401 },
      { date: 'Today 12:00', impressions: 45678, clicks: 1678, revenue: 4676 },
      { date: 'Today 16:00', impressions: 52340, clicks: 1923, revenue: 5359 },
      { date: 'Today 20:00', impressions: 44100, clicks: 1615, revenue: 4500 },
    ],
  },
  '7d': {
    metrics: {
      revenue: { value: 125430, change: 12.5, trend: 'up' as const },
      impressions: { value: 1247832, change: 8.2, trend: 'up' as const },
      clicks: { value: 45891, change: 15.3, trend: 'up' as const },
      ctr: { value: 0.0368, change: 3.1, trend: 'up' as const },
      conversions: { value: 2345, change: -2.3, trend: 'down' as const },
      cpa: { value: 53.5, change: -5.8, trend: 'down' as const },
      roas: { value: 4.2, change: 18.5, trend: 'up' as const },
      avgBid: { value: 1.85, change: 2.1, trend: 'up' as const },
    },
    performanceData: [
      { date: 'Jan 22', impressions: 156234, clicks: 5234, revenue: 15234 },
      { date: 'Jan 23', impressions: 178456, clicks: 6123, revenue: 17456 },
      { date: 'Jan 24', impressions: 189234, clicks: 6789, revenue: 19234 },
      { date: 'Jan 25', impressions: 165789, clicks: 5456, revenue: 16789 },
      { date: 'Jan 26', impressions: 198765, clicks: 7234, revenue: 21234 },
      { date: 'Jan 27', impressions: 212345, clicks: 7890, revenue: 23456 },
      { date: 'Jan 28', impressions: 187654, clicks: 6543, revenue: 19876 },
    ],
  },
  '30d': {
    metrics: {
      revenue: { value: 523450, change: 18.7, trend: 'up' as const },
      impressions: { value: 5234567, change: 12.4, trend: 'up' as const },
      clicks: { value: 192345, change: 21.2, trend: 'up' as const },
      ctr: { value: 0.0367, change: 5.8, trend: 'up' as const },
      conversions: { value: 9876, change: 8.4, trend: 'up' as const },
      cpa: { value: 53.0, change: -8.2, trend: 'down' as const },
      roas: { value: 4.5, change: 22.3, trend: 'up' as const },
      avgBid: { value: 1.88, change: 4.5, trend: 'up' as const },
    },
    performanceData: [
      { date: 'Week 1', impressions: 1123456, clicks: 41234, revenue: 112345 },
      { date: 'Week 2', impressions: 1234567, clicks: 45678, revenue: 123456 },
      { date: 'Week 3', impressions: 1345678, clicks: 49876, revenue: 134567 },
      { date: 'Week 4', impressions: 1530866, clicks: 55557, revenue: 153082 },
    ],
  },
  '90d': {
    metrics: {
      revenue: { value: 1523450, change: 24.5, trend: 'up' as const },
      impressions: { value: 15234567, change: 18.9, trend: 'up' as const },
      clicks: { value: 567890, change: 28.4, trend: 'up' as const },
      ctr: { value: 0.0373, change: 8.2, trend: 'up' as const },
      conversions: { value: 29876, change: 15.6, trend: 'up' as const },
      cpa: { value: 51.0, change: -12.5, trend: 'down' as const },
      roas: { value: 4.8, change: 28.7, trend: 'up' as const },
      avgBid: { value: 1.92, change: 6.8, trend: 'up' as const },
    },
    performanceData: [
      { date: 'Nov', impressions: 4567890, clicks: 167890, revenue: 456789 },
      { date: 'Dec', impressions: 5123456, clicks: 189234, revenue: 512345 },
      { date: 'Jan', impressions: 5543221, clicks: 210766, revenue: 554316 },
    ],
  },
}

const dateRangeLabels: Record<string, string> = {
  '1d': 'Today',
  '7d': 'Last 7 days',
  '30d': 'Last 30 days',
  '90d': 'Last 90 days',
}

export default function AnalyticsPage() {
  const [dateRange, setDateRange] = useState('7d')
  const [showDatePicker, setShowDatePicker] = useState(false)
  const [exportSuccess, setExportSuccess] = useState(false)

  // Get data based on selected date range
  const currentData = dataByRange[dateRange as keyof typeof dataByRange]
  const { metrics, performanceData } = currentData

  const topCampaigns = [
    { name: 'Summer Sale 2026', revenue: 32450, conversions: 1234, roas: 5.2 },
    { name: 'New Product Launch', revenue: 28920, conversions: 1089, roas: 4.8 },
    { name: 'Brand Awareness Q1', revenue: 24560, conversions: 956, roas: 3.9 },
    { name: 'Mobile App Promotion', revenue: 21340, conversions: 876, roas: 4.5 },
    { name: 'Holiday Special', revenue: 18180, conversions: 723, roas: 3.2 },
  ]

  // Geographic Performance Data
  const geoPerformanceData = [
    { country: 'United States', code: 'US', impressions: 2450000, clicks: 89000, revenue: 42500, ctr: 3.63, color: '#22c55e' },
    { country: 'United Kingdom', code: 'UK', impressions: 890000, clicks: 32000, revenue: 15800, ctr: 3.60, color: '#3b82f6' },
    { country: 'Germany', code: 'DE', impressions: 720000, clicks: 25000, revenue: 12400, ctr: 3.47, color: '#8b5cf6' },
    { country: 'Canada', code: 'CA', impressions: 560000, clicks: 19500, revenue: 9800, ctr: 3.48, color: '#f59e0b' },
    { country: 'Australia', code: 'AU', impressions: 380000, clicks: 13200, revenue: 6500, ctr: 3.47, color: '#ef4444' },
    { country: 'France', code: 'FR', impressions: 340000, clicks: 11500, revenue: 5200, ctr: 3.38, color: '#06b6d4' },
    { country: 'Japan', code: 'JP', impressions: 290000, clicks: 9800, revenue: 4800, ctr: 3.38, color: '#ec4899' },
    { country: 'India', code: 'IN', impressions: 850000, clicks: 28000, revenue: 8200, ctr: 3.29, color: '#84cc16' },
  ]

  // Format Performance Breakdown
  const formatPerformanceData = [
    { format: 'Rewarded Video', impressions: 1250000, installs: 45000, revenue: 38500, ecpm: 30.80, fillRate: 94.2 },
    { format: 'Playable Ads', impressions: 680000, installs: 28000, revenue: 24200, ecpm: 35.59, fillRate: 87.5 },
    { format: 'Native Ads', impressions: 920000, installs: 18000, revenue: 15400, ecpm: 16.74, fillRate: 91.8 },
    { format: 'Interstitial', impressions: 540000, installs: 12000, revenue: 9800, ecpm: 18.15, fillRate: 89.3 },
    { format: 'Banner', impressions: 2100000, installs: 8500, revenue: 6300, ecpm: 3.00, fillRate: 96.5 },
    { format: 'Offerwall', impressions: 180000, installs: 9200, revenue: 11200, ecpm: 62.22, fillRate: 78.4 },
  ]

  // Time of Day Performance
  const hourlyPerformance = [
    { hour: '00:00', impressions: 45000, ctr: 2.8, conversions: 120 },
    { hour: '03:00', impressions: 32000, ctr: 2.5, conversions: 85 },
    { hour: '06:00', impressions: 58000, ctr: 3.1, conversions: 180 },
    { hour: '09:00', impressions: 125000, ctr: 3.8, conversions: 420 },
    { hour: '12:00', impressions: 180000, ctr: 4.2, conversions: 680 },
    { hour: '15:00', impressions: 165000, ctr: 4.0, conversions: 590 },
    { hour: '18:00', impressions: 195000, ctr: 4.5, conversions: 750 },
    { hour: '21:00', impressions: 142000, ctr: 3.9, conversions: 480 },
  ]

  // Device & OEM Performance
  const devicePerformanceData = [
    { device: 'OPPO', impressions: 450000, installs: 18500, revenue: 15200, ecpm: 33.78 },
    { device: 'VIVO', impressions: 380000, installs: 15800, revenue: 12800, ecpm: 33.68 },
    { device: 'Xiaomi', impressions: 520000, installs: 21000, revenue: 17500, ecpm: 33.65 },
    { device: 'Samsung', impressions: 680000, installs: 24500, revenue: 19800, ecpm: 29.12 },
    { device: 'Huawei', impressions: 290000, installs: 11200, revenue: 9200, ecpm: 31.72 },
    { device: 'Other Android', impressions: 420000, installs: 14000, revenue: 10500, ecpm: 25.00 },
    { device: 'iOS', impressions: 850000, installs: 35000, revenue: 42000, ecpm: 49.41 },
  ]

  // Cross-Format Attribution Data
  const crossFormatAttributionData = [
    { 
      journey: 'Banner → Rewarded Video → Install',
      users: 12500,
      conversions: 3750,
      cvr: 30.0,
      avgTouchpoints: 2.3,
      revenue: 18750,
    },
    {
      journey: 'Native Ad → Playable → Install',
      users: 8200,
      conversions: 2870,
      cvr: 35.0,
      avgTouchpoints: 2.1,
      revenue: 14350,
    },
    {
      journey: 'Interstitial → Install (Direct)',
      users: 25000,
      conversions: 5000,
      cvr: 20.0,
      avgTouchpoints: 1.0,
      revenue: 25000,
    },
    {
      journey: 'Banner → Native → Rewarded Video → Install',
      users: 4500,
      conversions: 1800,
      cvr: 40.0,
      avgTouchpoints: 3.2,
      revenue: 10800,
    },
    {
      journey: 'Playable Ad → Install (Direct)',
      users: 15000,
      conversions: 5250,
      cvr: 35.0,
      avgTouchpoints: 1.0,
      revenue: 26250,
    },
  ]

  // Export to CSV function
  const handleExport = () => {
    // Create CSV content
    const headers = ['Metric', 'Value', 'Change %', 'Trend']
    const metricsRows = [
      ['Revenue', `$${metrics.revenue.value}`, `${metrics.revenue.change}%`, metrics.revenue.trend],
      ['Impressions', metrics.impressions.value.toString(), `${metrics.impressions.change}%`, metrics.impressions.trend],
      ['Clicks', metrics.clicks.value.toString(), `${metrics.clicks.change}%`, metrics.clicks.trend],
      ['CTR', `${(metrics.ctr.value * 100).toFixed(2)}%`, `${metrics.ctr.change}%`, metrics.ctr.trend],
      ['Conversions', metrics.conversions.value.toString(), `${metrics.conversions.change}%`, metrics.conversions.trend],
      ['CPA', `$${metrics.cpa.value}`, `${metrics.cpa.change}%`, metrics.cpa.trend],
      ['ROAS', `${metrics.roas.value}x`, `${metrics.roas.change}%`, metrics.roas.trend],
      ['Avg Bid', `$${metrics.avgBid.value}`, `${metrics.avgBid.change}%`, metrics.avgBid.trend],
    ]

    const performanceHeaders = ['Date', 'Impressions', 'Clicks', 'Revenue']
    const performanceRows = performanceData.map(d => [d.date, d.impressions.toString(), d.clicks.toString(), `$${d.revenue}`])

    const csvContent = [
      `Analytics Report - ${dateRangeLabels[dateRange]}`,
      '',
      'Key Metrics',
      headers.join(','),
      ...metricsRows.map(row => row.join(',')),
      '',
      'Performance Data',
      performanceHeaders.join(','),
      ...performanceRows.map(row => row.join(',')),
    ].join('\n')

    // Create and download file
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', `analytics_${dateRange}_${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    // Show success feedback
    setExportSuccess(true)
    setTimeout(() => setExportSuccess(false), 2000)
  }

  const MetricCard = ({
    title,
    icon,
    value,
    change,
    trend,
    format = 'number',
  }: {
    title: string
    icon: React.ReactNode
    value: number
    change: number
    trend: 'up' | 'down'
    format?: 'currency' | 'number' | 'percentage'
  }) => {
    const formatValue = (val: number) => {
      switch (format) {
        case 'currency':
          return formatCurrency(val)
        case 'percentage':
          return formatPercentage(val)
        default:
          return formatNumber(val)
      }
    }

    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">{title}</CardTitle>
          <div className="rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 p-2 text-white">
            {icon}
          </div>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{formatValue(value)}</div>
          <div className="flex items-center text-xs">
            {trend === 'up' ? (
              <TrendingUp className="mr-1 h-4 w-4 text-green-500" />
            ) : (
              <TrendingDown className="mr-1 h-4 w-4 text-red-500" />
            )}
            <span className={trend === 'up' ? 'text-green-500' : 'text-red-500'}>
              {Math.abs(change)}%
            </span>
            <span className="ml-1 text-muted-foreground">vs last period</span>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Analytics</h1>
          <p className="text-muted-foreground">
            Performance insights and reporting
          </p>
        </div>
        <div className="flex items-center space-x-2">
          {/* Date Range Picker */}
          <div className="relative">
            <Button 
              variant="outline" 
              size="sm"
              onClick={() => setShowDatePicker(!showDatePicker)}
            >
              <Calendar className="mr-2 h-4 w-4" />
              {dateRangeLabels[dateRange]}
              <ChevronDown className="ml-2 h-4 w-4" />
            </Button>
            {showDatePicker && (
              <div className="absolute right-0 top-full mt-1 z-50 bg-white rounded-lg shadow-lg border p-1 min-w-[150px]">
                {Object.entries(dateRangeLabels).map(([value, label]) => (
                  <button
                    key={value}
                    onClick={() => {
                      setDateRange(value)
                      setShowDatePicker(false)
                    }}
                    className={`w-full text-left px-3 py-2 text-sm rounded hover:bg-gray-100 flex items-center justify-between ${
                      dateRange === value ? 'bg-blue-50 text-blue-600' : ''
                    }`}
                  >
                    {label}
                    {dateRange === value && <Check className="h-4 w-4" />}
                  </button>
                ))}
              </div>
            )}
          </div>
          {/* Export Button */}
          <Button 
            variant="outline" 
            size="sm"
            onClick={handleExport}
            className={exportSuccess ? 'bg-green-50 text-green-600 border-green-200' : ''}
          >
            {exportSuccess ? (
              <>
                <Check className="mr-2 h-4 w-4" />
                Exported!
              </>
            ) : (
              <>
                <Download className="mr-2 h-4 w-4" />
                Export CSV
              </>
            )}
          </Button>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          title="Total Revenue"
          icon={<DollarSign className="h-4 w-4" />}
          value={metrics.revenue.value}
          change={metrics.revenue.change}
          trend={metrics.revenue.trend}
          format="currency"
        />
        <MetricCard
          title="Impressions"
          icon={<Users className="h-4 w-4" />}
          value={metrics.impressions.value}
          change={metrics.impressions.change}
          trend={metrics.impressions.trend}
        />
        <MetricCard
          title="Clicks"
          icon={<MousePointerClick className="h-4 w-4" />}
          value={metrics.clicks.value}
          change={metrics.clicks.change}
          trend={metrics.clicks.trend}
        />
        <MetricCard
          title="CTR"
          icon={<Target className="h-4 w-4" />}
          value={metrics.ctr.value}
          change={metrics.ctr.change}
          trend={metrics.ctr.trend}
          format="percentage"
        />
      </div>

      {/* Secondary Metrics */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Conversions
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatNumber(metrics.conversions.value)}</div>
            <div className="flex items-center text-xs">
              <TrendingDown className="mr-1 h-4 w-4 text-red-500" />
              <span className="text-red-500">{Math.abs(metrics.conversions.change)}%</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Cost Per Acquisition
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(metrics.cpa.value)}</div>
            <div className="flex items-center text-xs">
              <TrendingDown className="mr-1 h-4 w-4 text-green-500" />
              <span className="text-green-500">{Math.abs(metrics.cpa.change)}%</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Return on Ad Spend
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics.roas.value}x</div>
            <div className="flex items-center text-xs">
              <TrendingUp className="mr-1 h-4 w-4 text-green-500" />
              <span className="text-green-500">{metrics.roas.change}%</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Average Bid
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(metrics.avgBid.value)}</div>
            <div className="flex items-center text-xs">
              <TrendingUp className="mr-1 h-4 w-4 text-blue-500" />
              <span className="text-blue-500">{metrics.avgBid.change}%</span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Performance Chart */}
      <Card>
        <CardHeader>
          <CardTitle>Performance Overview</CardTitle>
          <CardDescription>
            {dateRange === '1d' ? 'Hourly metrics for today' :
             dateRange === '7d' ? 'Daily metrics for the last 7 days' :
             dateRange === '30d' ? 'Weekly metrics for the last 30 days' :
             'Monthly metrics for the last 90 days'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-80 w-full">
            {/* Simple bar chart visualization */}
            <div className="flex h-full items-end space-x-2">
              {performanceData.map((day, index) => {
                const maxRevenue = Math.max(...performanceData.map((d) => d.revenue))
                const height = (day.revenue / maxRevenue) * 100
                return (
                  <div key={index} className="flex flex-1 flex-col items-center">
                    <div className="relative w-full">
                      <div
                        className="w-full rounded-t-lg bg-gradient-to-t from-blue-600 to-purple-600 transition-all hover:opacity-80"
                        style={{ height: `${height * 2.5}px` }}
                        title={`${day.date}: ${formatCurrency(day.revenue)}`}
                      />
                    </div>
                    <div className="mt-2 text-xs text-muted-foreground">{day.date.split(' ')[1]}</div>
                  </div>
                )
              })}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Top Campaigns */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Top Performing Campaigns</CardTitle>
            <CardDescription>By revenue ({dateRangeLabels[dateRange].toLowerCase()})</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {topCampaigns.map((campaign, index) => (
                <div key={index} className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 text-sm font-bold text-white">
                      {index + 1}
                    </div>
                    <div>
                      <div className="font-medium">{campaign.name}</div>
                      <div className="text-xs text-muted-foreground">
                        {formatNumber(campaign.conversions)} conversions
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-semibold">{formatCurrency(campaign.revenue)}</div>
                    <div className="text-xs text-green-600">{campaign.roas}x ROAS</div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Conversion Funnel</CardTitle>
            <CardDescription>User journey breakdown</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {[
                { stage: 'Impressions', count: 1247832, percentage: 100 },
                { stage: 'Clicks', count: 45891, percentage: 3.68 },
                { stage: 'Landing Page Views', count: 42456, percentage: 3.40 },
                { stage: 'Add to Cart', count: 5678, percentage: 0.45 },
                { stage: 'Conversions', count: 2345, percentage: 0.19 },
              ].map((stage, index) => (
                <div key={index} className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="font-medium">{stage.stage}</span>
                    <span className="text-muted-foreground">
                      {formatNumber(stage.count)} ({formatPercentage(stage.percentage / 100)})
                    </span>
                  </div>
                  <div className="h-2 w-full overflow-hidden rounded-full bg-gray-200">
                    <div
                      className="h-full bg-gradient-to-r from-blue-600 to-purple-600 transition-all"
                      style={{ width: `${stage.percentage}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Geographic Performance */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Globe className="h-5 w-5 text-blue-500" />
                Geographic Performance
              </CardTitle>
              <CardDescription>Performance breakdown by country/region</CardDescription>
            </div>
            <select className="rounded-lg border bg-background px-3 py-2 text-sm">
              <option>All Regions</option>
              <option>North America</option>
              <option>Europe</option>
              <option>Asia Pacific</option>
            </select>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            {/* Country List */}
            <div className="space-y-3">
              {geoPerformanceData.map((geo, index) => (
                <div key={index} className="flex items-center justify-between rounded-lg border p-3 hover:bg-muted/50 transition-colors">
                  <div className="flex items-center gap-3">
                    <div 
                      className="h-3 w-3 rounded-full" 
                      style={{ backgroundColor: geo.color }}
                    />
                    <div>
                      <div className="font-medium">{geo.country}</div>
                      <div className="text-xs text-muted-foreground">
                        {formatNumber(geo.impressions)} impressions
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-semibold">{formatCurrency(geo.revenue)}</div>
                    <div className="text-xs text-muted-foreground">{geo.ctr}% CTR</div>
                  </div>
                </div>
              ))}
            </div>
            {/* Visual Summary */}
            <div className="space-y-4">
              <div className="rounded-lg bg-gradient-to-br from-blue-50 to-purple-50 dark:from-blue-950/20 dark:to-purple-950/20 p-4">
                <h4 className="font-semibold mb-3">Regional Summary</h4>
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Total Countries</span>
                    <span className="font-medium">{geoPerformanceData.length}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Total Impressions</span>
                    <span className="font-medium">{formatNumber(geoPerformanceData.reduce((sum, g) => sum + g.impressions, 0))}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Total Revenue</span>
                    <span className="font-medium">{formatCurrency(geoPerformanceData.reduce((sum, g) => sum + g.revenue, 0))}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Avg CTR</span>
                    <span className="font-medium">{(geoPerformanceData.reduce((sum, g) => sum + g.ctr, 0) / geoPerformanceData.length).toFixed(2)}%</span>
                  </div>
                </div>
              </div>
              <div className="rounded-lg border p-4">
                <h4 className="font-semibold mb-3">Top Performer</h4>
                <div className="flex items-center gap-3">
                  <MapPin className="h-8 w-8 text-green-500" />
                  <div>
                    <div className="font-medium">United States</div>
                    <div className="text-sm text-muted-foreground">
                      {formatCurrency(42500)} revenue • 3.63% CTR
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Format Performance Breakdown */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <BarChart2 className="h-5 w-5 text-purple-500" />
                Ad Format Performance
              </CardTitle>
              <CardDescription>Compare performance across different ad formats</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b">
                  <th className="pb-3 text-left font-medium">Format</th>
                  <th className="pb-3 text-right font-medium">Impressions</th>
                  <th className="pb-3 text-right font-medium">Installs</th>
                  <th className="pb-3 text-right font-medium">Revenue</th>
                  <th className="pb-3 text-right font-medium">eCPM</th>
                  <th className="pb-3 text-right font-medium">Fill Rate</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {formatPerformanceData.map((format, index) => (
                  <tr key={index} className="hover:bg-muted/50">
                    <td className="py-3">
                      <div className="flex items-center gap-2">
                        <div className={`h-2 w-2 rounded-full ${
                          format.format === 'Rewarded Video' ? 'bg-green-500' :
                          format.format === 'Playable Ads' ? 'bg-purple-500' :
                          format.format === 'Native Ads' ? 'bg-blue-500' :
                          format.format === 'Interstitial' ? 'bg-orange-500' :
                          format.format === 'Banner' ? 'bg-gray-500' : 'bg-pink-500'
                        }`} />
                        <span className="font-medium">{format.format}</span>
                      </div>
                    </td>
                    <td className="py-3 text-right">{formatNumber(format.impressions)}</td>
                    <td className="py-3 text-right">{formatNumber(format.installs)}</td>
                    <td className="py-3 text-right font-semibold">{formatCurrency(format.revenue)}</td>
                    <td className="py-3 text-right">
                      <span className={`rounded-full px-2 py-1 text-xs ${
                        format.ecpm > 30 ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' :
                        format.ecpm > 15 ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400' :
                        'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400'
                      }`}>
                        ${format.ecpm.toFixed(2)}
                      </span>
                    </td>
                    <td className="py-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <div className="h-2 w-16 rounded-full bg-gray-200 dark:bg-gray-700">
                          <div 
                            className="h-full rounded-full bg-blue-500"
                            style={{ width: `${format.fillRate}%` }}
                          />
                        </div>
                        <span className="text-sm">{format.fillRate}%</span>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {/* Time of Day & Device Performance */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* Time of Day Performance */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Clock className="h-5 w-5 text-orange-500" />
              Time of Day Performance
            </CardTitle>
            <CardDescription>Optimize delivery timing</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {hourlyPerformance.map((hour, index) => (
                <div key={index} className="flex items-center gap-3">
                  <span className="w-12 text-sm text-muted-foreground">{hour.hour}</span>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <div className="h-6 rounded bg-gradient-to-r from-blue-500 to-purple-500" 
                           style={{ width: `${(hour.impressions / 195000) * 100}%` }} />
                      <span className="text-xs text-muted-foreground">{formatNumber(hour.impressions)}</span>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm font-medium">{hour.ctr}% CTR</div>
                    <div className="text-xs text-muted-foreground">{hour.conversions} conv</div>
                  </div>
                </div>
              ))}
            </div>
            <div className="mt-4 rounded-lg bg-orange-50 dark:bg-orange-900/20 p-3">
              <div className="flex items-center gap-2 text-sm">
                <TrendingUp className="h-4 w-4 text-orange-500" />
                <span className="font-medium">Peak Hours:</span>
                <span className="text-muted-foreground">18:00 - 21:00 (4.5% CTR)</span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Device & OEM Performance */}
        <Card>
          <CardHeader>
            <CardTitle>Device & OEM Performance</CardTitle>
            <CardDescription>Performance by device manufacturer</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {devicePerformanceData.map((device, index) => (
                <div key={index} className="flex items-center justify-between rounded-lg border p-3 hover:bg-muted/50 transition-colors">
                  <div>
                    <div className="font-medium">{device.device}</div>
                    <div className="text-xs text-muted-foreground">
                      {formatNumber(device.impressions)} impressions • {formatNumber(device.installs)} installs
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-semibold">{formatCurrency(device.revenue)}</div>
                    <div className={`text-xs ${device.ecpm > 35 ? 'text-green-600' : device.ecpm > 25 ? 'text-yellow-600' : 'text-muted-foreground'}`}>
                      ${device.ecpm.toFixed(2)} eCPM
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Cross-Format Attribution */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Target className="h-5 w-5 text-indigo-500" />
                Cross-Format Attribution
              </CardTitle>
              <CardDescription>User journey analysis across ad formats</CardDescription>
            </div>
            <select className="rounded-lg border bg-background px-3 py-2 text-sm">
              <option>Last 7 Days</option>
              <option>Last 30 Days</option>
              <option>Last 90 Days</option>
            </select>
          </div>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b">
                  <th className="pb-3 text-left font-medium">User Journey</th>
                  <th className="pb-3 text-right font-medium">Users</th>
                  <th className="pb-3 text-right font-medium">Conversions</th>
                  <th className="pb-3 text-right font-medium">CVR</th>
                  <th className="pb-3 text-right font-medium">Avg Touchpoints</th>
                  <th className="pb-3 text-right font-medium">Revenue</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {crossFormatAttributionData.map((journey, index) => (
                  <tr key={index} className="hover:bg-muted/50">
                    <td className="py-3">
                      <div className="flex items-center gap-2">
                        <div className="flex items-center gap-1">
                          {journey.journey.split(' → ').map((step, stepIndex, arr) => (
                            <span key={stepIndex} className="flex items-center gap-1">
                              <span className={`rounded-full px-2 py-0.5 text-xs ${
                                step === 'Banner' ? 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300' :
                                step === 'Rewarded Video' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' :
                                step === 'Native Ad' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' :
                                step === 'Playable' || step === 'Playable Ad' ? 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400' :
                                step === 'Interstitial' ? 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400' :
                                step.includes('Install') ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' :
                                'bg-gray-100 text-gray-700'
                              }`}>
                                {step}
                              </span>
                              {stepIndex < arr.length - 1 && (
                                <span className="text-muted-foreground">→</span>
                              )}
                            </span>
                          ))}
                        </div>
                      </div>
                    </td>
                    <td className="py-3 text-right">{formatNumber(journey.users)}</td>
                    <td className="py-3 text-right font-medium">{formatNumber(journey.conversions)}</td>
                    <td className="py-3 text-right">
                      <span className={`rounded-full px-2 py-1 text-xs font-medium ${
                        journey.cvr >= 35 ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' :
                        journey.cvr >= 25 ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400' :
                        'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400'
                      }`}>
                        {journey.cvr.toFixed(1)}%
                      </span>
                    </td>
                    <td className="py-3 text-right">{journey.avgTouchpoints.toFixed(1)}</td>
                    <td className="py-3 text-right font-semibold">{formatCurrency(journey.revenue)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          
          {/* Attribution Insights */}
          <div className="mt-6 grid gap-4 md:grid-cols-3">
            <div className="rounded-lg bg-gradient-to-br from-green-50 to-emerald-50 dark:from-green-950/20 dark:to-emerald-950/20 p-4">
              <h4 className="font-semibold text-green-700 dark:text-green-400 mb-2">Best Performing Journey</h4>
              <div className="text-sm text-green-600 dark:text-green-500">
                Banner → Native → Rewarded Video
              </div>
              <div className="text-2xl font-bold text-green-700 dark:text-green-400 mt-1">40% CVR</div>
            </div>
            <div className="rounded-lg bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-950/20 dark:to-indigo-950/20 p-4">
              <h4 className="font-semibold text-blue-700 dark:text-blue-400 mb-2">Most Efficient</h4>
              <div className="text-sm text-blue-600 dark:text-blue-500">
                Playable Ad → Install (Direct)
              </div>
              <div className="text-2xl font-bold text-blue-700 dark:text-blue-400 mt-1">1.0 Touchpoints</div>
            </div>
            <div className="rounded-lg bg-gradient-to-br from-purple-50 to-violet-50 dark:from-purple-950/20 dark:to-violet-950/20 p-4">
              <h4 className="font-semibold text-purple-700 dark:text-purple-400 mb-2">Highest Revenue</h4>
              <div className="text-sm text-purple-600 dark:text-purple-500">
                Playable Ad → Install (Direct)
              </div>
              <div className="text-2xl font-bold text-purple-700 dark:text-purple-400 mt-1">{formatCurrency(26250)}</div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
