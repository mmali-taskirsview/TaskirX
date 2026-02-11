import React, { useState, useEffect } from 'react';
import { Settings as SettingsIcon, Copy, Eye, EyeOff, Save, Lock, Bell, User } from 'lucide-react';

const Settings = () => {
  const [activeTab, setActiveTab] = useState('profile');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [showApiKey, setShowApiKey] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);

  // Profile state
  const [profile, setProfile] = useState({
    name: '',
    email: '',
    company: '',
    phone: '',
    timezone: 'UTC',
    language: 'en'
  });

  // API Keys state
  const [apiKeys, setApiKeys] = useState([]);
  const [newKeyName, setNewKeyName] = useState('');
  const [showNewKeyModal, setShowNewKeyModal] = useState(false);

  // Password state
  const [passwords, setPasswords] = useState({
    current: '',
    new: '',
    confirm: ''
  });

  // Notification preferences
  const [notifications, setNotifications] = useState({
    campaignUpdates: true,
    bidNotifications: true,
    analyticsDaily: true,
    analyticsWeekly: false,
    performanceAlerts: true,
    systemNotifications: true,
    email: true,
    sms: false,
    inApp: true
  });

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      setLoading(true);
      const [profileRes, keysRes, notifRes] = await Promise.all([
        fetch('/api/users/profile', {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        }),
        fetch('/api/users/api-keys', {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        }),
        fetch('/api/users/notifications', {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
      ]);

      if (profileRes.ok) {
        const data = await profileRes.json();
        setProfile(data.profile);
      }

      if (keysRes.ok) {
        const data = await keysRes.json();
        setApiKeys(data.keys || []);
      }

      if (notifRes.ok) {
        const data = await notifRes.json();
        setNotifications(data.preferences);
      }

      setError(null);
    } catch (err) {
      console.error('Error fetching settings:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveProfile = async (e) => {
    e.preventDefault();
    try {
      setSaving(true);
      const response = await fetch('/api/users/profile', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(profile)
      });

      if (!response.ok) throw new Error('Failed to save profile');

      setSuccess('Profile updated successfully!');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const handleChangePassword = async (e) => {
    e.preventDefault();

    if (passwords.new !== passwords.confirm) {
      setError('New passwords do not match');
      return;
    }

    try {
      setSaving(true);
      const response = await fetch('/api/users/password', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({
          currentPassword: passwords.current,
          newPassword: passwords.new
        })
      });

      if (!response.ok) throw new Error('Failed to change password');

      setSuccess('Password changed successfully!');
      setPasswords({ current: '', new: '', confirm: '' });
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const handleCreateApiKey = async (e) => {
    e.preventDefault();

    if (!newKeyName.trim()) {
      setError('Please enter a key name');
      return;
    }

    try {
      setSaving(true);
      const response = await fetch('/api/users/api-keys', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ name: newKeyName })
      });

      if (!response.ok) throw new Error('Failed to create API key');

      const data = await response.json();
      setApiKeys([...apiKeys, data.key]);
      setNewKeyName('');
      setShowNewKeyModal(false);
      setSuccess('API key created successfully!');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteApiKey = async (keyId) => {
    if (!confirm('Are you sure you want to delete this API key?')) return;

    try {
      const response = await fetch(`/api/users/api-keys/${keyId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) throw new Error('Failed to delete API key');

      setApiKeys(apiKeys.filter(k => k.id !== keyId));
      setSuccess('API key deleted successfully!');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleRotateApiKey = async (keyId) => {
    try {
      const response = await fetch(`/api/users/api-keys/${keyId}/rotate`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) throw new Error('Failed to rotate API key');

      fetchSettings();
      setSuccess('API key rotated successfully!');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleSaveNotifications = async () => {
    try {
      setSaving(true);
      const response = await fetch('/api/users/notifications', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(notifications)
      });

      if (!response.ok) throw new Error('Failed to save preferences');

      setSuccess('Notification preferences updated!');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    setSuccess('Copied to clipboard!');
    setTimeout(() => setSuccess(null), 2000);
  };

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow p-8 text-center">
        <div className="animate-spin inline-block">
          <svg className="h-12 w-12 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          </svg>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <h2 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
        <SettingsIcon size={28} />
        ⚙️ Settings & Preferences
      </h2>

      {/* Alerts */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-800">
          {error}
        </div>
      )}
      {success && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-4 text-green-800">
          {success}
        </div>
      )}

      {/* Tab Navigation */}
      <div className="bg-white rounded-lg shadow">
        <div className="flex border-b border-gray-200">
          {[
            { id: 'profile', label: 'Profile', icon: '👤' },
            { id: 'api-keys', label: 'API Keys', icon: '🔑' },
            { id: 'password', label: 'Password', icon: '🔐' },
            { id: 'notifications', label: 'Notifications', icon: '🔔' }
          ].map(tab => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex-1 px-6 py-4 font-medium text-center transition-colors ${
                activeTab === tab.id
                  ? 'bg-blue-50 text-blue-600 border-b-2 border-blue-600'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              <span className="mr-2">{tab.icon}</span>
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Tab Content */}
      <div className="bg-white rounded-lg shadow p-8">
        {/* Profile Tab */}
        {activeTab === 'profile' && (
          <form onSubmit={handleSaveProfile} className="space-y-6">
            <h3 className="text-xl font-bold text-gray-900">User Profile</h3>

            <div className="grid grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Full Name</label>
                <input
                  type="text"
                  value={profile.name}
                  onChange={(e) => setProfile({...profile, name: e.target.value})}
                  className="w-full px-4 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
                <input
                  type="email"
                  value={profile.email}
                  disabled
                  className="w-full px-4 py-2 border border-gray-300 rounded text-gray-900 bg-gray-50 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Company</label>
                <input
                  type="text"
                  value={profile.company}
                  onChange={(e) => setProfile({...profile, company: e.target.value})}
                  className="w-full px-4 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Phone</label>
                <input
                  type="tel"
                  value={profile.phone}
                  onChange={(e) => setProfile({...profile, phone: e.target.value})}
                  className="w-full px-4 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Timezone</label>
                <select
                  value={profile.timezone}
                  onChange={(e) => setProfile({...profile, timezone: e.target.value})}
                  className="w-full px-4 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="UTC">UTC</option>
                  <option value="EST">EST (UTC-5)</option>
                  <option value="CST">CST (UTC-6)</option>
                  <option value="MST">MST (UTC-7)</option>
                  <option value="PST">PST (UTC-8)</option>
                  <option value="GMT">GMT (UTC+0)</option>
                  <option value="CET">CET (UTC+1)</option>
                  <option value="IST">IST (UTC+5:30)</option>
                  <option value="JST">JST (UTC+9)</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Language</label>
                <select
                  value={profile.language}
                  onChange={(e) => setProfile({...profile, language: e.target.value})}
                  className="w-full px-4 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="en">English</option>
                  <option value="es">Spanish</option>
                  <option value="fr">French</option>
                  <option value="de">German</option>
                  <option value="ja">Japanese</option>
                  <option value="zh">Chinese</option>
                </select>
              </div>
            </div>

            <div className="flex justify-end">
              <button
                type="submit"
                disabled={saving}
                className="bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-medium py-2 px-6 rounded flex items-center gap-2 transition-colors"
              >
                <Save size={20} />
                Save Profile
              </button>
            </div>
          </form>
        )}

        {/* API Keys Tab */}
        {activeTab === 'api-keys' && (
          <div className="space-y-6">
            <div className="flex justify-between items-center">
              <h3 className="text-xl font-bold text-gray-900">API Keys</h3>
              <button
                onClick={() => setShowNewKeyModal(true)}
                className="bg-green-600 hover:bg-green-700 text-white font-medium py-2 px-4 rounded transition-colors"
              >
                Generate New Key
              </button>
            </div>

            <div className="space-y-3">
              {apiKeys.map(key => (
                <div key={key.id} className="border border-gray-200 rounded-lg p-4 flex items-center justify-between">
                  <div className="flex-1">
                    <p className="font-medium text-gray-900">{key.name}</p>
                    <div className="flex items-center gap-2 mt-2">
                      <code className={`text-sm ${showApiKey && key.id === apiKeys[0].id ? 'text-gray-900' : 'text-gray-600'}`}>
                        {showApiKey && key.id === apiKeys[0].id ? key.key : key.key.substring(0, 20) + '****...'}
                      </code>
                      <button
                        onClick={() => copyToClipboard(key.key)}
                        className="text-blue-600 hover:text-blue-900 p-1"
                      >
                        <Copy size={16} />
                      </button>
                    </div>
                    <p className="text-xs text-gray-500 mt-2">
                      Created: {new Date(key.createdAt).toLocaleDateString()}
                    </p>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleRotateApiKey(key.id)}
                      className="text-orange-600 hover:text-orange-900 p-2"
                      title="Rotate Key"
                    >
                      🔄
                    </button>
                    <button
                      onClick={() => handleDeleteApiKey(key.id)}
                      className="text-red-600 hover:text-red-900 p-2"
                      title="Delete Key"
                    >
                      🗑️
                    </button>
                  </div>
                </div>
              ))}
            </div>

            {showNewKeyModal && (
              <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
                  <div className="p-6 border-b border-gray-200">
                    <h4 className="text-lg font-bold text-gray-900">Generate New API Key</h4>
                  </div>
                  <form onSubmit={handleCreateApiKey} className="p-6 space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Key Name</label>
                      <input
                        type="text"
                        value={newKeyName}
                        onChange={(e) => setNewKeyName(e.target.value)}
                        placeholder="e.g., Mobile App, Dashboard"
                        className="w-full px-3 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                    <div className="flex justify-end gap-3">
                      <button
                        type="button"
                        onClick={() => setShowNewKeyModal(false)}
                        className="px-4 py-2 text-gray-700 border border-gray-300 rounded hover:bg-gray-50"
                      >
                        Cancel
                      </button>
                      <button
                        type="submit"
                        disabled={saving}
                        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
                      >
                        Generate
                      </button>
                    </div>
                  </form>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Password Tab */}
        {activeTab === 'password' && (
          <form onSubmit={handleChangePassword} className="space-y-6 max-w-md">
            <h3 className="text-xl font-bold text-gray-900">Change Password</h3>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Current Password</label>
              <input
                type="password"
                value={passwords.current}
                onChange={(e) => setPasswords({...passwords, current: e.target.value})}
                className="w-full px-4 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">New Password</label>
              <div className="relative">
                <input
                  type={showNewPassword ? 'text' : 'password'}
                  value={passwords.new}
                  onChange={(e) => setPasswords({...passwords, new: e.target.value})}
                  className="w-full px-4 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <button
                  type="button"
                  onClick={() => setShowNewPassword(!showNewPassword)}
                  className="absolute right-3 top-2.5 text-gray-600"
                >
                  {showNewPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                </button>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Confirm New Password</label>
              <input
                type="password"
                value={passwords.confirm}
                onChange={(e) => setPasswords({...passwords, confirm: e.target.value})}
                className="w-full px-4 py-2 border border-gray-300 rounded text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <button
              type="submit"
              disabled={saving}
              className="bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-medium py-2 px-6 rounded flex items-center gap-2 transition-colors"
            >
              <Lock size={20} />
              Change Password
            </button>
          </form>
        )}

        {/* Notifications Tab */}
        {activeTab === 'notifications' && (
          <form onSubmit={(e) => { e.preventDefault(); handleSaveNotifications(); }} className="space-y-6">
            <h3 className="text-xl font-bold text-gray-900">Notification Preferences</h3>

            <div className="space-y-4">
              <div className="border-b border-gray-200 pb-4">
                <h4 className="font-medium text-gray-900 mb-3">Event Notifications</h4>
                <div className="space-y-2">
                  {[
                    { key: 'campaignUpdates', label: 'Campaign Updates' },
                    { key: 'bidNotifications', label: 'Bid Notifications' },
                    { key: 'performanceAlerts', label: 'Performance Alerts' }
                  ].map(item => (
                    <label key={item.key} className="flex items-center gap-3 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={notifications[item.key]}
                        onChange={(e) => setNotifications({...notifications, [item.key]: e.target.checked})}
                        className="w-4 h-4 rounded"
                      />
                      <span className="text-gray-700">{item.label}</span>
                    </label>
                  ))}
                </div>
              </div>

              <div className="border-b border-gray-200 pb-4">
                <h4 className="font-medium text-gray-900 mb-3">Analytics Reports</h4>
                <div className="space-y-2">
                  {[
                    { key: 'analyticsDaily', label: 'Daily Report' },
                    { key: 'analyticsWeekly', label: 'Weekly Report' }
                  ].map(item => (
                    <label key={item.key} className="flex items-center gap-3 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={notifications[item.key]}
                        onChange={(e) => setNotifications({...notifications, [item.key]: e.target.checked})}
                        className="w-4 h-4 rounded"
                      />
                      <span className="text-gray-700">{item.label}</span>
                    </label>
                  ))}
                </div>
              </div>

              <div>
                <h4 className="font-medium text-gray-900 mb-3">Notification Channels</h4>
                <div className="space-y-2">
                  {[
                    { key: 'email', label: 'Email Notifications' },
                    { key: 'sms', label: 'SMS Notifications' },
                    { key: 'inApp', label: 'In-App Notifications' }
                  ].map(item => (
                    <label key={item.key} className="flex items-center gap-3 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={notifications[item.key]}
                        onChange={(e) => setNotifications({...notifications, [item.key]: e.target.checked})}
                        className="w-4 h-4 rounded"
                      />
                      <span className="text-gray-700">{item.label}</span>
                    </label>
                  ))}
                </div>
              </div>
            </div>

            <div className="flex justify-end">
              <button
                type="submit"
                disabled={saving}
                className="bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-medium py-2 px-6 rounded flex items-center gap-2 transition-colors"
              >
                <Bell size={20} />
                Save Preferences
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
};

export default Settings;
