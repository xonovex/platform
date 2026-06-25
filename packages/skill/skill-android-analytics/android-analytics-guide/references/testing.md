# testing: Test tracking with a fake tracker

**Guideline:** Treat tracking as a verifiable output: inject an in-memory `FakeAnalyticsTracker` into the unit under test and assert the exact typed events it received.

**Rationale:** Tracking is a real behavioral contract — analytics dashboards, funnels, and experiments break silently when an event stops firing or fires with wrong properties. Without a test asserting the typed event, a refactor can drop tracking and no compiler or runtime error catches it. Asserting against the live SDK or network is slow, flaky, and tests the vendor, not your code.

**How to Apply:**

1. Depend on the `AnalyticsTracker` interface (not a concrete SDK). Provide a `FakeAnalyticsTracker` that records into an in-memory list and exposes `assertEvent(expected)`, `assertScreen(name)`, `assertNoEvent()`, and `clear()`.
2. Construct the fake, inject it into the ViewModel/use-case, perform the action, then assert the **exact typed event** by equality — not a string substring or a count alone.
3. Assert ordering only when it matters; back the fake with a Turbine channel to `await` and verify event sequence in suspend tests.
4. Test the negative: assert `assertNoEvent()` before user intent, and assert consent-withheld paths emit anonymized (or no) events. `clear()` between acts so one test's events don't leak into the next assertion.

**Example:**

```kotlin
class FakeAnalyticsTracker : AnalyticsTracker {
    private val recorded = mutableListOf<AnalyticsEvent>()
    private val screens = mutableListOf<ScreenName>()

    override fun trackEvent(event: AnalyticsEvent) { recorded += event }
    override fun trackScreen(name: ScreenName) { screens += name }

    fun assertEvent(expected: AnalyticsEvent) {
        assertTrue(expected in recorded, "expected $expected, got $recorded")
    }
    fun assertScreen(name: ScreenName) { assertTrue(name in screens) }
    fun assertNoEvent() { assertTrue(recorded.isEmpty(), "unexpected $recorded") }
    fun clear() { recorded.clear(); screens.clear() }
}

// Bad — no assertion that checkout tracked anything, or asserts a loose string
@Test fun checkout() {
    vm.onCheckoutClicked()
    verify(tracker).trackEvent(argThat { it.toString().contains("checkout") }) // brittle, untyped
}

// Good — exact typed event on a fake
@Test fun `checkout click tracks button event`() {
    val analytics = FakeAnalyticsTracker()
    val vm = CartViewModel(analytics)

    analytics.assertNoEvent()        // nothing before intent
    vm.onCheckoutClicked()
    analytics.assertEvent(ButtonClickEvent("cart_checkout"))
}
```

**Counter-Example:** End-to-end verification that events actually reach the vendor backend belongs in a separate, manually-run integration check against a staging pipeline — not in fast unit tests. Don't gate CI on the live SDK.

**Related:** ./architecture.md, ./where-to-track.md
