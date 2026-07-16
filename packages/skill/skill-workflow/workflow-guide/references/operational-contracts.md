# Operational Authorization and Lifecycle Contracts

## Acceptance and authorization boundary

Acceptance evidence assembly is advisory and may use deterministic tools, bounded models,
or bounded agents. Accountable Acceptance is a separate human decision. Bind the decision
to the exact Deliverable Publication reference and revision, intended target, accountable
actor and role, evidence references and revisions, policy/profile version, conditions,
decision time, and expiry. A summary, label, comment, model response, or passing check is
not human authorization.

Before relying on Acceptance or another authorization, re-resolve every binding. Subject,
target, evidence, policy, profile, actor authority, or expiry drift invalidates the
authorization. Preserve the stale decision as history; never rewrite it to fit the current
request.

## Privileged-operation protocol

Integration, target-changing Transition and Release, data deletion, secret or credential
mutation, and Retirement use this sequence:

1. Resolve the exact source revision, starting target revision or resource identity,
   requested semantic intent, current policy/profile version, and fresh prerequisite
   evidence.
2. Resolve an authorization whose actor, scope, action, target, bindings, conditions, and
   expiry exactly cover the request. Evaluate any exception or break-glass record
   separately.
3. Select a provider or external-enforcement adapter that declares this semantic intent,
   blocks at the target boundary, cannot be bypassed by the governed actor, and fails with
   the profile-required behavior. An ordinary shell, write, merge, deploy, delete, or tool
   call is not this capability.
4. Obtain least-privilege, short-lived credentials only after authorization. Verify
   protected target/environment state, immutable artifacts or source revisions, rollback
   readiness, idempotency, cancellation, concurrency, and irreversible-action handling.
5. Ask the selected adapter to execute. Preserve separate native references for the policy
   decision, enforcement, mutation, and post-operation verification.
6. Re-resolve source and target state. Publish the canonical lifecycle result with exact
   revisions, actor/executor origin, outcome, partial failure, rollback or recovery, and
   residual gaps.

Mandatory operations fail closed on an absent, expired, stale, mismatched, invalid, or
unverifiable authorization; unavailable enforcement; credential failure; target drift;
policy drift; cancellation; timeout; or ambiguous provider outcome. Never convert these
states to success or fall back to a weaker local path.

## Semantic intent mapping

| Semantic intent                 | Harness point when supported                  | Required target-side/external point                                | Mandatory failure behavior                                                 |
| ------------------------------- | --------------------------------------------- | ------------------------------------------------------------------ | -------------------------------------------------------------------------- |
| Record human Acceptance         | Human-task or permission-request presentation | Identity-backed result/approval provider                           | Reject non-human or unverifiable identity                                  |
| Integrate accepted revision     | Pre-tool or pre-action direct-mutation guard  | Protected repository/target rule or provider operation             | Fail closed on stale bindings, denial, outage, or bypass                   |
| Execute Transition              | Pre-action scope and authorization guard      | Data, identity, feature, service, or provider change control       | Fail closed for protected mutation; preserve partial state                 |
| Release or deploy               | Pre-action artifact/target guard              | Controlled automation plus protected environment/deployment policy | Stop progression and roll back/recover on missed gate                      |
| Contain an Incident             | Emergency-action and secret/tool scope guard  | Authoritative access, runtime, identity, or provider control       | Deny out-of-scope/expired access; preserve urgent safe actions             |
| Delete data or retire resources | Pre-action irreversible-operation guard       | Provider permission, protected API, admission, or resource policy  | Fail closed until exact scope, holds, evidence, and verification are valid |

Harness adapters publish whether each point blocks, advises, observes, or is unsupported,
including ordering, concurrency, bypass, context, and tested version. A harness guard may
provide earlier defense and evidence but cannot replace a required external point the
governed actor cannot bypass.

## Exceptions and break-glass

An exception records owner, authorized human approver, exact subjects/actions/controls,
rationale, compensating controls, start, expiry, provider evidence, and mandatory review.
Break-glass additionally records the emergency reason, explicit invocation, authoritative
access evidence, notification, containment, revocation, and post-use review. Effective
scope cannot exceed the approver's authority.

Expired, ownerless, evidence-free, unreviewed, out-of-scope, or silently persistent
emergency access fails closed. Publish use and revocation evidence separately from the
underlying privileged lifecycle result.

## Agent assistance

Agents and models may summarize exact evidence, identify missing bindings, correlate
signals, draft timelines, propose containment or corrective actions, and investigate
within explicit tool, data, time, and authority bounds. Their outputs retain source
references and an `advisory` authority class.

They may not create or impersonate a human decision, approve their own work, claim that a
regulation applies, decide a report is legally required, suppress a required escalation,
or turn inference into external/provider evidence. Applicability and detailed legal or
regulated reporting conclusions require current authoritative text and qualified review.

## Independent lifecycle results

Acceptance, Integration, Transition, Release, Observation, Incident, Corrective Action,
Retirement, and Learning remain independently publishable and reconstructable from their
provider-native references. A successful earlier result does not imply a later one, and a
rollback creates new evidence rather than erasing the failed operation.

Every operation declares retry, idempotency, timeout, cancellation, partial-result,
rollback/recovery, evidence-retention, sensitive-data, and failure behavior. Provider-native
evidence remains authoritative only for its exact subject, scope, source, version, and
freshness.
