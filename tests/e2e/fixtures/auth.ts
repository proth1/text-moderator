import { test as base, Page } from '@playwright/test';

/**
 * User credentials for testing
 */
export const users = {
  admin: {
    email: 'admin@civitas.test',
    apiKey: 'tk_admin_test_key_001',
    role: 'admin'
  },
  moderator: {
    email: 'moderator@civitas.test',
    apiKey: 'tk_mod_test_key_002',
    role: 'moderator'
  },
  viewer: {
    email: 'viewer@civitas.test',
    apiKey: 'tk_viewer_test_key_003',
    role: 'viewer'
  }
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
 * Authenticate a page with user credentials
 */
async function authenticatePage(page: Page, userRole: UserRole): Promise<void> {
  const user = users[userRole];

  // Set authentication cookie or localStorage
  await page.context().addCookies([
    {
      name: 'auth_token',
      value: user.apiKey,
      domain: 'localhost',
      path: '/',
      httpOnly: true,
      secure: false,
      sameSite: 'Lax'
    }
  ]);

  // Alternatively, set localStorage if the app uses that
  await page.goto('/');
  await page.evaluate((userData) => {
    localStorage.setItem('user', JSON.stringify(userData));
  }, user);
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
  }
});

export { expect } from '@playwright/test';
