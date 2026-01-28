import { test, expect } from '../fixtures/auth';

test.describe('Moderator Queue', () => {
  test.use({ storageState: 'moderator-auth.json' });

  test.beforeEach(async ({ moderatorPage }) => {
    await moderatorPage.goto('/moderation/queue');
  });

  test('queue displays pending decisions', async ({ moderatorPage }) => {
    await expect(moderatorPage.locator('h1')).toContainText('Moderation Queue');

    const queueItems = moderatorPage.locator('[data-testid="queue-item"]');
    await expect(queueItems).toHaveCount(2); // From seed data

    // Each item should show key information
    const firstItem = queueItems.first();
    await expect(firstItem.locator('[data-testid="category-scores"]')).toBeVisible();
    await expect(firstItem.locator('[data-testid="automated-action"]')).toBeVisible();
    await expect(firstItem.locator('[data-testid="timestamp"]')).toBeVisible();
  });

  test('moderator can approve a decision', async ({ moderatorPage }) => {
    const firstItem = moderatorPage.locator('[data-testid="queue-item"]').first();
    const approveButton = firstItem.locator('button:has-text("Approve")');
    const rationaleInput = moderatorPage.locator('textarea[name="rationale"]');

    await approveButton.click();

    // Should show rationale dialog
    await expect(rationaleInput).toBeVisible();
    await rationaleInput.fill('Confirmed violation of community guidelines');

    // Submit approval
    await moderatorPage.locator('button:has-text("Submit")').click();

    // Should show success message
    await expect(moderatorPage.locator('[data-testid="success-message"]')).toBeVisible();

    // Queue should update
    const queueItems = moderatorPage.locator('[data-testid="queue-item"]');
    await expect(queueItems).toHaveCount(1); // One less item
  });

  test('moderator can reject a decision', async ({ moderatorPage }) => {
    const firstItem = moderatorPage.locator('[data-testid="queue-item"]').first();
    const rejectButton = firstItem.locator('button:has-text("Reject")');

    await rejectButton.click();

    const rationaleInput = moderatorPage.locator('textarea[name="rationale"]');
    await expect(rationaleInput).toBeVisible();
    await rationaleInput.fill('False positive - context shows this is acceptable');

    await moderatorPage.locator('button:has-text("Submit")').click();

    await expect(moderatorPage.locator('[data-testid="success-message"]')).toBeVisible();
  });

  test('moderator can escalate a decision', async ({ moderatorPage }) => {
    const firstItem = moderatorPage.locator('[data-testid="queue-item"]').first();
    const escalateButton = firstItem.locator('button:has-text("Escalate")');

    await escalateButton.click();

    const rationaleInput = moderatorPage.locator('textarea[name="rationale"]');
    await rationaleInput.fill('Requires admin review - complex case');

    await moderatorPage.locator('button:has-text("Submit")').click();

    await expect(moderatorPage.locator('[data-testid="success-message"]')).toContainText(/escalated/i);
  });

  test('queue updates after action', async ({ moderatorPage }) => {
    const initialCount = await moderatorPage.locator('[data-testid="queue-item"]').count();

    const firstItem = moderatorPage.locator('[data-testid="queue-item"]').first();
    await firstItem.locator('button:has-text("Approve")').click();

    const rationaleInput = moderatorPage.locator('textarea[name="rationale"]');
    await rationaleInput.fill('Confirmed');
    await moderatorPage.locator('button:has-text("Submit")').click();

    // Wait for queue to refresh
    await moderatorPage.waitForTimeout(1000);

    const newCount = await moderatorPage.locator('[data-testid="queue-item"]').count();
    expect(newCount).toBe(initialCount - 1);
  });

  test('displays content hash instead of full content', async ({ moderatorPage }) => {
    const firstItem = moderatorPage.locator('[data-testid="queue-item"]').first();
    const contentDisplay = firstItem.locator('[data-testid="content-hash"]');

    await expect(contentDisplay).toBeVisible();
    await expect(contentDisplay).toContainText(/hash_/);
  });

  test('filters queue by action type', async ({ moderatorPage }) => {
    const filterSelect = moderatorPage.locator('select[name="action-filter"]');

    await filterSelect.selectOption('block');

    const queueItems = moderatorPage.locator('[data-testid="queue-item"]');
    const actions = await queueItems.locator('[data-testid="automated-action"]').allTextContents();

    for (const action of actions) {
      expect(action.toLowerCase()).toContain('block');
    }
  });

  test('sorts queue by timestamp', async ({ moderatorPage }) => {
    const sortSelect = moderatorPage.locator('select[name="sort"]');

    await sortSelect.selectOption('oldest');

    const timestamps = await moderatorPage
      .locator('[data-testid="timestamp"]')
      .allTextContents();

    // Verify ascending order
    const dates = timestamps.map(t => new Date(t).getTime());
    for (let i = 1; i < dates.length; i++) {
      expect(dates[i]).toBeGreaterThanOrEqual(dates[i - 1]);
    }
  });

  test('viewer cannot access queue', async ({ viewerPage }) => {
    await viewerPage.goto('/moderation/queue');

    // Should be redirected or show access denied
    await expect(viewerPage.locator('text=/Access Denied|Forbidden/i')).toBeVisible();
  });
});
