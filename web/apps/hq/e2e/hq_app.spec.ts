import { test, expect } from '@playwright/test';
import { checkA11y } from '../../../e2e-a11y-helper';

test.describe('Chain HQ Platform E2E Flow', () => {
  test('Phone + OTP login flow to Dashboard with Degraded Data Banner', async ({ page }) => {
    await page.goto('http://localhost:3012/login');
    await checkA11y(page);

    // Phone Step
    await page.fill('input[type="text"]', '+919876543210');
    await page.click('button[type="submit"]');

    // OTP Step
    await expect(page.locator('text=Enter 6-Digit OTP')).toBeVisible();
    await page.fill('input[type="text"]', '123456');
    await page.click('button[type="submit"]');

    // Dashboard Overview
    await expect(page).toHaveURL(/.*\/dashboard/);
    await expect(page.locator('text=Chain Executive Overview')).toBeVisible();
    await expect(page.locator('text=Total Stores')).toBeVisible();

    // Partial failure banner check
    const banner = page.locator('[data-testid="degraded-stores-banner"]');
    if (await banner.isVisible()) {
      await expect(banner).toContainText('Incomplete Metrics Notice');
    }
  });

  test('Bulk Catalog Import Flow with Per-Store Progress Rows', async ({ page }) => {
    await page.goto('http://localhost:3012/dashboard/catalog/bulk-import');

    await expect(page.locator('text=Bulk Catalog CSV Import')).toBeVisible();
    await page.click('text=Select Specific Stores');

    // Initiate import simulation
    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles({
      name: 'products.csv',
      mimeType: 'text/csv',
      buffer: Buffer.from('barcode,name,price_paise\n890123,Test Item,1000'),
    });

    await page.click('button[type="submit"]');

    // Per-store progress rows check
    await expect(page.locator('[data-testid="bulk-import-progress"]')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('text=Per-Store Import Progress')).toBeVisible();
  });
});
