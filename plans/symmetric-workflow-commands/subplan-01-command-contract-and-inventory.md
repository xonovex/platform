---
type: plan
has_subplans: false
parent_plan: ../symmetric-workflow-commands.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files: []
skills_to_consult:
  - command-guide
  - orthogonal-pattern-guide
  - git-guide
  - testing-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 01: Command contract and inventory

## Objective

Replace the 53 artifact- and lifecycle-specific command prompts with exactly
eight core operation commands and four workspace utilities. Freeze a compact,
provider-neutral prompt contract that keeps kind, perspective, method,
executor, trigger, provider, reference, and agent capability independent.

This subplan owns the command files and their immediate package dependencies.
It does not rewrite role documentation, planning-skill content, validation
tooling, or release versions.

## Tasks

### 1. Remove the legacy command inventory

- Delete every existing file under
  `packages/command/command-workflow/commands/`.
- Do not retain aliases, redirects, wrappers, deprecation notices, or a
  migration file.
- Confirm the replacement directory contains only:

```text
abandon.md
create.md
decide.md
execute.md
publish.md
review.md
revise.md
validate.md
workspace-abandon.md
workspace-cleanup.md
workspace-create.md
workspace-merge.md
```

### 2. Define one input/result grammar for the eight core operations

- Author `create.md`, `review.md`, `revise.md`, `decide.md`, `execute.md`,
  `validate.md`, `publish.md`, and `abandon.md` using the command-guide
  frontmatter and prompt conventions.
- Use one vocabulary consistently:
  `subject`, `reference`, `revision`, `kind`, `perspective`, `criteria`,
  `method`, `capability`, `provider`, and `result`.
- Keep kind, perspective, method, capability, and provider open selections;
  never turn them into a closed profile or lifecycle enum.
- Require commands to report any inferred kind/provider and refuse ambiguous
  inference instead of guessing.
- Make inline/session-only output valid. Persistence is explicit and returns a
  provider-native locator when a provider supports one.
- Keep `decide` descriptive: it records an outcome and rationale but sets no
  gate or authority state.
- Keep `publish` and destructive cleanup confirmable or dry-runnable unless the
  original request already explicitly authorized the effect.

Use this semantic shape, adapting syntax to the command format:

```text
Operation: <create|review|revise|decide|execute|validate|publish|abandon>
Subject: <inline content or provider-native reference>
Selection: kind?, perspective?, method?, capability?, provider?, revision?
Criteria/feedback: explicit when the operation requires them
Result: inline content and/or provider-native destination reference
```

### 3. Keep capability selection soft and explicit

- Each command owns only its universal operation semantics.
- Load a domain, method, or provider skill only when the caller selects it or
  when the subject makes the capability unambiguous.
- If an explicitly selected capability is unavailable, name it and stop with an
  actionable error.
- Do not introduce an umbrella workflow skill, central resolver, registry,
  runtime, policy engine, or implicit default provider.
- Do not branch behavior based on manual/hook/CI invocation, human/script/LLM/
  agent execution, role, or A1/A2/A3 labels.

The operation prompt may use logic equivalent to:

```text
Resolve only explicit or unambiguous capabilities.
Interpret references through the selected provider capability.
Apply this command's operation invariant.
Return the result and any provider-native reference/revision.
```

### 4. Distill workspace management into four sibling utilities

- Create `workspace-create.md`, `workspace-merge.md`,
  `workspace-abandon.md`, and `workspace-cleanup.md` from the useful mechanics
  in the former `plan-worktree-*` commands.
- Remove plan status, approval, lifecycle, and profile assumptions.
- Require an explicit workspace target and preserve repository-specific safety
  checks from the git guide.
- Keep core operations independent: none implicitly creates, merges, abandons,
  or cleans a workspace.

### 5. Remove hard skill dependencies from the command package

- Update `packages/command/command-workflow/package.json:5-11` so the package
  does not declare plan, code-quality, code-review, pull-request, or testing as
  universal runtime dependencies.
- Update `packages/command/command-workflow/moon.yml:4-9` to remove the matching
  project dependencies.
- Preserve formatting/test scripts and the cross-package link task. Any task
  input changes needed for the new docs belong to subplan 05.
- Keep `.claude-plugin/plugin.json` and `.codex-plugin/plugin.json` structurally
  valid, but defer version and catalog-description changes to subplan 05.

### 6. Perform command-level verification

- Count commands from disk and assert the exact 12-name inventory manually
  before the validator is strengthened in subplan 05.
- Inspect every new prompt for role names, `--profile`, workflow stages,
  approval gates, authority claims, provider-specific parsing, trigger modes,
  executor modes, and A1/A2/A3 behavior.
- Run package formatting and current documentation/link checks, accepting only
  failures caused by documentation that subplan 03 explicitly owns.

## Validation steps

1. `find packages/command/command-workflow/commands -maxdepth 1 -type f -name '*.md' | sort`
2. `test "$(find packages/command/command-workflow/commands -maxdepth 1 -type f -name '*.md' | wc -l)" -eq 12`
3. `rg -n -- '--profile|mandatory.*gate|approval.*gate|authority-reference|A1|A2|A3|AgentTrigger|AgentSchedule' packages/command/command-workflow/commands` must return no matches.
4. `npx moon run command-workflow:fmt-check`
5. `npx moon run command-workflow:build`
6. `git diff --check`

## Success criteria

- [ ] Exactly eight core operation prompts and four workspace prompts exist.
- [ ] Every prompt has one narrow operation invariant and uses the common
      subject/reference/result vocabulary.
- [ ] Kind, perspective, method, capability, provider, executor, trigger, role,
      and agent maturity remain orthogonal.
- [ ] Provider references stay opaque and persistence is never implicit.
- [ ] Workspace utilities contain no plan approval or lifecycle coupling.
- [ ] No hard skill dependency or umbrella workflow owner remains.
- [ ] No compatibility surface for the former commands exists.

## Files modified/created

- Delete: `packages/command/command-workflow/commands/*.md` at the parent-plan
  baseline.
- Create: the 12 command files listed in task 1.
- Modify: `packages/command/command-workflow/package.json`.
- Modify: `packages/command/command-workflow/moon.yml`.
- Verify only: `packages/command/command-workflow/.claude-plugin/plugin.json`.
- Verify only: `packages/command/command-workflow/.codex-plugin/plugin.json`.

## Dependencies

- No child-plan dependency.
- Can run in parallel with subplan 04 because their implementation files do not
  overlap.
- Subplans 02, 03, and 05 consume this command contract.

## Estimated duration

One focused implementation session. Most work is prompt-contract design and
inventory replacement; validation changes are intentionally deferred.
