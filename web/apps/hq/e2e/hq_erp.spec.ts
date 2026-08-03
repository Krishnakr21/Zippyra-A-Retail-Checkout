import { test, expect } from '@playwright/test';

test.describe('Chain HQ ERP Integration E2E Flow', () => {
  test('OWNER creates DIRECT-mode SAP connection -> PENDING_SETUP -> flips ACTIVE on activity', async ({ page }) => {
    let mockConnections: any[] = [];

    await page.route('**/v1/integration/connections?chain_id=*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ connections: mockConnections }),
      });
    });

    await page.route('**/v1/integration/connections', (route) => {
      if (route.request().method() === 'POST') {
        const body = JSON.parse(route.request().postData() || '{}');
        const created = {
          id: 'conn-sap-e2e-100',
          chain_id: 'chain-001',
          erp_type: body.erp_type,
          integration_mode: body.integration_mode,
          display_name: body.display_name,
          enabled_outbound_events: body.enabled_outbound_events,
          status: 'PENDING_SETUP',
          created_at: new Date().toISOString(),
        };
        mockConnections.push(created);

        route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            connection: created,
            webhook_secret: 'whsec_e2e_sap_secret_123',
            connector_setup_note: 'Configure SAP Cloud Connector with webhook_secret',
          }),
        });
      }
    });

    // 1. Navigate to HQ ERP page
    await page.goto('http://localhost:3002/dashboard/erp');
    await expect(page.locator('text=Enterprise ERP Integrations')).toBeVisible();

    // 2. Click "+ New ERP Connection" button
    await page.click('[data-testid="new-connection-btn"]');

    // 3. Fill connection form
    await page.fill('input[placeholder="e.g. Primary Store Tally ERP"]', 'SAP Global OData Production');
    await page.click('[data-testid="submit-connection-btn"]');

    // 4. Verify Credential Modal pops up
    await expect(page.locator('text=CRITICAL: One-Time Secret Credentials')).toBeVisible();
    await expect(page.locator('text=whsec_e2e_sap_secret_123')).toBeVisible();

    // Check confirmation checkbox and dismiss modal
    await page.click('[data-testid="confirm-download-checkbox"]');
    await page.click('[data-testid="close-erp-credential-modal-btn"]');

    // 5. Verify connection appears in list with PENDING_SETUP badge
    await expect(page.locator('text=SAP Global OData Production')).toBeVisible();
    await expect(page.locator('text=PENDING_SETUP')).toBeVisible();

    // 6. Simulate successful activity flipping status to ACTIVE
    mockConnections[0].status = 'ACTIVE';

    await page.reload();
    await expect(page.locator('text=ACTIVE')).toBeVisible();
  });
});
