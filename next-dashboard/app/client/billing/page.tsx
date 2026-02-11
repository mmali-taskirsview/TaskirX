'use client'

import { useState, useEffect } from 'react'
import { api } from '@/lib/api'
import { CreditCard, Download, Calendar, CheckCircle, AlertCircle, Clock, FileText, DollarSign, TrendingUp, Loader2 } from 'lucide-react'

interface Transaction {
  id: string;
  date: string;
  amount: number;
  type: string;
  description?: string;
  status: string;
}

interface Wallet {
  balance: number;
  creditLimit?: number;
  currency?: string;
}

const paymentMethods = [
  { id: 1, type: 'card', last4: '4242', brand: 'Visa', expiry: '12/27', isDefault: true },
  { id: 2, type: 'card', last4: '5555', brand: 'Mastercard', expiry: '08/26', isDefault: false },
]

export default function ClientBilling() {
  const [showAddCard, setShowAddCard] = useState(false)
  const [loading, setLoading] = useState(true)
  const [wallet, setWallet] = useState<Wallet | null>(null)
  const [transactions, setTransactions] = useState<Transaction[]>([])

  useEffect(() => {
    const fetchBillingData = async () => {
      try {
        // Fetch wallet balance
        const [walletRes, transRes] = await Promise.all([
          api.getWallet().catch(() => ({ data: null })),
          api.getTransactions(10).catch(() => ({ data: [] }))
        ])
        
        const walletData = walletRes.data || walletRes
        const transData = transRes.data || transRes || []
        
        setWallet(walletData || { balance: 45231, creditLimit: 100000 })
        
        // Transform transactions to invoice format
        const formattedTrans = transData.map((t: any, index: number) => ({
          id: t.id || `TXN-${Date.now()}-${index}`,
          date: t.createdAt || t.date || new Date().toISOString(),
          amount: Math.abs(Number(t.amount)) || 0,
          type: t.type || 'payment',
          description: t.description || 'Transaction',
          status: t.status || 'completed'
        }))
        
        setTransactions(formattedTrans.length > 0 ? formattedTrans : [
          { id: 'TXN-001', date: '2026-02-01', amount: 12450, type: 'deposit', description: 'Account Top-up', status: 'completed' },
          { id: 'TXN-002', date: '2026-01-15', amount: 8320, type: 'spend', description: 'Campaign Spend', status: 'completed' },
        ])
      } catch (error) {
        console.error('Failed to fetch billing data:', error)
        setWallet({ balance: 45231, creditLimit: 100000 })
      } finally {
        setLoading(false)
      }
    }
    
    fetchBillingData()
  }, [])

  const currentBalance = wallet?.balance || 0
  const creditLimit = wallet?.creditLimit || 100000
  const usedPercentage = creditLimit > 0 ? (currentBalance / creditLimit) * 100 : 0
  const totalSpend = transactions.filter(t => t.type === 'spend').reduce((sum, t) => sum + t.amount, 0)

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
        <span className="ml-2 text-gray-600">Loading billing data...</span>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Billing</h1>
          <p className="text-gray-500">Manage your billing and payment methods</p>
        </div>
        <button className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-white hover:bg-blue-700">
          <DollarSign className="h-5 w-5" />
          Add Funds
        </button>
      </div>

      {/* Balance Overview */}
      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 rounded-xl bg-gradient-to-br from-blue-600 to-blue-700 p-6 text-white">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-blue-100">Current Balance</p>
              <p className="mt-1 text-4xl font-bold">${currentBalance.toLocaleString()}</p>
            </div>
            <div className="rounded-xl bg-white/20 p-4">
              <CreditCard className="h-8 w-8" />
            </div>
          </div>
          <div className="mt-6">
            <div className="flex items-center justify-between text-sm">
              <span className="text-blue-100">Credit Used</span>
              <span className="text-white">${currentBalance.toLocaleString()} / ${creditLimit.toLocaleString()}</span>
            </div>
            <div className="mt-2 h-3 rounded-full bg-white/20">
              <div 
                className="h-3 rounded-full bg-white"
                style={{ width: `${usedPercentage}%` }}
              />
            </div>
          </div>
          <div className="mt-6 grid grid-cols-2 gap-4 border-t border-white/20 pt-4">
            <div>
              <p className="text-sm text-blue-100">This Month Spend</p>
              <p className="text-xl font-semibold">$12,450</p>
            </div>
            <div>
              <p className="text-sm text-blue-100">Last Month Spend</p>
              <p className="text-xl font-semibold">$15,780</p>
            </div>
          </div>
        </div>

        {/* Quick Stats */}
        <div className="space-y-4">
          <div className="rounded-xl bg-white p-5 shadow-sm">
            <div className="flex items-center gap-3">
              <div className="rounded-lg bg-green-100 p-2">
                <TrendingUp className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Total Spend (YTD)</p>
                <p className="text-xl font-bold text-gray-900">$46,000</p>
              </div>
            </div>
          </div>
          <div className="rounded-xl bg-white p-5 shadow-sm">
            <div className="flex items-center gap-3">
              <div className="rounded-lg bg-blue-100 p-2">
                <Calendar className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Next Invoice</p>
                <p className="text-xl font-bold text-gray-900">Feb 15, 2026</p>
              </div>
            </div>
          </div>
          <div className="rounded-xl bg-white p-5 shadow-sm">
            <div className="flex items-center gap-3">
              <div className="rounded-lg bg-purple-100 p-2">
                <CheckCircle className="h-5 w-5 text-purple-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Payment Status</p>
                <p className="text-xl font-bold text-green-600">All Paid</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Payment Methods */}
        <div className="rounded-xl bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">Payment Methods</h2>
            <button
              onClick={() => setShowAddCard(true)}
              className="text-sm font-medium text-blue-600 hover:text-blue-700"
            >
              + Add Card
            </button>
          </div>
          <div className="space-y-3">
            {paymentMethods.map((method) => (
              <div key={method.id} className={`flex items-center justify-between rounded-lg border p-4 ${method.isDefault ? 'border-blue-200 bg-blue-50' : 'border-gray-200'}`}>
                <div className="flex items-center gap-4">
                  <div className="flex h-10 w-14 items-center justify-center rounded bg-gray-100 text-xs font-bold text-gray-600">
                    {method.brand}
                  </div>
                  <div>
                    <p className="font-medium text-gray-900">•••• {method.last4}</p>
                    <p className="text-sm text-gray-500">Expires {method.expiry}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {method.isDefault && (
                    <span className="rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700">Default</span>
                  )}
                  <button className="text-sm text-gray-500 hover:text-gray-700">Edit</button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Billing Address */}
        <div className="rounded-xl bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">Billing Address</h2>
            <button className="text-sm font-medium text-blue-600 hover:text-blue-700">Edit</button>
          </div>
          <div className="rounded-lg border border-gray-200 p-4">
            <p className="font-medium text-gray-900">Demo Client Inc.</p>
            <p className="mt-1 text-sm text-gray-600">123 Business Street</p>
            <p className="text-sm text-gray-600">Suite 456</p>
            <p className="text-sm text-gray-600">Singapore 123456</p>
            <p className="mt-2 text-sm text-gray-600">Tax ID: SG123456789</p>
          </div>
        </div>
      </div>

      {/* Invoice History */}
      <div className="rounded-xl bg-white shadow-sm">
        <div className="border-b border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900">Invoice History</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200 bg-gray-50">
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">Transaction</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">Date</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">Amount</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">Type</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {transactions.map((transaction) => (
                <tr key={transaction.id} className="hover:bg-gray-50">
                  <td className="whitespace-nowrap px-6 py-4">
                    <div className="flex items-center gap-2">
                      <FileText className="h-4 w-4 text-gray-400" />
                      <span className="font-medium text-gray-900">{transaction.id}</span>
                    </div>
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-600">
                    {new Date(transaction.date).toLocaleDateString()}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm font-medium text-gray-900">
                    ${transaction.amount.toLocaleString()}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4">
                    <span className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium ${
                      transaction.status === 'completed' ? 'bg-green-100 text-green-700' :
                      transaction.status === 'pending' ? 'bg-yellow-100 text-yellow-700' :
                      'bg-blue-100 text-blue-700'
                    }`}>
                      {transaction.status === 'completed' && <CheckCircle className="h-3 w-3" />}
                      {transaction.status === 'pending' && <Clock className="h-3 w-3" />}
                      {transaction.status}
                    </span>
                  </td>
                  <td className="whitespace-nowrap px-6 py-4">
                    <button className="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700">
                      <Download className="h-4 w-4" /> Download
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
