'use client'

import { useState } from 'react'
import { api } from '@/lib/api'
import { X, Plus, Trash2, MapPin, Loader2 } from 'lucide-react'
// Assuming Button, Input, Label are available based on file structure
import { Button } from '@/components/ui/button' 
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface CreateCampaignModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
}

export function CreateCampaignModal({ open, onOpenChange, onSuccess }: CreateCampaignModalProps) {
  const [formData, setFormData] = useState({
    name: '',
    type: 'cpm', // Pricing Model
    adFormat: 'banner', // Creative Type
    creativeUrl: '',
    creativeTitle: '',
    creativeDesc: '',
    creativeIcon: '',
    creativeCta: 'Learn More',
    budget: 1000,
    bidPrice: 0.5,
    startDate: new Date().toISOString().split('T')[0],
    endDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    vertical: 'GAMING',
    targetCountries: 'US,CA',
    geoFences: [] as { lat: string; lon: string; radius: string; name: string }[]
  })
  const [isSubmitting, setIsSubmitting] = useState(false)

  if (!open) return null

  const handleGeoFenceAdd = () => {
    setFormData(prev => ({
      ...prev,
      geoFences: [...prev.geoFences, { lat: '', lon: '', radius: '10', name: 'New Zone' }]
    }))
  }

  const handleGeoFenceRemove = (index: number) => {
    const newFences = [...formData.geoFences]
    newFences.splice(index, 1)
    setFormData(prev => ({ ...prev, geoFences: newFences }))
  }

  const handleGeoFenceChange = (index: number, field: string, value: string) => {
    const newFences = [...formData.geoFences]
    // @ts-ignore
    newFences[index][field] = value
    setFormData(prev => ({ ...prev, geoFences: newFences }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    
    try {
      // Validate inputs
      const fences = formData.geoFences.map(f => ({
        lat: parseFloat(f.lat),
        lon: parseFloat(f.lon),
        radius: parseFloat(f.radius),
        name: f.name
      })).filter(f => !isNaN(f.lat) && !isNaN(f.lon))

      const payload = {
        name: formData.name,
        type: formData.type, // CPM/CPC model
        budget: Number(formData.budget),
        bidPrice: Number(formData.bidPrice),
        startDate: formData.startDate ? new Date(formData.startDate) : undefined,
        endDate: formData.endDate ? new Date(formData.endDate) : undefined,
        vertical: formData.vertical,
        targeting: {
          countries: formData.targetCountries.split(',').map(s => s.trim()),
          geoFences: fences.length > 0 ? fences : undefined
        },
        creative: {
          type: formData.adFormat,
          url: formData.creativeUrl,
          title: formData.adFormat === 'native' ? formData.creativeTitle : undefined,
          description: formData.adFormat === 'native' ? formData.creativeDesc : undefined,
          iconUrl: formData.adFormat === 'native' ? formData.creativeIcon : undefined,
          ctaText: formData.adFormat === 'native' ? formData.creativeCta : undefined,
          width: 300, // Default width
          height: 250, // Default height
          duration: formData.adFormat === 'video' ? 30 : 0
        }
      }

      await api.createCampaign(payload)
      onSuccess()
      onOpenChange(false)
    } catch (error) {
      console.error('Failed to create campaign', error)
      alert('Failed to create campaign. Please check inputs.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div 
        className="fixed inset-0 bg-black/60 backdrop-blur-sm" 
        onClick={() => onOpenChange(false)}
      />
      <div className="relative z-10 w-full max-w-2xl rounded-xl bg-white p-6 shadow-2xl max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-semibold text-gray-900">Create New Campaign</h2>
          <button
            onClick={() => onOpenChange(false)}
            className="rounded-full p-2 hover:bg-gray-100"
          >
            <X className="h-5 w-5 text-gray-500" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>Campaign Name</Label>
              <Input
                required
                value={formData.name}
                onChange={e => setFormData({...formData, name: e.target.value})}
                placeholder="e.g. Summer Sale 2026"
              />
            </div>
            <div className="space-y-2">
              <Label>Vertical</Label>
              <select
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={formData.vertical}
                onChange={e => setFormData({...formData, vertical: e.target.value})}
              >
                <option value="GAMING">Gaming</option>
                <option value="ECOMMERCE">E-Commerce</option>
                <option value="FINANCE">Finance</option>
                <option value="TRAVEL">Travel</option>
              </select>
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>Total Budget ($)</Label>
              <Input
                type="number"
                min="100"
                value={formData.budget}
                onChange={e => setFormData({...formData, budget: Number(e.target.value)})}
              />
            </div>
            <div className="space-y-2">
              <Label>Bid Price (CPM/CPC)</Label>
              <Input
                type="number"
                step="0.01"
                value={formData.bidPrice}
                onChange={e => setFormData({...formData, bidPrice: Number(e.target.value)})}
              />
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>Start Date</Label>
              <Input
                type="date"
                value={formData.startDate}
                onChange={e => setFormData({...formData, startDate: e.target.value})}
              />
            </div>
            <div className="space-y-2">
              <Label>End Date</Label>
              <Input
                type="date"
                value={formData.endDate}
                onChange={e => setFormData({...formData, endDate: e.target.value})}
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label>Target Countries (Comma separated)</Label>
            <Input
              value={formData.targetCountries}
              onChange={e => setFormData({...formData, targetCountries: e.target.value})}
              placeholder="US, CA, UK"
            />
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
           <div className="space-y-2">
              <Label>Ad Format</Label>
              <select
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={formData.adFormat}
                onChange={e => setFormData({...formData, adFormat: e.target.value})}
              >
                <option value="banner">Display Banner</option>
                <option value="video">Video</option>
                <option value="native">Native</option>
              </select>
            </div>
            <div className="space-y-2">
              <Label>Pricing Model</Label>
              <select
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={formData.type}
                onChange={e => setFormData({...formData, type: e.target.value})}
              >
                <option value="cpm">CPM (Cost per 1k)</option>
                <option value="cpc">CPC (Cost per Click)</option>
                <option value="cpa">CPA (Cost per Action)</option>
              </select>
            </div>
          </div>

          {/* Geo-Fencing Section */}
          <div className="rounded-lg border border-gray-200 p-4 bg-gray-50">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <MapPin className="h-5 w-5 text-blue-600" />
                <h3 className="font-medium text-gray-900">Geo-Fencing (Advanced)</h3>
              </div>
              <Button 
                type="button" 
                variant="outline" 
                size="sm" 
                onClick={handleGeoFenceAdd}
              >
                <Plus className="h-4 w-4 mr-1" /> Add Location
              </Button>
            </div>
            
            {formData.geoFences.length === 0 ? (
              <p className="text-sm text-gray-500 italic">No locations added. Campaign will target entire countries.</p>
            ) : (
              <div className="space-y-3">
                {formData.geoFences.map((fence, idx) => (
                  <div key={idx} className="flex gap-2 items-end bg-white p-3 rounded shadow-sm">
                    <div className="flex-1">
                      <Label className="text-xs">Latitude</Label>
                      <Input 
                        type="number" step="any"
                        value={fence.lat}
                        onChange={e => handleGeoFenceChange(idx, 'lat', e.target.value)}
                        placeholder="40.7128"
                      />
                    </div>
                    <div className="flex-1">
                      <Label className="text-xs">Longitude</Label>
                      <Input 
                        type="number" step="any"
                        value={fence.lon}
                        onChange={e => handleGeoFenceChange(idx, 'lon', e.target.value)}
                        placeholder="-74.0060"
                      />
                    </div>
                    <div className="w-20">
                      <Label className="text-xs">Radius (km)</Label>
                      <Input 
                        type="number" 
                        value={fence.radius}
                        onChange={e => handleGeoFenceChange(idx, 'radius', e.target.value)}
                      />
                    </div>
                    <Button 
                      type="button" 
                      variant="ghost" 
                      size="icon" 
                      onClick={() => handleGeoFenceRemove(idx)}
                      className="text-red-500 hover:text-red-700 hover:bg-red-50 h-10 w-10"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Creative Section */}
          <div className="rounded-lg border border-gray-200 p-4 bg-gray-50 space-y-4">
            <h3 className="font-medium text-gray-900 border-b pb-2">Creative Assets</h3>
            
            <div className="space-y-2">
              <Label>{formData.adFormat === 'video' ? 'Video File URL' : 'Main Image URL'}</Label>
              <Input 
                value={formData.creativeUrl} 
                onChange={e => setFormData({...formData, creativeUrl: e.target.value})}
                placeholder="https://example.com/asset.jpg"
              />
            </div>

            {formData.adFormat === 'native' && (
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label>Headline</Label>
                  <Input 
                    value={formData.creativeTitle} 
                    onChange={e => setFormData({...formData, creativeTitle: e.target.value})}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Description</Label>
                  <Input 
                    value={formData.creativeDesc} 
                    onChange={e => setFormData({...formData, creativeDesc: e.target.value})}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Icon URL</Label>
                  <Input 
                    value={formData.creativeIcon} 
                    onChange={e => setFormData({...formData, creativeIcon: e.target.value})}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Call to Action</Label>
                  <Input 
                    value={formData.creativeCta} 
                    onChange={e => setFormData({...formData, creativeCta: e.target.value})}
                  />
                </div>
              </div>
            )}
          </div>

          <div className="flex justify-end gap-3 pt-4 border-t border-gray-100">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Campaign
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
