# testing: Testing & enforcing accessibility

## Contents

- Assert semantics, do not just locate by them
- Test accessibility text, not just UI structure
- Locate by the accessibility semantic (Robot pattern)
- Enable the Accessibility Test Framework in instrumented tests
- Screenshot tests at large font scale and in dark mode
- Lint accessibility rules: abortOnError + burn down the baseline
- Wire lint and instrumented tests into CI

**Guideline:** Treat accessibility as a verified output, not a hope — assert semantics in Compose tests, unit-test the strings that build descriptions, add automated ATF checks where you can (real apps often ship without them — see below), exercise large font scale / dark mode in screenshot tests, keep accessibility lint rules error-level, and run all of it in CI.

**Rationale:** Accessibility regressions are invisible in normal QA because the screen still looks fine to a sighted tester at default font scale. Without enforcement, content descriptions get dropped, touch targets shrink, contrast drifts, and headings vanish — silently. Automated checks defend most criteria at once: 1.1.1 Non-text Content (labels), 1.4.3 Contrast (Minimum), 1.4.4 Resize Text and 1.4.10 Reflow (font scale / reflow), 1.4.11 Non-text Contrast, 2.5.8 Target Size (Minimum), and 4.1.2 Name, Role, Value (role + state). A test that merely finds a node by its label proves nothing about whether that label is correct or whether a screen reader can act on it.

**How to Apply:**

1. In Compose UI tests, assert the semantics — `assertContentDescriptionEquals`, role via a `SemanticsMatcher`, `stateDescription`, heading presence, and touch size — not just that a node exists.
2. Add the Accessibility Test Framework — `enableAccessibilityChecks()` (Compose) / `AccessibilityChecks.enable()` (Espresso) — to fail tests on low contrast, small targets, and missing labels. It is a cheap automatic net for whole classes of issues; large apps commonly omit it, which means those classes go uncaught unless asserted by hand.
3. Add screenshot/visual-regression cases at `fontScale` 1.5 and 2.0 and in dark mode to catch clipping, truncation, reflow, and contrast regressions.
4. Keep `lint.abortOnError = true` and treat `ContentDescription`, `KeyboardInaccessibleWidget`, and `LabelFor` as errors; burn the baseline down rather than re-suppressing.
5. Run `lint` plus `connectedAndroidTest` (or a Gradle-managed-device task) on every PR in CI so any of the above actually blocks a merge.

---

### Assert semantics, do not just locate by them

The locator and the assertion must be different things. Finding a node _by_ its content description and then doing nothing else only proves a string exists somewhere — it never checks that the string is the right one, that the element has a role, or that it meets target size.

**Example:**

```kotlin
// Bad — content description used purely as a locator; nothing about it is verified.
composeTestRule
    .onNodeWithContentDescription("Delete")
    .performClick()

// Good — locate by stable testTag, then assert the accessibility contract.
composeTestRule
    .onNodeWithTag("deleteButton")
    .assertContentDescriptionEquals("Delete trip to Berlin")
    .assert(SemanticsMatcher.expectValue(SemanticsProperties.Role, Role.Button))
    .assertTouchHeightIsEqualTo(48.dp)
    .assertTouchWidthIsEqualTo(48.dp)
```

For toggleable and selectable controls, assert the spoken state and the role together (4.1.2 Name, Role, Value):

```kotlin
composeTestRule
    .onNodeWithTag("wifiToggle")
    .assert(SemanticsMatcher.expectValue(SemanticsProperties.Role, Role.Switch))
    .assertIsOn()
    .assert(SemanticsMatcher.keyIsDefined(SemanticsProperties.StateDescription))
```

For a section header, assert the heading flag is set so rotor/heading navigation works (2.4.6 Headings and Labels, 1.3.1 Info and Relationships):

```kotlin
composeTestRule
    .onNodeWithText("Payment")
    .assert(SemanticsMatcher.keyIsDefined(SemanticsProperties.Heading))
```

Use `assertHeightIsAtLeast(48.dp)` when you only care about the rendered bounds; prefer `assertTouchHeightIsEqualTo` / `assertTouchWidthIsEqualTo` when the component extends its touch area beyond its visual bounds (2.5.8 Target Size (Minimum)).

---

### Test accessibility text, not just UI structure

Asserting a description on a _rendered_ node (above) and unit-testing the _logic that builds_ it are complementary. When the description comes from a pure function or mapper (see labelling.md, "carry the spoken label as data"), test it directly: assert the **complete localized sentence** under a pinned locale, across every branch (missing field, changed/unknown value, cancelled, multi-part join, trailing punctuation). A composable wrapper runs under Robolectric pinned to a language; a non-composable builder takes a fake/mock string-resolver and runs as a plain JVM test.

```kotlin
@Config(qualifiers = "nl") // pin locale; assert the exact translated sentence
@Test fun cancelledLeg() =
    assertEquals("Geannuleerd", describeLeg(context, leg.copy(cancelled = true)))
```

Keep a thin Compose test that the value actually lands on the node; this test covers the wording across locales without rendering.

---

### Locate by the accessibility semantic (Robot pattern)

Fold the accessibility property into the **locator**, not just the assertion: find a heading via `isHeading() and hasText(localizedTitle)`, an icon control via `hasClickAction() and hasContentDescription(localizedLabel)`, resolving strings from resources. A regression — a lost `heading()`, a missing `contentDescription` — then makes the test fail to even _locate_ the element, turning silent a11y loss into a hard failure. Related notes: locate a toggle by its localized `contentDescription` then `assertIsOn()` / `assertIsOff()` (this also guards the spoken label); assert terse glyph and full phrase independently (`assertTextEquals("+2")` vs `assertContentDescriptionEquals("2 minutes delay")`); pass `useUnmergedTree = true` to finders to reach children hidden by a merged composite — merging affects screen readers _and_ the test tree.

---

### Enable the Accessibility Test Framework in instrumented tests

The ATF runs the same engine behind Accessibility Scanner over the live view tree during a test and fails on a catalogue of issues (contrast, target size, missing labels, duplicate descriptions). It is the cheapest broad net you have. Enable it once per test class. Reality check: many large apps ship with no ATF wiring at all, relying on Robot semantics assertions plus accessibility-stress screenshots — which catch labels/headings/state but leave contrast and target-size regressions uncaught. That gap is the reason to add ATF, not a reason to skip it.

**Example:**

```kotlin
// Compose: opt every interaction in this rule into ATF validation.
@get:Rule
val composeTestRule = createAndroidComposeRule<ComponentActivity>().apply {
    enableAccessibilityChecks()
}

// Espresso / View interop: enable globally in setup.
@Before
fun enableA11yChecks() {
    AccessibilityChecks.enable().setRunChecksFromRootView(true)
}
```

When a legacy finding is genuinely a false positive (for example a decorative overlap), suppress that _one_ check by matcher — never disable the framework wholesale:

```kotlin
AccessibilityChecks.enable()
    .setSuppressingResultMatcher(
        allOf(
            matchesCheckNames(`is`("TouchTargetSizeCheck")),
            matchesViews(withId(R.id.legacy_inline_icon))
        )
    )
```

**Counter-Example:** ATF cannot judge whether a label is _meaningful_ — `"Button"` or `"image1"` passes the missing-label check. Semantic correctness still needs the explicit `assertContentDescriptionEquals` from the section above.

---

### Screenshot tests at large font scale and in dark mode

Most reflow and contrast regressions only surface under conditions a default-config test never reaches. Parameterize visual tests over font scale and theme so a 200% user and a dark-mode user are covered before release (1.4.4 Resize Text, 1.4.10 Reflow, 1.4.3 Contrast (Minimum)).

**Example:**

```kotlin
// Bad — single golden image at default scale, light theme only.
@Test
fun ticketCard() = snapshot { TicketCard(sampleTrip) }

// Good — drive the same composable across the conditions that break it.
private val configs = listOf(
    Config(fontScale = 1.0f, dark = false),
    Config(fontScale = 1.5f, dark = false),
    Config(fontScale = 2.0f, dark = false),
    Config(fontScale = 1.0f, dark = true),
)

@Test
fun ticketCard_acrossConfigs() {
    configs.forEach { cfg ->
        snapshot(name = "ticketCard_${cfg.fontScale}_${if (cfg.dark) "dark" else "light"}") {
            CompositionLocalProvider(
                LocalDensity provides Density(density = 2.625f, fontScale = cfg.fontScale)
            ) {
                AppTheme(darkTheme = cfg.dark) { TicketCard(sampleTrip) }
            }
        }
    }
}
```

Review the 2.0 and dark diffs deliberately: clipped text, ellipsized labels, and rows that stop scrolling are reflow failures, not cosmetic noise. Regenerate goldens only when the change is intended. `2.0` is the 200% WCAG/platform cap, so it is the ladder's top rung. Wrap the matrix in one reusable helper (e.g. `snapshotsAccessibility { }`) backed by a named font-scale ladder (1.0/1.25/1.5/2.0) plus landscape and dark, so a11y coverage is one call per screen instead of scattered `fontScale` tweaks.

---

### Lint accessibility rules: abortOnError + burn down the baseline

Android lint statically catches a meaningful slice — `ContentDescription` (missing description on `ImageView`/icon), `KeyboardInaccessibleWidget` (clickable with no focusability), and `LabelFor` (input with no associated label). These map to 1.1.1, 2.1.1 Keyboard, and 3.3.2 Labels or Instructions. They only protect you if they fail the build.

**Example:**

```kotlin
// Bad — every finding is swallowed and a baseline hides the existing debt forever.
android {
    lint {
        abortOnError = false
        baseline = file("lint-baseline.xml") // grows silently on each new violation
    }
}

// Good — accessibility findings are errors and block the build.
android {
    lint {
        abortOnError = true
        warningsAsErrors = true
        // Promote the a11y checks explicitly so config drift can't downgrade them.
        error += listOf("ContentDescription", "KeyboardInaccessibleWidget", "LabelFor")
        checkDependencies = true
    }
}
```

If a baseline already exists, treat it as a debt ledger to shrink, not a permanent mute: fix entries and delete them, and never let a re-run regenerate the baseline to absorb new violations. A baseline that keeps growing is indistinguishable from `abortOnError = false`. Keep genuine non-issue suppressions in a separate `lintConfig` XML — never list accessibility rules (`ContentDescription`, `SpUsage`) there — and make "the baseline must shrink" an explicit written policy.

**Counter-Example:** A short-lived baseline pinned to a known count, tracked in a follow-up issue and reduced every sprint, is acceptable. The anti-pattern is an unbounded baseline plus `abortOnError = false`, where new findings re-baseline silently.

---

### Wire lint and instrumented tests into CI

Local discipline does not survive a busy team. The semantics assertions, ATF checks, screenshot tests, and lint rules above only prevent regressions if a merge cannot happen while any of them is red.

**How to Apply:**

1. Run `./gradlew lint` (error-level a11y rules) on every PR.
2. Run instrumented UI tests via `connectedAndroidTest` or a Gradle-managed-device task so ATF and semantics assertions execute on a real/virtual device in CI.
3. Run the screenshot suite and publish the diff report as a build artifact for review.
4. Make all three required status checks; do not allow override-on-red as routine.

**Counter-Example:** A pipeline that compiles and runs only JVM unit tests exercises none of the accessibility surface — semantics, contrast, target size, and reflow all live in the instrumented and screenshot layers. "Green CI" without those layers is a false signal.

**Related:** the behaviors asserted here are defined in ./labelling.md, ./focus-order.md, ./state-and-announcements.md, ./contrast-and-color.md, and ./text-and-targets.md. Follow your design system's component test conventions (screenshot goldens, test-tag injection) where it provides them.
