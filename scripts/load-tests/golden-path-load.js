import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 5 },
    { duration: '2m', target: 20 }, // 20 simultaneous full customer journeys
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

export default function () {
  const storeId = 'store-001';
  const phone = `+919${Math.floor(100000000 + Math.random() * 900000000)}`;

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  // Step 1: Auth (OTP)
  const authRes = http.post(`${BASE_URL}/v1/auth/otp/send`, JSON.stringify({ channel: 'phone', identifier: phone }), params);
  check(authRes, { 'step 1 auth send status 200/429': (r) => r.status === 200 || r.status === 429 });

  const token = 'Bearer k6-golden-path-jwt-token';
  const authParams = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': token,
    },
  };

  // Step 2: Cart Scan
  const scanRes = http.post(`${BASE_URL}/v1/cart/scan`, JSON.stringify({ store_id: storeId, barcode: '8901030300011', quantity: 1 }), authParams);
  check(scanRes, { 'step 2 scan status 200': (r) => r.status === 200 });

  // Step 3: Payment Initiate
  const payRes = http.post(`${BASE_URL}/v1/payment/initiate`, JSON.stringify({ cart_id: `cart-gp-${__VU}`, amount_paise: 6500, integrity_token: 'valid-k6-token' }), authParams);
  check(payRes, { 'step 3 pay status 200': (r) => r.status === 200 });

  // Step 4: Exit Validation
  const exitRes = http.post(`${BASE_URL}/v1/exit/validate`, JSON.stringify({ token: `exit-gp-${__VU}`, gate_id: 'GATE_E2E_01' }), {
    headers: { 'Content-Type': 'application/json', 'X-Gate-Secret': 'zippyra-gate-hmac-secret-32bytes' },
  });
  check(exitRes, { 'step 4 exit status 200/400/409': (r) => r.status === 200 || r.status === 400 || r.status === 409 });

  sleep(1);
}
