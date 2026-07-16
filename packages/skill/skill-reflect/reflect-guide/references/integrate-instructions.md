# integrate-instructions: Fold Insights into AGENTS.md as Bullet Points

Convert a category's insights into concise AGENTS.md bullets.

## Workflow

1. **Gather** — by DEFAULT extract from the current session and hold in memory; `--from-reflections` reads `reflections/*.md`; `--persist` also writes them there. No `category` → all session insights. Group by topic.
2. Locate target AGENTS.md — specified file, or auto-detect from `applies_to` (match directory/package names, nearest AGENTS.md, ask if ambiguous).
3. Convert each insight to 1-2 bullets in AGENTS.md style (backtick names, `—` descriptions, `→` chains); merge into a related bullet group or append a new one.
4. Dedupe against existing bullets; keep only non-obvious details.
5. If insights were persisted, mark those files `applied: true`; append/merge only, never remove existing bullets.

## Gotchas

- Insights that restate code (e.g. "use `useMemo`") are filler — only keep what a fresh reader couldn't infer from the code
- Auto-detection on `applies_to: ["general"]` picks the root AGENTS.md, rarely what you want — require a specific routing key
- Don't bump `applied: true` before the actual write — it leaves orphaned insights
