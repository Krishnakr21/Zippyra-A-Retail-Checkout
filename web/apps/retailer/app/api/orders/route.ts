import { NextResponse } from 'next/server';

const ORDER_BASE = process.env.ORDER_SERVICE_URL || 'http://localhost:8085';

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const storeId = searchParams.get('store_id') || 'store-001';
  const page = searchParams.get('page') || '1';

  try {
    const backendRes = await fetch(`${ORDER_BASE}/v1/order/store?store_id=${storeId}&page=${page}`, {
      headers: { 'X-User-ID': 'staff-01', 'X-User-Role': 'MANAGER', 'X-Store-ID': storeId },
    });
    const data = await backendRes.json();
    return NextResponse.json(data, { status: backendRes.status });
  } catch (err: any) {
    return NextResponse.json({ orders: [], page: 1 });
  }
}
