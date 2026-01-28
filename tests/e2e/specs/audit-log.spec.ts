import { test, expect } from '../fixtures/auth';

test.describe('Audit Log', () => {
  test('displays audit log page with header', async ({ adminPage }) => {
    await adminPage.goto('/audit');
    await expect(adminPage.locator('main h1')).toContainText('Audit Log');
    await expect(adminPage.locator('p:has-text("Review all moderation actions")')).toBeVisible();
  });

  test('shows evidence records table', async ({ adminPage }) => {
    await adminPage.goto('/audit');
    const records = adminPage.locator('[data-testid="evidence-record"]');
    await expect(records.first()).toBeVisible({ timeout: 10000 });
  });

  test('evidence records display control ID', async ({ adminPage }) => {
    await adminPage.goto('/audit');

    const firstRecord = adminPage.locator('[data-testid="evidence-record"]').first();
    await expect(firstRecord).toBeVisible({ timeout: 10000 });

    const controlId = firstRecord.locator('[data-testid="control-id"]');
    await expect(controlId).toBeVisible();
    await expect(controlId).toContainText(/MOD-001|GOV-002/);
  });

  test('evidence records show immutable badge', async ({ adminPage }) => {
    await adminPage.goto('/audit');

    const firstRecord = adminPage.locator('[data-testid="evidence-record"]').first();
    await expect(firstRecord).toBeVisible({ timeout: 10000 });

    await expect(firstRecord.locator('[data-testid="immutable-badge"]')).toContainText('Immutable');
  });

  test('evidence records display decision ID', async ({ adminPage }) => {
    await adminPage.goto('/audit');

    const firstRecord = adminPage.locator('[data-testid="evidence-record"]').first();
    await expect(firstRecord).toBeVisible({ timeout: 10000 });

    await expect(firstRecord.locator('[data-testid="decision-id"]')).toBeVisible();
  });

  test('evidence records display automated action', async ({ adminPage }) => {
    await adminPage.goto('/audit');

    const firstRecord = adminPage.locator('[data-testid="evidence-record"]').first();
    await expect(firstRecord).toBeVisible({ timeout: 10000 });

    await expect(firstRecord.locator('[data-testid="automated-action"]')).toBeVisible();
  });

  test('shows export button', async ({ adminPage }) => {
    await adminPage.goto('/audit');
    await expect(adminPage.locator('[data-testid="export-button"]')).toBeVisible();
  });

  test('shows search input', async ({ adminPage }) => {
    await adminPage.goto('/audit');
    await expect(adminPage.locator('[data-testid="search-input"]')).toBeVisible();
  });

  test('search filters evidence records', async ({ adminPage }) => {
    await adminPage.goto('/audit');
    await expect(adminPage.locator('[data-testid="evidence-record"]').first()).toBeVisible({ timeout: 10000 });

    const searchInput = adminPage.locator('[data-testid="search-input"]');
    await searchInput.fill('MOD-001');
    await adminPage.waitForTimeout(500);

    const controlIDs = await adminPage
      .locator('[data-testid="control-id"]')
      .allTextContents();

    for (const id of controlIDs) {
      expect(id).toContain('MOD-001');
    }
  });

  test('displays correct number of seeded evidence records', async ({ adminPage }) => {
    await adminPage.goto('/audit');

    const records = adminPage.locator('[data-testid="evidence-record"]');
    await expect(records.first()).toBeVisible({ timeout: 10000 });

    const count = await records.count();
    expect(count).toBeGreaterThanOrEqual(3);
  });

  test('table has correct column headers', async ({ adminPage }) => {
    await adminPage.goto('/audit');

    await expect(adminPage.locator('th:has-text("Control ID")')).toBeVisible();
    await expect(adminPage.locator('th:has-text("Decision ID")')).toBeVisible();
    await expect(adminPage.locator('th:has-text("Action")')).toBeVisible();
    await expect(adminPage.locator('th:has-text("Model")')).toBeVisible();
    await expect(adminPage.locator('th:has-text("Immutable")')).toBeVisible();
    await expect(adminPage.locator('th:has-text("Created")')).toBeVisible();
  });
});
