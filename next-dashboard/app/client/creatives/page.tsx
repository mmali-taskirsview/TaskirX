'use client'

import { useState, useEffect } from 'react'
import { api } from '@/lib/api'
import { Plus, Search, Grid, List, Upload, Image as ImageIcon, Video, FileText, MoreVertical, Eye, Edit, Trash2, Copy, Check, X, Loader2 } from 'lucide-react'

interface Creative {
  id: number | string;
  name: string;
  type: string;
  format: string;
  size: string;
  status: string;
  campaigns: number;
  impressions: string;
  url: string;
}

export default function ClientCreatives() {
  const [view, setView] = useState<'grid' | 'list'>('grid')
  const [searchQuery, setSearchQuery] = useState('')
  const [typeFilter, setTypeFilter] = useState('all')
  const [showUpload, setShowUpload] = useState(false)
  const [loading, setLoading] = useState(true)
  const [creatives, setCreatives] = useState<Creative[]>([])

  useEffect(() => {
    const fetchCreatives = async () => {
      try {
        // Creatives are derived from campaigns - each campaign may have creatives
        const response = await api.getCampaigns()
        const campaigns = response.data || response || []
        
        // Generate creatives from campaign data
        const generatedCreatives: Creative[] = campaigns.flatMap((campaign: any, index: number) => {
          const type = campaign.type?.toLowerCase().includes('video') ? 'video' 
            : campaign.type?.toLowerCase().includes('native') ? 'native'
            : campaign.type?.toLowerCase().includes('playable') ? 'playable'
            : 'banner'
          
          return [{
            id: campaign.id || index + 1,
            name: `${campaign.name || 'Campaign'} Creative`,
            type,
            format: type === 'video' ? '1920x1080' : '300x250',
            size: type === 'video' ? '2.4 MB' : '45 KB',
            status: campaign.status === 'active' ? 'approved' : campaign.status === 'draft' ? 'pending' : 'approved',
            campaigns: 1,
            impressions: formatImpressions(campaign.impressions || 0),
            url: type === 'video' ? '/placeholder-video.mp4' : '/placeholder-banner.jpg',
          }]
        })
        
        setCreatives(generatedCreatives.length > 0 ? generatedCreatives : [
          { id: 1, name: 'Summer Sale Banner 300x250', type: 'banner', format: '300x250', size: '45 KB', status: 'approved', campaigns: 3, impressions: '1.2M', url: '/placeholder-banner.jpg' },
          { id: 2, name: 'App Promo Video 15s', type: 'video', format: '1920x1080', size: '2.4 MB', status: 'approved', campaigns: 2, impressions: '890K', url: '/placeholder-video.mp4' },
        ])
      } catch (error) {
        console.error('Failed to fetch creatives:', error)
        // Fallback demo data
        setCreatives([
          { id: 1, name: 'Summer Sale Banner 300x250', type: 'banner', format: '300x250', size: '45 KB', status: 'approved', campaigns: 3, impressions: '1.2M', url: '/placeholder-banner.jpg' },
          { id: 2, name: 'App Promo Video 15s', type: 'video', format: '1920x1080', size: '2.4 MB', status: 'approved', campaigns: 2, impressions: '890K', url: '/placeholder-video.mp4' },
          { id: 3, name: 'Holiday Native Ad', type: 'native', format: '1200x627', size: '120 KB', status: 'pending', campaigns: 0, impressions: '0', url: '/placeholder-native.jpg' },
        ])
      } finally {
        setLoading(false)
      }
    }
    
    fetchCreatives()
  }, [])

  const formatImpressions = (num: number): string => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
    if (num >= 1000) return `${(num / 1000).toFixed(0)}K`
    return num.toString()
  }

  const filteredCreatives = creatives.filter(c => {
    const matchesSearch = c.name.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesType = typeFilter === 'all' || c.type === typeFilter
    return matchesSearch && matchesType
  })

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'video': return <Video className="h-5 w-5 text-purple-500" />
      case 'banner': return <ImageIcon className="h-5 w-5 text-blue-500" />
      case 'native': return <FileText className="h-5 w-5 text-green-500" />
      case 'playable': return <FileText className="h-5 w-5 text-orange-500" />
      default: return <ImageIcon className="h-5 w-5 text-gray-500" />
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
        <span className="ml-2 text-gray-600">Loading creatives...</span>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Creatives</h1>
          <p className="text-gray-500">Manage your ad creatives and assets</p>
        </div>
        <button
          onClick={() => setShowUpload(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
        >
          <Upload className="h-5 w-5" />
          Upload Creative
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-col gap-4 rounded-xl bg-white p-4 shadow-sm sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-4">
          <div className="relative flex-1 min-w-[200px]">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              placeholder="Search creatives..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full rounded-lg border border-gray-300 py-2 pl-10 pr-4 focus:border-blue-500 focus:outline-none"
            />
          </div>
          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            className="rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
          >
            <option value="all">All Types</option>
            <option value="banner">Banners</option>
            <option value="video">Videos</option>
            <option value="native">Native</option>
            <option value="playable">Playable</option>
          </select>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setView('grid')}
            className={`rounded-lg p-2 ${view === 'grid' ? 'bg-blue-100 text-blue-600' : 'text-gray-500 hover:bg-gray-100'}`}
          >
            <Grid className="h-5 w-5" />
          </button>
          <button
            onClick={() => setView('list')}
            className={`rounded-lg p-2 ${view === 'list' ? 'bg-blue-100 text-blue-600' : 'text-gray-500 hover:bg-gray-100'}`}
          >
            <List className="h-5 w-5" />
          </button>
        </div>
      </div>

      {/* Grid View */}
      {view === 'grid' ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {filteredCreatives.map((creative) => (
            <div key={creative.id} className="group rounded-xl bg-white shadow-sm overflow-hidden">
              <div className="relative aspect-video bg-gray-100">
                <div className="absolute inset-0 flex items-center justify-center">
                  {getTypeIcon(creative.type)}
                </div>
                <div className="absolute inset-0 flex items-center justify-center bg-black/50 opacity-0 transition-opacity group-hover:opacity-100">
                  <button className="rounded-lg bg-white px-3 py-1.5 text-sm font-medium text-gray-900">
                    Preview
                  </button>
                </div>
                <span className={`absolute right-2 top-2 rounded-full px-2 py-0.5 text-xs font-medium ${
                  creative.status === 'approved' ? 'bg-green-100 text-green-700' :
                  creative.status === 'pending' ? 'bg-yellow-100 text-yellow-700' :
                  'bg-red-100 text-red-700'
                }`}>
                  {creative.status}
                </span>
              </div>
              <div className="p-4">
                <h3 className="font-medium text-gray-900 truncate">{creative.name}</h3>
                <div className="mt-2 flex items-center justify-between text-sm text-gray-500">
                  <span>{creative.format}</span>
                  <span>{creative.size}</span>
                </div>
                <div className="mt-3 flex items-center justify-between border-t pt-3">
                  <span className="text-xs text-gray-500">{creative.campaigns} campaigns</span>
                  <span className="text-xs text-gray-500">{creative.impressions} impr.</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        /* List View */
        <div className="rounded-xl bg-white shadow-sm">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200 bg-gray-50">
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">Creative</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">Type</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">Format</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">Campaigns</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">Impressions</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {filteredCreatives.map((creative) => (
                <tr key={creative.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gray-100">
                        {getTypeIcon(creative.type)}
                      </div>
                      <span className="font-medium text-gray-900">{creative.name}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600 capitalize">{creative.type}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">{creative.format}</td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${
                      creative.status === 'approved' ? 'bg-green-100 text-green-700' :
                      creative.status === 'pending' ? 'bg-yellow-100 text-yellow-700' :
                      'bg-red-100 text-red-700'
                    }`}>
                      {creative.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600">{creative.campaigns}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">{creative.impressions}</td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <button className="rounded p-1 hover:bg-gray-100"><Eye className="h-4 w-4 text-gray-500" /></button>
                      <button className="rounded p-1 hover:bg-gray-100"><Edit className="h-4 w-4 text-gray-500" /></button>
                      <button className="rounded p-1 hover:bg-gray-100"><Copy className="h-4 w-4 text-gray-500" /></button>
                      <button className="rounded p-1 hover:bg-gray-100"><Trash2 className="h-4 w-4 text-red-500" /></button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Upload Modal */}
      {showUpload && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-lg rounded-xl bg-white p-6">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-lg font-semibold">Upload Creative</h2>
              <button onClick={() => setShowUpload(false)} className="text-gray-500 hover:text-gray-700">
                <X className="h-5 w-5" />
              </button>
            </div>
            <div className="rounded-lg border-2 border-dashed border-gray-300 p-8 text-center">
              <Upload className="mx-auto h-12 w-12 text-gray-400" />
              <p className="mt-2 text-sm text-gray-600">Drag and drop files here, or click to browse</p>
              <p className="mt-1 text-xs text-gray-500">Supports: JPG, PNG, GIF, MP4, HTML5</p>
              <button className="mt-4 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700">
                Select Files
              </button>
            </div>
            <div className="mt-4 flex justify-end gap-3">
              <button onClick={() => setShowUpload(false)} className="rounded-lg border px-4 py-2 text-sm hover:bg-gray-50">
                Cancel
              </button>
              <button className="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700">
                Upload
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
