import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// Role-based route protection for Client and Admin dashboards
export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Get token from cookies (in production, validate JWT properly)
  const token = request.cookies.get('auth-token')?.value
  const userRole = request.cookies.get('user-role')?.value // 'advertiser' | 'publisher' | 'admin'

  // Public paths that don't require authentication
  const publicPaths = ['/login', '/register', '/forgot-password', '/']
  if (publicPaths.some(path => pathname === path || pathname.startsWith('/api/auth'))) {
    return NextResponse.next()
  }

  // Allow API routes to pass through without cookie-based auth redirects
  if (pathname.startsWith('/api')) {
    return NextResponse.next()
  }

  // Redirect to login if no token
  if (!token) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  // 1. Admin Routes
  if (pathname.startsWith('/admin')) {
     if (userRole !== 'admin') {
       return NextResponse.redirect(new URL('/login', request.url))
     }
    return NextResponse.next()
  }

  // 2. Advertiser Routes (Client)
  if (pathname.startsWith('/client') || pathname.startsWith('/dsp')) {
     if (userRole !== 'advertiser' && userRole !== 'admin') {
       return NextResponse.redirect(new URL('/login', request.url))
     }
    return NextResponse.next()
  }

  // 3. Publisher Routes
  if (pathname.startsWith('/publisher')) {
     if (userRole !== 'publisher' && userRole !== 'admin') {
       return NextResponse.redirect(new URL('/login', request.url))
     }
    return NextResponse.next()
  }

  // Legacy dashboard routes - allow access for backward compatibility
  if (pathname.startsWith('/dashboard')) {
    return NextResponse.next()
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    // Match all paths except static files and API
    '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
}
