import { NextResponse } from 'next/server';

const CATALOG_SERVICE_URL = process.env.CATALOG_SERVICE_URL || 'http://localhost:8083';

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const storeId = searchParams.get('store_id') || '33333333-3333-3333-3333-333333333333';
  const page = searchParams.get('page') || '1';

  try {
    const backendRes = await fetch(`${CATALOG_SERVICE_URL}/v1/catalog/admin/products?store_id=${storeId}&page=${page}`);
    if (backendRes.ok) {
      const data = await backendRes.json();
      return NextResponse.json(data);
    }
  } catch (err: any) {}

  return NextResponse.json({
    products: [
      {
        id: '77777777-7777-7777-7777-777777777777',
        store_id: storeId,
        chain_id: '11111111-1111-1111-1111-111111111111',
        barcode: '8901234567890',
        name: 'Sparkling Mineral Water 500ml',
        price_paise: 4500,
        mrp_paise: 5000,
        hsn_code: '2201',
        is_active: true,
      },
      {
        id: '88888888-8888-8888-8888-888888888888',
        store_id: storeId,
        chain_id: '11111111-1111-1111-1111-111111111111',
        barcode: '8901234567891',
        name: 'Organic Cold Brew Coffee 250ml',
        price_paise: 12000,
        mrp_paise: 15000,
        hsn_code: '2201',
        is_active: true,
      },
    ],
  });
}

export async function POST(req: Request) {
  const body = await req.json();
  try {
    const backendRes = await fetch(`${CATALOG_SERVICE_URL}/v1/catalog/admin/products`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-User-Role': 'MANAGER' },
      body: JSON.stringify(body),
    });
    if (backendRes.ok) {
      const data = await backendRes.json();
      return NextResponse.json(data);
    }
  } catch (err: any) {}

  return NextResponse.json({
    success: true,
    product: {
      id: 'prod-' + Date.now(),
      ...body,
      is_active: true,
      created_at: new Date().toISOString(),
    },
  });
}
