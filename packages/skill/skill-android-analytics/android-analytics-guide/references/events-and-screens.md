# events-and-screens: Typed events, screen names & experiments

**Guideline:** Never pass string-literal event or screen names to the tracker; model screens as a typed `ScreenName` and events as a sealed event hierarchy so every call site is compiler-checked and IDE-discoverable.

**Rationale:** String keys silently drift — a typo (`"buton_click"`), an inconsistent param name, or a renamed screen produces broken analytics that compile and ship fine, then surface as gaps in dashboards weeks later. A typed model makes the set of valid events finite, refactor-safe (rename propagates), and discoverable via autocomplete, so the call site cannot invent an undefined event.

**How to Apply:**

1. Define an enum (or sealed `object` constants) `ScreenName` for every trackable screen; fire a screen-view on entry, not on every recomposition or back-navigation.
2. Define a sealed `AnalyticsEvent` hierarchy. One concrete type per user action (`ButtonClickEvent`, `ToggleChangeEvent`), never a free-form `CustomEvent(name, map)`.
3. Give click-type events a stable `uniqueName` (snake_case, prefixed by area, e.g. `tickets_book`) plus optional structured params as typed fields — not a loose `Map`.
4. For A/B tests, do not branch the tracking call by hand. Pass the active experiments alongside the screen view via `trackScreenWithExperiments`, modelling each as `Experiment(key, value)` so several experiments ride one screen view.

**Example:**

```kotlin
// Bad — string keys, loose map, no compiler check
tracker.trackEvent("button_click", mapOf("name" to "book"))
tracker.trackScreen("tickets_overview")

// Good — typed, finite, discoverable
enum class ScreenName(val key: String) {
    TICKETS_OVERVIEW("tickets_overview"),
    TICKET_DETAIL("ticket_detail"),
}

sealed interface AnalyticsEvent {
    data class ButtonClickEvent(val uniqueName: String) : AnalyticsEvent
    data class ToggleChangeEvent(val uniqueName: String, val enabled: Boolean) : AnalyticsEvent
}

tracker.trackEvent(AnalyticsEvent.ButtonClickEvent(uniqueName = "tickets_book"))
tracker.trackScreen(ScreenName.TICKETS_OVERVIEW)

// A/B: one screen view carries every active experiment
data class Experiment(val key: String, val value: String)

tracker.trackScreenWithExperiments(
    screenName = ScreenName.TICKETS_OVERVIEW,
    experiments = listOf(
        Experiment("checkout_layout", "compact"),
        Experiment("price_badge", "variant_b"),
    ),
)
```

**Counter-Example:** A genuinely dynamic dimension (e.g. a server-supplied product id) belongs in a _typed param_ on a defined event (`ButtonClickEvent(uniqueName = "tickets_book", productId = id)`), not as a dynamically-built event _name_. The event type stays finite; only its data varies.

**Related:** Where this typed model comes from (code generation from a tracking plan) is an SDK/codegen concern outside this skill; this file is about using a typed model well. See ./where-to-track.md for which layer fires these calls and ./user-properties-and-identity.md for dimensions that persist across events rather than riding a single one.
