import { test, expect } from '@playwright/test'

/**
 * Smoke test: verifies the SPA mounts and renders visible content.
 * Requires the demo server to be running on BASE_URL (default: http://localhost:8181).
 * Fails if #app stays empty or only shows a raw spinner with no layout.
 */
test.describe('SPA smoke', () => {
  test('app mounts and dashboard layout renders', async ({ page }) => {
    const errors: string[] = []
    page.on('pageerror', err => errors.push(err.message))

    await page.goto('/', { waitUntil: 'networkidle' })

    // Top bar must be present
    await expect(page.locator('.v-app-bar')).toBeVisible({ timeout: 10000 })

    // Navigation drawer must be present
    await expect(page.locator('.v-navigation-drawer')).toBeVisible()

    // Main content area must have rendered children
    const main = page.locator('.v-main')
    await expect(main).toBeVisible()
    await expect(main.locator('.v-container')).toBeVisible({ timeout: 8000 })

    // The "IronCore" brand text must be in the top bar
    await expect(page.locator('text=IronCore')).toBeVisible()

    // No uncaught JS errors
    expect(errors).toHaveLength(0)
  })

  test('overview page renders stat cards', async ({ page }) => {
    await page.goto('/overview', { waitUntil: 'networkidle' })

    // Wait for stat cards to appear (loaded from API)
    await expect(page.locator('.stat-card').first()).toBeVisible({ timeout: 8000 })

    // At least 3 stat cards visible
    const cards = page.locator('.stat-card')
    await expect(cards).toHaveCount(3, { timeout: 8000 })
  })

  test('machines view renders list', async ({ page }) => {
    await page.goto('/machines', { waitUntil: 'networkidle' })
    // Table must render with at least one row
    await expect(page.locator('table tbody tr').first()).toBeVisible({ timeout: 8000 })
  })

  test('volumes view renders list', async ({ page }) => {
    await page.goto('/volumes', { waitUntil: 'networkidle' })
    await expect(page.locator('table tbody tr').first()).toBeVisible({ timeout: 8000 })
  })

  test('networks view renders list', async ({ page }) => {
    await page.goto('/networks', { waitUntil: 'networkidle' })
    await expect(page.locator('table tbody tr').first()).toBeVisible({ timeout: 8000 })
  })

  test('404 route shows not-found page', async ({ page }) => {
    await page.goto('/definitely-does-not-exist', { waitUntil: 'networkidle' })
    await expect(page.locator('text=404')).toBeVisible({ timeout: 5000 })
  })
})
