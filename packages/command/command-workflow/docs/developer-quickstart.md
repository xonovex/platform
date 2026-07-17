# Developer Quickstart

The everyday path through the workflow commands for a developer building a normal feature.
The [README](../README.md) documents all 59 commands and their contracts; this guide is the
subset you actually run, in order. See the [developer-workflow
diagram](../../../diagram/diagram-agent-workflow/developer-workflow.png) for the same flow at a
glance.

## How to invoke

Each command is a slash command that loads a guideline skill and runs one lifecycle stage:

```
/xonovex-workflow:plan-create
```

You will not touch most of the 59. Two things make the set learnable.

**Stages × verbs.** Commands are a handful of lifecycle stages crossed with a small verb set:

| Verb                  | Meaning                                                   |
| --------------------- | --------------------------------------------------------- |
| `-research`           | read-only investigation before anything exists            |
| `-create`             | produce the first version of a stage's result             |
| `-critique`           | independently stress-test it (run in a **fresh** session) |
| `-revise`             | apply feedback, publish a **new** revision                |
| `-accept` / `-reject` | an authority-bound decision against one exact revision    |
| `-run`                | execute the stage (develop, review, qa, release, …)       |
| `-continue`           | resume execution and complete one unit of work            |
| `-validate`           | check criteria are met, read-only, no mutation            |

Learn the shape once and `decision-{create,critique,revise,accept}`,
`plan-{create,critique,revise,accept}`, and the design families all read the same.

**Native references.** Each stage publishes a durable result (a plan doc, a PR, a report) that
later stages point at by an opaque handle — its id or path, returned by the previous command.
That handle is the resume mechanism: close your session and the next command reconstructs state
from the stored result, not from conversation memory.

## The happy path

Four stages: **Understand → Plan → Build → Ship.** Only Build writes code.

### 1 · Understand — optional

Skip this for anything you can describe in one sentence. Reach for it on fuzzy or greenfield
work where the requirements aren't settled:

```
/xonovex-workflow:research-run <question>       # durable, cited evidence report
/xonovex-workflow:formulation-run <subject>     # candidate requirements + examples
```

For a genuine fork (a choice you'd want recorded), add
`decision-create → decision-critique → decision-revise → decision-accept`. UI work can add the
`experience-design-*` family; non-trivial architecture the `solution-design-*` family. All of
these are optional and share the same `create → critique → revise → accept` shape.

### 2 · Plan

```
/xonovex-workflow:plan-research <requirements>   # read-only: codebase + web, versions, locations
/xonovex-workflow:plan-create                    # high-level plan → status: pending-approval
/xonovex-workflow:plan-critique <plan-ref>       # adversarial — run in a FRESH session
/xonovex-workflow:plan-revise <plan-ref>         # apply the findings → new revision
/xonovex-workflow:plan-accept <plan-ref>         # human gate — sets status: approved
/xonovex-workflow:plan-subplans-create <plan-ref> # detailed child plans + parallel groups
```

Key rules:

- **`plan-critique` runs in a fresh, independent session** — the session that wrote the plan
  defends it instead of attacking it.
- **`plan-critique → plan-revise` loops.** Re-critique the new revision until the plan is solid.
  Straightforward plans skip critique/revise entirely — at the cost of the adversarial safety net.
- **`plan-accept` is a mandatory human gate** and is **required before `plan-subplans-create`**
  (which needs `status: approved`). A model can summarize but cannot fabricate the approval, and
  approval of revision N never carries to N+1 — every `plan-revise` resets status to
  pending-approval, so re-run `plan-accept`. `plan-reject <plan-ref> --reason ...` is the
  negative branch and preserves the plan's history.
- Keep each subplan to ~5–7 tasks; larger ones risk silently dropping work.

### 3 · Build — the only stage that writes code

**Default (sequential):**

```
/xonovex-workflow:plan-continue                  # completes exactly ONE subplan, then stops
```

`plan-continue` reloads the plan and its required skills, baselines the toolchain
(typecheck/lint/build/tests), implements one child plan, and publishes its status. It
**deliberately stops after one subplan and never auto-chains** — run it again for the next one.

**Parallel / formal alternative:**

```
/xonovex-workflow:develop-run <plan-ref...> [--max-concurrency N]
/xonovex-workflow:develop-consolidate <dev-ref...> --target-workspace <ref>
```

`develop-run` executes planned assignments and publishes one Development result each, picking the
least-adaptive suitable executor; `develop-consolidate` merges parallel results into one
workspace and re-runs the full validation. Use this track only when you fanned out parallel work;
otherwise `plan-continue` is simpler. `develop-abandon` cleanly stops an assignment you're giving
up, preserving the partial work and reason.

**Check the work:**

```
/xonovex-workflow:plan-validate <plan-ref>       # read-only PASS/FAIL/WARN per criterion
/xonovex-workflow:plan-update <plan-ref>         # optional: record status + evidence in the plan
```

`plan-validate` reads the actual success criteria and Definition of Done — green tests alone don't
satisfy it, and it never changes plan status.

### 4 · Ship

```
/xonovex-workflow:git-commit                     # run repeatedly through the build
/xonovex-workflow:pr-create                      # open the PR/MR from the branch
```

- **`git-commit` stages everything (`git add -A`) and auto-infers the message and type from the
  diff** — on large or mixed-intent changes that guess is often wrong (it tends to mislabel as
  `refactor`). Use `--dry-run` or `--interactive` to review first, and `--type` to correct it.
  Per repo policy, **don't `--push` unless asked.**
- **`pr-create` detects the host from the git remote and stops if the matching host skill
  (`github-guide` / `gitlab-guide`) isn't installed.** It drafts the description from the real
  diff, adds CODEOWNERS for the changed paths as reviewers, and confirms before creating unless
  you pass `--yes`.

Optional AI review pipeline on the PR (one shared findings set flows through the four):

```
/xonovex-workflow:pr-review-analyze              # diff → labelled findings (nothing posted)
/xonovex-workflow:pr-review-refine               # tidy findings; --final marks them ready
/xonovex-workflow:pr-review-post                 # publish to the PR (blocking review for blockers)
/xonovex-workflow:pr-review-resolve              # resolve threads once genuinely fixed
```

Findings live in the session unless you `--out` / `--findings` a file — persist them to cross
sessions. For a normal solo feature this pipeline is optional; a human reviewer is the common path.

## What you can skip

- **Understand** — for well-scoped changes, jump straight to `plan-research`.
- **The critique/revise loops** — for a plan (or design/decision) you accept on first draft.
- **`develop-consolidate`** — only when you ran parallel `develop-run` assignments.
- **The git-worktree family** (`plan-worktree-{create,merge,abandon,cleanup}`) — only if you
  develop in isolated sibling worktrees instead of plain branches.
- **The governed/regulated tail** (below) — CI, ops, or a reviewer usually owns it.
- **Everything** — for a one-line change, implement directly and `git-commit`.

## Governed extension — skip unless regulated

Regulated or ops-owned flows continue past the PR with the abstract lifecycle equivalents, which
produce provider-native evidence and enforce human authority and external gates:

```
deliver-publish → review-run + qa-run
              → acceptance-validate → acceptance-decide   (authenticated human signs off)
              → integration-validate → integration-run    (protected external gate)
              → transition-run / release-run → observe-run
```

`review-run` and `qa-run` are the formal counterparts of the `pr-review-*` pipeline and produce
findings/test evidence bound to an exact revision. `assessment-run` and `inventory-generate` cover
compliance and SBOM/BOM needs. Acceptance is deliberately split — an agent may assemble evidence
(`acceptance-validate`), but only an authenticated human records the decision (`acceptance-decide`).
Integration is split the same way, and re-checks every binding at a non-bypassable external
enforcement point; ordinary tool access is not authorization. When something breaks:
`incident-run → corrective-action-run`. `retirement-run` decommissions resources much later.

The `workflow-inspect`, `workflow-governance-inspect`, `workflow-conformance`, `workflow-drift`,
and `workflow-modules` commands are out-of-band diagnostics you can run at any time.

## Jargon decoder

The command help is written abstractly so it can back any storage backend. Translated:

| Term             | What it means for you                                                                                                                                                       |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| native reference | an opaque handle/id to a stored result; pass what the previous command gave you                                                                                             |
| `--revision`     | pick one exact version of that result                                                                                                                                       |
| provider         | the storage backend for results, chosen by the profile (not hardcoded); commonly a local Markdown file under `plans/`. Git is optional and a separate workspace concern     |
| `--profile`      | a preset deciding which stages are required vs optional and which policies apply                                                                                            |
| `--method`       | the technique (user stories, BDD, example mapping, …); `neutral` is always the default                                                                                      |
| `--executor`     | who does the work — `deterministic`, `model`, `agent`, `human`, or `external`; a requested ceiling, never an escalation, and may be rejected (see `agent-governance-guide`) |

## See also

- [README](../README.md) — the full command table
- [PM Quickstart](pm-quickstart.md) · [QA Quickstart](qa-quickstart.md) ·
  [UX Quickstart](ux-quickstart.md) — the same actor-neutral commands read from the other seats
- [Architecture and composition](architecture-and-composition.md) — the two-planes model
- [`workflow-guide`](../../../skill/skill-workflow/workflow-guide/SKILL.md) and
  [`plan-guide`](../../../skill/skill-plan/plan-guide/SKILL.md) — the behavior each command runs
