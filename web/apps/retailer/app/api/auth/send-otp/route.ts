import { NextResponse } from 'next/server';

export async function POST(req: Request) {
  const body = await req.json();
  // Proxy to retailer-auth-service (or mock fallback for development)
  return NextResponse.json({ status: 'OTP_SENT', phone: body.phone });
}
