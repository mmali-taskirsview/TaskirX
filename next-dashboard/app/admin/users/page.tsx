'use client'

import { useState } from 'react'
import {
  Users,
  Search,
  Plus,
  Shield,
  Mail,
  MoreVertical,
  Edit,
  Trash2,
  Key,
  CheckCircle,
  Clock,
  XCircle,
  UserPlus,
} from 'lucide-react'

const users = [
  { id: 1, name: 'Admin User', email: 'admin@taskirx.com', role: 'Super Admin', status: 'active', lastLogin: '2 min ago', twoFactor: true },
  { id: 2, name: 'John Operations', email: 'john@taskirx.com', role: 'Admin', status: 'active', lastLogin: '1 hour ago', twoFactor: true },
  { id: 3, name: 'Sarah Analytics', email: 'sarah@taskirx.com', role: 'Analyst', status: 'active', lastLogin: '3 hours ago', twoFactor: false },
  { id: 4, name: 'Mike Support', email: 'mike@taskirx.com', role: 'Support', status: 'active', lastLogin: '1 day ago', twoFactor: true },
  { id: 5, name: 'Emily Finance', email: 'emily@taskirx.com', role: 'Finance', status: 'invited', lastLogin: 'Never', twoFactor: false },
  { id: 6, name: 'David Developer', email: 'david@taskirx.com', role: 'Developer', status: 'inactive', lastLogin: '2 weeks ago', twoFactor: false },
]

const roles = [
  { name: 'Super Admin', count: 1, permissions: 'Full access to all features' },
  { name: 'Admin', count: 2, permissions: 'Manage clients, users, and platform settings' },
  { name: 'Analyst', count: 3, permissions: 'View reports and analytics' },
  { name: 'Support', count: 4, permissions: 'Client support and ticket management' },
  { name: 'Finance', count: 2, permissions: 'Billing, invoices, and financial reports' },
  { name: 'Developer', count: 1, permissions: 'API access and integrations' },
]

export default function AdminUsers() {
  const [searchQuery, setSearchQuery] = useState('')
  const [roleFilter, setRoleFilter] = useState('All Roles')
  const [showInvite, setShowInvite] = useState(false)

  const filteredUsers = users.filter(user => {
    const matchesSearch = user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          user.email.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesRole = roleFilter === 'All Roles' || user.role === roleFilter
    return matchesSearch && matchesRole
  })

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <span className="flex items-center gap-1 rounded-full bg-green-900/50 px-2.5 py-0.5 text-xs font-medium text-green-400"><CheckCircle className="h-3 w-3" /> Active</span>
      case 'invited':
        return <span className="flex items-center gap-1 rounded-full bg-yellow-900/50 px-2.5 py-0.5 text-xs font-medium text-yellow-400"><Clock className="h-3 w-3" /> Invited</span>
      case 'inactive':
        return <span className="flex items-center gap-1 rounded-full bg-gray-700 px-2.5 py-0.5 text-xs font-medium text-gray-400"><XCircle className="h-3 w-3" /> Inactive</span>
      default:
        return null
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">User Management</h1>
          <p className="text-gray-400">Manage admin users and access permissions</p>
        </div>
        <button
          onClick={() => setShowInvite(true)}
          className="flex items-center gap-2 rounded-lg bg-purple-600 px-4 py-2 text-white hover:bg-purple-700"
        >
          <UserPlus className="h-5 w-5" />
          Invite User
        </button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-4">
        <div className="rounded-xl bg-gray-800 p-5">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-purple-900/50 p-2">
              <Users className="h-5 w-5 text-purple-400" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Total Users</p>
              <p className="text-2xl font-bold text-white">{users.length}</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-gray-800 p-5">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-green-900/50 p-2">
              <CheckCircle className="h-5 w-5 text-green-400" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Active</p>
              <p className="text-2xl font-bold text-white">{users.filter(u => u.status === 'active').length}</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-gray-800 p-5">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-blue-900/50 p-2">
              <Shield className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <p className="text-sm text-gray-400">2FA Enabled</p>
              <p className="text-2xl font-bold text-white">{users.filter(u => u.twoFactor).length}</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl bg-gray-800 p-5">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-yellow-900/50 p-2">
              <Mail className="h-5 w-5 text-yellow-400" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Pending Invites</p>
              <p className="text-2xl font-bold text-white">{users.filter(u => u.status === 'invited').length}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col gap-4 rounded-xl bg-gray-800 p-4 sm:flex-row sm:items-center">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-500" />
          <input
            type="text"
            placeholder="Search users..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full rounded-lg border border-gray-700 bg-gray-900 py-2 pl-10 pr-4 text-white placeholder-gray-500 focus:border-purple-500 focus:outline-none"
          />
        </div>
        <select
          value={roleFilter}
          onChange={(e) => setRoleFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-900 px-4 py-2 text-white focus:border-purple-500 focus:outline-none"
        >
          <option>All Roles</option>
          {roles.map(role => <option key={role.name} value={role.name}>{role.name}</option>)}
        </select>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Users Table */}
        <div className="lg:col-span-2 rounded-xl bg-gray-800">
          <div className="border-b border-gray-700 p-6">
            <h2 className="text-lg font-semibold text-white">Admin Users</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-700">
                  <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">User</th>
                  <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Role</th>
                  <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Status</th>
                  <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Last Login</th>
                  <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">2FA</th>
                  <th className="px-6 py-4 text-left text-xs font-medium uppercase text-gray-500">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700">
                {filteredUsers.map((user) => (
                  <tr key={user.id} className="hover:bg-gray-700/50">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-purple-900/50 font-bold text-purple-400">
                          {user.name.split(' ').map(n => n[0]).join('')}
                        </div>
                        <div>
                          <p className="font-medium text-white">{user.name}</p>
                          <p className="text-xs text-gray-500">{user.email}</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                        user.role === 'Super Admin' ? 'bg-red-900/50 text-red-400' :
                        user.role === 'Admin' ? 'bg-purple-900/50 text-purple-400' :
                        'bg-gray-700 text-gray-400'
                      }`}>
                        {user.role}
                      </span>
                    </td>
                    <td className="px-6 py-4">{getStatusBadge(user.status)}</td>
                    <td className="px-6 py-4 text-sm text-gray-400">{user.lastLogin}</td>
                    <td className="px-6 py-4">
                      {user.twoFactor ? (
                        <Shield className="h-5 w-5 text-green-400" />
                      ) : (
                        <Shield className="h-5 w-5 text-gray-600" />
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <button className="rounded p-1 text-gray-400 hover:bg-gray-700 hover:text-white">
                          <Edit className="h-4 w-4" />
                        </button>
                        <button className="rounded p-1 text-gray-400 hover:bg-gray-700 hover:text-white">
                          <Key className="h-4 w-4" />
                        </button>
                        <button className="rounded p-1 text-gray-400 hover:bg-gray-700 hover:text-red-400">
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Roles Panel */}
        <div className="rounded-xl bg-gray-800 p-6">
          <h2 className="text-lg font-semibold text-white mb-4">Roles & Permissions</h2>
          <div className="space-y-3">
            {roles.map((role) => (
              <div key={role.name} className="rounded-lg bg-gray-900 p-4">
                <div className="flex items-center justify-between mb-2">
                  <span className="font-medium text-white">{role.name}</span>
                  <span className="text-sm text-gray-500">{role.count} users</span>
                </div>
                <p className="text-xs text-gray-400">{role.permissions}</p>
              </div>
            ))}
          </div>
          <button className="mt-4 w-full rounded-lg border border-gray-700 px-4 py-2 text-sm text-gray-300 hover:bg-gray-700">
            Manage Roles
          </button>
        </div>
      </div>
    </div>
  )
}
