import type { Metadata } from 'next'
// import { Inter } from 'next/font/google'
import './globals.css'

// Use a simpler type for Inter since we can't load the font module
const inter = { className: 'font-sans' }

export const metadata: Metadata = {
  title: 'TaskirX - Ad Exchange Platform',
  description: 'Enterprise-grade programmatic advertising platform with AI-powered optimization',
  keywords: ['ad exchange', 'programmatic advertising', 'RTB', 'DSP', 'SSP'],
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        {children}
      </body>
    </html>
  )
}
