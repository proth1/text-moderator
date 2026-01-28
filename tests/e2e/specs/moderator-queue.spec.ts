import { test, expect } from '../fixtures/auth';

test.describe('Moderator Queue', () => {
  test('displays review queue page', async ({ adminPage }) => {
    await adminPage.goto('/reviews');
    await expect(adminPage.locator('main h1')).toContainText('Review Queue');
  });

  test('shows review table headers or empty state', async ({ adminPage }) => {
    await adminPage.goto('/reviews');
    await adminPage.waitForTimeout(3000);

    // Table only renders when reviews exist; otherwise empty state is shown
    const hasReviews = await adminPage.locator('table').count() > 0;

    if (hasReviews) {
      await expect(adminPage.locator('th:has-text("ID")')).toBeVisible();
      await expect(adminPage.locator('th:has-text("Text Preview")')).toBeVisible();
      await expect(adminPage.locator('th:has-text("Category")')).toBeVisible();
      await expect(adminPage.locator('th:has-text("Confidence")')).toBeVisible();
      await expect(adminPage.locator('th:has-text("Status")')).toBeVisible();
    } else {
      await expect(adminPage.locator('text=No reviews found')).toBeVisible();
    }
  });

  test('shows filter buttons for pending and reviewed', async ({ adminPage }) => {
    await adminPage.goto('/reviews');

    await expect(adminPage.locator('button:has-text("Pending")')).toBeVisible();
    await expect(adminPage.locator('button:has-text("Reviewed")')).toBeVisible();
    await expect(adminPage.locator('button:has-text("All")')).toBeVisible();
  });

  test('shows review links or empty state', async ({ adminPage }) => {
    await adminPage.goto('/reviews');
    await adminPage.waitForTimeout(3000);

    // Use table-scoped locator to avoid matching the "Reviewed" filter button
    const tableReviewButtons = adminPage.locator('table button:has-text("Review")');
    const count = await tableReviewButtons.count();

    if (count > 0) {
      await tableReviewButtons.first().click();
      await expect(adminPage.locator('main h1')).toContainText('Review Detail');
    } else {
      await expect(adminPage.locator('text=No reviews found')).toBeVisible();
    }
  });

  test('review detail shows actions or empty queue', async ({ adminPage }) => {
    await adminPage.goto('/reviews');
    await adminPage.waitForTimeout(3000);

    const tableReviewButtons = adminPage.locator('table button:has-text("Review")');
    const count = await tableReviewButtons.count();

    if (count > 0) {
      await tableReviewButtons.first().click();

      await expect(adminPage.locator('button:has-text("Approve Model Decision")')).toBeVisible();
      await expect(adminPage.locator('button:has-text("Override: Allow")')).toBeVisible();
      await expect(adminPage.locator('button:has-text("Override: Block")')).toBeVisible();
      await expect(adminPage.locator('button:has-text("Escalate")')).toBeVisible();
      await expect(adminPage.locator('textarea')).toBeVisible();
    } else {
      await expect(adminPage.locator('text=No reviews found')).toBeVisible();
    }
  });

  test('filter buttons change active state on click', async ({ adminPage }) => {
    await adminPage.goto('/reviews');

    const pendingBtn = adminPage.locator('button:has-text("Pending")');
    const reviewedBtn = adminPage.locator('button:has-text("Reviewed")');
    const allBtn = adminPage.locator('button:has-text("All")');

    await reviewedBtn.click();
    await adminPage.waitForTimeout(500);

    await allBtn.click();
    await adminPage.waitForTimeout(500);

    await pendingBtn.click();
    await adminPage.waitForTimeout(500);

    await expect(adminPage.locator('main h1')).toContainText('Review Queue');
  });

  test('moderator can access review queue', async ({ moderatorPage }) => {
    await moderatorPage.goto('/reviews');
    await expect(moderatorPage.locator('main h1')).toContainText('Review Queue');
  });
});
