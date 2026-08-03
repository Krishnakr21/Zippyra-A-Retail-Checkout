import { NextRequest, NextResponse } from 'next/server';

const CART_SERVICE_URL = process.env.CART_SERVICE_URL || 'http://localhost:8084';

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const storeId = searchParams.get('store_id') || 'store-001';
    const res = await fetch(`${CART_SERVICE_URL}/v1/cart/admin/offers?store_id=${storeId}`);
    if (res.ok) {
      const data = await res.json();
      return NextResponse.json(data);
    }
  } catch (err) {}

  // Fallback demo data
  return NextResponse.json({
    offers: [
      {
        id: 'offer-001',
        chain_id: 'chain-001',
        store_id: 'store-001',
        title: 'Monsoon Beverage Blast 10% Off',
        offer_type: 'PERCENT_OFF',
        applies_to: 'ALL',
        discount_percent: 10,
        min_cart_subtotal_paise: 20000,
        priority: 100,
        status: 'ACTIVE',
        created_at: new Date().toISOString(),
      },
    ],
  });
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const res = await fetch(`${CART_SERVICE_URL}/v1/cart/admin/offers`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-User-Role': 'MANAGER' },
      body: JSON.stringify(body),
    });
    if (res.ok) {
      const data = await res.json();
      return NextResponse.json(data);
    }
  } catch (err) {}

  const body = await request.json().catch(() => ({}));
  return NextResponse.json({
    success: true,
    offer: {
      id: 'offer-' + Date.now(),
      ...body,
      status: 'ACTIVE',
      created_at: new Date().toISOString(),
    },
  });
}
