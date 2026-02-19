'use client'

import { useState, useRef, useEffect } from 'react'
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
import { api } from '@/lib/api'

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
  // Backend fields
  tenantId?: string
  userId?: string
  description?: string
  width?: number
  height?: number
  fileName?: string
  mimeType?: string
  filePath?: string
  metadata?: any
  targetingRules?: any
  performanceMetrics?: {
    impressions: number
    clicks: number
    ctr: number
    conversions: number
    cvr: number
    spend: number
    revenue: number
    roi: number
  }
  createdAt?: string
  updatedAt?: string
}

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
  const [creatives, setCreatives] = useState<Creative[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [searchQuery, setSearchQuery] = useState('')
  const [filterType, setFilterType] = useState<'all' | 'image' | 'video' | 'html5' | 'playable'>('all')
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'pending' | 'rejected'>('all')
  const [showUploadModal, setShowUploadModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [editingCreative, setEditingCreative] = useState<Creative | null>(null)
  const [showPreview, setShowPreview] = useState<Creative | null>(null)
  const [selectedCreatives, setSelectedCreatives] = useState<string[]>([])
  const [selectedFiles, setSelectedFiles] = useState<File[]>([])
  const [uploadForm, setUploadForm] = useState({
    name: '',
    format: 'Banner 300x250',
    tags: '',
    destinationUrl: '',
    autoOptimize: false
  })
  const [uploading, setUploading] = useState(false)
  const [uploadProgress, setUploadProgress] = useState(0)
  const [uploadError, setUploadError] = useState<string | null>(null)
  const [editError, setEditError] = useState<string | null>(null)
  const [editForm, setEditForm] = useState({
    name: '',
    tags: '',
    status: 'pending' as Creative['status'],
    description: ''
  })
  const [updating, setUpdating] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [deletingCreative, setDeletingCreative] = useState<Creative | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Load creatives on component mount
  useEffect(() => {
    loadCreatives()
  }, [])

  const loadCreatives = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await api.getCreatives()
      setCreatives(response.data || [])
    } catch (err) {
      console.error('Failed to load creatives:', err)
      setError('Failed to load creatives. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(event.target.files || [])
    setSelectedFiles(files)
    setUploadError(null)
    
    // Auto-generate name from first file if not set
    if (!uploadForm.name && files.length > 0) {
      const fileName = files[0].name.split('.')[0]
      setUploadForm(prev => ({ ...prev, name: fileName }))
    }
  }

  const validateFiles = (files: File[]): string | null => {
    if (files.length === 0) return 'Please select at least one file'
    
    const maxSizes = {
      'image': 150 * 1024, // 150KB
      'video': 10 * 1024 * 1024, // 10MB
      'html5': 500 * 1024, // 500KB
      'playable': 5 * 1024 * 1024, // 5MB
    }

    for (const file of files) {
      const fileType = file.type.startsWith('image/') ? 'image' : 
                      file.type.startsWith('video/') ? 'video' :
                      file.name.endsWith('.html') || file.name.endsWith('.zip') ? 'html5' : 'playable'
      
      const maxSize = maxSizes[fileType as keyof typeof maxSizes] || maxSizes.image
      
      if (file.size > maxSize) {
        return `${file.name} exceeds maximum size of ${Math.round(maxSize / 1024)}KB`
      }
    }
    
    return null
  }

  const handleUpload = async () => {
    if (selectedFiles.length === 0) {
      setUploadError('Please select at least one file')
      return
    }

    const validationError = validateFiles(selectedFiles)
    if (validationError) {
      setUploadError(validationError)
      return
    }

    if (!uploadForm.name.trim()) {
      setUploadError('Please enter a creative name')
      return
    }

    try {
      setUploading(true)
      setUploadProgress(0)
      setUploadError(null)

      // Create FormData for file upload
      const formData = new FormData()
      formData.append('name', uploadForm.name.trim())
      formData.append('format', uploadForm.format)
      formData.append('tags', uploadForm.tags)
      formData.append('destinationUrl', uploadForm.destinationUrl)
      formData.append('autoOptimize', uploadForm.autoOptimize.toString())
      
      // Add files
      selectedFiles.forEach((file, index) => {
        formData.append('files', file)
      })

      // Note: Since our API client doesn't handle FormData well, we'll use fetch directly
      const baseUrl = process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:3000'
      const token = localStorage.getItem('auth_token') // Assuming JWT token is stored here
      
      const response = await fetch(`${baseUrl}/api/creatives/upload`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
        body: formData,
      })

      if (!response.ok) {
        throw new Error(`Upload failed: ${response.statusText}`)
      }

      const result = await response.json()
      
      // Refresh creatives list
      await loadCreatives()
      
      // Reset form and close modal
      setSelectedFiles([])
      setUploadForm({
        name: '',
        format: 'Banner 300x250',
        tags: '',
        destinationUrl: '',
        autoOptimize: false
      })
      setShowUploadModal(false)
      
    } catch (err) {
      console.error('Upload failed:', err)
      setUploadError(err instanceof Error ? err.message : 'Upload failed. Please try again.')
    } finally {
      setUploading(false)
      setUploadProgress(0)
    }
  }

  const handleEditCreative = (creative: Creative) => {
    setEditingCreative(creative)
    setEditForm({
      name: creative.name,
      tags: creative.tags.join(', '),
      status: creative.status,
      description: creative.description || ''
    })
    setEditError(null)
    setShowEditModal(true)
  }

  const handleUpdateCreative = async () => {
    if (!editingCreative) return

    if (!editForm.name.trim()) {
      setEditError('Please enter a creative name')
      return
    }

    try {
      setUpdating(true)
      setEditError(null)

      const updateData = {
        name: editForm.name.trim(),
        tags: editForm.tags.split(',').map(tag => tag.trim()).filter(tag => tag),
        status: editForm.status,
        description: editForm.description.trim() || undefined
      }

      await api.updateCreative(editingCreative.id, updateData)
      
      // Refresh creatives list
      await loadCreatives()
      
      // Reset form and close modal
      setEditingCreative(null)
      setEditForm({
        name: '',
        tags: '',
        status: 'pending',
        description: ''
      })
      
    } catch (err) {
      console.error('Update failed:', err)
      setEditError(err instanceof Error ? err.message : 'Update failed. Please try again.')
    } finally {
      setUpdating(false)
    }
  }

  const handleDeleteCreative = async () => {
    if (!deletingCreative) return

    try {
      setDeleting(true)
      setDeleteError(null)

      await api.deleteCreative(deletingCreative.id)
      
      // Refresh creatives list
      await loadCreatives()
      
      // Close dialog and reset state
      setShowDeleteDialog(false)
      setDeletingCreative(null)
      
    } catch (err) {
      console.error('Delete failed:', err)
      setDeleteError(err instanceof Error ? err.message : 'Delete failed. Please try again.')
    } finally {
      setDeleting(false)
    }
  }

  const handleBulkStatusChange = async (status: Creative['status']) => {
    if (selectedCreatives.length === 0) return

    try {
      setUpdating(true)
      setEditError(null)

      // Update all selected creatives
      await Promise.all(
        selectedCreatives.map(id => api.updateCreative(id, { status }))
      )

      // Refresh creatives list
      await loadCreatives()
      
      // Clear selection
      setSelectedCreatives([])
      
    } catch (err) {
      console.error('Bulk status update failed:', err)
      setEditError(err instanceof Error ? err.message : 'Bulk update failed. Please try again.')
    } finally {
      setUpdating(false)
    }
  }

  const handleBulkDelete = async () => {
    if (selectedCreatives.length === 0) return

    if (!confirm(`Are you sure you want to delete ${selectedCreatives.length} creative${selectedCreatives.length !== 1 ? 's' : ''}? This action cannot be undone.`)) {
      return
    }

    try {
      setDeleting(true)
      setDeleteError(null)

      // Delete all selected creatives
      await Promise.all(
        selectedCreatives.map(id => api.deleteCreative(id))
      )

      // Refresh creatives list
      await loadCreatives()
      
      // Clear selection
      setSelectedCreatives([])
      
    } catch (err) {
      console.error('Bulk delete failed:', err)
      setDeleteError(err instanceof Error ? err.message : 'Bulk delete failed. Please try again.')
    } finally {
      setDeleting(false)
    }
  }

  const handleDuplicateCreative = async (creative: Creative) => {
    try {
      setUploading(true)
      setUploadError(null)

      // For now, show that duplication needs backend implementation
      // In a full implementation, this would copy the file and create a new creative
      alert('Creative duplication requires backend implementation to copy files. This feature creates a metadata copy but needs file duplication logic.')

      // Create metadata-only duplicate for demonstration
      const duplicateData = {
        name: `Copy of ${creative.name}`,
        tags: creative.tags,
        status: 'draft' as Creative['status'],
        description: creative.description,
      }

      // This would need a backend endpoint to duplicate with file copying
      console.log('Would duplicate creative:', duplicateData)

    } catch (err) {
      console.error('Duplicate failed:', err)
      setUploadError(err instanceof Error ? err.message : 'Duplicate failed. Please try again.')
    } finally {
      setUploading(false)
    }
  }

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
  const totalImpressions = creatives.reduce((sum, c) => sum + (c.performanceMetrics?.impressions || c.impressions || 0), 0)
  const avgCTR = creatives.length > 0 ? 
    creatives.reduce((sum, c) => sum + (c.performanceMetrics?.ctr || c.ctr || 0), 0) / creatives.length : 0

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

  const toggleSelectAll = () => {
    if (selectedCreatives.length === filteredCreatives.length) {
      setSelectedCreatives([])
    } else {
      setSelectedCreatives(filteredCreatives.map(c => c.id))
    }
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

      {/* Loading State */}
      {loading && (
        <div className="flex items-center justify-center py-12">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-muted-foreground">Loading creatives...</p>
          </div>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 mb-6">
          <div className="flex items-center gap-3">
            <XCircle className="h-5 w-5 text-red-600" />
            <p className="text-red-700 dark:text-red-400">{error}</p>
            <button
              onClick={loadCreatives}
              className="ml-auto text-sm text-red-600 hover:text-red-800 underline"
            >
              Retry
            </button>
          </div>
        </div>
      )}

      {/* Summary Cards */}
      {!loading && !error && (
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
      )}

      {/* Bulk Actions Toolbar */}
      {selectedCreatives.length > 0 && (
        <Card className="mb-4">
          <CardContent className="pt-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <span className="text-sm font-medium">
                  {selectedCreatives.length} creative{selectedCreatives.length !== 1 ? 's' : ''} selected
                </span>
                <button
                  onClick={() => setSelectedCreatives([])}
                  className="text-xs text-muted-foreground hover:text-foreground"
                >
                  Clear selection
                </button>
              </div>
              <div className="flex items-center gap-2">
                <select
                  className="text-sm rounded border px-2 py-1"
                  onChange={(e) => {
                    if (e.target.value) {
                      handleBulkStatusChange(e.target.value as Creative['status'])
                      e.target.value = ''
                    }
                  }}
                  defaultValue=""
                >
                  <option value="">Change status</option>
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                  <option value="draft">Draft</option>
                  <option value="archived">Archived</option>
                </select>
                <button
                  onClick={handleBulkDelete}
                  className="inline-flex items-center gap-1 rounded bg-red-600 px-3 py-1 text-xs font-medium text-white hover:bg-red-700"
                >
                  <Trash2 className="h-3 w-3" />
                  Delete
                </button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Creatives Grid/List */}
      {!loading && !error && viewMode === 'grid' ? (
        filteredCreatives.length === 0 ? (
          <div className="text-center py-12">
            <ImageIcon className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-medium mb-2">No creatives found</h3>
            <p className="text-muted-foreground mb-4">
              {creatives.length === 0 ? 'Get started by uploading your first creative.' : 'Try adjusting your filters.'}
            </p>
            {creatives.length === 0 && (
              <button
                onClick={() => setShowUploadModal(true)}
                className="inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 px-4 py-2 text-sm font-medium text-white shadow-lg hover:opacity-90 transition-opacity"
              >
                <Upload className="h-4 w-4" />
                Upload Creative
              </button>
            )}
          </div>
        ) : (
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
                  {/* Selection Checkbox */}
                  <div className="absolute top-2 left-2 z-10">
                    <input
                      type="checkbox"
                      checked={selectedCreatives.includes(creative.id)}
                      onChange={(e) => {
                        e.stopPropagation()
                        if (e.target.checked) {
                          setSelectedCreatives(prev => [...prev, creative.id])
                        } else {
                          setSelectedCreatives(prev => prev.filter(id => id !== creative.id))
                        }
                      }}
                      className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                    />
                  </div>
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
                    <span className="text-xs text-muted-foreground">
                      {creative.width && creative.height ? `${creative.width}x${creative.height}` : creative.dimensions}
                    </span>
                  </div>

                  <div className="grid grid-cols-3 gap-2 text-center text-xs">
                    <div>
                      <div className="font-semibold">{formatNumber(creative.performanceMetrics?.impressions || creative.impressions || 0)}</div>
                      <div className="text-muted-foreground">Impr.</div>
                    </div>
                    <div>
                      <div className="font-semibold">{formatNumber(creative.performanceMetrics?.clicks || creative.clicks || 0)}</div>
                      <div className="text-muted-foreground">Clicks</div>
                    </div>
                    <div>
                      <div className="font-semibold">{(creative.performanceMetrics?.ctr || creative.ctr || 0).toFixed(1)}%</div>
                      <div className="text-muted-foreground">CTR</div>
                    </div>
                  </div>

                  <div className="mt-3 pt-3 border-t flex items-center justify-between">
                    <span className="text-xs text-muted-foreground">{creative.campaigns || 0} campaigns</span>
                    <div className="flex items-center gap-1">
                      <button 
                        onClick={() => setShowPreview(creative)}
                        className="p-1.5 rounded hover:bg-muted"
                      >
                        <Eye className="h-4 w-4" />
                      </button>
                      <button 
                        onClick={(e) => { e.stopPropagation(); handleEditCreative(creative); }}
                        className="p-1.5 rounded hover:bg-muted"
                      >
                        <Edit className="h-4 w-4" />
                      </button>
                      <button 
                        onClick={(e) => { e.stopPropagation(); handleDuplicateCreative(creative); }}
                        className="p-1.5 rounded hover:bg-muted"
                        title="Duplicate creative"
                      >
                        <Copy className="h-4 w-4" />
                      </button>
                      <button 
                        onClick={(e) => { e.stopPropagation(); setDeletingCreative(creative); setShowDeleteDialog(true); }}
                        className="p-1.5 rounded hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )
      ) : !loading && !error && viewMode === 'list' ? (
        filteredCreatives.length === 0 ? (
          <div className="text-center py-12">
            <List className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-medium mb-2">No creatives found</h3>
            <p className="text-muted-foreground mb-4">
              {creatives.length === 0 ? 'Get started by uploading your first creative.' : 'Try adjusting your filters.'}
            </p>
            {creatives.length === 0 && (
              <button
                onClick={() => setShowUploadModal(true)}
                className="inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 px-4 py-2 text-sm font-medium text-white shadow-lg hover:opacity-90 transition-opacity"
              >
                <Upload className="h-4 w-4" />
                Upload Creative
              </button>
            )}
          </div>
        ) : (
          <Card>
            <CardContent className="p-0">
              <table className="w-full">
                <thead>
                  <tr className="border-b">
                    <th className="p-4 text-left">
                      <input 
                        type="checkbox" 
                        className="rounded"
                        checked={selectedCreatives.length > 0 && selectedCreatives.length === filteredCreatives.length}
                        onChange={toggleSelectAll}
                      />
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
                            <div className="text-xs text-muted-foreground">{creative.fileSize || 'Unknown size'}</div>
                          </div>
                        </div>
                      </td>
                      <td className="p-4">
                        <span className="capitalize">{creative.type}</span>
                      </td>
                      <td className="p-4">
                        {creative.width && creative.height ? `${creative.width}x${creative.height}` : creative.dimensions}
                      </td>
                      <td className="p-4">{getStatusBadge(creative.status)}</td>
                      <td className="p-4 text-right">{formatNumber(creative.performanceMetrics?.impressions || creative.impressions || 0)}</td>
                      <td className="p-4 text-right">{(creative.performanceMetrics?.ctr || creative.ctr || 0).toFixed(1)}%</td>
                      <td className="p-4">{creative.campaigns || 0}</td>
                      <td className="p-4">
                        <div className="flex items-center gap-1">
                          <button 
                            onClick={() => setShowPreview(creative)}
                            className="p-1.5 rounded hover:bg-muted"
                          >
                            <Eye className="h-4 w-4" />
                          </button>
                          <button 
                            onClick={() => handleEditCreative(creative)}
                            className="p-1.5 rounded hover:bg-muted"
                          >
                            <Edit className="h-4 w-4" />
                          </button>
                          <button 
                            onClick={() => handleDuplicateCreative(creative)}
                            className="p-1.5 rounded hover:bg-muted"
                            title="Duplicate creative"
                          >
                            <Copy className="h-4 w-4" />
                          </button>
                          <button 
                            onClick={() => { setDeletingCreative(creative); setShowDeleteDialog(true); }}
                            className="p-1.5 rounded hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20"
                          >
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
        )
      ) : null}

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
                className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors ${
                  selectedFiles.length > 0 
                    ? 'border-green-500 bg-green-50 dark:bg-green-900/10' 
                    : 'hover:border-blue-500 hover:bg-blue-50/50 dark:hover:bg-blue-900/10'
                }`}
                onClick={() => fileInputRef.current?.click()}
              >
                <Upload className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                {selectedFiles.length > 0 ? (
                  <div>
                    <p className="font-medium text-green-600 dark:text-green-400">
                      {selectedFiles.length} file{selectedFiles.length > 1 ? 's' : ''} selected
                    </p>
                    <p className="text-sm text-muted-foreground mt-1">
                      {selectedFiles.map(f => f.name).join(', ')}
                    </p>
                  </div>
                ) : (
                  <div>
                    <p className="font-medium">Drop files here or click to upload</p>
                    <p className="text-sm text-muted-foreground mt-1">
                      Supports JPG, PNG, GIF, MP4, MOV, HTML5, ZIP
                    </p>
                  </div>
                )}
                <input
                  ref={fileInputRef}
                  type="file"
                  className="hidden"
                  multiple
                  accept="image/*,video/*,.html,.zip"
                  onChange={handleFileSelect}
                />
              </div>

              {/* Upload Error */}
              {uploadError && (
                <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3">
                  <p className="text-red-700 dark:text-red-400 text-sm">{uploadError}</p>
                </div>
              )}

              {/* Creative Details */}
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium mb-1">Creative Name *</label>
                  <input
                    type="text"
                    placeholder="e.g., Summer Sale Banner"
                    value={uploadForm.name}
                    onChange={(e) => setUploadForm(prev => ({ ...prev, name: e.target.value }))}
                    className="w-full rounded-lg border px-3 py-2"
                    disabled={uploading}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Format</label>
                  <select 
                    value={uploadForm.format}
                    onChange={(e) => setUploadForm(prev => ({ ...prev, format: e.target.value }))}
                    className="w-full rounded-lg border px-3 py-2"
                    disabled={uploading}
                  >
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
                  value={uploadForm.tags}
                  onChange={(e) => setUploadForm(prev => ({ ...prev, tags: e.target.value }))}
                  className="w-full rounded-lg border px-3 py-2"
                  disabled={uploading}
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Destination URL</label>
                <input
                  type="url"
                  placeholder="https://example.com/landing-page"
                  value={uploadForm.destinationUrl}
                  onChange={(e) => setUploadForm(prev => ({ ...prev, destinationUrl: e.target.value }))}
                  className="w-full rounded-lg border px-3 py-2"
                  disabled={uploading}
                />
              </div>

              <div>
                <label className="flex items-center gap-2">
                  <input 
                    type="checkbox" 
                    checked={uploadForm.autoOptimize}
                    onChange={(e) => setUploadForm(prev => ({ ...prev, autoOptimize: e.target.checked }))}
                    className="rounded"
                    disabled={uploading}
                  />
                  <span className="text-sm">Auto-optimize for different placements</span>
                </label>
              </div>

              {/* Upload Progress */}
              {uploading && (
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>Uploading...</span>
                    <span>{uploadProgress}%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2 dark:bg-gray-700">
                    <div 
                      className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                      style={{ width: `${uploadProgress}%` }}
                    ></div>
                  </div>
                </div>
              )}
            </div>

            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowUploadModal(false)}
                className="rounded-lg border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors"
                disabled={uploading}
              >
                Cancel
              </button>
              <button
                onClick={handleUpload}
                disabled={uploading || selectedFiles.length === 0}
                className="rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 px-4 py-2 text-sm font-medium text-white hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {uploading ? 'Uploading...' : 'Upload Creative'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Creative Modal */}
      {showEditModal && editingCreative && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-2xl rounded-lg bg-white p-6 shadow-xl dark:bg-gray-900">
            <h2 className="text-xl font-bold mb-4">Edit Creative</h2>
            
            <div className="space-y-4">
              {/* Edit Error */}
              {editError && (
                <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3">
                  <p className="text-red-700 dark:text-red-400 text-sm">{editError}</p>
                </div>
              )}

              {/* Creative Details */}
              <div>
                <label className="block text-sm font-medium mb-1">Creative Name *</label>
                <input
                  type="text"
                  placeholder="e.g., Summer Sale Banner"
                  value={editForm.name}
                  onChange={(e) => setEditForm(prev => ({ ...prev, name: e.target.value }))}
                  className="w-full rounded-lg border px-3 py-2"
                  disabled={updating}
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Tags</label>
                <input
                  type="text"
                  placeholder="e.g., summer, sale, promo (comma separated)"
                  value={editForm.tags}
                  onChange={(e) => setEditForm(prev => ({ ...prev, tags: e.target.value }))}
                  className="w-full rounded-lg border px-3 py-2"
                  disabled={updating}
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Status</label>
                <select 
                  value={editForm.status}
                  onChange={(e) => setEditForm(prev => ({ ...prev, status: e.target.value as Creative['status'] }))}
                  className="w-full rounded-lg border px-3 py-2"
                  disabled={updating}
                >
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                  <option value="draft">Draft</option>
                  <option value="archived">Archived</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Description</label>
                <textarea
                  placeholder="Optional description of the creative..."
                  value={editForm.description}
                  onChange={(e) => setEditForm(prev => ({ ...prev, description: e.target.value }))}
                  className="w-full rounded-lg border px-3 py-2 min-h-[80px] resize-none"
                  disabled={updating}
                />
              </div>

              {/* Update Progress */}
              {updating && (
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>Updating...</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2 dark:bg-gray-700">
                    <div className="bg-blue-600 h-2 rounded-full animate-pulse"></div>
                  </div>
                </div>
              )}
            </div>

            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowEditModal(false)}
                className="rounded-lg border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors"
                disabled={updating}
              >
                Cancel
              </button>
              <button
                onClick={handleUpdateCreative}
                disabled={updating || !editForm.name.trim()}
                className="rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 px-4 py-2 text-sm font-medium text-white hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {updating ? 'Updating...' : 'Update Creative'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {showDeleteDialog && deletingCreative && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl dark:bg-gray-900">
            <div className="flex items-center gap-3 mb-4">
              <div className="rounded-full bg-red-100 p-2 dark:bg-red-900/20">
                <Trash2 className="h-5 w-5 text-red-600" />
              </div>
              <div>
                <h3 className="text-lg font-semibold">Delete Creative</h3>
                <p className="text-sm text-muted-foreground">This action cannot be undone</p>
              </div>
            </div>

            {/* Delete Error */}
            {deleteError && (
              <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3 mb-4">
                <p className="text-red-700 dark:text-red-400 text-sm">{deleteError}</p>
              </div>
            )}

            <p className="text-sm text-muted-foreground mb-6">
              Are you sure you want to delete <strong>"{deletingCreative.name}"</strong>? 
              This will permanently remove the creative and all associated data.
            </p>

            {/* Delete Progress */}
            {deleting && (
              <div className="space-y-2 mb-4">
                <div className="flex justify-between text-sm">
                  <span>Deleting...</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2 dark:bg-gray-700">
                  <div className="bg-red-600 h-2 rounded-full animate-pulse"></div>
                </div>
              </div>
            )}

            <div className="flex justify-end gap-3">
              <button
                onClick={() => setShowDeleteDialog(false)}
                className="rounded-lg border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors"
                disabled={deleting}
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteCreative}
                disabled={deleting}
                className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {deleting ? 'Deleting...' : 'Delete Creative'}
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
                    <p className="text-sm text-gray-400">
                      {showPreview.width && showPreview.height ? `${showPreview.width}x${showPreview.height}` : showPreview.dimensions}
                    </p>
                  </div>
                )}
              </div>
              
              {/* Details */}
              <div className="p-6">
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h3 className="text-xl font-bold">{showPreview.name}</h3>
                    <p className="text-muted-foreground">{showPreview.format} • {showPreview.width && showPreview.height ? `${showPreview.width}x${showPreview.height}` : showPreview.dimensions}</p>
                  </div>
                  {getStatusBadge(showPreview.status)}
                </div>
                
                <div className="grid grid-cols-4 gap-4 mb-4">
                  <div className="text-center p-3 rounded-lg bg-muted/50">
                    <div className="text-2xl font-bold">{formatNumber(showPreview.performanceMetrics?.impressions || showPreview.impressions || 0)}</div>
                    <div className="text-sm text-muted-foreground">Impressions</div>
                  </div>
                  <div className="text-center p-3 rounded-lg bg-muted/50">
                    <div className="text-2xl font-bold">{formatNumber(showPreview.performanceMetrics?.clicks || showPreview.clicks || 0)}</div>
                    <div className="text-sm text-muted-foreground">Clicks</div>
                  </div>
                  <div className="text-center p-3 rounded-lg bg-muted/50">
                    <div className="text-2xl font-bold">{(showPreview.performanceMetrics?.ctr || showPreview.ctr || 0).toFixed(1)}%</div>
                    <div className="text-sm text-muted-foreground">CTR</div>
                  </div>
                  <div className="text-center p-3 rounded-lg bg-muted/50">
                    <div className="text-2xl font-bold">{showPreview.campaigns || 0}</div>
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
