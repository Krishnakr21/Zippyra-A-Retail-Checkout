import { test, expect } from '@playwright/test';

test.describe('Admin Platform Store KYC Compliance E2E Flow', () => {
  test('Store compliance page displays warning banner when KYC unverified -> updating status to VERIFIED removes banner', async ({ page }) => {
    // Intercept compliance API endpoints with initial UNVERIFIED state
    let kycStatus = 'PENDING';

    await page.route('**/v1/compliance/irn*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ records: [] }),
      });
    });

    await page.route('**/v1/compliance/velocity-alerts*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ alerts: [] }),
      });
    });

    await page.route('**/v1/compliance/kyc*', (route) => {
      if (route.request().method() === 'GET') {
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            store_id: 'store-e2e-01',
            gstin_verified: false,
            pan_verified: false,
            kyc_status: kycStatus,
          }),
        });
      } else if (route.request().method() === 'PUT' || route.request().method() === 'POST') {
        kycStatus = 'VERIFIED';
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            store_id: 'store-e2e-01',
            gstin_verified: true,
            pan_verified: true,
            kyc_status: 'VERIFIED',
          }),
        });
      }
    });

    // 1. Navigate to store compliance page
    await page.goto('http://localhost:3011/dashboard/stores/store-e2e-01/compliance');

    // 2. Verify mandatory unverified payment warning banner is visible
    const banner = page.locator('#kyc-unverified-banner');
    await expect(banner).toBeVisible();
    await expect(banner).toContainText('This store cannot legally process real payments until KYC is verified');

    // 3. Change KYC status dropdown to VERIFIED and click save
    await page.selectOption('#kyc-status-select', 'VERIFIED');
    await page.click('#save-kyc-btn');

    // 4. Verify banner is removed from DOM once KYC status is VERIFIED
    await expect(banner).not.toBeVisible();
  });
});
