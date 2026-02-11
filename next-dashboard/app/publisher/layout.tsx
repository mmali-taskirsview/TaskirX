'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  LayoutDashboard,
  Package,
  DollarSign,
  BarChart3,
  Settings,
  Shield,
  Wallet,
  Globe,
  Layers,
  TrendingUp,
  Bell,
  User,
  LogOut,
  ChevronDown,
  ChevronRight,
  Menu,
  X,
  Zap,
  FileText,
  PieChart
} from 'lucide-react';

// Publisher-specific navigation structure (SSP Focus)
const publisherNavigation = [
  {
    name: 'Dashboard',
    href: '/publisher',
    icon: LayoutDashboard,
  },
  {
    name: 'Inventory',
    icon: Package,
    children: [
      { name: 'Ad Units', href: '/publisher/inventory' },
      { name: 'Placements', href: '/publisher/placements' },
      { name: 'Ad Tags', href: '/publisher/ad-tags' },
    ]
  },
  {
    name: 'Yield Management',
    icon: TrendingUp,
    children: [
      { name: 'Floor Prices', href: '/publisher/floor-prices' },
      { name: 'Price Rules', href: '/publisher/price-rules' },
      { name: 'Demand Partners', href: '/publisher/demand-partners' },
    ]
  },
  {
    name: 'Analytics',
    icon: BarChart3,
    children: [
      { name: 'Performance', href: '/publisher/analytics' },
      { name: 'Fill Rate', href: '/publisher/fill-rate' },
      { name: 'Revenue Reports', href: '/publisher/reports' },
    ]
  },
  {
    name: 'Quality Controls',
    icon: Shield,
    children: [
      { name: 'Brand Safety', href: '/publisher/brand-safety' },
      { name: 'Ad Quality', href: '/publisher/ad-quality' },
      { name: 'Blocked Advertisers', href: '/publisher/blocked' },
    ]
  },
  {
    name: 'Payments',
    icon: Wallet,
    children: [
      { name: 'Earnings', href: '/publisher/earnings' },
      { name: 'Payouts', href: '/publisher/payouts' },
      { name: 'Payment Settings', href: '/publisher/payment-settings' },
    ]
  },
  {
    name: 'Settings',
    href: '/publisher/settings',
    icon: Settings,
  },
];

interface NavItemProps {
  item: {
    name: string;
    href?: string;
    icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
    children?: { name: string; href: string }[];
  };
  isCollapsed: boolean;
}

function NavItem({ item, isCollapsed }: NavItemProps) {
  const pathname = usePathname();
  const [isOpen, setIsOpen] = useState(false);
  const Icon = item.icon;

  const isActive = item.href
    ? pathname === item.href
    : item.children?.some((child) => pathname === child.href);

  if (item.children) {
    return (
      <div>
        <button
          onClick={() => setIsOpen(!isOpen)}
          className={`w-full flex items-center justify-between px-3 py-2 text-sm rounded-lg transition-colors ${
            isActive
              ? 'bg-emerald-50 text-emerald-700'
              : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
          }`}
        >
          <div className="flex items-center gap-3">
            <Icon className="w-5 h-5" />
            {!isCollapsed && <span>{item.name}</span>}
          </div>
          {!isCollapsed && (
            isOpen ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />
          )}
        </button>
        {isOpen && !isCollapsed && (
          <div className="ml-8 mt-1 space-y-1">
            {item.children.map((child) => (
              <Link
                key={child.href}
                href={child.href}
                className={`block px-3 py-2 text-sm rounded-lg transition-colors ${
                  pathname === child.href
                    ? 'bg-emerald-50 text-emerald-700 font-medium'
                    : 'text-gray-500 hover:bg-gray-50 hover:text-gray-700'
                }`}
              >
                {child.name}
              </Link>
            ))}
          </div>
        )}
      </div>
    );
  }

  return (
    <Link
      href={item.href!}
      className={`flex items-center gap-3 px-3 py-2 text-sm rounded-lg transition-colors ${
        isActive
          ? 'bg-emerald-50 text-emerald-700 font-medium'
          : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
      }`}
    >
      <Icon className="w-5 h-5" />
      {!isCollapsed && <span>{item.name}</span>}
    </Link>
  );
}

export default function PublisherLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Mobile menu button */}
      <div className="lg:hidden fixed top-4 left-4 z-50">
        <button
          onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          className="p-2 bg-white rounded-lg shadow-md"
        >
          {mobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
        </button>
      </div>

      {/* Sidebar */}
      <aside
        className={`fixed inset-y-0 left-0 z-40 bg-white border-r border-gray-200 transition-all duration-300 ${
          sidebarCollapsed ? 'w-16' : 'w-64'
        } ${mobileMenuOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}`}
      >
        {/* Logo */}
        <div className="h-16 flex items-center justify-between px-4 border-b border-gray-200">
          {!sidebarCollapsed && (
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 bg-gradient-to-br from-emerald-500 to-teal-600 rounded-lg flex items-center justify-center">
                <Globe className="w-5 h-5 text-white" />
              </div>
              <div>
                <span className="font-bold text-gray-900">TaskirX</span>
                <span className="text-xs text-emerald-600 block">Publisher SSP</span>
              </div>
            </div>
          )}
          <button
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            className="p-1.5 rounded-lg hover:bg-gray-100 hidden lg:block"
          >
            <Menu className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        {/* Navigation */}
        <nav className="p-3 space-y-1 overflow-y-auto h-[calc(100vh-8rem)]">
          {publisherNavigation.map((item) => (
            <NavItem key={item.name} item={item} isCollapsed={sidebarCollapsed} />
          ))}
        </nav>

        {/* User section */}
        <div className="absolute bottom-0 left-0 right-0 p-3 border-t border-gray-200 bg-white">
          {!sidebarCollapsed ? (
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 bg-emerald-100 rounded-full flex items-center justify-center">
                <User className="w-4 h-4 text-emerald-600" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-900 truncate">Publisher Demo</p>
                <p className="text-xs text-gray-500 truncate">pub_demo@taskirx.io</p>
              </div>
              <Link href="/login" className="p-1.5 rounded-lg hover:bg-gray-100">
                <LogOut className="w-4 h-4 text-gray-500" />
              </Link>
            </div>
          ) : (
            <Link href="/login" className="flex justify-center p-1.5 rounded-lg hover:bg-gray-100">
              <LogOut className="w-5 h-5 text-gray-500" />
            </Link>
          )}
        </div>
      </aside>

      {/* Main content */}
      <div className={`transition-all duration-300 ${sidebarCollapsed ? 'lg:ml-16' : 'lg:ml-64'}`}>
        {/* Top header */}
        <header className="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-6">
          <div className="flex items-center gap-4">
            <h1 className="text-lg font-semibold text-gray-900">Publisher Portal</h1>
            <span className="px-2 py-1 text-xs font-medium bg-emerald-100 text-emerald-700 rounded-full">
              SSP
            </span>
          </div>
          <div className="flex items-center gap-4">
            {/* Quick Stats */}
            <div className="hidden md:flex items-center gap-6 text-sm">
              <div className="flex items-center gap-2">
                <Zap className="w-4 h-4 text-emerald-500" />
                <span className="text-gray-600">Fill Rate:</span>
                <span className="font-semibold text-gray-900">94.2%</span>
              </div>
              <div className="flex items-center gap-2">
                <DollarSign className="w-4 h-4 text-emerald-500" />
                <span className="text-gray-600">eCPM:</span>
                <span className="font-semibold text-gray-900">$4.82</span>
              </div>
            </div>
            <button className="p-2 text-gray-500 hover:text-gray-700 relative">
              <Bell className="w-5 h-5" />
              <span className="absolute top-1 right-1 w-2 h-2 bg-emerald-500 rounded-full"></span>
            </button>
            <Link
              href="/"
              className="text-sm text-gray-600 hover:text-gray-900"
            >
              Switch Portal
            </Link>
          </div>
        </header>

        {/* Page content */}
        <main className="p-6">
          {children}
        </main>
      </div>

      {/* Mobile overlay */}
      {mobileMenuOpen && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 z-30 lg:hidden"
          onClick={() => setMobileMenuOpen(false)}
        />
      )}
    </div>
  );
}
