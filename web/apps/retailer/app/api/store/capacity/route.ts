import { NextResponse } from 'next/server';

const STORE_BASE = process.env.STORE_SERVICE_URL || 'http://localhost:8082';

export async function PUT(req: Request) {
  const body = await req.json();
  const storeId = body.store_id || '33333333-3333-3333-3333-333333333333';
  try {
    const backendRes = await fetch(`${STORE_BASE}/v1/store/self-manage/stores/${storeId}/capacity`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', 'X-Chain-ID': '11111111-1111-1111-1111-111111111111', 'X-User-Role': 'MANAGER', 'X-Store-ID': storeId },
      body: JSON.stringify({
        capacity_max: body.capacity_max,
      }),
    });
    if (backendRes.ok) {
      const data = await backendRes.json();
      return NextResponse.json(data);
    }
  } catch (err: any) {}

  return NextResponse.json({ status: 'UPDATED', capacity_max: body.capacity_max });
}
