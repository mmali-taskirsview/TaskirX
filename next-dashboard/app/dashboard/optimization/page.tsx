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
  Zap,
  TrendingUp,
  DollarSign,
  Target,
  Activity,
  Percent,
  RefreshCw,
  Check,
  ChevronDown,
  Settings2,
  Play,
  Pause,
} from 'lucide-react'
import { Button } from '@/components/button'
import { formatCurrency, formatNumber, formatPercentage } from '@/lib/utils'

type PaceType = 'even' | 'aggressive' | 'conservative' | 'asap';
type StatusType = 'healthy' | 'underspending' | 'overspending' | 'depleted';

type BudgetPacingItem = {
  id: string;
  campaign: string;
  status: StatusType;
  spent: number;
  budget: number;
  pace: PaceType;
  isOptimizing: boolean;
};

const initialBudgetPacing: BudgetPacingItem[] = [
  { id: '1', campaign: 'Summer Sale 2026', status: 'healthy', spent: 32450, budget: 50000, pace: 'even', isOptimizing: true },
  { id: '2', campaign: 'New Product Launch', status: 'underspending', spent: 48920, budget: 75000, pace: 'aggressive', isOptimizing: true },
  { id: '3', campaign: 'Brand Awareness Q1', status: 'healthy', spent: 67800, budget: 100000, pace: 'even', isOptimizing: true },
  { id: '4', campaign: 'Mobile App Promotion', status: 'overspending', spent: 41340, budget: 45000, pace: 'conservative', isOptimizing: true },
  { id: '5', campaign: 'Holiday Special', status: 'depleted', spent: 30000, budget: 30000, pace: 'asap', isOptimizing: false },
];

export default function OptimizationPage() {
  const [selectedCampaign, setSelectedCampaign] = useState('all')
  const [retraining, setRetraining] = useState(false)
  const [budgetPacing, setBudgetPacing] = useState<BudgetPacingItem[]>(initialBudgetPacing)
  const [selectedMultiplier, setSelectedMultiplier] = useState<number | null>(null)
  const [applyingMultiplier, setApplyingMultiplier] = useState(false)
  const [multiplierApplied, setMultiplierApplied] = useState(false)
  
  // Recent optimizations with state
  const [recentOptimizations, setRecentOptimizations] = useState([
    {
      id: '1',
      campaign: 'Summer Sale 2026',
      action: 'Increased bid multiplier',
      from: '1.0x',
      to: '1.15x',
      impact: '+12% ROI',
      timestamp: '5 minutes ago',
    },
    {
      id: '2',
      campaign: 'New Product Launch',
      action: 'Budget pacing adjusted',
      from: 'even',
      to: 'aggressive',
      impact: '+8% spend rate',
      timestamp: '15 minutes ago',
    },
    {
      id: '3',
      campaign: 'Brand Awareness Q1',
      action: 'Decreased bid multiplier',
      from: '1.3x',
      to: '1.15x',
      impact: '-15% CPA',
      timestamp: '32 minutes ago',
    },
    {
      id: '4',
      campaign: 'Mobile App Promotion',
      action: 'Budget pacing adjusted',
      from: 'aggressive',
      to: 'conservative',
      impact: '-20% overspend',
      timestamp: '1 hour ago',
    },
  ])

  const optimizationStats = {
    totalCampaigns: 24,
    activeOptimization: budgetPacing.filter(b => b.isOptimizing).length,
    roiImprovement: 0.245,
    budgetSaved: 18450,
    avgMultiplier: 1.23,
    explorationRate: 0.10,
  }

  const [thompsonSamplingData, setThompsonSamplingData] = useState({
    multipliers: [
      { value: 0.5, alpha: 245, beta: 156, winRate: 0.61, trials: 401 },
      { value: 0.7, alpha: 389, beta: 189, winRate: 0.67, trials: 578 },
      { value: 0.85, alpha: 512, beta: 167, winRate: 0.75, trials: 679 },
      { value: 1.0, alpha: 678, beta: 134, winRate: 0.84, trials: 812 },
      { value: 1.15, alpha: 589, beta: 156, winRate: 0.79, trials: 745 },
      { value: 1.3, alpha: 423, beta: 189, winRate: 0.69, trials: 612 },
      { value: 1.5, alpha: 312, beta: 234, winRate: 0.57, trials: 546 },
      { value: 2.0, alpha: 178, beta: 289, winRate: 0.38, trials: 467 },
    ],
  })

  // Retrain model
  const handleRetrainModel = () => {
    setRetraining(true)
    setTimeout(() => {
      // Simulate updated data after retraining
      setThompsonSamplingData({
        multipliers: thompsonSamplingData.multipliers.map(m => ({
          ...m,
          trials: m.trials + Math.floor(Math.random() * 50),
          winRate: Math.min(0.95, m.winRate + (Math.random() * 0.05 - 0.02))
        }))
      })
      setRetraining(false)
    }, 3000)
  }

  // Apply selected multiplier
  const handleApplyMultiplier = () => {
    if (!selectedMultiplier) return
    setApplyingMultiplier(true)
    
    setTimeout(() => {
      // Add to recent optimizations
      const newOpt = {
        id: Date.now().toString(),
        campaign: selectedCampaign === 'all' ? 'All Campaigns' : budgetPacing.find(b => b.id === selectedCampaign)?.campaign || 'Selected Campaign',
        action: 'Applied bid multiplier',
        from: `${optimizationStats.avgMultiplier}x`,
        to: `${selectedMultiplier}x`,
        impact: selectedMultiplier > optimizationStats.avgMultiplier ? '+5% ROI' : '-3% CPA',
        timestamp: 'Just now',
      }
      setRecentOptimizations([newOpt, ...recentOptimizations.slice(0, 3)])
      setApplyingMultiplier(false)
      setMultiplierApplied(true)
      setTimeout(() => {
        setMultiplierApplied(false)
        setSelectedMultiplier(null)
      }, 2000)
    }, 1500)
  }

  // Change pacing strategy
  const handlePacingChange = (campaignId: string, newPace: PaceType) => {
    const campaign = budgetPacing.find(b => b.id === campaignId)
    if (!campaign) return

    setBudgetPacing(budgetPacing.map(b => 
      b.id === campaignId ? { ...b, pace: newPace } : b
    ))

    // Add to recent optimizations
    const newOpt = {
      id: Date.now().toString(),
      campaign: campaign.campaign,
      action: 'Budget pacing adjusted',
      from: campaign.pace,
      to: newPace,
      impact: newPace === 'aggressive' ? '+15% spend rate' : newPace === 'conservative' ? '-10% spend rate' : 'Balanced',
      timestamp: 'Just now',
    }
    setRecentOptimizations([newOpt, ...recentOptimizations.slice(0, 3)])
  }

  // Toggle optimization for campaign
  const toggleOptimization = (campaignId: string) => {
    setBudgetPacing(budgetPacing.map(b => 
      b.id === campaignId ? { ...b, isOptimizing: !b.isOptimizing } : b
    ))
  }

  const getPacingStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'bg-green-100 text-green-700'
      case 'underspending':
        return 'bg-blue-100 text-blue-700'
      case 'overspending':
        return 'bg-orange-100 text-orange-700'
      case 'depleted':
        return 'bg-red-100 text-red-700'
      default:
        return 'bg-gray-100 text-gray-700'
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Bid Optimization</h1>
          <p className="text-muted-foreground">
            Thompson Sampling & Budget Pacing
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <select
            value={selectedCampaign}
            onChange={(e) => setSelectedCampaign(e.target.value)}
            className="h-10 rounded-lg border border-gray-300 bg-white px-4 text-sm"
          >
            <option value="all">All Campaigns</option>
            {budgetPacing.map(b => (
              <option key={b.id} value={b.id}>{b.campaign}</option>
            ))}
          </select>
          <Button 
            variant="outline"
            onClick={handleRetrainModel}
            disabled={retraining}
          >
            <RefreshCw className={`mr-2 h-4 w-4 ${retraining ? 'animate-spin' : ''}`} />
            {retraining ? 'Training...' : 'Retrain Model'}
          </Button>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                ROI Improvement
              </CardTitle>
              <TrendingUp className="h-4 w-4 text-green-500" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              +{formatPercentage(optimizationStats.roiImprovement)}
            </div>
            <p className="text-xs text-muted-foreground">
              Across {optimizationStats.activeOptimization} campaigns
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Budget Saved
              </CardTitle>
              <DollarSign className="h-4 w-4 text-blue-500" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">
              {formatCurrency(optimizationStats.budgetSaved)}
            </div>
            <p className="text-xs text-muted-foreground">
              Through smart pacing
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Avg Bid Multiplier
              </CardTitle>
              <Percent className="h-4 w-4 text-purple-500" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{optimizationStats.avgMultiplier}x</div>
            <p className="text-xs text-muted-foreground">
              {formatPercentage(optimizationStats.explorationRate)} exploration
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Thompson Sampling Status */}
      <Card className="border-l-4 border-l-purple-500">
        <CardHeader>
          <div className="flex items-center space-x-2">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-purple-500 to-pink-600">
              <Zap className="h-5 w-5 text-white" />
            </div>
            <div>
              <CardTitle>Thompson Sampling Algorithm</CardTitle>
              <CardDescription>Multi-Armed Bandit Optimization</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="grid gap-4 md:grid-cols-4">
              <div>
                <div className="text-sm text-muted-foreground">Algorithm</div>
                <div className="text-lg font-semibold">Beta Distribution</div>
              </div>
              <div>
                <div className="text-sm text-muted-foreground">Bid Arms</div>
                <div className="text-lg font-semibold">8 multipliers</div>
              </div>
              <div>
                <div className="text-sm text-muted-foreground">Total Trials</div>
                <div className="text-lg font-semibold">
                  {formatNumber(thompsonSamplingData.multipliers.reduce((sum, m) => sum + m.trials, 0))}
                </div>
              </div>
              <div>
                <div className="text-sm text-muted-foreground">Status</div>
                <div className="flex items-center">
                  <span className="mr-2 h-2 w-2 rounded-full bg-green-500" />
                  <span className="text-lg font-semibold text-green-600">Learning</span>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Bid Multiplier Performance */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Bid Multiplier Performance</CardTitle>
              <CardDescription>Click to select a multiplier, then apply</CardDescription>
            </div>
            {selectedMultiplier && (
              <Button 
                onClick={handleApplyMultiplier}
                disabled={applyingMultiplier}
                className={multiplierApplied ? 'bg-green-600' : 'bg-blue-600 hover:bg-blue-700'}
              >
                {multiplierApplied ? (
                  <>
                    <Check className="mr-2 h-4 w-4" />
                    Applied!
                  </>
                ) : applyingMultiplier ? (
                  <>
                    <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                    Applying...
                  </>
                ) : (
                  <>
                    <Zap className="mr-2 h-4 w-4" />
                    Apply {selectedMultiplier}x Multiplier
                  </>
                )}
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {thompsonSamplingData.multipliers.map((multiplier, index) => {
              const isOptimal = multiplier.winRate === Math.max(...thompsonSamplingData.multipliers.map(m => m.winRate))
              const isSelected = selectedMultiplier === multiplier.value
              return (
                <div 
                  key={index} 
                  onClick={() => setSelectedMultiplier(multiplier.value)}
                  className={`rounded-lg border-2 p-4 cursor-pointer transition-all ${
                    isSelected ? 'border-blue-500 bg-blue-50 ring-2 ring-blue-200' :
                    isOptimal ? 'border-green-500 bg-green-50 hover:bg-green-100' : 
                    'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-4">
                      <div className={`flex h-12 w-12 items-center justify-center rounded-lg ${
                        isSelected ? 'bg-blue-500' :
                        isOptimal ? 'bg-green-500' : 
                        'bg-gradient-to-br from-blue-500 to-purple-600'
                      } text-white font-bold`}>
                        {multiplier.value}x
                      </div>
                      <div>
                        <div className="flex items-center space-x-2">
                          <span className="font-semibold">
                            Win Rate: {formatPercentage(multiplier.winRate)}
                          </span>
                          {isOptimal && (
                            <span className="rounded-full bg-green-500 px-2 py-0.5 text-xs font-semibold text-white">
                              Optimal
                            </span>
                          )}
                          {isSelected && (
                            <span className="rounded-full bg-blue-500 px-2 py-0.5 text-xs font-semibold text-white">
                              Selected
                            </span>
                          )}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          α={multiplier.alpha}, β={multiplier.beta} | {multiplier.trials} trials
                        </div>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-sm font-medium text-muted-foreground">Confidence</div>
                      <div className="text-lg font-bold">
                        {formatPercentage(multiplier.alpha / (multiplier.alpha + multiplier.beta))}
                      </div>
                    </div>
                  </div>
                  <div className="mt-3">
                    <div className="h-2 w-full overflow-hidden rounded-full bg-gray-200">
                      <div
                        className={`h-full ${
                          isSelected ? 'bg-blue-500' :
                          isOptimal ? 'bg-green-500' : 
                          'bg-gradient-to-r from-blue-500 to-purple-600'
                        }`}
                        style={{ width: `${multiplier.winRate * 100}%` }}
                      />
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </CardContent>
      </Card>

      {/* Budget Pacing & Recent Optimizations */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Budget Pacing Status</CardTitle>
            <CardDescription>Click pacing to change strategy</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {budgetPacing.map((campaign) => (
                <div key={campaign.id} className="rounded-lg border border-gray-200 p-3">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{campaign.campaign}</span>
                        <button
                          onClick={() => toggleOptimization(campaign.id)}
                          className={`text-xs px-2 py-0.5 rounded-full ${
                            campaign.isOptimizing 
                              ? 'bg-green-100 text-green-700' 
                              : 'bg-gray-100 text-gray-600'
                          }`}
                          title={campaign.isOptimizing ? 'Click to pause' : 'Click to enable'}
                        >
                          {campaign.isOptimizing ? (
                            <span className="flex items-center gap-1">
                              <span className="h-1.5 w-1.5 rounded-full bg-green-500 animate-pulse" />
                              Active
                            </span>
                          ) : (
                            <span className="flex items-center gap-1">
                              <Pause className="h-3 w-3" />
                              Paused
                            </span>
                          )}
                        </button>
                      </div>
                      <div className="mt-1 flex items-center space-x-2">
                        <span
                          className={`rounded-full px-2 py-0.5 text-xs font-semibold ${getPacingStatusColor(
                            campaign.status
                          )}`}
                        >
                          {campaign.status}
                        </span>
                        <select
                          value={campaign.pace}
                          onChange={(e) => handlePacingChange(campaign.id, e.target.value as PaceType)}
                          className="text-xs border border-gray-200 rounded px-2 py-0.5 bg-white cursor-pointer hover:border-blue-400"
                        >
                          <option value="even">Even Pacing</option>
                          <option value="aggressive">Aggressive</option>
                          <option value="conservative">Conservative</option>
                          <option value="asap">ASAP</option>
                        </select>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-sm font-semibold">
                        {formatCurrency(campaign.spent)} / {formatCurrency(campaign.budget)}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {formatPercentage(campaign.spent / campaign.budget)} utilized
                      </div>
                    </div>
                  </div>
                  <div className="mt-2">
                    <div className="h-1.5 w-full overflow-hidden rounded-full bg-gray-200">
                      <div
                        className={`h-full ${
                          campaign.status === 'overspending' || campaign.status === 'depleted'
                            ? 'bg-red-500'
                            : campaign.status === 'underspending'
                            ? 'bg-blue-500'
                            : 'bg-green-500'
                        }`}
                        style={{ width: `${Math.min((campaign.spent / campaign.budget) * 100, 100)}%` }}
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Recent Optimizations</CardTitle>
            <CardDescription>Latest AI-driven adjustments</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {recentOptimizations.map((opt) => (
                <div key={opt.id} className="rounded-lg border border-gray-200 p-3">
                  <div className="flex items-start space-x-3">
                    <div className={`flex h-8 w-8 items-center justify-center rounded-lg ${
                      opt.timestamp === 'Just now' 
                        ? 'bg-gradient-to-br from-green-500 to-emerald-600' 
                        : 'bg-gradient-to-br from-purple-500 to-pink-600'
                    }`}>
                      <Activity className="h-4 w-4 text-white" />
                    </div>
                    <div className="flex-1">
                      <div className="font-medium">{opt.campaign}</div>
                      <div className="text-sm text-muted-foreground">{opt.action}</div>
                      <div className="mt-1 flex items-center space-x-2 text-xs">
                        <span className="rounded bg-gray-100 px-2 py-0.5 font-mono">
                          {opt.from}
                        </span>
                        <span>→</span>
                        <span className="rounded bg-blue-100 px-2 py-0.5 font-mono text-blue-700">
                          {opt.to}
                        </span>
                      </div>
                      <div className="mt-2 flex items-center justify-between">
                        <span className="text-xs font-semibold text-green-600">{opt.impact}</span>
                        <span className={`text-xs ${opt.timestamp === 'Just now' ? 'text-green-600 font-semibold' : 'text-muted-foreground'}`}>
                          {opt.timestamp}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
