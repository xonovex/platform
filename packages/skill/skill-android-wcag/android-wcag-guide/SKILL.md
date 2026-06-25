---
name: android-wcag-guide
description: Use when making Android Jetpack Compose UI accessible or meeting WCAG 2.2 AA — labelling for TalkBack, focus / reading order, headings, state and live-region announcements, color contrast, text scaling, touch-target size, and accessibility testing. Triggers on contentDescription, semantics / clearAndSetSemantics / mergeDescendants, traversalIndex / isTraversalGroup / heading(), stateDescription / liveRegion / announceForAccessibility, Role, CustomAccessibilityAction, TalkBack, screen reader, contrast, font scaling, tap target, or @Composable accessibility tests — even when the user doesn't say "WCAG" or "accessibility".
---

# Android (Compose) WCAG 2.2 AA — quick reference

Platform-level accessibility for Jetpack Compose: the `androidx.compose.ui.semantics` APIs and the AA criteria they satisfy. Framework-agnostic — a design system layered on top may bake some of this in, but the techniques here are the platform baseline.

When this skill fires:

1. Decide what the screen reader should perceive, in what order, and how state changes are announced — then reach for the matching semantics API below.
2. Source every user-facing label from `stringResource`, never a literal.
3. Load the `references/*.md` file matching the question, not everything upfront.

## Essentials

- **Label non-text content** - meaningful graphics get a localized `contentDescription`; decorative ones get `null`, see [references/labelling.md](references/labelling.md)
- **Control reading order and headings** - `traversalIndex`, `isTraversalGroup`, `heading()`, collection semantics, see [references/focus-order.md](references/focus-order.md)
- **Announce state and status** - `stateDescription`, `liveRegion`, `announceForAccessibility`, `CustomAccessibilityAction`, `onClickLabel`, see [references/state-and-announcements.md](references/state-and-announcements.md)
- **Meet contrast and never rely on color alone** - 4.5:1 text / 3:1 UI, pair color with icon or label, see [references/contrast-and-color.md](references/contrast-and-color.md)
- **Scale text and size touch targets** - `sp` text that honors `fontScale`, ≥48dp targets via `minimumInteractiveComponentSize`, see [references/text-and-targets.md](references/text-and-targets.md)
- **Assert and enforce accessibility** - semantics + a11y-text assertions, ATF (where present), lint + CI, see [references/testing.md](references/testing.md)

## Gotchas

- `contentDescription` is a localized user-facing string — source it from `stringResource`, never a literal. `testTag` is a separate dev-facing concern; never reuse one as the other.
- An empty `semantics(mergeDescendants = true) { }` with no `contentDescription` collapses children into raw concatenated order and drops the intended label — always set the consolidated description.
- `contentDescription` has an **accumulating** merge policy: setting it on a `mergeDescendants` parent does not replace children, it appends to whatever the children already expose (badge counts, labelled icons). When children carry their own descriptions, use `clearAndSetSemantics { }` to get one clean label instead of a doubled-up one.
- `graphicsLayer { alpha = 0f }` (or any fade) hides a node visually but leaves it in the accessibility tree, so TalkBack still reads it — common cause of a collapsing header announcing its title twice. Gate the hidden copy with `Modifier.semantics { hideFromAccessibility() }` (older Compose: `invisibleToUser()`).
- Putting the same text in an Assertive `liveRegion` **and** a `contentDescription` makes TalkBack read it twice — pick one.
- `clickable { }` on a `Row`/`Box`/`Card` exposes **no role** by default, so TalkBack does not announce it as actionable — pass `clickable(role = Role.Button) { }` (or set `role` in `semantics`). Without `onClickLabel` it also announces only a generic "double tap to activate" — name the action verb.
- `heading()` is not implied by text size or font weight — set it explicitly on screen titles and section headers, or heading navigation finds nothing.
- Don't put controls (tabs, buttons, chips) in the heading rotor — `heading()` is for content sections only; tabs already carry `Role.Tab`. Conversely, in a list of cards, mark **each item** a heading so users skim with the heading gesture instead of swiping every card and its button.

## Example — one focusable, labelled, actionable node for a composite tile

```kotlin
val recap = stringResource(R.string.booking_recap, route, date)   // localized, prebuilt
val newLabel = stringResource(R.string.badge_new)

Row(
    modifier = Modifier
        .clickable(onClickLabel = stringResource(R.string.action_open_booking)) { onOpen() }
        .clearAndSetSemantics {                       // replace children: announce as ONE node
            role = Role.Button
            contentDescription = recap
            if (isNew) stateDescription = newLabel
        },
) {
    Icon(painterResource(R.drawable.ic_train), contentDescription = null)  // decorative
    Column {
        Text(text = route)
        Text(text = date, style = MaterialTheme.typography.bodySmall)
    }
}
```

## Progressive Disclosure

Each reference is a trigger — read it only when the user's intent matches; do not preload everything.

- Read [references/labelling.md](references/labelling.md) - Load when adding or fixing `contentDescription`, deciding decorative vs meaningful, or consolidating a composite control with `clearAndSetSemantics` / `mergeDescendants` (WCAG 1.1.1, 4.1.2).
- Read [references/focus-order.md](references/focus-order.md) - Load when the screen reader reads elements in the wrong order, when distinguishing accessibility traversal order from keyboard focus order vs composition order, when adding headings / traversal grouping / list-collection semantics, or for programmatic focus, bring-into-view, focus-first-error, and IME traversal (WCAG 1.3.1, 1.3.2, 2.4.3, 2.4.6, 2.4.7, 3.3.1, 2.1.1).
- Read [references/state-and-announcements.md](references/state-and-announcements.md) - Load when a toggle/selection state isn't announced, when announcing loading / errors / transient events, or when adding custom or named actions (WCAG 4.1.2, 4.1.3, 2.1.1).
- Read [references/contrast-and-color.md](references/contrast-and-color.md) - Load when checking contrast ratios or when meaning is carried by color alone (WCAG 1.4.1, 1.4.3, 1.4.11).
- Read [references/text-and-targets.md](references/text-and-targets.md) - Load when text doesn't scale with system font size, content clips at large scale, or tap targets are too small (WCAG 1.4.4, 1.4.10, 2.5.5, 2.5.8).
- Read [references/testing.md](references/testing.md) - Load when writing accessibility tests, enabling automated a11y checks, or wiring lint / CI enforcement.
