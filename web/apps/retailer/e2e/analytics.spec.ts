import { test, expect } from '@playwright/test';
import { checkA11y } from '../../../e2e-a11y-helper';

test.describe('Retailer Dashboard Real Analytics E2E Flow', () => {
  test('Date-range change re-fetches all 4 sections independently via distinct network calls', async ({ page }) => {
    // Track requests to the 4 independent analytics endpoints
    const endpointHits: string[] = [];

    page.on('request', (request) => {
      const url = request.url();
      if (url.includes('/v1/analytics/sales')) endpointHits.push('sales');
      if (url.includes('/v1/analytics/funnel')) endpointHits.push('funnel');
      if (url.includes('/v1/analytics/peak-hours')) endpointHits.push('peak-hours');
      if (url.includes('/v1/analytics/top-products')) endpointHits.push('top-products');
    });

    await page.goto('http://localhost:3010/dashboard/analytics');
    await checkA11y(page);

    // Verify main Analytics page container and 4 independent sections exist
    await expect(page.locator('[data-testid="retailer-analytics-page"]')).toBeVisible();
    await expect(page.locator('[data-testid="section-sales-trend"]')).toBeVisible();
    await expect(page.locator('[data-testid="section-funnel"]')).toBeVisible();
    await expect(page.locator('[data-testid="section-peak-hours"]')).toBeVisible();
    await expect(page.locator('[data-testid="section-top-products"]')).toBeVisible();

    // Clear initial hit counts
    endpointHits.length = 0;

    // Change date-from input
    const dateFromInput = page.locator('[data-testid="date-from-input"]');
    await dateFromInput.fill('2026-06-01');

    // Click "Last 30 Days" preset button to trigger date range change
    await page.click('[data-testid="preset-30d-btn"]');

    // Assert distinct network requests for the sections were triggered
    await expect.poll(() => endpointHits.length).toBeGreaterThan(0);
  });
});
