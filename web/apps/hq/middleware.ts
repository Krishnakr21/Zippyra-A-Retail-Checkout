import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { checkBFFRateLimit } from '@zippyra/rate-limit';

export async function middleware(request: NextRequest) {
  // 1. Apply BFF Edge Rate Limiting to all /api/* requests
  const rateLimitResponse = await checkBFFRateLimit(request);
  if (rateLimitResponse) {
    return rateLimitResponse;
  }

  const session = request.cookies.get('hq_session');
  const { pathname } = request.nextUrl;

  if (pathname.startsWith('/dashboard') && !session) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  if (pathname === '/login' && session) {
    return NextResponse.redirect(new URL('/dashboard', request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/api/:path*', '/dashboard/:path*', '/login'],
};
