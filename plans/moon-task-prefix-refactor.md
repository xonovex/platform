---
type: plan
status: complete
---

# Moon task prefixing + ci-check aggregation

## Implementation notes (deviations from the original plan, user-directed)

- **`format` instead of `fmt`** — every task named `fmt`/`fmt-check`/`*-fmt`/`*-fmt-check` is now `format`/`format-check`/`*-format`/`*-format-check`. Underlying commands are unchanged (`go fmt`, `cargo fmt`, `npm run fmt`/`fmt:check`, `gofmt`, `shfmt`).
- **`ci-check` validates skills** — `skill-validate` is now a `ci-check` dep in `tag-skill.yml` and runs in CI (`runInCI:false` removed). `audit-sources`/`eval-triggers`/`eval-outputs` stay out (non-deterministic / LLM-based). All 72 skills pass `skill-validate`.
- **No comments** — all comments removed from every `.moon/tasks/*.yml` and every `moon.yml` (including the doc comment the plan had restored). `.moon/workspace.yml` operational comments were left in place.
- **flake** — `g.docker` (hadolint) added to the default devShell so `docker-lint`, now folded into `agent-operator-go`'s `ci-check` via the docker `lint` alias, resolves on PATH in CI.

## Goal

Prefix every validation/build task in `.moon/tasks/tag-*.yml` with a language
shorthand. Keep the generic task names (`build`, `lint`, `typecheck`, `test`,
`fmt-check`) as **no-op aliases** that depend on the project's language-prefixed
tasks, so `^:build` and multi-tag composition keep working. `ci-check` aggregates
the checks; switch the pre-commit hook to `:ci-check` to match CI.

## Context / why

- **CI already runs `:ci-check`** — `.github/workflows/ci.yml`: `nix develop -c "npm ci && npx moon ci ... :ci-check"`; release runs `:ci-publish`. The pre-commit hook (`.moon/workspace.yml` → `.hooks/run-quality-checks.sh --targets :lint,:typecheck,:test,:fmt-check`) does NOT match CI and misses `go-*`/`sh-*` prefixed tasks.
- `go` and `shell` already follow the prefixed pattern (`go-build`/`go-lint`/…, `sh-lint`/`sh-fmt-check`) and have `ci-check`. They are the reference model — leave them.
- `go-build` works because Go is homogeneous (`go-build` deps `^:go-build`). The TS family (`typescript`/`typescript-config`/`tsconfig`) is multi-tag and cross-depends, so a bare prefixed build breaks `^:build` — the **alias** solves this.

## Already done this session (starting state)

- Version bump `3.2.0 → 4.0.0` committed (`2fff00a`); marketplace prettier-formatted; `packages/skill/AGENTS.md` register doc updated.
- `2fff00a` also swept in an unintended removal of a 5-line doc comment above `validate:` in `.moon/tasks/tag-skill.yml` — **restore it** during this work.

## Confirmed decisions

- **Shorthands:** `ts-` (typescript + typescript-config + typescript-script), `tsconfig-`, `go-` (done), `sh-` (done), `rust-`, `skill-`, `command-`, `docker-` (done), `moon-` (moon-plugin), `npm-` (done, publish-only), `cli` (no tasks).
- **Scope:** prefix all validation tasks AND build (via the alias mechanism, so build-prefixing is safe).

## Architecture: prefixed tasks + generic aliases

Per language tag:

1. **Rename real tasks to prefixed** — `<p>-build`, `<p>-lint`, `<p>-lint-fix`, `<p>-typecheck`, `<p>-test`, `<p>-test-watch`, `<p>-fmt`, `<p>-fmt-check`, `<p>-coverage`, plus tag-specific (`skill-validate`, `skill-audit-sources`, `skill-eval-triggers`, `skill-eval-outputs`). Update intra-task deps to prefixed names. `<p>-build` keeps `deps: [^:build]` (depends on dependency projects' `build` **alias**, not `^:<p>-build`).
2. **Add generic alias tasks** (no-op, just deps) so cross-project `^:build` and multi-tag composition work:
   - `build: { command: noop, deps: [<p>-build] }`
   - `lint: { command: noop, deps: [<p>-lint] }`
   - `typecheck: { command: noop, deps: [<p>-typecheck] }`
   - `test: { command: noop, deps: [<p>-test] }`
   - `fmt-check: { command: noop, deps: [<p>-fmt-check] }`
     moon merges these across a project's tags → a go+typescript project's `build` aliases to `[go-build, ts-build]`, `lint` to `[go-lint, ts-lint]`, etc. (verify moon's merge appends deps; if it replaces, set `mergeDeps: append` on the aliases).
3. **`ci-check`** = `{ command: echo 'ci-check complete', deps: [build, lint, typecheck, test, fmt-check] }` via the aliases (auto-composes per language). Tasks with `runInCI: false` (skill `validate`/`audit-sources`/`eval-*`, the `fmt` fix variants) are NOT in ci-check.

## Per-file changes

- **tag-typescript.yml** — rename `build→ts-build` (deps `^:build`), `test→ts-test`, `test-watch→ts-test-watch`, `lint→ts-lint` (deps `[ts-build, ^:build]`), `lint-fix→ts-lint-fix`, `fmt→ts-fmt`, `fmt-check→ts-fmt-check`, `typecheck→ts-typecheck` (deps `[ts-build, ^:build]`), `coverage→ts-coverage`. Add aliases (build/lint/typecheck/test/fmt-check). `ci-check` deps `[build, lint, typecheck, test, fmt-check]`.
- **tag-typescript-config.yml** (extends typescript) — rename `test→ts-test` override; keep `clean`, `configs`. Inherits ts-* + aliases + ci-check from parent; drop its own `ci-check` override (or align to `[build, lint, typecheck, test, fmt-check]`).
- **tag-typescript-script.yml** (extends typescript) — rename the overrides `build→ts-build`, `test→ts-test`, `test-watch→ts-test-watch`, `lint→ts-lint`, `lint-fix→ts-lint-fix`, `typecheck→ts-typecheck`, `coverage→ts-coverage` (keep `mergeDeps: replace`). Inherits aliases + ci-check.
- **tag-tsconfig.yml** — rename `build→tsconfig-build`, `typecheck→tsconfig-typecheck`, `lint→tsconfig-lint`, `lint-fix→tsconfig-lint-fix`, `fmt→tsconfig-fmt`, `fmt-check→tsconfig-fmt-check`; keep `clean`, fileGroups. Add aliases. `ci-check` deps `[build, lint, typecheck, fmt-check]`.
- **tag-skill.yml** — rename `fmt→skill-fmt`, `fmt-check→skill-fmt-check`, `validate→skill-validate`, `audit-sources→skill-audit-sources`, `eval-triggers→skill-eval-triggers`, `eval-outputs→skill-eval-outputs`. Restore the removed doc comment above `skill-validate`. Add `fmt-check` alias → `[skill-fmt-check]` (and `fmt` alias → `[skill-fmt]` if wanted). Skill `build` is the package-local echo (stays; no `skill-build`). `ci-check` deps `[build, fmt-check]` (skill-validate/eval are runInCI:false, excluded).
- **tag-command.yml** — rename `fmt→command-fmt`, `fmt-check→command-fmt-check`. Add `fmt-check` alias. `build` is package-local echo. `ci-check` deps `[build, fmt-check]`.
- **tag-rust.yml** — rename `lint→rust-lint`, `fmt-check→rust-fmt-check`. Add `lint`/`fmt-check` aliases → `[rust-lint]`/`[rust-fmt-check]`. No build/test in this tag (rust crates that build live under moon-plugin).
- **tag-moon-plugin.yml** — rename `build→moon-build`, `test→moon-test`. Add `build`/`test` aliases → `[moon-build]`/`[moon-test]`. `ci-check` deps `[build, test, lint, fmt-check]` (lint/fmt-check resolve via the `rust` tag's aliases, since moon-plugin packages are also `rust`-tagged).
- **tag-docker.yml** — add `lint` alias → `[docker-lint]` and `ci-check` deps `[lint]`.
- **tag-go.yml / tag-shell.yml** — already prefixed + have ci-check. Add the generic aliases (`build→[go-build]`, `lint→[go-lint]`, …; `lint→[sh-lint]`, `fmt-check→[sh-fmt-check]`) so multi-tag composition and `^:build` are uniform, and point their `ci-check` at the aliases. Keep `go-build` deps `^:go-build` OR switch to `^:build` (alias) — prefer `^:build` for cross-language uniformity; verify Go-only chains still order correctly.
- **tag-npm.yml / tag-cli.yml** — publish-only / no tasks; no change (npm packages get checks from their typescript/tsconfig tag).

## Package `moon.yml` changes

- **agent-cli-go** (tags go+typescript+npm+cli) defines local `lint`/`typecheck`/`test` overriding the typescript tag's tasks. Rename these locals to `ts-lint`/`ts-typecheck`/`ts-test` (or `go-*` if they target Go) so the override still lands on the prefixed task. Audit the other 6 `agent-*-go*` `moon.yml` for local `lint`/`typecheck`/`test`/`fmt-check` and rename equivalently.
- 72 skill + 2 command packages define a local `build:` echo — leave as-is (the `build` alias is only added by language tags that have a real build; skill/command `build` is the placeholder the ci-check depends on).

## Hook switch

- `.moon/workspace.yml` → `vcs.hooks.pre-commit`: change `--targets :lint,:typecheck,:test,:fmt-check` to `--targets :ci-check`. Regenerate the generated hook (`npx moon sync hooks`), then confirm `.moon/hooks/pre-commit` calls `:ci-check`.

## Validation

1. `npx moon sync projects` / `npx moon action-graph :ci-check` — confirms all task deps resolve (catches "task not found" without running). Run inside `nix develop` for Nix tools.
2. `nix develop -c "npx moon ci :ci-check --concurrency 4"` on a clean tree (or `--affected`) — must pass for skill, command, typescript, typescript-script, tsconfig, go, rust/moon-plugin packages.
3. Stage a trivial skill markdown change and run `.hooks/run-quality-checks.sh --staged --targets :ci-check` — confirm it runs `skill-fmt-check` and passes.
4. `grep -rn ':lint\|:typecheck\|:fmt-check\|:test' .moon .hooks .github` — confirm no stale generic `:task` target references remain (all routed via `:ci-check`).

## Gotchas

- moon merges same-named tasks across a project's tags; verify the alias `deps` **append** across tags (else multi-tag projects only get one language's checks) — set `options.mergeDeps: append` on aliases if needed.
- Keep `runInCI: false` tasks (skill validate/audit/eval, fmt fix variants) OUT of `ci-check`.
- `^:build` must target the `build` **alias** (every buildable project has it); do not point it at a language-specific `<p>-build`.
- Restore the `tag-skill.yml` doc comment dropped in `2fff00a`.
- Do NOT reflow `.claude-plugin/marketplace.json` back to compact — it is now prettier-formatted (AGENTS.md updated).
- Verify inside `nix develop` (CI's tools: node/npm from `/nix/store`; note python3/uv/graphviz are NOT in the flake — irrelevant for moon tasks, which use node).

## As-built notes (deviations directed during implementation)

- **`fmt` → `format`:** all fmt task names renamed (`fmt`/`fmt-check` aliases and every `*-fmt`/`*-fmt-check` prefixed task → `format`/`format-check`). Underlying commands unchanged (`go fmt`, `cargo fmt`, `npm run fmt`, `gofmt`, `shfmt`, and the `fmt`/`fmt:check` npm scripts).
- **`skill-validate` folded into skill `ci-check`:** `skill-validate` is now `runInCI: true` and a dep of the skill `ci-check` (`deps: [build, format-check, skill-validate]`). All 72 skills pass validation. `skill-audit-sources`/`skill-eval-triggers`/`skill-eval-outputs` stay `runInCI: false` and excluded.
- **No comments:** every comment removed from `.moon/tasks/*.yml` and all project `moon.yml` (this supersedes the plan's "restore the tag-skill doc comment" step). `.moon/workspace.yml` comments were left in place.
- **flake support for `docker-lint`:** `g.docker` (hadolint) added to the default devShell in `flake.nix`, since the docker tag's `lint` alias folds `docker-lint` into `agent-operator-go`'s `ci-check` and hadolint was otherwise absent from CI's `nix develop`. The Dockerfile passes hadolint clean.
