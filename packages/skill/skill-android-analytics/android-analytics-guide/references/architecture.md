# architecture: Track behind an injected tracker abstraction

**Guideline:** Feature code depends on a hand-written `AnalyticsTracker` interface injected via DI; the analytics SDK is referenced only by one implementation module, never by ViewModels, composables, or other feature code.

**Rationale:** A directly-imported SDK couples every feature to a vendor: swapping or adding a destination becomes a project-wide edit, tests cannot assert what was tracked without a real network/SDK, and SDK initialization order leaks into UI code. An interface boundary localizes the SDK to one place and makes tracking a substitutable collaborator.

**How to Apply:**

1. Define `AnalyticsTracker` in an `api` module with no SDK dependency. Keep the surface small and typed: `trackScreen`, `trackEvent`, `setUserProperties`, plus optional helpers like `trackScreenWithExperiments` and a convenience `trackButtonClickEvent`.
2. Put the SDK-backed implementation in a separate `impl` module that depends on `api` and the SDK. Feature modules depend only on `api`.
3. Inject the tracker through the constructor (or DI graph). Never new-up the implementation or touch the SDK from a feature class.
4. The implementation may fan out to several destinations behind the one interface; callers stay unaware.
5. In tests, substitute a fake implementation of the interface (see ./testing.md).

**Example:**

```kotlin
// Bad — SDK imported directly in a ViewModel; vendor leaks into the feature,
// untestable without the real SDK.
import com.vendor.analytics.VendorSdk

class CheckoutViewModel : ViewModel() {
    fun onPay() {
        VendorSdk.shared.log("pay_tapped", mapOf("screen" to "checkout"))
    }
}

// Good — api module: hand-written, SDK-free boundary.
interface AnalyticsTracker {
    fun trackScreen(screenName: ScreenName)
    fun trackScreenWithExperiments(screenName: ScreenName, experiments: List<Experiment>)
    fun trackEvent(event: AnalyticsEvent)
    fun setUserProperties(properties: Map<AnalyticsUserProperty, String>)

    fun trackButtonClickEvent(uniqueName: String) =
        trackEvent(ButtonClickEvent(uniqueName))
}

// Good — feature depends only on the interface, injected.
class CheckoutViewModel(
    private val analytics: AnalyticsTracker,
) : ViewModel() {
    fun onPay() = analytics.trackButtonClickEvent(uniqueName = "checkout_pay")
}

// Good — impl module: the only place the SDK appears; can fan out.
internal class SdkAnalyticsTracker(
    private val sdk: VendorSdk,
    private val secondary: SecondaryDestination,
) : AnalyticsTracker {
    override fun trackEvent(event: AnalyticsEvent) {
        sdk.log(event.name, event.params)
        secondary.send(event)
    }
    // ...
}
```

**Counter-Example:** A thin, throwaway prototype with no tests and one destination may call the SDK inline. As soon as a second destination, a unit test, or a shared feature module appears, introduce the interface.

**Related:** ./where-to-track.md, ./testing.md. The concrete SDK-backed implementation conventions (a specific analytics SDK or codegen tool) are outside this platform-level skill.
