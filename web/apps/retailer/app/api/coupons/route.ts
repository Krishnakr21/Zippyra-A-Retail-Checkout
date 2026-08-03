import { NextRequest, NextResponse } from 'next/server';

const CART_SERVICE_URL = process.env.CART_SERVICE_URL || 'http://localhost:8084';

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const storeId = searchParams.get('store_id') || 'store-001';
    const chainId = searchParams.get('chain_id') || 'chain-001';

    const res = await fetch(`${CART_SERVICE_URL}/v1/cart/admin/coupons?chain_id=${chainId}&store_id=${storeId}&include_inactive=true`);
    if (res.ok) {
      const data = await res.json();
      return NextResponse.json(data);
    }
  } catch (err) {}

  // Fallback demo data
  return NextResponse.json({
    coupons: [
      {
        id: 'coupon-001',
        chain_id: 'chain-001',
        store_id: 'store-001',
        code: 'SAVE50',
        discount_type: 'FLAT_OFF',
        discount_value: 5000,
        min_cart_value_paise: 20000,
        max_uses: 100,
        max_uses_per_customer: 1,
        current_use_count: 14,
        is_active: true,
        active_from: new Date().toISOString(),
      },
      {
        id: 'coupon-002',
        chain_id: 'chain-001',
        store_id: 'store-001',
        code: 'WELCOME10',
        discount_type: 'PERCENT_OFF',
        discount_value: 10,
        min_cart_value_paise: 10000,
        max_uses_per_customer: 1,
        current_use_count: 42,
        is_active: true,
        active_from: new Date().toISOString(),
      },
    ],
  });
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const res = await fetch(`${CART_SERVICE_URL}/v1/cart/admin/coupons`, {
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
    coupon: {
      id: 'coupon-' + Date.now(),
      ...body,
      is_active: true,
      current_use_count: 0,
      active_from: new Date().toISOString(),
    },
  });
}
