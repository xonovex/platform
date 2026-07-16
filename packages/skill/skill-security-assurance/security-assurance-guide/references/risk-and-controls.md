# Risk and Control Selection

## Establish scope and context

Pin exact systems, components, data, identities, environments, dependencies, providers, revisions, lifecycle state, intended use, critical outcomes, legal or contractual context, and authority zones. Identify assets, entry and exit points, trust boundaries, attacker and insider capabilities, supply-chain paths, operational dependencies, and recovery assumptions.

Record confirmed facts, assumptions, unknowns, excluded scope, source versions, reviewers, and reassessment triggers. Unknown applicability or unobservable critical state remains a gap.

## Model credible risk

For each scenario, describe initiating conditions, threat actor or failure source, path, affected confidentiality/integrity/availability/privacy/safety outcome, likelihood or uncertainty, impact, detectability, existing controls, evidence, and owner. Include abuse, privilege escalation, injection, data exfiltration, tampering, dependency compromise, identity failure, denial of service, recovery failure, control bypass, and evidence loss as applicable.

Threat taxonomies help discovery; they do not replace system-specific paths or justify a numeric risk score by themselves.

## Select and tailor controls

Choose controls to meet explicit outcomes. For each control record:

- semantic identity, source and version, applicability, subject, owner, actor, and authority zone;
- preventive, detective, responsive, recovery, or compensating purpose;
- implementation owner and protected enforcement point;
- fail-closed, fail-visible, or advisory behavior plus dependency-outage behavior;
- evidence origin, exact subject/revision, evaluator version, freshness, limitations, and native references;
- dependencies, conflicts, exceptions, residual risk, verification, drift, update, rollback, and retirement.

Prefer defense in depth across independent layers. Do not count duplicated tools that share the same blind spot as independent controls.

## Map sources carefully

A crosswalk records source/target version, scope, relationship strength, rationale, gaps, conflicts, evidence expectation, assumptions, reviewer, and date. Controls and outcomes from different sources may overlap without being equivalent. Applicability, tailoring, organization-defined parameters, inherited controls, and implementation evidence remain explicit.

## Decide residual risk

An accountable authorized risk owner decides within declared authority using current evidence, known gaps, expiry, monitoring, incident, and reassessment conditions. An agent may assemble the record and recommend; it cannot create acceptance authority.

Exceptions and break-glass are separate scoped decisions. They do not edit the control baseline or become permanent defaults.
