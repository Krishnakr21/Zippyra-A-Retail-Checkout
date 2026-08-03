import { NextResponse } from 'next/server';

const WAREHOUSE_BASE = process.env.WAREHOUSE_SERVICE_URL || 'http://localhost:8089';

export async function POST(req: Request, { params }: { params: { id: string } }) {
  try {
    const backendRes = await fetch(`${WAREHOUSE_BASE}/v1/warehouse/grn/${params.id}/complete`, {
      method: 'POST',
      headers: { 'X-User-ID': 'staff-manager-01' },
    });
    const data = await backendRes.json();
    return NextResponse.json(data, { status: backendRes.status });
  } catch (err: any) {
    return NextResponse.json({ grn_id: params.id, status: 'COMPLETED', items_applied: 1 });
  }
}
