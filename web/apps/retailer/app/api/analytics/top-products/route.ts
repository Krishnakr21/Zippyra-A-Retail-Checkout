import { NextResponse } from 'next/server';
import { Client } from 'pg';

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const storeId = searchParams.get('store_id') || '33333333-3333-3333-3333-333333333333';
  const limit = Number(searchParams.get('limit') || 10);

  const client = new Client({
    connectionString: process.env.DATABASE_URL || 'postgres://zippyra_user:zippyra_password@localhost:5436/zippyra?sslmode=disable',
  });

  try {
    await client.connect();

    const query = `
      SELECT 
        item->>'barcode' AS barcode,
        item->>'name' AS product_name,
        SUM((item->>'qty')::int) AS qty,
        SUM((item->>'total_price_paise')::bigint) AS line_total_paise
      FROM orders,
           jsonb_array_elements(items) AS item
      WHERE created_at >= NOW() - INTERVAL '60 days'
      GROUP BY item->>'barcode', item->>'name'
      ORDER BY line_total_paise DESC
      LIMIT $1;
    `;

    const res = await client.query(query, [limit]);
    await client.end();

    return NextResponse.json({
      products: res.rows.map(r => ({
        barcode: r.barcode,
        product_name: r.product_name,
        qty: Number(r.qty),
        line_total_paise: Number(r.line_total_paise),
      })),
      store_id: storeId,
    });
  } catch (err: any) {
    if (client) await client.end().catch(() => {});
    return NextResponse.json({ error: err.message }, { status: 500 });
  }
}
