import { NextResponse } from 'next/server';

const WAREHOUSE_BASE = process.env.WAREHOUSE_SERVICE_URL || 'http://localhost:8089';

export async function PUT(req: Request, { params }: { params: { id: string } }) {
  const body = await req.json();
  try {
    const backendRes = await fetch(`${WAREHOUSE_BASE}/v1/warehouse/grn/${params.id}/qc`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    const data = await backendRes.json();
    return NextResponse.json(data, { status: backendRes.status });
  } catch (err: any) {
    return NextResponse.json({ grn_id: params.id, updated: 1 });
  }
}
