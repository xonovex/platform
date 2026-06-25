---
name: android-analytics-guide
description: Use when adding or reviewing analytics / event tracking in an Android app — tracking screen views and user-intent events, deciding where tracking belongs (ViewModel vs UI), typed events and screen names, user properties and identity, consent-gated and privacy-safe tracking, A/B experiments, and testing tracking with a fake tracker. Triggers on an injected analytics tracker, trackScreen / trackEvent / setUserProperties, screen-view or button-click events, user properties, cookie consent or PII in events, or asserting tracked events in tests — even when the user doesn't say "analytics".
---

# Android analytics — quick reference

SDK-agnostic patterns for event tracking in an Android app: depend on a typed tracker abstraction, own tracking in the presentation layer, and keep it typed, consent-aware, and tested. A concrete tracking-plan codegen tool and destination are layered on top in their own skill.

## Essentials

- **Track behind an injected interface** - depend on an `AnalyticsTracker`, never the SDK directly, see [references/architecture.md](references/architecture.md)
- **Track in the ViewModel, not the UI** - screen on entry, events on user intent; the UI raises callbacks, see [references/where-to-track.md](references/where-to-track.md)
- **Typed events and screen names** - no string literals; a `ScreenName` and typed events, see [references/events-and-screens.md](references/events-and-screens.md)
- **User properties and identity via the tracker** - typed `AnalyticsUserProperty`, login / session state, see [references/user-properties-and-identity.md](references/user-properties-and-identity.md)
- **Gate on consent, keep it PII-safe** - anonymize without consent, hash identifiers, see [references/consent-and-privacy.md](references/consent-and-privacy.md)
- **Assert tracking with a fake tracker** - inject a fake, assert events / screens, see [references/testing.md](references/testing.md)

## Gotchas

- Tracking from a composable (or a `LaunchedEffect` without a stable key) re-fires on recomposition / config change — own it in the ViewModel and key it to state.
- A screen view is just an event; re-tracking it every recomposition inflates counts — track once per actual screen entry.
- `AnalyticsTracker` belongs in an `api` module; the SDK lives only in `impl`. Features depend on `api`, so tests use a fake and the SDK never leaks into feature modules.
- Raw PII (email, name, customer id) in event parameters is a privacy defect — hash or omit it, and check consent before sending.

## Example — VM-owned tracking + a fake-tracker test

```kotlin
class TicketsViewModel(
    private val analytics: AnalyticsTracker,
) : ViewModel() {
    init { analytics.trackScreen(ScreenName.TICKETS_OVERVIEW) } // once, on entry

    fun onBookClicked() {
        analytics.trackEvent(ButtonClickEvent(uniqueName = "tickets_book"))
    }
}

// Test — inject the fake, assert the event; no real SDK
@Test
fun book_click_is_tracked() = runTest {
    val analytics = FakeAnalyticsTracker()
    TicketsViewModel(analytics).onBookClicked()
    analytics.assertEvent(ButtonClickEvent(uniqueName = "tickets_book"))
}
```

## Progressive Disclosure

Each reference is a trigger — read it only when the user's intent matches; do not preload everything.

- Read [references/architecture.md](references/architecture.md) - Load when introducing or reviewing the tracker abstraction, the api/impl module split, or dependency injection of analytics.
- Read [references/where-to-track.md](references/where-to-track.md) - Load when deciding where a tracking call belongs, or fixing duplicate / recomposition-driven events.
- Read [references/events-and-screens.md](references/events-and-screens.md) - Load when defining or naming events and screen views, button/toggle clicks, or A/B experiments on a screen.
- Read [references/user-properties-and-identity.md](references/user-properties-and-identity.md) - Load when setting user properties, login status, or user / session identity.
- Read [references/consent-and-privacy.md](references/consent-and-privacy.md) - Load when gating tracking on consent, anonymizing, or removing PII from events.
- Read [references/testing.md](references/testing.md) - Load when testing that something is (or is not) tracked, using a fake / test tracker.
