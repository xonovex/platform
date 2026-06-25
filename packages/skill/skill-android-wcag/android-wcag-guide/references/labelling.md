# labelling: Screen-reader labelling & non-text content

**Guideline:** Every meaningful graphic carries a localized text alternative, every decorative graphic is silenced, and every composite control exposes exactly one focusable node with one label and one role to TalkBack.

**Rationale:** TalkBack reads only what the semantics tree exposes. Missing or wrong alternatives violate 1.1.1 Non-text Content; a control with no name/role or an inconsistent one violates 4.1.2 Name, Role, Value. Over-exposed children turn one tile into a dozen swipe stops and read raw concatenated junk; under-exposed controls are invisible to the assistive layer entirely.

**How to Apply:**

1. For an `Image`/`Icon` that conveys information, set `contentDescription` from `stringResource(...)` with positional format args, never a literal. For a purely decorative graphic, set `contentDescription = null`.
2. For a composite control (a clickable tile, a card behaving as a button), wrap it in `Modifier.semantics(mergeDescendants = true) { ... }` and set `contentDescription`, `role`, and (if stateful) `stateDescription` inside the block. Never leave the block empty.
3. To collapse a control's existing descendant semantics and substitute a single curated label, use `Modifier.clearAndSetSemantics { contentDescription = ...; role = ... }`.
4. Drop child nodes whose meaning is already in the parent label by calling `invisibleToUser()` inside their semantics block.

**Example:**

```kotlin
// Bad — hardcoded literal description breaks localization;
//        decorative chevron read aloud as noise; no role.
Row(modifier = Modifier.clickable(onClick = onClick)) {
    Icon(Icons.Default.MusicNote, contentDescription = "Nummer") // literal, untranslated
    Text(title)
    Text(artist)
    Icon(Icons.Default.ChevronRight, contentDescription = "Chevron") // decorative noise
}

// Good — one focusable node, localized composite label, explicit role,
//        decorative chevron silenced.
Row(
    modifier = Modifier
        .clickable(onClick = onClick)
        .semantics(mergeDescendants = true) {
            contentDescription = label // localized at call site (see below)
            role = Role.Button
        },
) {
    Icon(Icons.Default.MusicNote, contentDescription = null)    // informational meaning is in `label`
    Text(title)
    Text(artist)
    Icon(Icons.Default.ChevronRight, contentDescription = null) // purely decorative
}
```

Build the composite label from a resource with positional args so word order survives translation:

```kotlin
val label = stringResource(R.string.track_by_artist, title, artist)
// strings.xml:    <string name="track_by_artist">%1$s by %2$s</string>
// strings-nl.xml: <string name="track_by_artist">%1$s van %2$s</string>
```

Stateful composite (assign a role and a separately-announced state):

```kotlin
Modifier
    .toggleable(value = checked, role = Role.Switch, onValueChange = onToggle)
    .semantics(mergeDescendants = true) {
        contentDescription = stringResource(R.string.notify_new_releases)
        stateDescription = if (checked) onLabel else offLabel // localized strings
    }
```

`clearAndSetSemantics` when a child subtree announces too much and you want one clean label:

```kotlin
// Replaces the whole descendant tree's semantics with a single label + role.
Modifier.clearAndSetSemantics {
    contentDescription = stringResource(R.string.price, amount)
    role = Role.Button
}
```

**Counter-Example (deliberate hide idiom):** To remove a node TalkBack should skip, the `invisibleToUser()` flag is what guarantees the skip — not an empty string. Setting only `contentDescription = ""` still leaves a focusable, empty node.

```kotlin
Modifier.clearAndSetSemantics {
    contentDescription = ""
    invisibleToUser()      // this flag does the work; the empty string alone does not
}
```

On newer Compose this flag is spelled `hideFromAccessibility()` (`Modifier.semantics { hideFromAccessibility() }`); `invisibleToUser()` is the older experimental name for the same thing. Reach for it whenever a node is **visually** gone but still in the tree — most often a `graphicsLayer { alpha = 0f }` fade. A collapsing header that cross-fades an expanded title into a toolbar title keeps **both** in the semantics tree, so TalkBack reads the title twice unless the faded-out copy is gated:

```kotlin
Column(
    modifier = Modifier
        .graphicsLayer { alpha = 1f - collapseProgress }
        .then(if (collapseProgress < 0.5f) Modifier else Modifier.semantics { hideFromAccessibility() }),
) { /* expanded title */ }
```

**Counter-Example (spell codes out for text-to-speech):** A confirmation/reference code rendered as visible text ("QPBMWPK") is read by TTS as a garbled run-on word. Keep the compact visible glyphs but expose a spelled-out `contentDescription` ("Q P B M W P K") so the code is intelligible spoken. In server-driven UI this is a backend concern — the BFF sends a dedicated spelled-out field; the client only applies it.

**Counter-Example (when `null` is wrong):** `contentDescription = null` is correct only for graphics whose meaning is fully carried by adjacent text or a parent merge label. A standalone icon-only button (no visible text) must carry a real localized description, or it has no accessible name (4.1.2).

## Carry the spoken label as data, build it where it is testable

**Guideline:** For data-driven rows, cards, and clusters, do not assemble the `contentDescription` inline in the composable. Build it where it is pure, testable logic; the composable only applies it.

**Rationale:** A description assembled inline is untested and drifts from the value it describes. Building it in the data/use-case/mapper layer makes it a regression-protected output (1.1.1, 1.3.1, 4.1.2) and is the natural seam for server-driven UI, where the backend supplies the string verbatim.

**How to Apply:**

1. Pair the visible value with its prebuilt description in one object so they cannot drift: `data class A11yText<T>(val value: T, val description: String)`.
2. Build the `description` in the data layer. Composable code may use `stringResource`; non-composable code depends on an injected string-resolver interface (`resolve(@StringRes): String`) rather than leaking `Context`, so it stays a plain JVM unit test.
3. Apply it with `clearAndSetSemantics { contentDescription = it.description }` on the cluster's sub-region — not blindly the whole row, so meaningful siblings (times, status) keep their own nodes.

```kotlin
// Domain/data layer — pure, unit-testable, no Compose
data class A11yText<T>(val value: T, val description: String)

class DescribePlaylist(private val res: StringResolver) {        // injected resolver, not Context
    operator fun invoke(tracks: List<Track>): A11yText<List<Track>> =
        A11yText(tracks, tracks.joinToString(res.get(R.string.then)) { it.name } + ".")
}

// UI — only applies the prebuilt string to one consolidated node
Box(Modifier.clearAndSetSemantics { contentDescription = playlist.description }) {
    PlaylistRow(playlist.value)
}
```

**Where to factor out:** centralize the _string_ (a pure function, an injected resolver, or a value+description model); keep `clearAndSetSemantics { ... }` inline at each call site rather than building a generic `Modifier.a11yLabel` helper. The localized branching text is the reuse- and test-worthy part; inline semantics let each site add its own `testTag` or extra properties. Then unit-test the built string — see the testing reference.

## Counter-examples to avoid

- Hardcoded literal `contentDescription` (English or Dutch): the string never localizes, so non-default-locale users hear the wrong language. Always go through `stringResource`.
- An empty `semantics(mergeDescendants = true) { }` block: merging without setting `contentDescription` concatenates children in raw layout order and silently drops any intended single label, heading, or role. Either set the label inside the block or do not merge.
- Leaving decorative icons with a description: TalkBack reads "Chevron", "Decorative line", etc. as standalone stops. Decorative graphics take `contentDescription = null`.

**Related:** ./focus-order.md (headings and grouping); ./state-and-announcements.md (announcing state vs naming a node). When a design system supplies pre-labelled components, compose them and re-label at the tile level rather than re-describing each child.
