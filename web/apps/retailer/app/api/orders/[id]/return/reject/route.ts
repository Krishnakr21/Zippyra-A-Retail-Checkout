import { NextResponse } from 'next/server';

const ORDER_BASE = process.env.ORDER_SERVICE_URL || 'http://localhost:8085';

export async function POST(req: Request, { params }: { params: { id: string } }) {
  const body = await req.json();
  try {
    const backendRes = await fetch(`${ORDER_BASE}/v1/order/${params.id}/return/reject`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-User-ID': 'staff-01', 'X-User-Role': 'MANAGER', 'X-Store-ID': 'store-001' },
      body: JSON.stringify(body),
    });
    const data = await backendRes.json();
    return NextResponse.json(data, { status: backendRes.status });
  } catch (err: any) {
    return NextResponse.json({ order_id: params.id, status: 'RETURN_REJECTED', reason: body.reason });
  }
}
