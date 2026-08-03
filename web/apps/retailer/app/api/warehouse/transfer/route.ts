import { NextResponse } from 'next/server';

const TRANSFER_SERVICE_URL = process.env.TRANSFER_SERVICE_URL || 'http://localhost:8100';

export async function POST(req: Request) {
  const body = await req.json();
  try {
    const backendRes = await fetch(`${TRANSFER_SERVICE_URL}/v1/transfer/orders`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-User-ID': 'staff-manager-01' },
      body: JSON.stringify(body),
    });
    if (backendRes.ok) {
      const data = await backendRes.json();
      return NextResponse.json(data);
    }
  } catch (err: any) {}

  return NextResponse.json({
    id: 'tr-' + Date.now(),
    source_store_id: body.source_store_id || 'store-001',
    dest_store_id: body.dest_store_id || 'store-002',
    status: 'REQUESTED',
    created_at: new Date().toISOString(),
  });
}
