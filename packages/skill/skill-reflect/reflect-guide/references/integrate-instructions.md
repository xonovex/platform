# integrate-instructions: Fold Insights into AGENTS.md as Bullet Points

Convert insights from a category into concise bullet points and integrate them into the appropriate AGENTS.md file.

## Core Workflow

1. **Gather insights** — by DEFAULT, extract lessons from the current session (see [extract.md](extract.md)) and hold them in memory; with `--from-reflections`, read `reflections/*.md` instead; with `--persist`, also write them to `reflections/`. No `category` → take all session insights. Group by topic.
2. Locate target AGENTS.md — use specified file or auto-detect from `applies_to` field
3. Convert each insight to a concise bullet point matching AGENTS.md style
4. Merge into existing file — append to relevant bullet group or create new group
5. If insights were persisted (`--persist` or `--from-reflections`), mark those files `applied: true`; in-memory insights have nothing to flag
6. Preview or write → report

## Conversion Rules

### Format

each insight becomes 1-2 bullet points in AGENTS.md style — backtick-wrapped names, `—` descriptions, `→` chains

### Deduplication

skip insights already covered by existing bullets; merge related insights into single bullets

### Placement

insert near related existing bullets; if no related section exists, append a new bullet group with a blank-line separator

### Brevity

only non-obvious details — skip anything inferable from directory name or project structure

## Auto-Detection

When no target file is provided:

- Use `applies_to` field to match directory names or package names
- Search for nearest AGENTS.md in the matching directory
- If ambiguous, ask the user

## Error Handling

- Missing category → take all session insights (don't block)
- No insights found in the session → report that no lessons were detected; with `--from-reflections`, suggest running `extract` first
- No matching AGENTS.md → ask user for target path
- AGENTS.md not found at path → verify and abort

## Safety

Preview before writing; preserve existing AGENTS.md content and structure; never remove existing bullets; only append or merge.

## Gotchas

- Insights that restate code (e.g. "use `useMemo`") are filler in AGENTS.md — only keep ones a fresh reader couldn't infer from the code
- Auto-detection on `applies_to: ["general"]` will pick the root AGENTS.md, which is rarely what you want — require a more specific routing key
- `applied: true` is the trail of what's already been integrated — bumping it prematurely (before the actual write) leaves orphaned insights
- Apply directly by default — extract straight into AGENTS.md; only write `reflections/*.md` when `--persist` is set or the user explicitly ran `extract` first. Don't store-then-integrate as a forced two-step.
