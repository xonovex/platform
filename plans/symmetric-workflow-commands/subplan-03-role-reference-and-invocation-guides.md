---
type: plan
has_subplans: false
parent_plan: ../symmetric-workflow-commands.md
parallel_group: 3
status: complete
updated: 2026-07-20
completed_date: "2026-07-20"
dependencies:
  plans:
    - subplan-01-command-contract-and-inventory.md
    - subplan-02-plan-skill-decoupling.md
    - subplan-04-external-trigger-boundary.md
  files:
    - packages/command/command-workflow/commands/create.md
    - packages/command/command-workflow/commands/review.md
    - packages/command/command-workflow/commands/revise.md
    - packages/command/command-workflow/commands/decide.md
    - packages/command/command-workflow/commands/execute.md
    - packages/command/command-workflow/commands/validate.md
    - packages/command/command-workflow/commands/publish.md
    - packages/command/command-workflow/commands/abandon.md
    - packages/skill/skill-plan/plan-guide/SKILL.md
    - packages/agent/agent-operator-go/README.md
skills_to_consult:
  - command-guide
  - orthogonal-pattern-guide
  - plan-guide
  - testing-guide
  - shell-scripting-guide
validation:
  type_check: not_applicable
  lint: passed
  build: passed
  tests: passed
  integration: passed
---

# Subplan 03: Role, reference, and invocation guides

## Objective

Replace the four prescriptive role workflows and diagrams with concise
documentation of one composable operation model. Explain role lenses,
provider-native references, and invocation mechanisms without defining gates,
authority, profiles, maturity policy, or a universal stage order.

This subplan owns published command documentation and the workflow-diagram
package. It consumes the finished command and plan-skill contracts and does not
change their behavior.

## Tasks

### 1. Rewrite the command package README around the dimensions

- Replace `packages/command/command-workflow/README.md` with an inventory derived
  from the eight core operations and four workspace utilities.
- Explain operation, kind, perspective, method, executor, agent capability,
  trigger, provider, and reference as independent axes.
- State that the examples are compositions, not a required workflow.
- Explain that a selected capability must be installed/available and failures
  are explicit; do not restore hard dependencies or an umbrella skill.
- Link to the three focused guides created below and to the single replacement
  diagram.

### 2. Replace role quickstarts with one role-lens guide

- Delete `docs/developer-quickstart.md`, `docs/pm-quickstart.md`,
  `docs/qa-quickstart.md`, and `docs/ux-quickstart.md`.
- Create `docs/role-lenses.md` with illustrative, non-gating compositions:

```text
PM/PO:             create -> review -> revise -> decide
UX:                create -> review -> revise -> decide
Developer:         create -> review -> revise -> execute -> validate -> publish
QA:                create -> review -> execute -> validate -> publish
Developer reviewer: review -> publish
```

- Clarify that the noun/kind and perspective change by role, while command
  semantics do not.
- State explicitly that QA validates behavior and evidence; a separate
  developer reviews and approves a pull request.
- Avoid permissions matrices, mandatory handoffs, authority claims, or
  role-specific command aliases.

### 3. Document provider-native references

- Create `docs/references.md`.
- Define a reference as an opaque locator interpreted by a selected provider,
  with an optional provider-native revision.
- Document unambiguous inference, explicit provider selection for ambiguous
  locators, inline results without persistence, and destination references
  returned by publish/copy operations.
- Include provider-shaped examples only as examples; do not normalize them:

```text
local: plans/example.md
git: repository URL + commit or branch
github: owner/repository issue or pull-request locator
tracker: provider-native item key
```

- State that there is no Xonovex ID, central resolver/store, or silent local
  fallback.

### 4. Document invocation and execution independently

- Create `docs/invocation.md`.
- Show that manual calls, harness hooks, CI/CD hooks, webhooks, and schedulers
  invoke the same command contract.
- Show that a human, deterministic script, LLM, or agent can execute the call
  without altering operation semantics.
- Treat A1/A2/A3 as optional harness/provider executor metadata only. Do not
  define the levels, pass them as command modes, or derive permissions/stages
  from them.
- Point Kubernetes callers to direct `AgentRun` creation after subplan 04,
  without presenting the operator as a workflow runtime.

### 5. Collapse four role diagrams into one operation diagram

- Delete `developer-workflow.dot/.png`, `pm-workflow.dot/.png`,
  `qa-workflow.dot/.png`, and `ux-workflow.dot/.png` under
  `packages/diagram/diagram-agent-workflow`.
- Create `operation-model.dot` and generated `operation-model.png` showing:
  - the eight sibling operations;
  - orthogonal selection inputs around them;
  - independent invocation/executor inputs;
  - provider-native result/reference output;
  - the four workspace utilities outside the operation flow.
- Do not draw a mandatory left-to-right lifecycle. Connections should mean
  composition possibilities, not gates.
- Update `packages/diagram/diagram-agent-workflow/moon.yml:3-47` to build only
  the new diagram.

### 6. Verify published documentation as an unopinionated model

- Resolve every internal documentation and image link.
- Derive all command lists/counts from the actual 12-file inventory or avoid
  claims that can drift.
- Search prose and DOT sources for former commands, profiles, gates, authority,
  maturity enforcement, QA PR ownership, and provider-neutral reference claims.
- Render the PNG with the existing Graphviz Moon target and inspect that labels
  are readable.

## Validation steps

1. `npx moon run command-workflow:ci-check`
2. `npx moon run diagram-agent-workflow:ci-check`
3. `npx moon run command-workflow:cross-package-links`
4. `rg -n -- '--profile|mandatory.*gate|authority-reference|governed|approval.*required|QA.*approve.*PR|AgentTrigger|AgentSchedule' packages/command/command-workflow/README.md packages/command/command-workflow/docs packages/diagram/diagram-agent-workflow --glob '!*.png'` must return no behavioral residue.
5. Verify the role guide contains all five lenses and only the developer-reviewer
   lens owns PR review/approval.
6. `git diff --check`

## Success criteria

- [x] One README, one role-lens guide, one reference guide, and one invocation
      guide explain the complete public model.
- [x] Every role composes the same operation vocabulary.
- [x] QA validation and developer PR review are separate responsibilities.
- [x] Provider-native locators remain opaque; no universal reference layer is
      implied.
- [x] Trigger, executor, agent maturity, and role do not change command
      semantics.
- [x] One non-lifecycle diagram replaces all four role workflows and renders
      successfully.
- [x] All links and documentation validation checks pass.

## Files modified/created

- Modify: `packages/command/command-workflow/README.md`.
- Delete: `packages/command/command-workflow/docs/*-quickstart.md`.
- Create: `packages/command/command-workflow/docs/role-lenses.md`.
- Create: `packages/command/command-workflow/docs/references.md`.
- Create: `packages/command/command-workflow/docs/invocation.md`.
- Delete: the eight role-specific `.dot`/`.png` diagram files.
- Create: `packages/diagram/diagram-agent-workflow/operation-model.dot`.
- Create: `packages/diagram/diagram-agent-workflow/operation-model.png`.
- Modify: `packages/diagram/diagram-agent-workflow/moon.yml`.
- Modify as needed: `packages/diagram/diagram-agent-workflow/package.json`.

## Dependencies

- Requires subplan 01's final command names and semantic contracts.
- Requires subplan 02's final plan-skill boundary for examples and links.
- Must finish before subplan 05 locks documentation inventory and release
  descriptions.

## Estimated duration

One focused implementation session, including Graphviz regeneration and visual
inspection.
