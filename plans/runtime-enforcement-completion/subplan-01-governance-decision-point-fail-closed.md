---
type: plan
has_subplans: false
parent_plan: ../runtime-enforcement-completion.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
    - package.json
    - packages/skill/skill-workflow/workflow-guide/scripts/*.mjs
    - packages/skill/skill-workflow/workflow-guide/scripts/*.ts
    - packages/skill/skill-workflow/package.json
    - packages/skill/skill-agent-governance/agent-governance-guide/scripts/*.mjs
    - packages/skill/skill-agent-governance/agent-governance-guide/scripts/*.ts
    - packages/skill/skill-agent-governance/package.json
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
existing infra `enforcePolicy` stays intact (defense in depth). Per parent Decision 1 the service
and the migrated validators are TypeScript (erasable syntax, run natively by Node `>=22.18.0`).
This is the foundation the harness hook (Phase 2), the runtime-real walking skeleton (Phase 3),
and A3 orchestration (Phase 4) all gate through.

## Tasks

1. Per the settled Decision below: raise the root `package.json` `engines` Node floor from
   `>=20.0.0` to `>=22.18.0` (target Node 24 LTS), migrate the validator helper files to
   erasable-syntax TypeScript run directly via native type stripping (updating the skills'
   `package.json` `test` scripts to invoke the `.ts` files), then scaffold a **TypeScript policy
   decision service** beside the validators — new `.ts` under
   `packages/skill/skill-workflow/workflow-guide/scripts/` and
   `packages/skill/skill-agent-governance/agent-governance-guide/scripts/` — that imports
   `checkIndependence` (`workflow-guide/scripts/independence-helpers`),
   `validateEmergencyAccess` and `validatePrivilegedOperation`
   (`workflow-guide/scripts/operational-lifecycle-helpers`), and `selectDevelopmentExecutor`
   / `validateDevelopment` (`workflow-guide/scripts/development-assurance-helpers`) **without
   forking them**, exposing one long-running service with a stable, versioned JSON
   request/response contract (subject reference, operation, inputs → decision, exact failure
   code, policy version, correlation id).
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
   decision service through a configurable endpoint (operator flag/env, e.g.
   `--decision-service-url`, defaulting to the pod-local sidecar per parent Decision 2;
   unreachable ⇒ deny mandatory), surface the failure code in the admission response, and keep the
   existing pod/infra `enforcePolicy` untouched.
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
- `go test ./...` in `packages/agent/agent-operator-go` and the workflow/governance `validate-*`
  scripts pass with no regression.
- The root `engines` floor is `>=22.18.0`, and the validator helpers plus decision service are
  `.ts` executed directly by `node` with no build step (the skills' `test` scripts invoke them).

## Dependencies

None — this is parallel group 1 and the load-bearing foundation. The parent orders the phases by
dependency: "Phase 1 (a runtime that invokes the validators and records a verdict) ... [is] the
load-bearing foundation" and every later phase gates through this verdict.

## Decisions (settled 2026-07-17)

- **Language seam: option (a), in TypeScript** (parent Decision 1). A long-running TypeScript
  policy-decision service over a versioned JSON contract; the validators migrate `.mjs` →
  erasable-syntax `.ts` run natively (root `engines` floor `>=22.18.0`, target Node 24 LTS); no
  forking, no port to Go, no embedded JS engine.
- **Endpoint topology** (parent Decision 2): the webhook reaches the service through a configurable
  endpoint defaulting to the pod-local sidecar; an unreachable endpoint denies mandatory operations
  regardless of topology.
