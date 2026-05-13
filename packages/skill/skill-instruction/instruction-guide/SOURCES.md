# Sources

## AGENTS.md Open Standard

- **URL:** https://agents.md/
- **Last reviewed:** 2026-05-13
- **Used for:**
  - `SKILL.md` → Core Principles, Gotchas
  - `references/init.md` → Output Format Include/Exclude lists, Gotchas
  - `references/consolidate.md` → Implementation, Gotchas
  - `references/sync.md` → Sync Strategy, Gotchas
- **Aspects extracted:**
  - "README for agents" framing — AGENTS.md is operational detail; README.md is human-onboarding → Core Principles role split
  - "Anything you'd tell a new teammate" heuristic → Core Principles + `init.md` Include list framing
  - Nested precedence rule ("closest AGENTS.md to the edited file wins") → Core Principles, `consolidate.md` Implementation + Gotchas, `sync.md` Sync Strategy + Gotchas
  - Programmatic checks: agents auto-execute listed commands → Core Principles, `init.md` Gotchas, `sync.md` Sync Strategy + Gotchas
  - Cross-harness compatibility (Codex, Jules, Cursor, VS Code, Copilot, Roo, Cline, OpenCode, Pi, etc.) → "Open standard" Core Principle
  - Living-documentation principle → Core Principles
  - Content categories to include: build/test commands, setup, code style, testing procedures, security, commit/PR formatting, dev env tips, deployment, dataset handling → `init.md` Include list
  - What to exclude (README-style human-onboarding content) → `init.md` Exclude list
  - "Standard Markdown, flexible headings, no required fields" — informed the loose-format approach but project convention prefers a stricter bullet-list format
