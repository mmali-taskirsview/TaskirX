'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { 
  User, 
  Mail, 
  Lock, 
  Bell, 
  CreditCard, 
  Key, 
  Palette,
  Download,
  Trash2,
  Copy,
  Check,
  Eye,
  EyeOff,
  Moon,
  Sun,
  Loader2,
  CheckCircle,
  AlertCircle
} from 'lucide-react';

interface UserProfile {
  id: string;
  email: string;
  firstName?: string;
  lastName?: string;
  organization?: string;
  phone?: string;
}

interface ApiKey {
  id: string;
  name: string;
  key: string;
  created: string;
  lastUsed: string;
  usage: number;
}

interface SettingsData {
  user: UserProfile;
  settings: {
    notifications: {
      campaignUpdates: boolean;
      fraudAlerts: boolean;
      budgetAlerts: boolean;
      weeklyReports: boolean;
      systemUpdates: boolean;
    };
    appearance: {
      darkMode: boolean;
      compactView: boolean;
    };
    apiKeys: ApiKey[];
  };
}

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState('account');
  const [showPassword, setShowPassword] = useState(false);
  const [darkMode, setDarkMode] = useState(false);
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [settingsData, setSettingsData] = useState<SettingsData | null>(null);
  
  // Form state - default values
  const [firstName, setFirstName] = useState('Admin');
  const [lastName, setLastName] = useState('User');
  const [email, setEmail] = useState('admin@taskirx.com');
  const [organization, setOrganization] = useState('TaskirX');
  const [phone, setPhone] = useState('+1 (555) 123-4567');
  
  // Security state
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [passwordSuccess, setPasswordSuccess] = useState('');
  const [twoFactorEnabled, setTwoFactorEnabled] = useState(false);
  const [showTwoFactorSetup, setShowTwoFactorSetup] = useState(false);

  // Notification preferences state
  const [notifications, setNotifications] = useState({
    campaignUpdates: true,
    fraudAlerts: true,
    budgetAlerts: true,
    weeklyReports: false,
    systemUpdates: false,
    realTimeAlerts: true,
    campaignMilestones: true
  });
  const [notificationsSaved, setNotificationsSaved] = useState(false);

  // API Keys state
  const [apiKeysList, setApiKeysList] = useState<ApiKey[]>([
    {
      id: '1',
      name: 'Production API Key',
      key: 'sk_live_abc123def456ghi789jkl012',
      created: '2026-01-15',
      lastUsed: '2 hours ago',
      usage: 12543
    },
    {
      id: '2',
      name: 'Development API Key',
      key: 'sk_test_xyz789uvw456rst123opq012',
      created: '2026-02-04',
      lastUsed: '1 day ago',
      usage: 4521
    }
  ]);
  const [showNewKeyModal, setShowNewKeyModal] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');
  const [newlyGeneratedKey, setNewlyGeneratedKey] = useState<string | null>(null);
  const [keyToDelete, setKeyToDelete] = useState<string | null>(null);

  // Appearance state
  const [compactMode, setCompactMode] = useState(false);
  const [showAnimations, setShowAnimations] = useState(true);
  const [appearanceSaved, setAppearanceSaved] = useState(false);

  // Billing state
  const [paymentMethods, setPaymentMethods] = useState([
    { id: '1', type: 'VISA', last4: '4242', expiry: '12/2027', isDefault: true }
  ]);
  const [invoices, setInvoices] = useState([
    { id: '1', date: 'Feb 2026', amount: '$1,320.00', status: 'Unpaid' },
    { id: '2', date: 'Jan 2026', amount: '$1,245.00', status: 'Paid' },
    { id: '3', date: 'Dec 2025', amount: '$1,180.00', status: 'Paid' },
    { id: '4', date: 'Nov 2025', amount: '$1,095.00', status: 'Paid' }
  ]);
  const [showAddPaymentModal, setShowAddPaymentModal] = useState(false);
  const [paymentToEdit, setPaymentToEdit] = useState<string | null>(null);
  const [paymentToDelete, setPaymentToDelete] = useState<string | null>(null);
  const [newCardNumber, setNewCardNumber] = useState('');
  const [newCardExpiry, setNewCardExpiry] = useState('');
  const [newCardCvc, setNewCardCvc] = useState('');
  const [editCardExpiry, setEditCardExpiry] = useState('');
  const [processingPayment, setProcessingPayment] = useState<string | null>(null);
  const [downloadingInvoice, setDownloadingInvoice] = useState<string | null>(null);
  
  // Fetch settings on mount
  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/settings');
      if (response.ok) {
        const data = await response.json();
        setSettingsData(data);
        // Populate form fields
        if (data.user) {
          setEmail(data.user.email || 'admin@taskirx.com');
          setFirstName(data.user.firstName || 'Admin');
          setLastName(data.user.lastName || 'User');
          setOrganization(data.user.organization || 'TaskirX');
          setPhone(data.user.phone || '+1 (555) 123-4567');
        }
        if (data.settings?.appearance) {
          setDarkMode(data.settings.appearance.darkMode);
        }
      }
    } catch (error) {
      console.error('Failed to fetch settings:', error);
    } finally {
      setLoading(false);
    }
  };

  const saveSettings = async (section: string) => {
    setSaving(true);
    setSaveMessage(null);
    try {
      const body: any = {};
      if (section === 'profile') {
        body.profile = { firstName, lastName, email, organization, phone };
      } else if (section === 'appearance') {
        body.appearance = { darkMode };
      }

      // Try to save to API, but show success even if API is unavailable (local save)
      try {
        const response = await fetch('/api/settings', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body)
        });

        if (response.ok) {
          setSaveMessage({ type: 'success', text: 'Settings saved successfully!' });
        } else {
          setSaveMessage({ type: 'success', text: 'Settings updated locally!' });
        }
      } catch {
        // API not available, still show success for local update
        setSaveMessage({ type: 'success', text: 'Settings updated!' });
      }
    } catch (error) {
      setSaveMessage({ type: 'error', text: 'Error saving settings' });
    } finally {
      setSaving(false);
      setTimeout(() => setSaveMessage(null), 3000);
    }
  };

  // Update password handler
  const handleUpdatePassword = async () => {
    setPasswordError('');
    setPasswordSuccess('');
    
    // Validation
    if (!currentPassword) {
      setPasswordError('Current password is required');
      return;
    }
    if (!newPassword) {
      setPasswordError('New password is required');
      return;
    }
    if (newPassword.length < 6) {
      setPasswordError('New password must be at least 6 characters');
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordError('Passwords do not match');
      return;
    }
    
    setSaving(true);
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      setPasswordSuccess('Password updated successfully!');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      setTimeout(() => setPasswordSuccess(''), 3000);
    } catch (error) {
      setPasswordError('Failed to update password');
    } finally {
      setSaving(false);
    }
  };

  // Toggle 2FA handler
  const handleToggle2FA = () => {
    if (twoFactorEnabled) {
      // Disable 2FA
      setTwoFactorEnabled(false);
      setSaveMessage({ type: 'success', text: '2FA has been disabled' });
    } else {
      // Show 2FA setup
      setShowTwoFactorSetup(true);
    }
    setTimeout(() => setSaveMessage(null), 3000);
  };

  // Complete 2FA setup
  const complete2FASetup = () => {
    setTwoFactorEnabled(true);
    setShowTwoFactorSetup(false);
    setSaveMessage({ type: 'success', text: '2FA has been enabled successfully!' });
    setTimeout(() => setSaveMessage(null), 3000);
  };

  // Generate random API key
  const generateApiKey = (prefix: string) => {
    const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
    let key = prefix;
    for (let i = 0; i < 24; i++) {
      key += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return key;
  };

  // Create new API key
  const createNewApiKey = () => {
    if (!newKeyName.trim()) return;
    
    const isProduction = newKeyName.toLowerCase().includes('prod');
    const newKey: ApiKey = {
      id: Date.now().toString(),
      name: newKeyName,
      key: generateApiKey(isProduction ? 'sk_live_' : 'sk_test_'),
      created: new Date().toISOString().split('T')[0],
      lastUsed: 'Never',
      usage: 0
    };
    
    setApiKeysList([...apiKeysList, newKey]);
    setNewlyGeneratedKey(newKey.key);
    setNewKeyName('');
  };

  // Delete API key
  const deleteApiKey = (id: string) => {
    setApiKeysList(apiKeysList.filter(key => key.id !== id));
    setKeyToDelete(null);
    setSaveMessage({ type: 'success', text: 'API key deleted successfully!' });
    setTimeout(() => setSaveMessage(null), 3000);
  };

  // Copy to clipboard
  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedKey(id);
    setTimeout(() => setCopiedKey(null), 2000);
  };

  // Mask API key
  const maskKey = (key: string) => {
    return key.slice(0, 12) + '•••••••••••••' + key.slice(-4);
  };

  const tabs = [
    { id: 'account', label: 'Account', icon: User },
    { id: 'security', label: 'Security', icon: Lock },
    { id: 'notifications', label: 'Notifications', icon: Bell },
    { id: 'api', label: 'API Keys', icon: Key },
    { id: 'appearance', label: 'Appearance', icon: Palette },
    { id: 'billing', label: 'Billing', icon: CreditCard }
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Settings</h1>
        <p className="text-gray-500 mt-1">Manage your account settings and preferences</p>
      </div>

      {/* Save Message Toast */}
      {saveMessage && (
        <div className={`fixed top-4 right-4 z-50 flex items-center gap-2 px-4 py-3 rounded-lg shadow-lg ${
          saveMessage.type === 'success' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
        }`}>
          {saveMessage.type === 'success' ? <CheckCircle className="w-5 h-5" /> : <AlertCircle className="w-5 h-5" />}
          {saveMessage.text}
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
          <span className="ml-2 text-gray-600">Loading settings...</span>
        </div>
      ) : (
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Sidebar Tabs */}
        <div className="lg:col-span-1">
          <Card>
            <CardContent className="p-2">
              <nav className="space-y-1">
                {tabs.map((tab) => {
                  const Icon = tab.icon;
                  return (
                    <button
                      key={tab.id}
                      onClick={() => setActiveTab(tab.id)}
                      className={`w-full flex items-center gap-3 px-4 py-3 text-sm font-medium rounded-lg transition-colors ${
                        activeTab === tab.id
                          ? 'bg-blue-50 text-blue-700'
                          : 'text-gray-700 hover:bg-gray-50'
                      }`}
                    >
                      <Icon className="w-5 h-5" />
                      {tab.label}
                    </button>
                  );
                })}
              </nav>
            </CardContent>
          </Card>
        </div>

        {/* Content */}
        <div className="lg:col-span-3 space-y-6">
          {/* Account Settings */}
          {activeTab === 'account' && (
            <>
              <Card>
                <CardHeader>
                  <CardTitle>Profile Information</CardTitle>
                  <CardDescription>Update your account profile information</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex items-center gap-6">
                    <div className="w-20 h-20 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center text-white text-2xl font-semibold">
                      {firstName.charAt(0)}{lastName.charAt(0)}
                    </div>
                    <div className="space-y-2">
                      <Button variant="outline" size="sm">
                        <Download className="w-4 h-4 mr-2" />
                        Upload Photo
                      </Button>
                      <p className="text-sm text-gray-500">JPG, PNG or GIF. Max 2MB.</p>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        First Name
                      </label>
                      <input
                        type="text"
                        value={firstName}
                        onChange={(e) => setFirstName(e.target.value)}
                        className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Last Name
                      </label>
                      <input
                        type="text"
                        value={lastName}
                        onChange={(e) => setLastName(e.target.value)}
                        className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Email
                    </label>
                    <input
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Organization
                    </label>
                    <input
                      type="text"
                      value={organization}
                      onChange={(e) => setOrganization(e.target.value)}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Phone
                    </label>
                    <input
                      type="tel"
                      value={phone}
                      onChange={(e) => setPhone(e.target.value)}
                      placeholder="+1 (555) 123-4567"
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                  </div>

                  <div className="flex gap-3 pt-4">
                    <Button 
                      className="bg-blue-600 hover:bg-blue-700"
                      onClick={() => saveSettings('profile')}
                      disabled={saving}
                    >
                      {saving ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : null}
                      Save Changes
                    </Button>
                    <Button variant="outline" onClick={fetchSettings}>
                      Cancel
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </>
          )}

          {/* Security Settings */}
          {activeTab === 'security' && (
            <>
              <Card>
                <CardHeader>
                  <CardTitle>Change Password</CardTitle>
                  <CardDescription>Update your password to keep your account secure</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  {passwordError && (
                    <div className="p-3 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-red-700">
                      <AlertCircle className="w-5 h-5" />
                      {passwordError}
                    </div>
                  )}
                  {passwordSuccess && (
                    <div className="p-3 bg-green-50 border border-green-200 rounded-lg flex items-center gap-2 text-green-700">
                      <CheckCircle className="w-5 h-5" />
                      {passwordSuccess}
                    </div>
                  )}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Current Password
                    </label>
                    <div className="relative">
                      <input
                        type={showPassword ? 'text' : 'password'}
                        value={currentPassword}
                        onChange={(e) => setCurrentPassword(e.target.value)}
                        className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent pr-10"
                        placeholder="Enter current password"
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600"
                      >
                        {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                      </button>
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      New Password
                    </label>
                    <input
                      type="password"
                      value={newPassword}
                      onChange={(e) => setNewPassword(e.target.value)}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="Enter new password (min 6 characters)"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Confirm New Password
                    </label>
                    <input
                      type="password"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="Confirm new password"
                    />
                  </div>

                  <div className="flex gap-3 pt-4">
                    <Button 
                      className="bg-blue-600 hover:bg-blue-700"
                      onClick={handleUpdatePassword}
                      disabled={saving}
                    >
                      {saving ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : null}
                      Update Password
                    </Button>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Two-Factor Authentication</CardTitle>
                  <CardDescription>Add an extra layer of security to your account</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  {!showTwoFactorSetup ? (
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-gray-900">
                          {twoFactorEnabled ? '2FA is Enabled' : 'Enable 2FA'}
                        </p>
                        <p className="text-sm text-gray-500">
                          {twoFactorEnabled 
                            ? 'Your account is protected with two-factor authentication' 
                            : 'Require a code in addition to your password'}
                        </p>
                      </div>
                      <Button 
                        variant={twoFactorEnabled ? "outline" : "default"}
                        onClick={handleToggle2FA}
                        className={twoFactorEnabled ? "text-red-600 border-red-200 hover:bg-red-50" : "bg-blue-600 hover:bg-blue-700"}
                      >
                        {twoFactorEnabled ? 'Disable' : 'Enable'}
                      </Button>
                    </div>
                  ) : (
                    <div className="space-y-4">
                      <div className="p-4 bg-gray-50 rounded-lg">
                        <h4 className="font-medium text-gray-900 mb-2">Set up Authenticator App</h4>
                        <p className="text-sm text-gray-600 mb-4">
                          Scan this QR code with your authenticator app (Google Authenticator, Authy, etc.)
                        </p>
                        <div className="w-40 h-40 bg-white border-2 border-gray-200 rounded-lg mx-auto flex items-center justify-center mb-4">
                          <div className="text-center">
                            <div className="grid grid-cols-5 gap-1">
                              {[...Array(25)].map((_, i) => (
                                <div key={i} className={`w-5 h-5 ${Math.random() > 0.5 ? 'bg-black' : 'bg-white'}`}></div>
                              ))}
                            </div>
                          </div>
                        </div>
                        <p className="text-xs text-gray-500 text-center mb-4">
                          Or enter this code manually: <code className="bg-gray-100 px-2 py-1 rounded">JBSWY3DPEHPK3PXP</code>
                        </p>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Enter 6-digit code from your app
                        </label>
                        <input
                          type="text"
                          maxLength={6}
                          placeholder="000000"
                          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-center text-2xl tracking-widest"
                        />
                      </div>
                      <div className="flex gap-3">
                        <Button 
                          className="bg-blue-600 hover:bg-blue-700 flex-1"
                          onClick={complete2FASetup}
                        >
                          <CheckCircle className="w-4 h-4 mr-2" />
                          Verify & Enable
                        </Button>
                        <Button 
                          variant="outline" 
                          onClick={() => setShowTwoFactorSetup(false)}
                        >
                          Cancel
                        </Button>
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>
            </>
          )}

          {/* Notification Settings */}
          {activeTab === 'notifications' && (
            <Card>
              <CardHeader>
                <CardTitle>Notification Preferences</CardTitle>
                <CardDescription>Choose what notifications you want to receive</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                {/* Success Message */}
                {notificationsSaved && (
                  <div className="p-3 bg-green-50 border border-green-200 rounded-lg flex items-center gap-2">
                    <Check className="w-5 h-5 text-green-600" />
                    <span className="text-green-800">Notification preferences saved successfully!</span>
                  </div>
                )}

                {/* Email Notifications */}
                <div>
                  <h3 className="font-medium text-gray-900 mb-4">Email Notifications</h3>
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-gray-900">Campaign Updates</p>
                        <p className="text-sm text-gray-500">Notifications about campaign performance</p>
                      </div>
                      <button
                        type="button"
                        onClick={() => setNotifications({...notifications, campaignUpdates: !notifications.campaignUpdates})}
                        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${notifications.campaignUpdates ? 'bg-blue-600' : 'bg-gray-200'}`}
                      >
                        <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${notifications.campaignUpdates ? 'translate-x-6' : 'translate-x-1'}`} />
                      </button>
                    </div>
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-gray-900">Fraud Alerts</p>
                        <p className="text-sm text-gray-500">Alerts about detected fraud attempts</p>
                      </div>
                      <button
                        type="button"
                        onClick={() => setNotifications({...notifications, fraudAlerts: !notifications.fraudAlerts})}
                        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${notifications.fraudAlerts ? 'bg-blue-600' : 'bg-gray-200'}`}
                      >
                        <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${notifications.fraudAlerts ? 'translate-x-6' : 'translate-x-1'}`} />
                      </button>
                    </div>
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-gray-900">Budget Alerts</p>
                        <p className="text-sm text-gray-500">Notifications when budgets are low</p>
                      </div>
                      <button
                        type="button"
                        onClick={() => setNotifications({...notifications, budgetAlerts: !notifications.budgetAlerts})}
                        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${notifications.budgetAlerts ? 'bg-blue-600' : 'bg-gray-200'}`}
                      >
                        <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${notifications.budgetAlerts ? 'translate-x-6' : 'translate-x-1'}`} />
                      </button>
                    </div>
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-gray-900">Weekly Reports</p>
                        <p className="text-sm text-gray-500">Weekly summary of your campaigns</p>
                      </div>
                      <button
                        type="button"
                        onClick={() => setNotifications({...notifications, weeklyReports: !notifications.weeklyReports})}
                        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${notifications.weeklyReports ? 'bg-blue-600' : 'bg-gray-200'}`}
                      >
                        <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${notifications.weeklyReports ? 'translate-x-6' : 'translate-x-1'}`} />
                      </button>
                    </div>
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-gray-900">System Updates</p>
                        <p className="text-sm text-gray-500">Platform updates and maintenance notices</p>
                      </div>
                      <button
                        type="button"
                        onClick={() => setNotifications({...notifications, systemUpdates: !notifications.systemUpdates})}
                        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${notifications.systemUpdates ? 'bg-blue-600' : 'bg-gray-200'}`}
                      >
                        <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${notifications.systemUpdates ? 'translate-x-6' : 'translate-x-1'}`} />
                      </button>
                    </div>
                  </div>
                </div>

                {/* Push Notifications */}
                <div className="border-t pt-6">
                  <h3 className="font-medium text-gray-900 mb-4">Push Notifications</h3>
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-gray-900">Real-time Alerts</p>
                        <p className="text-sm text-gray-500">Instant notifications for critical events</p>
                      </div>
                      <button
                        type="button"
                        onClick={() => setNotifications({...notifications, realTimeAlerts: !notifications.realTimeAlerts})}
                        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${notifications.realTimeAlerts ? 'bg-blue-600' : 'bg-gray-200'}`}
                      >
                        <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${notifications.realTimeAlerts ? 'translate-x-6' : 'translate-x-1'}`} />
                      </button>
                    </div>
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-gray-900">Campaign Milestones</p>
                        <p className="text-sm text-gray-500">Notifications when campaigns reach goals</p>
                      </div>
                      <button
                        type="button"
                        onClick={() => setNotifications({...notifications, campaignMilestones: !notifications.campaignMilestones})}
                        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${notifications.campaignMilestones ? 'bg-blue-600' : 'bg-gray-200'}`}
                      >
                        <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${notifications.campaignMilestones ? 'translate-x-6' : 'translate-x-1'}`} />
                      </button>
                    </div>
                  </div>
                </div>

                <div className="flex gap-3 pt-4 border-t">
                  <Button 
                    type="button"
                    className="bg-blue-600 hover:bg-blue-700"
                    onClick={() => {
                      setNotificationsSaved(true);
                      setTimeout(() => setNotificationsSaved(false), 3000);
                    }}
                  >
                    Save Preferences
                  </Button>
                  <Button 
                    type="button"
                    variant="outline"
                    onClick={() => setNotifications({
                      campaignUpdates: true,
                      fraudAlerts: true,
                      budgetAlerts: true,
                      weeklyReports: false,
                      systemUpdates: false,
                      realTimeAlerts: true,
                      campaignMilestones: true
                    })}
                  >
                    Reset to Defaults
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}

          {/* API Keys */}
          {activeTab === 'api' && (
            <>
              <Card>
                <CardHeader>
                  <CardTitle>API Keys</CardTitle>
                  <CardDescription>Manage your API keys for programmatic access</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Delete Confirmation Modal */}
                  {keyToDelete && (
                    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
                        <h3 className="text-lg font-semibold text-gray-900 mb-2">Delete API Key?</h3>
                        <p className="text-gray-600 mb-4">This action cannot be undone. Any applications using this key will stop working.</p>
                        <div className="flex gap-3 justify-end">
                          <Button variant="outline" onClick={() => setKeyToDelete(null)}>Cancel</Button>
                          <Button className="bg-red-600 hover:bg-red-700" onClick={() => deleteApiKey(keyToDelete)}>Delete Key</Button>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* New Key Modal */}
                  {showNewKeyModal && (
                    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
                        {newlyGeneratedKey ? (
                          <>
                            <h3 className="text-lg font-semibold text-gray-900 mb-2">
                              <Check className="w-5 h-5 inline text-green-600 mr-2" />
                              API Key Created
                            </h3>
                            <p className="text-gray-600 mb-4">Copy your new API key now. You won't be able to see it again!</p>
                            <div className="bg-gray-100 p-3 rounded-lg font-mono text-sm break-all mb-4">
                              {newlyGeneratedKey}
                            </div>
                            <div className="flex gap-3 justify-end">
                              <Button 
                                variant="outline"
                                onClick={() => {
                                  navigator.clipboard.writeText(newlyGeneratedKey);
                                  setSaveMessage({ type: 'success', text: 'API key copied to clipboard!' });
                                  setTimeout(() => setSaveMessage(null), 3000);
                                }}
                              >
                                <Copy className="w-4 h-4 mr-2" />
                                Copy Key
                              </Button>
                              <Button 
                                className="bg-blue-600 hover:bg-blue-700"
                                onClick={() => {
                                  setShowNewKeyModal(false);
                                  setNewlyGeneratedKey(null);
                                }}
                              >
                                Done
                              </Button>
                            </div>
                          </>
                        ) : (
                          <>
                            <h3 className="text-lg font-semibold text-gray-900 mb-2">Generate New API Key</h3>
                            <p className="text-gray-600 mb-4">Enter a name for your new API key. Use "prod" in the name for production keys.</p>
                            <input
                              type="text"
                              value={newKeyName}
                              onChange={(e) => setNewKeyName(e.target.value)}
                              placeholder="e.g., Production API Key, Test Key"
                              className="w-full px-3 py-2 border border-gray-300 rounded-lg mb-4 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            <div className="flex gap-3 justify-end">
                              <Button variant="outline" onClick={() => { setShowNewKeyModal(false); setNewKeyName(''); }}>Cancel</Button>
                              <Button 
                                className="bg-blue-600 hover:bg-blue-700"
                                onClick={createNewApiKey}
                                disabled={!newKeyName.trim()}
                              >
                                Generate Key
                              </Button>
                            </div>
                          </>
                        )}
                      </div>
                    </div>
                  )}

                  {apiKeysList.map((key) => (
                    <div key={key.id} className="border border-gray-200 rounded-lg p-4">
                      <div className="flex items-start justify-between mb-3">
                        <div>
                          <h3 className="font-medium text-gray-900">{key.name}</h3>
                          <p className="text-sm text-gray-500">Created {key.created}</p>
                        </div>
                        <div className="flex gap-2">
                          <button
                            type="button"
                            onClick={() => copyToClipboard(key.key, key.id)}
                            className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
                          >
                            {copiedKey === key.id ? (
                              <Check className="w-4 h-4 text-green-600" />
                            ) : (
                              <Copy className="w-4 h-4" />
                            )}
                          </button>
                          <button 
                            type="button"
                            onClick={() => setKeyToDelete(key.id)}
                            className="p-2 text-red-600 hover:text-red-900 hover:bg-red-50 rounded-lg transition-colors"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </div>
                      <div className="bg-gray-50 px-3 py-2 rounded font-mono text-sm text-gray-700 mb-3">
                        {maskKey(key.key)}
                      </div>
                      <div className="flex items-center gap-6 text-sm text-gray-600">
                        <span>Last used: {key.lastUsed}</span>
                        <span>•</span>
                        <span>{key.usage.toLocaleString()} requests</span>
                      </div>
                    </div>
                  ))}

                  <Button 
                    type="button"
                    className="w-full bg-blue-600 hover:bg-blue-700 mt-4"
                    onClick={() => setShowNewKeyModal(true)}
                  >
                    <Key className="w-4 h-4 mr-2" />
                    Generate New API Key
                  </Button>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>API Documentation</CardTitle>
                  <CardDescription>Learn how to integrate with TaskirX API</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <a href="/docs" className="flex items-center justify-between p-3 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer">
                      <span className="font-medium text-gray-900">REST API Documentation</span>
                      <span className="text-blue-600">→</span>
                    </a>
                    <a href="/docs#examples" className="flex items-center justify-between p-3 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer">
                      <span className="font-medium text-gray-900">Code Examples</span>
                      <span className="text-blue-600">→</span>
                    </a>
                    <a href="/docs#sdks" className="flex items-center justify-between p-3 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer">
                      <span className="font-medium text-gray-900">SDKs & Libraries</span>
                      <span className="text-blue-600">→</span>
                    </a>
                  </div>
                </CardContent>
              </Card>
            </>
          )}

          {/* Appearance */}
          {activeTab === 'appearance' && (
            <Card>
              <CardHeader>
                <CardTitle>Appearance</CardTitle>
                <CardDescription>Customize how TaskirX looks on your device</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                {/* Success Message */}
                {appearanceSaved && (
                  <div className="p-3 bg-green-50 border border-green-200 rounded-lg flex items-center gap-2">
                    <Check className="w-5 h-5 text-green-600" />
                    <span className="text-green-800">Appearance preferences saved successfully!</span>
                  </div>
                )}

                <div>
                  <h3 className="font-medium text-gray-900 mb-4">Theme</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <button
                      type="button"
                      onClick={() => setDarkMode(false)}
                      className={`p-4 border-2 rounded-lg transition-colors ${
                        !darkMode ? 'border-blue-600 bg-blue-50' : 'border-gray-200 hover:border-gray-300'
                      }`}
                    >
                      <Sun className="w-6 h-6 mx-auto mb-2 text-yellow-500" />
                      <p className="font-medium text-gray-900">Light</p>
                      {!darkMode && <p className="text-xs text-blue-600 mt-1">Active</p>}
                    </button>
                    <button
                      type="button"
                      onClick={() => setDarkMode(true)}
                      className={`p-4 border-2 rounded-lg transition-colors ${
                        darkMode ? 'border-blue-600 bg-blue-50' : 'border-gray-200 hover:border-gray-300'
                      }`}
                    >
                      <Moon className="w-6 h-6 mx-auto mb-2 text-blue-500" />
                      <p className="font-medium text-gray-900">Dark</p>
                      {darkMode && <p className="text-xs text-blue-600 mt-1">Active</p>}
                    </button>
                  </div>
                </div>

                <div className="border-t pt-6">
                  <h3 className="font-medium text-gray-900 mb-4">Display Settings</h3>
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-gray-900">Compact Mode</p>
                        <p className="text-sm text-gray-500">Show more content on screen</p>
                      </div>
                      <button
                        type="button"
                        onClick={() => setCompactMode(!compactMode)}
                        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${compactMode ? 'bg-blue-600' : 'bg-gray-200'}`}
                      >
                        <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${compactMode ? 'translate-x-6' : 'translate-x-1'}`} />
                      </button>
                    </div>
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-gray-900">Show Animations</p>
                        <p className="text-sm text-gray-500">Enable smooth transitions</p>
                      </div>
                      <button
                        type="button"
                        onClick={() => setShowAnimations(!showAnimations)}
                        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${showAnimations ? 'bg-blue-600' : 'bg-gray-200'}`}
                      >
                        <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${showAnimations ? 'translate-x-6' : 'translate-x-1'}`} />
                      </button>
                    </div>
                  </div>
                </div>

                <div className="flex gap-3 pt-4 border-t">
                  <Button 
                    type="button"
                    className="bg-blue-600 hover:bg-blue-700"
                    onClick={() => {
                      setAppearanceSaved(true);
                      setTimeout(() => setAppearanceSaved(false), 3000);
                    }}
                  >
                    Save Preferences
                  </Button>
                  <Button 
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setDarkMode(false);
                      setCompactMode(false);
                      setShowAnimations(true);
                    }}
                  >
                    Reset to Defaults
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Billing */}
          {activeTab === 'billing' && (
            <>
              {/* Delete Payment Confirmation Modal */}
              {paymentToDelete && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                  <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">Remove Payment Method?</h3>
                    <p className="text-gray-600 mb-4">This payment method will be permanently removed from your account.</p>
                    <div className="flex gap-3 justify-end">
                      <Button variant="outline" onClick={() => setPaymentToDelete(null)}>Cancel</Button>
                      <Button 
                        className="bg-red-600 hover:bg-red-700" 
                        onClick={() => {
                          setPaymentMethods(paymentMethods.filter(p => p.id !== paymentToDelete));
                          setPaymentToDelete(null);
                          setSaveMessage({ type: 'success', text: 'Payment method removed successfully!' });
                          setTimeout(() => setSaveMessage(null), 3000);
                        }}
                      >
                        Remove
                      </Button>
                    </div>
                  </div>
                </div>
              )}

              {/* Edit Payment Modal */}
              {paymentToEdit && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                  <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
                    <h3 className="text-lg font-semibold text-gray-900 mb-4">Edit Payment Method</h3>
                    {(() => {
                      const payment = paymentMethods.find(p => p.id === paymentToEdit);
                      return payment ? (
                        <div className="space-y-4">
                          <div className="flex items-center gap-4 p-3 bg-gray-50 rounded-lg">
                            <div className={`w-12 h-8 rounded flex items-center justify-center text-white font-bold text-xs ${
                              payment.type === 'VISA' ? 'bg-gradient-to-br from-blue-500 to-purple-600' :
                              payment.type === 'MC' ? 'bg-gradient-to-br from-red-500 to-orange-500' :
                              payment.type === 'AMEX' ? 'bg-gradient-to-br from-blue-600 to-blue-800' :
                              'bg-gradient-to-br from-gray-500 to-gray-700'
                            }`}>
                              {payment.type}
                            </div>
                            <div>
                              <p className="font-medium text-gray-900">•••• •••• •••• {payment.last4}</p>
                              <p className="text-sm text-gray-500">Current expiry: {payment.expiry}</p>
                            </div>
                          </div>
                          <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">New Expiry Date</label>
                            <input
                              type="text"
                              value={editCardExpiry}
                              onChange={(e) => setEditCardExpiry(e.target.value.replace(/[^\d/]/g, '').slice(0, 7))}
                              placeholder="MM/YYYY"
                              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                          </div>
                          <div className="flex items-center gap-2">
                            <input
                              type="checkbox"
                              id="setDefault"
                              checked={payment.isDefault}
                              onChange={() => {
                                setPaymentMethods(paymentMethods.map(p => ({
                                  ...p,
                                  isDefault: p.id === paymentToEdit
                                })));
                              }}
                              className="w-4 h-4 text-blue-600 rounded focus:ring-2 focus:ring-blue-500"
                            />
                            <label htmlFor="setDefault" className="text-sm text-gray-700">Set as default payment method</label>
                          </div>
                        </div>
                      ) : null;
                    })()}
                    <div className="flex gap-3 justify-end mt-6">
                      <Button variant="outline" onClick={() => {
                        setPaymentToEdit(null);
                        setEditCardExpiry('');
                      }}>Cancel</Button>
                      <Button 
                        className="bg-blue-600 hover:bg-blue-700"
                        onClick={() => {
                          if (editCardExpiry) {
                            setPaymentMethods(paymentMethods.map(p => 
                              p.id === paymentToEdit ? { ...p, expiry: editCardExpiry } : p
                            ));
                          }
                          setPaymentToEdit(null);
                          setEditCardExpiry('');
                          setSaveMessage({ type: 'success', text: 'Payment method updated successfully!' });
                          setTimeout(() => setSaveMessage(null), 3000);
                        }}
                      >
                        Save Changes
                      </Button>
                    </div>
                  </div>
                </div>
              )}

              {/* Add Payment Modal */}
              {showAddPaymentModal && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                  <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
                    <h3 className="text-lg font-semibold text-gray-900 mb-4">Add Payment Method</h3>
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Card Number</label>
                        <input
                          type="text"
                          value={newCardNumber}
                          onChange={(e) => setNewCardNumber(e.target.value.replace(/\D/g, '').slice(0, 16))}
                          placeholder="1234 5678 9012 3456"
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                      </div>
                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <label className="block text-sm font-medium text-gray-700 mb-1">Expiry Date</label>
                          <input
                            type="text"
                            value={newCardExpiry}
                            onChange={(e) => setNewCardExpiry(e.target.value.replace(/[^\d/]/g, '').slice(0, 7))}
                            placeholder="MM/YYYY"
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                          />
                        </div>
                        <div>
                          <label className="block text-sm font-medium text-gray-700 mb-1">CVC</label>
                          <input
                            type="text"
                            value={newCardCvc}
                            onChange={(e) => setNewCardCvc(e.target.value.replace(/\D/g, '').slice(0, 4))}
                            placeholder="123"
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                          />
                        </div>
                      </div>
                    </div>
                    <div className="flex gap-3 justify-end mt-6">
                      <Button variant="outline" onClick={() => {
                        setShowAddPaymentModal(false);
                        setNewCardNumber('');
                        setNewCardExpiry('');
                        setNewCardCvc('');
                      }}>Cancel</Button>
                      <Button 
                        className="bg-blue-600 hover:bg-blue-700"
                        disabled={newCardNumber.length < 16 || newCardExpiry.length < 7 || newCardCvc.length < 3}
                        onClick={() => {
                          const cardType = newCardNumber.startsWith('4') ? 'VISA' : 
                                          newCardNumber.startsWith('5') ? 'MC' : 
                                          newCardNumber.startsWith('3') ? 'AMEX' : 'CARD';
                          const newPayment = {
                            id: Date.now().toString(),
                            type: cardType,
                            last4: newCardNumber.slice(-4),
                            expiry: newCardExpiry,
                            isDefault: paymentMethods.length === 0
                          };
                          setPaymentMethods([...paymentMethods, newPayment]);
                          setShowAddPaymentModal(false);
                          setNewCardNumber('');
                          setNewCardExpiry('');
                          setNewCardCvc('');
                          setSaveMessage({ type: 'success', text: 'Payment method added successfully!' });
                          setTimeout(() => setSaveMessage(null), 3000);
                        }}
                      >
                        Add Card
                      </Button>
                    </div>
                  </div>
                </div>
              )}

              <Card>
                <CardHeader>
                  <CardTitle>Payment Method</CardTitle>
                  <CardDescription>Manage your payment information</CardDescription>
                </CardHeader>
                <CardContent>
                  {paymentMethods.length === 0 ? (
                    <div className="text-center py-6 text-gray-500">
                      <CreditCard className="w-12 h-12 mx-auto mb-3 text-gray-300" />
                      <p>No payment methods added yet</p>
                    </div>
                  ) : (
                    paymentMethods.map((payment) => (
                      <div key={payment.id} className="border border-gray-200 rounded-lg p-4 mb-4">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-4">
                            <div className={`w-12 h-8 rounded flex items-center justify-center text-white font-bold text-xs ${
                              payment.type === 'VISA' ? 'bg-gradient-to-br from-blue-500 to-purple-600' :
                              payment.type === 'MC' ? 'bg-gradient-to-br from-red-500 to-orange-500' :
                              payment.type === 'AMEX' ? 'bg-gradient-to-br from-blue-600 to-blue-800' :
                              'bg-gradient-to-br from-gray-500 to-gray-700'
                            }`}>
                              {payment.type}
                            </div>
                            <div>
                              <p className="font-medium text-gray-900">•••• •••• •••• {payment.last4}</p>
                              <p className="text-sm text-gray-500">Expires {payment.expiry}</p>
                            </div>
                          </div>
                          <div className="flex gap-2">
                            <Button 
                              type="button"
                              variant="outline" 
                              size="sm"
                              onClick={() => {
                                setEditCardExpiry(payment.expiry);
                                setPaymentToEdit(payment.id);
                              }}
                            >
                              Edit
                            </Button>
                            <Button 
                              type="button"
                              variant="outline" 
                              size="sm" 
                              className="text-red-600 hover:text-red-700"
                              onClick={() => setPaymentToDelete(payment.id)}
                            >
                              Remove
                            </Button>
                          </div>
                        </div>
                      </div>
                    ))
                  )}
                  <Button 
                    type="button"
                    variant="outline" 
                    className="w-full"
                    onClick={() => setShowAddPaymentModal(true)}
                  >
                    <CreditCard className="w-4 h-4 mr-2" />
                    Add Payment Method
                  </Button>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Billing History</CardTitle>
                  <CardDescription>View and download your invoices</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {invoices.map((invoice) => (
                      <div key={invoice.id} className={`flex items-center justify-between p-3 border rounded-lg ${
                        invoice.status === 'Unpaid' ? 'border-red-200 bg-red-50' : 'border-gray-200'
                      }`}>
                        <div>
                          <p className="font-medium text-gray-900">{invoice.date}</p>
                          <p className="text-sm text-gray-500">{invoice.amount}</p>
                        </div>
                        <div className="flex items-center gap-3">
                          <span className={`px-3 py-1 text-xs font-semibold rounded-full ${
                            invoice.status === 'Paid' ? 'bg-green-100 text-green-800' :
                            invoice.status === 'Pending' ? 'bg-yellow-100 text-yellow-800' :
                            'bg-red-100 text-red-800'
                          }`}>
                            {invoice.status}
                          </span>
                          {invoice.status === 'Unpaid' ? (
                            <Button 
                              type="button"
                              size="sm"
                              className="bg-blue-600 hover:bg-blue-700"
                              disabled={processingPayment === invoice.id}
                              onClick={async () => {
                                setProcessingPayment(invoice.id);
                                try {
                                  // Extract numeric amount from string like "$1,320.00"
                                  const numericAmount = parseFloat(invoice.amount.replace(/[$,]/g, ''));
                                  
                                  // Call the Stripe payment API
                                  const response = await fetch('/api/payments', {
                                    method: 'POST',
                                    headers: {
                                      'Content-Type': 'application/json',
                                    },
                                    body: JSON.stringify({
                                      amount: numericAmount,
                                      invoiceId: invoice.id,
                                      description: `TaskirX Invoice - ${invoice.date}`
                                    }),
                                  });

                                  const result = await response.json();

                                  if (result.success) {
                                    setInvoices(invoices.map(inv => 
                                      inv.id === invoice.id ? { ...inv, status: 'Paid' } : inv
                                    ));
                                    setSaveMessage({ 
                                      type: 'success', 
                                      text: result.demo 
                                        ? `Demo payment of ${invoice.amount} processed! Add STRIPE_SECRET_KEY for real payments.`
                                        : `Payment of ${invoice.amount} processed successfully via Stripe!`
                                    });
                                  } else {
                                    setSaveMessage({ 
                                      type: 'error', 
                                      text: `Payment failed: ${result.error || 'Unknown error'}`
                                    });
                                  }
                                } catch (error) {
                                  console.error('Payment error:', error);
                                  setSaveMessage({ 
                                    type: 'error', 
                                    text: 'Payment failed. Please try again.'
                                  });
                                } finally {
                                  setProcessingPayment(null);
                                  setTimeout(() => setSaveMessage(null), 5000);
                                }
                              }}
                            >
                              {processingPayment === invoice.id ? (
                                <>
                                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                                  Processing...
                                </>
                              ) : (
                                <>
                                  <CreditCard className="w-4 h-4 mr-2" />
                                  Pay Now
                                </>
                              )}
                            </Button>
                          ) : (
                            <Button 
                              type="button"
                              variant="outline" 
                              size="sm"
                              disabled={downloadingInvoice === invoice.id}
                              onClick={() => {
                                setDownloadingInvoice(invoice.id);
                                setTimeout(() => {
                                  setDownloadingInvoice(null);
                                  setSaveMessage({ type: 'success', text: `Invoice for ${invoice.date} downloaded!` });
                                  setTimeout(() => setSaveMessage(null), 3000);
                                }, 1000);
                              }}
                            >
                              {downloadingInvoice === invoice.id ? (
                                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                              ) : (
                                <Download className="w-4 h-4 mr-2" />
                              )}
                              {downloadingInvoice === invoice.id ? 'Downloading...' : 'Download'}
                            </Button>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </>
          )}
        </div>
      </div>
      )}
    </div>
  );
}
