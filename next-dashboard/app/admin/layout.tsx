'use client'

import { useState } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  LayoutDashboard,
  Users,
  DollarSign,
  TrendingUp,
  Building2,
  Settings,
  Bell,
  User,
  ChevronDown,
  Menu,
  X,
  LogOut,
  Shield,
  Server,
  Activity,
  FileText,
  AlertTriangle,
  Database,
  Key,
  Globe,
} from 'lucide-react'

const adminNavigation = [
  { name: 'Platform Overview', href: '/admin', icon: LayoutDashboard },
  { 
    name: 'Client Management',
    icon: Building2,
    children: [
      { name: 'All Clients', href: '/admin/clients' },
      { name: 'Client Onboarding', href: '/admin/clients/onboarding' },
      { name: 'Client Health', href: '/admin/clients/health' },
    ]
  },
  {
    name: 'User Management',
    icon: Users,
    children: [
      { name: 'All Users', href: '/admin/users' },
      { name: 'Roles & Permissions', href: '/admin/users/roles' },
      { name: 'Access Logs', href: '/admin/users/logs' },
    ]
  },
  {
    name: 'Revenue & Finance',
    icon: DollarSign,
    children: [
      { name: 'Revenue Dashboard', href: '/admin/revenue' },
      { name: 'Billing Management', href: '/admin/revenue/billing' },
      { name: 'Payouts', href: '/admin/revenue/payouts' },
    ]
  },
  {
    name: 'Yield Optimization',
    icon: TrendingUp,
    children: [
      { name: 'Price Floors', href: '/admin/price-floors' },
      { name: 'Demand Partners', href: '/admin/demand-partners' },
      { name: 'Supply Quality', href: '/admin/supply-quality' },
    ]
  },
  {
    name: 'Platform Operations',
    icon: Server,
    children: [
      { name: 'System Health', href: '/admin/system' },
      { name: 'API Keys', href: '/admin/api-keys' },
      { name: 'Integrations', href: '/admin/integrations' },
    ]
  },
  {
    name: 'Security & Fraud',
    icon: Shield,
    children: [
      { name: 'Fraud Detection', href: '/admin/fraud' },
      { name: 'Block Lists', href: '/admin/blocklists' },
      { name: 'Security Logs', href: '/admin/security' },
    ]
  },
  { name: 'Reports', href: '/admin/reports', icon: FileText },
  { name: 'Settings', href: '/admin/settings', icon: Settings },
]

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const pathname = usePathname()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [userMenuOpen, setUserMenuOpen] = useState(false)
  const [expandedMenus, setExpandedMenus] = useState<string[]>(['Client Management', 'Revenue & Finance'])

  const toggleMenu = (name: string) => {
    setExpandedMenus(prev => 
      prev.includes(name) 
        ? prev.filter(m => m !== name)
        : [...prev, name]
    )
  }

  const isActiveLink = (href: string) => {
    return pathname === href || (href !== '/admin' && pathname.startsWith(href))
  }

  return (
    <div className="flex h-screen bg-gray-900">
      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-40 bg-black bg-opacity-50 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={`fixed inset-y-0 left-0 z-50 w-72 transform bg-gray-900 shadow-xl transition-transform duration-300 ease-in-out lg:static lg:translate-x-0 ${
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        {/* Logo */}
        <div className="flex h-16 items-center justify-between border-b border-gray-800 px-6">
          <Link href="/admin" className="flex items-center space-x-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-purple-500 to-purple-600">
              <span className="text-lg font-bold text-white">T</span>
            </div>
            <span className="text-xl font-bold text-white">TaskirX</span>
          </Link>
          <button
            onClick={() => setSidebarOpen(false)}
            className="rounded-lg p-1 text-gray-400 hover:bg-gray-800 lg:hidden"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Admin Badge */}
        <div className="border-b border-gray-800 px-6 py-3">
          <div className="flex items-center space-x-2 rounded-lg bg-purple-900/50 px-3 py-2">
            <Shield className="h-4 w-4 text-purple-400" />
            <span className="text-sm font-medium text-purple-300">Admin Console</span>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 space-y-1 overflow-y-auto px-3 py-4">
          {adminNavigation.map((item) => {
            if ('children' in item && item.children) {
              const isExpanded = expandedMenus.includes(item.name)
              const hasActiveChild = item.children.some(child => isActiveLink(child.href))
              
              return (
                <div key={item.name}>
                  <button
                    onClick={() => toggleMenu(item.name)}
                    className={`flex w-full items-center justify-between rounded-lg px-3 py-2.5 text-sm font-medium transition-colors ${
                      hasActiveChild
                        ? 'bg-gray-800 text-white'
                        : 'text-gray-400 hover:bg-gray-800 hover:text-white'
                    }`}
                  >
                    <div className="flex items-center space-x-3">
                      <item.icon className={`h-5 w-5 ${hasActiveChild ? 'text-purple-400' : ''}`} />
                      <span>{item.name}</span>
                    </div>
                    <ChevronDown className={`h-4 w-4 transition-transform ${isExpanded ? 'rotate-180' : ''}`} />
                  </button>
                  
                  {isExpanded && (
                    <div className="ml-8 mt-1 space-y-1">
                      {item.children.map((child) => (
                        <Link
                          key={child.href}
                          href={child.href}
                          className={`block rounded-lg px-3 py-2 text-sm transition-colors ${
                            isActiveLink(child.href)
                              ? 'bg-purple-900/50 text-purple-300'
                              : 'text-gray-500 hover:bg-gray-800 hover:text-gray-300'
                          }`}
                        >
                          {child.name}
                        </Link>
                      ))}
                    </div>
                  )}
                </div>
              )
            }

            const isActive = isActiveLink(item.href)
            return (
              <Link
                key={item.name}
                href={item.href}
                className={`flex items-center space-x-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-purple-900/50 text-purple-300'
                    : 'text-gray-400 hover:bg-gray-800 hover:text-white'
                }`}
              >
                <item.icon className={`h-5 w-5 ${isActive ? 'text-purple-400' : ''}`} />
                <span>{item.name}</span>
              </Link>
            )
          })}
        </nav>

        {/* System Status */}
        <div className="border-t border-gray-800 p-4">
          <div className="rounded-lg bg-gray-800 p-3">
            <div className="flex items-center justify-between">
              <span className="text-xs text-gray-400">System Status</span>
              <span className="flex items-center text-xs text-green-400">
                <span className="mr-1.5 h-2 w-2 rounded-full bg-green-400" />
                All Systems Operational
              </span>
            </div>
            <div className="mt-2 grid grid-cols-3 gap-2 text-center">
              <div>
                <p className="text-lg font-semibold text-white">99.9%</p>
                <p className="text-xs text-gray-500">Uptime</p>
              </div>
              <div>
                <p className="text-lg font-semibold text-white">23ms</p>
                <p className="text-xs text-gray-500">Latency</p>
              </div>
              <div>
                <p className="text-lg font-semibold text-white">1.2M</p>
                <p className="text-xs text-gray-500">QPS</p>
              </div>
            </div>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Header */}
        <header className="flex h-16 items-center justify-between border-b border-gray-800 bg-gray-900 px-6">
          <div className="flex items-center space-x-4">
            <button
              onClick={() => setSidebarOpen(true)}
              className="rounded-lg p-2 text-gray-400 hover:bg-gray-800 lg:hidden"
            >
              <Menu className="h-5 w-5" />
            </button>
            <div className="flex items-center space-x-2">
              <Globe className="h-5 w-5 text-gray-500" />
              <select className="bg-transparent text-sm text-gray-300 focus:outline-none">
                <option value="all">All Regions</option>
                <option value="sea">Southeast Asia</option>
                <option value="na">North America</option>
                <option value="eu">Europe</option>
              </select>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            {/* Alerts */}
            <button className="relative rounded-lg p-2 text-gray-400 hover:bg-gray-800">
              <AlertTriangle className="h-5 w-5" />
              <span className="absolute right-1 top-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-xs text-white">
                3
              </span>
            </button>

            {/* Activity */}
            <button className="rounded-lg p-2 text-gray-400 hover:bg-gray-800">
              <Activity className="h-5 w-5" />
            </button>

            {/* Notifications */}
            <button className="relative rounded-lg p-2 text-gray-400 hover:bg-gray-800">
              <Bell className="h-5 w-5" />
              <span className="absolute right-1 top-1 h-2 w-2 rounded-full bg-purple-500" />
            </button>

            {/* Admin User Menu */}
            <div className="relative">
              <button
                onClick={() => setUserMenuOpen(!userMenuOpen)}
                className="flex items-center space-x-3 rounded-lg p-2 hover:bg-gray-800"
              >
                <div className="flex h-8 w-8 items-center justify-center rounded-full bg-purple-900">
                  <Shield className="h-4 w-4 text-purple-400" />
                </div>
                <div className="hidden text-left md:block">
                  <p className="text-sm font-medium text-white">Admin User</p>
                  <p className="text-xs text-gray-500">Super Admin</p>
                </div>
                <ChevronDown className="h-4 w-4 text-gray-400" />
              </button>

              {userMenuOpen && (
                <div className="absolute right-0 mt-2 w-56 rounded-lg bg-gray-800 py-1 shadow-lg ring-1 ring-gray-700">
                  <div className="border-b border-gray-700 px-4 py-3">
                    <p className="text-sm font-medium text-white">admin@taskirx.com</p>
                    <p className="text-xs text-gray-400">Super Administrator</p>
                  </div>
                  <Link
                    href="/admin/settings"
                    className="flex items-center space-x-2 px-4 py-2 text-sm text-gray-300 hover:bg-gray-700"
                  >
                    <Settings className="h-4 w-4" />
                    <span>Admin Settings</span>
                  </Link>
                  <Link
                    href="/admin/api-keys"
                    className="flex items-center space-x-2 px-4 py-2 text-sm text-gray-300 hover:bg-gray-700"
                  >
                    <Key className="h-4 w-4" />
                    <span>API Keys</span>
                  </Link>
                  <Link
                    href="/admin/security"
                    className="flex items-center space-x-2 px-4 py-2 text-sm text-gray-300 hover:bg-gray-700"
                  >
                    <Shield className="h-4 w-4" />
                    <span>Security</span>
                  </Link>
                  <hr className="my-1 border-gray-700" />
                  <button
                    onClick={() => window.location.href = '/login'}
                    className="flex w-full items-center space-x-2 px-4 py-2 text-sm text-red-400 hover:bg-gray-700"
                  >
                    <LogOut className="h-4 w-4" />
                    <span>Sign out</span>
                  </button>
                </div>
              )}
            </div>
          </div>
        </header>

        {/* Main Content */}
        <main className="flex-1 overflow-y-auto bg-gray-950 p-6">
          {children}
        </main>
      </div>
    </div>
  )
}
