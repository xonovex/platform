# where-to-track: Own tracking in the ViewModel, not the UI

## Guideline

Call the analytics tracker from the presentation layer (ViewModel/presenter) — track a screen view once on screen entry and track events from user-intent callbacks — never from inside a composable body.

## How to Apply

The ViewModel survives recomposition and config change; a composable body and `LaunchedEffect(Unit)` re-fire and double-count.

1. Track the screen view in the ViewModel's `init` (or the function that first loads the screen), so it fires once per ViewModel instance — i.e. once per genuine screen entry.
2. For events, have the composable expose an intent callback (`onCheckoutClicked`) and let the ViewModel call the tracker inside the handler. The composable never touches the tracker.
3. Re-track a screen only on a real re-entry (new ViewModel, new navigation destination), not on every state emission. Map state changes to UI; map user intents to events.
4. For Activity/Fragment screens, track the screen view once — in `init`/`onCreate`, or in `onResume` guarded by a flag so back-stack returns don't re-count.

## Example

```kotlin
// Bad — composable owns tracking; re-fires on every recomposition / config change
@Composable
fun CartScreen(state: CartState, analytics: AnalyticsTracker) {
    LaunchedEffect(Unit) {                       // re-fires on rotation -> inflated screen views
        analytics.trackScreen(ScreenName.Cart)
    }
    Button(onClick = {
        analytics.trackEvent(ButtonClickEvent("cart_checkout")) // event firing from the UI layer
    }) { /* ... */ }
}

// Good — ViewModel owns tracking; UI only raises intent
class CartViewModel(
    private val analytics: AnalyticsTracker,
) : ViewModel() {
    init {
        analytics.trackScreen(ScreenName.Cart) // once per ViewModel == once per entry
    }

    fun onCheckoutClicked() {
        analytics.trackEvent(ButtonClickEvent("cart_checkout"))
        // ...proceed to checkout
    }
}

@Composable
fun CartScreen(state: CartState, onCheckoutClicked: () -> Unit) {
    Button(onClick = onCheckoutClicked) { /* ... */ }
}
```

## Counter-Example

A truly UI-local interaction with no ViewModel and no business meaning (e.g. a self-contained reusable composable that must report its own "expanded" toggle) may track from a stably-keyed `LaunchedEffect` or a side-effect handler — but pass the tracker in and key the effect on the actual trigger, never `Unit`. Prefer hoisting the callback to a ViewModel whenever one exists.

## Related

./architecture.md, ./events-and-screens.md
