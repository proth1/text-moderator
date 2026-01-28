import { test, expect } from '../fixtures/auth';

test.describe('Policy Management', () => {
  test('displays policies page with header', async ({ adminPage }) => {
    await adminPage.goto('/policies');
    await expect(adminPage.locator('main h1')).toContainText('Policies');
    await expect(adminPage.locator('p:has-text("Manage moderation policies")')).toBeVisible();
  });

  test('shows New Policy button', async ({ adminPage }) => {
    await adminPage.goto('/policies');
    await expect(adminPage.locator('text=New Policy')).toBeVisible();
  });

  test('displays seeded policies from database', async ({ adminPage }) => {
    await adminPage.goto('/policies');

    const policyItems = adminPage.locator('[data-testid="policy-item"]');
    await expect(policyItems.first()).toBeVisible({ timeout: 10000 });

    const count = await policyItems.count();
    expect(count).toBeGreaterThanOrEqual(2);
  });

  test('policy items display name, status, and version', async ({ adminPage }) => {
    await adminPage.goto('/policies');

    const firstPolicy = adminPage.locator('[data-testid="policy-item"]').first();
    await expect(firstPolicy).toBeVisible({ timeout: 10000 });

    await expect(firstPolicy.locator('[data-testid="policy-name"]')).toBeVisible();
    await expect(firstPolicy.locator('[data-testid="policy-status"]')).toBeVisible();
    await expect(firstPolicy.locator('[data-testid="policy-version"]')).toBeVisible();
  });

  test('shows Standard Community Guidelines from seed data', async ({ adminPage }) => {
    await adminPage.goto('/policies');
    await expect(adminPage.locator('text=Standard Community Guidelines')).toBeVisible({ timeout: 10000 });
  });

  test('shows Youth Safe Mode from seed data', async ({ adminPage }) => {
    await adminPage.goto('/policies');
    await expect(adminPage.locator('text=Youth Safe Mode')).toBeVisible({ timeout: 10000 });
  });

  test('policy status badges are displayed', async ({ adminPage }) => {
    await adminPage.goto('/policies');
    await expect(adminPage.locator('[data-testid="policy-item"]').first()).toBeVisible({ timeout: 10000 });
    await expect(adminPage.locator('text=PUBLISHED').first()).toBeVisible();
  });

  test('clicking New Policy navigates to policy editor', async ({ adminPage }) => {
    await adminPage.goto('/policies');
    await adminPage.locator('text=New Policy').click();
    await expect(adminPage.locator('main h1')).toContainText('Create New Policy');
  });

  test('policy editor shows category configuration', async ({ adminPage }) => {
    await adminPage.goto('/policies/new');

    await expect(adminPage.locator('text=Policy Name')).toBeVisible();
    await expect(adminPage.locator('text=Description')).toBeVisible();
    await expect(adminPage.locator('text=Category Configuration')).toBeVisible();
    await expect(adminPage.locator('button:has-text("Save as Draft")')).toBeVisible();
    await expect(adminPage.locator('button:has-text("Save & Publish")')).toBeVisible();
  });

  test('clicking Edit on a policy navigates to editor', async ({ adminPage }) => {
    await adminPage.goto('/policies');
    await expect(adminPage.locator('[data-testid="policy-item"]').first()).toBeVisible({ timeout: 10000 });

    await adminPage.locator('[data-testid="policy-item"]').first().locator('text=Edit').click();
    await expect(adminPage.locator('main h1')).toContainText('Edit Policy');
  });
});
