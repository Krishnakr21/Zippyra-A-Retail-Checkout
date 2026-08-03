import { NextResponse } from 'next/server';

export async function POST() {
  const res = NextResponse.json({ status: 'LOGGED_OUT' });
  res.cookies.delete('staff_session');
  return res;
}
