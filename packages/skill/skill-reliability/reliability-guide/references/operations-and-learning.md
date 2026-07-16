# Operations, Incidents, Learning, and Retirement

## Define ownership before operation

Every service, policy, module, dependency, data flow, telemetry path, recovery mechanism, and exception has an owner, support status, service expectation, escalation, security contact, change authority, documentation, runbooks, provider dependencies, maintenance window, and retirement path.

Unsupported or ownerless critical components fail profile validation; they do not become reliable because they have not failed recently.

## Prepare and run incidents

Define detection, severity, command, roles, communication, accessible channels, evidence protection, containment, continuity, recovery, notification decision, external coordination, and closure criteria. During response, record exact affected versions, time window, decisions, authority, actions, provider-native evidence, impact, unknowns, and next update.

Verify recovery against user and control outcomes, not only green infrastructure. Preserve failed operations and successful recovery as separate results.

## Close corrective actions on effectiveness

Each corrective action records source incident/finding, root and contributing conditions, owner, target version, planned control change, risk, due date, verification, effectiveness window, rollback, evidence, and closure authority. Closing a ticket or deploying a change is not effectiveness proof.

## Learn without auto-promoting

Capture candidate lessons from lifecycle work, onboarding, denials, incidents, exceptions, drift, rollbacks, support, and module failures. Deduplicate, identify scope, test the causal claim, assess privacy and incentives, route to the owning policy/skill/module, and reject one-off noise.

Promotion is reviewed, versioned, conflict-checked, canaried where executable, authorized, measurable, reversible, and removable. Store provenance from lesson to change and rollback. Never change global instructions or enforcement automatically from one metric or incident.

## Use balanced operational metrics

Combine outcome reliability, control coverage, evidence freshness, false positives/negatives, bypass attempts, drift age, exception age, change failure, recovery and rollback success, onboarding time, user friction, accessibility, security/privacy incidents, cost, and DORA outcomes where relevant.

Pair each measure with purpose, audience, source, quality, aggregation, retention, interpretation limits, likely gaming behavior, counter-metrics, and qualitative review. Do not rank individuals or make automatic punitive decisions.

## Retire safely

Inventory consumers, dependencies, alternatives, mandatory coverage, data/evidence retention, credentials, provider configuration, support, and rollback. Announce and migrate, stop new adoption, disable safely, verify authoritative removal, revoke access, preserve required evidence, monitor for stale callers/configuration, and close only after replacement and recovery paths are proven.
