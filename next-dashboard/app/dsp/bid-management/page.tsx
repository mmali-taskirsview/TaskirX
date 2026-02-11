'use client'

import { useState, useEffect } from 'react'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { 
  Plus, 
  Edit, 
  Trash2, 
  Target,
  TrendingUp,
  DollarSign,
  Clock,
  Zap,
  RefreshCw,
  Settings
} from 'lucide-react'

interface BidStrategy {
  id: string
  name: string
  status: string
  type: string
  baseBid: number
  maxBid: number
  minBid: number
  targetCpa: number
  targetRoas: number
  frequencyCap: {
    impressions: number
    period: 'hour' | 'day' | 'week' | 'month'
    perUser: boolean
  }
  pacing: {
    type: 'even' | 'accelerated' | 'front_loaded'
    dailyBudget?: number
    hourlyBudget?: number
  }
  bidAdjustments: {
    device?: Record<string, number>
    geo?: Record<string, number>
    daypart?: Record<string, number>
  }
  totalImpressions: number
  totalSpend: number
  avgCpm: number
  createdAt: string
}

export default function BidManagementPage() {
  const [strategies, setStrategies] = useState<BidStrategy[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingStrategy, setEditingStrategy] = useState<BidStrategy | null>(null)
  const [formData, setFormData] = useState<Partial<BidStrategy>>({
    name: '',
    type: 'manual',
    baseBid: 2.0,
    maxBid: 10.0,
    minBid: 0.5,
    targetCpa: 25,
    targetRoas: 4.0,
    frequencyCap: {
        impressions: 0,
        period: 'day',
        perUser: true
    },
    pacing: {
        type: 'even',
        dailyBudget: 1000
    }
  })

  const fetchStrategies = async () => {
    try {
      setLoading(true)
      const response = await api.getBidStrategies()
      setStrategies(response.data)
      setError(null)
    } catch (err: any) {
      setError(err.message || 'Failed to load bid strategies')
      // Demo data
      setStrategies([
        {
          id: '1',
          name: 'Aggressive CPA Targeting',
          status: 'active',
          type: 'target_cpa',
          baseBid: 3.5,
          maxBid: 15.0,
          minBid: 1.0,
          targetCpa: 25,
          targetRoas: 0,
          frequencyCap: { impressions: 3, period: 'day', perUser: true },
          pacing: { type: 'accelerated', dailyBudget: 5000 },
          bidAdjustments: {
            device: { mobile: 1.2, desktop: 1.0, tablet: 0.8 },
            geo: { 'US': 1.3, 'UK': 1.1, 'CA': 1.0 }
          },
          totalImpressions: 2847583,
          totalSpend: 28475.83,
          avgCpm: 10.0,
          createdAt: '2024-01-15'
        },
        {
          id: '2',
          name: 'ROAS Optimizer',
          status: 'active',
          type: 'target_roas',
          baseBid: 2.5,
          maxBid: 12.0,
          minBid: 0.5,
          targetCpa: 0,
          targetRoas: 4.0,
          frequencyCap: { impressions: 5, period: 'day', perUser: true },
          pacing: { type: 'even', dailyBudget: 3000 },
          bidAdjustments: {
            device: { mobile: 1.0, desktop: 1.1, tablet: 0.9 }
          },
          totalImpressions: 1847293,
          totalSpend: 18472.93,
          avgCpm: 10.0,
          createdAt: '2024-01-20'
        },
        {
          id: '3',
          name: 'Max Conversions',
          status: 'active',
          type: 'maximize_conversions',
          baseBid: 4.0,
          maxBid: 20.0,
          minBid: 1.5,
          targetCpa: 0,
          targetRoas: 0,
          frequencyCap: { impressions: 10, period: 'week', perUser: true },
          pacing: { type: 'accelerated', dailyBudget: 10000 },
          bidAdjustments: {},
          totalImpressions: 3584729,
          totalSpend: 35847.29,
          avgCpm: 10.0,
          createdAt: '2024-02-01'
        },
        {
          id: '4',
          name: 'Manual Control',
          status: 'paused',
          type: 'manual',
          baseBid: 2.0,
          maxBid: 8.0,
          minBid: 0.5,
          targetCpa: 0,
          targetRoas: 0,
          frequencyCap: { impressions: 7, period: 'day', perUser: true },
          pacing: { type: 'even', dailyBudget: 2000 },
          bidAdjustments: {
            device: { mobile: 1.1, desktop: 1.0, tablet: 0.7 },
            daypart: { '0-6': 0.5, '6-12': 0.8, '12-18': 1.2, '18-24': 1.0 }
          },
          totalImpressions: 584729,
          totalSpend: 5847.29,
          avgCpm: 10.0,
          createdAt: '2024-02-10'
        }
      ])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchStrategies()
  }, [])

  const handleSubmit = async () => {
    try {
      if (editingStrategy) {
        await api.updateBidStrategy(editingStrategy.id, formData)
      } else {
        await api.createBidStrategy(formData)
      }
      setDialogOpen(false)
      setEditingStrategy(null)
      resetForm()
      fetchStrategies()
    } catch (err: any) {
      alert(err.message || 'Failed to save bid strategy')
    }
  }

  const handleEdit = (strategy: BidStrategy) => {
    setEditingStrategy(strategy)
    setFormData({
      name: strategy.name,
      type: strategy.type,
      baseBid: strategy.baseBid,
      maxBid: strategy.maxBid,
      minBid: strategy.minBid,
      targetCpa: strategy.targetCpa,
      targetRoas: strategy.targetRoas,
      frequencyCap: strategy.frequencyCap,
      pacing: strategy.pacing
    })
    setDialogOpen(true)
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this bid strategy?')) return
    try {
      await api.deleteBidStrategy(id)
      fetchStrategies()
    } catch (err: any) {
      alert(err.message || 'Failed to delete bid strategy')
    }
  }

  const resetForm = () => {
    setFormData({
      name: '',
      type: 'manual',
      baseBid: 2.0,
      maxBid: 10.0,
      minBid: 0.5,
      targetCpa: 25,
      targetRoas: 4.0,
      frequencyCap: { impressions: 5, period: 'day', perUser: true },
      pacing: { type: 'even', dailyBudget: 1000 }
    })
  }

  const formatCurrency = (num: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(num)
  }

  const formatNumber = (num: number) => {
    if (num >= 1000000) return (num / 1000000).toFixed(2) + 'M'
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
    return num.toLocaleString()
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <Badge className="bg-green-500">Active</Badge>
      case 'paused':
        return <Badge variant="secondary">Paused</Badge>
      default:
        return <Badge variant="destructive">Inactive</Badge>
    }
  }

  const getTypeBadge = (type: string) => {
    const types: Record<string, { label: string; color: string }> = {
      manual: { label: 'Manual', color: 'text-gray-500 border-gray-500' },
      auto_optimize: { label: 'Auto Optimize', color: 'text-blue-500 border-blue-500' },
      target_cpa: { label: 'Target CPA', color: 'text-green-500 border-green-500' },
      target_roas: { label: 'Target ROAS', color: 'text-purple-500 border-purple-500' },
      maximize_conversions: { label: 'Max Conversions', color: 'text-orange-500 border-orange-500' }
    }
    const config = types[type] || { label: type, color: '' }
    return <Badge variant="outline" className={config.color}>{config.label}</Badge>
  }

  const getPacingBadge = (pacing: { type: string }) => {
    if (!pacing) return <Badge variant="outline">Default</Badge>
    return pacing.type === 'accelerated' 
      ? <Badge variant="outline" className="text-red-500 border-red-500">Accelerated</Badge>
      : <Badge variant="outline" className="text-blue-500 border-blue-500">Even</Badge>
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Bid Management</h1>
          <p className="text-muted-foreground">
            Configure bidding strategies and optimization algorithms
          </p>
        </div>
        <div className="flex gap-2">
          <Button onClick={fetchStrategies} variant="outline">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button onClick={() => { setEditingStrategy(null); resetForm(); }}>
                <Plus className="h-4 w-4 mr-2" />
                New Strategy
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle>{editingStrategy ? 'Edit' : 'Create'} Bid Strategy</DialogTitle>
                <DialogDescription>
                  Configure bidding parameters and optimization goals
                </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4 max-h-[70vh] overflow-y-auto">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Strategy Name</Label>
                    <Input
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      placeholder="e.g., Aggressive CPA"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Strategy Type</Label>
                    <Select
                      value={formData.type}
                      onValueChange={(value) => setFormData({ ...formData, type: value })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="manual">Manual Bidding</SelectItem>
                        <SelectItem value="auto_optimize">Auto Optimize</SelectItem>
                        <SelectItem value="target_cpa">Target CPA</SelectItem>
                        <SelectItem value="target_roas">Target ROAS</SelectItem>
                        <SelectItem value="maximize_conversions">Maximize Conversions</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-4">
                  <div className="space-y-2">
                    <Label>Base Bid ($)</Label>
                    <Input
                      type="number"
                      step="0.01"
                      value={formData.baseBid}
                      onChange={(e) => setFormData({ ...formData, baseBid: parseFloat(e.target.value) })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Min Bid ($)</Label>
                    <Input
                      type="number"
                      step="0.01"
                      value={formData.minBid}
                      onChange={(e) => setFormData({ ...formData, minBid: parseFloat(e.target.value) })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Max Bid ($)</Label>
                    <Input
                      type="number"
                      step="0.01"
                      value={formData.maxBid}
                      onChange={(e) => setFormData({ ...formData, maxBid: parseFloat(e.target.value) })}
                    />
                  </div>
                </div>

                {formData.type === 'target_cpa' && (
                  <div className="space-y-2">
                    <Label>Target CPA ($)</Label>
                    <Input
                      type="number"
                      step="0.01"
                      value={formData.targetCpa}
                      onChange={(e) => setFormData({ ...formData, targetCpa: parseFloat(e.target.value) })}
                    />
                  </div>
                )}

                {formData.type === 'target_roas' && (
                  <div className="space-y-2">
                    <Label>Target ROAS</Label>
                    <Input
                      type="number"
                      step="0.1"
                      value={formData.targetRoas}
                      onChange={(e) => setFormData({ ...formData, targetRoas: parseFloat(e.target.value) })}
                    />
                  </div>
                )}

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Max Impressions</Label>
                    <Input
                      type="number"
                      value={formData.frequencyCap?.impressions || 0}
                      onChange={(e) => setFormData({ 
                        ...formData, 
                        frequencyCap: { ...(formData.frequencyCap as any), impressions: parseInt(e.target.value) } 
                      })}
                    />
                    <p className="text-xs text-muted-foreground">Per user limit</p>
                  </div>
                  <div className="space-y-2">
                    <Label>Period</Label>
                    <Select
                      value={formData.frequencyCap?.period || 'day'}
                      onValueChange={(value: any) => setFormData({ 
                        ...formData, 
                        frequencyCap: { ...(formData.frequencyCap as any), period: value }
                      })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="hour">Hour</SelectItem>
                        <SelectItem value="day">Day</SelectItem>
                        <SelectItem value="week">Week</SelectItem>
                        <SelectItem value="month">Month</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                   <div className="space-y-2">
                      <Label>Pacing Type</Label>
                      <Select
                        value={formData.pacing?.type || 'even'}
                        onValueChange={(value: any) => setFormData({ 
                            ...formData, 
                            pacing: { ...(formData.pacing as any), type: value } 
                        })}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="even">Even (Spread budget)</SelectItem>
                          <SelectItem value="accelerated">Accelerated (ASAP)</SelectItem>
                          <SelectItem value="front_loaded">Front Loaded</SelectItem>
                        </SelectContent>
                      </Select>
                   </div>
                   <div className="space-y-2">
                      <Label>Daily Budget ($)</Label>
                      <Input
                        type="number"
                        value={formData.pacing?.dailyBudget || 0}
                        onChange={(e) => setFormData({ 
                            ...formData, 
                            pacing: { ...(formData.pacing as any), dailyBudget: parseFloat(e.target.value) } 
                        })}
                      />
                   </div>
                </div>

                <Button onClick={handleSubmit} className="w-full">
                  {editingStrategy ? 'Update' : 'Create'} Strategy
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {error && (
        <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
          <p className="text-yellow-800 dark:text-yellow-200 text-sm">
            Using demo data - {error}
          </p>
        </div>
      )}

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Target className="h-4 w-4 text-purple-500" />
              Active Strategies
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {strategies.filter(s => s.status === 'active').length}
            </div>
            <p className="text-xs text-muted-foreground">
              of {strategies.length} total
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <DollarSign className="h-4 w-4 text-green-500" />
              Total Spend
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(strategies.reduce((sum, s) => sum + s.totalSpend, 0))}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Zap className="h-4 w-4 text-blue-500" />
              Total Impressions
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatNumber(strategies.reduce((sum, s) => sum + s.totalImpressions, 0))}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <TrendingUp className="h-4 w-4 text-orange-500" />
              Avg CPM
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(
                strategies.reduce((sum, s) => sum + s.totalSpend, 0) / 
                strategies.reduce((sum, s) => sum + s.totalImpressions, 0) * 1000 || 0
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Strategies Table */}
      <Card>
        <CardHeader>
          <CardTitle>Bid Strategies</CardTitle>
          <CardDescription>Manage bidding algorithms and pacing rules</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Strategy</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Bid Range</TableHead>
                <TableHead>Freq Cap</TableHead>
                <TableHead>Pacing</TableHead>
                <TableHead>Impressions</TableHead>
                <TableHead>Spend</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {strategies.map((strategy) => (
                <TableRow key={strategy.id}>
                  <TableCell>
                    <div className="font-medium">{strategy.name}</div>
                  </TableCell>
                  <TableCell>{getStatusBadge(strategy.status)}</TableCell>
                  <TableCell>{getTypeBadge(strategy.type)}</TableCell>
                  <TableCell>
                    <span className="text-sm">
                      {formatCurrency(strategy.minBid)} - {formatCurrency(strategy.maxBid)}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm">
                      {strategy.frequencyCap?.impressions || 0} / {strategy.frequencyCap?.period || 'day'}
                    </span>
                  </TableCell>
                  <TableCell>{getPacingBadge(strategy.pacing)}</TableCell>
                  <TableCell>{formatNumber(strategy.totalImpressions)}</TableCell>
                  <TableCell>{formatCurrency(strategy.totalSpend)}</TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleEdit(strategy)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleDelete(strategy.id)}
                      >
                        <Trash2 className="h-4 w-4 text-red-500" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
