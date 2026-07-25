# Sources

## WCAG 2.2 (W3C Recommendation)

- **URL:** https://www.w3.org/TR/WCAG22/
- **Last reviewed:** 2026-06-16
- **Used for:** success-criteria definitions and AA scope across all references
- **References:** all
- **Aspects extracted:**
  - Perceivable / Operable / Understandable / Robust success criteria → criteria mapping in every reference
  - AA thresholds (1.4.3 contrast 4.5:1 / 3:1, 1.4.4 resize 200%, 1.4.10 reflow) → `contrast-and-color.md`, `text-and-targets.md`

## Jetpack Compose accessibility (Android Developers)

- **URL:** https://developer.android.com/develop/ui/compose/accessibility
- **Last reviewed:** 2026-06-16
- **Used for:** Compose semantics APIs and their correct application
- **Aspects extracted:**
  - `semantics` / `clearAndSetSemantics` / `mergeDescendants` / Role / `invisibleToUser` → `labelling.md`
  - `traversalIndex` / `isTraversalGroup` / `heading()` / `CollectionInfo` → `focus-order.md`
  - `stateDescription` / `liveRegion` / `toggleable` / `CustomAccessibilityAction` → `state-and-announcements.md`
  - font scaling / text reflow → `text-and-targets.md`
  - Compose test semantics assertions → `testing.md`

## Production Android codebase (applied patterns)

- **Provenance:** Repository-local production Android implementation reviewed for reusable accessibility patterns
- **Last reviewed:** 2026-06-16
- **Used for:** real-world Compose accessibility patterns and the recurring mistakes mined into `Counter-Example` / `Gotchas` (hardcoded contentDescription literals, missing headings on app bars, ad-hoc spoken-text joining, double-announcing errors)
- **References:** references/labelling.md, references/state-and-announcements.md

## Flagship Android app codebase (applied patterns)

- **Provenance:** Repository-local Android application reviewed for reusable Compose accessibility patterns
- **Last reviewed:** 2026-06-16
- **Used for:** the flagship app's stronger patterns folded into the references — accessibility text built in a tested data/use-case layer (pure function + injected string resolver), `Modifier.moveIntoViewOnFocus` plus programmatic-focus / focus-first-error / IME-traversal recipes, the `isHeading()` Robot-pattern test idiom, conditional `liveRegion` for in-flight state, and the real-world observation that ATF is commonly absent (Robot semantics assertions + accessibility-stress screenshots used instead)
- **References:** references/focus-order.md, references/labelling.md, references/state-and-announcements.md, references/testing.md

## Refresh Workflow

1. Re-check the WCAG 2.2 criteria list and the Compose accessibility docs for new or changed APIs
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
