import assert from 'node:assert';
import { test, beforeEach } from 'node:test';
import { NextRequest } from 'next/server';
import { checkBFFRateLimit, memoryRateLimiter } from './rate-limiter';

beforeEach(() => {
  memoryRateLimiter.clear();
});

test('Auth-adjacent BFF route allows up to 10 requests and rejects 11th with 429', async () => {
  const url = 'http://localhost:3000/api/auth/send-otp';

  // First 10 requests should pass
  for (let i = 1; i <= 10; i++) {
    const req = new NextRequest(url, {
      headers: { 'x-forwarded-for': '192.168.1.100' },
    });
    const res = await checkBFFRateLimit(req);
    assert.strictEqual(res, null, `Request ${i} should be allowed`);
  }

  // 11th request should be rejected with 429 Too Many Requests
  const req11 = new NextRequest(url, {
    headers: { 'x-forwarded-for': '192.168.1.100' },
  });
  const res11 = await checkBFFRateLimit(req11);
  assert.notStrictEqual(res11, null, '11th request should be blocked');
  assert.strictEqual(res11?.status, 429);

  const retryAfter = res11?.headers.get('Retry-After');
  assert.ok(retryAfter, 'Retry-After header should be present');

  const body = await res11?.json();
  assert.strictEqual(body.error.code, 'TOO_MANY_REQUESTS');
  assert.strictEqual(body.error.message, 'Too many requests. Please try again shortly.');
});

test('Legitimate traffic from different IPs are tracked independently', async () => {
  const url = 'http://localhost:3000/api/auth/send-otp';

  // Fill up IP A
  for (let i = 1; i <= 10; i++) {
    const reqA = new NextRequest(url, {
      headers: { 'x-forwarded-for': '10.0.0.1' },
    });
    await checkBFFRateLimit(reqA);
  }

  // IP A 11th is blocked
  const reqA11 = new NextRequest(url, {
    headers: { 'x-forwarded-for': '10.0.0.1' },
  });
  const resA11 = await checkBFFRateLimit(reqA11);
  assert.strictEqual(resA11?.status, 429);

  // IP B 1st is allowed
  const reqB1 = new NextRequest(url, {
    headers: { 'x-forwarded-for': '10.0.0.2' },
  });
  const resB1 = await checkBFFRateLimit(reqB1);
  assert.strictEqual(resB1, null, 'IP B should be allowed');
});

test('General API routes have a 120 req/min limit', async () => {
  const url = 'http://localhost:3000/api/catalog/products';

  for (let i = 1; i <= 120; i++) {
    const req = new NextRequest(url, {
      headers: { 'x-forwarded-for': '192.168.1.50' },
    });
    const res = await checkBFFRateLimit(req);
    assert.strictEqual(res, null, `General request ${i} should be allowed`);
  }

  // 121st request blocked
  const req121 = new NextRequest(url, {
    headers: { 'x-forwarded-for': '192.168.1.50' },
  });
  const res121 = await checkBFFRateLimit(req121);
  assert.strictEqual(res121?.status, 429);
});
