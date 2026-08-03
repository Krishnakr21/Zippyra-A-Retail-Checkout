import { NextResponse } from 'next/server';
import { Client } from 'pg';

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const storeId = searchParams.get('store_id') || '33333333-3333-3333-3333-333333333333';
  const dateFrom = searchParams.get('date_from') || '2026-07-01';
  const dateTo = searchParams.get('date_to') || '2026-08-31';
  const granularity = (searchParams.get('granularity') || 'day').toLowerCase();

  const client = new Client({
    connectionString: process.env.DATABASE_URL || 'postgres://zippyra_user:zippyra_password@localhost:5436/zippyra?sslmode=disable',
  });

  try {
    await client.connect();

    let trunc = 'day';
    if (granularity === 'week') trunc = 'week';
    if (granularity === 'month') trunc = 'month';

    const query = `
      SELECT 
        TO_CHAR(DATE_TRUNC('${trunc}', created_at), 'YYYY-MM-DD') AS period,
        COALESCE(SUM(total_paise), 0) AS revenue_paise,
        COUNT(id) AS order_count,
        COALESCE(SUM(discount_paise), 0) AS discount_paise
      FROM orders
      WHERE created_at >= $1::timestamptz AND created_at <= ($2::date + INTERVAL '1 day')::timestamptz
      GROUP BY DATE_TRUNC('${trunc}', created_at)
      ORDER BY DATE_TRUNC('${trunc}', created_at) ASC;
    `;

    const res = await client.query(query, [dateFrom, dateTo]);
    await client.end();

    return NextResponse.json({
      data: res.rows.map(r => ({
        period: r.period,
        revenue_paise: Number(r.revenue_paise),
        order_count: Number(r.order_count),
        discount_paise: Number(r.discount_paise),
      })),
      granularity,
      store_id: storeId,
    });
  } catch (err: any) {
    if (client) await client.end().catch(() => {});
    return NextResponse.json({ error: err.message }, { status: 500 });
  }
}
