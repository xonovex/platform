# text-and-targets: Text resize, reflow & touch-target size

---

## Size text in scalable units (`sp`), never `dp`

**Guideline:** Express every text size and `lineHeight` in `sp` (or inherit them from the design-system type scale); never use `dp`, raw `TextUnit`, or scale-defeating helpers for text.

**Rationale:** `dp` ignores the user's font-size preference, so text cannot scale and fails 1.4.4 Resize Text (text must reach 200% without loss of content or function). Hardcoding a size that bypasses the type scale also tends to drag a fixed `lineHeight` with it, clipping ascenders/descenders at large scales.

**How to Apply:**

1. Prefer a type-scale token from your theme (e.g. `MaterialTheme.typography.bodyLarge` / `labelLarge`) over any literal size — it already carries a scalable `fontSize` and a proportional `lineHeight`.
2. If you must set a literal, use `sp` (`16.sp`), and set `lineHeight` in `sp` or `em` so it scales with the text.
3. Treat `TextUnit.Sp`-defeating helpers (a `nonScaledSp` extension, `pxToSp` math, density-corrected sizing) as red flags — they exist for measuring/drawing, not for body copy.

**Example:**

```kotlin
// Bad — dp is not user-scalable; lineHeight in dp clips when scaled
Text(
    text = label,
    fontSize = 14.dp.value.sp,           // value-laundering a dp into sp
    lineHeight = 18.dp.value.sp,
)

// Good — scalable units, or just inherit the scale token
Text(
    text = label,
    style = MaterialTheme.typography.labelLarge, // fontSize + lineHeight scale together
)
// or, if a literal is unavoidable:
Text(text = label, fontSize = 14.sp, lineHeight = 20.sp)
```

**Counter-Example:** Decorative glyphs drawn onto a `Canvas` (e.g. a route dot whose diameter equals the cap height) legitimately convert text metrics to px; that is rendering math, not content sizing, and is out of scope for 1.4.4.

---

## Let text reflow — no `maxLines=1` + ellipsize that hides content

**Guideline:** Allow text to wrap and its container to grow; do not pin interactive or informational text to a single line that truncates at large font scales or narrow widths.

**Rationale:** A fixed-height row with `maxLines = 1` + `TextOverflow.Ellipsis` drops content the moment the user enlarges text, violating 1.4.4 Resize Text, and breaks 1.4.10 Reflow (content must be usable at 320dp-equivalent width without two-dimensional scrolling for the loss of information).

**How to Apply:**

1. Default to wrapping: omit `maxLines`, or set it generously, and let height be `wrapContentHeight` rather than a fixed `dp`.
2. Replace fixed row heights with `Modifier.heightIn(min = …)` so the row can grow when text wraps.
3. Reserve `maxLines = 1` for genuinely single-token content (a time, a short code); never for free text or labels that vary by locale.

**Example:**

```kotlin
// Bad — clips at 200% font scale and on long translations
Row(modifier = Modifier.height(48.dp)) {
    Text(title, maxLines = 1, overflow = TextOverflow.Ellipsis)
}

// Good — row grows, text wraps
Row(modifier = Modifier.heightIn(min = 48.dp)) {
    Text(title) // wraps freely
}
```

---

## Never defeat font scaling globally (`fontScale`, density)

**Guideline:** Do not force `Configuration.fontScale` to a constant, override `LocalDensity` to cancel the user's scale, or otherwise normalize text size across the app or a screen.

**Rationale:** Pinning `fontScale = 1f` or rewriting density makes 200% impossible everywhere it is applied, a blanket 1.4.4 Resize Text failure that no per-`Text` fix can recover.

**How to Apply:**

1. Never wrap content in a `CompositionLocalProvider(LocalDensity provides Density(density, fontScale = 1f))` to "stabilize" layout — fix the layout to flex instead.
2. Do not set `android:configChanges` or a custom `Resources` to clamp `fontScale`; let the system value flow through.
3. If a snapshot/preview needs a fixed scale, scope it to the `@Preview`/test only, never to shipped composition.

**Example:**

```kotlin
// Bad — cancels the user's font preference for everything below
CompositionLocalProvider(
    LocalDensity provides Density(LocalDensity.current.density, fontScale = 1f),
) { AppContent() }

// Good — keep the real density; verify with a large-font preview instead
AppContent()

@Preview(name = "Large font", fontScale = 2.0f)
@Composable private fun ScreenLargeFontPreview() { MaterialTheme { AppContent() } }
```

**Counter-Example:** A `@Preview(fontScale = 2.0f)` or a screenshot test that pins a scale is correct — it exercises scaling rather than suppressing it.

---

## Touch targets ≥ 48dp without enlarging the visual

**Guideline:** Every clickable element must expose a touch target of at least 48x48dp; expand the hit area, not the icon, when the visual is smaller.

**Rationale:** Small targets are hard to actuate for users with motor or vision impairments. WCAG 2.5.8 Target Size (Minimum) requires 24x24 CSS px at AA, and 2.5.5 Target Size (Enhanced, AAA) wants 44x44; Android Material guidance and `minimumInteractiveComponentSize()` standardize on 48dp, which satisfies both.

**How to Apply:**

1. On any element carrying `Modifier.clickable {}` whose own size may fall below 48dp, add `Modifier.minimumInteractiveComponentSize()` (it grows only the touch area, leaving the drawn content centered at its visual size).
2. Where you control the box directly, use `Modifier.sizeIn(minWidth = 48.dp, minHeight = 48.dp)` or `Modifier.defaultMinSize(minWidth = 48.dp, minHeight = 48.dp)`.
3. Order matters: apply `.clickable` (then `.padding` inside it) so the padding is part of the ripple/hit area, not outside it. Keep the icon itself at its design size (e.g. 24dp) inside the larger target. With `minimumInteractiveComponentSize()`, place it _before_ `.clickable`/`.toggleable` (`.minimumInteractiveComponentSize().clickable {}`) so the enlarged region is the interactive one — the reverse leaves the extra area untappable.
4. Prefer a design-system control that already guarantees the minimum touch target over a bare `Icon` + `clickable`.

**Example:**

```kotlin
// Bad — 32dp icon button: hit area is below the 48dp minimum
Box(
    modifier = Modifier
        .size(32.dp)
        .clickable(onClick = onClose),
    contentAlignment = Alignment.Center,
) {
    Icon(Icons.Default.Close, contentDescription = "Close")
}

// Good — visual stays 24dp, touch area is >= 48dp
Box(
    modifier = Modifier
        .minimumInteractiveComponentSize() // expands hit area only
        .clickable(
            onClick = onClose,
            role = Role.Button,
        ),
    contentAlignment = Alignment.Center,
) {
    Icon(
        imageVector = Icons.Default.Close,
        contentDescription = "Close",
        modifier = Modifier.size(24.dp),
    )
}
```

**Counter-Example:** 2.5.8 exempts targets whose function is also reachable via an equivalent control on the same screen, and inline targets in a sentence/text block; a word styled as a link inside flowing copy need not hit 48dp. Prefer giving it room anyway.

**Related:** ./contrast-and-color.md (text legibility). Prefer design-system type-scale tokens and controls that already meet the touch-target minimum.
