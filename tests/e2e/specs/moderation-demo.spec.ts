import { test, expect } from '@playwright/test';

test.describe('Moderation Demo Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/demo');
  });

  test('should display moderation demo interface', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Text Moderation Demo');
    await expect(page.locator('textarea[name="content"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('user can type text and see real-time feedback', async ({ page }) => {
    const textarea = page.locator('textarea[name="content"]');
    const feedbackArea = page.locator('[data-testid="moderation-feedback"]');

    // Type safe text
    await textarea.fill('Hello, this is a friendly message!');
    await page.waitForTimeout(500); // Wait for debounce

    // Should show positive feedback
    await expect(feedbackArea).toContainText('safe');
    await expect(feedbackArea).toHaveClass(/success|allowed|safe/);
  });

  test('warning state shows inline explanation', async ({ page }) => {
    const textarea = page.locator('textarea[name="content"]');
    const feedbackArea = page.locator('[data-testid="moderation-feedback"]');
    const explanation = page.locator('[data-testid="explanation"]');

    // Type borderline text
    await textarea.fill('I strongly disagree with your opinion and think it\'s misguided');
    await page.waitForTimeout(500);

    // Should show warning
    await expect(feedbackArea).toContainText(/warn|caution/i);
    await expect(explanation).toBeVisible();
    await expect(explanation).toContainText(/moderate|warning/i);
  });

  test('blocked state disables submission', async ({ page }) => {
    const textarea = page.locator('textarea[name="content"]');
    const submitButton = page.locator('button[type="submit"]');
    const feedbackArea = page.locator('[data-testid="moderation-feedback"]');

    // Type toxic text
    await textarea.fill('You are absolutely terrible and worthless');
    await page.waitForTimeout(500);

    // Should show blocked state
    await expect(feedbackArea).toContainText(/block|rejected/i);
    await expect(submitButton).toBeDisabled();
  });

  test('user can revise text after warning', async ({ page }) => {
    const textarea = page.locator('textarea[name="content"]');
    const submitButton = page.locator('button[type="submit"]');
    const feedbackArea = page.locator('[data-testid="moderation-feedback"]');

    // Type problematic text
    await textarea.fill('This is mildly toxic content');
    await page.waitForTimeout(500);
    await expect(feedbackArea).toContainText(/warn/i);

    // Revise to safe text
    await textarea.clear();
    await textarea.fill('This is completely safe content');
    await page.waitForTimeout(500);

    // Should now be allowed
    await expect(feedbackArea).toContainText(/safe|allow/i);
    await expect(submitButton).toBeEnabled();
  });

  test('displays category scores', async ({ page }) => {
    const textarea = page.locator('textarea[name="content"]');
    const scoresDisplay = page.locator('[data-testid="category-scores"]');

    await textarea.fill('Hello, this is a test message');
    await page.waitForTimeout(500);

    // Should display individual category scores
    await expect(scoresDisplay).toBeVisible();
    await expect(scoresDisplay).toContainText(/toxicity/i);
    await expect(scoresDisplay).toContainText(/hate/i);
    await expect(scoresDisplay).toContainText(/harassment/i);
  });

  test('real-time feedback updates as user types', async ({ page }) => {
    const textarea = page.locator('textarea[name="content"]');
    const feedbackArea = page.locator('[data-testid="moderation-feedback"]');

    // Start with safe text
    await textarea.fill('Hello');
    await page.waitForTimeout(500);
    await expect(feedbackArea).toContainText(/safe/i);

    // Add toxic content
    await textarea.fill('Hello you terrible person');
    await page.waitForTimeout(500);
    await expect(feedbackArea).toContainText(/warn|block/i);
  });

  test('handles API errors gracefully', async ({ page }) => {
    // Mock API error
    await page.route('**/api/moderate', (route) => {
      route.fulfill({
        status: 500,
        body: JSON.stringify({ error: 'Internal server error' })
      });
    });

    const textarea = page.locator('textarea[name="content"]');
    const errorMessage = page.locator('[data-testid="error-message"]');

    await textarea.fill('Test content');
    await page.waitForTimeout(500);

    await expect(errorMessage).toBeVisible();
    await expect(errorMessage).toContainText(/error|unavailable/i);
  });
});
