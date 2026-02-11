'use client'

import { useState } from 'react'
import { Calculator, DollarSign, TrendingUp, Users, Target, BarChart3, Download, RefreshCw, Lightbulb } from 'lucide-react'

const benchmarks = {
  'E-commerce': { avgCtr: 1.8, avgCvr: 2.5, avgCpa: 28, avgRoas: 4.2 },
  'Gaming': { avgCtr: 2.4, avgCvr: 3.2, avgCpa: 18, avgRoas: 5.8 },
  'Finance': { avgCtr: 1.2, avgCvr: 1.8, avgCpa: 65, avgRoas: 3.1 },
  'Travel': { avgCtr: 1.5, avgCvr: 2.1, avgCpa: 42, avgRoas: 3.8 },
  'Health': { avgCtr: 1.6, avgCvr: 2.3, avgCpa: 35, avgRoas: 4.0 },
}

export default function ClientROICalculator() {
  const [industry, setIndustry] = useState('E-commerce')
  const [budget, setBudget] = useState(10000)
  const [cpm, setCpm] = useState(5)
  const [ctr, setCtr] = useState(2.0)
  const [cvr, setCvr] = useState(2.5)
  const [aov, setAov] = useState(75)

  const benchmark = benchmarks[industry as keyof typeof benchmarks]

  // Calculations
  const impressions = (budget / cpm) * 1000
  const clicks = impressions * (ctr / 100)
  const conversions = clicks * (cvr / 100)
  const revenue = conversions * aov
  const cpc = budget / clicks
  const cpa = budget / conversions
  const roas = revenue / budget
  const profit = revenue - budget
  const roi = ((revenue - budget) / budget) * 100

  // LTV Projection (assuming 2.5x multiplier)
  const ltvMultiplier = 2.5
  const projectedLtv = revenue * ltvMultiplier
  const projectedRoas = projectedLtv / budget

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">ROI Calculator</h1>
          <p className="text-gray-500">Project campaign performance and ROI</p>
        </div>
        <div className="flex gap-2">
          <button className="flex items-center gap-2 rounded-lg border border-gray-300 px-4 py-2 text-sm hover:bg-gray-50">
            <RefreshCw className="h-4 w-4" /> Reset
          </button>
          <button className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700">
            <Download className="h-4 w-4" /> Export Report
          </button>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Input Panel */}
        <div className="lg:col-span-1 space-y-6">
          <div className="rounded-xl bg-white p-6 shadow-sm">
            <h2 className="mb-4 text-lg font-semibold text-gray-900">Campaign Inputs</h2>
            
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium text-gray-700">Industry</label>
                <select
                  value={industry}
                  onChange={(e) => setIndustry(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none"
                >
                  {Object.keys(benchmarks).map(ind => (
                    <option key={ind} value={ind}>{ind}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-sm font-medium text-gray-700">Budget ($)</label>
                <input
                  type="number"
                  value={budget}
                  onChange={(e) => setBudget(Number(e.target.value))}
                  className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="text-sm font-medium text-gray-700">CPM ($)</label>
                <input
                  type="number"
                  step="0.1"
                  value={cpm}
                  onChange={(e) => setCpm(Number(e.target.value))}
                  className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="flex items-center justify-between text-sm font-medium text-gray-700">
                  <span>CTR (%)</span>
                  <span className="text-xs text-gray-500">Benchmark: {benchmark.avgCtr}%</span>
                </label>
                <input
                  type="range"
                  min="0.1"
                  max="5"
                  step="0.1"
                  value={ctr}
                  onChange={(e) => setCtr(Number(e.target.value))}
                  className="mt-2 w-full"
                />
                <div className="mt-1 text-center text-lg font-semibold text-blue-600">{ctr}%</div>
              </div>

              <div>
                <label className="flex items-center justify-between text-sm font-medium text-gray-700">
                  <span>Conversion Rate (%)</span>
                  <span className="text-xs text-gray-500">Benchmark: {benchmark.avgCvr}%</span>
                </label>
                <input
                  type="range"
                  min="0.1"
                  max="10"
                  step="0.1"
                  value={cvr}
                  onChange={(e) => setCvr(Number(e.target.value))}
                  className="mt-2 w-full"
                />
                <div className="mt-1 text-center text-lg font-semibold text-blue-600">{cvr}%</div>
              </div>

              <div>
                <label className="text-sm font-medium text-gray-700">Avg. Order Value ($)</label>
                <input
                  type="number"
                  value={aov}
                  onChange={(e) => setAov(Number(e.target.value))}
                  className="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none"
                />
              </div>
            </div>
          </div>

          {/* AI Recommendation */}
          <div className="rounded-xl bg-gradient-to-br from-purple-500 to-purple-600 p-6 text-white">
            <div className="flex items-center gap-2 mb-3">
              <Lightbulb className="h-5 w-5" />
              <h3 className="font-semibold">AI Recommendation</h3>
            </div>
            <p className="text-sm text-purple-100">
              Based on {industry} benchmarks, your CTR of {ctr}% is 
              {ctr >= benchmark.avgCtr ? ' above' : ' below'} average. 
              Consider {ctr < benchmark.avgCtr ? 'A/B testing creatives to improve CTR' : 'increasing budget to scale winning campaigns'}.
            </p>
          </div>
        </div>

        {/* Results Panel */}
        <div className="lg:col-span-2 space-y-6">
          {/* Key Metrics */}
          <div className="grid gap-4 sm:grid-cols-4">
            <div className="rounded-xl bg-white p-5 shadow-sm">
              <div className="flex items-center gap-2 text-gray-500">
                <Users className="h-4 w-4" />
                <span className="text-sm">Impressions</span>
              </div>
              <p className="mt-2 text-2xl font-bold text-gray-900">
                {impressions >= 1000000 ? (impressions / 1000000).toFixed(1) + 'M' : (impressions / 1000).toFixed(0) + 'K'}
              </p>
            </div>
            <div className="rounded-xl bg-white p-5 shadow-sm">
              <div className="flex items-center gap-2 text-gray-500">
                <Target className="h-4 w-4" />
                <span className="text-sm">Clicks</span>
              </div>
              <p className="mt-2 text-2xl font-bold text-gray-900">{clicks.toLocaleString(undefined, { maximumFractionDigits: 0 })}</p>
            </div>
            <div className="rounded-xl bg-white p-5 shadow-sm">
              <div className="flex items-center gap-2 text-gray-500">
                <BarChart3 className="h-4 w-4" />
                <span className="text-sm">Conversions</span>
              </div>
              <p className="mt-2 text-2xl font-bold text-gray-900">{conversions.toLocaleString(undefined, { maximumFractionDigits: 0 })}</p>
            </div>
            <div className="rounded-xl bg-white p-5 shadow-sm">
              <div className="flex items-center gap-2 text-gray-500">
                <DollarSign className="h-4 w-4" />
                <span className="text-sm">Revenue</span>
              </div>
              <p className="mt-2 text-2xl font-bold text-gray-900">${revenue.toLocaleString(undefined, { maximumFractionDigits: 0 })}</p>
            </div>
          </div>

          {/* ROI Summary */}
          <div className="rounded-xl bg-white p-6 shadow-sm">
            <h2 className="mb-4 text-lg font-semibold text-gray-900">ROI Summary</h2>
            <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
              <div className="text-center rounded-lg bg-gray-50 p-4">
                <p className="text-sm text-gray-500">CPC</p>
                <p className="mt-1 text-2xl font-bold text-gray-900">${cpc.toFixed(2)}</p>
              </div>
              <div className="text-center rounded-lg bg-gray-50 p-4">
                <p className="text-sm text-gray-500">CPA</p>
                <p className={`mt-1 text-2xl font-bold ${cpa <= benchmark.avgCpa ? 'text-green-600' : 'text-orange-600'}`}>
                  ${cpa.toFixed(2)}
                </p>
                <p className="text-xs text-gray-400">Benchmark: ${benchmark.avgCpa}</p>
              </div>
              <div className="text-center rounded-lg bg-gray-50 p-4">
                <p className="text-sm text-gray-500">ROAS</p>
                <p className={`mt-1 text-2xl font-bold ${roas >= benchmark.avgRoas ? 'text-green-600' : 'text-orange-600'}`}>
                  {roas.toFixed(1)}x
                </p>
                <p className="text-xs text-gray-400">Benchmark: {benchmark.avgRoas}x</p>
              </div>
              <div className="text-center rounded-lg bg-gray-50 p-4">
                <p className="text-sm text-gray-500">ROI</p>
                <p className={`mt-1 text-2xl font-bold ${roi > 0 ? 'text-green-600' : 'text-red-600'}`}>
                  {roi > 0 ? '+' : ''}{roi.toFixed(0)}%
                </p>
              </div>
            </div>

            {/* Profit/Loss */}
            <div className={`mt-6 rounded-lg p-4 ${profit >= 0 ? 'bg-green-50' : 'bg-red-50'}`}>
              <div className="flex items-center justify-between">
                <span className={`text-sm font-medium ${profit >= 0 ? 'text-green-700' : 'text-red-700'}`}>
                  {profit >= 0 ? 'Projected Profit' : 'Projected Loss'}
                </span>
                <span className={`text-2xl font-bold ${profit >= 0 ? 'text-green-700' : 'text-red-700'}`}>
                  {profit >= 0 ? '+' : ''}${profit.toLocaleString(undefined, { maximumFractionDigits: 0 })}
                </span>
              </div>
            </div>
          </div>

          {/* LTV Projection */}
          <div className="rounded-xl bg-white p-6 shadow-sm">
            <h2 className="mb-4 text-lg font-semibold text-gray-900">LTV Projection (12-month)</h2>
            <p className="text-sm text-gray-500 mb-4">
              Based on industry average LTV multiplier of {ltvMultiplier}x
            </p>
            <div className="grid gap-4 sm:grid-cols-3">
              <div className="rounded-lg border border-blue-200 bg-blue-50 p-4 text-center">
                <p className="text-sm text-blue-600">First Purchase Revenue</p>
                <p className="mt-1 text-xl font-bold text-blue-700">${revenue.toLocaleString(undefined, { maximumFractionDigits: 0 })}</p>
              </div>
              <div className="rounded-lg border border-purple-200 bg-purple-50 p-4 text-center">
                <p className="text-sm text-purple-600">Projected LTV Revenue</p>
                <p className="mt-1 text-xl font-bold text-purple-700">${projectedLtv.toLocaleString(undefined, { maximumFractionDigits: 0 })}</p>
              </div>
              <div className="rounded-lg border border-green-200 bg-green-50 p-4 text-center">
                <p className="text-sm text-green-600">LTV-based ROAS</p>
                <p className="mt-1 text-xl font-bold text-green-700">{projectedRoas.toFixed(1)}x</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
