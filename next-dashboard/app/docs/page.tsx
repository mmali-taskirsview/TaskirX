'use client'

import Link from 'next/link'
import { ArrowLeft, Book, Code, Server, Shield, Database, Zap, FileText } from 'lucide-react'

const docs = [
  {
    title: 'Getting Started',
    description: 'Quick start guide to set up and use TaskirX',
    icon: <Book className="w-6 h-6" />,
    sections: [
      'Installation',
      'Configuration',
      'First Campaign',
      'Dashboard Overview',
    ],
  },
  {
    title: 'API Reference',
    description: 'Complete REST API documentation',
    icon: <Code className="w-6 h-6" />,
    sections: [
      'Authentication',
      'Campaigns',
      'Analytics',
      'Billing',
    ],
  },
  {
    title: 'Backend Architecture',
    description: 'NestJS backend with TypeScript',
    icon: <Server className="w-6 h-6" />,
    sections: [
      'Module Structure',
      'Database Schema',
      'Authentication Flow',
      'Error Handling',
    ],
  },
  {
    title: 'AI & ML Services',
    description: 'Fraud detection, ad matching, bid optimization',
    icon: <Zap className="w-6 h-6" />,
    sections: [
      'Fraud Detection (Random Forest)',
      'Ad Matching (TF-IDF)',
      'Bid Optimization (Thompson)',
      'Real-time Processing',
    ],
  },
  {
    title: 'Security',
    description: 'Authentication, authorization, and data protection',
    icon: <Shield className="w-6 h-6" />,
    sections: [
      'JWT Authentication',
      'Role-based Access',
      'Data Encryption',
      'Rate Limiting',
    ],
  },
  {
    title: 'Database',
    description: 'PostgreSQL, Redis, and ClickHouse setup',
    icon: <Database className="w-6 h-6" />,
    sections: [
      'PostgreSQL Schema',
      'Redis Caching',
      'ClickHouse Analytics',
      'Migrations',
    ],
  },
]

export default function DocsPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50">
      {/* Header */}
      <header className="border-b bg-white/80 backdrop-blur-sm sticky top-0 z-50">
        <div className="container mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <Link href="/" className="flex items-center space-x-2 text-gray-600 hover:text-gray-900">
                <ArrowLeft className="w-5 h-5" />
                <span>Back</span>
              </Link>
              <div className="h-6 w-px bg-gray-300" />
              <div className="flex items-center space-x-2">
                <div className="w-8 h-8 bg-gradient-to-br from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
                  <FileText className="w-4 h-4 text-white" />
                </div>
                <span className="text-xl font-bold">Documentation</span>
              </div>
            </div>
            <nav className="flex items-center space-x-4">
              <Link 
                href="/dashboard"
                className="px-4 py-2 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg hover:shadow-lg transition"
              >
                Go to Dashboard
              </Link>
            </nav>
          </div>
        </div>
      </header>

      {/* Hero */}
      <section className="container mx-auto px-6 py-12">
        <div className="text-center max-w-3xl mx-auto">
          <h1 className="text-4xl font-bold mb-4 bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
            TaskirX Documentation
          </h1>
          <p className="text-lg text-gray-600">
            Everything you need to build, deploy, and scale your programmatic advertising platform
          </p>
        </div>
      </section>

      {/* Documentation Grid */}
      <section className="container mx-auto px-6 py-8 pb-20">
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {docs.map((doc) => (
            <div 
              key={doc.title}
              className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 hover:shadow-lg transition cursor-pointer"
            >
              <div className="flex items-center space-x-3 mb-4">
                <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-purple-500 rounded-lg flex items-center justify-center text-white">
                  {doc.icon}
                </div>
                <div>
                  <h3 className="text-lg font-semibold">{doc.title}</h3>
                  <p className="text-sm text-gray-500">{doc.description}</p>
                </div>
              </div>
              <ul className="space-y-2">
                {doc.sections.map((section) => (
                  <li key={section} className="flex items-center text-sm text-gray-600">
                    <span className="w-1.5 h-1.5 bg-blue-500 rounded-full mr-2" />
                    {section}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </section>

      {/* Quick Links */}
      <section className="bg-gray-900 text-white py-16">
        <div className="container mx-auto px-6">
          <h2 className="text-2xl font-bold mb-8 text-center">Quick Links</h2>
          <div className="grid md:grid-cols-4 gap-4 max-w-4xl mx-auto">
            <Link href="/dashboard" className="bg-gray-800 p-4 rounded-lg text-center hover:bg-gray-700 transition">
              Dashboard
            </Link>
            <Link href="/dashboard/campaigns" className="bg-gray-800 p-4 rounded-lg text-center hover:bg-gray-700 transition">
              Campaigns
            </Link>
            <Link href="/dashboard/analytics" className="bg-gray-800 p-4 rounded-lg text-center hover:bg-gray-700 transition">
              Analytics
            </Link>
            <Link href="/dashboard/settings" className="bg-gray-800 p-4 rounded-lg text-center hover:bg-gray-700 transition">
              Settings
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-gray-400 py-8 border-t border-gray-800">
        <div className="container mx-auto px-6 text-center">
          <p>© 2026 TaskirX. All rights reserved.</p>
        </div>
      </footer>
    </div>
  )
}
