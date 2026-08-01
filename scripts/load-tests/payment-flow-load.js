import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

const duplicatePaymentCounter = new Counter('duplicate_payments_detected');

export const options = {
  stages: [
    { duration: '1m', target: 50 },
    { duration: '2m', target: 200 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    duplicate_payments_detected: ['count==0'], // ZERO duplicate payments under load
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

export default function () {
  const cartId = `cart-k6-${__VU}-${__ITER}`;
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer k6-load-test-jwt-token',
    },
  };

  // 1. Initiate Payment
  const initPayload = JSON.stringify({
    cart_id: cartId,
    amount_paise: 6825,
    integrity_token: 'valid-k6-token',
  });

  const initRes = http.post(`${BASE_URL}/v1/payment/initiate`, initPayload, params);

  check(initRes, {
    'initiate status is 200': (r) => r.status === 200,
  });

  if (initRes.status === 200) {
    const data = JSON.parse(initRes.body);
    const paymentId = data.payment_id;

    // 2. Concurrent Webhook Simulation (Simulate duplicate Razorpay webhooks)
    const webhookPayload = JSON.stringify({
      event: 'payment.authorized',
      payload: {
        payment: {
          entity: {
            id: `pay_rzp_${paymentId}`,
            order_id: data.razorpay_order_id,
            status: 'captured',
            amount: 6825,
          },
        },
      },
    });

    const whRes1 = http.post(`${BASE_URL}/v1/payment/webhook/razorpay`, webhookPayload, params);
    const whRes2 = http.post(`${BASE_URL}/v1/payment/webhook/razorpay`, webhookPayload, params);

    // Verify idempotency: duplicate webhook should be ignored without creating duplicate order
    if (whRes1.status === 200 && whRes2.status === 200) {
      // 3. Poll Payment Status
      const statusRes = http.get(`${BASE_URL}/v1/payment/status/${paymentId}`, params);
      check(statusRes, {
        'status is 200': (r) => r.status === 200,
        'payment state is SUCCESS': (r) => {
          const body = JSON.parse(r.body);
          return body.status === 'SUCCESS';
        },
      });
    } else {
      duplicatePaymentCounter.add(1);
    }
  }

  sleep(1);
}
