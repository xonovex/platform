# Sources

## Pull request best practices (web, 2025-2026)

- **URLs:**
  - https://www.deployhq.com/blog/the-perfect-pull-request-best-practices-for-collaborative-development
  - https://graphite.com/guides/best-practices-managing-pr-size
  - https://www.em-tools.io/engineering-metrics/pull-request-size
  - https://gitrolysis.com/posts/2026/01/how-to-write-better-pull-request-descriptions-templates-and-examples/
  - https://willowvoice.com/blog/how-to-write-good-pull-request-description
  - https://www.pullrequest.com/blog/writing-a-great-pull-request-description/
- **Last reviewed:** 2026-06-24
- **Used for:**
  - `SKILL.md` → Essentials, Gotchas, Example
  - All files under `references/`
- **Aspects extracted:**
  - Description what/why/how, length-to-scope, links, screenshots → `references/description.md`
  - PR size limits, atomicity, the infra/scaffolding exception → `references/size-and-atomicity.md`
  - Testing evidence beyond "tested locally" → `references/testing-evidence.md`
  - Surfacing tradeoffs and limitations early → `references/tradeoffs.md`
  - Self-review, green CI before assigning → `references/self-review.md`
  - Lean templates, related-PR links → `references/templates.md`

## Refresh Workflow

1. Re-fetch the upstream source(s)
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
