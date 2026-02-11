'use client';

import React, { useState, useEffect } from 'react';
import {
  Settings,
  User,
  Building,
  Globe,
  Key,
  Bell,
  Shield,
  CreditCard,
  FileText,
  Mail,
  Phone,
  MapPin,
  Copy,
  Check,
  Eye,
  EyeOff,
  RefreshCw,
  Save,
  Loader2
} from 'lucide-react';
import { api } from '@/lib/api';

interface Publisher {
  id: string;
  name: string;
  domain: string;
  contactEmail: string;
  contactName?: string;
  status: string;
  apiKey?: string;
  settings?: Record<string, any>;
  createdAt: string;
}

export default function PublisherSettingsPage() {
  const [activeTab, setActiveTab] = useState('account');
  const [showApiKey, setShowApiKey] = useState(false);
  const [copiedKey, setCopiedKey] = useState(false);
  const [publisher, setPublisher] = useState<Publisher | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  // Fetch publisher data
  useEffect(() => {
    async function fetchPublisher() {
      try {
        setLoading(true);
        const response = await api.getPublishers();
        if (response.data && response.data.length > 0) {
          setPublisher(response.data[0]); // Get first publisher for demo
        }
      } catch (err) {
        console.error('Error fetching publisher:', err);
      } finally {
        setLoading(false);
      }
    }
    fetchPublisher();
  }, []);

  const apiKey = publisher?.apiKey || 'pub_live_sk_a8f2g9h3j4k5l6m7n8o9p0q1r2s3t4u5';

  const copyApiKey = () => {
    navigator.clipboard.writeText(apiKey);
    setCopiedKey(true);
    setTimeout(() => setCopiedKey(false), 2000);
  };

  const handleSavePublisher = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!publisher) return;
    
    try {
      setSaving(true);
      await api.updatePublisher(publisher.id, {
        name: publisher.name,
        domain: publisher.domain,
        contactEmail: publisher.contactEmail,
        contactName: publisher.contactName
      });
    } catch (err) {
      console.error('Error saving publisher:', err);
    } finally {
      setSaving(false);
    }
  };

  const tabs = [
    { id: 'account', name: 'Account', icon: User },
    { id: 'company', name: 'Company', icon: Building },
    { id: 'api', name: 'API Keys', icon: Key },
    { id: 'notifications', name: 'Notifications', icon: Bell },
    { id: 'security', name: 'Security', icon: Shield },
  ];

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <Loader2 className="w-8 h-8 animate-spin text-emerald-600" />
        <span className="ml-2 text-gray-600">Loading settings...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
        <p className="text-gray-600 mt-1">Manage your publisher account settings</p>
      </div>

      <div className="flex gap-6">
        {/* Sidebar */}
        <div className="w-64 flex-shrink-0">
          <nav className="space-y-1">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`w-full flex items-center gap-3 px-4 py-3 text-sm font-medium rounded-lg transition-colors ${
                  activeTab === tab.id
                    ? 'bg-emerald-50 text-emerald-600'
                    : 'text-gray-600 hover:bg-gray-50'
                }`}
              >
                <tab.icon className="w-5 h-5" />
                {tab.name}
              </button>
            ))}
          </nav>
        </div>

        {/* Content */}
        <div className="flex-1">
          {/* Account Tab */}
          {activeTab === 'account' && (
            <div className="bg-white rounded-xl border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-6">Account Information</h2>
              <form className="space-y-6">
                <div className="flex items-center gap-6">
                  <div className="w-20 h-20 bg-emerald-100 rounded-full flex items-center justify-center">
                    <User className="w-10 h-10 text-emerald-600" />
                  </div>
                  <div>
                    <button type="button" className="px-4 py-2 text-sm text-emerald-600 border border-emerald-200 rounded-lg hover:bg-emerald-50">
                      Upload Photo
                    </button>
                    <p className="text-sm text-gray-500 mt-1">JPG, PNG up to 2MB</p>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-6">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">First Name</label>
                    <input
                      type="text"
                      defaultValue="John"
                      className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Last Name</label>
                    <input
                      type="text"
                      defaultValue="Publisher"
                      className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Email Address</label>
                  <div className="flex items-center gap-2">
                    <Mail className="w-5 h-5 text-gray-400" />
                    <input
                      type="email"
                      defaultValue="john@publisher.com"
                      className="flex-1 px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Phone Number</label>
                  <div className="flex items-center gap-2">
                    <Phone className="w-5 h-5 text-gray-400" />
                    <input
                      type="tel"
                      defaultValue="+1 (555) 123-4567"
                      className="flex-1 px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Timezone</label>
                  <select className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500">
                    <option>America/New_York (EST)</option>
                    <option>America/Los_Angeles (PST)</option>
                    <option>Europe/London (GMT)</option>
                    <option>Asia/Tokyo (JST)</option>
                  </select>
                </div>

                <div className="pt-4 border-t border-gray-100">
                  <button type="submit" className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700">
                    <Save className="w-4 h-4" />
                    Save Changes
                  </button>
                </div>
              </form>
            </div>
          )}

          {/* Company Tab */}
          {activeTab === 'company' && (
            <div className="bg-white rounded-xl border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-6">Company Information</h2>
              <form onSubmit={handleSavePublisher} className="space-y-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Publisher ID</label>
                  <input
                    type="text"
                    value={publisher?.id || 'N/A'}
                    disabled
                    className="w-full px-4 py-2 border border-gray-200 rounded-lg bg-gray-50 text-gray-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Company Name</label>
                  <input
                    type="text"
                    value={publisher?.name || ''}
                    onChange={(e) => setPublisher(publisher ? {...publisher, name: e.target.value} : null)}
                    className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Website URL</label>
                  <div className="flex items-center gap-2">
                    <Globe className="w-5 h-5 text-gray-400" />
                    <input
                      type="url"
                      value={publisher?.domain || ''}
                      onChange={(e) => setPublisher(publisher ? {...publisher, domain: e.target.value} : null)}
                      className="flex-1 px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Contact Email</label>
                  <div className="flex items-center gap-2">
                    <Mail className="w-5 h-5 text-gray-400" />
                    <input
                      type="email"
                      value={publisher?.contactEmail || ''}
                      onChange={(e) => setPublisher(publisher ? {...publisher, contactEmail: e.target.value} : null)}
                      className="flex-1 px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Contact Name</label>
                  <input
                    type="text"
                    value={publisher?.contactName || ''}
                    onChange={(e) => setPublisher(publisher ? {...publisher, contactName: e.target.value} : null)}
                    className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
                  <span className={`inline-flex px-3 py-1 text-sm font-medium rounded-full ${
                    publisher?.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                  }`}>
                    {publisher?.status || 'Unknown'}
                  </span>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Member Since</label>
                  <p className="text-gray-600">
                    {publisher?.createdAt ? new Date(publisher.createdAt).toLocaleDateString() : 'N/A'}
                  </p>
                </div>

                <div className="pt-4 border-t border-gray-100">
                  <button 
                    type="submit" 
                    disabled={saving}
                    className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 disabled:opacity-50"
                  >
                    {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
                    {saving ? 'Saving...' : 'Save Changes'}
                  </button>
                </div>
              </form>
            </div>
          )}

          {/* API Keys Tab */}
          {activeTab === 'api' && (
            <div className="space-y-6">
              <div className="bg-white rounded-xl border border-gray-200 p-6">
                <h2 className="text-lg font-semibold text-gray-900 mb-2">API Keys</h2>
                <p className="text-sm text-gray-500 mb-6">Manage your API keys for programmatic access</p>

                <div className="space-y-4">
                  <div className="p-4 bg-gray-50 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-sm font-medium text-gray-700">Live API Key</span>
                      <span className="px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded-full">Active</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <input
                        type={showApiKey ? 'text' : 'password'}
                        value={apiKey}
                        readOnly
                        className="flex-1 px-4 py-2 border border-gray-200 rounded-lg bg-white font-mono text-sm"
                      />
                      <button
                        onClick={() => setShowApiKey(!showApiKey)}
                        className="p-2 text-gray-400 hover:text-gray-600"
                      >
                        {showApiKey ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                      </button>
                      <button
                        onClick={copyApiKey}
                        className="p-2 text-gray-400 hover:text-gray-600"
                      >
                        {copiedKey ? <Check className="w-5 h-5 text-green-500" /> : <Copy className="w-5 h-5" />}
                      </button>
                    </div>
                    <p className="text-xs text-gray-500 mt-2">Created: Jan 15, 2026 • Last used: 2 minutes ago</p>
                  </div>

                  <button className="flex items-center gap-2 px-4 py-2 text-emerald-600 border border-emerald-200 rounded-lg hover:bg-emerald-50">
                    <RefreshCw className="w-4 h-4" />
                    Regenerate API Key
                  </button>
                </div>
              </div>

              <div className="bg-amber-50 border border-amber-200 rounded-xl p-4">
                <h4 className="font-medium text-amber-800">Security Warning</h4>
                <p className="text-sm text-amber-700 mt-1">
                  Keep your API keys secure. Never expose them in client-side code or public repositories.
                  If you believe your key has been compromised, regenerate it immediately.
                </p>
              </div>
            </div>
          )}

          {/* Notifications Tab */}
          {activeTab === 'notifications' && (
            <div className="bg-white rounded-xl border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-6">Notification Preferences</h2>
              <div className="space-y-6">
                {[
                  { title: 'Daily Revenue Reports', description: 'Get daily summaries of your earnings' },
                  { title: 'Weekly Performance Digest', description: 'Weekly overview of key metrics' },
                  { title: 'Payout Notifications', description: 'Alerts when payouts are processed' },
                  { title: 'Fill Rate Alerts', description: 'Notify when fill rate drops below threshold' },
                  { title: 'New Partner Opportunities', description: 'Updates about new demand partners' },
                  { title: 'System Maintenance', description: 'Scheduled maintenance notifications' },
                ].map((item, index) => (
                  <div key={index} className="flex items-center justify-between">
                    <div>
                      <p className="font-medium text-gray-900">{item.title}</p>
                      <p className="text-sm text-gray-500">{item.description}</p>
                    </div>
                    <button className="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none bg-emerald-500">
                      <span className="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out translate-x-5" />
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Security Tab */}
          {activeTab === 'security' && (
            <div className="space-y-6">
              <div className="bg-white rounded-xl border border-gray-200 p-6">
                <h2 className="text-lg font-semibold text-gray-900 mb-6">Password</h2>
                <form className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Current Password</label>
                    <input
                      type="password"
                      className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">New Password</label>
                    <input
                      type="password"
                      className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Confirm New Password</label>
                    <input
                      type="password"
                      className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    />
                  </div>
                  <button type="submit" className="px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700">
                    Update Password
                  </button>
                </form>
              </div>

              <div className="bg-white rounded-xl border border-gray-200 p-6">
                <h2 className="text-lg font-semibold text-gray-900 mb-2">Two-Factor Authentication</h2>
                <p className="text-sm text-gray-500 mb-4">Add an extra layer of security to your account</p>
                <div className="flex items-center justify-between p-4 bg-green-50 rounded-lg border border-green-200">
                  <div className="flex items-center gap-3">
                    <Shield className="w-6 h-6 text-green-600" />
                    <div>
                      <p className="font-medium text-green-800">2FA is enabled</p>
                      <p className="text-sm text-green-600">Using authenticator app</p>
                    </div>
                  </div>
                  <button className="px-3 py-1 text-sm text-green-700 border border-green-300 rounded-lg hover:bg-green-100">
                    Manage
                  </button>
                </div>
              </div>

              <div className="bg-white rounded-xl border border-gray-200 p-6">
                <h2 className="text-lg font-semibold text-gray-900 mb-4">Active Sessions</h2>
                <div className="space-y-3">
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div>
                      <p className="font-medium text-gray-900">Chrome on Windows</p>
                      <p className="text-sm text-gray-500">New York, US • Current session</p>
                    </div>
                    <span className="px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded-full">Active</span>
                  </div>
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div>
                      <p className="font-medium text-gray-900">Safari on iPhone</p>
                      <p className="text-sm text-gray-500">New York, US • 2 hours ago</p>
                    </div>
                    <button className="text-sm text-red-600 hover:text-red-700">Revoke</button>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
