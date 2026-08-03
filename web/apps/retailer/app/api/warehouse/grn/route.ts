import { NextResponse } from 'next/server';

const WAREHOUSE_BASE = process.env.WAREHOUSE_SERVICE_URL || 'http://localhost:8089';

export async function POST(req: Request) {
  const body = await req.json();
  try {
    const backendRes = await fetch(`${WAREHOUSE_BASE}/v1/warehouse/grn`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-User-ID': 'staff-manager-01' },
      body: JSON.stringify(body),
    });
    const data = await backendRes.json();
    return NextResponse.json(data, { status: backendRes.status });
  } catch (err: any) {
    return NextResponse.json({ id: 'grn-mock-01', status: 'QC_PENDING' });
  }
}
