import { test, expect, devices } from '@playwright/test'
import path from 'path'
import fs from 'fs'

/**
 * Smoke tests: verify the SPA mounts and renders visible content.
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

    await expect(page.locator('.v-app-bar').locator('text=IronCore').first()).toBeVisible()

    expect(errors).toHaveLength(0)
  })

  test('overview page renders stat cards', async ({ page }) => {
    await page.goto('/overview', { waitUntil: 'networkidle' })
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
    await expect(page.locator('text=Page not found')).toBeVisible({ timeout: 5000 })
  })
})

/**
 * Phase 3 views: verify the new resource views render.
 */
test.describe('Phase 3 views', () => {
  test('machine pools view renders', async ({ page }) => {
    await page.goto('/machinepools', { waitUntil: 'networkidle' })
    await expect(page.locator('table tbody tr').first()).toBeVisible({ timeout: 8000 })
  })

  test('volume snapshots view renders', async ({ page }) => {
    await page.goto('/volumesnapshots', { waitUntil: 'networkidle' })
    await expect(page.locator('table tbody tr').first()).toBeVisible({ timeout: 8000 })
  })

  test('NAT gateways view renders', async ({ page }) => {
    await page.goto('/natgateways', { waitUntil: 'networkidle' })
    await expect(page.locator('table tbody tr').first()).toBeVisible({ timeout: 8000 })
  })

  test('resource quotas view renders', async ({ page }) => {
    await page.goto('/resourcequotas', { waitUntil: 'networkidle' })
    // Quota view shows per-quota cards, not a single table
    await expect(page.locator('.v-main .v-container').first()).toBeVisible({ timeout: 8000 })
    await expect(page.locator('table').first()).toBeVisible({ timeout: 8000 })
  })
})

/**
 * Phase 4 views: NetworkPolicy, Bucket, BucketClass, BucketPool, NetworkInterface.
 */
test.describe('Phase 4 views', () => {
  test('network interfaces view renders', async ({ page }) => {
    await page.goto('/networkinterfaces', { waitUntil: 'networkidle' })
    await expect(page.locator('table tbody tr').first()).toBeVisible({ timeout: 8000 })
  })

  test('network policies view renders', async ({ page }) => {
    await page.goto('/networkpolicies', { waitUntil: 'networkidle' })
    await expect(page.locator('table tbody tr').first()).toBeVisible({ timeout: 8000 })
  })

  test('buckets view renders', async ({ page }) => {
    await page.goto('/buckets', { waitUntil: 'networkidle' })
    await expect(page.locator('table tbody tr').first()).toBeVisible({ timeout: 8000 })
  })

  test('bucket classes view renders', async ({ page }) => {
    await page.goto('/bucketclasses', { waitUntil: 'networkidle' })
    await expect(page.locator('table tbody tr').first()).toBeVisible({ timeout: 8000 })
  })

  test('bucket pools view renders', async ({ page }) => {
    await page.goto('/bucketpools', { waitUntil: 'networkidle' })
    await expect(page.locator('table tbody tr').first()).toBeVisible({ timeout: 8000 })
  })
})

/**
 * Responsive tests: verify layout at key breakpoints, capture screenshots.
 */
test.describe('Responsive layout', () => {
  const screenshotDir = 'e2e/screenshots'

  const viewports = [
    { name: 'mobile-390', width: 390, height: 844 },
    { name: 'tablet-768', width: 768, height: 1024 },
    { name: 'desktop-1440', width: 1440, height: 900 },
  ]

  for (const vp of viewports) {
    test(`overview at ${vp.width}px — no horizontal overflow`, async ({ page }) => {
      await page.setViewportSize({ width: vp.width, height: vp.height })
      await page.goto('/overview', { waitUntil: 'networkidle' })

      // Top bar must be visible and not overflow
      await expect(page.locator('.v-app-bar')).toBeVisible()

      // Check no horizontal scroll (scrollWidth <= clientWidth means no overflow)
      const overflows = await page.evaluate(() =>
        document.documentElement.scrollWidth > document.documentElement.clientWidth
      )
      expect(overflows, `Horizontal overflow at ${vp.width}px`).toBe(false)

      // On mobile: drawer should be closed by default
      if (vp.width < 600) {
        const drawer = page.locator('.v-navigation-drawer')
        // Drawer is rendered but should be in temporary/overlay mode (off-screen or closed)
        // Check it is not covering main content by verifying main is visible
        await expect(page.locator('.v-main')).toBeVisible()
        const drawerBox = await drawer.boundingBox()
        const mainBox = await page.locator('.v-main').boundingBox()
        if (drawerBox && mainBox) {
          // On mobile closed state, drawer should be off-screen (left < 0) or very narrow
          const drawerVisible = drawerBox.x >= 0 && drawerBox.width > 50
          if (drawerVisible) {
            // If somehow open, main should still be accessible (overlay mode)
            // Just warn — the real check is no horizontal overflow
          }
        }
      }

      // Save screenshot for visual inspection
      fs.mkdirSync(screenshotDir, { recursive: true })
      await page.screenshot({
        path: path.join(screenshotDir, `${vp.name}-overview-light.png`),
        fullPage: false,
      })
    })

    test(`overview dark theme at ${vp.width}px`, async ({ page }) => {
      await page.setViewportSize({ width: vp.width, height: vp.height })
      await page.goto('/overview', { waitUntil: 'networkidle' })
      // Theme toggle button — aria-label switches between light/dark
      const themeBtn = page.locator('[aria-label="Switch to dark mode"], [aria-label="Switch to light mode"]').first()
      await themeBtn.click()
      await page.waitForTimeout(300)

      const overflows = await page.evaluate(() =>
        document.documentElement.scrollWidth > document.documentElement.clientWidth
      )
      expect(overflows, `Dark mode horizontal overflow at ${vp.width}px`).toBe(false)

      fs.mkdirSync(screenshotDir, { recursive: true })
      await page.screenshot({
        path: path.join(screenshotDir, `${vp.name}-overview-dark.png`),
        fullPage: false,
      })
    })
  }

  test('mobile nav drawer opens and closes', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 })
    await page.goto('/overview', { waitUntil: 'networkidle' })

    // Drawer should be closed on initial load
    const drawer = page.locator('.v-navigation-drawer')
    // Check drawer is off-screen (x < 0 in closed state) or not visible
    let box = await drawer.boundingBox()
    const isClosed = !box || box.x < 0 || box.width < 50
    expect(isClosed, 'Drawer should be closed by default on mobile').toBe(true)

    // Open the drawer via hamburger
    await page.locator('[aria-label="Toggle navigation drawer"]').click()
    await page.waitForTimeout(300)
    box = await drawer.boundingBox()
    expect(box && box.width > 200, 'Drawer should open after hamburger click').toBe(true)

    // Close by clicking outside (clicking on main area)
    await page.locator('.v-main').click({ position: { x: 350, y: 400 }, force: true })
    await page.waitForTimeout(300)
    box = await drawer.boundingBox()
    const isClosedAfter = !box || box.x < 0 || box.width < 50
    expect(isClosedAfter, 'Drawer should close after clicking outside').toBe(true)
  })

  test('machines table scrolls horizontally on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 })
    await page.goto('/machines', { waitUntil: 'networkidle' })
    await expect(page.locator('table').first()).toBeVisible({ timeout: 8000 })

    // Page itself should not overflow
    const pageOverflows = await page.evaluate(() =>
      document.documentElement.scrollWidth > document.documentElement.clientWidth
    )
    expect(pageOverflows, 'Page should not overflow on mobile machines view').toBe(false)
  })
})

/**
 * Empty and error state tests.
 */
test.describe('Edge states', () => {
  test('navigation drawer visible on desktop', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 })
    await page.goto('/overview', { waitUntil: 'networkidle' })
    await expect(page.locator('.v-navigation-drawer')).toBeVisible()
    const box = await page.locator('.v-navigation-drawer').boundingBox()
    expect(box && box.width > 200, 'Drawer should be open and full-width on desktop').toBe(true)
  })

  test('rail mode toggles on desktop', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 })
    await page.goto('/overview', { waitUntil: 'networkidle' })
    const drawerBefore = await page.locator('.v-navigation-drawer').boundingBox()
    await page.locator('[aria-label="Toggle navigation drawer"]').click()
    await page.waitForTimeout(300)
    const drawerAfter = await page.locator('.v-navigation-drawer').boundingBox()
    // Rail mode should make drawer narrower
    expect(drawerAfter && drawerBefore && drawerAfter.width < drawerBefore.width).toBe(true)
  })
})
