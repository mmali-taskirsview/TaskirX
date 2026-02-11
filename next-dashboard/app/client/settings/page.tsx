'use client'

import { useState, useEffect } from 'react'
import { api } from '@/lib/api'
import { User, Bell, Shield, Key, Globe, Mail, Save, Eye, EyeOff, Loader2 } from 'lucide-react'

interface UserProfile {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
  company?: string;
  role?: string;
  timezone?: string;
}

export default function ClientSettings() {
  const [showPassword, setShowPassword] = useState(false)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [profile, setProfile] = useState<UserProfile>({
    id: '',
    firstName: 'Demo',
    lastName: 'Client',
    email: 'demo@client.com',
    company: 'Demo Client Inc.',
    role: 'advertiser',
    timezone: 'Asia/Singapore',
  })
  const [notifications, setNotifications] = useState({
    email: true,
    push: true,
    campaignAlerts: true,
    budgetAlerts: true,
    weeklyReport: true,
    monthlyReport: true,
  })

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        // Try to fetch user profile from API
        const response = await api.getUsers().catch(() => ({ data: [] }))
        const users = response.data || response || []
        
        // Find current user or use first advertiser
        const user = users.find((u: any) => u.role === 'advertiser') || users[0]
        
        if (user) {
          setProfile({
            id: user.id,
            firstName: user.firstName || user.name?.split(' ')[0] || 'Demo',
            lastName: user.lastName || user.name?.split(' ')[1] || 'Client',
            email: user.email || 'demo@client.com',
            company: user.company || user.tenantId || 'Demo Client Inc.',
            role: user.role || 'advertiser',
            timezone: user.timezone || 'Asia/Singapore',
          })
        }
      } catch (error) {
        console.error('Failed to fetch profile:', error)
      } finally {
        setLoading(false)
      }
    }
    
    fetchProfile()
  }, [])

  const handleSave = async () => {
    setSaving(true)
    try {
      if (profile.id) {
        await api.updateUser(profile.id, {
          firstName: profile.firstName,
          lastName: profile.lastName,
          email: profile.email,
        })
      }
      // Show success feedback
    } catch (error) {
      console.error('Failed to save profile:', error)
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
        <span className="ml-2 text-gray-600">Loading settings...</span>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
        <p className="text-gray-500">Manage your account preferences</p>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Sidebar Navigation */}
        <div className="lg:col-span-1">
          <nav className="space-y-1 rounded-xl bg-white p-4 shadow-sm">
            {[
              { icon: User, label: 'Profile', active: true },
              { icon: Bell, label: 'Notifications', active: false },
              { icon: Shield, label: 'Security', active: false },
              { icon: Key, label: 'API Keys', active: false },
              { icon: Globe, label: 'Preferences', active: false },
            ].map((item) => (
              <button
                key={item.label}
                className={`flex w-full items-center gap-3 rounded-lg px-4 py-2.5 text-sm font-medium ${
                  item.active ? 'bg-blue-50 text-blue-700' : 'text-gray-700 hover:bg-gray-50'
                }`}
              >
                <item.icon className={`h-5 w-5 ${item.active ? 'text-blue-600' : 'text-gray-400'}`} />
                {item.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Profile Section */}
          <div className="rounded-xl bg-white p-6 shadow-sm">
            <h2 className="text-lg font-semibold text-gray-900 mb-6">Profile Information</h2>
            
            <div className="flex items-center gap-6 mb-6">
              <div className="flex h-20 w-20 items-center justify-center rounded-full bg-blue-100">
                <User className="h-10 w-10 text-blue-600" />
              </div>
              <div>
                <button className="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700">
                  Upload Photo
                </button>
                <p className="mt-1 text-xs text-gray-500">JPG, PNG. Max 2MB</p>
              </div>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <label className="text-sm font-medium text-gray-700">First Name</label>
                <input
                  type="text"
                  value={profile.firstName}
                  onChange={(e) => setProfile({ ...profile, firstName: e.target.value })}
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-sm font-medium text-gray-700">Last Name</label>
                <input
                  type="text"
                  value={profile.lastName}
                  onChange={(e) => setProfile({ ...profile, lastName: e.target.value })}
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-sm font-medium text-gray-700">Email</label>
                <input
                  type="email"
                  value={profile.email}
                  onChange={(e) => setProfile({ ...profile, email: e.target.value })}
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-sm font-medium text-gray-700">Phone</label>
                <input
                  type="tel"
                  defaultValue="+65 9123 4567"
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                />
              </div>
              <div className="sm:col-span-2">
                <label className="text-sm font-medium text-gray-700">Company</label>
                <input
                  type="text"
                  value={profile.company}
                  onChange={(e) => setProfile({ ...profile, company: e.target.value })}
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                />
              </div>
            </div>

            <div className="mt-6 flex justify-end">
              <button 
                onClick={handleSave}
                disabled={saving}
                className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
              >
                {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
                {saving ? 'Saving...' : 'Save Changes'}
              </button>
            </div>
          </div>

          {/* Notifications Section */}
          <div className="rounded-xl bg-white p-6 shadow-sm">
            <h2 className="text-lg font-semibold text-gray-900 mb-6">Notification Preferences</h2>
            
            <div className="space-y-4">
              {[
                { key: 'email', label: 'Email Notifications', desc: 'Receive notifications via email' },
                { key: 'push', label: 'Push Notifications', desc: 'Receive browser push notifications' },
                { key: 'campaignAlerts', label: 'Campaign Alerts', desc: 'Get notified about campaign status changes' },
                { key: 'budgetAlerts', label: 'Budget Alerts', desc: 'Get notified when budget thresholds are reached' },
                { key: 'weeklyReport', label: 'Weekly Report', desc: 'Receive weekly performance summary' },
                { key: 'monthlyReport', label: 'Monthly Report', desc: 'Receive monthly detailed report' },
              ].map((item) => (
                <div key={item.key} className="flex items-center justify-between rounded-lg border border-gray-200 p-4">
                  <div>
                    <p className="font-medium text-gray-900">{item.label}</p>
                    <p className="text-sm text-gray-500">{item.desc}</p>
                  </div>
                  <button
                    onClick={() => setNotifications(prev => ({ ...prev, [item.key]: !prev[item.key as keyof typeof prev] }))}
                    className={`relative h-6 w-11 rounded-full transition-colors ${
                      notifications[item.key as keyof typeof notifications] ? 'bg-blue-600' : 'bg-gray-300'
                    }`}
                  >
                    <span className={`absolute top-0.5 h-5 w-5 rounded-full bg-white transition-transform ${
                      notifications[item.key as keyof typeof notifications] ? 'left-5' : 'left-0.5'
                    }`} />
                  </button>
                </div>
              ))}
            </div>
          </div>

          {/* Security Section */}
          <div className="rounded-xl bg-white p-6 shadow-sm">
            <h2 className="text-lg font-semibold text-gray-900 mb-6">Security</h2>
            
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium text-gray-700">Current Password</label>
                <div className="relative mt-1">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    placeholder="••••••••"
                    className="w-full rounded-lg border border-gray-300 px-4 py-2 pr-10 focus:border-blue-500 focus:outline-none"
                  />
                  <button
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400"
                  >
                    {showPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                  </button>
                </div>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-700">New Password</label>
                <input
                  type="password"
                  placeholder="••••••••"
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-sm font-medium text-gray-700">Confirm New Password</label>
                <input
                  type="password"
                  placeholder="••••••••"
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                />
              </div>
            </div>

            <div className="mt-6 flex items-center justify-between">
              <button className="text-sm text-blue-600 hover:text-blue-700">
                Enable Two-Factor Authentication
              </button>
              <button className="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700">
                Update Password
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
