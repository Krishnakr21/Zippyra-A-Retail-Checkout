import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 50 },   // 1 Store (50 VUs)
    { duration: '2m', target: 500 },  // 10 Stores (500 VUs)
    { duration: '2m', target: 2500 }, // 50 Stores (2500 VUs - Breaking Point Stress)
  ],
  thresholds: {
    http_req_duration: ['p(99)<100'], // P99 response time must stay under 100ms
    http_req_failed: ['rate<0.01'],    // Error rate must be under 1%
  },
};

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

export default function () {
  const storeId = `store-${Math.floor(Math.random() * 50) + 1}`;
  const barcodes = ['8901030300011', '4006381333931', '8901234567890'];
  const barcode = barcodes[Math.floor(Math.random() * barcodes.length)];

  const payload = JSON.stringify({
    store_id: storeId,
    barcode: barcode,
    quantity: 1,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer k6-load-test-jwt-token',
    },
  };

  const res = http.post(`${BASE_URL}/v1/cart/scan`, payload, params);

  check(res, {
    'status is 200': (r) => r.status === 200,
    'latency under 100ms': (r) => r.timings.duration < 100,
  });

  sleep(0.5);
}
