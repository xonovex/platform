# contrast-and-color: Contrast & use of color

## Guideline

Meet the WCAG AA contrast thresholds, and never let color be the only carrier of meaning — pair it with an icon, a text label, and a status-naming `contentDescription`.

## Rationale

Low contrast makes text and controls unreadable for low-vision users and in bright sunlight. Conveying state by hue alone is invisible to the ~8% of users with color-vision deficiency and to screen-reader users. Serves 1.4.3 Contrast (Minimum), 1.4.11 Non-text Contrast, and 1.4.1 Use of Color.

## Contrast ratios (1.4.3, 1.4.11)

Compute the ratio against the color actually composited behind the element, not the window background.

| Element                                      | Minimum ratio | SC     |
| -------------------------------------------- | ------------- | ------ |
| Body text                                    | 4.5:1         | 1.4.3  |
| Large text: ≥ 24sp, or ≥ 18.66sp bold        | 3:1           | 1.4.3  |
| UI component boundary / focus / on-off state | 3:1           | 1.4.11 |
| Icon/graphic needed to understand content    | 3:1           | 1.4.11 |
| Disabled controls, pure decoration           | exempt        | —      |

### How to Apply

1. Pick foreground and background from the same theme (`MaterialTheme.colorScheme.*`, or your design system's semantic tokens) so the pair is designed to pass; do not mix a token foreground over an ad-hoc background.
2. For text, classify size first — a 16sp label needs 4.5:1, a 24sp heading only 3:1.
3. For a control's meaningful boundary (a switch track, an unfilled checkbox, an input outline) check the boundary-to-adjacent-color ratio, not the label.
4. Hint and placeholder text counts as text: it must clear 4.5:1, not the faint gray that "looks like a placeholder".

### Example

```kotlin
// Bad - de-emphasis color used as primary readable text; often below 4.5:1
Text(
    text = stringResource(R.string.search_hint),
    color = MaterialTheme.colorScheme.onSurfaceVariant, // for secondary text, not body copy
)

// Good - readable on-surface color; reserve "variant"/subtle tones for genuinely secondary lines
Text(
    text = stringResource(R.string.search_hint),
    color = MaterialTheme.colorScheme.onSurface,
)
```

## Never rely on color alone (1.4.1)

A status must be decodable with color removed. Encode it three ways: a shape/icon, a visible text label, and a `contentDescription` that names the state in words.

### How to Apply

1. Add a distinguishing icon next to the colored element (check vs. cross, not green dot vs. red dot).
2. Add a visible text label, or merge the status into the semantics of the field it describes.
3. Set a `contentDescription` (or `Modifier.semantics { contentDescription = ... }`) that states the meaning, so the screen reader announces "Error: ..." not just the message text.
4. For form validation, surface the error as text adjacent to the field and reference it in the field's semantics — color on the border is supplementary only.

### Example

```kotlin
// Bad - validity signalled only by border hue; invisible without color
OutlinedTextField(
    value = email,
    onValueChange = onEmailChange,
    isError = isInvalid, // turns the outline red and nothing else
)

// Good - icon + label + spoken state; color is one of several cues
Column {
    OutlinedTextField(
        value = email,
        onValueChange = onEmailChange,
        isError = isInvalid,
    )
    if (isInvalid) {
        val message = stringResource(R.string.email_invalid)
        Row(
            verticalAlignment = Alignment.CenterVertically,
            modifier = Modifier.semantics(mergeDescendants = true) {
                contentDescription = "Error: $message"
                error(message)
            },
        ) {
            Icon(
                imageVector = Icons.Default.Error,
                contentDescription = null, // decorative; meaning is in the row semantics
                tint = MaterialTheme.colorScheme.error,
            )
            Text(text = message, color = MaterialTheme.colorScheme.error)
        }
    }
}
```

## Source colors from semantic theme tokens

### Guideline

Read every color from a semantic theme token; never hardcode a `Color(0xFF...)` literal.

### Rationale

Hardcoded hex bypasses the theme's vetted contrast pairs and silently breaks dark mode — a value that passes 4.5:1 on a light background often collapses on a dark one. Theme tokens swap together per mode; raw hex does not. Serves 1.4.3 and 1.4.11.

### How to Apply

1. Use the theme's semantic roles (`MaterialTheme.colorScheme.onSurface`, `error`, `outline`, …), or your design system's equivalent named tokens.
2. If a needed semantic role does not exist, add/request a token — do not inline a hex value to "fix it for now".
3. Tint icons and draw borders from the same token family as their context so the pair is mode-consistent.

### Example

```kotlin
// Bad - literal that passes in light, fails on the dark surface
Text(
    text = label,
    color = Color(0xFF767676), // 4.5:1 on white, far below 4.5:1 on a dark background
)

// Good - semantic token, contrast guaranteed in both modes
Text(
    text = label,
    color = MaterialTheme.colorScheme.onSurface,
)
```

## Verify both light and dark themes

### Guideline

Check contrast in light AND dark; if dark mode is unsupported, opt out explicitly and treat its contrast set as untested.

### How to Apply

1. Render screenshot/preview cases under both the light and dark theme and eyeball low-contrast text, borders, and status icons.
2. When a screen genuinely cannot support dark mode, force the light theme at that scope rather than letting an untested dark palette render.
3. Re-verify after any token rename or theme override — an override can satisfy one mode while regressing the other.

```kotlin
@Preview(name = "Light") @Composable
private fun ScreenLight() = AppTheme(darkTheme = false) { Screen() }

@Preview(name = "Dark", uiMode = UI_MODE_NIGHT_YES) @Composable
private fun ScreenDark() = AppTheme(darkTheme = true) { Screen() }
```

### Counter-Example

Disabled controls and purely decorative graphics are exempt from 1.4.3 / 1.4.11. Do not boost a disabled button's contrast to "pass" — that misrepresents its state. Keep it visibly muted and expose its disabled status in semantics instead.

### Related

./text-and-targets.md (text legibility and resize); ./labelling.md (the status-naming `contentDescription` that backs up color). Prefer a design system whose semantic tokens are pre-paired for AA on colored surfaces.
