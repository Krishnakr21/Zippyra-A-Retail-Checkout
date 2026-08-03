import { test, expect } from '@playwright/test';
import { checkA11y } from '../../../e2e-a11y-helper';

test.describe('Retailer Dashboard Devices & Alerts E2E Flow', () => {
  test('Resolve hardware alert -> alert row disappears from unresolved alerts view', async ({ page }) => {
    await page.goto('http://localhost:3010/dashboard/devices');
    await checkA11y(page);

    await expect(page.locator('text=Store Hardware Devices')).toBeVisible();

    // Verify unresolved alert section is visible
    await expect(page.locator('[data-testid="unresolved-alerts-section"]')).toBeVisible();
    await expect(page.locator('[data-testid="alert-row-alt-001"]')).toBeVisible();

    // Click Resolve Alert button
    await page.click('[data-testid="resolve-alert-btn-alt-001"]');

    // Verify alert row disappears without reload
    await expect(page.locator('[data-testid="alert-row-alt-001"]')).not.toBeVisible();
  });
});
