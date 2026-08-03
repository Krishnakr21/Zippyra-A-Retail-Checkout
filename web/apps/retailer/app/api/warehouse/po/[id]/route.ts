import { NextResponse } from 'next/server';

const WAREHOUSE_BASE = process.env.WAREHOUSE_SERVICE_URL || 'http://localhost:8089';

export async function GET(req: Request, { params }: { params: { id: string } }) {
  try {
    const backendRes = await fetch(`${WAREHOUSE_BASE}/v1/warehouse/po/${params.id}`);
    const data = await backendRes.json();
    return NextResponse.json(data, { status: backendRes.status });
  } catch (err: any) {
    return NextResponse.json({ error: { code: 'WAREHOUSE_OFFLINE', message: 'Warehouse service offline' } }, { status: 503 });
  }
}
