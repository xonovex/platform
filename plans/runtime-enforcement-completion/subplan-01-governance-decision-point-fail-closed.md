---
type: plan
has_subplans: false
parent_plan: ../runtime-enforcement-completion.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
    - packages/skill/skill-workflow/workflow-guide/scripts/*.mjs
    - packages/skill/skill-agent-governance/agent-governance-guide/scripts/*.mjs
    - packages/agent/agent-operator-go/api/v1alpha1/agentpolicy_types.go
    - packages/agent/agent-operator-go/internal/webhook/agentrun_webhook.go
    - packages/agent/agent-operator-go/internal/webhook/*_test.go
skills_to_consult:
  - plan-guide
  - hexagonal-pattern-guide
  - microkernel-pattern-guide
  - typescript-guide
  - kubernetes-guide
  - testing-guide
  - connascence-guide
validation:
  type_check: not_run
  lint: not_run
  build: not_run
  tests: not_run
updated: "2026-07-17"
---

# Governance Decision Point That Invokes the Validators and Fails Closed

## Objective

Make the reference validators load-bearing. A running decision point evaluates a real lifecycle
operation with `checkIndependence` / `validateEmergencyAccess` / `validatePrivilegedOperation` /
`selectDevelopmentExecutor`, returns an allow/deny verdict with the exact failure code, records the
verdict as provider-native evidence, and fails closed on any error or missing input. The verdict is
added to the operator's already-live admission surface as a second, independent deny path — the
existing infra `enforcePolicy` stays intact (defense in depth). This is the foundation the harness
hook (Phase 2), the runtime-real walking skeleton (Phase 3), and A3 orchestration (Phase 4) all gate
through.

## Tasks

1. Resolve the Open decision below (language seam) and record it in this subplan, then scaffold a
   **policy decision service** beside the validators — new `.mjs` under
   `packages/skill/skill-workflow/workflow-guide/scripts/` and
   `packages/skill/skill-agent-governance/agent-governance-guide/scripts/` — that imports
   `checkIndependence` (`workflow-guide/scripts/independence-helpers.mjs:34`),
   `validateEmergencyAccess` and `validatePrivilegedOperation`
   (`workflow-guide/scripts/operational-lifecycle-helpers.mjs:186,247`), and `selectDevelopmentExecutor`
   / `validateDevelopment` (`workflow-guide/scripts/development-assurance-helpers.mjs:33,65`) **without
   forking them**, exposing one stable request/response contract (subject reference, operation, inputs
   → decision, exact failure code, policy version, correlation id).
2. Emit a **verdict evidence record** on every check (allow or deny) reusing the walking skeleton's
   `record_evidence` shape as the schema seed
   (`agent-governance-guide/assets/walking-skeleton/run-skeleton.sh:44-52,62-68`): decision, failure
   code, policy version, subject/correlation id; idempotent per correlation id.
3. Fail closed in the decision service: any validator error, unreachable service, or absent required
   input **denies** a mandatory-profile operation and **observes** an advisory one, mirroring
   `agent-governance-guide/assets/walking-skeleton/guard.sh:34-41`.
4. Extend `AgentPolicyEnforced` (`api/v1alpha1/agentpolicy_types.go:13-41`) with a governance field
   (e.g. `RequireGovernanceVerdict`) and model a governed lifecycle operation carried on an `AgentRun`
   annotation/label; hand-maintain the `zz_generated.deepcopy.go` and CRD manifest edits per the
   operator's controller-gen constraint (`agent-operator-go/AGENTS.md`, controller-gen broken on Go 1.25+).
5. Add the governance verdict as a **second, independent deny path** in the AgentRun webhook alongside
   `enforcePolicy` (`internal/webhook/agentrun_webhook.go:216-304`), at the live admission point wired
   by `cmd/operator/main.go:79` (`(&webhook.AgentRunWebhook{}).SetupWebhookWithManager`): call the
   decision service, surface the failure code in the admission response, and keep the existing
   pod/infra `enforcePolicy` untouched.
6. Add a **governance verdict test** beside `internal/webhook/agentpolicy_verdict_test.go` — a table of
   governed operations (independent vs self-approved release; in-scope vs expired emergency access;
   mechanical vs adaptive executor request) asserting the exact failure code and the accept/deny
   verdict, plus a decision-service-outage case (mandatory denies, advisory observes).
7. Add a decision-service unit test / `validate-*` wiring under the two `scripts/` directories so the
   validator symbols gain a non-fixture, non-vocabulary importer, and confirm the existing
   `agent-governance-guide/scripts/validate-executor-vocabulary.mjs` still passes.

## Acceptance criteria

- A repo-wide grep for `checkIndependence`, `validateEmergencyAccess`, `validatePrivilegedOperation`,
  and `selectDevelopmentExecutor` across `*.mjs`/`*.ts`/`*.js`/`*.go` (excluding `node_modules`/`dist`)
  shows at least one importer that is neither a `validate-*-fixtures.mjs` nor
  `validate-executor-vocabulary.mjs`: the new decision service.
- Running the decision point against a self-approved release returns `deny` with the exact independence
  / emergency-access failure code, and against a compliant one returns `allow` — both asserted by an
  automated test.
- A webhook test proves the admission path denies a governed operation whose verdict is `deny` (failure
  code present in the admission response) and admits a passing one, while the existing
  `agentpolicy_verdict_test.go` infra verdicts remain green.
- Every check writes exactly one verdict evidence record per correlation id (re-running the same
  correlation id does not append a duplicate); a simulated decision-service outage denies a mandatory
  operation and observes an advisory one.
- `go test ./...` in `packages/agent/agent-operator-go` and the workflow/governance `validate-*.mjs`
  scripts pass with no regression.

## Dependencies

None — this is parallel group 1 and the load-bearing foundation. The parent orders the phases by
dependency: "Phase 1 (a runtime that invokes the validators and records a verdict) ... [is] the
load-bearing foundation" and every later phase gates through this verdict.

## Open decisions

- **Language seam (Go operator vs Node validators).** The validators are `.mjs`; the live enforcement
  point is Go. The parent leaves this explicitly unresolved: "**Open — must be decided before Phase 1.**" The options it records are (a) a Node policy-decision service the operator calls over a local
  contract, (b) port the validators to Go (rejected by the no-fork constraint — risks divergence from
  the authoritative source), or (c) embed a JS runtime. The parent's recommended default is **(a)**,
  keeping decision and enforcement decoupled per `composable-workflow-phases.md` plane 2. This subplan
  is written against (a); confirm or override before starting Task 1.
