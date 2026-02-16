const { test, expect, request } = require('@playwright/test');

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:3000/api';
const DASHBOARD_BASE_URL = process.env.DASHBOARD_BASE_URL || 'http://localhost:3001';

async function registerUser(apiRequest, role) {
  const email = `${role}.${Date.now()}@example.com`;
  const payload = {
    email,
    password: 'Test123!',
    role,
    companyName: 'TestCo',
  };
  const response = await apiRequest.post(`${API_BASE_URL}/auth/register`, { data: payload });
  expect(response.ok()).toBeTruthy();
  const data = await response.json();
  return { email, password: payload.password, auth: data };
}

async function authenticate(page, apiRequest, role) {
  const user = await registerUser(apiRequest, role);
  const token = user.auth.access_token;
  const roleValue = user.auth.user?.role || role;
  const cookieUrl = DASHBOARD_BASE_URL.endsWith('/')
    ? DASHBOARD_BASE_URL
    : `${DASHBOARD_BASE_URL}/`;

  await page.context().addCookies([
    {
      name: 'auth-token',
      value: token,
      url: cookieUrl,
    },
    {
      name: 'user-role',
      value: roleValue,
      url: cookieUrl,
    },
  ]);

  await page.addInitScript(({ authToken, userRole, user }) => {
    localStorage.setItem('auth_token', authToken);
    localStorage.setItem('user', JSON.stringify(user));
    document.cookie = `auth-token=${authToken}; path=/; max-age=86400; SameSite=Strict`;
    document.cookie = `user-role=${userRole}; path=/; max-age=86400; SameSite=Strict`;
  }, { authToken: token, userRole: roleValue, user: user.auth.user });

  return user;
}

test('admin role redirects to /admin', async ({ page, request: apiRequest }) => {
  await authenticate(page, apiRequest, 'admin');
  await page.goto('/admin');
  await expect(page).toHaveURL(/\/admin/);
});

test('advertiser role redirects to /client', async ({ page, request: apiRequest }) => {
  await authenticate(page, apiRequest, 'advertiser');
  await page.goto('/client');
  await expect(page).toHaveURL(/\/client/);
});

test('publisher role redirects to /publisher', async ({ page, request: apiRequest }) => {
  await authenticate(page, apiRequest, 'publisher');
  await page.goto('/publisher');
  await expect(page).toHaveURL(/\/publisher/);
});

test('core pages render for authenticated admin', async ({ page, request: apiRequest }) => {
  await authenticate(page, apiRequest, 'admin');

  const routes = [
    '/admin',
    '/dashboard',
    '/client',
    '/publisher',
    '/dsp',
    '/docs',
  ];

  for (const route of routes) {
    await page.goto(route);
    await expect(page).toHaveURL(new RegExp(route));
    await expect(page.locator('body')).toBeVisible();
  }
});
