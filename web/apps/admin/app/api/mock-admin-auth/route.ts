import { NextResponse } from 'next/server';

/**
 * ⚠ MOCK AUTH ROUTE — FOR LOCAL DEV ONLY
 *
 * Accepts any email ending in ALLOWED_ADMIN_EMAIL_DOMAIN (default: @zippyra.com)
 * and any 6-digit TOTP code. Sets a simple cookie for session tracking.
 *
 * Replace with real admin-auth-service integration once built.
 */

const ALLOWED_DOMAIN = process.env.ALLOWED_ADMIN_EMAIL_DOMAIN || '@zippyra.com';

export async function POST(request: Request) {
  const body = await request.json();
  const { step, email, totp_code } = body;

  if (step === 'email') {
    if (!email || !email.endsWith(ALLOWED_DOMAIN)) {
      return NextResponse.json(
        { error: `Only ${ALLOWED_DOMAIN} emails are allowed` },
        { status: 403 }
      );
    }
    // In mock mode, any password works
    return NextResponse.json({ status: 'totp_required', email });
  }

  if (step === 'totp') {
    if (!totp_code || totp_code.length !== 6 || !/^\d{6}$/.test(totp_code)) {
      return NextResponse.json({ error: 'Invalid TOTP code format' }, { status: 400 });
    }

    // Mock: accept any 6-digit code
    const sessionData = {
      admin_id: 'mock-admin-001',
      email: email || 'admin@zippyra.com',
      name: 'Dev Admin',
      role: 'ADMIN',
      created_at: new Date().toISOString(),
    };

    const response = NextResponse.json({ status: 'authenticated', session: sessionData });
    response.cookies.set('admin_session', JSON.stringify(sessionData), {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 30 * 60, // 30 minutes
      path: '/',
    });

    return response;
  }

  return NextResponse.json({ error: 'Invalid step' }, { status: 400 });
}
