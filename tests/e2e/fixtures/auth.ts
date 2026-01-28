import { test as base, Page } from '@playwright/test';

/**
 * User credentials for testing - matches seed-data.sql
 */
export const users = {
  admin: {
    id: 'a0000000-0000-0000-0000-000000000001',
    email: 'admin@civitas.test',
    apiKey: 'tk_admin_test_key_001',
    role: 'admin',
    created_at: '2026-01-01T00:00:00.000Z',
  },
  moderator: {
    id: 'a0000000-0000-0000-0000-000000000002',
    email: 'moderator@civitas.test',
    apiKey: 'tk_mod_test_key_002',
    role: 'moderator',
    created_at: '2026-01-01T00:00:00.000Z',
  },
  viewer: {
    id: 'a0000000-0000-0000-0000-000000000003',
    email: 'viewer@civitas.test',
    apiKey: 'tk_viewer_test_key_003',
    role: 'viewer',
    created_at: '2026-01-01T00:00:00.000Z',
  },
} as const;

export type UserRole = keyof typeof users;

/**
 * Extended test fixtures with authenticated pages
 */
type AuthFixtures = {
  adminPage: Page;
  moderatorPage: Page;
  viewerPage: Page;
};

/**
 * Authenticate a page by setting localStorage values the app reads on boot.
 * The app's authStore reads `api_key` and `user` from localStorage.
 */
async function authenticatePage(page: Page, userRole: UserRole): Promise<void> {
  const user = users[userRole];

  // Navigate to origin so we can set localStorage (needs same origin)
  await page.goto('/login');
  await page.waitForLoadState('domcontentloaded');

  // Set localStorage with the format the app expects
  await page.evaluate(
    ({ apiKey, userData }) => {
      localStorage.setItem('api_key', apiKey);
      localStorage.setItem(
        'user',
        JSON.stringify({
          id: userData.id,
          email: userData.email,
          role: userData.role,
          created_at: userData.created_at,
        })
      );
    },
    { apiKey: user.apiKey, userData: user }
  );

  // Navigate to dashboard - checkAuth will pick up localStorage values
  await page.goto('/');

  // Wait for the sidebar to render (confirms we're authenticated and in MainLayout)
  await page.waitForSelector('nav', { timeout: 10000 });
}

/**
 * Extended test with authenticated user fixtures
 */
export const test = base.extend<AuthFixtures>({
  adminPage: async ({ browser }, use) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    await authenticatePage(page, 'admin');
    await use(page);
    await context.close();
  },

  moderatorPage: async ({ browser }, use) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    await authenticatePage(page, 'moderator');
    await use(page);
    await context.close();
  },

  viewerPage: async ({ browser }, use) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    await authenticatePage(page, 'viewer');
    await use(page);
    await context.close();
  },
});

export { expect } from '@playwright/test';
