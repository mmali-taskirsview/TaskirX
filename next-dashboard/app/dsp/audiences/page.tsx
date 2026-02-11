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
  Users,
  Target,
  TrendingUp,
  RefreshCw,
  UserCheck,
  Globe,
  Clock,
  Code
} from 'lucide-react'

interface AudienceSegment {
  id: string
  name: string
  description: string
  status: string
  type: string
  size: number
  matchRate: number
  cpmModifier: number
  rules: any
  demographics: {
    ageRange?: string
    gender?: string
    income?: string
  }
  lookbackDays: number
  createdAt: string
}

export default function AudiencesPage() {
  const [audiences, setAudiences] = useState<AudienceSegment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingAudience, setEditingAudience] = useState<AudienceSegment | null>(null)
  const [pixelDialogOpen, setPixelDialogOpen] = useState(false)
  const [selectedAudienceForPixel, setSelectedAudienceForPixel] = useState<AudienceSegment | null>(null)
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    type: 'first_party',
    cpmModifier: 1.0,
    lookbackDays: 30,
    ageRange: '',
    gender: '',
    income: ''
  })
  
  // New: Audience Types from System Specs
  const AUDIENCE_TYPES = [
      { value: 'first_party', label: 'First Party Data (CRM/Pixel)' },
      { value: 'third_party', label: 'Third Party Data Provider' },
      { value: 'lookalike', label: 'Lookalike Audience' },
      { value: 'contextual', label: 'Contextual Interest' },
      { value: 'retargeting', label: 'Retargeting' },
      { value: 'demographic', label: 'Demographic (Age/Gender/Income)' },
      { value: 'psychographic', label: 'Psychographic (Lifestyle/Values)' },
      { value: 'behavioral', label: 'Behavioral (Intent/Purchase)' },
      { value: 'b2b', label: 'B2B (Firmographic)' },
      { value: 'intent', label: 'In-Market Intent' }
  ];

  const fetchAudiences = async () => {
    try {
      setLoading(true)
      const response = await api.getAudiences()
      setAudiences(response.data)
      setError(null)
    } catch (err: any) {
      setError(err.message || 'Failed to load audiences')
      // Demo data
      setAudiences([
        {
          id: '1',
          name: 'High-Value Shoppers',
          description: 'Users who have made purchases over $500 in the last 30 days',
          status: 'active',
          type: 'first_party',
          size: 2450000,
          matchRate: 78.5,
          cpmModifier: 1.8,
          rules: { purchaseAmount: { min: 500 }, recency: 30 },
          demographics: { ageRange: '25-54', income: 'high' },
          lookbackDays: 30,
          createdAt: '2024-01-15'
        },
        {
          id: '2',
          name: 'Cart Abandoners',
          description: 'Users who added items to cart but didn\'t complete purchase',
          status: 'active',
          type: 'retargeting',
          size: 890000,
          matchRate: 92.3,
          cpmModifier: 2.2,
          rules: { cartAbandoned: true, recency: 7 },
          demographics: {},
          lookbackDays: 7,
          createdAt: '2024-01-18'
        },
        {
          id: '3',
          name: 'Tech Enthusiasts',
          description: 'Users interested in technology and gadgets',
          status: 'active',
          type: 'third_party',
          size: 15700000,
          matchRate: 65.2,
          cpmModifier: 1.4,
          rules: { interests: ['technology', 'gadgets', 'electronics'] },
          demographics: { ageRange: '18-44', gender: 'all' },
          lookbackDays: 90,
          createdAt: '2024-01-20'
        },
        {
          id: '4',
          name: 'Lookalike - Top Converters',
          description: 'Similar users to top 5% converters',
          status: 'active',
          type: 'lookalike',
          size: 8500000,
          matchRate: 45.8,
          cpmModifier: 1.6,
          rules: { seedAudience: 'top_converters', similarity: 0.95 },
          demographics: {},
          lookbackDays: 60,
          createdAt: '2024-02-01'
        },
        {
          id: '5',
          name: 'Finance Contextual',
          description: 'Users viewing finance and investment content',
          status: 'active',
          type: 'contextual',
          size: 12300000,
          matchRate: 88.1,
          cpmModifier: 1.5,
          rules: { contentCategories: ['finance', 'investing', 'banking'] },
          demographics: { income: 'medium-high' },
          lookbackDays: 1,
          createdAt: '2024-02-05'
        },
        {
          id: '6',
          name: 'Fitness & Health',
          description: 'Health-conscious users interested in fitness',
          status: 'paused',
          type: 'third_party',
          size: 9800000,
          matchRate: 71.4,
          cpmModifier: 1.3,
          rules: { interests: ['fitness', 'health', 'wellness', 'nutrition'] },
          demographics: { ageRange: '25-54' },
          lookbackDays: 60,
          createdAt: '2024-02-10'
        }
      ])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchAudiences()
  }, [])

  const generatePixelCode = (audienceId: string) => {
    const baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000/api'
    const trackingUrl = `${baseUrl}/dsp/pixel?id=${audienceId}&evt=page_view`
    return `<img src="${trackingUrl}" width="1" height="1" style="display:none;" />`
  }

  const handleSubmit = async () => {
    try {
      const payload = {
        ...formData,
        demographics: {
          ageRange: formData.ageRange || undefined,
          gender: formData.gender || undefined,
          income: formData.income || undefined
        }
      }
      if (editingAudience) {
        await api.updateAudience(editingAudience.id, payload)
      } else {
        await api.createAudience(payload)
      }
      setDialogOpen(false)
      setEditingAudience(null)
      resetForm()
      fetchAudiences()
    } catch (err: any) {
      alert(err.message || 'Failed to save audience')
    }
  }

  const handleEdit = (audience: AudienceSegment) => {
    setEditingAudience(audience)
    setFormData({
      name: audience.name,
      description: audience.description,
      type: audience.type,
      cpmModifier: audience.cpmModifier,
      lookbackDays: audience.lookbackDays,
      ageRange: audience.demographics?.ageRange || '',
      gender: audience.demographics?.gender || '',
      income: audience.demographics?.income || ''
    })
    setDialogOpen(true)
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this audience?')) return
    try {
      await api.deleteAudience(id)
      fetchAudiences()
    } catch (err: any) {
      alert(err.message || 'Failed to delete audience')
    }
  }

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      type: 'first_party',
      cpmModifier: 1.0,
      lookbackDays: 30,
      ageRange: '',
      gender: '',
      income: ''
    })
  }

  const formatNumber = (num: number) => {
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
    return num.toLocaleString()
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <Badge className="bg-green-500">Active</Badge>
      case 'paused':
        return <Badge variant="secondary">Paused</Badge>
      case 'building':
        return <Badge variant="outline">Building</Badge>
      default:
        return <Badge variant="destructive">Inactive</Badge>
    }
  }

  const getTypeBadge = (type: string) => {
    const types: Record<string, { label: string; color: string; icon: any }> = {
      first_party: { label: '1st Party', color: 'text-green-500 border-green-500', icon: UserCheck },
      third_party: { label: '3rd Party', color: 'text-blue-500 border-blue-500', icon: Globe },
      lookalike: { label: 'Lookalike', color: 'text-purple-500 border-purple-500', icon: Users },
      contextual: { label: 'Contextual', color: 'text-orange-500 border-orange-500', icon: Target },
      retargeting: { label: 'Retargeting', color: 'text-red-500 border-red-500', icon: Clock },
      // New types
      demographic: { label: 'Demographic', color: 'text-indigo-500 border-indigo-500', icon: Users },
      psychographic: { label: 'Psychographic', color: 'text-pink-500 border-pink-500', icon: Target },
      behavioral: { label: 'Behavioral', color: 'text-amber-500 border-amber-500', icon: TrendingUp },
      b2b: { label: 'B2B', color: 'text-slate-500 border-slate-500', icon: Globe },
      intent: { label: 'In-Market', color: 'text-emerald-500 border-emerald-500', icon: Target }
    }
    const config = types[type] || { label: type, color: '', icon: Users }
    return <Badge variant="outline" className={config.color}>{config.label}</Badge>
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
          <h1 className="text-3xl font-bold">Audience Segments</h1>
          <p className="text-muted-foreground">
            Manage targeting audiences for campaign optimization
          </p>
        </div>
        <div className="flex gap-2">
          <Button onClick={fetchAudiences} variant="outline">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button onClick={() => { setEditingAudience(null); resetForm(); }}>
                <Plus className="h-4 w-4 mr-2" />
                New Audience
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle>{editingAudience ? 'Edit' : 'Create'} Audience Segment</DialogTitle>
                <DialogDescription>
                  Define audience targeting rules and demographics
                </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="space-y-2">
                  <Label>Audience Name</Label>
                  <Input
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="e.g., High-Value Shoppers"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Description</Label>
                  <Textarea
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    placeholder="Describe your audience segment..."
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Audience Type</Label>
                    <Select
                      value={formData.type}
                      onValueChange={(value) => setFormData({ ...formData, type: value })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {AUDIENCE_TYPES.map((type) => (
                          <SelectItem key={type.value} value={type.value}>
                            {type.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label>CPM Modifier</Label>
                    <Input
                      type="number"
                      step="0.1"
                      value={formData.cpmModifier}
                      onChange={(e) => setFormData({ ...formData, cpmModifier: parseFloat(e.target.value) })}
                    />
                    <p className="text-xs text-muted-foreground">1.0 = no change, 1.5 = +50% bid</p>
                  </div>
                </div>
                <div className="space-y-2">
                  <Label>Lookback Days</Label>
                  <Input
                    type="number"
                    value={formData.lookbackDays}
                    onChange={(e) => setFormData({ ...formData, lookbackDays: parseInt(e.target.value) })}
                  />
                </div>
                <div className="border-t pt-4">
                  <Label className="text-lg">Demographics</Label>
                </div>
                <div className="grid grid-cols-3 gap-4">
                  <div className="space-y-2">
                    <Label>Age Range</Label>
                    <Select
                      value={formData.ageRange}
                      onValueChange={(value) => setFormData({ ...formData, ageRange: value })}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Any" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="">Any</SelectItem>
                        <SelectItem value="18-24">18-24</SelectItem>
                        <SelectItem value="25-34">25-34</SelectItem>
                        <SelectItem value="35-44">35-44</SelectItem>
                        <SelectItem value="45-54">45-54</SelectItem>
                        <SelectItem value="55-64">55-64</SelectItem>
                        <SelectItem value="65+">65+</SelectItem>
                        <SelectItem value="18-44">18-44</SelectItem>
                        <SelectItem value="25-54">25-54</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label>Gender</Label>
                    <Select
                      value={formData.gender}
                      onValueChange={(value) => setFormData({ ...formData, gender: value })}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Any" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="">Any</SelectItem>
                        <SelectItem value="male">Male</SelectItem>
                        <SelectItem value="female">Female</SelectItem>
                        <SelectItem value="all">All</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label>Income</Label>
                    <Select
                      value={formData.income}
                      onValueChange={(value) => setFormData({ ...formData, income: value })}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Any" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="">Any</SelectItem>
                        <SelectItem value="low">Low</SelectItem>
                        <SelectItem value="medium">Medium</SelectItem>
                        <SelectItem value="medium-high">Medium-High</SelectItem>
                        <SelectItem value="high">High</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <Button onClick={handleSubmit} className="w-full">
                  {editingAudience ? 'Update' : 'Create'} Audience
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
              <Users className="h-4 w-4 text-purple-500" />
              Total Audiences
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{audiences.length}</div>
            <p className="text-xs text-muted-foreground">
              {audiences.filter(a => a.status === 'active').length} active
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Target className="h-4 w-4 text-blue-500" />
              Total Reach
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatNumber(audiences.reduce((sum, a) => sum + a.size, 0))}
            </div>
            <p className="text-xs text-muted-foreground">unique users</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <TrendingUp className="h-4 w-4 text-green-500" />
              Avg Match Rate
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {(audiences.reduce((sum, a) => sum + a.matchRate, 0) / audiences.length || 0).toFixed(1)}%
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <UserCheck className="h-4 w-4 text-orange-500" />
              1st Party
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {audiences.filter(a => a.type === 'first_party').length}
            </div>
            <p className="text-xs text-muted-foreground">audiences</p>
          </CardContent>
        </Card>
      </div>

      {/* Audiences Table */}
      <Card>
        <CardHeader>
          <CardTitle>Audience Segments</CardTitle>
          <CardDescription>All targeting audiences and their performance</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Audience</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Size</TableHead>
                <TableHead>Match Rate</TableHead>
                <TableHead>CPM Modifier</TableHead>
                <TableHead>Lookback</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {audiences.map((audience) => (
                <TableRow key={audience.id}>
                  <TableCell>
                    <div>
                      <div className="font-medium">{audience.name}</div>
                      <div className="text-xs text-muted-foreground truncate max-w-[250px]">
                        {audience.description}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>{getStatusBadge(audience.status)}</TableCell>
                  <TableCell>{getTypeBadge(audience.type)}</TableCell>
                  <TableCell>{formatNumber(audience.size)}</TableCell>
                  <TableCell>
                    <span className={audience.matchRate >= 70 ? 'text-green-500' : audience.matchRate >= 50 ? 'text-yellow-500' : 'text-red-500'}>
                      {audience.matchRate.toFixed(1)}%
                    </span>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">
                      {audience.cpmModifier > 1 ? '+' : ''}{((audience.cpmModifier - 1) * 100).toFixed(0)}%
                    </Badge>
                  </TableCell>
                  <TableCell>{audience.lookbackDays} days</TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      {audience.type === 'retargeting' && (
                        <Button
                          size="sm"
                          variant="ghost"
                          title="Get Pixel Code"
                          onClick={() => {
                            setSelectedAudienceForPixel(audience)
                            setPixelDialogOpen(true)
                          }}
                        >
                          <Code className="h-4 w-4 text-blue-500" />
                        </Button>
                      )}
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleEdit(audience)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleDelete(audience.id)}
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

      <Dialog open={pixelDialogOpen} onOpenChange={setPixelDialogOpen}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>Retargeting Pixel Code</DialogTitle>
            <DialogDescription>
              Install this pixel on your website to populate the <strong>{selectedAudienceForPixel?.name}</strong> audience.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>Pixel HTML Code</Label>
              <Textarea 
                readOnly
                className="font-mono text-xs h-32"
                value={selectedAudienceForPixel ? generatePixelCode(selectedAudienceForPixel.id) : ''}
              />
            </div>
            <div className="rounded-md bg-muted p-4 text-sm">
              <p className="font-semibold mb-2">Instructions:</p>
              <ol className="list-decimal list-inside space-y-1">
                <li>Copy the HTML code above.</li>
                <li>Paste it into the &lt;body&gt; of your website pages.</li>
                <li>The pixel will fire properly for every page load.</li>
                <li>Audience size will update hourly.</li>
              </ol>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
