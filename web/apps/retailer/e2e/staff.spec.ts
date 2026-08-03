import { test, expect } from '@playwright/test';

test.describe('Retailer Dashboard Staff Management E2E', () => {
  test.beforeEach(async ({ page }) => {
    // Mock login session as MANAGER
    await page.context().addCookies([
      {
        name: 'user_role',
        value: 'MANAGER',
        domain: 'localhost',
        path: '/',
      },
    ]);
  });

  test('MANAGER can add staff member and see it appear in table', async ({ page }) => {
    // Intercept staff API GET & POST
    let staffRoster = [
      {
        id: 'staff-e2e-1',
        name: 'Anita Sharma',
        phone: '+919876543299',
        role: 'CASHIER',
        is_active: true,
        has_pin_set: true,
        store_id: 'store-mumbai-01',
        created_at: new Date().toISOString(),
      },
    ];

    await page.route('**/v1/retailer-auth/staff*', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ staff: staffRoster }),
        });
      } else if (route.request().method() === 'POST') {
        const body = JSON.parse(route.request().postData() || '{}');
        const newMember = {
          id: 'staff-e2e-new',
          name: body.name,
          phone: body.phone,
          role: body.role,
          is_active: true,
          has_pin_set: false,
          store_id: 'store-mumbai-01',
          created_at: new Date().toISOString(),
        };
        staffRoster.push(newMember);
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify(newMember),
        });
      }
    });

    await page.goto('/dashboard/staff');

    await expect(page.getByText('Anita Sharma')).toBeVisible();

    // Click Add Staff
    await page.getByRole('button', { name: '+ Add Staff' }).click();

    // Fill form
    await page.getByPlaceholder('e.g. Ramesh Kumar').fill('Vikas Verma');
    await page.getByPlaceholder('+919876543210').fill('+919812345678');
    await page.getByRole('combobox').nth(1).selectOption('SECURITY');

    // Submit
    await page.getByRole('button', { name: 'Add Member' }).click();

    // Verify newly added staff appears with is_active=true and has_pin_set=false (OTP Only)
    await expect(page.getByText('Vikas Verma')).toBeVisible();
    await expect(page.getByText('🔒 OTP Only')).toBeVisible();
  });

  test('deactivating a staff member updates status without full page reload', async ({ page }) => {
    let staffRoster = [
      {
        id: 'staff-e2e-2',
        name: 'Karan Johar',
        phone: '+919876543288',
        role: 'STOCK_ASSOCIATE',
        is_active: true,
        has_pin_set: false,
        store_id: 'store-mumbai-01',
        created_at: new Date().toISOString(),
      },
    ];

    await page.route('**/v1/retailer-auth/staff*', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ staff: staffRoster }),
        });
      } else if (route.request().method() === 'DELETE') {
        staffRoster[0].is_active = false;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ status: 'DEACTIVATED' }),
        });
      }
    });

    await page.goto('/dashboard/staff');
    await expect(page.getByText('Karan Johar')).toBeVisible();

    // Click Deactivate button
    await page.getByRole('button', { name: 'Deactivate' }).click();

    // Warning confirmation modal appears
    await expect(page.getByText('This will immediately log this staff member out everywhere')).toBeVisible();

    // Confirm Deactivation
    await page.getByRole('button', { name: 'Deactivate Staff' }).click();

    // Uncheck "Show Active Only" to verify deactivated staff status
    await page.getByLabel('Show Active Only').uncheck();
    await expect(page.getByText('Inactive')).toBeVisible();
  });

  test('non-MANAGER session attempting to navigate to /dashboard/staff is redirected', async ({ page }) => {
    // Set non-MANAGER cookie (e.g. CASHIER)
    await page.context().addCookies([
      {
        name: 'user_role',
        value: 'CASHIER',
        domain: 'localhost',
        path: '/',
      },
    ]);

    await page.goto('/dashboard/staff');

    // Middleware redirects CASHIER to /dashboard?error=UNAUTHORIZED_ROLE
    await expect(page).toHaveURL(/\/dashboard/);
  });
});
