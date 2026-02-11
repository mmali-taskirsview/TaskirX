'use client'

import { useState } from 'react'
import { Plus, Play, Pause, TrendingUp, TrendingDown, BarChart3, Users, Target, CheckCircle, Clock, AlertCircle } from 'lucide-react'

const abTests = [
  {
    id: 1,
    name: 'Banner Color Test',
    status: 'running',
    campaign: 'Summer Sale 2026',
    startDate: '2026-01-20',
    variants: [
      { name: 'Control (Blue)', traffic: 50, conversions: 245, ctr: 2.1, revenue: 12450 },
      { name: 'Variant A (Green)', traffic: 50, conversions: 312, ctr: 2.8, revenue: 15890 },
    ],
    winner: 'Variant A (Green)',
    confidence: 95,
    lift: '+27.3%',
  },
  {
    id: 2,
    name: 'CTA Text Test',
    status: 'running',
    campaign: 'App Install Campaign',
    startDate: '2026-01-25',
    variants: [
      { name: 'Control (Download Now)', traffic: 33, conversions: 189, ctr: 2.4, revenue: 8900 },
      { name: 'Variant A (Get Started)', traffic: 33, conversions: 167, ctr: 2.1, revenue: 7800 },
      { name: 'Variant B (Try Free)', traffic: 34, conversions: 234, ctr: 2.9, revenue: 11200 },
    ],
    winner: 'Variant B (Try Free)',
    confidence: 89,
    lift: '+23.8%',
  },
  {
    id: 3,
    name: 'Video Length Test',
    status: 'completed',
    campaign: 'Brand Awareness CTV',
    startDate: '2026-01-10',
    endDate: '2026-01-24',
    variants: [
      { name: 'Control (15s)', traffic: 50, conversions: 456, ctr: 0.6, revenue: 22000 },
      { name: 'Variant A (30s)', traffic: 50, conversions: 523, ctr: 0.5, revenue: 28500 },
    ],
    winner: 'Variant A (30s)',
    confidence: 98,
    lift: '+14.7%',
  },
  {
    id: 4,
    name: 'Landing Page Test',
    status: 'draft',
    campaign: 'Holiday Promo',
    variants: [
      { name: 'Control', traffic: 50, conversions: 0, ctr: 0, revenue: 0 },
      { name: 'Variant A', traffic: 50, conversions: 0, ctr: 0, revenue: 0 },
    ],
    winner: null,
    confidence: 0,
    lift: null,
  },
]

export default function ClientABTesting() {
  const [showCreate, setShowCreate] = useState(false)
  const [filter, setFilter] = useState('all')

  const filteredTests = abTests.filter(test => filter === 'all' || test.status === filter)

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'running':
        return <span className="flex items-center gap-1 rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-700"><span className="h-1.5 w-1.5 rounded-full bg-green-500 animate-pulse" /> Running</span>
      case 'completed':
        return <span className="flex items-center gap-1 rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-700"><CheckCircle className="h-3 w-3" /> Completed</span>
      case 'draft':
        return <span className="flex items-center gap-1 rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-700"><Clock className="h-3 w-3" /> Draft</span>
      default:
        return null
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">A/B Testing</h1>
          <p className="text-gray-500">Create and manage split tests for your campaigns</p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
        >
          <Plus className="h-5 w-5" />
          New Test
        </button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-4">
        <div className="rounded-xl bg-white p-5 shadow-sm">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-green-100 p-2"><Play className="h-5 w-5 text-green-600" /></div>
            <div>
              <p className="text-sm text-gray-500">Running Tests</p>
              <p className="text-2xl font-bold text-gray-900">2</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-white p-5 shadow-sm">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-blue-100 p-2"><CheckCircle className="h-5 w-5 text-blue-600" /></div>
            <div>
              <p className="text-sm text-gray-500">Completed</p>
              <p className="text-2xl font-bold text-gray-900">1</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-white p-5 shadow-sm">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-purple-100 p-2"><TrendingUp className="h-5 w-5 text-purple-600" /></div>
            <div>
              <p className="text-sm text-gray-500">Avg. Lift</p>
              <p className="text-2xl font-bold text-gray-900">+21.9%</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-white p-5 shadow-sm">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-orange-100 p-2"><Target className="h-5 w-5 text-orange-600" /></div>
            <div>
              <p className="text-sm text-gray-500">Avg. Confidence</p>
              <p className="text-2xl font-bold text-gray-900">94%</p>
            </div>
          </div>
        </div>
      </div>

      {/* Filter */}
      <div className="flex gap-2">
        {['all', 'running', 'completed', 'draft'].map(status => (
          <button
            key={status}
            onClick={() => setFilter(status)}
            className={`rounded-lg px-4 py-2 text-sm font-medium capitalize ${
              filter === status ? 'bg-blue-600 text-white' : 'bg-white text-gray-700 hover:bg-gray-50'
            }`}
          >
            {status === 'all' ? 'All Tests' : status}
          </button>
        ))}
      </div>

      {/* Tests List */}
      <div className="space-y-4">
        {filteredTests.map(test => (
          <div key={test.id} className="rounded-xl bg-white p-6 shadow-sm">
            <div className="flex items-start justify-between">
              <div>
                <div className="flex items-center gap-3">
                  <h3 className="text-lg font-semibold text-gray-900">{test.name}</h3>
                  {getStatusBadge(test.status)}
                </div>
                <p className="mt-1 text-sm text-gray-500">Campaign: {test.campaign}</p>
                <p className="text-xs text-gray-400">
                  Started: {test.startDate} {test.endDate && `• Ended: ${test.endDate}`}
                </p>
              </div>
              {test.status === 'running' && (
                <button className="flex items-center gap-1 rounded-lg border border-yellow-300 bg-yellow-50 px-3 py-1.5 text-sm font-medium text-yellow-700 hover:bg-yellow-100">
                  <Pause className="h-4 w-4" /> Pause
                </button>
              )}
            </div>

            {/* Variants */}
            <div className="mt-4 space-y-3">
              {test.variants.map((variant, idx) => {
                const isWinner = test.winner === variant.name
                return (
                  <div key={idx} className={`rounded-lg border p-4 ${isWinner ? 'border-green-300 bg-green-50' : 'border-gray-200'}`}>
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className={`font-medium ${isWinner ? 'text-green-700' : 'text-gray-900'}`}>
                          {variant.name}
                        </span>
                        {isWinner && (
                          <span className="flex items-center gap-1 rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">
                            <TrendingUp className="h-3 w-3" /> Winner
                          </span>
                        )}
                      </div>
                      <span className="text-sm text-gray-500">{variant.traffic}% traffic</span>
                    </div>
                    <div className="mt-2 grid grid-cols-4 gap-4 text-sm">
                      <div>
                        <p className="text-gray-500">Conversions</p>
                        <p className="font-semibold text-gray-900">{variant.conversions.toLocaleString()}</p>
                      </div>
                      <div>
                        <p className="text-gray-500">CTR</p>
                        <p className="font-semibold text-gray-900">{variant.ctr}%</p>
                      </div>
                      <div>
                        <p className="text-gray-500">Revenue</p>
                        <p className="font-semibold text-gray-900">${variant.revenue.toLocaleString()}</p>
                      </div>
                      <div>
                        <p className="text-gray-500">Confidence</p>
                        <div className="flex items-center gap-2">
                          <div className="h-2 flex-1 rounded-full bg-gray-200">
                            <div 
                              className={`h-2 rounded-full ${test.confidence >= 95 ? 'bg-green-500' : test.confidence >= 80 ? 'bg-yellow-500' : 'bg-gray-400'}`}
                              style={{ width: `${test.confidence}%` }}
                            />
                          </div>
                          <span className="font-semibold text-gray-900">{test.confidence}%</span>
                        </div>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>

            {/* Result Summary */}
            {test.lift && (
              <div className="mt-4 flex items-center justify-between rounded-lg bg-gray-50 px-4 py-3">
                <span className="text-sm text-gray-600">
                  {test.status === 'completed' ? 'Final Result:' : 'Current Result:'} <strong>{test.winner}</strong> is outperforming
                </span>
                <span className="flex items-center gap-1 text-lg font-bold text-green-600">
                  <TrendingUp className="h-5 w-5" /> {test.lift}
                </span>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
