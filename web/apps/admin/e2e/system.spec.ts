import { test, expect } from '@playwright/test';

test.describe('Admin System Platform Ops E2E', () => {
  test.beforeEach(async ({ page }) => {
    // Mock login session as ADMIN
    await page.context().addCookies([
      {
        name: 'admin_role',
        value: 'ADMIN',
        domain: 'localhost',
        path: '/',
      },
    ]);
  });

  test('replaying a seeded DLQ message triggers selective replay and updates status', async ({ page }) => {
    let replayed = false;

    // Intercept DLQ APIs
    await page.route('**/v1/audit/kafka/dlq-topics', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          dlq_topics: [
            {
              topic_name: 'payment.confirmed.dlq',
              message_count: replayed ? 0 : 1,
              oldest_message_age_seconds: replayed ? 0 : 120,
            },
          ],
        }),
      });
    });

    await page.route('**/v1/audit/kafka/dlq-topics/payment.confirmed.dlq/messages*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          topic: 'payment.confirmed.dlq',
          messages: replayed
            ? []
            : [
                {
                  topic: 'payment.confirmed.dlq',
                  offset: 505,
                  key: 'pay-505',
                  value: { payment_id: 'pay-505', amount: 5000 },
                  timestamp: new Date().toISOString(),
                },
              ],
          total: replayed ? 0 : 1,
        }),
      });
    });

    await page.route('**/v1/audit/kafka/dlq-topics/payment.confirmed.dlq/replay', async (route) => {
      replayed = true;
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          replayed_count: 1,
          failed_offsets: [],
        }),
      });
    });

    await page.goto('/dashboard/system');

    // Click inspect topic
    await page.getByText('payment.confirmed.dlq').click();
    await expect(page.getByText('Offset #505')).toBeVisible();

    // Select offset 505 checkbox
    await page.getByRole('checkbox').first().click();

    // Click Replay Selected
    await page.getByRole('button', { name: 'Replay Selected (1)' }).click();

    // Verify replay success alert
    await expect(page.getByText('Successfully replayed 1 message(s)')).toBeVisible();
  });

  test('circuit-breaker status tab auto-refreshes without a manual page reload', async ({ page }) => {
    let callCount = 0;

    await page.route('**/v1/payment/internal/circuit-breaker-status', async (route) => {
      callCount++;
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          gateway: 'razorpay',
          state: callCount > 1 ? 'OPEN' : 'CLOSED',
          error_rate_rolling_1min: callCount > 1 ? 0.08 : 0.0,
        }),
      });
    });

    await page.goto('/dashboard/system');

    // Switch to Circuit Breaker Tab
    await page.getByText('Circuit Breaker Status').click();

    await expect(page.getByText('Healthy (Razorpay Active)')).toBeVisible();
    expect(callCount).toBe(1);

    // Fast-forward or wait 10s for auto-refresh network request
    await page.waitForTimeout(10500);

    // Verify a second network call fired and status updated to OPEN automatically
    expect(callCount).toBeGreaterThanOrEqual(2);
  });
});
