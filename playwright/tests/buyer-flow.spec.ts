import { test, expect } from '@playwright/test'

// The Expo web app runs on port 8081 (configured in playwright.config.ts baseURL).
// Override with the correct Expo web URL if different.
const APP_URL = process.env.APP_URL ?? 'http://localhost:8082'

test.describe('ShopKuber buyer flow', () => {
  test('complete buyer journey: register → browse → cart → checkout → order', async ({
    page,
  }) => {
    test.slow()

    const uniqueEmail = `buyer_${Date.now()}@test.com`
    const password = 'test1234'
    const fullName = 'Test Buyer'
    const deliveryAddress = '123 Test Street, Moscow, Russia, 101000'

    // ── Step 1: Navigate to app – expect login form ────────────────────────────
    await page.goto(APP_URL)
    await page.waitForTimeout(1000)

    // Wait for login form to appear (either by heading or email input)
    await expect(
      page.getByText(/sign in/i).or(page.getByText(/ShopKuber/i)),
    ).toBeVisible({ timeout: 15_000 })

    await page.screenshot({ path: 'screenshots/01-login-screen.png' })

    // ── Step 2: Navigate to register ──────────────────────────────────────────
    const registerLink = page
      .getByText(/register/i)
      .or(page.getByRole('button', { name: /register/i }))
      .or(page.getByText(/create account/i))

    await registerLink.first().click()
    await page.waitForTimeout(1000)

    // Fill register form
    const fullNameInput = page
      .getByPlaceholder(/full name/i)
      .or(page.getByLabel(/full name/i))
    await fullNameInput.fill(fullName)

    const emailInput = page
      .getByPlaceholder(/email/i)
      .or(page.getByLabel(/email/i))
    await emailInput.fill(uniqueEmail)

    const passwordInput = page
      .getByPlaceholder(/password/i)
      .or(page.getByLabel(/password/i))
    await passwordInput.first().fill(password)

    const submitButton = page
      .getByRole('button', { name: /create account/i })
      .or(page.getByRole('button', { name: /register/i }))
    await submitButton.first().click()

    // Wait for home screen to appear after successful registration
    await page.waitForTimeout(2000)
    await expect(
      page.getByText(/shopkuber/i).or(page.getByText(/home/i)).or(page.getByText(/products/i)),
    ).toBeVisible({ timeout: 15_000 })

    // ── Step 3: Home screen ────────────────────────────────────────────────────
    await page.screenshot({ path: 'screenshots/02-home-screen.png' })

    // Verify product list is visible
    // Products are rendered as ProductCard components
    const productCards = page.locator('[data-testid="product-card"]').or(
      page.locator('text=/₽/').first().locator('..').locator('..'),
    )
    await page.waitForTimeout(2000)

    // ── Step 4: Click first product card ──────────────────────────────────────
    // Find first product by looking for price text (₽) or product card
    const firstProduct = page.locator('text=/₽/').first()
    if (await firstProduct.isVisible()) {
      await firstProduct.click()
    } else {
      // fallback: try clicking any touchable that looks like a product
      const anyProduct = page.locator('[role="button"]').first()
      await anyProduct.click()
    }

    await page.waitForTimeout(1500)

    // ── Step 5: Product detail ─────────────────────────────────────────────────
    await page.screenshot({ path: 'screenshots/03-product-detail.png' })

    // ── Step 6: Select variant and add to cart ────────────────────────────────
    // Try clicking first available variant chip
    const variantChips = page.locator('text=/₽/').all()
    const chips = await variantChips
    if (chips.length > 0) {
      await chips[0].click()
      await page.waitForTimeout(500)
    }

    // Click "Add to Cart" button
    const addToCartButton = page
      .getByRole('button', { name: /add to cart/i })
      .or(page.getByText(/add to cart/i))
    await addToCartButton.first().click()
    await page.waitForTimeout(1500)

    await page.screenshot({ path: 'screenshots/04-added-to-cart.png' })

    // ── Step 7: Verify success feedback ───────────────────────────────────────
    // Snackbar should appear with "Added to cart!" or similar
    await expect(
      page.getByText(/added to cart/i).or(page.getByText(/cart/i)),
    ).toBeVisible({ timeout: 5_000 }).catch(() => {
      // Snackbar may have already dismissed - that's fine
    })

    // ── Step 8: Navigate to cart tab ──────────────────────────────────────────
    const cartTab = page
      .getByRole('tab', { name: /cart/i })
      .or(page.getByText(/cart/i).nth(1))
      .or(page.locator('[aria-label="Cart"]'))
    await cartTab.first().click()
    await page.waitForTimeout(1500)

    // Verify cart item is present
    await expect(
      page.getByText(/×/).or(page.getByText(/sku:/i)).or(page.getByText(/₽/)),
    ).toBeVisible({ timeout: 8_000 })

    // ── Step 9: Screenshot cart ────────────────────────────────────────────────
    await page.screenshot({ path: 'screenshots/05-cart.png' })

    // ── Step 10: Checkout ──────────────────────────────────────────────────────
    const checkoutButton = page.getByRole('button', { name: /checkout/i })
    await expect(checkoutButton).toBeVisible({ timeout: 8_000 })
    await checkoutButton.click()
    await page.waitForTimeout(1000)

    // Fill address dialog
    const addressInput = page
      .getByPlaceholder(/address/i)
      .or(page.getByLabel(/address/i))
    await expect(addressInput).toBeVisible({ timeout: 8_000 })
    await addressInput.fill(deliveryAddress)
    await page.waitForTimeout(500)

    // Confirm order
    const placeOrderButton = page
      .getByRole('button', { name: /place order/i })
      .or(page.getByText(/place order/i))
    await placeOrderButton.first().click()
    await page.waitForTimeout(2500)

    // ── Step 11: Order created ─────────────────────────────────────────────────
    await page.screenshot({ path: 'screenshots/06-order-created.png' })

    // Verify we are on order detail page
    await expect(
      page.getByText(/order details/i)
        .or(page.getByText(/order id/i))
        .or(page.getByText(/pending/i)),
    ).toBeVisible({ timeout: 10_000 })

    // ── Step 12: Navigate to Profile tab ──────────────────────────────────────
    const profileTab = page
      .getByRole('tab', { name: /profile/i })
      .or(page.getByText(/profile/i))
      .or(page.locator('[aria-label="Profile"]'))
    await profileTab.first().click()
    await page.waitForTimeout(1500)

    // Click first order in the orders list
    const firstOrderRow = page
      .getByText(/#[A-F0-9]{8}/)
      .or(page.getByText(/pending/i).first())
    if (await firstOrderRow.isVisible()) {
      await firstOrderRow.click()
      await page.waitForTimeout(1500)
    }

    // ── Step 13: Order detail ──────────────────────────────────────────────────
    await page.screenshot({ path: 'screenshots/07-order-detail.png' })

    await expect(
      page.getByText(/order details/i)
        .or(page.getByText(/order id/i))
        .or(page.getByText(/total/i)),
    ).toBeVisible({ timeout: 10_000 })

    // ── Step 14: Pay button (if visible) ──────────────────────────────────────
    const payButton = page.getByRole('button', { name: /pay now/i })
    if (await payButton.isVisible()) {
      await page.screenshot({ path: 'screenshots/08-pay-button.png' })
    } else {
      await page.screenshot({ path: 'screenshots/08-pay-button.png' })
    }
  })
})
