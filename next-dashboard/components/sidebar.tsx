'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  LayoutDashboard,
  Target,
  BarChart3,
  Settings,
  Shield,
  Zap,
  Users,
  FlaskConical,
  Image,
  Calculator,
  DollarSign,
  Building2,
  PieChart,
  ChevronDown,
  ChevronRight,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useState } from 'react'

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Campaigns', href: '/dashboard/campaigns', icon: Target },
  { name: 'Analytics', href: '/dashboard/analytics', icon: BarChart3 },
  { name: 'A/B Testing', href: '/dashboard/ab-testing', icon: FlaskConical },
  { name: 'Creatives', href: '/dashboard/creatives', icon: Image },
  { name: 'Fraud Detection', href: '/dashboard/fraud', icon: Shield },
  { name: 'Optimization', href: '/dashboard/optimization', icon: Zap },
  { name: 'ROI Calculator', href: '/dashboard/roi-calculator', icon: Calculator },
  { name: 'Users', href: '/dashboard/users', icon: Users },
  { name: 'Settings', href: '/dashboard/settings', icon: Settings },
]

const adminNavigation = [
  { name: 'Revenue Breakdown', href: '/dashboard/admin/revenue', icon: DollarSign },
  { name: 'Price Floors', href: '/dashboard/admin/price-floors', icon: PieChart },
  { name: 'Client Portfolio', href: '/dashboard/admin/clients', icon: Building2 },
]

export function Sidebar() {
  const pathname = usePathname()
  const [adminOpen, setAdminOpen] = useState(pathname.includes('/admin'))

  return (
    <div className="flex h-full w-64 flex-col bg-gray-900 text-white">
      {/* Logo */}
      <div className="flex h-16 items-center border-b border-gray-800 px-6">
        <div className="flex items-center space-x-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600">
            <span className="text-sm font-bold">T</span>
          </div>
          <span className="text-xl font-bold">TaskirX</span>
          <span className="rounded-full bg-blue-500/20 px-2 py-0.5 text-xs font-semibold text-blue-400">
            v3.0
          </span>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 px-3 py-4 overflow-y-auto">
        {navigation.map((item) => {
          const isActive = pathname === item.href
          return (
            <Link
              key={item.name}
              href={item.href}
              className={cn(
                'group flex items-center rounded-lg px-3 py-2 text-sm font-medium transition',
                isActive
                  ? 'bg-gray-800 text-white'
                  : 'text-gray-400 hover:bg-gray-800 hover:text-white'
              )}
            >
              <item.icon
                className={cn(
                  'mr-3 h-5 w-5 flex-shrink-0',
                  isActive ? 'text-blue-400' : 'text-gray-500 group-hover:text-gray-400'
                )}
              />
              {item.name}
            </Link>
          )
        })}

        {/* Admin Section */}
        <div className="pt-4 mt-4 border-t border-gray-800">
          <button
            onClick={() => setAdminOpen(!adminOpen)}
            className="w-full group flex items-center justify-between rounded-lg px-3 py-2 text-sm font-medium text-gray-400 hover:bg-gray-800 hover:text-white transition"
          >
            <div className="flex items-center">
              <Settings className="mr-3 h-5 w-5 flex-shrink-0 text-gray-500 group-hover:text-gray-400" />
              <span>Admin</span>
            </div>
            {adminOpen ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </button>
          {adminOpen && (
            <div className="ml-4 mt-1 space-y-1">
              {adminNavigation.map((item) => {
                const isActive = pathname === item.href
                return (
                  <Link
                    key={item.name}
                    href={item.href}
                    className={cn(
                      'group flex items-center rounded-lg px-3 py-2 text-sm font-medium transition',
                      isActive
                        ? 'bg-gray-800 text-white'
                        : 'text-gray-400 hover:bg-gray-800 hover:text-white'
                    )}
                  >
                    <item.icon
                      className={cn(
                        'mr-3 h-4 w-4 flex-shrink-0',
                        isActive ? 'text-blue-400' : 'text-gray-500 group-hover:text-gray-400'
                      )}
                    />
                    {item.name}
                  </Link>
                )
              })}
            </div>
          )}
        </div>
      </nav>

      {/* Footer */}
      <div className="border-t border-gray-800 p-4">
        <div className="flex items-center space-x-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-purple-600">
            <span className="text-sm font-bold">A</span>
          </div>
          <div className="flex-1">
            <p className="text-sm font-medium">Admin User</p>
            <p className="text-xs text-gray-400">admin@taskir.com</p>
          </div>
        </div>
      </div>
    </div>
  )
}
