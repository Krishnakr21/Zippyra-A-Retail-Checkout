import { NextResponse } from 'next/server';

const INVENTORY_BASE = process.env.INVENTORY_SERVICE_URL || 'http://localhost:8084';

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const storeId = searchParams.get('store_id') || 'store-001';

  try {
    const backendRes = await fetch(`${INVENTORY_BASE}/v1/inventory/low-stock?store_id=${storeId}`);
    const data = await backendRes.json();
    return NextResponse.json(data, { status: backendRes.status });
  } catch (err: any) {
    return NextResponse.json({ items: [] });
  }
}
