import { NextResponse } from 'next/server';

const ORDER_BASE = process.env.ORDER_SERVICE_URL || 'http://localhost:8085';

export async function GET(req: Request, { params }: { params: { id: string } }) {
  try {
    const backendRes = await fetch(`${ORDER_BASE}/v1/order/${params.id}`, {
      headers: { 'X-User-ID': 'staff-01', 'X-User-Role': 'MANAGER', 'X-Store-ID': 'store-001' },
    });
    const data = await backendRes.json();
    return NextResponse.json(data, { status: backendRes.status });
  } catch (err: any) {
    return NextResponse.json({ error: { code: 'ORDER_OFFLINE', message: 'Order service offline' } }, { status: 503 });
  }
}
