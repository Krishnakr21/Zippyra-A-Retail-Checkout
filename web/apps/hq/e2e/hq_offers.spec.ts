import { test, expect } from '@playwright/test';

test.describe('Chain HQ Offer Authoring & Retailer Cross-App Visibility E2E Flow', () => {
  test('OWNER creates chain-wide offer -> Retailer offers page shows it as CHAIN_WIDE non-editable row', async ({ page }) => {
    let mockChainOffers: any[] = [];

    // Intercept HQ offer API endpoints
    await page.route('**/v1/cart/admin/offers?chain_id=*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ offers: mockChainOffers }),
      });
    });

    await page.route('**/v1/cart/admin/offers', (route) => {
      if (route.request().method() === 'POST') {
        const body = JSON.parse(route.request().postData() || '{}');
        const created = {
          id: 'off-cw-e2e-200',
          chain_id: 'chain-001',
          store_id: null,
          type: body.type,
          applies_to: body.applies_to,
          rule_config: body.rule_config,
          min_cart_value_paise: body.min_cart_value_paise,
          priority: body.priority,
          active_from: body.active_from,
          is_active: true,
          scope: 'CHAIN_WIDE',
        };
        mockChainOffers.push(created);

        route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({ offer: created, warnings: [] }),
        });
      }
    });

    // 1. Navigate to Chain HQ Offers page as OWNER
    await page.goto('http://localhost:3002/dashboard/offers');
    await expect(page.locator('text=Chain HQ Offer Campaign Authoring')).toBeVisible();

    // 2. OWNER clicks "+ New Chain-Wide Offer" button
    const newChainBtn = page.locator('[data-testid="new-chain-offer-btn"]');
    await expect(newChainBtn).toBeVisible();
    await newChainBtn.click();

    // 3. Fill form for Flat ₹50 Off
    await page.selectOption('#offer-type-select', 'FLAT_OFF');
    await page.fill('#flat-amount-input', '50');
    await page.click('#submit-offer-btn');

    // 4. Verify chain-wide offer appears in HQ table
    await expect(page.locator('text=₹50.00 Flat Off')).toBeVisible();

    // 5. Navigate to Retailer Dashboard Offers page to verify cross-app visibility
    await page.route('**/v1/cart/admin/offers?store_id=*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ offers: mockChainOffers }),
      });
    });

    await page.goto('http://localhost:3000/dashboard/offers');

    // 6. Verify row displays CHAIN_WIDE badge and disabled edit button
    await expect(page.locator('text=₹50.00 Flat Off')).toBeVisible();
    await expect(page.locator('text=CHAIN_WIDE')).toBeVisible();

    const editBtn = page.locator('[data-testid="edit-offer-btn-off-cw-e2e-200"]');
    await expect(editBtn).toBeDisabled();
  });
});
