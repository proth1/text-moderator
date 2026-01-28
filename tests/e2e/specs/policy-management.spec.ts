import { test, expect } from '../fixtures/auth';

test.describe('Policy Management', () => {
  test('admin can create a new policy', async ({ adminPage }) => {
    await adminPage.goto('/admin/policies');

    await adminPage.locator('button:has-text("Create Policy")').click();

    // Fill in policy details
    await adminPage.locator('input[name="name"]').fill('Test Community Policy');

    // Set thresholds
    await adminPage.locator('input[name="threshold-toxicity"]').fill('0.8');
    await adminPage.locator('select[name="action-toxicity"]').selectOption('block');

    await adminPage.locator('input[name="threshold-hate"]').fill('0.7');
    await adminPage.locator('select[name="action-hate"]').selectOption('block');

    await adminPage.locator('input[name="threshold-harassment"]').fill('0.75');
    await adminPage.locator('select[name="action-harassment"]').selectOption('warn');

    // Set scope
    await adminPage.locator('select[name="scope-region"]').selectOption('global');
    await adminPage.locator('select[name="scope-content-type"]').selectOption('user_generated');

    // Submit
    await adminPage.locator('button[type="submit"]').click();

    // Should show success and new policy
    await expect(adminPage.locator('[data-testid="success-message"]')).toContainText(/created/i);
    await expect(adminPage.locator('text=Test Community Policy')).toBeVisible();
    await expect(adminPage.locator('text=Draft')).toBeVisible();
  });

  test('admin can publish a draft policy', async ({ adminPage }) => {
    await adminPage.goto('/admin/policies');

    // Find draft policy
    const draftPolicy = adminPage.locator('[data-testid="policy-item"]:has-text("Relaxed Forum Policy")');
    await draftPolicy.locator('button:has-text("Publish")').click();

    // Confirm publication
    await adminPage.locator('button:has-text("Confirm")').click();

    // Should update status
    await expect(adminPage.locator('[data-testid="success-message"]')).toContainText(/published/i);
    await expect(draftPolicy).toContainText(/published/i);
  });

  test('published policy shows effective date', async ({ adminPage }) => {
    await adminPage.goto('/admin/policies');

    const publishedPolicy = adminPage.locator('[data-testid="policy-item"]:has-text("Standard Community Guidelines")');
    const effectiveDate = publishedPolicy.locator('[data-testid="effective-date"]');

    await expect(effectiveDate).toBeVisible();
    await expect(effectiveDate).not.toBeEmpty();
  });

  test('admin can create new version of existing policy', async ({ adminPage }) => {
    await adminPage.goto('/admin/policies');

    const existingPolicy = adminPage.locator('[data-testid="policy-item"]:has-text("Standard Community Guidelines")');
    await existingPolicy.locator('button:has-text("New Version")').click();

    // Should open form with current values pre-filled
    await expect(adminPage.locator('input[name="name"]')).toHaveValue('Standard Community Guidelines');

    // Modify threshold
    await adminPage.locator('input[name="threshold-toxicity"]').fill('0.85');

    await adminPage.locator('button[type="submit"]').click();

    // Should show new version
    await expect(adminPage.locator('text=Version 2')).toBeVisible();
    await expect(adminPage.locator('text=Version 1')).toBeVisible();
  });

  test('policy version history is displayed', async ({ adminPage }) => {
    await adminPage.goto('/admin/policies');

    const policy = adminPage.locator('[data-testid="policy-item"]').first();
    await policy.locator('button:has-text("View History")').click();

    // Should show version list
    const versionList = adminPage.locator('[data-testid="version-list"]');
    await expect(versionList).toBeVisible();
    await expect(versionList.locator('[data-testid="version-item"]')).toHaveCount(1); // At least 1
  });

  test('moderator can view but not edit policies', async ({ moderatorPage }) => {
    await moderatorPage.goto('/admin/policies');

    // Should see policies
    const policies = moderatorPage.locator('[data-testid="policy-item"]');
    await expect(policies).toHaveCount(3); // From seed data

    // Should not see create button
    await expect(moderatorPage.locator('button:has-text("Create Policy")')).not.toBeVisible();

    // Should not see edit buttons
    const editButtons = moderatorPage.locator('button:has-text("Edit")');
    await expect(editButtons).toHaveCount(0);
  });

  test('viewer cannot access policy management', async ({ viewerPage }) => {
    await viewerPage.goto('/admin/policies');

    // Should be denied
    await expect(viewerPage.locator('text=/Access Denied|Forbidden/i')).toBeVisible();
  });

  test('validates threshold values are between 0 and 1', async ({ adminPage }) => {
    await adminPage.goto('/admin/policies');
    await adminPage.locator('button:has-text("Create Policy")').click();

    await adminPage.locator('input[name="name"]').fill('Invalid Policy');
    await adminPage.locator('input[name="threshold-toxicity"]').fill('1.5');

    await adminPage.locator('button[type="submit"]').click();

    // Should show validation error
    await expect(adminPage.locator('[data-testid="validation-error"]')).toContainText(/between 0 and 1/i);
  });

  test('policy list shows key metrics', async ({ adminPage }) => {
    await adminPage.goto('/admin/policies');

    const policyItem = adminPage.locator('[data-testid="policy-item"]').first();

    // Should display metadata
    await expect(policyItem.locator('[data-testid="policy-name"]')).toBeVisible();
    await expect(policyItem.locator('[data-testid="policy-version"]')).toBeVisible();
    await expect(policyItem.locator('[data-testid="policy-status"]')).toBeVisible();
  });

  test('can archive a policy', async ({ adminPage }) => {
    await adminPage.goto('/admin/policies');

    const policy = adminPage.locator('[data-testid="policy-item"]').first();
    await policy.locator('button:has-text("Archive")').click();

    await adminPage.locator('button:has-text("Confirm")').click();

    await expect(adminPage.locator('[data-testid="success-message"]')).toContainText(/archived/i);
    await expect(policy).toContainText(/archived/i);
  });
});
