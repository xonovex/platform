# state-and-announcements: State, live regions, announcements & custom actions

## Contents

- Toggle / selection state via `stateDescription` + the right `Role`
- Live regions for non-transient updates (4.1.3)
- Announce a transient in-flight state on the control
- Imperative `announceForAccessibility` for transient events
- Custom actions and named primary verbs (`onClickLabel`)
- Disabled controls
- Validate on blur, not on first render
- Counter-Examples (double-announcing, generic prompts, non-localized state)

**Guideline:** Expose every dynamic state change to accessibility services with a localized `stateDescription`, the correct interaction `Role`, a live region only for content that persists on screen, an imperative announcement only for transient events, and named verbs for every actionable element.

**Rationale:** TalkBack reads a node from its role + state + name. Toggling, selecting, expanding, loading, or erroring without surfacing the new state leaves a blind user with no feedback that anything happened. This serves 4.1.2 Name, Role, Value (state must be programmatically determinable and announced when it changes) and 4.1.3 Status Messages (AA-only — content changes that are not focus changes must be announced without moving focus). Named verbs serve 2.1.1 Keyboard / general operability — a switch-access or screen-reader user must know what activating an element does before doing it.

## Toggle / selection state

**How to Apply:**

1. Prefer the foundation modifiers — `Modifier.toggleable`, `Modifier.selectable`, `Modifier.triStateToggleable` — they set role, click handling, and `toggleableState`/`selected` semantics in one place. Reach for raw `semantics { }` only when you cannot use them.
2. Pass the matching `Role`: `Role.Switch` for on/off, `Role.Checkbox` for independent multi-select (or tri-state), `Role.RadioButton` for single-select. The role drives the spoken state phrasing ("on" vs "ticked" vs "selected").
3. For custom controls that the foundation modifiers cannot cover, set `stateDescription` yourself and **always pull it from `stringResource`** — never a hardcoded English literal. A live region is appropriate when the state text changes in place (see below).

```kotlin
// Bad - state is invisible to TalkBack; role missing; English baked in
Row(
    modifier = Modifier.clickable { onCheckedChange(!checked) }
) {
    Text(text = label)
    Switch(checked = checked, onCheckedChange = null) // decorative, state lost
}

// Good - one modifier carries value, role, click, and state
Row(
    modifier = Modifier.toggleable(
        value = checked,
        role = Role.Switch,
        onValueChange = onCheckedChange,
    )
) {
    Text(text = label)
    Switch(checked = checked, onCheckedChange = null) // visual only; Row owns semantics
}

// Good - tri-state checkbox ("parent" of a group)
Modifier.triStateToggleable(
    state = parentState, // ToggleableState.On / Off / Indeterminate
    role = Role.Checkbox,
    onClick = onParentClick,
)

// Good - genuinely custom control: localized stateDescription, no hardcoded string
val stateText = if (expanded) {
    stringResource(R.string.expanded)
} else {
    stringResource(R.string.collapsed)
}
Modifier.semantics { stateDescription = stateText }
```

## Live regions (4.1.3 Status Messages)

Use a live region when there is a **persistent node on screen** whose text changes — a loading row, an inline validation message, a value that updates in place.

**How to Apply:**

1. `liveRegion = LiveRegionMode.Polite` for non-urgent updates (loading started/finished, a count changed). Polite waits for TalkBack to finish the current utterance.
2. `liveRegion = LiveRegionMode.Assertive` for errors and limits the user must hear now — it interrupts. Use sparingly; reserve for failures and hard boundaries.
3. Put the live region on the node whose **text content** changes, not on a sibling. Compose announces the diff automatically when that node's text/`stateDescription` updates while it is composed.

```kotlin
// Good - inline error interrupts; loading status is polite
Text(
    text = errorText,
    modifier = Modifier.semantics { liveRegion = LiveRegionMode.Assertive },
)

Text(
    text = if (loading) stringResource(R.string.loading) else resultLabel,
    modifier = Modifier.semantics { liveRegion = LiveRegionMode.Polite },
)
```

**Counter-Example:** A value that the _user_ is actively driving with a stepper or slider is often better announced via `stateDescription` on the value node (with `liveRegion` `Assertive` only at limits) than a chatty Polite region on every keypress.

## Announce a transient in-flight state on the control

For a control that enters a loading/submitting state in place, attach a `liveRegion` + a state-specific `contentDescription` **only for the duration** of that state by branching the modifier, so it announces once and then reverts to a plain button.

```kotlin
val mod = if (submitting) {
    Modifier.semantics { contentDescription = submittingLabel; liveRegion = LiveRegionMode.Assertive }
} else {
    Modifier // idle: keeps its normal onClick + Role.Button + label
}
```

Conversely, gate a value's `liveRegion` so it announces only the **settled** value, not the loading placeholder it passes through:

```kotlin
Modifier.semantics(mergeDescendants = true) {
    if (state != Loading) liveRegion = LiveRegionMode.Polite
}
```

## Imperative announcements for transient events

When there is **no persistent node** to mark — a page changed in a pager, a bottom sheet opened, a snackbar-like transient event — announce it imperatively from an effect keyed on the event.

**How to Apply:**

1. Capture the view once: `val view = LocalView.current`.
2. Fire inside a `LaunchedEffect` keyed on the changing value so it runs exactly once per change, not on every recomposition.
3. Build the spoken string from `stringResource` / resource formatting — never concatenate untranslated fragments.

```kotlin
// Bad - announces on every recomposition; English literal
val view = LocalView.current
view.announceForAccessibility("Page " + page) // fires repeatedly, untranslated

// Good - one announcement per actual change, localized
val view = LocalView.current
LaunchedEffect(pagerState.currentPage) {
    view.announceForAccessibility(
        context.getString(R.string.page_x_of_y, pagerState.currentPage + 1, pageCount)
    )
}
```

Prefer a live region whenever a persistent node exists; imperative announcements bypass the accessibility tree and cannot be re-read by the user, so they are a last resort for genuinely transient events.

## Custom actions and the primary verb

`Role.Button` + `contentDescription` alone makes TalkBack say only "double tap to activate" — generic and useless on a row that does something specific. Name the verb.

**How to Apply:**

1. Keep `contentDescription` a stable **noun** (the control's name) and put the state-dependent **verb** in `onClickLabel` (the word TalkBack appends after "double tap to"). For a toggle, hold the noun fixed and switch the verb by state — never bake the verb into `contentDescription` ("add to saved, button, double tap to add to saved").
2. Name the **primary** action with `onClickLabel` on `clickable` / `toggleable` (e.g. "show item details"). TalkBack then says "double tap to show item details".
3. Expose **secondary** and **gesture-only** actions — swipe/drag/long-press-revealed Edit or Delete are otherwise unreachable by TalkBack — as `CustomAccessibilityAction(label, action)` in `customActions`; build them per item so a row only advertises the actions that apply. The `action` lambda must return `true` when handled.
4. Label every action with a localized verb phrase.

```kotlin
// Bad - generic "double tap to activate"; secondary action unreachable
Modifier
    .clickable(role = Role.Button, onClick = onOpen)
    .semantics { contentDescription = title }

// Good - named primary verb + discoverable secondary action
Modifier
    .clickable(
        role = Role.Button,
        onClick = onOpen,
        onClickLabel = stringResource(R.string.show_item_details),
    )
    .semantics {
        customActions = listOf(
            CustomAccessibilityAction(
                label = stringResource(R.string.remove_from_favorites),
                action = { onRemove(); true },
            )
        )
    }
```

## Disabled controls

Expose a disabled control instead of hiding it, so the user learns it exists and (ideally) why.

```kotlin
// Good - control is announced as disabled, with a reason
Modifier.semantics {
    disabled()
    stateDescription = stringResource(R.string.unavailable_select_a_date_first)
}
```

## Standardize one join helper

Spoken strings assembled from parts (label + state + hint) must use **one** join helper across the app, not ad-hoc `"$a, $b"` here and `"$a - $b"` there. Inconsistent separators make TalkBack pacing erratic and unreviewable. Centralize the join (and its localized separator) in a single function. A concrete shape: a sealed localized-string type with a `Multi(parts, separator)` case and one `resolve()` that does `parts.joinToString(separator)`, deferring resolution to render time. Use `listOfNotNull(...).joinToString(separator)` so optional sub-sections leave no empty separators; a `", "` separator reads as a spoken pause.

## Validate on blur, not on first render

Validate a field when it loses focus (`Modifier.onFocusChanged`), and guard the first callback with a remembered boolean — the initial composition fires `onFocusChanged`, and flagging an untouched field makes a screen-reader user hear an error before doing anything (3.3.1).

```kotlin
var touched by remember { mutableStateOf(false) }
Modifier.onFocusChanged { if (touched && !it.isFocused) validate() else touched = true }
```

## Counter-Examples (anti-patterns to reject in review)

1. **Double-announcing.** The same error in an `Assertive` `liveRegion` _and_ a `contentDescription` is read twice. Pick one: live region for the announcement, no duplicate description on a focusable sibling.
2. **Generic activation prompt.** `Role.Button` + `contentDescription` with no `onClickLabel` — TalkBack says "double tap to activate". Always name the primary verb.
3. **Non-localized state.** `stateDescription = "On"` / `announceForAccessibility("Saved")`. State and announcement strings must come from string resources.
4. **Inconsistent separators.** Hand-concatenated spoken text with different joiners per screen. Standardize one join helper.

**Related:** ./labelling.md (naming nodes; `contentDescription` vs `stateDescription`); ./focus-order.md (when to move focus instead of announce). A design system may already wire correct roles and live regions into its controls — prefer those over hand-rolled semantics.
