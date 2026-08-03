import { NextResponse } from 'next/server';
import { Client } from 'pg';

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const storeId = searchParams.get('store_id') || '33333333-3333-3333-3333-333333333333';
  const role = searchParams.get('role');
  const activeOnly = searchParams.get('active_only') !== 'false';

  const client = new Client({
    connectionString: process.env.DATABASE_URL || 'postgres://zippyra_user:zippyra_password@localhost:5436/zippyra?sslmode=disable',
  });

  try {
    await client.connect();

    let query = `
      SELECT id, store_id, chain_id, name, phone, role, is_active, 
             (pin_hash IS NOT NULL) AS has_pin_set, created_at
      FROM staff_members
      WHERE (store_id::text = $1 OR $1 = 'ALL')
    `;
    const params: any[] = [storeId];

    if (activeOnly) {
      query += ` AND is_active = true`;
    }
    if (role && role !== 'ALL') {
      params.push(role);
      query += ` AND role = $${params.length}`;
    }

    query += ` ORDER BY created_at DESC`;

    const res = await client.query(query, params);
    await client.end();

    return NextResponse.json({ staff: res.rows });
  } catch (err: any) {
    if (client) await client.end().catch(() => {});
    return NextResponse.json({ error: err.message }, { status: 500 });
  }
}

export async function POST(req: Request) {
  const body = await req.json();
  const name = body.name?.trim();
  const phone = body.phone?.trim();
  const role = body.role || 'CASHIER';
  const storeId = body.store_id || '33333333-3333-3333-3333-333333333333';
  const chainId = body.chain_id || '11111111-1111-1111-1111-111111111111';

  if (!name || !phone) {
    return NextResponse.json({ error: 'Name and phone are required' }, { status: 400 });
  }

  const client = new Client({
    connectionString: process.env.DATABASE_URL || 'postgres://zippyra_user:zippyra_password@localhost:5436/zippyra?sslmode=disable',
  });

  try {
    await client.connect();

    const insertQuery = `
      INSERT INTO staff_members (store_id, chain_id, name, phone, role, is_active)
      VALUES ($1, $2, $3, $4, $5, true)
      RETURNING id, store_id, chain_id, name, phone, role, is_active, false AS has_pin_set, created_at;
    `;

    const res = await client.query(insertQuery, [storeId, chainId, name, phone, role]);
    await client.end();

    return NextResponse.json(res.rows[0]);
  } catch (err: any) {
    if (client) await client.end().catch(() => {});
    if (err.code === '23505') { // Unique constraint violation on phone
      return NextResponse.json({ code: 'PHONE_ALREADY_STAFF', message: 'This mobile number is already registered as staff' }, { status: 409 });
    }
    return NextResponse.json({ error: err.message }, { status: 500 });
  }
}
