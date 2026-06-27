# accessibility: Accessibility for an Immediate-Mode GUI

## Guideline

An immediate-mode GUI can be accessible — each frame, as you draw a control, also register it (role, label, rect, state) into a small side list that _is_ a retained semantic tree, expose that list to platform accessibility APIs and to automation/virtual input, and reuse the DPI-scale and theme machinery you already have for zoom and high contrast.

## Rationale

The common claim that immediate-mode GUIs can't support screen readers confuses the _API style_ (declare desired state each frame) with the _data_: nothing stops you from also recording, per visible control, the few facts a screen reader needs. Building that side list during the same draw call you already make is cheap and always in sync, and it doubles as an automation surface for UI tests. The same insight that lets a canvas-rendered web app stay accessible (keep a parallel semantic model) applies here — and zoom/contrast aren't new systems at all, they're the DPI-scale factor and theme colors you already expose.

## How to Apply

1. As each control is drawn, call `register_control(ui, role, label, rect)` with its role (button, checkbox, static text, custom), accessible label, and screen rect — accumulating a per-frame list of semantic objects.
2. Extend the registration to carry state the reader needs: enabled/disabled, value (checked/unchecked, slider value), selected/focused, and parent for hierarchy.
3. Expose the list through query entry points (`automation_controls()`, `find_control()`) and accept virtual input (`mouse_move`, `mouse_button_state`, `keyboard_key_state`, `text_input`) so a screen reader or test harness can drive the UI.
4. Bridge that list into the OS accessibility API (UIA/AT-SPI/AX) so real assistive tech sees it.
5. Provide full keyboard navigation and a visible focus model (see [references/events-and-focus.md](./events-and-focus.md)) — accessibility needs every action reachable without a mouse.
6. Get zoom for free by exposing the existing DPI scale factor as a user control; ship a high-contrast theme variant through the existing theming system.

## Example

```c
// Same draw call that renders the button also records its semantics for this frame.
bool button(ui_o *ui, rect_t r, const char *label) {
    bool clicked = draw_and_test_button(ui, r, label);
    register_control(ui, ROLE_BUTTON, label, r);   // side list = retained a11y tree
    return clicked;
}

// Screen reader / UI-test harness reads the list and injects virtual input.
control_t *c = find_control(ui, "Play");
if (c) { mouse_move(ui, center(c->rect)); mouse_button_state(ui, MB_LEFT, true); }
```

## Gotchas

- Registering only role + label + rect is not enough for a usable reader — without state (checked/disabled), value, focus, and hierarchy it can announce a control but not its condition; capture those too.
- The side list must be rebuilt each frame in lockstep with drawing; a control drawn but not registered is invisible to assistive tech, and one registered but not drawn is a ghost.
- Having the semantic list is only half the job — it must actually be fed into the platform accessibility API; an internal list no AT can see helps no one.
- Latin-language localization (see [references/localization.md](./localization.md)) is part of accessibility, but right-to-left and complex scripts (Arabic, Tamil) need real bidi/shaping support, not just string swap.
- Pseudo-localization or a quick language-toggle hotkey to reveal untranslated text doubles as an accessibility audit aid; wire it in early.

## Related

[references/events-and-focus.md](./events-and-focus.md), [references/dpi-scaling.md](./dpi-scaling.md), [references/localization.md](./localization.md)
