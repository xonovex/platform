# consent-and-privacy: Consent-gated, PII-safe tracking

## Guideline

Apply consent centrally in the tracker layer and never put raw PII in event params or user properties.

## Rationale

If features individually decide whether to track, one missed call ships identified data without consent — a legal/privacy breach. Raw PII (email, name, customer id, phone) in analytics payloads leaks into third-party pipelines, logs, and warehouses you do not control, and is irreversible once sent.

## How to Apply

1. Expose consent as a `Flow` of the user's current choices (categories: analytics, personalization, ad storage). Read it once inside the tracker implementation; do not pass consent down into feature code.
2. In the central tracker, branch on consent before dispatch. When analytics consent is withheld, set an anonymize/nullify flag so identifiers are stripped — only fully drop events if the product explicitly does not want anonymized collection.
3. Never accept a raw identifier in an event or property. Hash it with a stable one-way hash (e.g. SHA-256) at the call site or omit it. Treat any raw PII param as a defect to reject in review.

## Example

```kotlin
// Bad — raw PII in params, no central consent gate
class LoginViewModel(private val tracker: AnalyticsTracker) {
    fun onLoggedIn(user: User) {
        tracker.trackEvent(LoginEvent(email = user.email)) // leaks email
    }
}

// Good — hashed id at the call site; consent handled centrally
class LoginViewModel(private val tracker: AnalyticsTracker) {
    fun onLoggedIn(user: User) {
        tracker.trackEvent(LoginEvent(hashedCustomerId = sha256(user.id)))
    }
}

// Central tracker observes consent and anonymizes — features never gate
class DefaultAnalyticsTracker(
    private val consent: Flow<ConsentState>,
    private val scope: CoroutineScope,
    private val sink: AnalyticsSink,
) : AnalyticsTracker {

    private val current = AtomicReference(ConsentState.UNKNOWN)

    init {
        consent.onEach { current.set(it) }.launchIn(scope)
    }

    override fun trackEvent(event: AnalyticsEvent) {
        val state = current.get()
        if (!state.analyticsGranted) {
            // Anonymize rather than guess or silently send identified data.
            sink.send(event, anonymizeUser = true)
            return
        }
        sink.send(event, anonymizeUser = false)
    }
}
```

## Counter-Example

Strictly necessary, non-identified product telemetry (crash-free signals, anonymous feature counters) that your privacy/legal classification places outside analytics consent may bypass the gate — but it must still carry zero PII and should be documented as such, not assumed.

## Related

./user-properties-and-identity.md, ./testing.md
