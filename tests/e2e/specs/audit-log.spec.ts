import { test, expect } from '../fixtures/auth';

test.describe('Audit Log', () => {
  test('admin can view evidence records', async ({ adminPage }) => {
    await adminPage.goto('/admin/audit');

    await expect(adminPage.locator('h1')).toContainText('Audit Log');

    const evidenceRecords = adminPage.locator('[data-testid="evidence-record"]');
    await expect(evidenceRecords.first()).toBeVisible();
  });

  test('evidence records display required fields', async ({ adminPage }) => {
    await adminPage.goto('/admin/audit');

    const firstRecord = adminPage.locator('[data-testid="evidence-record"]').first();

    // Should show all required audit fields
    await expect(firstRecord.locator('[data-testid="control-id"]')).toBeVisible();
    await expect(firstRecord.locator('[data-testid="policy-id"]')).toBeVisible();
    await expect(firstRecord.locator('[data-testid="decision-id"]')).toBeVisible();
    await expect(firstRecord.locator('[data-testid="timestamp"]')).toBeVisible();
    await expect(firstRecord.locator('[data-testid="immutable-badge"]')).toContainText('Immutable');
  });

  test('search evidence by control ID', async ({ adminPage }) => {
    await adminPage.goto('/admin/audit');

    const searchInput = adminPage.locator('input[name="control-id-search"]');
    await searchInput.fill('MOD-001');

    await adminPage.locator('button:has-text("Search")').click();

    // All results should have MOD-001
    const controlIDs = await adminPage
      .locator('[data-testid="control-id"]')
      .allTextContents();

    for (const id of controlIDs) {
      expect(id).toContain('MOD-001');
    }
  });

  test('filter evidence by date range', async ({ adminPage }) => {
    await adminPage.goto('/admin/audit');

    const startDate = adminPage.locator('input[name="start-date"]');
    const endDate = adminPage.locator('input[name="end-date"]');

    await startDate.fill('2026-01-01');
    await endDate.fill('2026-01-31');

    await adminPage.locator('button:has-text("Filter")').click();

    // Should show filtered results
    const records = adminPage.locator('[data-testid="evidence-record"]');
    await expect(records).toHaveCountGreaterThan(0);
  });

  test('export evidence records', async ({ adminPage }) => {
    await adminPage.goto('/admin/audit');

    const downloadPromise = adminPage.waitForEvent('download');

    await adminPage.locator('button:has-text("Export CSV")').click();

    const download = await downloadPromise;

    // Verify download
    expect(download.suggestedFilename()).toMatch(/evidence.*\.csv/);
  });

  test('export produces correct CSV format', async ({ adminPage }) => {
    await adminPage.goto('/admin/audit');

    const downloadPromise = adminPage.waitForEvent('download');
    await adminPage.locator('button:has-text("Export CSV")').click();
    const download = await downloadPromise;

    // Save and read file
    const path = await download.path();
    const fs = require('fs');
    const content = fs.readFileSync(path, 'utf-8');

    // Verify CSV headers
    expect(content).toContain('control_id');
    expect(content).toContain('policy_id');
    expect(content).toContain('decision_id');
    expect(content).toContain('created_at');
    expect(content).toContain('immutable');
  });

  test('evidence records are sorted by timestamp descending', async ({ adminPage }) => {
    await adminPage.goto('/admin/audit');

    const timestamps = await adminPage
      .locator('[data-testid="timestamp"]')
      .allTextContents();

    // Verify descending order
    const dates = timestamps.map(t => new Date(t).getTime());
    for (let i = 1; i < dates.length; i++) {
      expect(dates[i]).toBeLessThanOrEqual(dates[i - 1]);
    }
  });

  test('pagination works correctly', async ({ adminPage }) => {
    await adminPage.goto('/admin/audit');

    const pageInfo = adminPage.locator('[data-testid="page-info"]');
    await expect(pageInfo).toContainText('Page 1');

    // Go to next page
    await adminPage.locator('button:has-text("Next")').click();

    await expect(pageInfo).toContainText('Page 2');
  });

  test('moderator can view audit log read-only', async ({ moderatorPage }) => {
    await moderatorPage.goto('/admin/audit');

    // Should see records
    const records = moderatorPage.locator('[data-testid="evidence-record"]');
    await expect(records.first()).toBeVisible();

    // Should not see delete or edit buttons
    await expect(moderatorPage.locator('button:has-text("Delete")')).not.toBeVisible();
    await expect(moderatorPage.locator('button:has-text("Edit")')).not.toBeVisible();
  });

  test('viewer cannot access audit log', async ({ viewerPage }) => {
    await viewerPage.goto('/admin/audit');

    await expect(viewerPage.locator('text=/Access Denied|Forbidden/i')).toBeVisible();
  });

  test('evidence detail view shows full information', async ({ adminPage }) => {
    await adminPage.goto('/admin/audit');

    const firstRecord = adminPage.locator('[data-testid="evidence-record"]').first();
    await firstRecord.locator('button:has-text("View Details")').click();

    // Should show modal with full details
    const modal = adminPage.locator('[data-testid="evidence-detail-modal"]');
    await expect(modal).toBeVisible();

    await expect(modal.locator('[data-testid="full-category-scores"]')).toBeVisible();
    await expect(modal.locator('[data-testid="submission-hash"]')).toBeVisible();
    await expect(modal.locator('[data-testid="model-name"]')).toBeVisible();
    await expect(modal.locator('[data-testid="model-version"]')).toBeVisible();
  });

  test('evidence chain shows related records', async ({ adminPage }) => {
    await adminPage.goto('/admin/audit');

    // Click on a record with related evidence
    const recordWithReview = adminPage.locator('[data-testid="evidence-record"]:has-text("hash_hate_001")').first();
    await recordWithReview.locator('button:has-text("View Chain")').click();

    // Should show all related evidence
    const chainModal = adminPage.locator('[data-testid="evidence-chain-modal"]');
    await expect(chainModal).toBeVisible();

    const chainItems = chainModal.locator('[data-testid="chain-item"]');
    await expect(chainItems).toHaveCountGreaterThanOrEqual(2); // MOD-001 and GOV-002
  });
});
