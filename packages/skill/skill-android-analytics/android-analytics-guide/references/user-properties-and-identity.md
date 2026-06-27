# user-properties-and-identity: User properties & identity

## Guideline

Set slowly-changing user dimensions once via a typed `setUserProperties` call keyed by an enum, and keep identity (user id, session id) inside the tracker layer — never re-derive them into per-event params.

## Rationale

Stuffing `isLoggedIn`, `appLanguage`, or a user id into every event's params means each call site must re-derive state, the values drift between events, and renaming a dimension forces edits across dozens of tracking sites. User properties model dimensions that change rarely and apply to _all_ subsequent events; pushing them per-event inflates payloads, splits the source of truth, and risks leaking identifiers into events that should not carry them. A typed key set also stops two call sites from disagreeing on the spelling of a property key.

## How to Apply

1. Define a closed key set (an `enum class AnalyticsUserProperty`) and expose `fun setUserProperties(properties: Map<AnalyticsUserProperty, String>)` on the tracker — no raw `String` keys.
2. Set each property at its natural lifecycle moment: environment and language at app start; login status, account type, and entitlement flags on login/logout; update them when they change rather than re-reading per event.
3. Keep user id and session id behind tracker methods (`identify`/`resetIdentity`); never pass them as event params.
4. Decide property vs. param by cardinality of change: a value that changes per action is a param; a value that persists across many events is a user property.

## Example

```kotlin
enum class AnalyticsUserProperty {
    ENVIRONMENT, APP_LANGUAGE, LOGGED_IN_CONSUMER, LOGGED_IN_BUSINESS, HAS_PAS
}

interface AnalyticsTracker {
    fun trackEvent(event: AnalyticsEvent)
    fun setUserProperties(properties: Map<AnalyticsUserProperty, String>)
    fun identify(userId: String)
    fun resetIdentity()
}

// Bad — login status (and the user id) ride along on every event
data class ButtonClickEvent(
    val label: String,
    val isLoggedIn: Boolean,   // slowly-changing dimension smuggled into a per-action event
    val userId: String,        // identity duplicated into params
) : AnalyticsEvent

tracker.trackEvent(ButtonClickEvent(label = "buy", isLoggedIn = true, userId = user.id))

// Good — set the dimension and identity once, keep the event about the action
fun onConsumerLoggedIn(tracker: AnalyticsTracker, user: User) {
    tracker.identify(user.id)
    tracker.setUserProperties(
        mapOf(
            AnalyticsUserProperty.LOGGED_IN_CONSUMER to "true",
            AnalyticsUserProperty.HAS_PAS to user.hasPas.toString(),
        ),
    )
}

fun onLoggedOut(tracker: AnalyticsTracker) {
    tracker.setUserProperties(mapOf(AnalyticsUserProperty.LOGGED_IN_CONSUMER to "false"))
    tracker.resetIdentity()
}

// The event now carries only per-action data
data class ButtonClickEvent(val label: String) : AnalyticsEvent
tracker.trackEvent(ButtonClickEvent(label = "buy"))
```

## Counter-Example

A dimension that genuinely differs per action is not a user property — record it as an event param. For example, the _screen_ a button lives on belongs on the event, not on the user, because it changes with every navigation.

## Related

./consent-and-privacy.md, ./events-and-screens.md
