import { NextResponse } from 'next/server';

const API_BASE = process.env.API_INTERNAL_BASE_URL || 'http://localhost:8081';

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const storeId = searchParams.get('store_id') || 'store-001';
  const barcode = searchParams.get('barcode');

  if (!barcode) {
    return NextResponse.json({ error: { code: 'INVALID_REQUEST', message: 'Barcode is required' } }, { status: 400 });
  }

  try {
    const backendRes = await fetch(`${API_BASE}/v1/catalog/barcode/${barcode}?store_id=${storeId}`);
    const data = await backendRes.json();
    return NextResponse.json(data, { status: backendRes.status });
  } catch (err: any) {
    return NextResponse.json({ error: { code: 'BACKEND_OFFLINE', message: 'Catalog service offline' } }, { status: 503 });
  }
}
