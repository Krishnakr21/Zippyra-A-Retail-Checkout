import { NextResponse } from 'next/server';

const WAREHOUSE_BASE = process.env.WAREHOUSE_SERVICE_URL || 'http://localhost:8089';

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const storeId = searchParams.get('store_id') || 'store-001';
  const status = searchParams.get('status');

  let url = `${WAREHOUSE_BASE}/v1/warehouse/po?store_id=${storeId}`;
  if (status) url += `&status=${status}`;

  try {
    const backendRes = await fetch(url);
    const data = await backendRes.json();
    return NextResponse.json(data, { status: backendRes.status });
  } catch (err: any) {
    return NextResponse.json({ items: [], page: 1 });
  }
}

export async function POST(req: Request) {
  const body = await req.json();
  try {
    const backendRes = await fetch(`${WAREHOUSE_BASE}/v1/warehouse/po`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-User-ID': 'staff-manager-01' },
      body: JSON.stringify(body),
    });
    const data = await backendRes.json();
    return NextResponse.json(data, { status: backendRes.status });
  } catch (err: any) {
    return NextResponse.json({ error: { code: 'WAREHOUSE_OFFLINE', message: 'Warehouse service offline' } }, { status: 503 });
  }
}
