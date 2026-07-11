# localization: Localizing an Immediate-Mode GUI

## Guideline

Wrap every user-visible string in a `LOCALIZE(...)` marker that hashes the _source string itself_ as the lookup key, resolve it each frame through a swappable localizer interface that falls back to the source string on a miss, and verify coverage with a pseudo-localization ("gibberish") mode plus an extraction tool — rather than managing numeric string IDs.

## Rationale

Numeric/string-ID schemes need a maintained ID table: new strings add header entries, IDs collide in merges, removed strings orphan entries. Hashing the source string makes the call site _the_ key — no table to sync, no merge conflict, and a missing translation degrades to readable English instead of a blank or ID. The per-frame hash cost is negligible because only the handful of strings visible this frame are localized. The marker macro also gives a tool one token to scan, making coverage and dead-entry detection mechanical.

## How to Apply

1. Mark every displayed string with `LOCALIZE("text")`; the macro hashes the literal to a key and looks it up in the current localizer.
2. For identical strings with different meanings, add a context argument — `LOCALIZE_CTX("File", "menubar")` — and fold the context into the hash so homonyms resolve separately.
3. Resolve through an abstract localizer interface (current-language table → hashmap), so a plugin can supply a different translator; on a miss, return the source string.
4. Build per-language tables in data; load the active language's table into a hashmap at startup or language switch, and hot-swap it to change language live.
5. Run an extraction tool over the source that finds all `LOCALIZE` calls, diffs them against the tables, and reports untranslated and unused entries in CI.
6. Add a pseudo-localization mode that transforms each string into recognizable nonsense _and_ pads it ~40% longer, to surface both unmarked strings and layout that breaks under expansion.

## Example

```c
// Source string is the key — no ID table to maintain.
ui_label(ui, LOCALIZE("Play"));
ui_menu_item(ui, LOCALIZE_CTX("File", "menubar"));   // distinct from "File" the verb

// Localizer interface: swappable, falls back to the source string on a miss.
typedef struct localizer_i {
    const char *(*translate)(uint64_t key, const char *source);
} localizer_i;

static const char *table_lookup(uint64_t key, const char *source) {
    const char *hit = hashmap_get(&g_active_lang, key);
    return hit ? hit : source;                 // missing translation -> readable English
}
// Pseudo-loc build swaps in a localizer that returns padded gibberish for every key.
```

## Gotchas

- Hashing the source string means _editing the English text changes the key_ — a copy tweak silently orphans the existing translation; treat string edits as re-translation work.
- Any string built by concatenation or `printf` from fragments can't be localized as a unit (word order differs per language) — localize whole format templates, not pieces.
- Forgetting to mark a string fails silently; the gibberish mode is what catches it — run it regularly, not once.
- A side table of localized literals can drift from the code; the extraction tool, run in CI, is what keeps unused/missing entries from accumulating.
- This pattern covers Latin text swap only; bidirectional and complex-script layout and non-Latin glyph atlases are separate problems (see the font-atlas handling in [references/dpi-scaling.md](./dpi-scaling.md)).

## Related

[references/ids-and-state.md](./ids-and-state.md), [references/accessibility.md](./accessibility.md), [references/dpi-scaling.md](./dpi-scaling.md)
