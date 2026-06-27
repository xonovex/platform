# focus-order: Reading order, grouping & headings

## Contents

- Three independent orders (accessibility vs keyboard vs composition)
- Accessibility traversal order (`traversalIndex`)
- Grouping (`isTraversalGroup`, `mergeDescendants`)
- Headings (`heading()`)
- Collections (`collectionInfo` / `collectionItemInfo`)
- Named traversal-index enum (maintainability)
- Section box: jump to a heading, then descend into it
- Programmatic focus & bringing content into view

### Guideline

Make the accessibility traversal order match the visual reading order, expose headings and collection structure to the semantics tree, and group composites into single swipe stops.

### Rationale

Screen-reader users navigate sequentially (swipe) and by structure (heading rotor, "item X of Y"). If the semantics tree does not encode order, grouping, headings, and collection counts, those navigation modes break: focus jumps around, related elements scatter into separate swipes, heading jump-navigation finds nothing, and lists give no position feedback. Serves 1.3.1 Info and Relationships, 1.3.2 Meaningful Sequence, 2.4.3 Focus Order, and 2.4.6 Headings and Labels.

## Three independent orders (accessibility vs keyboard vs composition)

Android has **three distinct orderings**, controlled by different APIs; no single property syncs them:

- **Accessibility traversal order** — what TalkBack reads. Default = visual reading order (top-to-bottom, left-to-right). Tuned with `isTraversalGroup` + `traversalIndex` (semantics).
- **Keyboard / D-pad focus order** — where Tab / arrows move focus. Default = the order composables are _called_ (a depth-first walk of the composition), **not** the layout. Tuned with `Modifier.focusProperties { next/previous, up/down/left/right }` + `FocusRequester` (see "Programmatic focus" below).
- **Composition / index order** — the order you declare children. It is the _default basis_ for keyboard 1D (Tab) focus, but **not** for accessibility traversal.

`traversalIndex` changes only the first and does nothing for keyboard focus; `focusProperties` changes only the second and does nothing for TalkBack. To give screen-reader **and** keyboard users the same custom order, set **both**.

## Accessibility traversal order (`traversalIndex`)

### How to Apply

1. Default order is the visual reading order (top-to-bottom, left-to-right) — note this differs from the keyboard default (composable-call order). Only override when the visual order differs from composition order (common with a top bar drawn after body content, overlays, or `Box` z-stacks).
2. Set `Modifier.semantics { traversalIndex = <Float> }`. Lower reads earlier; the implicit default is `0f`; negatives read before everything at `0f`.
3. Keep the spread small and intentional: e.g. body content `-10f`, top-bar title `-1f`. Do not assign a unique index to every node, only the ones whose order you must correct.

```kotlin
// Bad - title composed first, so TalkBack reads "Inbox" before the message the user just tapped
Box {
    TopAppBar(title = { Text("Inbox") })
    MessageList(modifier = Modifier.padding(top = 64.dp))
}

// Good - body reads first, then the bar
Box {
    TopAppBar(
        title = { Text("Inbox") },
        modifier = Modifier.semantics { traversalIndex = TraversalOrder.TopBar.index }, // -1f
    )
    MessageList(
        modifier = Modifier.semantics { traversalIndex = TraversalOrder.Body.index }, // -10f
    )
}
```

## Grouping (`isTraversalGroup`, `mergeDescendants`)

### How to Apply

1. `Modifier.semantics { isTraversalGroup = true }` scopes ordering: children sort by `traversalIndex` _within_ the group before the group is placed among its siblings. Use it so a reordered region stays self-contained.
2. `isTraversalGroup` does NOT merge children into one focus stop. To collapse a card (icon + title + subtitle) into a single swipe stop, use `Modifier.semantics(mergeDescendants = true) { }`. They are orthogonal and apply at **different levels** — group a section with `isTraversalGroup` at the container, merge an atomic card with `mergeDescendants` at its own node — but combining both on the **same** node is pointless: a merged node has no individually focusable children left to order.
3. Merge only when the children carry no independent action. A card with a nested button must stay unmerged so the button keeps its own focus and click action.

```kotlin
// Bad - each line is a separate swipe stop; user hears three fragments
Column {
    Text("Maria Jansen")
    Text("Project update")
    Text("2 minutes ago")
}

// Good - one stop, read as a sentence
Column(
    modifier = Modifier
        .semantics(mergeDescendants = true) {
            contentDescription = "Maria Jansen, Project update, 2 minutes ago"
        },
) {
    Text("Maria Jansen")
    Text("Project update")
    Text("2 minutes ago")
}
```

## Headings (`heading()`)

### How to Apply

1. Mark the screen title AND every section header with `Modifier.semantics { heading() }`. TalkBack's heading-navigation control jumps between these; without it the user must swipe through every element.
2. Apply it to the text/label node, not its container, so the announced "heading" text is the visible title.
3. A heading is still a normal focus stop. Do not add a redundant "heading" word to the label, TalkBack appends the role itself.
4. `heading()` is a flat boolean — there is **no** `heading(level = …)`, and TalkBack does not announce heading levels on Android. A screen's outline is simply the ordered set of `heading()` nodes (see "Section box" below); convey depth with typography and reading order, not a level.
5. **Don't mark controls as headings.** Tabs, buttons, and chips are operable controls that already expose a `Role` (e.g. `Role.Tab` with a selected state) — adding `heading()` pollutes the heading rotor, which is for jumping between _content_ sections, and is semantically wrong (fails the intent of 1.3.1 / 2.4.6). A tab row is navigated by its tab role, not by heading navigation.
6. **In a list, mark each item a heading.** For a list of cards/rows, put `heading()` on each item (alongside its consolidated label) so a screen-reader user skims the list with the heading gesture/rotor instead of swiping through every card and its inner button. This is the conventional hierarchy: a section title is a heading, and each item under it is a heading too.

```kotlin
// Bad — tab row marked as a heading; the rotor lands on a control
TabRow(/* ... */, modifier = Modifier.semantics { heading() })

// Good — tabs keep their tab role; each list card is the heading-nav target
TabRow(/* ... */)                                   // a control, not a heading
items(messages) { message ->
    Card(
        modifier = Modifier.semantics(mergeDescendants = true) {
            heading()                               // skimmable with the heading gesture
            isTraversalGroup = true
            contentDescription = message.label
        },
    ) { /* card content; interactive children stay separately focusable */ }
}
```

```kotlin
// Bad - title is announced as plain text; no heading-nav target on the screen
Text("Your messages", style = MaterialTheme.typography.titleLarge)

// Good
Text(
    text = "Your messages",
    style = MaterialTheme.typography.titleLarge,
    modifier = Modifier.semantics { heading() },
)
```

## Collections (`collectionInfo` / `collectionItemInfo`)

### How to Apply

1. On the list/grid container, set `Modifier.semantics { collectionInfo = CollectionInfo(rowCount = n, columnCount = c) }`. For an unknown/streaming count, use `-1`.
2. On each item, set `collectionItemInfo = CollectionItemInfo(rowIndex, rowSpan, columnIndex, columnSpan)`. This is what makes TalkBack announce "item X of Y".
3. `LazyColumn` / `LazyRow` populate this automatically. Add it by hand only when you build a collection out of a plain `Column`/`Row`/`Box` (or a flow layout) that the framework cannot infer.

```kotlin
// Bad - hand-rolled list in a Column; TalkBack gives no "X of Y" position
Column {
    people.forEach { Text(it.name) }
}

// Good
Column(
    modifier = Modifier.semantics {
        collectionInfo = CollectionInfo(rowCount = people.size, columnCount = 1)
    },
) {
    people.forEachIndexed { index, person ->
        Text(
            text = person.name,
            modifier = Modifier.semantics {
                collectionItemInfo = CollectionItemInfo(
                    rowIndex = index, rowSpan = 1, columnIndex = 0, columnSpan = 1,
                )
            },
        )
    }
}
```

## Named traversal-index enum (maintainability)

### Guideline

Replace scattered magic `Float` traversal indices with one named enum per screen so the ordering is self-documenting and edits stay consistent.

### How to Apply

1. Define an enum (or sealed value list) whose entries name the regions in reading order and expose their index.
2. Reference the enum everywhere instead of literals, so adding a region in the middle is a single, reviewable change.
3. Keep one enum per screen/feature, not a global one, ordering is local to a layout.

```kotlin
// Bad - unexplained literals drift apart over edits
Modifier.semantics { traversalIndex = -10f } // why -10? what is at -9?

// Good
enum class TraversalOrder(val index: Float) {
    Body(-10f),
    Banner(-5f),
    TopBar(-1f),
    // default content sits at 0f
}
Modifier.semantics { traversalIndex = TraversalOrder.Body.index }
```

## Section box: jump to a heading, then descend into it

A screen's sections form a navigable outline through two TalkBack motions: the **heading gesture** jumps between section titles, then **linear swipe descends into** a section, stepping through its children, before reaching the next. To support both, give a section two semantics on **different nodes** — and do **not** merge:

- `heading()` on the **title** → the jump target for heading navigation.
- `isTraversalGroup = true` on the **container** → keeps the section's children grouped and **individually focusable**, so swiping after the title descends through this section's content before leaving it.

`isTraversalGroup` does not merge — children stay focusable, which is exactly what lets the user descend into the section. This is Android's documented "navigable section/heading structure".

```kotlin
@Composable
fun AccessibleSection(
    title: String,
    modifier: Modifier = Modifier,
    content: @Composable () -> Unit,
) = Column(modifier = modifier.semantics { isTraversalGroup = true }) {
    Text(title, modifier = Modifier.semantics { heading() }) // jump target
    content() // stays individually focusable — descendable
}
```

Stack several down a screen and you have its heading outline. Compose has no heading _levels_, so depth is conveyed by typography and reading order, not a level.

**The opposite choice — collapse a section to one stop.** `semantics(mergeDescendants = true) { heading(); contentDescription = summary }` merges the whole section into a **single** node: TalkBack reads one summary and the user can **not** descend into its static children. Only interactive children (buttons, links — anything `clickable` or self-merging) escape the merge and stay separately focusable. Use this for an atomic tile (a card that is one tap target with a couple of actions), **not** for a section the user should explore. If you merge, drop `isTraversalGroup` — there is nothing left to order. (This is why a merged "section box" still lets you reach its buttons but not its body text.)

**Visual headings are not semantic headings.** Text merely styled with a heading typography token, or markdown rendered into one `AnnotatedString` (where `#`/`##`/`###` map only to a `TextStyle`), is invisible to heading navigation — it looks like an outline but exposes none. Each outline entry needs its own `heading()` node; `AnnotatedString` spans cannot carry it, so render such headings as real composables (or add the semantic per node).

### Counter-Example

On a simple linear screen where composition order already equals visual order, do not add `traversalIndex` at all, an unnecessary index is one more thing to keep in sync and can introduce the very mismatch it is meant to prevent. Likewise, never set `collectionInfo` on a `LazyColumn`/`LazyRow`, it is already provided and a manual count will fight the framework's.

## Programmatic focus & bringing content into view

Reading order is only half of operability. These patterns move or reveal focus so keyboard, switch, and screen-reader users are never stranded. (WCAG 2.4.3 Focus Order, 2.4.7 Focus Visible, 3.3.1 Error Identification, 2.1.1 Keyboard.) `isTraversalGroup`/`traversalIndex` (above) set the _accessibility reading order_; the modifiers here set _keyboard/IME focus_ — different APIs, often both needed on the same screen.

- **Bring a focused field into view (keyboard overlap).** Compose does not auto-scroll a focused field above the IME — fix it with the `BringIntoView` relocation API, packaged as one reusable modifier:

```kotlin
fun Modifier.moveIntoViewOnFocus(req: BringIntoViewRequester, scope: CoroutineScope) =
    onFocusEvent { if (it.isFocused) scope.launch { req.bringIntoView() } }
// per field:  val req = remember { BringIntoViewRequester() }
// container:  Modifier.bringIntoViewRequester(req)
// field:      Modifier.moveIntoViewOnFocus(req, rememberCoroutineScope())
```

- **Scroll on error / on submit.** Drive scrolling from _state_, not only focus: `LaunchedEffect(errorMessage, submitSignal) { if (...) { req.bringIntoView(); onConsumed() } }`, with a one-shot consumed flag so recomposition does not re-scroll. Pair with announcing the error via a `liveRegion` (see state-and-announcements.md) so it is both perceivable and on-screen.
- **Move focus to the first error on submit.** Hold a `FocusRequester` per field; in `LaunchedEffect(errorState)` focus the first field in error so the user lands on the problem instead of hunting for it. With an if-cascade, request focus in _reverse_ visual order so the earliest error wins; fall back to a `liveRegion` error banner when no field-level error exists. (3.3.1)
- **Auto-focus on entry — sparingly.** `remember { FocusRequester() }` + `Modifier.focusRequester(it)` + `LaunchedEffect(Unit) { it.requestFocus(); keyboard?.show() }`. Only the first/primary field, and only when the screen is genuinely a single-input task — force-showing the IME disrupts TalkBack users.
- **IME-driven traversal.** Set `imeAction = ImeAction.Next` on all but the last field and advance with `focusManager.moveFocus(FocusDirection.Next)` (uses Compose's computed order — prefer it) or an explicit `nextField.requestFocus()` (manual order). Call `focusManager.clearFocus(force = true)` / `keyboard?.hide()` on the terminal action so the IME closes and focus is not stranded.
- **Focus into transient surfaces.** When a sheet or dialog appears, `requestFocus()` a `FocusRequester` attached to its root content and wrap children in `Modifier.focusGroup()`. Note: `focusGroup()` is keyboard/D-pad focus traversal; `isTraversalGroup` is the accessibility reading order — different APIs.
- **Reorder keyboard / D-pad focus.** Default Tab order follows composable-call order, not the layout — fix it first by grouping/sequencing composables correctly (e.g. a `Row` of two `Column`s so each column is exhausted before the next), then override with `Modifier.focusProperties { next = reqB; previous = reqA }` (1D Tab) or `{ up/down/left/right = ... }` (2D arrows/D-pad), each target carrying a `FocusRequester`. Caveats: only the **topmost** `focusProperties` modifier applies, parents override children, and `FocusRequester.Default` resets to default. This is independent of `traversalIndex` — set both if both audiences need the order.

### Related

./labelling.md for `contentDescription` and merge labelling; ./state-and-announcements.md for announcing changes (and validate-on-blur) without moving focus; ./text-and-targets.md for per-item sizing within collections.
