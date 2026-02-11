import { NextRequest, NextResponse } from 'next/server';
import { PHASE_PRODUCTION_BUILD } from 'next/constants';

export const dynamic = 'force-dynamic';

const publicBackendUrl = process.env.NEXT_PUBLIC_BACKEND_URL;
const normalizedPublicBackendUrl = publicBackendUrl ? publicBackendUrl.replace(/\/api$/, '') : undefined;
const BACKEND_URL = process.env.BACKEND_URL || normalizedPublicBackendUrl || 'http://taskir-nestjs:3000';
const isBuild = process.env.NEXT_PHASE === PHASE_PRODUCTION_BUILD;

// Helper to get auth token
async function getAuthToken() {
  if (isBuild) {
    return null;
  }
  try {
    const response = await fetch(`${BACKEND_URL}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: 'admin@taskirx.com',
        password: 'Admin123!'
      })
    });
    if (response.ok) {
      const data = await response.json();
      return data.access_token;
    }
  } catch (error) {
    if (!isBuild) {
      console.error('Auth error:', error);
    }
  }
  return null;
}

// GET /api/campaigns - List all campaigns
export async function GET(request: NextRequest) {
  if (isBuild) {
    return NextResponse.json([]);
  }
  try {
    let token = request.headers.get('authorization')?.replace('Bearer ', '');
    
    if (!token) {
      token = await getAuthToken();
    }
    
    const response = await fetch(`${BACKEND_URL}/api/campaigns`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (error) {
    if (!isBuild) {
      console.error('Error fetching campaigns:', error);
    }
    return NextResponse.json({ error: 'Failed to fetch campaigns' }, { status: 500 });
  }
}

// POST /api/campaigns - Create a new campaign
export async function POST(request: NextRequest) {
  if (isBuild) {
    return NextResponse.json({ error: 'Backend unavailable during build' }, { status: 503 });
  }
  try {
    let token = request.headers.get('authorization')?.replace('Bearer ', '');
    
    if (!token) {
      token = await getAuthToken();
    }
    
    const body = await request.json();
    
    const response = await fetch(`${BACKEND_URL}/api/campaigns`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(body)
    });
    
    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (error) {
    if (!isBuild) {
      console.error('Error creating campaign:', error);
    }
    return NextResponse.json({ error: 'Failed to create campaign' }, { status: 500 });
  }
}
