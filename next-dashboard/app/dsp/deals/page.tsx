'use client'

import { useState, useEffect } from 'react'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
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
  FileText,
  DollarSign,
  Target,
  Calendar,
  RefreshCw,
  Building,
  TrendingUp
} from 'lucide-react'

interface Deal {
  id: string
  name: string
  dealId: string
  status: string
  type: string
  supplyPartnerId: string
  supplyPartnerName: string
  floorPrice: number
  fixedPrice: number
  budget: number
  budgetSpent: number
  impressionsGoal: number
  impressionsDelivered: number
  inventory: {
    formats?: string[]
    sizes?: string[]
    geos?: string[]
  }
  startDate: string
  endDate: string
  createdAt: string
}

export default function DealsPage() {
  const [deals, setDeals] = useState<Deal[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingDeal, setEditingDeal] = useState<Deal | null>(null)
  const [formData, setFormData] = useState({
    name: '',
    dealId: '',
    type: 'preferred',
    supplyPartnerId: '',
    floorPrice: 2.0,
    fixedPrice: 0,
    budget: 10000,
    impressionsGoal: 1000000,
    startDate: '',
    endDate: ''
  })

  const fetchDeals = async () => {
    try {
      setLoading(true)
      const response = await api.getDeals()
      setDeals(response.data)
      setError(null)
    } catch (err: any) {
      setError(err.message || 'Failed to load deals')
      // Demo data
      setDeals([
        {
          id: '1',
          name: 'Premium Sports Inventory',
          dealId: 'DEAL-ESPN-2024-001',
          status: 'active',
          type: 'preferred',
          supplyPartnerId: '1',
          supplyPartnerName: 'ESPN Digital',
          floorPrice: 8.5,
          fixedPrice: 0,
          budget: 50000,
          budgetSpent: 32847,
          impressionsGoal: 5000000,
          impressionsDelivered: 3284729,
          inventory: { formats: ['video', 'display'], sizes: ['300x250', '728x90'], geos: ['US'] },
          startDate: '2024-01-01',
          endDate: '2024-03-31',
          createdAt: '2024-01-01'
        },
        {
          id: '2',
          name: 'News PMP - Breaking News',
          dealId: 'DEAL-CNN-PMP-2024',
          status: 'active',
          type: 'private_auction',
          supplyPartnerId: '2',
          supplyPartnerName: 'CNN Digital',
          floorPrice: 12.0,
          fixedPrice: 0,
          budget: 75000,
          budgetSpent: 45283,
          impressionsGoal: 6000000,
          impressionsDelivered: 3583928,
          inventory: { formats: ['display', 'native'], sizes: ['300x250', '300x600'], geos: ['US', 'UK'] },
          startDate: '2024-01-15',
          endDate: '2024-04-15',
          createdAt: '2024-01-15'
        },
        {
          id: '3',
          name: 'Finance Guaranteed - Q1',
          dealId: 'DEAL-BLOOM-PG-Q1',
          status: 'active',
          type: 'programmatic_guaranteed',
          supplyPartnerId: '3',
          supplyPartnerName: 'Bloomberg',
          floorPrice: 0,
          fixedPrice: 25.0,
          budget: 125000,
          budgetSpent: 87500,
          impressionsGoal: 5000000,
          impressionsDelivered: 3500000,
          inventory: { formats: ['display', 'video'], sizes: ['970x250', '300x250'], geos: ['US', 'UK', 'SG'] },
          startDate: '2024-01-01',
          endDate: '2024-03-31',
          createdAt: '2024-01-01'
        },
        {
          id: '4',
          name: 'Entertainment Video',
          dealId: 'DEAL-HULU-VID-2024',
          status: 'pending',
          type: 'preferred',
          supplyPartnerId: '4',
          supplyPartnerName: 'Hulu',
          floorPrice: 18.0,
          fixedPrice: 0,
          budget: 100000,
          budgetSpent: 0,
          impressionsGoal: 4000000,
          impressionsDelivered: 0,
          inventory: { formats: ['video'], sizes: ['pre-roll', 'mid-roll'], geos: ['US'] },
          startDate: '2024-02-01',
          endDate: '2024-05-31',
          createdAt: '2024-01-20'
        },
        {
          id: '5',
          name: 'Tech News Open Auction',
          dealId: 'DEAL-TC-OPEN-2024',
          status: 'active',
          type: 'open_auction',
          supplyPartnerId: '5',
          supplyPartnerName: 'TechCrunch',
          floorPrice: 5.0,
          fixedPrice: 0,
          budget: 25000,
          budgetSpent: 18472,
          impressionsGoal: 3000000,
          impressionsDelivered: 2847293,
          inventory: { formats: ['display', 'native'], geos: ['US', 'CA', 'UK', 'DE'] },
          startDate: '2024-01-01',
          endDate: '2024-06-30',
          createdAt: '2024-01-01'
        }
      ])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchDeals()
  }, [])

  const handleSubmit = async () => {
    try {
      if (editingDeal) {
        await api.updateDeal(editingDeal.id, formData)
      } else {
        await api.createDeal(formData)
      }
      setDialogOpen(false)
      setEditingDeal(null)
      resetForm()
      fetchDeals()
    } catch (err: any) {
      alert(err.message || 'Failed to save deal')
    }
  }

  const handleEdit = (deal: Deal) => {
    setEditingDeal(deal)
    setFormData({
      name: deal.name,
      dealId: deal.dealId,
      type: deal.type,
      supplyPartnerId: deal.supplyPartnerId,
      floorPrice: deal.floorPrice,
      fixedPrice: deal.fixedPrice,
      budget: deal.budget,
      impressionsGoal: deal.impressionsGoal,
      startDate: deal.startDate,
      endDate: deal.endDate
    })
    setDialogOpen(true)
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this deal?')) return
    try {
      await api.deleteDeal(id)
      fetchDeals()
    } catch (err: any) {
      alert(err.message || 'Failed to delete deal')
    }
  }

  const resetForm = () => {
    setFormData({
      name: '',
      dealId: '',
      type: 'preferred',
      supplyPartnerId: '',
      floorPrice: 2.0,
      fixedPrice: 0,
      budget: 10000,
      impressionsGoal: 1000000,
      startDate: '',
      endDate: ''
    })
  }

  const formatNumber = (num: number) => {
    if (num >= 1000000) return (num / 1000000).toFixed(2) + 'M'
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
    return num.toLocaleString()
  }

  const formatCurrency = (num: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(num)
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <Badge className="bg-green-500">Active</Badge>
      case 'pending':
        return <Badge variant="outline" className="text-yellow-500 border-yellow-500">Pending</Badge>
      case 'completed':
        return <Badge variant="secondary">Completed</Badge>
      case 'paused':
        return <Badge variant="secondary">Paused</Badge>
      default:
        return <Badge variant="destructive">Inactive</Badge>
    }
  }

  const getTypeBadge = (type: string) => {
    const types: Record<string, { label: string; color: string }> = {
      preferred: { label: 'Preferred', color: 'text-blue-500 border-blue-500' },
      private_auction: { label: 'Private Auction', color: 'text-purple-500 border-purple-500' },
      programmatic_guaranteed: { label: 'PG', color: 'text-green-500 border-green-500' },
      open_auction: { label: 'Open', color: 'text-gray-500 border-gray-500' }
    }
    const config = types[type] || { label: type, color: '' }
    return <Badge variant="outline" className={config.color}>{config.label}</Badge>
  }

  const getDeliveryProgress = (delivered: number, goal: number) => {
    const progress = (delivered / goal) * 100
    return (
      <div className="w-full">
        <div className="flex justify-between text-xs mb-1">
          <span>{formatNumber(delivered)}</span>
          <span className={progress >= 80 ? 'text-green-500' : progress >= 50 ? 'text-yellow-500' : 'text-red-500'}>
            {progress.toFixed(1)}%
          </span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-2">
          <div 
            className={`h-2 rounded-full ${progress >= 80 ? 'bg-green-500' : progress >= 50 ? 'bg-yellow-500' : 'bg-red-500'}`}
            style={{ width: `${Math.min(progress, 100)}%` }}
          />
        </div>
      </div>
    )
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
          <h1 className="text-3xl font-bold">Deals</h1>
          <p className="text-muted-foreground">
            Manage PMP, Preferred, and Programmatic Guaranteed deals
          </p>
        </div>
        <div className="flex gap-2">
          <Button onClick={fetchDeals} variant="outline">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button onClick={() => { setEditingDeal(null); resetForm(); }}>
                <Plus className="h-4 w-4 mr-2" />
                New Deal
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle>{editingDeal ? 'Edit' : 'Create'} Deal</DialogTitle>
                <DialogDescription>
                  Configure deal terms and inventory access
                </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4 max-h-[70vh] overflow-y-auto">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Deal Name</Label>
                    <Input
                      value={formData.name}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, name: e.target.value })}
                      placeholder="e.g., Premium Sports Inventory"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Deal ID</Label>
                    <Input
                      value={formData.dealId}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, dealId: e.target.value })}
                      placeholder="e.g., DEAL-ESPN-2024-001"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Deal Type</Label>
                    <Select
                      value={formData.type}
                      onValueChange={(value: string) => setFormData({ ...formData, type: value })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="preferred">Preferred Deal</SelectItem>
                        <SelectItem value="private_auction">Private Auction (PMP)</SelectItem>
                        <SelectItem value="programmatic_guaranteed">Programmatic Guaranteed</SelectItem>
                        <SelectItem value="open_auction">Open Auction</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label>Supply Partner ID</Label>
                    <Input
                      value={formData.supplyPartnerId}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, supplyPartnerId: e.target.value })}
                      placeholder="Partner ID"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Floor Price ($)</Label>
                    <Input
                      type="number"
                      step="0.01"
                      value={formData.floorPrice}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, floorPrice: parseFloat(e.target.value) })}
                    />
                    <p className="text-xs text-muted-foreground">For auctions</p>
                  </div>
                  <div className="space-y-2">
                    <Label>Fixed Price ($)</Label>
                    <Input
                      type="number"
                      step="0.01"
                      value={formData.fixedPrice}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, fixedPrice: parseFloat(e.target.value) })}
                    />
                    <p className="text-xs text-muted-foreground">For PG deals</p>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Budget ($)</Label>
                    <Input
                      type="number"
                      value={formData.budget}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, budget: parseFloat(e.target.value) })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Impressions Goal</Label>
                    <Input
                      type="number"
                      value={formData.impressionsGoal}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, impressionsGoal: parseInt(e.target.value) })}
                    />
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Start Date</Label>
                    <Input
                      type="date"
                      value={formData.startDate}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, startDate: e.target.value })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>End Date</Label>
                    <Input
                      type="date"
                      value={formData.endDate}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, endDate: e.target.value })}
                    />
                  </div>
                </div>

                <Button onClick={handleSubmit} className="w-full">
                  {editingDeal ? 'Update' : 'Create'} Deal
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
              <FileText className="h-4 w-4 text-blue-500" />
              Total Deals
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{deals.length}</div>
            <p className="text-xs text-muted-foreground">
              {deals.filter(d => d.status === 'active').length} active
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <DollarSign className="h-4 w-4 text-green-500" />
              Total Budget
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(deals.reduce((sum, d) => sum + d.budget, 0))}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <TrendingUp className="h-4 w-4 text-purple-500" />
              Budget Spent
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(deals.reduce((sum, d) => sum + d.budgetSpent, 0))}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Target className="h-4 w-4 text-orange-500" />
              Impressions Delivered
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatNumber(deals.reduce((sum, d) => sum + d.impressionsDelivered, 0))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Deals Table */}
      <Card>
        <CardHeader>
          <CardTitle>Active Deals</CardTitle>
          <CardDescription>All deals and their delivery status</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Deal</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Partner</TableHead>
                <TableHead>Price</TableHead>
                <TableHead>Budget</TableHead>
                <TableHead className="w-[150px]">Delivery</TableHead>
                <TableHead>Dates</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {deals.map((deal) => (
                <TableRow key={deal.id}>
                  <TableCell>
                    <div>
                      <div className="font-medium">{deal.name}</div>
                      <div className="text-xs text-muted-foreground font-mono">
                        {deal.dealId}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>{getStatusBadge(deal.status)}</TableCell>
                  <TableCell>{getTypeBadge(deal.type)}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <Building className="h-3 w-3" />
                      {deal.supplyPartnerName}
                    </div>
                  </TableCell>
                  <TableCell>
                    {deal.fixedPrice > 0 
                      ? <span className="text-green-500">{formatCurrency(deal.fixedPrice)} fixed</span>
                      : <span>{formatCurrency(deal.floorPrice)} floor</span>
                    }
                  </TableCell>
                  <TableCell>
                    <div className="text-sm">
                      <div>{formatCurrency(deal.budgetSpent)}</div>
                      <div className="text-xs text-muted-foreground">of {formatCurrency(deal.budget)}</div>
                    </div>
                  </TableCell>
                  <TableCell>
                    {getDeliveryProgress(deal.impressionsDelivered, deal.impressionsGoal)}
                  </TableCell>
                  <TableCell>
                    <div className="text-xs">
                      <div>{deal.startDate}</div>
                      <div className="text-muted-foreground">to {deal.endDate}</div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleEdit(deal)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleDelete(deal.id)}
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
