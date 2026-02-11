'use client';

import React, { useState, useEffect } from 'react';
import {
  Shield,
  Plus,
  Search,
  Eye,
  EyeOff,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Globe,
  Link as LinkIcon,
  Filter,
  Trash2,
  Edit,
  BarChart3,
  Ban,
  Tag,
  Loader2
} from 'lucide-react';
import { api } from '@/lib/api';

// Brand safety rule types from backend
type BrandSafetyRuleType = 'blocklist' | 'allowlist' | 'category_block' | 'keyword_block';
type BrandSafetyTarget = 'advertiser' | 'domain' | 'category' | 'keyword';

interface BrandSafetyRule {
  id: string;
  name: string;
  description: string;
  ruleType: BrandSafetyRuleType;
  target: BrandSafetyTarget;
  values: string[];
  publisherId: string;
  adUnitId?: string;
  isActive: boolean;
  priority: number;
  matchCount: number;
  createdAt: string;
  updatedAt: string;
}

// Content filter presets (UI only - not in DB)
const contentFilters = [
  { name: 'GARM Brand Safety', enabled: true, description: 'Global Alliance for Responsible Media standards' },
  { name: 'IAB Content Taxonomy', enabled: true, description: 'IAB Tech Lab content classification' },
  { name: 'Custom Keyword Blocking', enabled: true, description: 'Block ads containing specific keywords' },
  { name: 'Competitive Separation', enabled: false, description: 'Prevent competitor ads on same page' }
];

export default function BrandSafetyPage() {
  const [activeTab, setActiveTab] = useState('categories');
  const [showAddBlockModal, setShowAddBlockModal] = useState(false);
  const [rules, setRules] = useState<BrandSafetyRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [newDomain, setNewDomain] = useState('');
  
  // Form state for new rule
  const [newRule, setNewRule] = useState({
    name: '',
    description: '',
    ruleType: 'blocklist' as BrandSafetyRuleType,
    target: 'category' as BrandSafetyTarget,
    values: '',
    publisherId: 'pub_001'
  });

  // Fetch brand safety rules
  useEffect(() => {
    async function fetchRules() {
      try {
        setLoading(true);
        const response = await api.getBrandSafetyRules();
        setRules(response.data || []);
        setError(null);
      } catch (err) {
        console.error('Error fetching brand safety rules:', err);
        setError('Failed to load brand safety rules');
      } finally {
        setLoading(false);
      }
    }
    fetchRules();
  }, []);

  // Filter rules by target type
  const categoryRules = rules.filter(r => r.target === 'category');
  const advertiserRules = rules.filter(r => r.target === 'advertiser');
  const domainRules = rules.filter(r => r.target === 'domain');
  const keywordRules = rules.filter(r => r.target === 'keyword');

  // Calculate stats
  const totalMatchCount = rules.reduce((sum, r) => sum + Number(r.matchCount || 0), 0);
  const blockedCategories = categoryRules.filter(r => r.ruleType === 'blocklist' || r.ruleType === 'category_block');

  // Toggle rule active state
  const handleToggleRule = async (rule: BrandSafetyRule) => {
    try {
      await api.toggleBrandSafetyRule(rule.id);
      setRules(rules.map(r => 
        r.id === rule.id ? { ...r, isActive: !r.isActive } : r
      ));
    } catch (err) {
      console.error('Error toggling rule:', err);
    }
  };

  // Delete rule
  const handleDeleteRule = async (id: string) => {
    try {
      await api.deleteBrandSafetyRule(id);
      setRules(rules.filter(r => r.id !== id));
    } catch (err) {
      console.error('Error deleting rule:', err);
    }
  };

  // Create new rule
  const handleCreateRule = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await api.createBrandSafetyRule({
        name: newRule.name,
        description: newRule.description,
        ruleType: newRule.ruleType,
        target: newRule.target,
        values: newRule.values.split(',').map(v => v.trim()),
        publisherId: newRule.publisherId
      });
      setRules([...rules, response.data]);
      setShowAddBlockModal(false);
      setNewRule({
        name: '',
        description: '',
        ruleType: 'blocklist',
        target: 'category',
        values: '',
        publisherId: 'pub_001'
      });
    } catch (err) {
      console.error('Error creating rule:', err);
    }
  };

  // Add domain block
  const handleAddDomain = async () => {
    if (!newDomain) return;
    try {
      const response = await api.createBrandSafetyRule({
        name: `Block ${newDomain}`,
        description: 'Domain block rule',
        ruleType: 'blocklist',
        target: 'domain',
        values: [newDomain],
        publisherId: 'pub_001'
      });
      setRules([...rules, response.data]);
      setNewDomain('');
    } catch (err) {
      console.error('Error adding domain:', err);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <Loader2 className="w-8 h-8 animate-spin text-emerald-600" />
        <span className="ml-2 text-gray-600">Loading brand safety rules...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Brand Safety</h1>
          <p className="text-gray-600 mt-1">Control what ads appear on your properties</p>
        </div>
        <button
          onClick={() => setShowAddBlockModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
        >
          <Plus className="w-4 h-4" />
          Add Block Rule
        </button>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-red-100 rounded-lg">
              <Ban className="w-5 h-5 text-red-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Blocked Categories</p>
              <p className="text-xl font-bold text-gray-900">{blockedCategories.length}</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-orange-100 rounded-lg">
              <XCircle className="w-5 h-5 text-orange-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Blocked Advertisers</p>
              <p className="text-xl font-bold text-gray-900">{advertiserRules.length}</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-purple-100 rounded-lg">
              <Globe className="w-5 h-5 text-purple-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Blocked Domains</p>
              <p className="text-xl font-bold text-gray-900">{domainRules.length}</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl p-4 border border-gray-200">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-100 rounded-lg">
              <BarChart3 className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-sm text-gray-600">Total Matches</p>
              <p className="text-xl font-bold text-gray-900">{totalMatchCount > 1000 ? `${(totalMatchCount / 1000).toFixed(0)}K` : totalMatchCount}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="flex gap-8">
          {['categories', 'advertisers', 'domains', 'filters'].map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`pb-4 px-1 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab
                  ? 'border-emerald-500 text-emerald-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              {tab.charAt(0).toUpperCase() + tab.slice(1)}
            </button>
          ))}
        </nav>
      </div>

      {/* Categories Tab */}
      {activeTab === 'categories' && (
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div className="p-4 border-b border-gray-100">
            <h3 className="font-semibold text-gray-900">Content Categories</h3>
            <p className="text-sm text-gray-500 mt-1">Block entire categories of advertising content</p>
          </div>
          {categoryRules.length === 0 ? (
            <div className="p-8 text-center text-gray-500">
              <Ban className="w-12 h-12 mx-auto mb-4 text-gray-300" />
              <p>No category rules configured</p>
              <button 
                onClick={() => setShowAddBlockModal(true)}
                className="mt-4 text-emerald-600 hover:text-emerald-700"
              >
                Add your first category block rule
              </button>
            </div>
          ) : (
            <div className="divide-y divide-gray-100">
              {categoryRules.map((rule) => (
                <div key={rule.id} className="p-4 flex items-center justify-between hover:bg-gray-50">
                  <div className="flex items-center gap-4">
                    <button
                      onClick={() => handleToggleRule(rule)}
                      className={`relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                        rule.isActive ? 'bg-red-500' : 'bg-gray-200'
                      }`}
                    >
                      <span
                        className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
                          rule.isActive ? 'translate-x-5' : 'translate-x-0'
                        }`}
                      />
                    </button>
                    <div>
                      <p className="font-medium text-gray-900">{rule.name}</p>
                      <p className="text-sm text-gray-500">
                        {rule.values.join(', ')} • {Number(rule.matchCount)} matches
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {rule.isActive ? (
                      <span className="flex items-center gap-1 text-sm text-red-600">
                        <Ban className="w-4 h-4" />
                        Blocking
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-sm text-gray-500">
                        <CheckCircle className="w-4 h-4" />
                        Inactive
                      </span>
                    )}
                    <button 
                      onClick={() => handleDeleteRule(rule.id)}
                      className="p-1 text-gray-400 hover:text-red-600"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Advertisers Tab */}
      {activeTab === 'advertisers' && (
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div className="p-4 border-b border-gray-100">
            <h3 className="font-semibold text-gray-900">Blocked Advertisers</h3>
            <p className="text-sm text-gray-500 mt-1">Specific advertisers that cannot serve ads on your inventory</p>
          </div>
          {advertiserRules.length === 0 ? (
            <div className="p-8 text-center text-gray-500">
              <XCircle className="w-12 h-12 mx-auto mb-4 text-gray-300" />
              <p>No advertiser rules configured</p>
              <button 
                onClick={() => setShowAddBlockModal(true)}
                className="mt-4 text-emerald-600 hover:text-emerald-700"
              >
                Add your first advertiser block
              </button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Rule Name</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Blocked Values</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Matches</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {advertiserRules.map((rule) => (
                    <tr key={rule.id} className="hover:bg-gray-50">
                      <td className="px-4 py-4 text-sm font-medium text-gray-900">{rule.name}</td>
                      <td className="px-4 py-4 text-sm text-gray-600">
                        <div className="flex items-center gap-1">
                          <LinkIcon className="w-4 h-4 text-gray-400" />
                          {rule.values.join(', ')}
                        </div>
                      </td>
                      <td className="px-4 py-4 text-sm text-gray-600">{Number(rule.matchCount)}</td>
                      <td className="px-4 py-4">
                        <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                          rule.isActive ? 'bg-red-100 text-red-700' : 'bg-gray-100 text-gray-700'
                        }`}>
                          {rule.isActive ? 'Blocking' : 'Inactive'}
                        </span>
                      </td>
                      <td className="px-4 py-4 text-right">
                        <button 
                          onClick={() => handleDeleteRule(rule.id)}
                          className="p-1 text-gray-400 hover:text-red-600"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Domains Tab */}
      {activeTab === 'domains' && (
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div className="p-4 border-b border-gray-100">
            <h3 className="font-semibold text-gray-900">Blocked Domains</h3>
            <p className="text-sm text-gray-500 mt-1">Domain patterns that are blocked from serving ads</p>
          </div>
          {domainRules.length === 0 ? (
            <div className="p-8 text-center text-gray-500">
              <Globe className="w-12 h-12 mx-auto mb-4 text-gray-300" />
              <p>No domain rules configured</p>
            </div>
          ) : (
            <div className="divide-y divide-gray-100">
              {domainRules.map((rule) => (
                <div key={rule.id} className="p-4 flex items-center justify-between hover:bg-gray-50">
                  <div className="flex items-center gap-4">
                    <div className="p-2 bg-gray-100 rounded-lg">
                      <Globe className="w-4 h-4 text-gray-600" />
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">{rule.name}</p>
                      <p className="font-mono text-sm text-gray-500">{rule.values.join(', ')}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className="text-sm text-gray-500">
                      {Number(rule.matchCount)} matches • {new Date(rule.createdAt).toLocaleDateString()}
                    </span>
                    <button 
                      onClick={() => handleDeleteRule(rule.id)}
                      className="p-1 text-gray-400 hover:text-red-600"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
          <div className="p-4 bg-gray-50 border-t border-gray-100">
            <div className="flex items-center gap-2">
              <input
                type="text"
                value={newDomain}
                onChange={(e) => setNewDomain(e.target.value)}
                placeholder="Enter domain to block (e.g., *.example.com)"
                className="flex-1 px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
              />
              <button 
                onClick={handleAddDomain}
                className="px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700"
              >
                Add Domain
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Filters Tab */}
      {activeTab === 'filters' && (
        <div className="space-y-4">
          {contentFilters.map((filter, index) => (
            <div key={index} className="bg-white rounded-xl border border-gray-200 p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <button
                    className={`relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                      filter.enabled ? 'bg-emerald-500' : 'bg-gray-200'
                    }`}
                  >
                    <span
                      className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
                        filter.enabled ? 'translate-x-5' : 'translate-x-0'
                      }`}
                    />
                  </button>
                  <div>
                    <p className="font-medium text-gray-900">{filter.name}</p>
                    <p className="text-sm text-gray-500">{filter.description}</p>
                  </div>
                </div>
                <button className="px-3 py-1 text-sm text-emerald-600 hover:text-emerald-700">
                  Configure
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Add Block Rule Modal */}
      {showAddBlockModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-lg">
            <h2 className="text-xl font-bold text-gray-900 mb-4">Add Block Rule</h2>
            <form onSubmit={handleCreateRule} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Rule Name</label>
                <input
                  type="text"
                  value={newRule.name}
                  onChange={(e) => setNewRule({ ...newRule, name: e.target.value })}
                  placeholder="E.g., Block Gambling Content"
                  className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Block Type</label>
                <select 
                  value={newRule.target}
                  onChange={(e) => setNewRule({ ...newRule, target: e.target.value as BrandSafetyTarget })}
                  className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                >
                  <option value="category">Category</option>
                  <option value="advertiser">Advertiser</option>
                  <option value="domain">Domain</option>
                  <option value="keyword">Keyword</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Values (comma-separated)
                </label>
                <input
                  type="text"
                  value={newRule.values}
                  onChange={(e) => setNewRule({ ...newRule, values: e.target.value })}
                  placeholder="gambling, casino, betting"
                  className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                <input
                  type="text"
                  value={newRule.description}
                  onChange={(e) => setNewRule({ ...newRule, description: e.target.value })}
                  placeholder="Why are you blocking this?"
                  className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                />
              </div>
              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowAddBlockModal(false)}
                  className="flex-1 px-4 py-2 border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700"
                >
                  Add Block Rule
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
