# Sources

## Flagship Android app codebase (applied patterns)

- **Last reviewed:** 2026-06-16
- **Used for:** the SDK-agnostic Android analytics patterns this skill generalizes
- **Aspects extracted:**
  - `core/analytics/api` + `core/analytics/impl` module split; the injected `AnalyticsTracker` interface (`trackScreen`, `trackEvent`, `setUserProperties`, `trackScreenWithExperiments`) → `architecture.md`
  - Tracking performed in ViewModels (screen on load, events on user intent) → `where-to-track.md`
  - Typed events + `ScreenName` constants, button/toggle click conventions, A/B `Experiment` → `events-and-screens.md`
  - `AnalyticsUserProperty` enum + `setUserProperties`; user/session identity → `user-properties-and-identity.md`
  - Consent-gated tracking and user anonymization (cookie-consent → consent context, `nullifiedUser`), PII hashing → `consent-and-privacy.md`
  - `FakeAnalyticsTracker` test fixture (Turbine), event/screen assertions → `testing.md`

## Android analytics general knowledge

- **Used for:** platform-level patterns (DI of a tracker interface, ViewModel-owned tracking, typed events, consent) that hold regardless of the analytics SDK. SDK-specific details (e.g. a particular analytics codegen tool) are out of scope for this platform-level skill.

## Refresh Workflow

1. Re-scan `core/analytics/api` for changes to the `AnalyticsTracker` contract and `FakeAnalyticsTracker`
2. Diff against the prior pull
3. Update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** above
