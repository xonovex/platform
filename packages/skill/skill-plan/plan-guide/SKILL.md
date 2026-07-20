---
name: plan-guide
description: "Use when discovering, researching, formulating, designing, deciding, planning, revising, critiquing, continuing, or validating work — covers neutral early-lifecycle results, optional experience/solution design, authority-bound decisions, provider-native planning, and code research. Triggers on problem discovery, evidence and uncertainty, requirements formulation, UX or architecture results, decision briefs, plans/subplans, fresh-context resume, code cleanup/hardening/alignment, even when the user doesn't say 'plan'."
compatibility: "Node.js 22+ is required only for the bundled early-lifecycle fixture validator, which reads repository fixtures without network access."
allowed-tools: "Read Bash(node:*)"
---

# Early Lifecycle, Planning, and Code-Research Guidelines

Produce Discovery, Research, Formulation, optional Design, Decision, and Planning results without binding their meaning to one method, executor, provider, policy runtime, file layout, or agent session. Also run code-research operations that feed those results.

## Core Principles

- **Neutral semantics** — canonical result meaning stays stable while methods, artifacts, executors, providers, profiles, and optional runtime composition vary independently
- **Native persistence** — consume and publish opaque provider-native references; local files and Git are selectable providers, not prerequisites
- **Least-adaptive execution** — prefer deterministic collection, bound model synthesis, reserve agents for branching exploration, and preserve human/qualified authority
- **Research first** — distinguish evidence, provenance, uncertainty, and synthesis before authoring or deciding
- **Reports, not code** — all research/analysis operations are read-only and generate reports for `plan-create` to consume; only `plan-continue` modifies the codebase
- **Validation required** — every plan's success criteria back-checks with typecheck / lint / build / tests
- **Skills to consult** — every plan lists which guideline skills implementers must read first
- **Use available tools** — prefer environment-provided task tracking, file search, and code analysis over working from memory

## Early Lifecycle Operations

- **Discovery run** — iterate observations, assumptions, affected context, and unknowns — see [references/discovery-run.md](references/discovery-run.md)
- **Research run** — collect and synthesize evidence with provenance and uncertainty — see [references/research-run.md](references/research-run.md)
- **Formulation run** — express candidate behavior, examples, constraints, and ambiguities — see [references/formulation-run.md](references/formulation-run.md)
- **Experience Design** — create, independently [critique](references/experience-design-critique.md), [revise](references/experience-design-revise.md), and [accept](references/experience-design-accept.md) optional experience results — see [references/experience-design-create.md](references/experience-design-create.md)
- **Solution Design** — create, independently [critique](references/solution-design-critique.md), [revise](references/solution-design-revise.md), and [accept](references/solution-design-accept.md) optional solution results — see [references/solution-design-create.md](references/solution-design-create.md)
- **Decision** — [create](references/decision-create.md), independently [critique](references/decision-critique.md), [revise](references/decision-revise.md), and [accept](references/decision-accept.md) authority-bound results
- **Onboarding advice** — recommend workflow methods, skills, providers, and executors without applying changes — see [references/lifecycle-onboard-advise.md](references/lifecycle-onboard-advise.md)

## Gotchas

- A method such as user stories or BDD is selectable — making it the entry contract excludes valid workflows and creates a hard dependency
- An unavailable explicitly selected provider is an error — silently writing a local file changes the user's source of truth
- A model can prepare a decision brief but cannot fabricate stakeholder statements, human acceptance, professional qualification, or delegated authority
- Skipping Research and going straight to Planning produces vague plans built on assumed context
- A plan without `skills_to_consult` leaves implementers ignoring project conventions
- Auto-detecting toolchain via `package.json` only misses Moon/Makefile-driven projects — check both
- Approve the parent with `plan-accept` before `plan-subplans-create` — it requires `status: approved`
- `plan-critique` must run as an independent agent (fresh session), not the plan's author — self-critique defends instead of attacks
- A change describable in one sentence (a one-line diff) skips discovery and planning — implement it directly; heavy up-front spec on a trivial edit is waste
- Workflow authority requirements remain explicit result semantics; an executor or maturity label cannot silently satisfy them
- A plan is done when the team's Definition of Done is met (review, docs, no regressions, NFRs), not when tests merely pass — `plan-validate` checks the DoD, not just success criteria
- "Tests pass" doesn't mean "success criteria met" — `plan-validate` reads the criteria, not just exit codes
- Auto-continuing to the next plan after completion silently chains work — `plan-continue` STOPS after one
- Subplans with >7 tasks risk silent drops — target 5–7 tasks each
- Skipping the verification re-read before marking complete is the #1 cause of incomplete work

## Plan Lifecycle

1. **Create** — `plan-create` resolves Research, Decision, Formulation, and Design native references and publishes the parent Planning result; selected test/acceptance methods load matching installed skills
2. **Revise** — `plan-revise` applies explicit feedback to an exact Planning revision and publishes a new native revision
3. **Critique** — `plan-critique` independently stress-tests an exact Planning revision, feeding a separate result into revise
4. **Accept / Reject** — `plan-accept` or `plan-reject` records an authorized status decision against the exact revision
5. **Expand** — `plan-subplans-create` publishes detailed child Planning results and provider-native relationships
6. **Execute** — `plan-continue` reconstructs native handles and works through one child Planning result at a time
7. **Update** — `plan-update` publishes current status and validation results as a new revision
8. **Validate** — `plan-validate` checks success criteria and Definition of Done against exact revisions and evidence

## Progressive Disclosure

### Early lifecycle

- Read [references/early-lifecycle-contracts.md](references/early-lifecycle-contracts.md) - Load when selecting methods/executors/providers, handling authority, publishing native results, or resuming after context loss
- Read [references/discovery-run.md](references/discovery-run.md) - Load when discovering a problem/opportunity through neutral iterative observations, assumptions, and unknowns
- Read [references/research-run.md](references/research-run.md) - Load when producing reusable evidence with provenance, confidence, uncertainty, limitations, and bounded exploration
- Read [references/formulation-run.md](references/formulation-run.md) - Load when formulating candidate behavior, examples, constraints, and ambiguities without mandating stories or Gherkin
- Read [references/experience-design-create.md](references/experience-design-create.md) - Load when creating an optional Experience Design result
- Read [references/experience-design-critique.md](references/experience-design-critique.md) - Load when independently critiquing an exact Experience Design revision
- Read [references/experience-design-revise.md](references/experience-design-revise.md) - Load when revising Experience Design from explicit feedback or critique references
- Read [references/experience-design-accept.md](references/experience-design-accept.md) - Load when recording an authorized decision on an exact Experience Design revision
- Read [references/solution-design-create.md](references/solution-design-create.md) - Load when creating an optional Solution Design result
- Read [references/solution-design-critique.md](references/solution-design-critique.md) - Load when independently critiquing an exact Solution Design revision
- Read [references/solution-design-revise.md](references/solution-design-revise.md) - Load when revising Solution Design from explicit feedback or critique references
- Read [references/solution-design-accept.md](references/solution-design-accept.md) - Load when recording an authorized decision on an exact Solution Design revision
- Read [references/decision-create.md](references/decision-create.md) - Load when preparing and recording one evidence-grounded, authority-bound Decision result
- Read [references/decision-critique.md](references/decision-critique.md) - Load when independently critiquing a Decision without acting as its authority
- Read [references/decision-revise.md](references/decision-revise.md) - Load when revising an exact Decision while preserving evidence, authority, and supersession
- Read [references/decision-accept.md](references/decision-accept.md) - Load when an authorized actor records a decision against an exact native revision
- Read [references/lifecycle-onboard-advise.md](references/lifecycle-onboard-advise.md) - Load when recommending workflow methods, skills, providers, and executors without applying changes

### Research

- Read [references/plan-research.md](references/plan-research.md) - Load when researching codebase + web for a future plan, or running a read-only code-quality audit (harden / simplify / align — applies code-quality-guide)
- Read [references/code-barrels-remove.md](references/code-barrels-remove.md) - Load when analyzing barrel exports for removal
- Read [references/code-comments-remove.md](references/code-comments-remove.md) - Load when identifying non-essential comments
- Read [references/code-shared-extract.md](references/code-shared-extract.md) - Load when finding duplicated patterns to extract
- Read [references/code-template-extract.md](references/code-template-extract.md) - Load when creating reusable templates from existing code
- Read [references/code-template-scaffold.md](references/code-template-scaffold.md) - Load when generating new code from templates
- Read [references/todos.md](references/todos.md) - Load when scanning and grouping TODO comments

### Plan lifecycle

- Read [references/plan-create.md](references/plan-create.md) - Load when publishing a high-level Planning result from opaque native lifecycle references
- Read [references/plan-revise.md](references/plan-revise.md) - Load when applying native feedback, critique references, annotations, or prompt instructions to an exact Planning revision
- Read [references/plan-critique.md](references/plan-critique.md) - Load when adversarially stress-testing a plan to expose weaknesses (red-team / pre-mortem / falsify / steelman), read-only
- Read [references/plan-accept.md](references/plan-accept.md) - Load when approving a plan for execution (sanity-check, then set status: approved)
- Read [references/plan-reject.md](references/plan-reject.md) - Load when rejecting a plan with a reason (set status: rejected, record why, keep the plan)
- Read [references/plan-subplans-create.md](references/plan-subplans-create.md) - Load when expanding an approved plan into detailed parallelizable subplans
- Read [references/plan-continue.md](references/plan-continue.md) - Load when resuming implementation work from an existing plan
- Read [references/plan-update.md](references/plan-update.md) - Load when refreshing a plan with current status / validation / progress
- Read [references/plan-validate.md](references/plan-validate.md) - Load when verifying a plan's success criteria are met (read-only)
