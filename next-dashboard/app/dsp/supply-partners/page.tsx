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
  Activity, 
  Zap, 
  DollarSign,
  Clock,
  TrendingUp,
  RefreshCw
} from 'lucide-react'

interface SupplyPartner {
  id: string
  name: string
  status: string
  type: string
  endpointUrl: string
  qpsLimit: number
  timeoutMs: number
  minBid: number
  maxBid: number
  dailyBudget: number
  totalRequests: number
  totalBids: number
  totalWins: number
  totalSpend: number
  createdAt: string
}

export default function SupplyPartnersPage() {
  const [partners, setPartners] = useState<SupplyPartner[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingPartner, setEditingPartner] = useState<SupplyPartner | null>(null)
  const [formData, setFormData] = useState({
    name: '',
    type: 'rtb',
    endpointUrl: '',
    qpsLimit: 10000,
    timeoutMs: 100,
    minBid: 0.01,
    maxBid: 50,
    dailyBudget: 10000
  })

  const fetchPartners = async () => {
    try {
      setLoading(true)
      const response = await api.getSupplyPartners()
      setPartners(response.data)
      setError(null)
    } catch (err: any) {
      setError(err.message || 'Failed to load supply partners')
      // Demo data
      setPartners([
        {
          id: '1',
          name: 'Google Ad Exchange',
          status: 'active',
          type: 'rtb',
          endpointUrl: 'https://adx.google.com/openrtb',
          qpsLimit: 50000,
          timeoutMs: 100,
          minBid: 0.01,
          maxBid: 100,
          dailyBudget: 50000,
          totalRequests: 15847293,
          totalBids: 14567382,
          totalWins: 2847584,
          totalSpend: 85423.47,
          createdAt: '2024-01-15'
        },
        {
          id: '2',
          name: 'AppNexus',
          status: 'active',
          type: 'rtb',
          endpointUrl: 'https://ib.adnxs.com/openrtb2',
          qpsLimit: 30000,
          timeoutMs: 80,
          minBid: 0.02,
          maxBid: 75,
          dailyBudget: 30000,
          totalRequests: 8472934,
          totalBids: 7832847,
          totalWins: 1423847,
          totalSpend: 42847.29,
          createdAt: '2024-01-20'
        },
        {
          id: '3',
          name: 'Rubicon Project',
          status: 'active',
          type: 'rtb',
          endpointUrl: 'https://optimized-by.rubiconproject.com/openrtb',
          qpsLimit: 25000,
          timeoutMs: 90,
          minBid: 0.01,
          maxBid: 60,
          dailyBudget: 25000,
          totalRequests: 5847293,
          totalBids: 5234847,
          totalWins: 987432,
          totalSpend: 28473.18,
          createdAt: '2024-02-01'
        },
        {
          id: '4',
          name: 'PubMatic',
          status: 'paused',
          type: 'header_bidding',
          endpointUrl: 'https://hbopenbid.pubmatic.com/openrtb',
          qpsLimit: 20000,
          timeoutMs: 100,
          minBid: 0.02,
          maxBid: 50,
          dailyBudget: 15000,
          totalRequests: 3847293,
          totalBids: 3234847,
          totalWins: 534728,
          totalSpend: 15847.93,
          createdAt: '2024-02-10'
        }
      ])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchPartners()
  }, [])

  const handleSubmit = async () => {
    try {
      if (editingPartner) {
        await api.updateSupplyPartner(editingPartner.id, formData)
      } else {
        await api.createSupplyPartner(formData)
      }
      setDialogOpen(false)
      setEditingPartner(null)
      resetForm()
      fetchPartners()
    } catch (err: any) {
      alert(err.message || 'Failed to save supply partner')
    }
  }

  const handleEdit = (partner: SupplyPartner) => {
    setEditingPartner(partner)
    setFormData({
      name: partner.name,
      type: partner.type,
      endpointUrl: partner.endpointUrl,
      qpsLimit: partner.qpsLimit,
      timeoutMs: partner.timeoutMs,
      minBid: partner.minBid,
      maxBid: partner.maxBid,
      dailyBudget: partner.dailyBudget
    })
    setDialogOpen(true)
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this supply partner?')) return
    try {
      await api.deleteSupplyPartner(id)
      fetchPartners()
    } catch (err: any) {
      alert(err.message || 'Failed to delete supply partner')
    }
  }

  const resetForm = () => {
    setFormData({
      name: '',
      type: 'rtb',
      endpointUrl: '',
      qpsLimit: 10000,
      timeoutMs: 100,
      minBid: 0.01,
      maxBid: 50,
      dailyBudget: 10000
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
      case 'paused':
        return <Badge variant="secondary">Paused</Badge>
      case 'pending':
        return <Badge variant="outline">Pending</Badge>
      default:
        return <Badge variant="destructive">Inactive</Badge>
    }
  }

  const getTypeBadge = (type: string) => {
    switch (type) {
      case 'rtb':
        return <Badge variant="outline" className="text-blue-500 border-blue-500">RTB</Badge>
      case 'header_bidding':
        return <Badge variant="outline" className="text-purple-500 border-purple-500">Header Bidding</Badge>
      case 'direct':
        return <Badge variant="outline" className="text-green-500 border-green-500">Direct</Badge>
      default:
        return <Badge variant="outline">{type}</Badge>
    }
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
          <h1 className="text-3xl font-bold">Supply Partners</h1>
          <p className="text-muted-foreground">
            Manage SSP connections and RTB endpoints
          </p>
        </div>
        <div className="flex gap-2">
          <Button onClick={fetchPartners} variant="outline">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button onClick={() => { setEditingPartner(null); resetForm(); }}>
                <Plus className="h-4 w-4 mr-2" />
                Add Partner
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle>{editingPartner ? 'Edit' : 'Add'} Supply Partner</DialogTitle>
                <DialogDescription>
                  Configure SSP connection and bidding parameters
                </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Partner Name</Label>
                    <Input
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      placeholder="e.g., Google Ad Exchange"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Type</Label>
                    <Select
                      value={formData.type}
                      onValueChange={(value) => setFormData({ ...formData, type: value })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="rtb">RTB (OpenRTB)</SelectItem>
                        <SelectItem value="header_bidding">Header Bidding</SelectItem>
                        <SelectItem value="direct">Direct</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="space-y-2">
                  <Label>Endpoint URL</Label>
                  <Input
                    value={formData.endpointUrl}
                    onChange={(e) => setFormData({ ...formData, endpointUrl: e.target.value })}
                    placeholder="https://ssp.example.com/openrtb"
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>QPS Limit</Label>
                    <Input
                      type="number"
                      value={formData.qpsLimit}
                      onChange={(e) => setFormData({ ...formData, qpsLimit: parseInt(e.target.value) })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Timeout (ms)</Label>
                    <Input
                      type="number"
                      value={formData.timeoutMs}
                      onChange={(e) => setFormData({ ...formData, timeoutMs: parseInt(e.target.value) })}
                    />
                  </div>
                </div>
                <div className="grid grid-cols-3 gap-4">
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
                  <div className="space-y-2">
                    <Label>Daily Budget ($)</Label>
                    <Input
                      type="number"
                      value={formData.dailyBudget}
                      onChange={(e) => setFormData({ ...formData, dailyBudget: parseFloat(e.target.value) })}
                    />
                  </div>
                </div>
                <Button onClick={handleSubmit} className="w-full">
                  {editingPartner ? 'Update' : 'Create'} Supply Partner
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
            <CardTitle className="text-sm font-medium">Total Partners</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{partners.length}</div>
            <p className="text-xs text-muted-foreground">
              {partners.filter(p => p.status === 'active').length} active
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatNumber(partners.reduce((sum, p) => sum + p.totalRequests, 0))}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Wins</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatNumber(partners.reduce((sum, p) => sum + p.totalWins, 0))}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Spend</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(partners.reduce((sum, p) => sum + p.totalSpend, 0))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Partners Table */}
      <Card>
        <CardHeader>
          <CardTitle>Supply Partners</CardTitle>
          <CardDescription>SSP connections and their performance metrics</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Partner</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>QPS</TableHead>
                <TableHead>Requests</TableHead>
                <TableHead>Wins</TableHead>
                <TableHead>Win Rate</TableHead>
                <TableHead>Spend</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {partners.map((partner) => (
                <TableRow key={partner.id}>
                  <TableCell>
                    <div>
                      <div className="font-medium">{partner.name}</div>
                      <div className="text-xs text-muted-foreground truncate max-w-[200px]">
                        {partner.endpointUrl}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>{getStatusBadge(partner.status)}</TableCell>
                  <TableCell>{getTypeBadge(partner.type)}</TableCell>
                  <TableCell>{formatNumber(partner.qpsLimit)}</TableCell>
                  <TableCell>{formatNumber(partner.totalRequests)}</TableCell>
                  <TableCell>{formatNumber(partner.totalWins)}</TableCell>
                  <TableCell>
                    {((partner.totalWins / partner.totalBids) * 100).toFixed(2)}%
                  </TableCell>
                  <TableCell>{formatCurrency(partner.totalSpend)}</TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleEdit(partner)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleDelete(partner.id)}
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
