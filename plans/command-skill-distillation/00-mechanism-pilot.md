---
type: plan
has_subplans: false
parent_plan: plans/command-skill-distillation.md
parallel_group: 1
status: pending
feature: command-skill-distillation
dependencies:
  plans: []
  files:
    - packages/command/command-utility/commands/insights-extract.md
    - packages/command/command-utility/commands/insights-instructions-integrate.md
    - packages/command/command-utility/commands/insights-skills-integrate.md
    - packages/command/command-utility/.claude-plugin/plugin.json
    - packages/command/command-utility/.codex-plugin/plugin.json
skills_to_consult: [command-guide, git-guide, pull-request-guide, code-review-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 00 — Mechanism Pilot (insights-* distillation + the four runtime gates)

## Objective

Distill ONE command family end-to-end — the three `insights-*` commands in
`command-utility` — to the thin-delegator shape, wire `xonovex-skill-insights` into the
plugin's `dependencies` array in both manifests, and then PROVE the install-time +
runtime contract holds before any sibling subplan starts. This subplan is the GATE for
the whole feature: if any of the four verification gates fails, mass rollout is blocked
and the recorded workaround (or no-go) governs what subplans 01–05 are allowed to do.

The three command rewrites are near-mechanical; the four gates (dependency
auto-install/`dependency-unsatisfied`, runtime `Skill`-tool load, Codex `.codex-plugin`
parity, `--plugin-dir` dev-loop resolution) are the real work and each gets its own task.

The skill twin already exists and is the source of truth: `insights-guide`
(plugin `xonovex-skill-insights`) exposes the operations
`extract` → `references/extract.md`, `integrate-instructions` →
`references/integrate-instructions.md`, `integrate-skills` →
`references/integrate-skills.md`.

## Tasks

1. **Distill the three `insights-*` commands to the thin shape.** Replace the full body of
   each command with frontmatter (add `Skill` to `allowed-tools`, keep `description` /
   `argument-hint` verbatim), the existing `## Arguments` flag contract verbatim, and a
   `## Delegation` block. Delete Goal / Usage / Workflow / Output Format / Frontmatter
   Fields / Output / Examples / Error Handling / Gotchas / Safety entirely.
   - `packages/command/command-utility/commands/insights-extract.md` → operation **extract**.
     Keep the two `## Arguments` bullets (`category` optional, `--out-dir` default
     `insights/`).
   - `packages/command/command-utility/commands/insights-instructions-integrate.md` →
     operation **integrate-instructions**. Keep the three `## Arguments` bullets
     (`category` required, `--dry-run`, `--agents-file <path>`).
   - `packages/command/command-utility/commands/insights-skills-integrate.md` → operation
     **integrate-skills**. Keep the four `## Arguments` bullets (`category` required,
     `--dry-run`, `--force`, `--output <path>`).

   Target shape (insights-extract shown; mirror for the other two with their own
   arguments + operation):
   ```markdown
   ---
   description: Analyze the session for development mistakes and lessons learned
   allowed-tools:
     - Read
     - Write
     - Glob
     - Grep
     - Skill
   argument-hint: "[category] [--out-dir <dir>]"
   ---

   # /xonovex-utility:insights-extract — Extract Development Lessons

   ## Arguments

   - `category` (optional): Focus on a specific mistake category (e.g., `tool-usage`, `dependencies`, `validation`).
   - `--out-dir` (optional): The directory to save insight files in. Defaults to `insights/`.

   ## Delegation

   Load the `insights-guide` skill (plugin `xonovex-skill-insights`) and perform its
   **extract** operation with these arguments. The skill is the source of truth for the
   procedure, output format, and gotchas — do not restate them.
   ```

2. **Wire `xonovex-skill-insights` into both manifests (STARTS the dependencies array).**
   Add a `dependencies` array to each manifest — subplan 01 appends the remaining utility
   skills (content, instruction, skill, command) to this same array, so introduce it
   exactly once here with the single insights entry.
   - `packages/command/command-utility/.claude-plugin/plugin.json`
   - `packages/command/command-utility/.codex-plugin/plugin.json`

   Delta (bare string; same `xonovex-marketplace`; version omitted):
   ```json
   {
     "name": "xonovex-utility",
     "version": "3.0.0",
     "description": "Xonovex utility commands — skills, instructions, insights, content",
     "author": {"name": "Xonovex"},
     "dependencies": ["xonovex-skill-insights"]
   }
   ```
   (Preserve each file's existing `author` formatting — `.codex-plugin` uses the expanded
   object form; only add the `dependencies` key.)

3. **GATE (a) — dependency enforcement.** Install `xonovex-utility` from the
   `xonovex-marketplace` and confirm `xonovex-skill-insights` auto-installs at the same
   scope. Then uninstall `xonovex-skill-insights` and confirm all three `insights-*`
   commands report `dependency-unsatisfied`. Verify via every surface the parent plan
   names: `claude plugin list`, the `/plugin` UI, and `/doctor`. Record the exact status
   strings each surface prints (they are the acceptance evidence for subplans 01–05).

4. **GATE (b) — runtime `Skill`-tool load + output parity.** With the dependency
   satisfied, invoke `/xonovex-utility:insights-extract` and confirm it loads
   `insights-guide` via the `Skill` tool at run time (not by implicit ambient trigger) and
   that the produced insight files + summary match the pre-distill behavior captured from
   the original fat command. Spot-check `/xonovex-utility:insights-instructions-integrate`
   and `/xonovex-utility:insights-skills-integrate` load their operations too. Record
   whether the `Skill` load is deterministic across repeated invocations.

5. **GATE (c) — Codex `.codex-plugin` parity.** Install/run the same family under Codex and
   confirm `.codex-plugin/plugin.json` `dependencies` triggers auto-install + the
   `dependency-unsatisfied` state, AND that a thin Codex command loads its skill at run
   time. If either is unsupported, record the concrete workaround the parent plan allows:
   keep Codex commands fat, or gate the distillation rollout to Claude Code only. This
   decision constrains every later subplan, so capture it explicitly.

6. **GATE (d) — `--plugin-dir` local dev loop.** Install `command-utility` from a local
   directory via `--plugin-dir` and confirm the `xonovex-skill-insights` dependency
   resolves the same way a marketplace install does. If local-dir installs do NOT resolve
   dependencies, document the marketplace-install dev workflow that subplans 01–05 must use
   instead.

7. **Record results + go/no-go for rollout.** In this subplan's body (Validation Steps /
   Success Criteria evidence — do NOT create a separate report file), capture the verbatim
   outcome of gates (a)–(d): pass/fail, the exact status strings, any Codex/`--plugin-dir`
   workaround, and an explicit GO or NO-GO verdict for subplans 01–05. A NO-GO on any gate
   blocks group 2 until the workaround is adopted.

## Validation Steps

- **type_check**: n/a — markdown/JSON edits only, no app code.
- **lint**: `npx moon run command-utility:fmt-check` — Prettier `fmt:check` passes on the
  touched `command-utility` package (the three commands + both manifests).
- **build**: `npx moon run command-utility:build` — the `command-utility` project builds.
- **tests**: n/a — no unit tests for command/manifest markdown.
- **integration**: invoke `/xonovex-utility:insights-extract` (and spot-check the other two
  `insights-*` commands); confirm each loads `insights-guide` via the `Skill` tool and the
  output is unchanged from the pre-distill fat command. Confirm gates (a)–(d) per tasks
  3–6 across Claude Code and Codex.

## Success Criteria

- [ ] All three `insights-*` commands are thin delegators (~15–20 lines): frontmatter +
      `## Arguments` + `## Delegation`, with `Skill` added to `allowed-tools`; no restated
      Goal/Workflow/Output/Examples/Error Handling/Gotchas.
- [ ] Each command's `## Delegation` names the correct operation
      (`extract` / `integrate-instructions` / `integrate-skills`) on `insights-guide`
      (plugin `xonovex-skill-insights`).
- [ ] Both `.claude-plugin/plugin.json` and `.codex-plugin/plugin.json` carry
      `"dependencies": ["xonovex-skill-insights"]` (array introduced once, ready for
      subplan 01 to extend).
- [ ] GATE (a): marketplace install auto-installs the skill; uninstalling it yields
      `dependency-unsatisfied` in `claude plugin list`, `/plugin`, and `/doctor` — exact
      strings recorded.
- [ ] GATE (b): `insights-extract` loads `insights-guide` via the `Skill` tool at run time
      and output matches pre-distill behavior.
- [ ] GATE (c): Codex parity verified, or a documented workaround (fat Codex commands /
      Claude-Code-only rollout) recorded.
- [ ] GATE (d): `--plugin-dir` dependency resolution verified, or the marketplace-install
      dev workflow documented.
- [ ] An explicit GO / NO-GO verdict for subplans 01–05 is recorded.
- [ ] `fmt:check` and `build` are green on `command-utility`.

## Files Modified / Created

- `packages/command/command-utility/commands/insights-extract.md` — distilled to thin shape.
- `packages/command/command-utility/commands/insights-instructions-integrate.md` — distilled.
- `packages/command/command-utility/commands/insights-skills-integrate.md` — distilled.
- `packages/command/command-utility/.claude-plugin/plugin.json` — add `dependencies` array.
- `packages/command/command-utility/.codex-plugin/plugin.json` — add `dependencies` array.

## Dependencies

None. This is the GATE subplan in parallel_group 1 and depends on no sibling. The skill
twin (`xonovex-skill-insights` / `insights-guide` with its three references) already exists
and is unchanged here. Every other subplan (01–05, groups 2–4) is blocked until this
subplan's four gates pass and its go/no-go verdict is recorded; subplan 01 specifically
extends the same `dependencies` array this subplan introduces, so it must land after this.

## Estimated Duration

~half a day: three small near-mechanical command rewrites plus the two-line manifest delta
are quick; the four verification gates across Claude Code and Codex are the bulk of the
effort.
