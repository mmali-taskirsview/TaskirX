import { NextResponse } from 'next/server';
import { PHASE_PRODUCTION_BUILD } from 'next/constants';

export const dynamic = 'force-dynamic';

const publicBackendUrl = process.env.NEXT_PUBLIC_BACKEND_URL;
const normalizedPublicBackendUrl = publicBackendUrl ? publicBackendUrl.replace(/\/api$/, '') : undefined;
const BACKEND_URL = process.env.BACKEND_URL || normalizedPublicBackendUrl || 'http://taskir-nestjs:3000';
const isBuild = process.env.NEXT_PHASE === PHASE_PRODUCTION_BUILD;

export async function GET() {
  if (isBuild) {
    return NextResponse.json([]);
  }
  try {
    const response = await fetch(`${BACKEND_URL}/api/integrations/catalog`, {
      headers: { 'Content-Type': 'application/json' },
      cache: 'no-store',
    });

    if (!response.ok) {
      return NextResponse.json({ error: 'Failed to load integrations catalog' }, { status: 502 });
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    return NextResponse.json({ error: 'Integrations service unavailable' }, { status: 503 });
  }
}
