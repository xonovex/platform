# Sources

## {Tech} {Documentation type, e.g. 'Full Documentation' / 'API Reference'}

- **URL:** {https://example.com/llms-full.txt | https://example.com/docs}
- **Last reviewed:** {YYYY-MM-DD}
- **Used for:**
  - `SKILL.md` → {all sections | specific sections}
  - All files under `references/`
- **Aspects extracted:**
  - {Topic 1} → `references/{topic-1}.md`
  - {Topic 2} → `references/{topic-2}.md`
  - {Topic 3} → `references/{topic-3}.md`

## Refresh Workflow

1. Re-fetch the upstream source(s)
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
