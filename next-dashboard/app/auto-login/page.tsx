'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'

export default function AutoLoginPage() {
  const router = useRouter()
  const [status, setStatus] = useState('Connecting to backend...')
  const [messages, setMessages] = useState<{text: string, type: string}[]>([])
  const [error, setError] = useState(false)

  const addMessage = (text: string, type: string) => {
    setMessages(prev => [...prev, { text, type }])
  }

  useEffect(() => {
    const login = async () => {
      try {
        addMessage('📡 Connecting to backend...', 'info')
        
        const response = await fetch('http://localhost:3000/api/auth/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
          },
          body: JSON.stringify({
            email: 'admin@taskirx.com',
            password: 'Admin123!'
          })
        })

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }

        const data = await response.json()
        addMessage('✅ Login successful!', 'success')
        addMessage(`👤 User: ${data.user.email}`, 'success')
        addMessage(`🎯 Role: ${data.user.role}`, 'success')

        // Save to localStorage
        localStorage.setItem('auth_token', data.access_token)
        localStorage.setItem('token', data.access_token)
        localStorage.setItem('user', JSON.stringify(data.user))

        addMessage('💾 Saved to localStorage', 'success')
        setStatus('✨ Success! Redirecting to dashboard...')

        // Redirect to dashboard
        setTimeout(() => {
          router.push('/dashboard')
        }, 1500)

      } catch (err: any) {
        console.error('Login error:', err)
        addMessage('❌ Login failed!', 'error')
        addMessage(err.message || 'Unknown error', 'error')
        setStatus('❌ Login failed. See details below.')
        setError(true)

        // Retry after 3 seconds
        setTimeout(() => {
          window.location.reload()
        }, 3000)
      }
    }

    login()
  }, [router])

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-500 to-blue-600 flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-2xl p-8 max-w-md w-full text-center">
        <h1 className="text-2xl font-bold text-gray-800 mb-6">🚀 TaskirX Auto Login</h1>
        
        {!error && (
          <div className="w-10 h-10 border-4 border-gray-200 border-t-purple-500 rounded-full animate-spin mx-auto mb-4" />
        )}
        
        <p className={`text-sm mb-4 ${error ? 'text-red-600' : 'text-gray-600'}`}>
          {status}
        </p>

        <div className="space-y-2 mb-4">
          {messages.map((msg, i) => (
            <div
              key={i}
              className={`text-sm p-2 rounded ${
                msg.type === 'success' ? 'bg-green-100 text-green-800' :
                msg.type === 'error' ? 'bg-red-100 text-red-800' :
                'bg-blue-100 text-blue-800'
              }`}
            >
              {msg.text}
            </div>
          ))}
        </div>

        <div className="text-left bg-gray-100 rounded p-3 text-xs text-gray-600">
          <p><strong>POST</strong> http://localhost:3000/api/auth/login</p>
          <p><strong>Email:</strong> admin@taskirx.com</p>
        </div>
      </div>
    </div>
  )
}
