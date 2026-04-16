---
description: Create an AGENTS.md file for a directory by analyzing its structure and contents
model: sonnet
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Bash
  - TodoWrite
  - AskUserQuestion
argument-hint: "[directory] [--dry-run] [--recursive]"
---

# /xonovex-utility:instructions-init – Create AGENTS.md

Generate a concise AGENTS.md for a directory by scanning its structure, config files, and code.

## Arguments

- `directory` (required): Target directory
- `--dry-run` (optional): Preview without writing
- `--recursive` (optional): Also create AGENTS.md for subdirectories with unique content

## Core Workflow

1. Use TodoWrite to track steps
2. Verify no AGENTS.md exists (abort if present — use sync/simplify instead)
3. Scan structure, configs, code patterns, git history
4. Generate AGENTS.md in standard format → preview or write → report

## Discovery

**Configs:** package.json, moon.yml, CMakeLists.txt, Dockerfile, go.mod, tsconfig.json, Cargo.toml, pyproject.toml — detect project type (TypeScript, Go, C, Python, Rust), extract build/test commands, entry points, exports

**Structure:** 1 level deep, exclude (node_modules, .git, build, dist, coverage, .artifacts, **pycache**); scan README.md for architectural context

## Output Format

`# Title` (humanized dir name, e.g., `agent-operator-go` → `Agent Operator Go`) + flat bullet list — no `##` headings, no prose

- Backtick-wrap commands/files/keys; `—` descriptions, `→` workflow chains
- Parenthetical file lists: `(main.tf, vars)`; dash-separated examples: `- local, staging`
- Subdirectory bullets: `` `<dir>/` `` — description (key files) - examples
- **Include:** Caveats, non-obvious config, build/test commands, key exports, architectural decisions
- **Exclude:** Descriptions restating dir name, predictable guideline links, boilerplate
- **Brevity:** 5-15 lines; nothing non-obvious → skip, report "No unique content found"

## Examples

```bash
/xonovex-utility:instructions-init packages/agent/agent-cli
/xonovex-utility:instructions-init infrastructure/ --recursive
/xonovex-utility:instructions-init services/api --dry-run
```

## Error Handling

- AGENTS.md exists: abort, suggest sync or simplify
- Empty/no unique content: report and skip
- Permission errors: warn and skip

## Safety

Never overwrite existing AGENTS.md, only create files with genuinely useful content, use `--dry-run`, recommend git commit after `--recursive` runs.
