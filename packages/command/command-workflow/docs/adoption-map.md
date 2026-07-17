# Adoption Map

This platform is adopted by **composition, not progression**. There is no ladder from
hobbyist to enterprise: a solo developer protecting their dotfiles with a harness hook and a
regulated team running admission-controlled agents use the same pieces and the same
selection logic — they just answer the same questions differently, control by control. Pick
any subset; every piece states what it gives you and what its absence means. Nothing here
requires the Kubernetes agent operator; runtime admission is one optional enforcement type,
never a prerequisite.

The sections are ordered so each uses only vocabulary the previous ones introduced — **read
in order to learn; adopt in any order to use**. The one genuinely ordered thing in the
platform is the [autonomy ladder](#the-one-real-ladder-autonomy), and it orders oversight,
not adoption.

The [adoption-map diagram](../../../diagram/diagram-agent-workflow/adoption-map.png) shows
the two tracks and the per-control choice at a glance; the
[enforcement-types diagram](../../../diagram/diagram-agent-workflow/enforcement-types.png)
shows how one policy decision reaches every enforcement type.

Find your section:

| You want to…                                                                 | Section                                       |
| ---------------------------------------------------------------------------- | --------------------------------------------- |
| Get better code out of your agent, nothing else                              | [Guideline skills](#guideline-skills)         |
| Run a structured lifecycle (plan → build → ship) with durable results        | [Workflow commands](#workflow-commands)       |
| Tune which stages are required, which methods and storage backends are used  | [Profiles and composition](#profiles)         |
| Add governance semantics — policy, executors, autonomy — without enforcement | [Governance contracts](#governance-contracts) |
| Have the agent harness itself block denied actions                           | [Harness hooks](#harness-hooks)               |
| Enforce independently of any agent, in CI and provider controls              | [External enforcement](#external-enforcement) |
| Run agents as admission-controlled jobs (optional)                           | [Runtime admission](#runtime-admission)       |
| Decide which enforcement each control needs                                  | [Choosing per control](#choosing-per-control) |
| See what a composition looks like for a team like yours                      | [Worked compositions](#worked-compositions)   |
| See every independent choice — all the axes — on one page                    | [The axes at a glance](#axes)                 |

## The pieces

| Piece                 | What it is                                                                                      | Where                                                    |
| --------------------- | ----------------------------------------------------------------------------------------------- | -------------------------------------------------------- |
| Guideline skills      | Coding/craft knowledge an agent loads on trigger (`c99-guide`, `zod-guide`, `testing-guide`, …) | `packages/skill/skill-*`                                 |
| Lifecycle skills      | The two contract owners: `workflow-guide` and `agent-governance-guide`, plus `plan-guide`       | `skill-workflow`, `skill-agent-governance`, `skill-plan` |
| Harness skills        | One capability matrix + onboarding guide per agent product (Claude Code, codex, copilot, …)     | `skill-claude-code` and siblings                         |
| Platform skills       | External-control owners (GitHub, GitLab, Azure DevOps, Bitbucket, Bitrise, AWS, Datadog)        | `skill-github` and siblings                              |
| Commands              | 59 slash commands (`/xonovex-workflow:<stage>-<verb>`) that load a skill and run one operation  | `packages/command/command-workflow`                      |
| Fixtures + validators | Machine-checked conformance for every contract, run in CI                                       | each guide's `assets/` + `scripts/`                      |

Everything installs through your harness's plugin mechanism; the marketplace manifest lists
every skill and command plugin, and versions move in lockstep.

## The model in five rules

The whole platform follows five rules; every section below is an application of them. The
full statement is [architecture-and-composition.md](architecture-and-composition.md).

1. **Two independent planes.** The _workflow_ plane owns lifecycle capabilities, results,
   profiles, and completion. The _governance_ plane owns applicability, policy decisions,
   authority, executors, events, enforcement adapters, and drift. They integrate through
   semantic references; neither requires the other.
2. **Decision before enforcement.** A policy decision point returns a versioned
   allow/deny/ask outcome. A separate enforcement point applies it and records its own
   evidence. No part of the architecture embeds the only copy of policy meaning.
3. **Provider-native results.** Every result is a semantic type plus an opaque native
   reference and revision token — a Markdown file, a PR, a work item, a CI run, a database
   row are peer providers. There is no central workflow database, no universal envelope, and
   no silent fallback to local files when a hosted provider was selected.
4. **Least-adaptive execution.** Deterministic tools establish authoritative facts; bounded
   models do narrow transforms; bounded agents do genuinely adaptive work; humans hold
   accountable judgment; external systems own the facts they already own. A model never
   silently replaces an authoritative check.
5. **Honesty over coverage.** Installing a skill or module is not evidence a control runs.
   Documentation-verified support is never reported as runtime conformance. Every adoption
   mode states its expected absences. Where a capability is missing, that is a named gap, not
   a best-effort match.

## The workflow track

Three pick-up points, each a complete, honest stopping point. Only the last two are
related — profiles shape the commands; guideline skills are independent of both.

<a id="guideline-skills"></a>

### Guideline skills only

Install any subset of the guideline skills and stop. Each skill triggers on its own file
types and vocabulary and teaches your agent the house style for one concern — C99,
TypeScript, Vitest, Kubernetes manifests, code review, test doubles, planning discipline.
Skills are composable by construction: one concept has exactly one owner skill, and
specific skills cross-reference the general ones by name instead of copying, so installing
ten skills never gives you ten conflicting copies of one rule.

There is no lifecycle, no result storage, no policy, and no enforcement in this piece.

**What you get:** better, more consistent agent output per file type.
**Expected absence:** no workflow results, no gates, no governance or enforcement claims of
any kind.

<a id="workflow-commands"></a>

### Workflow commands, workflow-only

Install the `xonovex-workflow` command plugin (which loads `workflow-guide` and `plan-guide`)
and run the lifecycle as slash commands. This is the **workflow-only adoption mode**: full
structured delivery with no harness policy and no enforcement guarantee claimed.

The 59 commands are a small set of lifecycle stages crossed with a small verb set —
`-research` (read-only), `-create`, `-critique` (fresh session), `-revise`, `-accept` /
`-reject` (authority-bound, one exact revision), `-run`, `-continue` (one unit, then stop),
`-validate` (read-only). Learn the shape once; `plan-*`, `decision-*`, and the design
families all read the same. The everyday subset, in order, is the
[developer quickstart](developer-quickstart.md); PM, QA, and UX seats have their own.

Two mechanics matter more than any individual command:

- **Native references are the resume mechanism.** Every stage publishes a durable result (a
  plan document, a PR, a report) and later stages point at it by the opaque handle the
  previous command returned. Close the session; the next command reconstructs state from the
  stored result, not from conversation memory. A runtime trace or session ID is never a
  workflow identity.
- **Gates are human and revision-exact.** `plan-accept` is a mandatory human gate; approval
  of revision N never carries to N+1, and a model can summarize evidence but cannot fabricate
  the approval. The same split repeats at every high-impact gate: an agent may _assemble_
  evidence (`acceptance-validate`), only an accountable human _decides_
  (`acceptance-decide`).

Results commonly land as Markdown under `plans/` — but that is one provider choice, not the
architecture. See the jargon decoder in the quickstart for `--profile`, `--method`,
`--executor`, and native references translated to plain terms.

**What you get:** a repeatable lifecycle with durable, revision-exact results and human
gates.
**Expected absence (stated by the mode itself):** no harness policy or external enforcement
guarantee is claimed. The gates bind behavior because you follow them, not because anything
blocks you.

<a id="profiles"></a>

### Profiles and composition

Stay on the same commands but take control of _which_ stages are required, in what order,
with which methods and storage backends. This piece is the composition grammar, owned by
[`workflow-guide` profiles](../../../skill/skill-workflow/workflow-guide/references/profiles.md).

**Independent axes, resolved separately.** Workflow profile, method, artifact provider,
work-item provider, code-host provider, workspace provider, policy, and learning policy each
resolve in the same order: explicit command argument → selected workflow profile → project
instructions → unambiguous environment detection → axis default → otherwise fail visibly and
ask. An explicitly selected but unavailable provider is an error, never a silent fallback.

**A profile declares semantics, not a config format:** identity, version, owner, scope;
included capabilities and preserved result kinds; allowed sequence, concurrency, iteration,
and composite presentation; publication boundaries, exit rules, completion evidence; per-axis
requirements; actor and independence requirements; the guarantee an enforcement point must
provide; and failure behavior when a dependency is unavailable. The default software-delivery
presentation (Discover/Plan/Develop/Deliver/Review/Accept, with Integrate, Release, Observe
as explicit extensions) is one presentation, not the ontology — a profile may split
composites, run Review and QA concurrently, loop, or stop after an accepted deliverable.

**Composition is strengthening-only.** Composing profiles unions required evidence,
authority, and guarantees; incompatible requirements fail visibly (never last-writer-wins);
weakening requires an authorized, scoped, expiring exception; and a mandatory control with no
enforcement point that can guarantee it is rejected outright. Organization-managed
requirements cannot be silently weakened by project, user, or session configuration.

Inspect the effective composition at any time with `workflow-inspect` (resolved capabilities,
topology, freshness, gaps) and validate it with `workflow-conformance`. Worked profile
compositions ship today as conformance fixtures in the guides' `assets/`; author your own
against the profile contract and validate it the same way.

**What you get:** the lifecycle shaped to your team — solo and lightweight through regulated
and evidence-heavy — with conflicts surfaced instead of absorbed.
**Expected absence:** declaring a requirement does not enforce it; enforcement is the
governance track's per-control choice.

## The governance track

<a id="governance-contracts"></a>

### Governance contracts — semantics without enforcement

Install [`agent-governance-guide`](../../../skill/skill-agent-governance/agent-governance-guide/SKILL.md)
and adopt its semantics — **governance-only** (protecting ordinary agent activity without any
lifecycle commands) or **enablement-only** (transactional setup with no ongoing guarantee).
This piece supplies the vocabulary and discipline that the enforcement types below bind to
native controls.

**Executor classes.** Every task selects the least adaptive executor that fits:
deterministic script/API → script plus bounded model → bounded agent → human → external
system, each with a required boundary (pinned inputs, closed schemas, budgets, attenuated
tools, recorded identity, preserved native evidence). `--executor` on any command is a
requested ceiling, never an escalation, and may be rejected.

**Modules and transactional enablement.** Every installed governance module declares its
classification (from knowledge-only through advisory, evidence-producing, enforcing, and
configuration-changing to privileged), and all configuration change follows one transactional
recipe: discover → assess → recommend → **preview** the exact native change (identity,
checksum, permissions, failure mode, rollback) → **authorize** → apply idempotently →
**verify** by re-reading native state → monitor **drift** → roll back cleanly. Consent is a
real gate: without it, the operation previews and stops. Harness _module patterns_ are the
one place this track is ordered — by authority, not maturity: the module authority ladder
(knowledge-only → advisory hook → enforcing hook → script-plus-model → agent launcher →
organization-managed) with the rule _choose the lowest-authority composition that meets the
requirement_.

**The autonomy ladder.** [autonomy.md](../../../skill/skill-agent-governance/agent-governance-guide/references/autonomy.md)
grades a _posture_ — how far a run advances before a human must act — from `A0` (human
invokes everything) through `A1` (adds an independent critique the author cannot suppress),
`A2` (headless to the next human gate; requires a run journal, asynchronous exact-revision
approval, cancellation, and a kill switch) to `A3` (non-human triggers under admission
control; an eventual goal an adopter builds and proves, not a description of what exists) —
drawn in the
[autonomy-ladder diagram](../../../diagram/diagram-agent-workflow/autonomy-ladder.png).
**Start at `A1`.** Two rules govern the whole ladder: raising a level never widens what a
task may do — only who is watching and when — and **never raise autonomy without matching
oversight**: where a level's required oversight is absent, unverified, or degraded, that
level is simply unavailable, and degradation demotes you to the highest level whose oversight
still holds. Unattended escalations are bounded — window and safe default declared before
the run starts; silence is never approval.

**The walking skeleton.** Prove the whole composition locally before trusting any of it:
`assets/walking-skeleton/run-skeleton.sh` (in the governance guide) executes the full
onboarding lifecycle against a throwaway workspace — 17 checks covering consent, idempotent
apply, an allowed and a denied operation with reasons, drift detection and remediation,
tamper refusal, fail-closed policy outage, depth-limited agent launch, expired exceptions,
and clean rollback. It never touches your repository.

**What you get:** governance semantics you can inspect (`workflow-governance-inspect`),
validate, and drift-check — and a graded, honest answer to "how autonomous are we?"
**Expected absence (stated by the modes):** lifecycle results are not prerequisites
(governance-only), and there is no ongoing policy guarantee unless an enforcing module is
separately selected (enablement-only).

### Enforcement types — peers, chosen per control

Enforcement is not a sequence to climb. The platform has three **types** of enforcement,
and they are peers: independent mechanisms with different reach, selected per control. A
hobbyist and an enterprise choose from the same three; multiple types may back the same
control when its guarantee demands independence.

| Type                 | Who it binds                                | Typical mechanisms                                                                          |
| -------------------- | ------------------------------------------- | ------------------------------------------------------------------------------------------- |
| Harness hook         | the agent session it is configured in       | native hook events selected through a versioned capability matrix                           |
| External enforcement | every actor — human or agent, hooked or not | required CI checks, rulesets, protected environments, provider permissions, policy services |
| Runtime admission    | workloads entering a cluster                | validating/admission webhooks (`AgentPolicy` or any equivalent control)                     |

<a id="harness-hooks"></a>

#### Harness hooks — enforcement in the session

Bind governance intents to your agent harness's native hook mechanism so a denied action is
refused _in the session_, before it happens.

**Semantic intents, native capabilities.** Policy binds to twelve stable intent families —
session, prompt, model, tool, permission, capability, result, configuration, compaction,
subagent, workspace, and privileged operation
([events-and-capabilities.md](../../../skill/skill-agent-governance/agent-governance-guide/references/events-and-capabilities.md)) —
never to a native hook name. Each harness skill publishes a **versioned capability matrix**
recording, per intent: the native event, support state, whether the handler executes,
blocking vs advisory behavior, ordering and concurrency, configuration precedence, data
exposure, and a conformance probe. A mandatory intent is satisfiable only when support is
`supported`, the handler executes, coverage is complete, blocking is real, failure behavior
matches the profile, and lower scopes cannot weaken the configuration. Experimental or
unknown support can serve advisory profiles only.

**Worked example — Claude Code.** The
[capability matrix](../../../skill/skill-claude-code/code-harness-guide/references/capabilities.md)
maps _tool before use_ → `PreToolUse`, _prompt before submission_ → `UserPromptSubmit`,
_permission request_ → `PermissionRequest`, and so on. The guard contract is deliberately
minimal and deterministic: the hook handler receives one JSON event on stdin and answers
with exit `0` (allow) or exit `2` (deny) plus a JSON decision — the same decision-point shape
the walking skeleton's `guard.sh` implements. Note the matrix's own honesty discipline: it
records its observed runtime version and states exactly which rows were exercised at runtime
versus verified against documentation — and documentation-verified rows are never reported
as runtime conformance. Treat your own deployment the same way: probe, record the version and
date, and re-probe on the matrix's refresh trigger.

**Onboard hooks transactionally**, like any module: preview the exact settings change,
authorize, apply, verify with the harness's own diagnostics, monitor drift, roll back. Six
harness skills ship (Claude Code, codex, copilot, kiro, pi, opencode); each owns its native
mechanics.

**Know what a hook is not.** A hook is one enforcement point, not a security boundary: it is
not an independent sandbox, a denial does not roll back concurrently started sibling
handlers, and a user-scope hook can be reconfigured by the user it governs. That is why
external enforcement exists as its own type — for a **mandatory** control, the hook is early
feedback and an independent, non-bypassable external point is the authority.

**What you get:** immediate, in-session blocking of denied tool calls and prompts, with
separate decision and enforcement evidence per event.
**Expected absence:** no protection against actors who can edit the hook configuration or act
outside the harness; mandatory guarantees additionally require external enforcement.

<a id="external-enforcement"></a>

#### Pipelines and external enforcement — no agent runtime required

Enforce the same semantic policy decisions through controls that exist entirely outside any
agent — CI, repository rules, deployment approvals, provider permissions, and policy
services. This is the **external-enforcement-only adoption mode**: it works with zero harness
hooks and zero lifecycle commands installed, and it is where mandatory controls live.
Contract: [external-enforcement.md](../../../skill/skill-agent-governance/agent-governance-guide/references/external-enforcement.md).

**The intent → native-point mapping:**

| Semantic intent               | Suitable native points                                             |
| ----------------------------- | ------------------------------------------------------------------ |
| Validate an exact change      | Required CI job/check, merge gate, pipeline policy                 |
| Protect a repository target   | Ruleset, branch/tag protection, approval rule, provider permission |
| Authorize a deployment        | Protected environment, deployment approval, deployment policy      |
| Admit a runtime workload      | Admission webhook, validating policy, namespace policy             |
| Restrict a provider operation | Provider role/token scope, protected API, secret policy            |
| Obtain a policy decision      | Deterministic rules provider or policy decision service            |
| Assist locally                | Tracked Git hook / pre-commit — **advisory only**, never mandatory |

A mandatory control is conformant only when at least one selected point **cannot be bypassed
by the governed actor**, binds to the exact subject revision, fails with its declared
behavior (fail-closed for mandatory privileged operations), and produces independently
resolvable evidence. Every bypass actor and exception path is recorded; a control with an
undisclosed bypass is unsupported for mandatory use. Skipped, cancelled, or missing checks
are never coerced to success.

**Privileged operations** — Integration, Release, production deployment, secret
access/rotation, infrastructure mutation, data deletion, Retirement — carry the full bar:
exact revisions, authorized actor with segregation of duties, fresh prerequisite evidence,
least-privilege short-lived credentials released only after approval, a protected
target-side gate, digest/provenance verification for anything executed or deployed, explicit
rollback handling, and retained decision/enforcement/mutation/verification evidence.
Client-side hooks, labels, comments, chat approvals, or a passing harness hook alone never
authorize a privileged change.

**Policy decision services** are optional adapters, not an architecture: OPA is one
implementation. Send minimized canonical facts, record the decision reference and
policy/data version, cache only authenticated decisions with expiry, and fail closed on
outage for mandatory operations.

**Platform owners.** Concrete native mechanics — GitHub SHA-pinned reusable workflows and
rulesets, GitLab pipeline execution policies and compliance frameworks, Azure DevOps
branch policies and checks, Bitbucket, Bitrise, AWS IAM/Organizations/CloudTrail, Datadog
CI visibility — belong to the platform skills, catalogued with their capability boundaries in
[platform-onboarding.md](platform-onboarding.md). Mixed estates are first-class: Azure
Boards can own work items while Bitbucket owns source, Bitrise owns CI, AWS owns runtime
governance, and Datadog observes — correlated by exact native references, never coerced into
one identity.

**Independence is the point, not redundancy.** The walking skeleton proves it concretely: it
disables the harness hook and shows the CI-shaped required check still denying the protected
mutation. The hook gives the agent early, in-session feedback; the external control is the
independent authority the mandatory guarantee actually rests on — the flow the
[enforcement-types diagram](../../../diagram/diagram-agent-workflow/enforcement-types.png)
draws end to end.

**What you get:** enforcement that binds every actor — human or agent, hooked or not — at
exact revisions, with native evidence.
**Expected absence (stated by the mode):** agent hooks and lifecycle commands are optional;
nothing in this type observes or improves the _in-session_ behavior of an agent.

<a id="runtime-admission"></a>

#### Runtime admission — optional, for agent workloads on clusters

If — and only if — you run agents as isolated workloads on Kubernetes, the platform's
`AgentPolicy` admission adapter adds one more enforcement type: namespace-scoped admission
for `AgentRun` workloads covering runtime-class isolation, security-context hardening,
egress restriction, duration and resource bounds, image allowlists, and pinned toolchains,
failing closed on lookup failure.

This type is optional infrastructure, not an upper stage. It is one adapter in the
external-enforcement mapping's "admit a runtime workload" row — any equivalent admission or
validating policy control satisfies the same intent — and a homelab k3s cluster qualifies
just as much as an enterprise estate. Nothing else on this map depends on it, and reaching
`A3` autonomy is defined by _requirements_ (an enforced policy verdict at a non-bypassable
point, protected targets, escalation routing with accountable recipients, and per-run
provenance), not by this or any particular component. Where those requirements are met by
other controls, the operator adds nothing; where they are not met at all, `A3` is
unavailable — run at `A2` or below, which needs no cluster whatsoever.

<a id="choosing-per-control"></a>

## Choosing enforcement per control

The same three questions for every adopter, asked once per control — not once per
organization:

1. **Who must this control bind?** Only your own agent, in its session → a harness hook is
   enough. Every actor — including you, other humans, and unhooked agents → an external,
   non-bypassable point. Unattended workloads entering a cluster → admission.
2. **What failure behavior does it need?** Advisory (observe and report — honest for
   low-impact concerns) → any type, even experimental hook support. Mandatory → at least one
   selected point the governed actor cannot bypass, bound to the exact revision, failing
   closed, with resolvable evidence.
3. **What is the impact?** Low — a wrong outcome is detectable without a human and
   reversible → advisory is legitimate. Privileged — Integration, Release, deployment,
   secrets, data deletion, Retirement → the full external bar, whatever your team size.

Mix freely. "Protect `~/.ssh` from my agent" is a hook, for a hobbyist and an enterprise
alike. "Protect `main`" is branch protection, for both. "Release to production" carries the
privileged-operations bar wherever releasing to production matters — the bar follows the
impact, not the organization chart. A mandatory control may select several types at once;
their value is independence: one failing open does not silence the other.

<a id="worked-compositions"></a>

## Worked compositions

Team shape — solo, small-team, regulated — is an axis of the profile contract, not a level
of the platform. These are illustrations, not tiers:

- **Hobbyist / solo.** Guideline skills, a `PreToolUse` hook protecting dotfiles and
  secrets, branch protection on `main`. That is already two enforcement types on day one —
  enforcement was never the "advanced" part. Mode: governance-only (or skills-only), at
  `A0`–`A1`.
- **Small team.** Workflow commands with a profile fitting the team, required CI checks as
  the external authority, per-developer hooks for early feedback, `plan-accept` and
  `acceptance-decide` as the human gates. Mode: integrated (light), typically `A1`–`A2`.
- **Regulated.** Integrated mode with the privileged-operations bar: segregation of duties,
  protected environments, evidence retention, drift monitoring — and admission control where
  unattended agent workloads run. `A2` where oversight holds; `A3` only where its full
  requirements are met and observed.
- **Enterprise without governed agents.** External-enforcement-only — rulesets, pipeline
  policies, provider permissions, no harness hooks, no lifecycle commands. A complete,
  legitimate composition, not a partial adoption.

## The one real ladder: autonomy

Autonomy (`A0`–`A3`) is the platform's only genuine ladder, because its levels have real
prerequisites — the oversight coupling. It decides which pieces of this map are
load-bearing for you:

| Level     | What must already be in place (and observed to work)                                                              |
| --------- | ----------------------------------------------------------------------------------------------------------------- |
| `A0`–`A1` | Nothing on this map is load-bearing; every piece is optional craft                                                |
| `A2`      | Governance-contract oversight: run journal, asynchronous exact-revision approval, cancellation, a kill switch     |
| `A3`      | An enforced verdict at a non-bypassable external point, protected targets, escalation routing, per-run provenance |

Running unattended against advisory enforcement is precisely what the oversight coupling
forbids. Degraded oversight demotes the level; silence never advances a gate.

<a id="axes"></a>

## The axes at a glance

Every independent choice in the platform, in one place — drawn in the
[variation-axes diagram](../../../diagram/diagram-agent-workflow/variation-axes.png). Each
set is owned by its contract; this section is a view, not a second owner. Axes are
orthogonal: any value on one composes with any value on the others, and nothing here is a
sequence.

**Plane level** — the root orthogonality
([architecture-and-composition.md](architecture-and-composition.md)): **workflow plane ⊥
governance plane**. Adopt either, both, or neither; they integrate only through semantic
references.

**Workflow variation axes** — owned by
[`workflow-guide` architecture](../../../skill/skill-workflow/workflow-guide/references/architecture.md).
Every axis resolves independently through the same order: explicit argument → selected
profile → project instructions → unambiguous environment detection → axis default → fail
visibly and ask.

| Axis                                     | Varies                                                    |
| ---------------------------------------- | --------------------------------------------------------- |
| Workflow profile                         | required capabilities, topology, completion               |
| Method                                   | user stories, BDD, example mapping, … (`neutral` default) |
| Artifact provider                        | where deliverable results live                            |
| Work-item provider                       | where work items live                                     |
| Code-host provider                       | where source and PRs live                                 |
| Workspace provider                       | where working state lives                                 |
| Policy (quality / security / governance) | which rules apply                                         |
| Learning policy                          | how feedback and learning are promoted                    |

**Governance profile facets** — owned by the governance
[profiles contract](../../../skill/skill-agent-governance/agent-governance-guide/references/profiles.md):
`lifecycle · governance · executor · enforcement · data · telemetry · distribution`.

**Cross-cutting pairs** — each side varies independently; the coupling rule is what makes
the pair safe:

| Pair                         | Coupling rule                                                            |
| ---------------------------- | ------------------------------------------------------------------------ |
| Autonomy ⊥ oversight         | never raise autonomy without matching, observed oversight                |
| Autonomy ⊥ executor class    | raising autonomy never widens a task's permitted executor                |
| Team shape ⊥ everything else | solo / small-team / regulated is a profile input, never a platform level |

**Per-control enforcement dimensions** — the three questions of
[Choosing enforcement per control](#choosing-per-control): who must it bind · what failure
behavior · what impact.

**Agent sandbox confinement axes** — owned by the agent package
([`packages/agent/AGENTS.md`](../../../agent/AGENTS.md)), drawn in the sandbox-isolation
diagrams:

| Axis      | Values                  |
| --------- | ----------------------- |
| Isolation | `none · bwrap · docker` |
| Provision | `none · nix · command`  |
| Network   | `host · none · proxy`   |

plus the `hostPassthrough` knob and four independently-requestable fail-closed guarantees
(`RequirePinnedProvision`, `RequireHostToolsUnreachable`, `RequireEgressRestricted`,
`RequireKernelIsolation`) — filesystem reach and network egress are deliberately separate
axes: "host tools unreachable" ≠ "host unreachable".

**Not axes — the ordered scales:** autonomy `A0`–`A3` (oversight prerequisites), the module
authority ladder (least-privilege module selection), and isolation _strength_ within the
sandbox (an ordered spectrum whose sibling axes are orthogonal to it). The platform's
pattern in one line: everything independent is an axis you compose; everything ordered is a
ladder you must earn.

## Modes are named compositions

The five adoption modes are the platform's names for common compositions — presets, never
tiers — and every mode's absence report is part of its contract:

| Adoption mode             | Composition                                         | The absence you accept                                                        |
| ------------------------- | --------------------------------------------------- | ----------------------------------------------------------------------------- |
| Workflow-only             | Commands + profiles, selected methods and providers | No harness policy or external enforcement guarantee is claimed                |
| Governance-only           | Contracts + selected enforcement types              | Lifecycle results are not prerequisites                                       |
| Enablement-only           | Transactional setup and previews                    | No ongoing policy guarantee unless an enforcing module is separately selected |
| External-enforcement-only | CI, repository, deployment, and provider controls   | Agent hooks and lifecycle commands are optional                               |
| Integrated                | Any dependency-valid mix of the above               | Unselected modules remain explicit gaps, not hidden defaults                  |

**Validate, don't assume.** At any depth: `workflow-inspect` and
`workflow-governance-inspect` report the effective composition and its gaps;
`workflow-conformance` validates results, profiles, providers, and cross-plane requirements;
`workflow-drift` compares intended and observed state; the walking skeleton replays the whole
composition locally. And keep the platform's evidence discipline as your own: record what
was actually exercised, at which version, on which date — never report installed as
enforced, or documented as observed.

## See also

- [Architecture and composition](architecture-and-composition.md) — the two-planes model and
  effective profile composition
- [Developer quickstart](developer-quickstart.md) · [PM](pm-quickstart.md) ·
  [QA](qa-quickstart.md) · [UX](ux-quickstart.md) — the command surface per seat
- [Platform onboarding](platform-onboarding.md) — external platform owners and the
  transactional recipe
- [Security and policy](security-and-policy.md) · [Validation and traceability](validation-and-traceability.md)
- [`workflow-guide`](../../../skill/skill-workflow/workflow-guide/SKILL.md) and
  [`agent-governance-guide`](../../../skill/skill-agent-governance/agent-governance-guide/SKILL.md)
  — the contract owners every section defers to
