import { test, expect } from '@playwright/test';
import { checkA11y } from '../../../e2e-a11y-helper';

test.describe('Retailer Dashboard Real Authentication & PIN E2E Flow', () => {
  test('OTP Auth Flow: Send OTP -> Verify -> Dashboard redirect', async ({ page }) => {
    await page.goto('http://localhost:3010/login');
    await checkA11y(page);

    await expect(page.locator('text=Zippyra Retailer Dashboard')).toBeVisible();

    // Fill phone number
    await page.fill('[data-testid="identifier-input"]', '+919876543210');
    await page.click('[data-testid="send-otp-btn"]');

    // OTP Verify step
    await expect(page.locator('[data-testid="otp-input"]')).toBeVisible();
    await page.fill('[data-testid="otp-input"]', '123456');
    await page.click('[data-testid="verify-otp-btn"]');
  });

  test('Quick Staff PIN Auth Flow: Switch mode -> Store ID + PIN -> Login', async ({ page }) => {
    await page.goto('http://localhost:3010/login');

    // Switch to Quick Staff PIN tab
    await page.click('[data-testid="mode-pin-btn"]');
    await expect(page.locator('[data-testid="pin-input"]')).toBeVisible();

    await page.fill('[data-testid="store-id-input"]', 'store-001');
    await page.fill('[data-testid="pin-input"]', '123456');
    await page.click('[data-testid="pin-login-btn"]');
  });
});
