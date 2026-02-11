'use client';

import React, { useState, useEffect } from 'react';
import {
  Code,
  Plus,
  Search,
  Copy,
  Check,
  Eye,
  Download,
  Settings,
  Monitor,
  Smartphone,
  Globe,
  Zap,
  AlertCircle,
  CheckCircle,
  Loader2
} from 'lucide-react';
import { api } from '@/lib/api';

interface AdUnit {
  id: string;
  name: string;
  publisherId: string;
  placementId?: string;
  adFormat: string;
  sizes: string[];
  status: string;
  floorPrice: number;
  impressions: number;
  requests: number;
  revenue: number;
  createdAt: string;
}

// Generate ad tag code for an ad unit
function generateAdTagCode(adUnit: AdUnit, type: string): string {
  const size = adUnit.sizes[0] || '300x250';
  const publisherId = adUnit.publisherId;
  
  switch (type) {
    case 'javascript':
      return `<script async src="https://cdn.taskirx.com/tags/${publisherId}/${adUnit.id}.js"></script>
<div id="taskirx-ad-${adUnit.id}" data-size="${size}"></div>`;
    case 'sdk':
      return `// iOS SDK
TaskirXAds.showBanner(placementId: "${adUnit.id}", size: .banner${size.replace('x', 'x')})

// Android SDK
TaskirXAds.showBanner("${adUnit.id}", AdSize.BANNER_${size.replace('x', 'X')})`;
    case 'vast':
      return `https://ads.taskirx.com/vast/v2?tag=${adUnit.id}&pub=${publisherId}&w=${size.split('x')[0]}&h=${size.split('x')[1]}`;
    case 'amp':
      return `<amp-ad width="${size.split('x')[0]}" height="${size.split('x')[1]}"
  type="taskirx"
  data-publisher-id="${publisherId}"
  data-tag-id="${adUnit.id}">
</amp-ad>`;
    default:
      return `<!-- TaskirX Ad Tag: ${adUnit.id} -->`;
  }
}

// Convert ad unit to ad tag format
function adUnitToTag(adUnit: AdUnit) {
  const tagType = adUnit.adFormat === 'video' ? 'vast' : 
                  adUnit.adFormat === 'native' ? 'javascript' : 'javascript';
  return {
    id: adUnit.id,
    name: adUnit.name,
    placement: adUnit.placementId || 'Default',
    size: adUnit.sizes[0] || 'responsive',
    type: tagType,
    status: adUnit.status,
    impressions: adUnit.impressions,
    lastServed: adUnit.createdAt,
    code: generateAdTagCode(adUnit, tagType)
  };
}

// Define tag type
interface AdTag {
  id: string;
  name: string;
  placement: string;
  size: string;
  type: string;
  status: string;
  impressions: number;
  lastServed: string;
  code: string;
}

export default function AdTagsPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');
  const [selectedTag, setSelectedTag] = useState<AdTag | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [adTags, setAdTags] = useState<AdTag[]>([]);
  const [loading, setLoading] = useState(true);

  // Fetch ad units and convert to tags
  useEffect(() => {
    async function fetchAdUnits() {
      try {
        setLoading(true);
        const response = await api.getAdUnits();
        if (response.data) {
          const tags = response.data.map((unit: AdUnit) => adUnitToTag(unit));
          setAdTags(tags);
        }
      } catch (err) {
        console.error('Error fetching ad units:', err);
      } finally {
        setLoading(false);
      }
    }
    fetchAdUnits();
  }, []);

  const copyToClipboard = (code: string, id: string) => {
    navigator.clipboard.writeText(code);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'javascript': return <Code className="w-4 h-4" />;
      case 'sdk': return <Smartphone className="w-4 h-4" />;
      case 'vast': return <Monitor className="w-4 h-4" />;
      case 'amp': return <Zap className="w-4 h-4" />;
      default: return <Globe className="w-4 h-4" />;
    }
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'javascript': return 'bg-yellow-100 text-yellow-800';
      case 'sdk': return 'bg-blue-100 text-blue-800';
      case 'vast': return 'bg-purple-100 text-purple-800';
      case 'amp': return 'bg-orange-100 text-orange-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const filteredTags = adTags.filter((tag: AdTag) => {
    const matchesSearch = tag.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                          tag.placement.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesType = filterType === 'all' || tag.type === filterType;
    return matchesSearch && matchesType;
  });

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <Loader2 className="w-8 h-8 animate-spin text-emerald-600" />
        <span className="ml-2 text-gray-600">Loading ad tags...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Ad Tags</h1>
          <p className="text-gray-600 mt-1">Generate and manage ad tags for your placements</p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors">
          <Plus className="w-4 h-4" />
          Generate New Tag
        </button>
      </div>

      {/* Quick Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-emerald-100 rounded-lg">
              <Code className="w-5 h-5 text-emerald-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Total Tags</p>
              <p className="text-xl font-bold text-gray-900">{adTags.length}</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-green-100 rounded-lg">
              <CheckCircle className="w-5 h-5 text-green-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Active Tags</p>
              <p className="text-xl font-bold text-gray-900">{adTags.filter((t: AdTag) => t.status === 'active').length}</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-100 rounded-lg">
              <Eye className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Total Impressions</p>
              <p className="text-xl font-bold text-gray-900">1.83M</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-purple-100 rounded-lg">
              <Zap className="w-5 h-5 text-purple-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Tag Types</p>
              <p className="text-xl font-bold text-gray-900">4</p>
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-xl p-4 border border-gray-200">
        <div className="flex flex-wrap gap-4 items-center">
          <div className="relative flex-1 min-w-[200px]">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
            <input
              type="text"
              placeholder="Search ad tags..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
            />
          </div>
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            className="px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
          >
            <option value="all">All Types</option>
            <option value="javascript">JavaScript</option>
            <option value="sdk">Mobile SDK</option>
            <option value="vast">VAST/VPAID</option>
            <option value="amp">AMP</option>
          </select>
        </div>
      </div>

      {/* Tags Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {filteredTags.length === 0 ? (
          <div className="col-span-2 text-center py-12 bg-white rounded-xl border border-gray-200">
            <Code className="w-12 h-12 mx-auto mb-4 text-gray-300" />
            <p className="text-gray-500">No ad tags found</p>
            <p className="text-sm text-gray-400 mt-1">Create ad units to generate tags</p>
          </div>
        ) : (
          filteredTags.map((tag: AdTag) => (
          <div key={tag.id} className="bg-white rounded-xl border border-gray-200 overflow-hidden">
            <div className="p-4 border-b border-gray-100">
              <div className="flex items-start justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="font-semibold text-gray-900">{tag.name}</h3>
                    {tag.status === 'active' ? (
                      <span className="px-2 py-0.5 text-xs font-medium bg-green-100 text-green-700 rounded-full">
                        Active
                      </span>
                    ) : (
                      <span className="px-2 py-0.5 text-xs font-medium bg-yellow-100 text-yellow-700 rounded-full">
                        Testing
                      </span>
                    )}
                  </div>
                  <p className="text-sm text-gray-500 mt-1">{tag.placement} • {tag.size}</p>
                </div>
                <span className={`flex items-center gap-1 px-2 py-1 text-xs font-medium rounded-full ${getTypeColor(tag.type)}`}>
                  {getTypeIcon(tag.type)}
                  {tag.type.toUpperCase()}
                </span>
              </div>
              <div className="flex items-center gap-4 mt-3 text-sm text-gray-500">
                <span>{tag.impressions.toLocaleString()} impressions</span>
                <span>•</span>
                <span>Last served: {new Date(tag.lastServed).toLocaleTimeString()}</span>
              </div>
            </div>
            <div className="p-4 bg-gray-50">
              <div className="relative">
                <pre className="text-xs text-gray-700 bg-gray-900 text-gray-100 p-3 rounded-lg overflow-x-auto max-h-24">
                  <code>{tag.code}</code>
                </pre>
                <button
                  onClick={() => copyToClipboard(tag.code, tag.id)}
                  className="absolute top-2 right-2 p-1.5 bg-gray-700 hover:bg-gray-600 rounded text-white transition-colors"
                  title="Copy to clipboard"
                >
                  {copiedId === tag.id ? (
                    <Check className="w-4 h-4 text-green-400" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </button>
              </div>
              <div className="flex items-center gap-2 mt-3">
                <button
                  onClick={() => setSelectedTag(tag)}
                  className="flex-1 flex items-center justify-center gap-2 px-3 py-2 text-sm text-emerald-600 border border-emerald-200 rounded-lg hover:bg-emerald-50 transition-colors"
                >
                  <Eye className="w-4 h-4" />
                  Preview
                </button>
                <button className="flex-1 flex items-center justify-center gap-2 px-3 py-2 text-sm text-gray-600 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors">
                  <Download className="w-4 h-4" />
                  Download
                </button>
                <button className="flex items-center justify-center gap-2 px-3 py-2 text-sm text-gray-600 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors">
                  <Settings className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
          ))
        )}
      </div>

      {/* Integration Guide */}
      <div className="bg-gradient-to-r from-emerald-500 to-teal-600 rounded-xl p-6 text-white">
        <h3 className="text-lg font-semibold mb-2">Need help integrating?</h3>
        <p className="text-emerald-100 mb-4">
          Check out our comprehensive integration guides for JavaScript, Mobile SDKs, VAST/VPAID, and AMP.
        </p>
        <div className="flex flex-wrap gap-3">
          <button className="px-4 py-2 bg-white/20 hover:bg-white/30 rounded-lg transition-colors">
            JavaScript Guide
          </button>
          <button className="px-4 py-2 bg-white/20 hover:bg-white/30 rounded-lg transition-colors">
            Mobile SDK Docs
          </button>
          <button className="px-4 py-2 bg-white/20 hover:bg-white/30 rounded-lg transition-colors">
            VAST Integration
          </button>
          <button className="px-4 py-2 bg-white/20 hover:bg-white/30 rounded-lg transition-colors">
            AMP Setup
          </button>
        </div>
      </div>

      {/* Preview Modal */}
      {selectedTag && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-bold text-gray-900">Tag Preview: {selectedTag.name}</h2>
              <button
                onClick={() => setSelectedTag(null)}
                className="text-gray-400 hover:text-gray-600"
              >
                ✕
              </button>
            </div>
            
            <div className="space-y-4">
              <div className="bg-gray-100 rounded-lg p-4 flex items-center justify-center min-h-[200px]">
                <div className="text-center text-gray-500">
                  <Monitor className="w-12 h-12 mx-auto mb-2 text-gray-400" />
                  <p>Ad Preview ({selectedTag.size})</p>
                  <p className="text-sm">Sample ad would render here</p>
                </div>
              </div>
              
              <div>
                <h4 className="font-medium text-gray-900 mb-2">Full Tag Code</h4>
                <pre className="text-sm bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto">
                  <code>{selectedTag.code}</code>
                </pre>
              </div>
              
              <div className="flex gap-3">
                <button
                  onClick={() => copyToClipboard(selectedTag.code, selectedTag.id + '_modal')}
                  className="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700"
                >
                  <Copy className="w-4 h-4" />
                  Copy Code
                </button>
                <button
                  onClick={() => setSelectedTag(null)}
                  className="px-4 py-2 border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50"
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
