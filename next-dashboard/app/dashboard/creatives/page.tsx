'use client'

import { useState, useRef } from 'react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/card'
import {
  Image as ImageIcon,
  Video,
  FileCode,
  Upload,
  Trash2,
  Edit,
  Eye,
  Copy,
  Download,
  Filter,
  Search,
  Plus,
  CheckCircle2,
  XCircle,
  Clock,
  Folder,
  Grid3X3,
  List,
  MoreVertical,
  Play,
  Pause,
  Volume2,
  VolumeX,
} from 'lucide-react'
import { formatNumber, formatCurrency } from '@/lib/utils'

// Types
interface Creative {
  id: string
  name: string
  type: 'image' | 'video' | 'html5' | 'playable'
  format: string
  dimensions: string
  fileSize: string
  status: 'active' | 'pending' | 'rejected' | 'archived'
  campaigns: number
  impressions: number
  clicks: number
  ctr: number
  uploadDate: string
  thumbnail: string
  url: string
  tags: string[]
}

// Sample data
const creatives: Creative[] = [
  {
    id: '1',
    name: 'Summer Sale Banner - 300x250',
    type: 'image',
    format: 'Banner',
    dimensions: '300x250',
    fileSize: '45 KB',
    status: 'active',
    campaigns: 3,
    impressions: 1250000,
    clicks: 37500,
    ctr: 3.0,
    uploadDate: '2026-01-15',
    thumbnail: '/creatives/banner-summer.jpg',
    url: 'https://cdn.taskirx.com/creatives/banner-summer-300x250.jpg',
    tags: ['summer', 'sale', 'banner'],
  },
  {
    id: '2',
    name: 'App Promo Video - 15s',
    type: 'video',
    format: 'Rewarded Video',
    dimensions: '1920x1080',
    fileSize: '2.4 MB',
    status: 'active',
    campaigns: 2,
    impressions: 850000,
    clicks: 42500,
    ctr: 5.0,
    uploadDate: '2026-01-12',
    thumbnail: '/creatives/video-app-promo.jpg',
    url: 'https://cdn.taskirx.com/creatives/app-promo-15s.mp4',
    tags: ['video', 'app', 'promo'],
  },
  {
    id: '3',
    name: 'Interactive Playable - Puzzle Game',
    type: 'playable',
    format: 'Playable Ad',
    dimensions: '320x480',
    fileSize: '1.8 MB',
    status: 'active',
    campaigns: 1,
    impressions: 320000,
    clicks: 28800,
    ctr: 9.0,
    uploadDate: '2026-01-10',
    thumbnail: '/creatives/playable-puzzle.jpg',
    url: 'https://cdn.taskirx.com/creatives/playable-puzzle.html',
    tags: ['playable', 'game', 'interactive'],
  },
  {
    id: '4',
    name: 'Native Article Card',
    type: 'html5',
    format: 'Native Ad',
    dimensions: 'Responsive',
    fileSize: '120 KB',
    status: 'active',
    campaigns: 4,
    impressions: 980000,
    clicks: 34300,
    ctr: 3.5,
    uploadDate: '2026-01-08',
    thumbnail: '/creatives/native-card.jpg',
    url: 'https://cdn.taskirx.com/creatives/native-article.html',
    tags: ['native', 'article', 'responsive'],
  },
  {
    id: '5',
    name: 'Holiday Special Banner - 728x90',
    type: 'image',
    format: 'Leaderboard',
    dimensions: '728x90',
    fileSize: '38 KB',
    status: 'pending',
    campaigns: 0,
    impressions: 0,
    clicks: 0,
    ctr: 0,
    uploadDate: '2026-01-20',
    thumbnail: '/creatives/banner-holiday.jpg',
    url: 'https://cdn.taskirx.com/creatives/holiday-728x90.jpg',
    tags: ['holiday', 'banner', 'leaderboard'],
  },
  {
    id: '6',
    name: 'Brand Intro Video - 30s',
    type: 'video',
    format: 'Interstitial Video',
    dimensions: '1080x1920',
    fileSize: '5.2 MB',
    status: 'rejected',
    campaigns: 0,
    impressions: 0,
    clicks: 0,
    ctr: 0,
    uploadDate: '2026-01-18',
    thumbnail: '/creatives/video-brand.jpg',
    url: 'https://cdn.taskirx.com/creatives/brand-intro-30s.mp4',
    tags: ['video', 'brand', 'intro'],
  },
  {
    id: '7',
    name: 'Product Showcase - MREC',
    type: 'image',
    format: 'MREC',
    dimensions: '300x250',
    fileSize: '52 KB',
    status: 'active',
    campaigns: 2,
    impressions: 620000,
    clicks: 18600,
    ctr: 3.0,
    uploadDate: '2026-01-05',
    thumbnail: '/creatives/product-showcase.jpg',
    url: 'https://cdn.taskirx.com/creatives/product-mrec.jpg',
    tags: ['product', 'mrec', 'showcase'],
  },
  {
    id: '8',
    name: 'Offerwall Banner Set',
    type: 'html5',
    format: 'Offerwall',
    dimensions: 'Multiple',
    fileSize: '890 KB',
    status: 'active',
    campaigns: 1,
    impressions: 180000,
    clicks: 16200,
    ctr: 9.0,
    uploadDate: '2026-01-02',
    thumbnail: '/creatives/offerwall-set.jpg',
    url: 'https://cdn.taskirx.com/creatives/offerwall-set.zip',
    tags: ['offerwall', 'set', 'multiple'],
  },
]

// Creative specifications
const creativeSpecs = [
  { format: 'Banner 300x250', dimensions: '300x250', maxSize: '150 KB', types: ['JPG', 'PNG', 'GIF'] },
  { format: 'Banner 320x50', dimensions: '320x50', maxSize: '100 KB', types: ['JPG', 'PNG', 'GIF'] },
  { format: 'Banner 728x90', dimensions: '728x90', maxSize: '150 KB', types: ['JPG', 'PNG', 'GIF'] },
  { format: 'Interstitial', dimensions: '320x480 / 480x320', maxSize: '200 KB', types: ['JPG', 'PNG', 'GIF'] },
  { format: 'Rewarded Video', dimensions: '1920x1080', maxSize: '10 MB', types: ['MP4', 'MOV'] },
  { format: 'Playable Ad', dimensions: '320x480 / 480x320', maxSize: '5 MB', types: ['HTML5', 'ZIP'] },
  { format: 'Native Ad', dimensions: 'Responsive', maxSize: '500 KB', types: ['HTML5', 'JSON'] },
]

export default function CreativesPage() {
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [searchQuery, setSearchQuery] = useState('')
  const [filterType, setFilterType] = useState<'all' | 'image' | 'video' | 'html5' | 'playable'>('all')
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'pending' | 'rejected'>('all')
  const [showUploadModal, setShowUploadModal] = useState(false)
  const [showPreview, setShowPreview] = useState<Creative | null>(null)
  const [selectedCreatives, setSelectedCreatives] = useState<string[]>([])
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Filter creatives
  const filteredCreatives = creatives.filter(creative => {
    const matchesSearch = creative.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          creative.tags.some(tag => tag.toLowerCase().includes(searchQuery.toLowerCase()))
    const matchesType = filterType === 'all' || creative.type === filterType
    const matchesStatus = filterStatus === 'all' || creative.status === filterStatus
    return matchesSearch && matchesType && matchesStatus
  })

  // Stats
  const totalCreatives = creatives.length
  const activeCreatives = creatives.filter(c => c.status === 'active').length
  const totalImpressions = creatives.reduce((sum, c) => sum + c.impressions, 0)
  const avgCTR = creatives.filter(c => c.ctr > 0).reduce((sum, c) => sum + c.ctr, 0) / 
                 creatives.filter(c => c.ctr > 0).length

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'image': return <ImageIcon className="h-4 w-4" />
      case 'video': return <Video className="h-4 w-4" />
      case 'html5': return <FileCode className="h-4 w-4" />
      case 'playable': return <Play className="h-4 w-4" />
      default: return <ImageIcon className="h-4 w-4" />
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900/30 dark:text-green-400">
            <CheckCircle2 className="h-3 w-3" />
            Active
          </span>
        )
      case 'pending':
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-yellow-100 px-2 py-0.5 text-xs font-medium text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400">
            <Clock className="h-3 w-3" />
            Pending
          </span>
        )
      case 'rejected':
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-red-100 px-2 py-0.5 text-xs font-medium text-red-700 dark:bg-red-900/30 dark:text-red-400">
            <XCircle className="h-3 w-3" />
            Rejected
          </span>
        )
      default:
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700 dark:bg-gray-800 dark:text-gray-400">
            Archived
          </span>
        )
    }
  }

  const toggleSelect = (id: string) => {
    setSelectedCreatives(prev => 
      prev.includes(id) ? prev.filter(i => i !== id) : [...prev, id]
    )
  }

  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Creative Library</h1>
          <p className="text-muted-foreground">
            Manage and organize your ad creatives across all formats
          </p>
        </div>
        <button
          onClick={() => setShowUploadModal(true)}
          className="inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 px-4 py-2 text-sm font-medium text-white shadow-lg hover:opacity-90 transition-opacity"
        >
          <Upload className="h-4 w-4" />
          Upload Creative
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Creatives</p>
                <p className="text-2xl font-bold">{totalCreatives}</p>
              </div>
              <div className="rounded-full bg-blue-100 p-3 dark:bg-blue-900/30">
                <Folder className="h-5 w-5 text-blue-600" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Active Creatives</p>
                <p className="text-2xl font-bold">{activeCreatives}</p>
              </div>
              <div className="rounded-full bg-green-100 p-3 dark:bg-green-900/30">
                <CheckCircle2 className="h-5 w-5 text-green-600" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Impressions</p>
                <p className="text-2xl font-bold">{formatNumber(totalImpressions)}</p>
              </div>
              <div className="rounded-full bg-purple-100 p-3 dark:bg-purple-900/30">
                <Eye className="h-5 w-5 text-purple-600" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Avg CTR</p>
                <p className="text-2xl font-bold">{avgCTR.toFixed(1)}%</p>
              </div>
              <div className="rounded-full bg-orange-100 p-3 dark:bg-orange-900/30">
                <Play className="h-5 w-5 text-orange-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Filters and Search */}
      <div className="flex flex-wrap items-center gap-4">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search creatives..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full rounded-lg border bg-background pl-10 pr-4 py-2"
          />
        </div>
        
        <select
          value={filterType}
          onChange={(e) => setFilterType(e.target.value as typeof filterType)}
          className="rounded-lg border bg-background px-3 py-2"
        >
          <option value="all">All Types</option>
          <option value="image">Images</option>
          <option value="video">Videos</option>
          <option value="html5">HTML5</option>
          <option value="playable">Playable</option>
        </select>

        <select
          value={filterStatus}
          onChange={(e) => setFilterStatus(e.target.value as typeof filterStatus)}
          className="rounded-lg border bg-background px-3 py-2"
        >
          <option value="all">All Status</option>
          <option value="active">Active</option>
          <option value="pending">Pending</option>
          <option value="rejected">Rejected</option>
        </select>

        <div className="flex items-center gap-1 rounded-lg border p-1">
          <button
            onClick={() => setViewMode('grid')}
            className={`rounded p-2 ${viewMode === 'grid' ? 'bg-muted' : 'hover:bg-muted/50'}`}
          >
            <Grid3X3 className="h-4 w-4" />
          </button>
          <button
            onClick={() => setViewMode('list')}
            className={`rounded p-2 ${viewMode === 'list' ? 'bg-muted' : 'hover:bg-muted/50'}`}
          >
            <List className="h-4 w-4" />
          </button>
        </div>

        {selectedCreatives.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">{selectedCreatives.length} selected</span>
            <button className="rounded-lg border px-3 py-1.5 text-sm hover:bg-muted">
              <Trash2 className="h-4 w-4 text-red-500" />
            </button>
            <button className="rounded-lg border px-3 py-1.5 text-sm hover:bg-muted">
              <Download className="h-4 w-4" />
            </button>
          </div>
        )}
      </div>

      {/* Creatives Grid/List */}
      {viewMode === 'grid' ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {filteredCreatives.map((creative) => (
            <Card 
              key={creative.id} 
              className={`overflow-hidden cursor-pointer transition-all hover:shadow-lg ${
                selectedCreatives.includes(creative.id) ? 'ring-2 ring-blue-500' : ''
              }`}
            >
              {/* Thumbnail */}
              <div 
                className="relative aspect-video bg-gradient-to-br from-gray-100 to-gray-200 dark:from-gray-800 dark:to-gray-900"
                onClick={() => setShowPreview(creative)}
              >
                <div className="absolute inset-0 flex items-center justify-center">
                  {creative.type === 'video' ? (
                    <div className="rounded-full bg-black/50 p-3">
                      <Play className="h-8 w-8 text-white" />
                    </div>
                  ) : creative.type === 'playable' ? (
                    <div className="rounded-full bg-purple-500/80 p-3">
                      <Play className="h-8 w-8 text-white" />
                    </div>
                  ) : (
                    <ImageIcon className="h-12 w-12 text-gray-400" />
                  )}
                </div>
                {/* Type Badge */}
                <div className="absolute top-2 left-2 rounded-full bg-black/50 px-2 py-1 text-xs text-white flex items-center gap-1">
                  {getTypeIcon(creative.type)}
                  {creative.type.toUpperCase()}
                </div>
                {/* Select Checkbox */}
                <div 
                  className="absolute top-2 right-2"
                  onClick={(e) => { e.stopPropagation(); toggleSelect(creative.id); }}
                >
                  <div className={`h-5 w-5 rounded border-2 flex items-center justify-center ${
                    selectedCreatives.includes(creative.id) 
                      ? 'bg-blue-500 border-blue-500' 
                      : 'bg-white/80 border-gray-300'
                  }`}>
                    {selectedCreatives.includes(creative.id) && (
                      <CheckCircle2 className="h-3 w-3 text-white" />
                    )}
                  </div>
                </div>
              </div>
              
              <CardContent className="p-4">
                <div className="flex items-start justify-between mb-2">
                  <h3 className="font-medium text-sm line-clamp-2">{creative.name}</h3>
                  <button className="p-1 hover:bg-muted rounded">
                    <MoreVertical className="h-4 w-4" />
                  </button>
                </div>
                
                <div className="flex items-center gap-2 mb-3">
                  {getStatusBadge(creative.status)}
                  <span className="text-xs text-muted-foreground">{creative.dimensions}</span>
                </div>

                <div className="grid grid-cols-3 gap-2 text-center text-xs">
                  <div>
                    <div className="font-semibold">{formatNumber(creative.impressions)}</div>
                    <div className="text-muted-foreground">Impr.</div>
                  </div>
                  <div>
                    <div className="font-semibold">{formatNumber(creative.clicks)}</div>
                    <div className="text-muted-foreground">Clicks</div>
                  </div>
                  <div>
                    <div className="font-semibold">{creative.ctr}%</div>
                    <div className="text-muted-foreground">CTR</div>
                  </div>
                </div>

                <div className="mt-3 pt-3 border-t flex items-center justify-between">
                  <span className="text-xs text-muted-foreground">{creative.campaigns} campaigns</span>
                  <div className="flex items-center gap-1">
                    <button 
                      onClick={() => setShowPreview(creative)}
                      className="p-1.5 rounded hover:bg-muted"
                    >
                      <Eye className="h-4 w-4" />
                    </button>
                    <button className="p-1.5 rounded hover:bg-muted">
                      <Edit className="h-4 w-4" />
                    </button>
                    <button className="p-1.5 rounded hover:bg-muted">
                      <Copy className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <Card>
          <CardContent className="p-0">
            <table className="w-full">
              <thead>
                <tr className="border-b">
                  <th className="p-4 text-left">
                    <input type="checkbox" className="rounded" />
                  </th>
                  <th className="p-4 text-left font-medium">Creative</th>
                  <th className="p-4 text-left font-medium">Type</th>
                  <th className="p-4 text-left font-medium">Dimensions</th>
                  <th className="p-4 text-left font-medium">Status</th>
                  <th className="p-4 text-right font-medium">Impressions</th>
                  <th className="p-4 text-right font-medium">CTR</th>
                  <th className="p-4 text-left font-medium">Campaigns</th>
                  <th className="p-4 text-left font-medium">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {filteredCreatives.map((creative) => (
                  <tr key={creative.id} className="hover:bg-muted/50">
                    <td className="p-4">
                      <input 
                        type="checkbox" 
                        className="rounded"
                        checked={selectedCreatives.includes(creative.id)}
                        onChange={() => toggleSelect(creative.id)}
                      />
                    </td>
                    <td className="p-4">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
                          {getTypeIcon(creative.type)}
                        </div>
                        <div>
                          <div className="font-medium">{creative.name}</div>
                          <div className="text-xs text-muted-foreground">{creative.fileSize}</div>
                        </div>
                      </div>
                    </td>
                    <td className="p-4">
                      <span className="capitalize">{creative.type}</span>
                    </td>
                    <td className="p-4">{creative.dimensions}</td>
                    <td className="p-4">{getStatusBadge(creative.status)}</td>
                    <td className="p-4 text-right">{formatNumber(creative.impressions)}</td>
                    <td className="p-4 text-right">{creative.ctr}%</td>
                    <td className="p-4">{creative.campaigns}</td>
                    <td className="p-4">
                      <div className="flex items-center gap-1">
                        <button 
                          onClick={() => setShowPreview(creative)}
                          className="p-1.5 rounded hover:bg-muted"
                        >
                          <Eye className="h-4 w-4" />
                        </button>
                        <button className="p-1.5 rounded hover:bg-muted">
                          <Edit className="h-4 w-4" />
                        </button>
                        <button className="p-1.5 rounded hover:bg-muted text-red-500">
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      )}

      {/* Creative Specifications */}
      <Card>
        <CardHeader>
          <CardTitle>Creative Specifications</CardTitle>
          <CardDescription>Supported formats and requirements</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="pb-3 text-left font-medium">Format</th>
                  <th className="pb-3 text-left font-medium">Dimensions</th>
                  <th className="pb-3 text-left font-medium">Max Size</th>
                  <th className="pb-3 text-left font-medium">File Types</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {creativeSpecs.map((spec, index) => (
                  <tr key={index}>
                    <td className="py-3 font-medium">{spec.format}</td>
                    <td className="py-3">{spec.dimensions}</td>
                    <td className="py-3">{spec.maxSize}</td>
                    <td className="py-3">
                      <div className="flex gap-1">
                        {spec.types.map((type) => (
                          <span key={type} className="rounded bg-gray-100 px-2 py-0.5 text-xs dark:bg-gray-800">
                            {type}
                          </span>
                        ))}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {/* Upload Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-2xl rounded-lg bg-white p-6 shadow-xl dark:bg-gray-900">
            <h2 className="text-xl font-bold mb-4">Upload Creative</h2>
            
            <div className="space-y-4">
              {/* Drop Zone */}
              <div 
                className="border-2 border-dashed rounded-lg p-8 text-center cursor-pointer hover:border-blue-500 hover:bg-blue-50/50 dark:hover:bg-blue-900/10 transition-colors"
                onClick={() => fileInputRef.current?.click()}
              >
                <Upload className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                <p className="font-medium">Drop files here or click to upload</p>
                <p className="text-sm text-muted-foreground mt-1">
                  Supports JPG, PNG, GIF, MP4, MOV, HTML5, ZIP
                </p>
                <input
                  ref={fileInputRef}
                  type="file"
                  className="hidden"
                  multiple
                  accept="image/*,video/*,.html,.zip"
                />
              </div>

              {/* Creative Details */}
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium mb-1">Creative Name</label>
                  <input
                    type="text"
                    placeholder="e.g., Summer Sale Banner"
                    className="w-full rounded-lg border px-3 py-2"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Format</label>
                  <select className="w-full rounded-lg border px-3 py-2">
                    <option>Banner 300x250</option>
                    <option>Banner 320x50</option>
                    <option>Banner 728x90</option>
                    <option>Interstitial</option>
                    <option>Rewarded Video</option>
                    <option>Playable Ad</option>
                    <option>Native Ad</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Tags</label>
                <input
                  type="text"
                  placeholder="e.g., summer, sale, promo (comma separated)"
                  className="w-full rounded-lg border px-3 py-2"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Destination URL</label>
                <input
                  type="url"
                  placeholder="https://example.com/landing-page"
                  className="w-full rounded-lg border px-3 py-2"
                />
              </div>

              <div>
                <label className="flex items-center gap-2">
                  <input type="checkbox" className="rounded" />
                  <span className="text-sm">Auto-optimize for different placements</span>
                </label>
              </div>
            </div>

            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowUploadModal(false)}
                className="rounded-lg border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() => setShowUploadModal(false)}
                className="rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 px-4 py-2 text-sm font-medium text-white hover:opacity-90 transition-opacity"
              >
                Upload Creative
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Preview Modal */}
      {showPreview && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80">
          <div className="relative w-full max-w-4xl p-4">
            <button
              onClick={() => setShowPreview(null)}
              className="absolute top-2 right-2 rounded-full bg-white/10 p-2 text-white hover:bg-white/20"
            >
              <XCircle className="h-6 w-6" />
            </button>
            
            <div className="bg-white dark:bg-gray-900 rounded-lg overflow-hidden">
              {/* Preview Area */}
              <div className="aspect-video bg-gray-900 flex items-center justify-center">
                {showPreview.type === 'video' ? (
                  <div className="text-center text-white">
                    <Play className="h-16 w-16 mx-auto mb-4" />
                    <p>Video Preview</p>
                    <p className="text-sm text-gray-400">{showPreview.name}</p>
                  </div>
                ) : showPreview.type === 'playable' ? (
                  <div className="text-center text-white">
                    <Play className="h-16 w-16 mx-auto mb-4 text-purple-500" />
                    <p>Playable Ad Preview</p>
                    <p className="text-sm text-gray-400">Interactive demo would load here</p>
                  </div>
                ) : (
                  <div className="text-center text-white">
                    <ImageIcon className="h-16 w-16 mx-auto mb-4" />
                    <p>Image Preview</p>
                    <p className="text-sm text-gray-400">{showPreview.dimensions}</p>
                  </div>
                )}
              </div>
              
              {/* Details */}
              <div className="p-6">
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h3 className="text-xl font-bold">{showPreview.name}</h3>
                    <p className="text-muted-foreground">{showPreview.format} • {showPreview.dimensions}</p>
                  </div>
                  {getStatusBadge(showPreview.status)}
                </div>
                
                <div className="grid grid-cols-4 gap-4 mb-4">
                  <div className="text-center p-3 rounded-lg bg-muted/50">
                    <div className="text-2xl font-bold">{formatNumber(showPreview.impressions)}</div>
                    <div className="text-sm text-muted-foreground">Impressions</div>
                  </div>
                  <div className="text-center p-3 rounded-lg bg-muted/50">
                    <div className="text-2xl font-bold">{formatNumber(showPreview.clicks)}</div>
                    <div className="text-sm text-muted-foreground">Clicks</div>
                  </div>
                  <div className="text-center p-3 rounded-lg bg-muted/50">
                    <div className="text-2xl font-bold">{showPreview.ctr}%</div>
                    <div className="text-sm text-muted-foreground">CTR</div>
                  </div>
                  <div className="text-center p-3 rounded-lg bg-muted/50">
                    <div className="text-2xl font-bold">{showPreview.campaigns}</div>
                    <div className="text-sm text-muted-foreground">Campaigns</div>
                  </div>
                </div>
                
                <div className="flex items-center gap-2">
                  <span className="text-sm text-muted-foreground">Tags:</span>
                  {showPreview.tags.map((tag) => (
                    <span key={tag} className="rounded-full bg-blue-100 px-2 py-0.5 text-xs text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
