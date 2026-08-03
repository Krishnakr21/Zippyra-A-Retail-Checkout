import { NextResponse } from 'next/server';

export async function POST(req: Request) {
  const body = await req.json();
  const res = NextResponse.json({ status: 'SUCCESS' });
  res.cookies.set('staff_session', 'mock-staff-jwt-token-store-001', {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    path: '/',
  });
  return res;
}
