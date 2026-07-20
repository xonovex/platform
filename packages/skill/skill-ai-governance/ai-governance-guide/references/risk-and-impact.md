# Risk and Impact Management

## Establish context and affected people

Describe intended benefits, decisions influenced, autonomy, scale, duration, reversibility, environments, dependencies, misuse, reasonably foreseeable use, and people or groups affected directly and indirectly. Include people who cannot use, contest, opt out of, or recover from the system.

Use participatory input where appropriate with consent, accessibility, compensation, privacy, and safe escalation. Do not collect sensitive attributes merely to make a governance dashboard look complete.

## Build a versioned risk record

For each risk, record:

- scenario, initiating conditions, affected people/assets/rights, harm, severity, likelihood or uncertainty, and detectability;
- supporting evidence, assumptions, disagreements, data or evaluation gaps, and source versions;
- preventive, detective, responsive, and recovery controls with owners and enforcement points;
- control limitations, verification evidence, residual risk, acceptance authority, expiry, and reassessment trigger.

Cover product and operational risks such as invalid output, discrimination, exclusion, privacy loss, data leakage, security abuse, prompt injection, tool misuse, excessive agency, automation bias, opacity, manipulation, unsafe content, intellectual-property or provenance gaps, provider concentration, resource/cost abuse, and environmental or workforce effects where relevant.

## Select controls by outcome

Prefer controls that prevent or bound harm at the strongest available layer: product constraints, data governance, least privilege, deterministic validation, human authorization, protected provider operations, monitoring, rollback, and safe shutdown. Prompts and user instructions are advisory controls unless an adequate enforcement point guarantees them.

Every mandatory control declares fail-closed, fail-visible, or advisory behavior. Telemetry failure alone does not automatically block unless the selected profile requires that evidence and no adequate alternative exists.

## Accept residual risk explicitly

Only an authorized accountable role accepts residual risk within its authority. Bind the decision to exact system and evidence versions, scope, conditions, expiry, monitoring, intervention, incident, reassessment, and rollback. A model score, completed template, or passing policy check is not residual-risk acceptance.

Reassess after material changes to purpose, users, model, data, prompts, retrieval, tools, autonomy, provider, permissions, jurisdiction, environment, incidents, harmful outcomes, control effectiveness, or source requirements.
