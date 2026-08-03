import { test, expect } from '@playwright/test';

test.describe('Retailer Dashboard Offer Authoring E2E Flow', () => {
  test('Create store-specific offer -> compiled rules in OfferPreviewPanel reflect it', async ({ page }) => {
    let mockOffers: any[] = [];
    let mockPreviewRules: any[] = [];

    // Intercept backend offer API endpoints
    await page.route('**/v1/cart/admin/offers?store_id=*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ offers: mockOffers }),
      });
    });

    await page.route('**/v1/cart/admin/offers/store-001/preview', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockPreviewRules),
      });
    });

    await page.route('**/v1/cart/admin/offers', (route) => {
      if (route.request().method() === 'POST') {
        const body = JSON.parse(route.request().postData() || '{}');
        const created = {
          id: 'off-e2e-100',
          chain_id: 'chain-001',
          store_id: 'store-001',
          type: body.type,
          applies_to: body.applies_to,
          target_ids: body.target_ids,
          rule_config: body.rule_config,
          min_cart_value_paise: body.min_cart_value_paise,
          priority: body.priority,
          active_from: body.active_from,
          is_active: true,
          scope: 'STORE_SPECIFIC',
        };
        mockOffers.push(created);
        mockPreviewRules.push({
          id: 'off-e2e-100',
          type: body.type,
          value: body.rule_config?.percent || 10,
          applies_to: body.applies_to,
          target_ids: body.target_ids,
          min_cart_value_paise: body.min_cart_value_paise,
        });

        route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({ offer: created, warnings: [] }),
        });
      }
    });

    // 1. Navigate to Retailer Offers page
    await page.goto('http://localhost:3000/dashboard/offers');

    await expect(page.locator('text=Store Promotional Offers & Campaigns')).toBeVisible();

    // 2. Click "+ New Store Offer" button
    await page.click('[data-testid="new-store-offer-btn"]');

    // 3. Select CATEGORY_PERCENT_OFF & CATEGORY applies_to
    await page.selectOption('#offer-type-select', 'CATEGORY_PERCENT_OFF');
    await page.selectOption('#applies-to-select', 'CATEGORY');
    await page.fill('#percent-input', '10');
    await page.fill('#target-ids-input', 'cat-beverages');

    // 4. Submit form
    await page.click('#submit-offer-btn');

    // 5. Verify offer appears in table
    await expect(page.locator('text=10% Off')).toBeVisible();

    // 6. Verify compiled rule appears in OfferPreviewPanel
    await expect(page.locator('[data-testid="compiled-rule-row-off-e2e-100"]')).toBeVisible();
    await expect(page.locator('text=CATEGORY_PERCENT_OFF')).toBeVisible();
  });
});
