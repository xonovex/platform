# using-scripts: Bundling Executables in Skills

## When to Bundle a Script vs Inline a Command

- **One-off command** — existing tool, ≤2-3 flags, easy to get right inline (`npx eslint@9 --fix .`)
- **Bundled script** (`scripts/{name}.{ext}`) — complex command, repeated logic, custom validation, or anything tested
- Signal to bundle: agent re-derives the same logic across runs (seen in execution traces)

## One-Off Commands

- **Pin versions** — `npx eslint@9.0.0`, `uvx ruff@0.8.0`, `go run pkg@v1.2.3`
- **State prerequisites** in SKILL.md or use the `compatibility` frontmatter field
- Pick a single default runtime, mention alternatives briefly (don't list 4 equivalent options)

## Self-Contained Scripts

Declare dependencies **inline** so the agent runs the script with one command — no install step.

- **Python (PEP 723)** — `# /// script\n# dependencies = ["pkg"]\n# ///` → `uv run scripts/x.py`
- **Deno** — `import x from "npm:cheerio@1.0.0"` → `deno run scripts/x.ts`
- **Bun** — `import x from "cheerio@1.0.0"` → `bun run scripts/x.ts` (no `node_modules` parent dir)
- **Ruby** — `bundler/inline` `gemfile do … end` → `ruby scripts/x.rb`

Pin versions; document required runtime in `compatibility`.

## Reference Scripts from SKILL.md

- Use **relative paths from the skill root** — agent resolves these automatically
- List scripts in an "Available scripts" section so the agent knows they exist:

```markdown
## Available scripts

- **`scripts/validate.sh`** — Validates configuration files
- **`scripts/process.py`** — Processes input data
```

## Designing Scripts for Agentic Use

- **No interactive prompts** — agents run in non-interactive shells; blocking on TTY hangs forever. Accept input via flags, env vars, or stdin.
- **`--help`** is the primary interface doc — brief description, flags, examples
- **Helpful error messages** — say what went wrong, what was expected, what to try
- **Structured output** — JSON / CSV / TSV to stdout; diagnostics to stderr
- **Idempotent** — "create if not exists" survives retries
- **Closed-set inputs** — reject ambiguous input with a clear error rather than guessing
- **`--dry-run`** for destructive / stateful operations
- **Meaningful exit codes** — distinct codes per failure mode; document in `--help`
- **Safe defaults** — destructive ops require `--confirm` / `--force`
- **Bounded output** — many harnesses truncate at 10-30K chars; default to summary, support `--offset` / `--output`

## Gotchas

- An existing `node_modules` anywhere up the tree disables Bun's auto-install
- Native-addon packages (node-gyp) often fail under Deno — prefer packages shipping prebuilt binaries
- `pipx` and `uvx` are not bundled with Python — state the prereq
- A long-running script that ignores `--output` and floods stdout can blow past the harness's truncation limit silently

## Error / Output Examples

```
# Bad: hangs on input
$ python scripts/deploy.py
Target environment: _

# Good: clear, actionable error
$ python scripts/deploy.py
Error: --env is required. Options: development, staging, production.
Usage: python scripts/deploy.py --env staging --tag v1.2.3
```

```
# Bad: opaque
Error: invalid input

# Good: directs the next attempt
Error: --format must be one of: json, csv, table.
       Received: "xml"
```
