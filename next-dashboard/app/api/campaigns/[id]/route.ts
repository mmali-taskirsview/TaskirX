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

// GET /api/campaigns/[id] - Get campaign by ID
export async function GET(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  if (isBuild) {
    return NextResponse.json({ error: 'Backend unavailable during build' }, { status: 503 });
  }
  try {
    let token = request.headers.get('authorization')?.replace('Bearer ', '');
    
    if (!token) {
      token = await getAuthToken();
    }
    
    const response = await fetch(`${BACKEND_URL}/api/campaigns/${params.id}`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (error) {
    if (!isBuild) {
      console.error('Error fetching campaign:', error);
    }
    return NextResponse.json({ error: 'Failed to fetch campaign' }, { status: 500 });
  }
}

// PUT /api/campaigns/[id] - Update campaign
export async function PUT(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  if (isBuild) {
    return NextResponse.json({ error: 'Backend unavailable during build' }, { status: 503 });
  }
  try {
    let token = request.headers.get('authorization')?.replace('Bearer ', '');
    
    if (!token) {
      token = await getAuthToken();
    }
    
    const body = await request.json();
    
    const response = await fetch(`${BACKEND_URL}/api/campaigns/${params.id}`, {
      method: 'PUT',
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
      console.error('Error updating campaign:', error);
    }
    return NextResponse.json({ error: 'Failed to update campaign' }, { status: 500 });
  }
}

// DELETE /api/campaigns/[id] - Delete campaign
export async function DELETE(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  if (isBuild) {
    return NextResponse.json({ error: 'Backend unavailable during build' }, { status: 503 });
  }
  try {
    let token = request.headers.get('authorization')?.replace('Bearer ', '');
    
    if (!token) {
      token = await getAuthToken();
    }
    
    const response = await fetch(`${BACKEND_URL}/api/campaigns/${params.id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    if (response.status === 204) {
      return new NextResponse(null, { status: 204 });
    }
    
    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (error) {
    if (!isBuild) {
      console.error('Error deleting campaign:', error);
    }
    return NextResponse.json({ error: 'Failed to delete campaign' }, { status: 500 });
  }
}
