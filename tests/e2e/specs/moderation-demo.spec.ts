import { test, expect } from '@playwright/test';

test.describe('Moderation Demo Page', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to login page first so we can set localStorage
    await page.goto('/login');
    await page.waitForLoadState('domcontentloaded');

    // Set up auto-login with admin credentials
    await page.evaluate(() => {
      localStorage.setItem('api_key', 'tk_admin_test_key_001');
      localStorage.setItem(
        'user',
        JSON.stringify({
          id: 'a0000000-0000-0000-0000-000000000001',
          email: 'admin@civitas.test',
          role: 'admin',
          created_at: '2026-01-01T00:00:00.000Z',
        })
      );
    });

    // Navigate to demo page
    await page.goto('/demo');
    await page.waitForSelector('nav', { timeout: 10000 });
  });

  test('should display moderation demo interface', async ({ page }) => {
    await expect(page.locator('main h1')).toContainText('Moderation Demo');
    await expect(page.locator('textarea[name="content"]')).toBeVisible();
    await expect(page.locator('button:has-text("Submit for Moderation")')).toBeVisible();
  });

  test('submit button is disabled when textarea is empty', async ({ page }) => {
    const submitButton = page.locator('button:has-text("Submit for Moderation")');
    await expect(submitButton).toBeDisabled();
  });

  test('submit button enables when text is entered', async ({ page }) => {
    const textarea = page.locator('textarea[name="content"]');
    const submitButton = page.locator('button:has-text("Submit for Moderation")');

    await textarea.fill('Hello, this is a friendly message!');
    await expect(submitButton).toBeEnabled();
  });

  test('submitting safe text shows allow action', async ({ page }) => {
    const textarea = page.locator('textarea[name="content"]');
    const submitButton = page.locator('button:has-text("Submit for Moderation")');

    await textarea.fill('Hello, this is a friendly message!');
    await submitButton.click();

    // Wait for results to appear (real HuggingFace API call)
    const feedback = page.locator('[data-testid="moderation-feedback"]');
    await expect(feedback).toBeVisible({ timeout: 30000 });

    // Should show ALLOW action for safe content
    await expect(feedback).toContainText(/ALLOW/i);
  });

  test('submitting toxic text shows block action', async ({ page }) => {
    const textarea = page.locator('textarea[name="content"]');
    const submitButton = page.locator('button:has-text("Submit for Moderation")');

    await textarea.fill('You are absolutely terrible and worthless');
    await submitButton.click();

    // Wait for results (real HuggingFace API call)
    const feedback = page.locator('[data-testid="moderation-feedback"]');
    await expect(feedback).toBeVisible({ timeout: 30000 });

    // Should show BLOCK action for toxic content
    await expect(feedback).toContainText(/BLOCK/i);
  });

  test('displays category scores after submission', async ({ page }) => {
    const textarea = page.locator('textarea[name="content"]');
    const submitButton = page.locator('button:has-text("Submit for Moderation")');

    await textarea.fill('Hello, this is a test message');
    await submitButton.click();

    const scoresDisplay = page.locator('[data-testid="category-scores"]');
    await expect(scoresDisplay).toBeVisible({ timeout: 30000 });

    // Should display toxicity category
    await expect(scoresDisplay).toContainText(/toxicity/i);
  });

  test('clear button resets the form', async ({ page }) => {
    const textarea = page.locator('textarea[name="content"]');
    const clearButton = page.locator('button:has-text("Clear")');

    await textarea.fill('Some text to clear');
    await expect(textarea).toHaveValue('Some text to clear');

    await clearButton.click();
    await expect(textarea).toHaveValue('');
  });

  test('handles API errors gracefully', async ({ page }) => {
    // Intercept the moderation API call and return an error
    await page.route('**/api/v1/moderate', (route) => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal server error' }),
      });
    });

    const textarea = page.locator('textarea[name="content"]');
    const submitButton = page.locator('button:has-text("Submit for Moderation")');

    await textarea.fill('Test content');
    await submitButton.click();

    const errorMessage = page.locator('[data-testid="error-message"]');
    await expect(errorMessage).toBeVisible({ timeout: 10000 });
    await expect(errorMessage).toContainText(/failed/i);
  });
});
