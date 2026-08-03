import { NextResponse } from 'next/server';
import { Client } from 'pg';

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const storeId = searchParams.get('store_id') || '33333333-3333-3333-3333-333333333333';

  const client = new Client({
    connectionString: process.env.DATABASE_URL || 'postgres://zippyra_user:zippyra_password@localhost:5436/zippyra?sslmode=disable',
  });

  try {
    await client.connect();

    const query = `
      SELECT 
        t.id,
        t.source_store_id,
        t.dest_store_id,
        t.chain_id,
        t.status,
        t.requested_by,
        t.created_at,
        COALESCE(
          json_agg(
            json_build_object(
              'id', li.id,
              'barcode', li.barcode,
              'qty_requested', li.qty_requested,
              'qty_shipped', li.qty_shipped,
              'qty_received', li.qty_received
            )
          ) FILTER (WHERE li.id IS NOT NULL), '[]'
        ) AS line_items
      FROM transfer_orders t
      LEFT JOIN transfer_line_items li ON li.transfer_id = t.id
      WHERE t.source_store_id = $1 OR t.dest_store_id = $1
      GROUP BY t.id
      ORDER BY t.created_at DESC;
    `;

    const res = await client.query(query, [storeId]);
    await client.end();

    return NextResponse.json({ transfers: res.rows });
  } catch (err: any) {
    if (client) await client.end().catch(() => {});
    return NextResponse.json({ error: err.message }, { status: 500 });
  }
}

export async function POST(req: Request) {
  const body = await req.json();
  const sourceStoreId = body.source_store_id || '33333333-3333-3333-3333-333333333333';
  const destStoreId = body.dest_store_id || '44444444-4444-4444-4444-444444444444';
  const chainId = body.chain_id || '11111111-1111-1111-1111-111111111111';
  const lineItems = body.line_items || [];

  const client = new Client({
    connectionString: process.env.DATABASE_URL || 'postgres://zippyra_user:zippyra_password@localhost:5436/zippyra?sslmode=disable',
  });

  try {
    await client.connect();

    const insertOrderQuery = `
      INSERT INTO transfer_orders (source_store_id, dest_store_id, chain_id, status, requested_by)
      VALUES ($1, $2, $3, 'REQUESTED', '99999999-9999-9999-9999-999999999901')
      RETURNING id, source_store_id, dest_store_id, chain_id, status, requested_by, created_at;
    `;

    const res = await client.query(insertOrderQuery, [sourceStoreId, destStoreId, chainId]);
    const transferOrder = res.rows[0];

    for (const item of lineItems) {
      await client.query(`
        INSERT INTO transfer_line_items (transfer_id, barcode, qty_requested)
        VALUES ($1, $2, $3)
      `, [transferOrder.id, item.barcode || '8901030012345', item.qty_requested || 10]);
    }

    await client.end();
    return NextResponse.json({ ...transferOrder, line_items: lineItems });
  } catch (err: any) {
    if (client) await client.end().catch(() => {});
    return NextResponse.json({ error: err.message }, { status: 500 });
  }
}
