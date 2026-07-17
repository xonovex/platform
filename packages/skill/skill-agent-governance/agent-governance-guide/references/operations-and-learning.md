# Governance Operations, Effectiveness, and Learning

## Own every operational module

Record owner, support status, service expectations, escalation/on-call, security contact, change authority, documentation/runbooks, dependency/provider contacts, evidence retention, incident plan, update channel, emergency disable, and retirement plan. Ownerless or unsupported modules cannot satisfy mandatory organization profiles.

## Update, disable, and retire safely

For every governance module or policy change:

1. pin current, candidate, and rollback versions plus provenance and compatibility;
2. preview policy/control/capability/permission/data-flow/failure changes and conflicts;
3. test deterministic fixtures, intended allows/denials, concurrency, trust, data leakage, stale versions, missing enforcement, exception abuse, telemetry outage, rollback, and fresh-context recovery;
4. canary representative adoption modes, authority zones, platforms, and risk levels with explicit success, observation, abort, and promotion criteria;
5. authorize, apply idempotently, verify authoritative native state, monitor, and record evidence;
6. promote gradually, roll back, or disable; preserve an independently reachable emergency path and required control coverage.

Disable stops execution; rollback restores a pinned prior version; retirement removes adoption after migration and evidence retention. Keep them separate.

## Detect drift across the plane

Compare intended and observed harness configuration, CI policy, module/source versions, provider capability/edition/tier, profile/control/source versions, managed settings, permissions, data flows, evidence freshness, and support/lifecycle state. Classify strengthening, neutral, weakening, unsupported, stale, and unknown changes.

Fail visibly when mandatory guarantees become invalid. Never auto-promote drift into desired policy; remediation follows preview, authorization, apply, verify, and rollback.

## Extract learning candidates

Capture candidates from lifecycle results, onboarding, policy denials, incidents, exceptions, emergency exceptions, drift, update/rollback, support, and module failures. Each candidate records source evidence, scope, owner, affected outcomes, proposed owning skill/policy/module, expected benefit, risks, privacy/incentive limits, conflicts, expiry, and review status.

Deduplicate and test the causal claim. Route general guidance to its owning skill, organization-specific rules to governed policy/instructions, and executable changes to their module. Do not create a new owner for one weak observation.

Promotion is reviewed, versioned, conflict-checked, canaried where executable, authorized, measurable, reversible, and removable. Preserve candidate → decision → version → evidence → rollback provenance. Never auto-promote one denial, incident, exception, metric, or model inference into global instructions or enforcement.

Use **reflect-guide** for extracting reusable session lessons; its promotion must obey this governance gate when the target changes managed or executable behavior.

## Measure effectiveness without surveillance

Use balanced system-level measures such as control/enforcement coverage, evidence freshness, false positives/negatives, bypass attempts, exception/emergency-exception age, drift age, update/rollback/recovery success, incident outcomes, onboarding time, user friction, accessibility, cost, and DORA outcomes where relevant.

Each measure declares purpose, audience, source/quality, aggregation/minimum cohort, access, retention/deletion, interpretation limits, prohibited uses, likely gaming behavior, counter-metrics, and qualitative review. Never rank individuals, infer productivity from activity, or auto-punish/auto-promote from a metric.

## Respond to governance incidents

Treat control bypass, unauthorized weakening, permission expansion, supply-chain compromise, sensitive logging, stale mandatory policy, broken rollback/emergency disable, exception abuse, evidence tampering, and widespread false denial as governance incidents according to impact. Preserve exact versions, contain, restore safe coverage, communicate, verify recovery, create corrective actions, and retire obsolete modules through normal change authority.
