import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export interface RateLimitResult {
  success: boolean;
  limit: number;
  remaining: number;
  resetSeconds: number;
}

interface WindowEntry {
  timestamps: number[];
}

// In-Memory Sliding Window Store for Edge / Serverless Execution
class MemorySlidingWindowStore {
  private store = new Map<string, WindowEntry>();

  check(key: string, limit: number, windowSeconds: number): RateLimitResult {
    const now = Date.now();
    const windowMs = windowSeconds * 1000;
    const windowStart = now - windowMs;

    let entry = this.store.get(key);
    if (!entry) {
      entry = { timestamps: [] };
      this.store.set(key, entry);
    }

    // Filter out timestamps outside current window
    entry.timestamps = entry.timestamps.filter((ts) => ts > windowStart);

    if (entry.timestamps.length >= limit) {
      const oldestInWindow = entry.timestamps[0];
      const resetSeconds = Math.max(1, Math.ceil((oldestInWindow + windowMs - now) / 1000));
      return {
        success: false,
        limit,
        remaining: 0,
        resetSeconds,
      };
    }

    entry.timestamps.push(now);
    const remaining = limit - entry.timestamps.length;
    return {
      success: true,
      limit,
      remaining,
      resetSeconds: windowSeconds,
    };
  }

  clear() {
    this.store.clear();
  }
}

export const memoryRateLimiter = new MemorySlidingWindowStore();

export function getClientIP(request: NextRequest): string {
  const xff = request.headers.get('x-forwarded-for');
  if (xff) {
    return xff.split(',')[0].trim();
  }
  const xreal = request.headers.get('x-real-ip');
  if (xreal) {
    return xreal.trim();
  }
  return '127.0.0.1';
}

export function isAuthAdjacentRoute(pathname: string): boolean {
  const lower = pathname.toLowerCase();
  return (
    lower.startsWith('/api/auth') ||
    lower.startsWith('/api/otp') ||
    lower.startsWith('/api/login') ||
    lower.includes('/staff/otp') ||
    lower.includes('/auth/send-otp') ||
    lower.includes('/auth/verify-otp')
  );
}

/**
 * Main BFF Edge Rate Limiting Middleware Function
 */
export async function checkBFFRateLimit(request: NextRequest): Promise<NextResponse | null> {
  const { pathname } = request.nextUrl;

  // Only apply rate limiting to /api/* BFF routes
  if (!pathname.startsWith('/api')) {
    return null;
  }

  const clientIP = getClientIP(request);
  const isAuth = isAuthAdjacentRoute(pathname);

  // Rate Limits:
  // - Auth-adjacent BFF routes: 10 requests per 60 seconds
  // - General authenticated BFF routes: 120 requests per 60 seconds
  const limit = isAuth ? 10 : 120;
  const windowSeconds = 60;
  const key = isAuth ? `ratelimit:auth:${clientIP}` : `ratelimit:general:${clientIP}`;

  const res = memoryRateLimiter.check(key, limit, windowSeconds);

  if (!res.success) {
    return new NextResponse(
      JSON.stringify({
        error: {
          code: 'TOO_MANY_REQUESTS',
          message: 'Too many requests. Please try again shortly.',
          retry_after_seconds: res.resetSeconds,
        },
      }),
      {
        status: 429,
        headers: {
          'Content-Type': 'application/json',
          'Retry-After': String(res.resetSeconds),
          'X-RateLimit-Limit': String(res.limit),
          'X-RateLimit-Remaining': '0',
          'X-RateLimit-Reset': String(res.resetSeconds),
        },
      }
    );
  }

  return null;
}
