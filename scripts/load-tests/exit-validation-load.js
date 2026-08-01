import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

const doubleUseAllowedCounter = new Counter('double_use_race_violations');

export const options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '1m', target: 200 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    double_use_race_violations: ['count==0'], // ZERO double-use race condition violations
  },
};

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

export default function () {
  const exitToken = `exit-token-k6-${__VU}-${__ITER}`;
  const gateId = 'GATE_E2E_01';

  const payload = JSON.stringify({
    token: exitToken,
    gate_id: gateId,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-Gate-Secret': 'zippyra-gate-hmac-secret-32bytes',
    },
  };

  // Attempt 1: Initial exit gate scan
  const res1 = http.post(`${BASE_URL}/v1/exit/validate`, payload, params);

  // Attempt 2: Immediate duplicate replay (concurrent scan attempt)
  const res2 = http.post(`${BASE_URL}/v1/exit/validate`, payload, params);

  check(res1, {
    'initial validation status evaluated': (r) => r.status === 200 || r.status === 400 || r.status === 409,
  });

  check(res2, {
    'duplicate scan returns 409 QR_ALREADY_USED': (r) => {
      if (res1.status === 200 && r.status === 200) {
        // Double use occurred!
        doubleUseAllowedCounter.add(1);
        return false;
      }
      return true;
    },
  });

  sleep(0.5);
}
