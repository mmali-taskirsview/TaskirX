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
  DollarSign,
  TrendingUp,
  TrendingDown,
  Settings,
  AlertTriangle,
  CheckCircle2,
  Edit,
  Save,
  X,
  Plus,
  Trash2,
  Globe,
  Smartphone,
  BarChart3,
  Zap,
} from 'lucide-react'
import { formatCurrency, formatNumber } from '@/lib/utils'

// Price floor rules
interface PriceFloorRule {
  id: string
  name: string
  format: string
  geo: string
  device: string
  floor: number
  status: 'active' | 'paused'
  fillRate: number
  revenue: number
  impressions: number
  recommendation?: number
}

const priceFloorRules: PriceFloorRule[] = [
  {
    id: '1',
    name: 'US Rewarded Video - Premium',
    format: 'Rewarded Video',
    geo: 'United States',
    device: 'All',
    floor: 25.00,
    status: 'active',
    fillRate: 78.5,
    revenue: 125000,
    impressions: 4200000,
    recommendation: 28.00,
  },
  {
    id: '2',
    name: 'US Playable Ads',
    format: 'Playable Ads',
    geo: 'United States',
    device: 'All',
    floor: 35.00,
    status: 'active',
    fillRate: 65.2,
    revenue: 89000,
    impressions: 1800000,
    recommendation: 32.00,
  },
  {
    id: '3',
    name: 'EU Interstitial',
    format: 'Interstitial',
    geo: 'Europe',
    device: 'All',
    floor: 15.00,
    status: 'active',
    fillRate: 82.3,
    revenue: 67000,
    impressions: 3500000,
  },
  {
    id: '4',
    name: 'Global Banner - Mobile',
    format: 'Banner',
    geo: 'Global',
    device: 'Mobile',
    floor: 2.50,
    status: 'active',
    fillRate: 92.1,
    revenue: 45000,
    impressions: 15000000,
  },
  {
    id: '5',
    name: 'APAC Native Ads',
    format: 'Native Ads',
    geo: 'Asia Pacific',
    device: 'All',
    floor: 12.00,
    status: 'active',
    fillRate: 75.8,
    revenue: 52000,
    impressions: 3200000,
    recommendation: 14.00,
  },
  {
    id: '6',
    name: 'US CTV Premium',
    format: 'CTV/OTT',
    geo: 'United States',
    device: 'CTV',
    floor: 50.00,
    status: 'active',
    fillRate: 58.4,
    revenue: 78000,
    impressions: 980000,
    recommendation: 45.00,
  },
  {
    id: '7',
    name: 'LATAM Rewarded Video',
    format: 'Rewarded Video',
    geo: 'Latin America',
    device: 'All',
    floor: 8.00,
    status: 'paused',
    fillRate: 45.2,
    revenue: 12000,
    impressions: 890000,
    recommendation: 6.00,
  },
]

// Format options
const formatOptions = ['Rewarded Video', 'Playable Ads', 'Interstitial', 'Native Ads', 'Banner', 'CTV/OTT', 'Offerwall', 'Audio']
const geoOptions = ['Global', 'United States', 'Europe', 'Asia Pacific', 'Latin America', 'Middle East', 'Africa']
const deviceOptions = ['All', 'Mobile', 'Desktop', 'Tablet', 'CTV']

export default function PriceFloorManagementPage() {
  const [rules, setRules] = useState(priceFloorRules)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editValue, setEditValue] = useState<number>(0)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [newRule, setNewRule] = useState({
    name: '',
    format: 'Rewarded Video',
    geo: 'Global',
    device: 'All',
    floor: 10.00,
  })

  const totalRevenue = rules.filter(r => r.status === 'active').reduce((sum, r) => sum + r.revenue, 0)
  const avgFillRate = rules.filter(r => r.status === 'active').reduce((sum, r) => sum + r.fillRate, 0) / rules.filter(r => r.status === 'active').length
  const totalImpressions = rules.filter(r => r.status === 'active').reduce((sum, r) => sum + r.impressions, 0)
  const rulesWithRecommendations = rules.filter(r => r.recommendation).length

  const handleSaveFloor = (id: string) => {
    setRules(rules.map(r => r.id === id ? { ...r, floor: editValue } : r))
    setEditingId(null)
  }

  const handleToggleStatus = (id: string) => {
    setRules(rules.map(r => r.id === id ? { ...r, status: r.status === 'active' ? 'paused' : 'active' } : r))
  }

  const handleApplyRecommendation = (id: string) => {
    const rule = rules.find(r => r.id === id)
    if (rule?.recommendation) {
      setRules(rules.map(r => r.id === id ? { ...r, floor: r.recommendation! } : r))
    }
  }

  const handleDeleteRule = (id: string) => {
    setRules(rules.filter(r => r.id !== id))
  }

  const handleCreateRule = () => {
    const newId = (rules.length + 1).toString()
    setRules([...rules, {
      id: newId,
      ...newRule,
      status: 'active',
      fillRate: 70.0,
      revenue: 0,
      impressions: 0,
    }])
    setShowCreateModal(false)
    setNewRule({ name: '', format: 'Rewarded Video', geo: 'Global', device: 'All', floor: 10.00 })
  }

  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Price Floor Management</h1>
          <p className="text-muted-foreground">
            Optimize yield with dynamic price floor rules
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 px-4 py-2 text-sm font-medium text-white shadow-lg hover:opacity-90 transition-opacity"
        >
          <Plus className="h-4 w-4" />
          Create Rule
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Active Rules</p>
                <p className="text-2xl font-bold">{rules.filter(r => r.status === 'active').length}</p>
              </div>
              <div className="rounded-full bg-blue-100 p-3 dark:bg-blue-900/30">
                <Settings className="h-5 w-5 text-blue-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Avg Fill Rate</p>
                <p className="text-2xl font-bold">{avgFillRate.toFixed(1)}%</p>
              </div>
              <div className="rounded-full bg-green-100 p-3 dark:bg-green-900/30">
                <BarChart3 className="h-5 w-5 text-green-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Revenue</p>
                <p className="text-2xl font-bold">{formatCurrency(totalRevenue)}</p>
              </div>
              <div className="rounded-full bg-purple-100 p-3 dark:bg-purple-900/30">
                <DollarSign className="h-5 w-5 text-purple-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className={rulesWithRecommendations > 0 ? 'border-yellow-500/50' : ''}>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Recommendations</p>
                <p className="text-2xl font-bold">{rulesWithRecommendations}</p>
              </div>
              <div className="rounded-full bg-yellow-100 p-3 dark:bg-yellow-900/30">
                <Zap className="h-5 w-5 text-yellow-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* AI Recommendations Banner */}
      {rulesWithRecommendations > 0 && (
        <Card className="border-yellow-500/50 bg-yellow-50/50 dark:bg-yellow-900/10">
          <CardContent className="pt-6">
            <div className="flex items-start gap-4">
              <div className="rounded-full bg-yellow-100 p-2 dark:bg-yellow-900/30">
                <Zap className="h-5 w-5 text-yellow-600" />
              </div>
              <div className="flex-1">
                <h3 className="font-semibold text-yellow-800 dark:text-yellow-400">AI-Powered Recommendations Available</h3>
                <p className="text-sm text-yellow-700 dark:text-yellow-500 mt-1">
                  {rulesWithRecommendations} price floor rules have optimization recommendations based on market analysis and fill rate patterns.
                  Review and apply recommendations to maximize yield.
                </p>
              </div>
              <button className="rounded-lg bg-yellow-600 px-4 py-2 text-sm font-medium text-white hover:bg-yellow-700 transition-colors">
                Review All
              </button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Price Floor Rules Table */}
      <Card>
        <CardHeader>
          <CardTitle>Price Floor Rules</CardTitle>
          <CardDescription>Manage minimum bid prices by format, geo, and device</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b">
                  <th className="pb-3 text-left font-medium">Rule Name</th>
                  <th className="pb-3 text-left font-medium">Format</th>
                  <th className="pb-3 text-left font-medium">Geo</th>
                  <th className="pb-3 text-left font-medium">Device</th>
                  <th className="pb-3 text-right font-medium">Floor (eCPM)</th>
                  <th className="pb-3 text-right font-medium">Fill Rate</th>
                  <th className="pb-3 text-right font-medium">Revenue</th>
                  <th className="pb-3 text-center font-medium">Status</th>
                  <th className="pb-3 text-center font-medium">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {rules.map((rule) => (
                  <tr key={rule.id} className={`hover:bg-muted/50 ${rule.status === 'paused' ? 'opacity-60' : ''}`}>
                    <td className="py-4">
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{rule.name}</span>
                        {rule.recommendation && (
                          <span className="rounded-full bg-yellow-100 px-2 py-0.5 text-xs text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400">
                            AI Rec
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="py-4">{rule.format}</td>
                    <td className="py-4">
                      <div className="flex items-center gap-1">
                        <Globe className="h-4 w-4 text-muted-foreground" />
                        {rule.geo}
                      </div>
                    </td>
                    <td className="py-4">
                      <div className="flex items-center gap-1">
                        <Smartphone className="h-4 w-4 text-muted-foreground" />
                        {rule.device}
                      </div>
                    </td>
                    <td className="py-4 text-right">
                      {editingId === rule.id ? (
                        <div className="flex items-center justify-end gap-2">
                          <span className="text-muted-foreground">$</span>
                          <input
                            type="number"
                            step="0.01"
                            value={editValue}
                            onChange={(e) => setEditValue(parseFloat(e.target.value))}
                            className="w-20 rounded border px-2 py-1 text-right"
                          />
                          <button 
                            onClick={() => handleSaveFloor(rule.id)}
                            className="p-1 text-green-600 hover:bg-green-100 rounded"
                          >
                            <Save className="h-4 w-4" />
                          </button>
                          <button 
                            onClick={() => setEditingId(null)}
                            className="p-1 text-red-600 hover:bg-red-100 rounded"
                          >
                            <X className="h-4 w-4" />
                          </button>
                        </div>
                      ) : (
                        <div className="flex items-center justify-end gap-2">
                          <span className="font-semibold">${rule.floor.toFixed(2)}</span>
                          {rule.recommendation && rule.recommendation !== rule.floor && (
                            <span className={`text-xs ${rule.recommendation > rule.floor ? 'text-green-600' : 'text-red-600'}`}>
                              → ${rule.recommendation.toFixed(2)}
                            </span>
                          )}
                          <button 
                            onClick={() => { setEditingId(rule.id); setEditValue(rule.floor); }}
                            className="p-1 hover:bg-muted rounded"
                          >
                            <Edit className="h-4 w-4" />
                          </button>
                        </div>
                      )}
                    </td>
                    <td className="py-4 text-right">
                      <span className={`rounded-full px-2 py-1 text-xs font-medium ${
                        rule.fillRate >= 80 ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' :
                        rule.fillRate >= 60 ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400' :
                        'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                      }`}>
                        {rule.fillRate.toFixed(1)}%
                      </span>
                    </td>
                    <td className="py-4 text-right font-medium">{formatCurrency(rule.revenue)}</td>
                    <td className="py-4 text-center">
                      <button
                        onClick={() => handleToggleStatus(rule.id)}
                        className={`rounded-full px-3 py-1 text-xs font-medium ${
                          rule.status === 'active'
                            ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                            : 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400'
                        }`}
                      >
                        {rule.status === 'active' ? 'Active' : 'Paused'}
                      </button>
                    </td>
                    <td className="py-4">
                      <div className="flex items-center justify-center gap-1">
                        {rule.recommendation && (
                          <button
                            onClick={() => handleApplyRecommendation(rule.id)}
                            className="p-1.5 rounded hover:bg-yellow-100 text-yellow-600"
                            title="Apply recommendation"
                          >
                            <Zap className="h-4 w-4" />
                          </button>
                        )}
                        <button
                          onClick={() => handleDeleteRule(rule.id)}
                          className="p-1.5 rounded hover:bg-red-100 text-red-600"
                          title="Delete rule"
                        >
                          <Trash2 className="h-4 w-4" />
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

      {/* Optimization Tips */}
      <Card>
        <CardHeader>
          <CardTitle>Optimization Tips</CardTitle>
          <CardDescription>Best practices for price floor management</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <div className="rounded-lg border p-4">
              <div className="flex items-center gap-2 mb-2">
                <TrendingUp className="h-5 w-5 text-green-500" />
                <h4 className="font-semibold">High Fill Rate (&gt;80%)</h4>
              </div>
              <p className="text-sm text-muted-foreground">
                Consider increasing floor prices by 10-15% to capture more value while maintaining demand.
              </p>
            </div>
            <div className="rounded-lg border p-4">
              <div className="flex items-center gap-2 mb-2">
                <AlertTriangle className="h-5 w-5 text-yellow-500" />
                <h4 className="font-semibold">Low Fill Rate (&lt;60%)</h4>
              </div>
              <p className="text-sm text-muted-foreground">
                Floor price may be too high. Test 10-20% reduction to improve fill rate and overall revenue.
              </p>
            </div>
            <div className="rounded-lg border p-4">
              <div className="flex items-center gap-2 mb-2">
                <CheckCircle2 className="h-5 w-5 text-blue-500" />
                <h4 className="font-semibold">Optimal Range (60-80%)</h4>
              </div>
              <p className="text-sm text-muted-foreground">
                Good balance between yield and fill. Monitor regularly and make incremental adjustments.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Create Rule Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-lg rounded-lg bg-white p-6 shadow-xl dark:bg-gray-900">
            <h2 className="text-xl font-bold mb-4">Create Price Floor Rule</h2>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">Rule Name</label>
                <input
                  type="text"
                  value={newRule.name}
                  onChange={(e) => setNewRule({ ...newRule, name: e.target.value })}
                  placeholder="e.g., US Rewarded Video - Premium"
                  className="w-full rounded-lg border px-3 py-2"
                />
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium mb-1">Ad Format</label>
                  <select
                    value={newRule.format}
                    onChange={(e) => setNewRule({ ...newRule, format: e.target.value })}
                    className="w-full rounded-lg border px-3 py-2"
                  >
                    {formatOptions.map(f => <option key={f} value={f}>{f}</option>)}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Geography</label>
                  <select
                    value={newRule.geo}
                    onChange={(e) => setNewRule({ ...newRule, geo: e.target.value })}
                    className="w-full rounded-lg border px-3 py-2"
                  >
                    {geoOptions.map(g => <option key={g} value={g}>{g}</option>)}
                  </select>
                </div>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium mb-1">Device Type</label>
                  <select
                    value={newRule.device}
                    onChange={(e) => setNewRule({ ...newRule, device: e.target.value })}
                    className="w-full rounded-lg border px-3 py-2"
                  >
                    {deviceOptions.map(d => <option key={d} value={d}>{d}</option>)}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Floor Price (eCPM)</label>
                  <div className="relative">
                    <span className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">$</span>
                    <input
                      type="number"
                      step="0.01"
                      value={newRule.floor}
                      onChange={(e) => setNewRule({ ...newRule, floor: parseFloat(e.target.value) })}
                      className="w-full rounded-lg border pl-7 pr-3 py-2"
                    />
                  </div>
                </div>
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
                onClick={handleCreateRule}
                disabled={!newRule.name}
                className="rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 px-4 py-2 text-sm font-medium text-white hover:opacity-90 transition-opacity disabled:opacity-50"
              >
                Create Rule
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
