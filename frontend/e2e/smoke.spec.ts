import { test, expect } from '@playwright/test'

/**
 * Smoke tests: verifies the SPA mounts and renders visible content.
 * Requires the demo server to be running on BASE_URL (default: http://localhost:8181).
 */
test.describe('SPA smoke', () => {
  test('app mounts and dashboard layout renders', async ({ page }) => {
    const errors: string[] = []
    page.on('pageerror', err => errors.push(err.message))

    await page.goto('/', { waitUntil: 'networkidle' })

    await expect(page.locator('.v-app-bar')).toBeVisible({ timeout: 10000 })
    await expect(page.locator('.v-navigation-drawer')).toBeVisible()

    const main = page.locator('.v-main')
    await expect(main).toBeVisible()
    await expect(main.locator('.v-container')).toBeVisible({ timeout: 8000 })

    // Brand text in the top bar specifically (avoid strict-mode fail from footer text)
    await expect(page.locator('.v-app-bar').locator('text=IronCore').first()).toBeVisible()

    expect(errors).toHaveLength(0)
  })

  test('overview page renders stat cards', async ({ page }) => {
    await page.goto('/overview', { waitUntil: 'networkidle' })

    // At least one stat card is visible
    await expect(page.locator('.stat-card').first()).toBeVisible({ timeout: 10000 })
  })

  test('machines view renders list', async ({ page }) => {
    await page.goto('/machines', { waitUntil: 'networkidle' })
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
    // "Page not found" heading is the visible element; the "404" has opacity:0.12
    await expect(page.locator('text=Page not found')).toBeVisible({ timeout: 5000 })
  })
})
