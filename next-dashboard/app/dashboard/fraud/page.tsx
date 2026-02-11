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
  Shield,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Activity,
  Globe,
  Smartphone,
  Download,
  RefreshCw,
  Ban,
  Check,
  ShieldCheck,
  ShieldX,
} from 'lucide-react'
import { Button } from '@/components/button'
import { formatNumber, formatPercentage } from '@/lib/utils'

// Time range data
const dataByTimeRange = {
  '1h': {
    totalRequests: 103652,
    fraudDetected: 372,
    fraudRate: 0.0036,
    blocked: 343,
    flagged: 29,
    accuracy: 0.952,
    falsePositives: 5,
    processingTime: 84,
  },
  '24h': {
    totalRequests: 2487653,
    fraudDetected: 8942,
    fraudRate: 0.0036,
    blocked: 8234,
    flagged: 708,
    accuracy: 0.952,
    falsePositives: 127,
    processingTime: 87,
  },
  '7d': {
    totalRequests: 17413571,
    fraudDetected: 62594,
    fraudRate: 0.0036,
    blocked: 57638,
    flagged: 4956,
    accuracy: 0.948,
    falsePositives: 889,
    processingTime: 89,
  },
  '30d': {
    totalRequests: 74629877,
    fraudDetected: 268667,
    fraudRate: 0.0036,
    blocked: 247178,
    flagged: 21489,
    accuracy: 0.945,
    falsePositives: 3821,
    processingTime: 91,
  },
}

const timeRangeLabels: Record<string, string> = {
  '1h': 'Last Hour',
  '24h': 'Last 24 Hours',
  '7d': 'Last 7 Days',
  '30d': 'Last 30 Days',
}

type Threat = {
  id: string;
  type: string;
  severity: 'critical' | 'high' | 'medium' | 'low';
  ip: string;
  country: string;
  timestamp: string;
  blocked: boolean;
}

const initialThreats: Threat[] = [
  {
    id: '1',
    type: 'Click Fraud',
    severity: 'high',
    ip: '192.168.1.100',
    country: 'Unknown',
    timestamp: '2 minutes ago',
    blocked: true,
  },
  {
    id: '2',
    type: 'Bot Traffic',
    severity: 'critical',
    ip: '10.0.0.45',
    country: 'Russia',
    timestamp: '5 minutes ago',
    blocked: true,
  },
  {
    id: '3',
    type: 'Invalid Traffic',
    severity: 'medium',
    ip: '172.16.0.23',
    country: 'China',
    timestamp: '8 minutes ago',
    blocked: false,
  },
  {
    id: '4',
    type: 'Click Farm',
    severity: 'high',
    ip: '203.45.67.89',
    country: 'Vietnam',
    timestamp: '12 minutes ago',
    blocked: true,
  },
  {
    id: '5',
    type: 'Suspicious Pattern',
    severity: 'medium',
    ip: '45.123.78.90',
    country: 'India',
    timestamp: '15 minutes ago',
    blocked: false,
  },
]

export default function FraudDetectionPage() {
  const [timeRange, setTimeRange] = useState('24h')
  const [threats, setThreats] = useState<Threat[]>(initialThreats)
  const [exportSuccess, setExportSuccess] = useState(false)
  const [retraining, setRetraining] = useState(false)
  const [modelStatus, setModelStatus] = useState<'Online' | 'Training' | 'Offline'>('Online')

  // Get stats based on time range
  const fraudStats = dataByTimeRange[timeRange as keyof typeof dataByTimeRange]

  // Toggle block status
  const toggleBlockStatus = (threatId: string) => {
    setThreats(threats.map(t => 
      t.id === threatId ? { ...t, blocked: !t.blocked } : t
    ))
  }

  // Block all unblocked threats
  const blockAllThreats = () => {
    setThreats(threats.map(t => ({ ...t, blocked: true })))
  }

  // Export report
  const handleExport = () => {
    const headers = ['ID', 'Type', 'Severity', 'IP Address', 'Country', 'Timestamp', 'Status']
    const rows = threats.map(t => [
      t.id,
      t.type,
      t.severity,
      t.ip,
      t.country,
      t.timestamp,
      t.blocked ? 'Blocked' : 'Flagged'
    ])

    const statsSection = [
      '',
      `Fraud Report - ${timeRangeLabels[timeRange]}`,
      '',
      'Summary Statistics',
      `Total Requests,${fraudStats.totalRequests}`,
      `Fraud Detected,${fraudStats.fraudDetected}`,
      `Fraud Rate,${(fraudStats.fraudRate * 100).toFixed(2)}%`,
      `Blocked,${fraudStats.blocked}`,
      `Flagged,${fraudStats.flagged}`,
      `ML Accuracy,${(fraudStats.accuracy * 100).toFixed(1)}%`,
      '',
      'Recent Threats',
    ]

    const csvContent = [
      ...statsSection,
      headers.join(','),
      ...rows.map(row => row.join(','))
    ].join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = `fraud_report_${timeRange}_${new Date().toISOString().split('T')[0]}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    setExportSuccess(true)
    setTimeout(() => setExportSuccess(false), 2000)
  }

  // Retrain model
  const handleRetrainModel = () => {
    setRetraining(true)
    setModelStatus('Training')
    
    // Simulate training process
    setTimeout(() => {
      setRetraining(false)
      setModelStatus('Online')
    }, 3000)
  }

  const fraudTypes = [
    { type: 'Click Fraud', count: 3456, percentage: 38.7 },
    { type: 'Bot Traffic', count: 2789, percentage: 31.2 },
    { type: 'Invalid Traffic', count: 1567, percentage: 17.5 },
    { type: 'Click Farm', count: 834, percentage: 9.3 },
    { type: 'Other', count: 296, percentage: 3.3 },
  ]

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-100 text-red-700'
      case 'high':
        return 'bg-orange-100 text-orange-700'
      case 'medium':
        return 'bg-yellow-100 text-yellow-700'
      case 'low':
        return 'bg-blue-100 text-blue-700'
      default:
        return 'bg-gray-100 text-gray-700'
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Fraud Detection</h1>
          <p className="text-muted-foreground">
            AI-powered fraud detection and prevention
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <select
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value)}
            className="h-10 rounded-lg border border-gray-300 bg-white px-4 text-sm"
          >
            <option value="1h">Last Hour</option>
            <option value="24h">Last 24 Hours</option>
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
          </select>
          <Button 
            variant="outline"
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
                Export Report
              </>
            )}
          </Button>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Total Requests
              </CardTitle>
              <Activity className="h-4 w-4 text-blue-500" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatNumber(fraudStats.totalRequests)}</div>
            <p className="text-xs text-muted-foreground">Last 24 hours</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Fraud Detected
              </CardTitle>
              <AlertTriangle className="h-4 w-4 text-red-500" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">
              {formatNumber(fraudStats.fraudDetected)}
            </div>
            <p className="text-xs text-muted-foreground">
              {formatPercentage(fraudStats.fraudRate)} fraud rate
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                ML Accuracy
              </CardTitle>
              <CheckCircle className="h-4 w-4 text-green-500" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {formatPercentage(fraudStats.accuracy)}
            </div>
            <p className="text-xs text-muted-foreground">
              {fraudStats.falsePositives} false positives
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Avg Processing
              </CardTitle>
              <Shield className="h-4 w-4 text-purple-500" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{fraudStats.processingTime}ms</div>
            <p className="text-xs text-green-600">Target: &lt;100ms</p>
          </CardContent>
        </Card>
      </div>

      {/* ML Model Status */}
      <Card className={`border-l-4 ${modelStatus === 'Online' ? 'border-l-green-500' : modelStatus === 'Training' ? 'border-l-yellow-500' : 'border-l-red-500'}`}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <div className={`flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br ${
                modelStatus === 'Online' ? 'from-green-500 to-emerald-600' : 
                modelStatus === 'Training' ? 'from-yellow-500 to-orange-500' : 
                'from-red-500 to-red-600'
              }`}>
                <Shield className="h-5 w-5 text-white" />
              </div>
              <div>
                <CardTitle>ML Model Status</CardTitle>
                <CardDescription>Random Forest Classifier</CardDescription>
              </div>
            </div>
            <Button 
              variant="outline"
              onClick={handleRetrainModel}
              disabled={retraining}
              className={retraining ? 'opacity-50' : ''}
            >
              <RefreshCw className={`mr-2 h-4 w-4 ${retraining ? 'animate-spin' : ''}`} />
              {retraining ? 'Training...' : 'Retrain Model'}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-5">
            <div>
              <div className="text-sm text-muted-foreground">Model Version</div>
              <div className="text-lg font-semibold">v2.3.1</div>
            </div>
            <div>
              <div className="text-sm text-muted-foreground">Features</div>
              <div className="text-lg font-semibold">15 inputs</div>
            </div>
            <div>
              <div className="text-sm text-muted-foreground">Accuracy</div>
              <div className="text-lg font-semibold text-green-600">{formatPercentage(fraudStats.accuracy)}</div>
            </div>
            <div>
              <div className="text-sm text-muted-foreground">Last Trained</div>
              <div className="text-lg font-semibold">{retraining ? 'Training now...' : '2 days ago'}</div>
            </div>
            <div>
              <div className="text-sm text-muted-foreground">Status</div>
              <div className="flex items-center">
                <span className={`mr-2 h-2 w-2 rounded-full ${
                  modelStatus === 'Online' ? 'bg-green-500' : 
                  modelStatus === 'Training' ? 'bg-yellow-500 animate-pulse' : 
                  'bg-red-500'
                }`} />
                <span className={`text-lg font-semibold ${
                  modelStatus === 'Online' ? 'text-green-600' : 
                  modelStatus === 'Training' ? 'text-yellow-600' : 
                  'text-red-600'
                }`}>{modelStatus}</span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Fraud Types & Recent Threats */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Fraud Distribution</CardTitle>
            <CardDescription>By type (last 24 hours)</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {fraudTypes.map((fraud, index) => (
                <div key={index} className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="font-medium">{fraud.type}</span>
                    <span className="text-muted-foreground">
                      {formatNumber(fraud.count)} ({fraud.percentage}%)
                    </span>
                  </div>
                  <div className="h-2 w-full overflow-hidden rounded-full bg-gray-200">
                    <div
                      className="h-full bg-gradient-to-r from-red-500 to-orange-500"
                      style={{ width: `${fraud.percentage}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Recent Threats</CardTitle>
                <CardDescription>Latest fraud attempts detected</CardDescription>
              </div>
              {threats.some(t => !t.blocked) && (
                <Button 
                  size="sm" 
                  onClick={blockAllThreats}
                  className="bg-red-600 hover:bg-red-700"
                >
                  <Ban className="mr-2 h-4 w-4" />
                  Block All
                </Button>
              )}
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {threats.map((threat) => (
                <div
                  key={threat.id}
                  className="flex items-start space-x-3 rounded-lg border border-gray-200 p-3"
                >
                  {threat.blocked ? (
                    <ShieldCheck className="h-5 w-5 text-red-500" />
                  ) : (
                    <AlertTriangle className="h-5 w-5 text-yellow-500" />
                  )}
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <span className="font-medium">{threat.type}</span>
                      <span
                        className={`rounded-full px-2 py-0.5 text-xs font-semibold ${getSeverityColor(
                          threat.severity
                        )}`}
                      >
                        {threat.severity}
                      </span>
                    </div>
                    <div className="mt-1 flex items-center space-x-3 text-xs text-muted-foreground">
                      <span className="flex items-center">
                        <Globe className="mr-1 h-3 w-3" />
                        {threat.ip}
                      </span>
                      <span>{threat.country}</span>
                      <span>{threat.timestamp}</span>
                    </div>
                    <div className="mt-2 flex items-center justify-between">
                      {threat.blocked ? (
                        <span className="text-xs font-semibold text-red-600">
                          ✓ Blocked
                        </span>
                      ) : (
                        <span className="text-xs font-semibold text-yellow-600">
                          ⚠ Flagged for Review
                        </span>
                      )}
                      <button
                        onClick={() => toggleBlockStatus(threat.id)}
                        className={`text-xs px-2 py-1 rounded ${
                          threat.blocked 
                            ? 'bg-green-100 text-green-700 hover:bg-green-200' 
                            : 'bg-red-100 text-red-700 hover:bg-red-200'
                        }`}
                      >
                        {threat.blocked ? 'Unblock' : 'Block Now'}
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Actions Taken */}
      <Card>
        <CardHeader>
          <CardTitle>Actions Summary</CardTitle>
          <CardDescription>Fraud prevention measures (last 24 hours)</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <div className="rounded-lg bg-red-50 p-4">
              <div className="flex items-center space-x-2">
                <XCircle className="h-5 w-5 text-red-600" />
                <span className="text-sm font-medium text-red-900">Blocked</span>
              </div>
              <div className="mt-2 text-3xl font-bold text-red-600">
                {formatNumber(fraudStats.blocked)}
              </div>
              <p className="mt-1 text-xs text-red-700">
                {formatPercentage(fraudStats.blocked / fraudStats.fraudDetected)} of detected
              </p>
            </div>

            <div className="rounded-lg bg-yellow-50 p-4">
              <div className="flex items-center space-x-2">
                <AlertTriangle className="h-5 w-5 text-yellow-600" />
                <span className="text-sm font-medium text-yellow-900">Flagged</span>
              </div>
              <div className="mt-2 text-3xl font-bold text-yellow-600">
                {formatNumber(fraudStats.flagged)}
              </div>
              <p className="mt-1 text-xs text-yellow-700">
                Pending manual review
              </p>
            </div>

            <div className="rounded-lg bg-green-50 p-4">
              <div className="flex items-center space-x-2">
                <CheckCircle className="h-5 w-5 text-green-600" />
                <span className="text-sm font-medium text-green-900">Allowed</span>
              </div>
              <div className="mt-2 text-3xl font-bold text-green-600">
                {formatNumber(fraudStats.totalRequests - fraudStats.fraudDetected)}
              </div>
              <p className="mt-1 text-xs text-green-700">
                {formatPercentage(1 - fraudStats.fraudRate)} legitimate
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
