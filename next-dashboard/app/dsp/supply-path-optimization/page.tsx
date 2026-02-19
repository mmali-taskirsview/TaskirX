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
  CheckCircle,
  Route,
  Network,
  Timer,
  Coins
} from 'lucide-react'

interface SupplyPathMetrics {
  timeRange: string
  totalRequests: number
  successfulBids: number
  winRate: number
  avgLatencyMs: number
  avgTotalFees: number
  serviceMetrics: {
    [serviceName: string]: {
      serviceName: string
      totalCalls: number
      successRate: number
      avgLatencyMs: number
      errorRate: number
      totalFees: number
    }
  }
  pathEfficiency: number
  timestamp: string
}

interface OptimizationSuggestion {
  type: string
  service: string
  description: string
  priority: 'high' | 'medium' | 'low'
  savings: number
}

interface SupplyPathOptimization {
  optimizations: OptimizationSuggestion[]
  estimatedSavings: number
}

interface DirectPublisherOpportunity {
  serviceName: string
  currentFeeRate: number
  estimatedDirectFee: number
  successRate: number
  monthlyVolume: number
  priority: string
  riskLevel: string
  estimatedSavings: number
  roi: number
}

interface DirectPublisherAnalysis {
  timeRange: string
  timestamp: string
  currentHops: number
  opportunities: DirectPublisherOpportunity[]
}

interface OptimizationScenario {
  name: string
  description: string
  estimatedCostReduction: number
  estimatedLatencyReduction: number
  riskLevel: string
  implementationEffort: string
  timeToValue: string
  netBenefit: number
  breakEvenMonths: number
}

interface CostBenefitAnalysis {
  timeRange: string
  timestamp: string
  currentTotalCost: number
  currentWinRate: number
  currentAvgLatency: number
  scenarios: OptimizationScenario[]
}

export default function SupplyPathOptimizationPage() {
  const [metrics, setMetrics] = useState<SupplyPathMetrics | null>(null)
  const [optimizations, setOptimizations] = useState<SupplyPathOptimization | null>(null)
  const [directPublisherAnalysis, setDirectPublisherAnalysis] = useState<DirectPublisherAnalysis | null>(null)
  const [costBenefitAnalysis, setCostBenefitAnalysis] = useState<CostBenefitAnalysis | null>(null)
  const [loading, setLoading] = useState(true)
  const [timeRange, setTimeRange] = useState('1h')

  const fetchData = async () => {
    try {
      setLoading(true)

      // TODO: Implement backend endpoints
      // For now, use mock data
      const mockMetrics: SupplyPathMetrics = {
        timeRange: timeRange,
        totalRequests: 12543,
        successfulBids: 8921,
        winRate: 0.71,
        avgLatencyMs: 45.2,
        avgTotalFees: 0.0032,
        serviceMetrics: {
          'fraud-detection': {
            serviceName: 'fraud-detection',
            totalCalls: 12543,
            successRate: 0.98,
            avgLatencyMs: 12.5,
            errorRate: 0.02,
            totalFees: 0.001
          },
          'ad-matching': {
            serviceName: 'ad-matching',
            totalCalls: 12301,
            successRate: 0.95,
            avgLatencyMs: 18.7,
            errorRate: 0.05,
            totalFees: 0.002
          },
          'bid-optimizer': {
            serviceName: 'bid-optimizer',
            totalCalls: 8921,
            successRate: 0.97,
            avgLatencyMs: 14.3,
            errorRate: 0.03,
            totalFees: 0.0015
          }
        },
        pathEfficiency: 0.89,
        timestamp: new Date().toISOString()
      }

      const mockOptimizations: SupplyPathOptimization = {
        optimizations: [
          {
            type: 'cache',
            service: 'fraud-detection',
            description: 'High latency detected (12.5ms avg). Consider caching responses.',
            priority: 'medium',
            savings: 2.34
          },
          {
            type: 'circuit_breaker',
            service: 'ad-matching',
            description: 'Low success rate (95%). Circuit breaker may be tripping too frequently.',
            priority: 'high',
            savings: 0
          },
          {
            type: 'fee_negotiation',
            service: 'bid-optimizer',
            description: 'High fees detected ($0.0015 per call). Consider direct integration.',
            priority: 'low',
            savings: 1.12
          }
        ],
        estimatedSavings: 3.46
      }

      setMetrics(mockMetrics)
      setOptimizations(mockOptimizations)

      // Mock data for advanced analytics
      const mockDirectPublisherAnalysis: DirectPublisherAnalysis = {
        timeRange: timeRange,
        timestamp: new Date().toISOString(),
        currentHops: 3,
        opportunities: [
          {
            serviceName: 'fraud-detection',
            currentFeeRate: 0.0012,
            estimatedDirectFee: 0.0008,
            successRate: 0.92,
            monthlyVolume: 45000,
            priority: 'high',
            riskLevel: 'medium',
            estimatedSavings: 1800,
            roi: 75.0
          },
          {
            serviceName: 'ad-matching',
            currentFeeRate: 0.0021,
            estimatedDirectFee: 0.0014,
            successRate: 0.89,
            monthlyVolume: 42000,
            priority: 'high',
            riskLevel: 'medium',
            estimatedSavings: 2940,
            roi: 82.5
          }
        ]
      }

      const mockCostBenefitAnalysis: CostBenefitAnalysis = {
        timeRange: timeRange,
        timestamp: new Date().toISOString(),
        currentTotalCost: 12500,
        currentWinRate: 0.71,
        currentAvgLatency: 45.2,
        scenarios: [
          {
            name: 'Direct Publisher Connections',
            description: 'Bypass intermediaries and connect directly with publishers',
            estimatedCostReduction: 0.6,
            estimatedLatencyReduction: 18.08,
            riskLevel: 'medium',
            implementationEffort: 'high',
            timeToValue: '3-6 months',
            netBenefit: 22500,
            breakEvenMonths: 2
          },
          {
            name: 'Service Performance Optimization',
            description: 'Optimize existing services for better performance and cost efficiency',
            estimatedCostReduction: 0.25,
            estimatedLatencyReduction: 9.04,
            riskLevel: 'low',
            implementationEffort: 'medium',
            timeToValue: '1-3 months',
            netBenefit: 9375,
            breakEvenMonths: 1
          },
          {
            name: 'Hybrid Optimization',
            description: 'Combine direct connections with service optimizations',
            estimatedCostReduction: 0.75,
            estimatedLatencyReduction: 22.6,
            riskLevel: 'medium',
            implementationEffort: 'high',
            timeToValue: '2-4 months',
            netBenefit: 28125,
            breakEvenMonths: 3
          }
        ]
      }

      setDirectPublisherAnalysis(mockDirectPublisherAnalysis)
      setCostBenefitAnalysis(mockCostBenefitAnalysis)

      // Fetch real data from backend
      try {
        const metricsResponse = await api.getSupplyChainMetrics(timeRange)
        setMetrics(metricsResponse.data)
        const directPublisherResponse = await api.getDirectPublisherAnalysis(timeRange)
        setDirectPublisherAnalysis(directPublisherResponse.data)
        const costBenefitResponse = await api.getCostBenefitAnalysis(timeRange)
        setCostBenefitAnalysis(costBenefitResponse.data)
      } catch (apiError) {
        console.warn('API calls failed, using mock data:', apiError)
        // Keep mock data as fallback
      }

    } catch (error) {
      console.error('Failed to fetch SPO data:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [timeRange])

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'bg-red-100 text-red-800'
      case 'medium': return 'bg-yellow-100 text-yellow-800'
      case 'low': return 'bg-green-100 text-green-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const getOptimizationIcon = (type: string) => {
    switch (type) {
      case 'cache': return <Zap className="h-4 w-4" />
      case 'circuit_breaker': return <AlertTriangle className="h-4 w-4" />
      case 'direct_connection': return <Network className="h-4 w-4" />
      case 'fee_negotiation': return <Coins className="h-4 w-4" />
      default: return <Target className="h-4 w-4" />
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <RefreshCw className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Supply Path Optimization</h1>
          <p className="text-gray-600 mt-1">Monitor and optimize your bidding supply chain performance</p>
        </div>
        <div className="flex items-center gap-4">
          <select
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md text-sm"
          >
            <option value="1h">Last Hour</option>
            <option value="24h">Last 24 Hours</option>
            <option value="7d">Last 7 Days</option>
          </select>
          <Button onClick={fetchData} variant="outline" size="sm">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Key Metrics */}
      {metrics && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{metrics.totalRequests.toLocaleString()}</div>
              <p className="text-xs text-muted-foreground">
                Supply chain requests
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Win Rate</CardTitle>
              <Target className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{(metrics.winRate * 100).toFixed(1)}%</div>
              <p className="text-xs text-muted-foreground">
                Successful bids
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Avg Latency</CardTitle>
              <Clock className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{metrics.avgLatencyMs.toFixed(0)}ms</div>
              <p className="text-xs text-muted-foreground">
                End-to-end response time
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Fees</CardTitle>
              <DollarSign className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">${metrics.avgTotalFees.toFixed(4)}</div>
              <p className="text-xs text-muted-foreground">
                Average per request
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Service Performance */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5" />
              Service Performance
            </CardTitle>
            <CardDescription>
              Performance metrics for each service in the supply chain
            </CardDescription>
          </CardHeader>
          <CardContent>
            {metrics && Object.keys(metrics.serviceMetrics).length > 0 ? (
              <div className="space-y-4">
                {Object.entries(metrics.serviceMetrics).map(([serviceName, service]) => (
                  <div key={serviceName} className="flex items-center justify-between p-3 border rounded-lg">
                    <div className="flex-1">
                      <div className="font-medium text-sm">{service.serviceName}</div>
                      <div className="text-xs text-gray-500">
                        {service.totalCalls} calls • {(service.successRate * 100).toFixed(1)}% success
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-sm font-medium">{service.avgLatencyMs.toFixed(0)}ms</div>
                      <div className="text-xs text-gray-500">${service.totalFees.toFixed(4)}</div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-gray-500">
                No service metrics available
              </div>
            )}
          </CardContent>
        </Card>

        {/* Optimization Suggestions */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Target className="h-5 w-5" />
              Optimization Opportunities
            </CardTitle>
            <CardDescription>
              AI-powered recommendations to improve supply chain efficiency
            </CardDescription>
          </CardHeader>
          <CardContent>
            {optimizations && optimizations.optimizations.length > 0 ? (
              <div className="space-y-4">
                <div className="flex items-center justify-between p-3 bg-green-50 border border-green-200 rounded-lg">
                  <div className="flex items-center gap-2">
                    <CheckCircle className="h-4 w-4 text-green-600" />
                    <span className="text-sm font-medium text-green-800">Estimated Savings</span>
                  </div>
                  <span className="text-lg font-bold text-green-800">
                    ${optimizations.estimatedSavings.toFixed(2)}/hour
                  </span>
                </div>

                {optimizations.optimizations.map((opt, index) => (
                  <div key={index} className="flex items-start gap-3 p-3 border rounded-lg">
                    <div className="mt-0.5">
                      {getOptimizationIcon(opt.type)}
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium text-sm">{opt.service}</span>
                        <Badge className={`text-xs ${getPriorityColor(opt.priority)}`}>
                          {opt.priority}
                        </Badge>
                      </div>
                      <p className="text-sm text-gray-600 mb-2">{opt.description}</p>
                      {opt.savings > 0 && (
                        <div className="text-xs text-green-600 font-medium">
                          Potential savings: ${opt.savings.toFixed(2)}/hour
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-gray-500">
                No optimization opportunities found
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Supply Chain Visualization */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Route className="h-5 w-5" />
            Supply Chain Flow
          </CardTitle>
          <CardDescription>
            Visual representation of your bidding supply chain
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <Network className="h-16 w-16 text-gray-300 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">Supply Chain Visualization</h3>
              <p className="text-gray-500 max-w-md">
                Interactive supply chain flow diagram will be available in the next update.
                This will show the complete path from publisher to winning campaign.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Advanced Analytics Sections */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Direct Publisher Analysis */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Target className="h-5 w-5" />
              Direct Publisher Opportunities
            </CardTitle>
            <CardDescription>
              Identify opportunities to bypass intermediaries and connect directly with publishers
            </CardDescription>
          </CardHeader>
          <CardContent>
            {directPublisherAnalysis && directPublisherAnalysis.opportunities.length > 0 ? (
              <div className="space-y-4">
                <div className="flex items-center justify-between p-3 bg-blue-50 border border-blue-200 rounded-lg">
                  <div className="flex items-center gap-2">
                    <Coins className="h-4 w-4 text-blue-600" />
                    <span className="text-sm font-medium text-blue-800">Current Supply Chain</span>
                  </div>
                  <span className="text-sm font-bold text-blue-800">
                    {directPublisherAnalysis.currentHops} hops
                  </span>
                </div>

                {directPublisherAnalysis.opportunities.map((opportunity, index) => (
                  <div key={index} className="flex items-start gap-3 p-3 border rounded-lg">
                    <div className="mt-0.5">
                      <TrendingUp className="h-4 w-4 text-green-600" />
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium text-sm">{opportunity.serviceName}</span>
                        <Badge className={`text-xs ${opportunity.priority === 'high' ? 'bg-red-100 text-red-800' : 'bg-yellow-100 text-yellow-800'}`}>
                          {opportunity.priority}
                        </Badge>
                        <Badge variant="outline" className="text-xs">
                          {opportunity.riskLevel} risk
                        </Badge>
                      </div>
                      <div className="text-xs text-gray-600 mb-2">
                        Current: ${(opportunity.currentFeeRate * 10000).toFixed(2)}/1k calls → Direct: ${(opportunity.estimatedDirectFee * 10000).toFixed(2)}/1k calls
                      </div>
                      <div className="flex items-center justify-between">
                        <div className="text-xs text-green-600 font-medium">
                          ${opportunity.estimatedSavings.toLocaleString()}/month savings
                        </div>
                        <div className="text-xs text-blue-600 font-medium">
                          {opportunity.roi.toFixed(1)}% ROI
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-gray-500">
                No direct publisher opportunities identified
              </div>
            )}
          </CardContent>
        </Card>

        {/* Cost-Benefit Analysis */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5" />
              Optimization Scenarios
            </CardTitle>
            <CardDescription>
              Compare different optimization strategies with detailed cost-benefit analysis
            </CardDescription>
          </CardHeader>
          <CardContent>
            {costBenefitAnalysis && costBenefitAnalysis.scenarios.length > 0 ? (
              <div className="space-y-4">
                <div className="flex items-center justify-between p-3 bg-gray-50 border rounded-lg">
                  <div className="flex items-center gap-2">
                    <DollarSign className="h-4 w-4 text-gray-600" />
                    <span className="text-sm font-medium text-gray-800">Current Monthly Cost</span>
                  </div>
                  <span className="text-lg font-bold text-gray-800">
                    ${costBenefitAnalysis.currentTotalCost.toLocaleString()}
                  </span>
                </div>

                {costBenefitAnalysis.scenarios.map((scenario, index) => (
                  <div key={index} className="p-4 border rounded-lg">
                    <div className="flex items-start justify-between mb-3">
                      <div>
                        <h4 className="font-medium text-sm mb-1">{scenario.name}</h4>
                        <p className="text-xs text-gray-600">{scenario.description}</p>
                      </div>
                      <div className="text-right">
                        <div className="text-sm font-bold text-green-600">
                          ${scenario.netBenefit.toLocaleString()}/month
                        </div>
                        <div className="text-xs text-gray-500">
                          {scenario.breakEvenMonths} month breakeven
                        </div>
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4 text-xs">
                      <div>
                        <span className="text-gray-500">Cost Reduction:</span>
                        <span className="ml-1 font-medium text-green-600">
                          {(scenario.estimatedCostReduction * 100).toFixed(0)}%
                        </span>
                      </div>
                      <div>
                        <span className="text-gray-500">Latency Reduction:</span>
                        <span className="ml-1 font-medium text-blue-600">
                          {scenario.estimatedLatencyReduction.toFixed(1)}ms
                        </span>
                      </div>
                      <div>
                        <span className="text-gray-500">Risk:</span>
                        <Badge className={`text-xs ml-1 ${scenario.riskLevel === 'low' ? 'bg-green-100 text-green-800' : scenario.riskLevel === 'medium' ? 'bg-yellow-100 text-yellow-800' : 'bg-red-100 text-red-800'}`}>
                          {scenario.riskLevel}
                        </Badge>
                      </div>
                      <div>
                        <span className="text-gray-500">Time to Value:</span>
                        <span className="ml-1 font-medium">{scenario.timeToValue}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-gray-500">
                No optimization scenarios available
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}