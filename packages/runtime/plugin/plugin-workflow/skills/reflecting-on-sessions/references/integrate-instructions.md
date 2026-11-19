# integrate-instructions: Fold Insights into AGENTS.md as Bullet Points

Convert a category's insights into concise AGENTS.md bullets.

## Inputs

- **Category** (optional): one category to focus on, such as testing; the default takes all session insights and routes each to the nearest AGENTS.md through its `applies_to`
- **From reflections** (optional): take the insights from existing `reflections/*.md` files in place of the session, which is the second step after a separate extraction
- **Persist** (optional): also write the insights as `reflections/*.md` for an audit trail, in a named directory when one is given; off by default
- **Instruction file** (optional): the target AGENTS.md; detected from `applies_to` when absent
- **Preview** (optional): show the result without modifying a file

## Workflow

1. **Gather**: by DEFAULT extract from the current session and hold in memory; the from-reflections input reads `reflections/*.md`; the persist input also writes them there. No category → all session insights. Group by topic.
2. Locate target AGENTS.md: specified file, or auto-detect from `applies_to` (match directory/package names, nearest AGENTS.md, ask if ambiguous).
3. Convert each insight to 1-2 bullets in AGENTS.md style (backtick names, `:` descriptions, `→` chains); merge into a related bullet group or append a new one.
4. Dedupe against existing bullets; keep only non-obvious details.
5. If insights were persisted, mark those files `applied: true`; append/merge only, never remove existing bullets.

## Gotchas

- Insights that restate code (e.g. "use `useMemo`") are filler, only keep what a fresh reader couldn't infer from the code
- Auto-detection on `applies_to: ["general"]` picks the root AGENTS.md, rarely what you want: require a specific routing key
- Don't bump `applied: true` before the actual write: it leaves orphaned insights
