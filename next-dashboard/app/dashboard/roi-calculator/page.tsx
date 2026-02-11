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
  Calculator,
  DollarSign,
  TrendingUp,
  TrendingDown,
  Target,
  PieChart,
  BarChart3,
  RefreshCw,
  Download,
  Info,
  Lightbulb,
  AlertTriangle,
  CheckCircle2,
} from 'lucide-react'
import { formatCurrency, formatNumber, formatPercentage } from '@/lib/utils'

// Industry benchmarks
const industryBenchmarks = {
  gaming: { ctr: 2.8, cvr: 3.5, cpa: 2.50, roas: 3.2, ltv: 8.50 },
  ecommerce: { ctr: 1.9, cvr: 2.8, cpa: 15.00, roas: 4.5, ltv: 65.00 },
  fintech: { ctr: 1.2, cvr: 1.5, cpa: 45.00, roas: 5.2, ltv: 250.00 },
  travel: { ctr: 2.1, cvr: 2.2, cpa: 35.00, roas: 3.8, ltv: 120.00 },
  utilities: { ctr: 3.5, cvr: 8.0, cpa: 1.80, roas: 2.5, ltv: 4.50 },
  social: { ctr: 2.4, cvr: 4.2, cpa: 3.20, roas: 2.8, ltv: 9.00 },
}

export default function ROICalculatorPage() {
  // Input states
  const [adSpend, setAdSpend] = useState(10000)
  const [impressions, setImpressions] = useState(1000000)
  const [clicks, setClicks] = useState(25000)
  const [conversions, setConversions] = useState(750)
  const [revenue, setRevenue] = useState(45000)
  const [avgOrderValue, setAvgOrderValue] = useState(60)
  const [customerLTV, setCustomerLTV] = useState(180)
  const [industry, setIndustry] = useState<keyof typeof industryBenchmarks>('gaming')
  
  // Scenario comparison
  const [compareMode, setCompareMode] = useState(false)
  const [scenarioB, setScenarioB] = useState({
    adSpend: 15000,
    impressions: 1500000,
    clicks: 40000,
    conversions: 1100,
    revenue: 72000,
  })

  // Calculated metrics
  const metrics = useMemo(() => {
    const ctr = (clicks / impressions) * 100
    const cvr = (conversions / clicks) * 100
    const cpc = adSpend / clicks
    const cpm = (adSpend / impressions) * 1000
    const cpa = adSpend / conversions
    const roas = revenue / adSpend
    const roi = ((revenue - adSpend) / adSpend) * 100
    const profit = revenue - adSpend
    const revenuePerConversion = revenue / conversions
    const ltvRoas = (customerLTV * conversions) / adSpend
    const breakEvenConversions = Math.ceil(adSpend / avgOrderValue)
    const breakEvenRoas = 1

    return {
      ctr,
      cvr,
      cpc,
      cpm,
      cpa,
      roas,
      roi,
      profit,
      revenuePerConversion,
      ltvRoas,
      breakEvenConversions,
      breakEvenRoas,
    }
  }, [adSpend, impressions, clicks, conversions, revenue, avgOrderValue, customerLTV])

  // Scenario B metrics
  const scenarioBMetrics = useMemo(() => {
    const ctr = (scenarioB.clicks / scenarioB.impressions) * 100
    const cvr = (scenarioB.conversions / scenarioB.clicks) * 100
    const cpc = scenarioB.adSpend / scenarioB.clicks
    const cpm = (scenarioB.adSpend / scenarioB.impressions) * 1000
    const cpa = scenarioB.adSpend / scenarioB.conversions
    const roas = scenarioB.revenue / scenarioB.adSpend
    const roi = ((scenarioB.revenue - scenarioB.adSpend) / scenarioB.adSpend) * 100
    const profit = scenarioB.revenue - scenarioB.adSpend

    return { ctr, cvr, cpc, cpm, cpa, roas, roi, profit }
  }, [scenarioB])

  // Benchmark comparison
  const benchmark = industryBenchmarks[industry]
  
  const getBenchmarkStatus = (metric: string, value: number) => {
    const benchmarkValue = benchmark[metric as keyof typeof benchmark]
    if (!benchmarkValue) return 'neutral'
    
    if (metric === 'cpa') {
      return value < benchmarkValue ? 'good' : value > benchmarkValue * 1.2 ? 'bad' : 'neutral'
    }
    return value > benchmarkValue ? 'good' : value < benchmarkValue * 0.8 ? 'bad' : 'neutral'
  }

  // Recommendations
  const recommendations = useMemo(() => {
    const recs = []
    
    if (metrics.ctr < benchmark.ctr * 0.8) {
      recs.push({
        type: 'warning',
        title: 'Low CTR',
        description: `Your CTR (${metrics.ctr.toFixed(2)}%) is below industry average (${benchmark.ctr}%). Consider testing new creatives or improving ad relevance.`,
      })
    }
    
    if (metrics.cvr < benchmark.cvr * 0.8) {
      recs.push({
        type: 'warning',
        title: 'Low Conversion Rate',
        description: `Your CVR (${metrics.cvr.toFixed(2)}%) is below industry average (${benchmark.cvr}%). Review your landing page experience and targeting.`,
      })
    }
    
    if (metrics.cpa > benchmark.cpa * 1.2) {
      recs.push({
        type: 'warning',
        title: 'High CPA',
        description: `Your CPA (${formatCurrency(metrics.cpa)}) is above industry average (${formatCurrency(benchmark.cpa)}). Optimize targeting or bid strategies.`,
      })
    }
    
    if (metrics.roas > benchmark.roas * 1.2) {
      recs.push({
        type: 'success',
        title: 'Strong ROAS',
        description: `Your ROAS (${metrics.roas.toFixed(2)}x) exceeds industry average (${benchmark.roas}x). Consider scaling your budget.`,
      })
    }
    
    if (metrics.ltvRoas > metrics.roas * 2) {
      recs.push({
        type: 'info',
        title: 'High LTV Opportunity',
        description: `Your LTV-based ROAS (${metrics.ltvRoas.toFixed(2)}x) is significantly higher than immediate ROAS. Consider longer attribution windows.`,
      })
    }
    
    if (conversions < metrics.breakEvenConversions) {
      recs.push({
        type: 'warning',
        title: 'Below Break-Even',
        description: `You need ${metrics.breakEvenConversions} conversions to break even, but only have ${conversions}. Focus on improving conversion rate.`,
      })
    }

    return recs
  }, [metrics, benchmark, conversions])

  const resetCalculator = () => {
    setAdSpend(10000)
    setImpressions(1000000)
    setClicks(25000)
    setConversions(750)
    setRevenue(45000)
    setAvgOrderValue(60)
    setCustomerLTV(180)
  }

  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">ROI Calculator</h1>
          <p className="text-muted-foreground">
            Calculate and optimize your advertising return on investment
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={resetCalculator}
            className="inline-flex items-center gap-2 rounded-lg border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors"
          >
            <RefreshCw className="h-4 w-4" />
            Reset
          </button>
          <button
            className="inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 px-4 py-2 text-sm font-medium text-white shadow-lg hover:opacity-90 transition-opacity"
          >
            <Download className="h-4 w-4" />
            Export Report
          </button>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Input Panel */}
        <div className="lg:col-span-1 space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Calculator className="h-5 w-5 text-blue-500" />
                Campaign Inputs
              </CardTitle>
              <CardDescription>Enter your campaign data</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">Industry</label>
                <select
                  value={industry}
                  onChange={(e) => setIndustry(e.target.value as keyof typeof industryBenchmarks)}
                  className="w-full rounded-lg border bg-background px-3 py-2"
                >
                  <option value="gaming">Gaming / Apps</option>
                  <option value="ecommerce">E-Commerce</option>
                  <option value="fintech">Fintech</option>
                  <option value="travel">Travel</option>
                  <option value="utilities">Utilities</option>
                  <option value="social">Social / Entertainment</option>
                </select>
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">Ad Spend ($)</label>
                <input
                  type="number"
                  value={adSpend}
                  onChange={(e) => setAdSpend(Number(e.target.value))}
                  className="w-full rounded-lg border bg-background px-3 py-2"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Impressions</label>
                <input
                  type="number"
                  value={impressions}
                  onChange={(e) => setImpressions(Number(e.target.value))}
                  className="w-full rounded-lg border bg-background px-3 py-2"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Clicks</label>
                <input
                  type="number"
                  value={clicks}
                  onChange={(e) => setClicks(Number(e.target.value))}
                  className="w-full rounded-lg border bg-background px-3 py-2"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Conversions / Installs</label>
                <input
                  type="number"
                  value={conversions}
                  onChange={(e) => setConversions(Number(e.target.value))}
                  className="w-full rounded-lg border bg-background px-3 py-2"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Revenue ($)</label>
                <input
                  type="number"
                  value={revenue}
                  onChange={(e) => setRevenue(Number(e.target.value))}
                  className="w-full rounded-lg border bg-background px-3 py-2"
                />
              </div>

              <div className="pt-4 border-t">
                <h4 className="font-medium mb-3">Advanced Inputs</h4>
                
                <div className="space-y-3">
                  <div>
                    <label className="block text-sm font-medium mb-1">Avg Order Value ($)</label>
                    <input
                      type="number"
                      value={avgOrderValue}
                      onChange={(e) => setAvgOrderValue(Number(e.target.value))}
                      className="w-full rounded-lg border bg-background px-3 py-2"
                    />
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium mb-1">Customer LTV ($)</label>
                    <input
                      type="number"
                      value={customerLTV}
                      onChange={(e) => setCustomerLTV(Number(e.target.value))}
                      className="w-full rounded-lg border bg-background px-3 py-2"
                    />
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Scenario Comparison Toggle */}
          <Card>
            <CardContent className="pt-6">
              <label className="flex items-center justify-between">
                <span className="font-medium">Compare Scenarios</span>
                <button
                  onClick={() => setCompareMode(!compareMode)}
                  className={`relative h-6 w-11 rounded-full transition-colors ${
                    compareMode ? 'bg-blue-600' : 'bg-gray-300'
                  }`}
                >
                  <span
                    className={`absolute top-1 h-4 w-4 rounded-full bg-white transition-transform ${
                      compareMode ? 'translate-x-6' : 'translate-x-1'
                    }`}
                  />
                </button>
              </label>
            </CardContent>
          </Card>
        </div>

        {/* Results Panel */}
        <div className="lg:col-span-2 space-y-4">
          {/* Key Metrics */}
          <div className="grid gap-4 md:grid-cols-4">
            <Card className={metrics.roi > 0 ? 'border-green-500/50' : 'border-red-500/50'}>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-muted-foreground">ROI</p>
                    <p className={`text-2xl font-bold ${metrics.roi > 0 ? 'text-green-600' : 'text-red-600'}`}>
                      {metrics.roi.toFixed(1)}%
                    </p>
                  </div>
                  {metrics.roi > 0 ? (
                    <TrendingUp className="h-8 w-8 text-green-500" />
                  ) : (
                    <TrendingDown className="h-8 w-8 text-red-500" />
                  )}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-muted-foreground">ROAS</p>
                    <p className="text-2xl font-bold">{metrics.roas.toFixed(2)}x</p>
                    <p className="text-xs text-muted-foreground">
                      Benchmark: {benchmark.roas}x
                    </p>
                  </div>
                  <div className={`rounded-full p-2 ${
                    getBenchmarkStatus('roas', metrics.roas) === 'good' 
                      ? 'bg-green-100 dark:bg-green-900/30' 
                      : getBenchmarkStatus('roas', metrics.roas) === 'bad'
                        ? 'bg-red-100 dark:bg-red-900/30'
                        : 'bg-gray-100 dark:bg-gray-800'
                  }`}>
                    <Target className={`h-6 w-6 ${
                      getBenchmarkStatus('roas', metrics.roas) === 'good' 
                        ? 'text-green-600' 
                        : getBenchmarkStatus('roas', metrics.roas) === 'bad'
                          ? 'text-red-600'
                          : 'text-gray-600'
                    }`} />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-muted-foreground">CPA</p>
                    <p className="text-2xl font-bold">{formatCurrency(metrics.cpa)}</p>
                    <p className="text-xs text-muted-foreground">
                      Benchmark: {formatCurrency(benchmark.cpa)}
                    </p>
                  </div>
                  <DollarSign className="h-8 w-8 text-blue-500" />
                </div>
              </CardContent>
            </Card>

            <Card className={metrics.profit > 0 ? 'border-green-500/50' : 'border-red-500/50'}>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-muted-foreground">Net Profit</p>
                    <p className={`text-2xl font-bold ${metrics.profit > 0 ? 'text-green-600' : 'text-red-600'}`}>
                      {formatCurrency(metrics.profit)}
                    </p>
                  </div>
                  <div className={`rounded-full p-2 ${metrics.profit > 0 ? 'bg-green-100' : 'bg-red-100'}`}>
                    {metrics.profit > 0 ? (
                      <TrendingUp className="h-6 w-6 text-green-600" />
                    ) : (
                      <TrendingDown className="h-6 w-6 text-red-600" />
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Detailed Metrics */}
          <Card>
            <CardHeader>
              <CardTitle>Performance Breakdown</CardTitle>
              <CardDescription>Detailed metrics vs industry benchmarks</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {[
                  { label: 'CTR', value: `${metrics.ctr.toFixed(2)}%`, benchmark: `${benchmark.ctr}%`, status: getBenchmarkStatus('ctr', metrics.ctr) },
                  { label: 'CVR', value: `${metrics.cvr.toFixed(2)}%`, benchmark: `${benchmark.cvr}%`, status: getBenchmarkStatus('cvr', metrics.cvr) },
                  { label: 'CPC', value: formatCurrency(metrics.cpc), benchmark: '-', status: 'neutral' },
                  { label: 'CPM', value: formatCurrency(metrics.cpm), benchmark: '-', status: 'neutral' },
                  { label: 'LTV ROAS', value: `${metrics.ltvRoas.toFixed(2)}x`, benchmark: '-', status: metrics.ltvRoas > 3 ? 'good' : 'neutral' },
                  { label: 'Break-Even', value: `${metrics.breakEvenConversions} conv`, benchmark: '-', status: conversions >= metrics.breakEvenConversions ? 'good' : 'bad' },
                ].map((metric, index) => (
                  <div 
                    key={index}
                    className={`rounded-lg border p-4 ${
                      metric.status === 'good' ? 'border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-900/10' :
                      metric.status === 'bad' ? 'border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-900/10' :
                      ''
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground">{metric.label}</span>
                      {metric.status === 'good' && <CheckCircle2 className="h-4 w-4 text-green-600" />}
                      {metric.status === 'bad' && <AlertTriangle className="h-4 w-4 text-red-600" />}
                    </div>
                    <div className="mt-1 text-xl font-bold">{metric.value}</div>
                    {metric.benchmark !== '-' && (
                      <div className="text-xs text-muted-foreground">Industry: {metric.benchmark}</div>
                    )}
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* Scenario Comparison */}
          {compareMode && (
            <Card>
              <CardHeader>
                <CardTitle>Scenario Comparison</CardTitle>
                <CardDescription>Compare current vs alternative scenario</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid gap-6 md:grid-cols-2">
                  {/* Scenario A (Current) */}
                  <div className="rounded-lg border p-4 bg-blue-50 dark:bg-blue-900/10">
                    <h4 className="font-semibold mb-3 text-blue-700 dark:text-blue-400">Scenario A (Current)</h4>
                    <div className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <span>Ad Spend</span>
                        <span className="font-medium">{formatCurrency(adSpend)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Revenue</span>
                        <span className="font-medium">{formatCurrency(revenue)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span>ROAS</span>
                        <span className="font-medium">{metrics.roas.toFixed(2)}x</span>
                      </div>
                      <div className="flex justify-between">
                        <span>ROI</span>
                        <span className={`font-medium ${metrics.roi > 0 ? 'text-green-600' : 'text-red-600'}`}>
                          {metrics.roi.toFixed(1)}%
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span>Profit</span>
                        <span className={`font-medium ${metrics.profit > 0 ? 'text-green-600' : 'text-red-600'}`}>
                          {formatCurrency(metrics.profit)}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Scenario B */}
                  <div className="rounded-lg border p-4">
                    <h4 className="font-semibold mb-3">Scenario B (Alternative)</h4>
                    <div className="space-y-2">
                      <div>
                        <label className="text-xs text-muted-foreground">Ad Spend</label>
                        <input
                          type="number"
                          value={scenarioB.adSpend}
                          onChange={(e) => setScenarioB(s => ({ ...s, adSpend: Number(e.target.value) }))}
                          className="w-full rounded border px-2 py-1 text-sm"
                        />
                      </div>
                      <div>
                        <label className="text-xs text-muted-foreground">Revenue</label>
                        <input
                          type="number"
                          value={scenarioB.revenue}
                          onChange={(e) => setScenarioB(s => ({ ...s, revenue: Number(e.target.value) }))}
                          className="w-full rounded border px-2 py-1 text-sm"
                        />
                      </div>
                      <div className="pt-2 border-t mt-2 space-y-1 text-sm">
                        <div className="flex justify-between">
                          <span>ROAS</span>
                          <span className="font-medium">{scenarioBMetrics.roas.toFixed(2)}x</span>
                        </div>
                        <div className="flex justify-between">
                          <span>ROI</span>
                          <span className={`font-medium ${scenarioBMetrics.roi > 0 ? 'text-green-600' : 'text-red-600'}`}>
                            {scenarioBMetrics.roi.toFixed(1)}%
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span>Profit</span>
                          <span className={`font-medium ${scenarioBMetrics.profit > 0 ? 'text-green-600' : 'text-red-600'}`}>
                            {formatCurrency(scenarioBMetrics.profit)}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                
                {/* Comparison Summary */}
                <div className="mt-4 p-4 rounded-lg bg-muted/50">
                  <h4 className="font-semibold mb-2">Summary</h4>
                  <div className="text-sm">
                    Scenario B {scenarioBMetrics.profit > metrics.profit ? 'generates' : 'loses'}{' '}
                    <span className={scenarioBMetrics.profit > metrics.profit ? 'text-green-600 font-semibold' : 'text-red-600 font-semibold'}>
                      {formatCurrency(Math.abs(scenarioBMetrics.profit - metrics.profit))}
                    </span>{' '}
                    {scenarioBMetrics.profit > metrics.profit ? 'more' : 'less'} profit than Scenario A.
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Recommendations */}
          {recommendations.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Lightbulb className="h-5 w-5 text-yellow-500" />
                  Recommendations
                </CardTitle>
                <CardDescription>AI-powered optimization suggestions</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {recommendations.map((rec, index) => (
                    <div 
                      key={index}
                      className={`flex items-start gap-3 rounded-lg p-3 ${
                        rec.type === 'warning' ? 'bg-yellow-50 dark:bg-yellow-900/10' :
                        rec.type === 'success' ? 'bg-green-50 dark:bg-green-900/10' :
                        'bg-blue-50 dark:bg-blue-900/10'
                      }`}
                    >
                      {rec.type === 'warning' ? (
                        <AlertTriangle className="h-5 w-5 text-yellow-600 mt-0.5" />
                      ) : rec.type === 'success' ? (
                        <CheckCircle2 className="h-5 w-5 text-green-600 mt-0.5" />
                      ) : (
                        <Info className="h-5 w-5 text-blue-600 mt-0.5" />
                      )}
                      <div>
                        <h4 className="font-medium">{rec.title}</h4>
                        <p className="text-sm text-muted-foreground">{rec.description}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Funnel Visualization */}
          <Card>
            <CardHeader>
              <CardTitle>Conversion Funnel</CardTitle>
              <CardDescription>Visualize your marketing funnel efficiency</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {[
                  { stage: 'Impressions', value: impressions, percentage: 100, color: 'from-blue-500 to-blue-600' },
                  { stage: 'Clicks', value: clicks, percentage: (clicks / impressions) * 100, color: 'from-purple-500 to-purple-600' },
                  { stage: 'Conversions', value: conversions, percentage: (conversions / impressions) * 100, color: 'from-green-500 to-green-600' },
                ].map((stage, index) => (
                  <div key={index}>
                    <div className="flex items-center justify-between text-sm mb-1">
                      <span className="font-medium">{stage.stage}</span>
                      <span className="text-muted-foreground">
                        {formatNumber(stage.value)} ({stage.percentage.toFixed(2)}%)
                      </span>
                    </div>
                    <div className="h-8 w-full overflow-hidden rounded-lg bg-gray-100 dark:bg-gray-800">
                      <div
                        className={`h-full bg-gradient-to-r ${stage.color} transition-all flex items-center justify-end pr-2`}
                        style={{ width: `${Math.max(stage.percentage, 2)}%` }}
                      >
                        {stage.percentage > 10 && (
                          <span className="text-xs font-medium text-white">{formatNumber(stage.value)}</span>
                        )}
                      </div>
                    </div>
                    {index < 2 && (
                      <div className="flex items-center gap-2 mt-1 text-xs text-muted-foreground">
                        <span>Drop-off:</span>
                        <span className="text-red-600">
                          {index === 0 
                            ? `${((1 - clicks / impressions) * 100).toFixed(1)}%`
                            : `${((1 - conversions / clicks) * 100).toFixed(1)}%`
                          }
                        </span>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
