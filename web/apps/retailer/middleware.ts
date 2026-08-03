import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { checkBFFRateLimit } from '@zippyra/rate-limit';

export async function middleware(request: NextRequest) {
  // 1. Apply BFF Edge Rate Limiting to all /api/* requests
  const rateLimitResponse = await checkBFFRateLimit(request);
  if (rateLimitResponse) {
    return rateLimitResponse;
  }

  const { pathname } = request.nextUrl;

  // 2. Role gating for staff management route (/dashboard/staff)
  if (pathname.startsWith('/dashboard/staff')) {
    const roleCookie = request.cookies.get('user_role')?.value;
    const roleHeader = request.headers.get('X-User-Role');
    const role = roleCookie || roleHeader;

    // Only MANAGER is allowed to access /dashboard/staff
    if (role && role !== 'MANAGER' && role !== 'SUPER_ADMIN') {
      const redirectUrl = new URL('/dashboard', request.url);
      redirectUrl.searchParams.set('error', 'UNAUTHORIZED_ROLE');
      return NextResponse.redirect(redirectUrl);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/api/:path*', '/dashboard/staff/:path*'],
};
