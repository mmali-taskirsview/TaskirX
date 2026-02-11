'use client';

import React, { useState, useEffect } from 'react';
import { Plus, Pencil, Trash2, Play, Pause, Eye, RefreshCw, ChevronDown, ChevronUp, Globe, Users, Smartphone, Target, Zap } from 'lucide-react';

interface Campaign {
  id: string;
  name: string;
  type: string;
  vertical?: string;
  adFormat: string;
  pricingModel: string;
  status: string;
  budget: number;
  spent: number;
  impressions: number;
  clicks: number;
  conversions: number;
  ctr: number;
  cpc: number;
  startDate: string;
  endDate: string;
  createdAt: string;
}

// Ad Format Categories (Unified with Backend Definitions)
const AD_FORMATS = {
  'Display & Mobile': [
    { value: 'banner', label: 'Banner Ads', desc: 'Standard Display (Leaderboard, Med Rec)' },
    { value: 'interstitial', label: 'Interstitial', desc: 'Full-screen overlay ads' },
    { value: 'native', label: 'Native Ads', desc: 'Seamlessly matches app/site content' },
    { value: 'rich_media', label: 'Rich Media', desc: 'Expandable, Lightbox, Push-down' },
  ],
  'Video & CTV': [
    { value: 'video_instream', label: 'In-Stream Video', desc: 'Pre/Mid/Post-roll' },
    { value: 'video_outstream', label: 'Out-Stream Video', desc: 'In-banner or In-article video' },
    { value: 'ctv', label: 'Connected TV (CTV)', desc: 'OTT Apps, Smart TV, Consoles' },
    { value: 'rewarded', label: 'Rewarded Video', desc: 'User opts-in for rewards (Gaming)' },
  ],
  'Audio & Emerging': [
    { value: 'audio_digital', label: 'Digital Audio', desc: 'Spotify, Podcasts, Internet Radio' },
    { value: 'audio_programmatic', label: 'Programmatic Audio', desc: 'DAI, Voice Assistant Ads' },
    { value: 'in_game', label: 'In-Game Ads', desc: 'Billboards inside 3D game worlds' },
    { value: 'vr_ar', label: 'VR / AR Ads', desc: 'Immersive ad experiences' },
  ],
  'Performance & Special': [
    { value: 'push', label: 'Push Notifications', desc: 'Web & Mobile Push' },
    { value: 'popunder', label: 'Popunders', desc: 'Background window ads' },
    { value: 'playable', label: 'Playable Ads', desc: 'Interactive demo before install' },
    { value: 'dco', label: 'DCO', desc: 'Dynamic Creative Optimization' },
  ],
  // Legacy/Specific Support
  'OEM & Social': [
    { value: 'ipush', label: 'iPUSH/System Push', desc: 'OEM: OPPO, VIVO, Xiaomi' },
    { value: 'in_feed', label: 'Social In-Feed', desc: 'Facebook/TikTok style feeds' },
    { value: 'stories', label: 'Stories Ads', desc: 'Vertical video format' },
  ]
};

// Pricing Models
const PRICING_MODELS = [
  { value: 'cpm', label: 'CPM', desc: 'Cost Per Mille (1,000 impressions)' },
  { value: 'cpc', label: 'CPC', desc: 'Cost Per Click' },
  { value: 'cpa', label: 'CPA', desc: 'Cost Per Action' },
  { value: 'cpi', label: 'CPI', desc: 'Cost Per Install' },
  { value: 'cps', label: 'CPS', desc: 'Cost Per Sale' },
  { value: 'cpr', label: 'CPR', desc: 'Cost Per Registration' },
  { value: 'cpv', label: 'CPV', desc: 'Cost Per View (Video/CTV)' },
  { value: 'cpcv', label: 'CPCV', desc: 'Cost Per Completed View (CTV/OTT)' },
];

// Industry Verticals
const INDUSTRY_VERTICALS = {
  'Gaming & Entertainment': [
    'Mobile Games', 'Console Games', 'Esports', 'Streaming Services', 'Music'
  ],
  'Finance & Business': [
    'Personal Finance', 'B2B SaaS', 'Investments', 'Banking', 'Insurance'
  ],
  'E-commerce & Retail': [
    'Fashion', 'Electronics', 'Health & Beauty', 'Home & Garden', 'Direct-to-Consumer'
  ],
  'Health & Lifestyle': [
    'Fitness', 'Wellness', 'Travel', 'Education', 'Real Estate'
  ],
  'Tech & Software': [
    'Mobile Apps', 'Cybersecurity', 'Cloud Services', 'Hardware', 'Developer Tools'
  ]
};

// Campaign Objectives
const CAMPAIGN_OBJECTIVES = [
  { value: 'app_installs', label: 'App Installs (CPI)', desc: 'Drive app downloads', icon: '📱' },
  { value: 'website_conversions', label: 'Website Conversions (CPC/CPA)', desc: 'Drive website actions', icon: '🌐' },
  { value: 'brand_awareness', label: 'Brand Awareness (CPM)', desc: 'Maximize reach', icon: '📢' },
  { value: 'video_views', label: 'Video Views (CPV)', desc: 'Drive video engagement', icon: '🎬' },
  { value: 'lead_generation', label: 'Lead Generation', desc: 'Collect leads', icon: '📋' },
];

// Bid Strategies
const BID_STRATEGIES = [
  { value: 'maximize_installs', label: 'Maximize Installs', desc: 'Get most installs within budget' },
  { value: 'target_cpa', label: 'Target CPA', desc: 'Optimize for target cost per action' },
  { value: 'lowest_cost', label: 'Lowest Cost', desc: 'Get lowest cost per result' },
  { value: 'manual', label: 'Manual Bidding', desc: 'Set your own bid price' },
];

// Geographic Targeting Options
const COUNTRIES = [
  { code: 'US', name: 'United States', region: 'North America' },
  { code: 'CA', name: 'Canada', region: 'North America' },
  { code: 'GB', name: 'United Kingdom', region: 'Europe' },
  { code: 'DE', name: 'Germany', region: 'Europe' },
  { code: 'FR', name: 'France', region: 'Europe' },
  { code: 'ID', name: 'Indonesia', region: 'Southeast Asia' },
  { code: 'PH', name: 'Philippines', region: 'Southeast Asia' },
  { code: 'TH', name: 'Thailand', region: 'Southeast Asia' },
  { code: 'VN', name: 'Vietnam', region: 'Southeast Asia' },
  { code: 'MY', name: 'Malaysia', region: 'Southeast Asia' },
  { code: 'SG', name: 'Singapore', region: 'Southeast Asia' },
  { code: 'IN', name: 'India', region: 'South Asia' },
  { code: 'JP', name: 'Japan', region: 'East Asia' },
  { code: 'KR', name: 'South Korea', region: 'East Asia' },
  { code: 'CN', name: 'China', region: 'East Asia' },
  { code: 'BR', name: 'Brazil', region: 'South America' },
  { code: 'MX', name: 'Mexico', region: 'North America' },
  { code: 'AU', name: 'Australia', region: 'Oceania' },
];

// Demographics
const AGE_RANGES = [
  { value: '13-17', label: '13-17' },
  { value: '18-24', label: '18-24' },
  { value: '25-34', label: '25-34' },
  { value: '35-44', label: '35-44' },
  { value: '45-54', label: '45-54' },
  { value: '55-64', label: '55-64' },
  { value: '65+', label: '65+' },
];

const GENDERS = [
  { value: 'all', label: 'All Genders' },
  { value: 'male', label: 'Male' },
  { value: 'female', label: 'Female' },
];

const INCOME_BRACKETS = [
  { value: 'low', label: 'Low Income' },
  { value: 'medium', label: 'Medium Income' },
  { value: 'high', label: 'High Income' },
  { value: 'premium', label: 'Premium/Affluent' },
];

// Device Targeting
const DEVICE_TYPES = [
  { value: 'mobile', label: 'Mobile', icon: '📱' },
  { value: 'tablet', label: 'Tablet', icon: '📲' },
  { value: 'desktop', label: 'Desktop', icon: '💻' },
  { value: 'ctv', label: 'Connected TV', icon: '📺' },
];

const OS_OPTIONS = [
  { value: 'android', label: 'Android' },
  { value: 'ios', label: 'iOS' },
  { value: 'windows', label: 'Windows' },
  { value: 'macos', label: 'macOS' },
];

// OEM Targeting (OPPO, VIVO, Xiaomi, etc.)
const OEM_BRANDS = [
  { value: 'oppo', label: 'OPPO', premium: true },
  { value: 'vivo', label: 'VIVO', premium: true },
  { value: 'xiaomi', label: 'Xiaomi', premium: false },
  { value: 'samsung', label: 'Samsung', premium: false },
  { value: 'huawei', label: 'Huawei', premium: false },
  { value: 'realme', label: 'Realme', premium: false },
  { value: 'oneplus', label: 'OnePlus', premium: true },
];

const CARRIERS = [
  { value: 'any', label: 'Any Carrier' },
  { value: 'wifi', label: 'WiFi Only' },
  { value: '5g', label: '5G Networks' },
  { value: '4g', label: '4G/LTE Networks' },
];

// Behavioral Targeting
const INTEREST_CATEGORIES = [
  { value: 'gaming', label: 'Gaming', icon: '🎮' },
  { value: 'ecommerce', label: 'E-commerce/Shopping', icon: '🛒' },
  { value: 'finance', label: 'Finance/Banking', icon: '💰' },
  { value: 'entertainment', label: 'Entertainment', icon: '🎬' },
  { value: 'news', label: 'News & Media', icon: '📰' },
  { value: 'social', label: 'Social Media', icon: '💬' },
  { value: 'health', label: 'Health & Fitness', icon: '💪' },
  { value: 'travel', label: 'Travel', icon: '✈️' },
  { value: 'food', label: 'Food & Dining', icon: '🍔' },
  { value: 'education', label: 'Education', icon: '📚' },
  { value: 'sports', label: 'Sports', icon: '⚽' },
  { value: 'technology', label: 'Technology', icon: '💻' },
];

// Use Next.js API routes for proxy
const API_BASE = '/api';

export default function CampaignsPage() {
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingCampaign, setEditingCampaign] = useState<Campaign | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    type: 'banner',
    vertical: 'Mobile Games',
    pricingModel: 'cpm',
    budget: 1000,
    dailyBudget: 100,
    bidPrice: 0.5,
    bidStrategy: 'lowest_cost',
    objective: 'app_installs',
    // Geographic Targeting
    targetingCountries: ['US'],
    // Demographic Targeting
    ageRanges: ['18-24', '25-34', '35-44'],
    gender: 'all',
    incomeBrackets: [] as string[],
    // Device Targeting
    deviceTypes: ['mobile', 'desktop'],
    operatingSystems: ['android', 'ios'],
    // OEM Targeting
    oemBrands: [] as string[],
    carrierTargeting: 'any',
    batteryLevel: 20,
    // Behavioral Targeting
    interests: [] as string[],
    // Schedule
    startDate: '',
    endDate: '',
    // Optimization
    autoOptimize: true,
  });
  
  const [showAdvancedTargeting, setShowAdvancedTargeting] = useState(false);
  const [activeTargetingTab, setActiveTargetingTab] = useState('geographic');

  const fetchCampaigns = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(`${API_BASE}/campaigns`);
      
      if (!response.ok) {
        throw new Error(`Failed to fetch campaigns: ${response.status}`);
      }
      
      const data = await response.json();
      setCampaigns(Array.isArray(data) ? data : data.data || []);
    } catch (err) {
      console.error('Error fetching campaigns:', err);
      setError(err instanceof Error ? err.message : 'Failed to load campaigns');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCampaigns();
  }, []);

  const handleCreateCampaign = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await fetch(`${API_BASE}/campaigns`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          ...formData,
          status: 'draft'
        })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || `Failed to create campaign: ${response.status}`);
      }

      await fetchCampaigns();
      setShowCreateModal(false);
      resetForm();
    } catch (err) {
      console.error('Error creating campaign:', err);
      alert(err instanceof Error ? err.message : 'Failed to create campaign');
    }
  };

  const handleUpdateCampaign = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingCampaign) return;

    try {
      const response = await fetch(`${API_BASE}/campaigns/${editingCampaign.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData)
      });

      if (!response.ok) {
        throw new Error(`Failed to update campaign: ${response.status}`);
      }

      await fetchCampaigns();
      setEditingCampaign(null);
      resetForm();
    } catch (err) {
      console.error('Error updating campaign:', err);
      alert(err instanceof Error ? err.message : 'Failed to update campaign');
    }
  };

  const handleDeleteCampaign = async (id: string) => {
    if (!confirm('Are you sure you want to delete this campaign?')) return;

    try {
      const response = await fetch(`${API_BASE}/campaigns/${id}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        throw new Error(`Failed to delete campaign: ${response.status}`);
      }

      await fetchCampaigns();
    } catch (err) {
      console.error('Error deleting campaign:', err);
      alert(err instanceof Error ? err.message : 'Failed to delete campaign');
    }
  };

  const handleStatusChange = async (id: string, newStatus: string) => {
    try {
      const response = await fetch(`${API_BASE}/campaigns/${id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ status: newStatus })
      });

      if (!response.ok) {
        throw new Error(`Failed to update status: ${response.status}`);
      }

      await fetchCampaigns();
    } catch (err) {
      console.error('Error updating status:', err);
      alert(err instanceof Error ? err.message : 'Failed to update campaign status');
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      type: 'banner',
      vertical: 'Mobile Games',
      pricingModel: 'cpm',
      budget: 1000,
      dailyBudget: 100,
      bidPrice: 0.5,
      bidStrategy: 'lowest_cost',
      objective: 'app_installs',
      targetingCountries: ['US'],
      ageRanges: ['18-24', '25-34', '35-44'],
      gender: 'all',
      incomeBrackets: [],
      deviceTypes: ['mobile', 'desktop'],
      operatingSystems: ['android', 'ios'],
      oemBrands: [],
      carrierTargeting: 'any',
      batteryLevel: 20,
      interests: [],
      startDate: '',
      endDate: '',
      autoOptimize: true,
    });
    setShowAdvancedTargeting(false);
  };

  // Campaign Quality Score Calculator (A/B/C/D grading)
  const calculateQualityScore = (campaign: Campaign): { grade: string; score: number; color: string; bgColor: string } => {
    let score = 0;
    
    // CTR scoring (max 30 points)
    const ctr = campaign.clicks && campaign.impressions ? (campaign.clicks / campaign.impressions) * 100 : 0;
    if (ctr >= 2.0) score += 30;
    else if (ctr >= 1.0) score += 22;
    else if (ctr >= 0.5) score += 15;
    else if (ctr > 0) score += 8;
    
    // Budget utilization scoring (max 25 points)
    const utilization = campaign.spent && campaign.budget ? (campaign.spent / campaign.budget) * 100 : 0;
    if (utilization >= 70 && utilization <= 100) score += 25;
    else if (utilization >= 50 && utilization < 70) score += 18;
    else if (utilization >= 30 && utilization < 50) score += 12;
    else if (utilization > 0) score += 5;
    
    // Impressions scoring (max 20 points)
    const impressions = campaign.impressions || 0;
    if (impressions >= 100000) score += 20;
    else if (impressions >= 50000) score += 15;
    else if (impressions >= 10000) score += 10;
    else if (impressions > 0) score += 5;
    
    // Clicks scoring (max 15 points)
    const clicks = campaign.clicks || 0;
    if (clicks >= 5000) score += 15;
    else if (clicks >= 1000) score += 11;
    else if (clicks >= 100) score += 7;
    else if (clicks > 0) score += 3;
    
    // Status bonus (max 10 points)
    if (campaign.status === 'active') score += 10;
    else if (campaign.status === 'paused') score += 5;
    
    // Calculate grade based on score
    if (score >= 85) return { grade: 'A', score, color: 'text-emerald-700', bgColor: 'bg-emerald-100' };
    if (score >= 70) return { grade: 'B', score, color: 'text-blue-700', bgColor: 'bg-blue-100' };
    if (score >= 50) return { grade: 'C', score, color: 'text-amber-700', bgColor: 'bg-amber-100' };
    return { grade: 'D', score, color: 'text-red-700', bgColor: 'bg-red-100' };
  };

  const openEditModal = (campaign: Campaign) => {
    setEditingCampaign(campaign);
    setFormData({
      name: campaign.name,
      type: campaign.type || 'banner',
      vertical: campaign.vertical || 'Mobile Games',
      pricingModel: campaign.pricingModel || 'cpm',
      budget: campaign.budget,
      dailyBudget: 100,
      bidPrice: 0.5,
      bidStrategy: 'lowest_cost',
      objective: 'app_installs',
      targetingCountries: ['US'],
      ageRanges: ['18-24', '25-34', '35-44'],
      gender: 'all',
      incomeBrackets: [],
      deviceTypes: ['mobile', 'desktop'],
      operatingSystems: ['android', 'ios'],
      oemBrands: [],
      carrierTargeting: 'any',
      batteryLevel: 20,
      interests: [],
      startDate: campaign.startDate?.split('T')[0] || '',
      endDate: campaign.endDate?.split('T')[0] || '',
      autoOptimize: true,
    });
  };

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'active': return 'bg-green-100 text-green-800';
      case 'paused': return 'bg-yellow-100 text-yellow-800';
      case 'draft': return 'bg-gray-100 text-gray-800';
      case 'completed': return 'bg-blue-100 text-blue-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value || 0);
  };

  const formatNumber = (value: number) => {
    return new Intl.NumberFormat('en-US').format(value || 0);
  };

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Campaigns</h1>
          <p className="text-gray-600">Manage your advertising campaigns</p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={fetchCampaigns}
            className="flex items-center gap-2 px-4 py-2 text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
          >
            <RefreshCw className="w-4 h-4" />
            Refresh
          </button>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-4 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700"
          >
            <Plus className="w-4 h-4" />
            Create Campaign
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded-lg">
          {error}
          <button onClick={fetchCampaigns} className="ml-4 underline">Retry</button>
        </div>
      )}

      {loading ? (
        <div className="flex justify-center items-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      ) : campaigns.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
          <div className="text-gray-500 mb-4">No campaigns yet</div>
          <button
            onClick={() => setShowCreateModal(true)}
            className="inline-flex items-center gap-2 px-4 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700"
          >
            <Plus className="w-4 h-4" />
            Create Your First Campaign
          </button>
        </div>
      ) : (
        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Campaign</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Quality</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Budget</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Spent</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Impressions</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Clicks</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">CTR</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {campaigns.map((campaign) => {
                const qualityScore = calculateQualityScore(campaign);
                return (
                <tr key={campaign.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="font-medium text-gray-900">{campaign.name}</div>
                    <div className="text-sm text-gray-500">{campaign.type}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(campaign.status)}`}>
                      {campaign.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2">
                      <span className={`px-2 py-1 text-sm font-bold rounded ${qualityScore.bgColor} ${qualityScore.color}`}>
                        {qualityScore.grade}
                      </span>
                      <span className="text-xs text-gray-500">{qualityScore.score}/100</span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {formatCurrency(campaign.budget)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {formatCurrency(campaign.spent)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {formatNumber(campaign.impressions)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {formatNumber(campaign.clicks)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {(campaign.ctr || 0).toFixed(2)}%
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <div className="flex justify-end gap-2">
                      {campaign.status === 'active' ? (
                        <button
                          onClick={() => handleStatusChange(campaign.id, 'paused')}
                          className="p-1 text-yellow-600 hover:text-yellow-800"
                          title="Pause"
                        >
                          <Pause className="w-4 h-4" />
                        </button>
                      ) : campaign.status !== 'completed' && (
                        <button
                          onClick={() => handleStatusChange(campaign.id, 'active')}
                          className="p-1 text-green-600 hover:text-green-800"
                          title="Activate"
                        >
                          <Play className="w-4 h-4" />
                        </button>
                      )}
                      <button
                        onClick={() => openEditModal(campaign)}
                        className="p-1 text-blue-600 hover:text-blue-800"
                        title="Edit"
                      >
                        <Pencil className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => handleDeleteCampaign(campaign.id)}
                        className="p-1 text-red-600 hover:text-red-800"
                        title="Delete"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              )})}
            </tbody>
          </table>
        </div>
      )}

      {/* Create/Edit Modal */}
      {(showCreateModal || editingCampaign) && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg p-6 w-full max-w-4xl mx-4 max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold mb-4">
              {editingCampaign ? 'Edit Campaign' : 'Create New Campaign'}
            </h2>
            <form onSubmit={editingCampaign ? handleUpdateCampaign : handleCreateCampaign}>
              <div className="space-y-6">
                {/* Step 1: Campaign Basics */}
                <div className="bg-gray-50 rounded-lg p-4">
                  <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                    <span className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm">1</span>
                    Campaign Basics
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Campaign Name *</label>
                      <input
                        type="text"
                        required
                        value={formData.name}
                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                        placeholder="Enter campaign name"
                      />
                    </div>
                    <div>
                         <label className="block text-sm font-medium text-gray-700 mb-1">Industry Vertical *</label>
                         <select
                           value={formData.vertical}
                           onChange={(e) => setFormData({ ...formData, vertical: e.target.value })}
                           className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                         >
                           {Object.entries(INDUSTRY_VERTICALS).map(([category, verticals]) => (
                             <optgroup key={category} label={category}>
                               {verticals.map((v) => (
                                 <option key={v} value={v}>{v}</option>
                               ))}
                             </optgroup>
                           ))}
                         </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Objective *</label>
                      <select
                        value={formData.objective}
                        onChange={(e) => setFormData({ ...formData, objective: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                      >
                        {CAMPAIGN_OBJECTIVES.map((obj) => (
                          <option key={obj.value} value={obj.value}>{obj.icon} {obj.label}</option>
                        ))}
                      </select>
                    </div>
                  </div>
                </div>

                {/* Step 2: Ad Format & Pricing */}
                <div className="bg-gray-50 rounded-lg p-4">
                  <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                    <span className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm">2</span>
                    Format & Pricing
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Ad Format *</label>
                      <select
                        value={formData.type}
                        onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                      >
                        {Object.entries(AD_FORMATS).map(([category, formats]) => (
                          <optgroup key={category} label={category}>
                            {formats.map((format) => (
                              <option key={format.value} value={format.value}>
                                {format.label}
                              </option>
                            ))}
                          </optgroup>
                        ))}
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Pricing Model *</label>
                      <select
                        value={formData.pricingModel}
                        onChange={(e) => setFormData({ ...formData, pricingModel: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                      >
                        {PRICING_MODELS.map((model) => (
                          <option key={model.value} value={model.value}>{model.label} - {model.desc}</option>
                        ))}
                      </select>
                    </div>
                  </div>
                </div>

                {/* Step 3: Budget & Bidding */}
                <div className="bg-gray-50 rounded-lg p-4">
                  <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                    <span className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm">3</span>
                    Budget & Bidding
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Total Budget (USD) *</label>
                      <input
                        type="number"
                        required
                        min="1"
                        step="0.01"
                        value={formData.budget}
                        onChange={(e) => setFormData({ ...formData, budget: parseFloat(e.target.value) })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Daily Cap (USD)</label>
                      <input
                        type="number"
                        min="1"
                        step="0.01"
                        value={formData.dailyBudget}
                        onChange={(e) => setFormData({ ...formData, dailyBudget: parseFloat(e.target.value) })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Bid Strategy</label>
                      <select
                        value={formData.bidStrategy}
                        onChange={(e) => setFormData({ ...formData, bidStrategy: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                      >
                        {BID_STRATEGIES.map((strategy) => (
                          <option key={strategy.value} value={strategy.value}>{strategy.label}</option>
                        ))}
                      </select>
                    </div>
                  </div>
                  {formData.bidStrategy === 'manual' && (
                    <div className="mt-4">
                      <label className="block text-sm font-medium text-gray-700 mb-1">Manual Bid Price (USD)</label>
                      <input
                        type="number"
                        min="0.01"
                        step="0.01"
                        value={formData.bidPrice}
                        onChange={(e) => setFormData({ ...formData, bidPrice: parseFloat(e.target.value) })}
                        className="w-full max-w-xs px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                  )}
                  <div className="mt-4 flex items-center gap-2">
                    <input
                      type="checkbox"
                      id="autoOptimize"
                      checked={formData.autoOptimize}
                      onChange={(e) => setFormData({ ...formData, autoOptimize: e.target.checked })}
                      className="w-4 h-4 text-blue-600 rounded"
                    />
                    <label htmlFor="autoOptimize" className="text-sm text-gray-700">
                      Auto-optimize budget across formats
                    </label>
                  </div>
                </div>

                {/* Step 4: Advanced Targeting */}
                <div className="bg-gray-50 rounded-lg p-4">
                  <button
                    type="button"
                    onClick={() => setShowAdvancedTargeting(!showAdvancedTargeting)}
                    className="w-full flex items-center justify-between text-lg font-semibold"
                  >
                    <span className="flex items-center gap-2">
                      <span className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm">4</span>
                      Advanced Targeting
                    </span>
                    {showAdvancedTargeting ? <ChevronUp className="w-5 h-5" /> : <ChevronDown className="w-5 h-5" />}
                  </button>
                  
                  {showAdvancedTargeting && (
                    <div className="mt-4 space-y-4">
                      {/* Targeting Tabs */}
                      <div className="flex gap-2 border-b">
                        {[
                          { id: 'geographic', label: 'Geographic', icon: Globe },
                          { id: 'demographic', label: 'Demographic', icon: Users },
                          { id: 'device', label: 'Device & OEM', icon: Smartphone },
                          { id: 'behavioral', label: 'Behavioral', icon: Target },
                        ].map((tab) => (
                          <button
                            key={tab.id}
                            type="button"
                            onClick={() => setActiveTargetingTab(tab.id)}
                            className={`flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 transition ${
                              activeTargetingTab === tab.id
                                ? 'border-blue-600 text-blue-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700'
                            }`}
                          >
                            <tab.icon className="w-4 h-4" />
                            {tab.label}
                          </button>
                        ))}
                      </div>

                      {/* Geographic Targeting */}
                      {activeTargetingTab === 'geographic' && (
                        <div className="space-y-4">
                          <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">Target Countries</label>
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-2 max-h-48 overflow-y-auto border rounded-lg p-3">
                              {COUNTRIES.map((country) => (
                                <label key={country.code} className="flex items-center gap-2 text-sm">
                                  <input
                                    type="checkbox"
                                    checked={formData.targetingCountries.includes(country.code)}
                                    onChange={(e) => {
                                      if (e.target.checked) {
                                        setFormData({ ...formData, targetingCountries: [...formData.targetingCountries, country.code] });
                                      } else {
                                        setFormData({ ...formData, targetingCountries: formData.targetingCountries.filter(c => c !== country.code) });
                                      }
                                    }}
                                    className="w-4 h-4 text-blue-600 rounded"
                                  />
                                  <span>{country.name}</span>
                                </label>
                              ))}
                            </div>
                          </div>
                        </div>
                      )}

                      {/* Demographic Targeting */}
                      {activeTargetingTab === 'demographic' && (
                        <div className="space-y-4">
                          <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">Age Ranges</label>
                            <div className="flex flex-wrap gap-2">
                              {AGE_RANGES.map((age) => (
                                <label key={age.value} className="flex items-center gap-2 px-3 py-1 border rounded-full text-sm cursor-pointer hover:bg-gray-100">
                                  <input
                                    type="checkbox"
                                    checked={formData.ageRanges.includes(age.value)}
                                    onChange={(e) => {
                                      if (e.target.checked) {
                                        setFormData({ ...formData, ageRanges: [...formData.ageRanges, age.value] });
                                      } else {
                                        setFormData({ ...formData, ageRanges: formData.ageRanges.filter(a => a !== age.value) });
                                      }
                                    }}
                                    className="w-4 h-4 text-blue-600 rounded"
                                  />
                                  {age.label}
                                </label>
                              ))}
                            </div>
                          </div>
                          <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">Gender</label>
                            <div className="flex gap-4">
                              {GENDERS.map((g) => (
                                <label key={g.value} className="flex items-center gap-2 text-sm">
                                  <input
                                    type="radio"
                                    name="gender"
                                    value={g.value}
                                    checked={formData.gender === g.value}
                                    onChange={(e) => setFormData({ ...formData, gender: e.target.value })}
                                    className="w-4 h-4 text-blue-600"
                                  />
                                  {g.label}
                                </label>
                              ))}
                            </div>
                          </div>
                          <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">Income Brackets</label>
                            <div className="flex flex-wrap gap-2">
                              {INCOME_BRACKETS.map((income) => (
                                <label key={income.value} className="flex items-center gap-2 px-3 py-1 border rounded-full text-sm cursor-pointer hover:bg-gray-100">
                                  <input
                                    type="checkbox"
                                    checked={formData.incomeBrackets.includes(income.value)}
                                    onChange={(e) => {
                                      if (e.target.checked) {
                                        setFormData({ ...formData, incomeBrackets: [...formData.incomeBrackets, income.value] });
                                      } else {
                                        setFormData({ ...formData, incomeBrackets: formData.incomeBrackets.filter(i => i !== income.value) });
                                      }
                                    }}
                                    className="w-4 h-4 text-blue-600 rounded"
                                  />
                                  {income.label}
                                </label>
                              ))}
                            </div>
                          </div>
                        </div>
                      )}

                      {/* Device & OEM Targeting */}
                      {activeTargetingTab === 'device' && (
                        <div className="space-y-4">
                          <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">Device Types</label>
                            <div className="flex flex-wrap gap-2">
                              {DEVICE_TYPES.map((device) => (
                                <label key={device.value} className="flex items-center gap-2 px-3 py-2 border rounded-lg text-sm cursor-pointer hover:bg-gray-100">
                                  <input
                                    type="checkbox"
                                    checked={formData.deviceTypes.includes(device.value)}
                                    onChange={(e) => {
                                      if (e.target.checked) {
                                        setFormData({ ...formData, deviceTypes: [...formData.deviceTypes, device.value] });
                                      } else {
                                        setFormData({ ...formData, deviceTypes: formData.deviceTypes.filter(d => d !== device.value) });
                                      }
                                    }}
                                    className="w-4 h-4 text-blue-600 rounded"
                                  />
                                  <span>{device.icon}</span>
                                  {device.label}
                                </label>
                              ))}
                            </div>
                          </div>
                          <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">Operating Systems</label>
                            <div className="flex flex-wrap gap-2">
                              {OS_OPTIONS.map((os) => (
                                <label key={os.value} className="flex items-center gap-2 px-3 py-1 border rounded-full text-sm cursor-pointer hover:bg-gray-100">
                                  <input
                                    type="checkbox"
                                    checked={formData.operatingSystems.includes(os.value)}
                                    onChange={(e) => {
                                      if (e.target.checked) {
                                        setFormData({ ...formData, operatingSystems: [...formData.operatingSystems, os.value] });
                                      } else {
                                        setFormData({ ...formData, operatingSystems: formData.operatingSystems.filter(o => o !== os.value) });
                                      }
                                    }}
                                    className="w-4 h-4 text-blue-600 rounded"
                                  />
                                  {os.label}
                                </label>
                              ))}
                            </div>
                          </div>
                          <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">OEM Brand Targeting (Premium)</label>
                            <div className="flex flex-wrap gap-2">
                              {OEM_BRANDS.map((oem) => (
                                <label key={oem.value} className={`flex items-center gap-2 px-3 py-1 border rounded-full text-sm cursor-pointer hover:bg-gray-100 ${oem.premium ? 'border-yellow-400 bg-yellow-50' : ''}`}>
                                  <input
                                    type="checkbox"
                                    checked={formData.oemBrands.includes(oem.value)}
                                    onChange={(e) => {
                                      if (e.target.checked) {
                                        setFormData({ ...formData, oemBrands: [...formData.oemBrands, oem.value] });
                                      } else {
                                        setFormData({ ...formData, oemBrands: formData.oemBrands.filter(o => o !== oem.value) });
                                      }
                                    }}
                                    className="w-4 h-4 text-blue-600 rounded"
                                  />
                                  {oem.label}
                                  {oem.premium && <span className="text-yellow-600 text-xs">★</span>}
                                </label>
                              ))}
                            </div>
                          </div>
                          <div className="grid grid-cols-2 gap-4">
                            <div>
                              <label className="block text-sm font-medium text-gray-700 mb-2">Network/Carrier</label>
                              <select
                                value={formData.carrierTargeting}
                                onChange={(e) => setFormData({ ...formData, carrierTargeting: e.target.value })}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                              >
                                {CARRIERS.map((carrier) => (
                                  <option key={carrier.value} value={carrier.value}>{carrier.label}</option>
                                ))}
                              </select>
                            </div>
                            <div>
                              <label className="block text-sm font-medium text-gray-700 mb-2">Min Battery Level: {formData.batteryLevel}%</label>
                              <input
                                type="range"
                                min="0"
                                max="100"
                                value={formData.batteryLevel}
                                onChange={(e) => setFormData({ ...formData, batteryLevel: parseInt(e.target.value) })}
                                className="w-full"
                              />
                            </div>
                          </div>
                        </div>
                      )}

                      {/* Behavioral Targeting */}
                      {activeTargetingTab === 'behavioral' && (
                        <div className="space-y-4">
                          <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">Interest Categories</label>
                            <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
                              {INTEREST_CATEGORIES.map((interest) => (
                                <label key={interest.value} className="flex items-center gap-2 px-3 py-2 border rounded-lg text-sm cursor-pointer hover:bg-gray-100">
                                  <input
                                    type="checkbox"
                                    checked={formData.interests.includes(interest.value)}
                                    onChange={(e) => {
                                      if (e.target.checked) {
                                        setFormData({ ...formData, interests: [...formData.interests, interest.value] });
                                      } else {
                                        setFormData({ ...formData, interests: formData.interests.filter(i => i !== interest.value) });
                                      }
                                    }}
                                    className="w-4 h-4 text-blue-600 rounded"
                                  />
                                  <span>{interest.icon}</span>
                                  {interest.label}
                                </label>
                              ))}
                            </div>
                          </div>
                        </div>
                      )}
                    </div>
                  )}
                </div>

                {/* Step 5: Schedule */}
                <div className="bg-gray-50 rounded-lg p-4">
                  <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                    <span className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm">5</span>
                    Schedule
                  </h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
                      <input
                        type="date"
                        value={formData.startDate}
                        onChange={(e) => setFormData({ ...formData, startDate: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">End Date</label>
                      <input
                        type="date"
                        value={formData.endDate}
                        onChange={(e) => setFormData({ ...formData, endDate: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                  </div>
                </div>
              </div>

              <div className="flex justify-end gap-3 mt-6 pt-4 border-t">
                <button
                  type="button"
                  onClick={() => {
                    setShowCreateModal(false);
                    setEditingCampaign(null);
                    resetForm();
                  }}
                  className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-6 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700 flex items-center gap-2"
                >
                  <Zap className="w-4 h-4" />
                  {editingCampaign ? 'Update Campaign' : 'Create Campaign'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
