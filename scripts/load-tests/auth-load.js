import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 200 },
    { duration: '1m', target: 1000 }, // 1000 concurrent OTP cycles
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    'http_req_duration{name:VerifyOTP}': ['p(99)<300'], // P99 < 300ms for OTP verify
  },
};

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

export default function () {
  const phone = `+919${Math.floor(100000000 + Math.random() * 900000000)}`;

  // 1. Send OTP
  const sendPayload = JSON.stringify({
    channel: 'phone',
    identifier: phone,
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const sendRes = http.post(`${BASE_URL}/v1/auth/otp/send`, sendPayload, Object.assign({}, params, { tags: { name: 'SendOTP' } }));

  check(sendRes, {
    'send status 200 or 429': (r) => r.status === 200 || r.status === 429,
  });

  if (sendRes.status === 200) {
    // 2. Verify OTP
    const verifyPayload = JSON.stringify({
      channel: 'phone',
      identifier: phone,
      otp: '123456',
      device_id: 'k6-device-id',
    });

    const verifyRes = http.post(`${BASE_URL}/v1/auth/otp/verify`, verifyPayload, Object.assign({}, params, { tags: { name: 'VerifyOTP' } }));

    check(verifyRes, {
      'verify latency under 300ms': (r) => r.timings.duration < 300,
    });
  }

  sleep(1);
}
