import { test, expect } from '@playwright/test';

test.describe('Chain HQ Dashboard Real Revenue & Analytics E2E', () => {
  test('Dashboard displays real revenue KPI figures when analytics is available', async ({ page }) => {
    // Mock successful dashboard response with real revenue metrics
    await page.route('**/v1/chain-hq/dashboard', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          total_stores: 12,
          active_stores: 11,
          stores_with_low_stock_count: 3,
          total_low_stock_items: 24,
          degraded_stores: [],
          total_revenue_paise: 125000000, // ₹12,50,000
          total_orders: 4500,
          as_of: new Date().toISOString(),
        }),
      });
    });

    await page.goto('http://localhost:3012/dashboard');

    await expect(page.locator('[data-testid="hq-dashboard-overview"]')).toBeVisible();
    await expect(page.locator('[data-testid="kpi-total-revenue"]')).toContainText('₹1,250,000');
    await expect(page.locator('[data-testid="kpi-total-orders"]')).toContainText('4,500');
  });

  test('Simulating analytics_unavailable=true renders non-alarming notice without breaking rest of dashboard', async ({ page }) => {
    // Mock response with analytics_unavailable: true
    await page.route('**/v1/chain-hq/dashboard', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          total_stores: 12,
          active_stores: 11,
          stores_with_low_stock_count: 3,
          total_low_stock_items: 24,
          degraded_stores: [],
          analytics_unavailable: true,
          as_of: new Date().toISOString(),
        }),
      });
    });

    await page.goto('http://localhost:3012/dashboard');

    await expect(page.locator('[data-testid="hq-dashboard-overview"]')).toBeVisible();
    // Notice banner must be visible
    await expect(page.locator('[data-testid="analytics-unavailable-banner"]')).toBeVisible();
    await expect(page.locator('[data-testid="analytics-unavailable-banner"]')).toContainText(
      'Revenue Analytics Temporarily Unavailable'
    );

    // KPI total revenue must indicate Unavailable instead of zero
    await expect(page.locator('[data-testid="kpi-total-revenue"]')).toContainText('Unavailable');

    // Rest of dashboard KPI cards must remain intact
    await expect(page.locator('text=Active Stores')).toBeVisible();
    await expect(page.locator('text=Low Stock Stores')).toBeVisible();
  });
});
