# validate: Validate Skill Against Spec and Best Practices

## Contents

[Spec Constraints](#spec-constraints) · [Core Workflow](#core-workflow) · [Frontmatter Checks](#frontmatter-checks) · [Body Checks](#body-checks) · [Reference Checks](#reference-checks) · [Content Quality Checks](#content-quality-checks) · [Structural-Pattern Hints](#structural-pattern-hints-soft-signals) · [Harness Neutrality Checks](#harness-neutrality-checks) · [Output](#output) · [Error Handling](#error-handling) · [Safety](#safety)

Read-only audit of a SKILL.md against the Agent Skills spec, project conventions, and authoring best practices. Reports pass/fail per check with line numbers.

A reference implementation lives at [scripts/validate.py](../scripts/validate.py) (PEP 723 self-contained Python): `uv run scripts/validate.py <skill-dir>` runs all checks below and exits non-zero on errors.

## Spec Constraints

- `name`: 1-64 chars, lowercase kebab-case, no consecutive/leading/trailing hyphens, not the reserved words `anthropic`/`claude`, no XML tags, matches parent dir
- `description`: 1-1024 chars, non-empty, imperative "Use when..." phrasing, covers what + when, third person
- Body: target <500 lines / ~5000 tokens; push detail to `references/`
- Optional: `license` (string), `compatibility` (≤500 chars), `metadata` (string→string map), `allowed-tools` (experimental, space-separated)

## Core Workflow

1. Track checks in a task list
2. Resolve target: SKILL.md path or skill directory
3. Parse frontmatter (YAML) and body separately
4. Run all checks (frontmatter, body, references, content quality, harness neutrality)
5. Report pass/fail with file:line evidence and remediation hints
6. Read-only — never modify files

## Frontmatter Checks

- `name` present, 1-64 chars, matches `^[a-z0-9]+(-[a-z0-9]+)*$`, equals parent directory name
- `description` present, 1-1024 chars
- `description` starts with imperative cue ("Use when...", "Use this skill when...")
- `description` includes trigger contexts ("Triggers on...", "even when the user doesn't say...")
- `description` includes skip/scope boundary clauses where adjacent skills exist
- Optional fields obey their limits (`compatibility` ≤500 chars)
- No unknown top-level fields

## Body Checks

- Body ≤500 lines and ~5000 tokens (warn at 80%)
- Has at least one heading
- No trailing whitespace, no Windows line endings
- Code blocks have language markers
- No empty sections

## Reference Checks

- Every `references/{file}.md` link resolves to an existing file
- Each reference link has an explicit load-when trigger (e.g., "Load when X")
- Reference files live one level deep under `references/`
- Reference filenames are kebab-case

## Content Quality Checks

- Has a **Gotchas** or equivalent non-obvious-facts section
- No menus of 3+ equal options (prefer a default + brief alternative)
- No general-knowledge filler (warn on prose that explains what the technology is)

## Structural-Pattern Hints (soft signals)

Soft warnings only — these patterns aid agent execution but aren't required. See [instruction-patterns.md](instruction-patterns.md).

- Multi-step workflow (>3 ordered steps) without a checklist (`- [ ]`) — consider adding one
- Output-producing skill without an output template — consider adding one
- Fragile edits / destructive operations without a validation loop or plan-validate-execute pattern — consider adding one

## Harness Neutrality Checks

- No proprietary tool / function names from any agent harness — describe the capability instead
- No vendor model IDs or model names — describe the role (e.g. "exploration agent")
- No hardcoded vendor-namespaced paths — use placeholders like `<skills-dir>/`, `<commands-dir>/`
- No vendor-prefixed frontmatter keys
- No vendor-specific instruction filenames — use `AGENTS.md` (the open standard)

## Output

```
Validation: <skills-dir>/{skill-name}/SKILL.md

[PASS] name: kebab-case, matches parent dir
[PASS] description: 312 chars, imperative, includes triggers
[PASS] body: 124 lines (~1240 tokens)
[FAIL] references: 2 broken links
  - SKILL.md:42 → references/missing-topic.md (file not found)
  - SKILL.md:58 → references/old-name.md (file not found)
[WARN] references: 3 links lack load-when triggers
  - SKILL.md:42, SKILL.md:51, SKILL.md:60
[FAIL] harness neutrality: 1 vendor-specific reference
  - SKILL.md:30 → proprietary tool name detected

Result: FAIL (2 errors, 3 warnings)

Remediation:
1. Create missing reference files or fix paths
2. Add "Load when X" to bare reference links
3. Replace vendor-specific tool names with generic phrasing
```

## Error Handling

- Target not found → error with suggested paths
- Invalid YAML frontmatter → report line and parse error
- Non-skill markdown (no frontmatter) → skip with info
- Directory without SKILL.md → error

## Safety

- Read-only; never edits the skill
- Recommend running before commit
- Print full diff context (file:line) so remediation is unambiguous
