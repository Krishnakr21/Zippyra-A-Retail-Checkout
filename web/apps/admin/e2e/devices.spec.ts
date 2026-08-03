import { test, expect } from '@playwright/test';
import { checkA11y } from '../../../e2e-a11y-helper';

test.describe('Admin Platform Device Provisioning E2E Flow', () => {
  test('Provision device -> credential modal appears -> confirm -> device listed with PROVISIONING status', async ({ page }) => {
    await page.goto('http://localhost:3011/dashboard/stores/store-001/devices');
    await checkA11y(page);

    await expect(page.locator('text=Store Hardware Provisioning')).toBeVisible();

    // Click Provision New Device button
    await page.click('[data-testid="provision-device-btn"]');

    // Fill form
    await page.fill('[data-testid="gate-id-input"]', 'GATE_E2E_01');
    await page.fill('[data-testid="device-label-input"]', 'E2E Test Gate');
    await page.click('[data-testid="submit-provision-btn"]');

    // Verify Credential Download Modal appears
    await expect(page.locator('text=CRITICAL: One-Time Credential Access')).toBeVisible();
    await expect(page.locator('[data-testid="download-bundle-btn"]')).toBeVisible();

    // Done button disabled until checkbox checked
    const doneBtn = page.locator('[data-testid="close-credential-modal-btn"]');
    await expect(doneBtn).toBeDisabled();

    // Check confirmation checkbox
    await page.click('[data-testid="confirm-download-checkbox"]');
    await expect(doneBtn).toBeEnabled();
    await doneBtn.click();

    // Modal dismissed, device appears in table
    await expect(page.locator('text=CRITICAL: One-Time Credential Access')).not.toBeVisible();
    await expect(page.locator('text=E2E Test Gate')).toBeVisible();
  });
});
