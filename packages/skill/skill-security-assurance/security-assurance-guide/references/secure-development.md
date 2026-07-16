# Secure Development Assurance

## Make security part of each change

Select practices appropriate to the product, threat model, impact, delivery model, and profile:

- define security requirements, misuse/abuse cases, data and identity boundaries, and verification before implementation;
- design least privilege, separation, safe defaults, validation, error handling, logging, secrets, dependency, isolation, update, and recovery behavior;
- review architecture and high-risk changes independently where required;
- use language/framework-specific secure implementation guidance through the owning skill;
- verify deterministic properties, unit/integration/system behavior, negative paths, fuzz/property tests, static/dynamic/composition scans, penetration or red-team cases, and operational recovery as applicable;
- triage vulnerabilities by exploitability, reachability, impact, exposure, compensating controls, and affected exact versions rather than score alone;
- protect source, reviews, CI, builders, artifacts, deployments, credentials, environments, and release authorization;
- monitor, patch, disclose, communicate, roll back, and retire safely.

## Preserve evidence at native owners

Return opaque source revision, review, build, test, finding, artifact, deployment, vulnerability, exception, and incident references with exact subjects and versions. Do not copy sensitive payloads into a universal report merely to normalize them.

Evidence remains valid only for the recorded subject, policy, evaluator, environment, and time. Re-run affected assurance when code, dependencies, build, configuration, data, permissions, provider capabilities, controls, or threat assumptions change.

## Make gates explainable

A gate reports the policy decision, enforcement action, evidence, reason, remediation, exception path, and failure behavior separately. A mandatory gate fails closed or requires independently adequate fresh evidence when its profile says so. Advisory findings fail visibly without pretending to block.

Avoid impossible gates that teams bypass: define ownership, service objectives, reliable tooling, actionable findings, bounded latency, support, emergency handling, and false-positive review.

## Retire vulnerable or obsolete behavior

Retirement verifies dependents, replacement coverage, data and secret handling, vulnerability/incident records, evidence retention, rollback, client migration, configuration removal, and absence of stale enforcement. Removing a scanner or policy without replacing required coverage is a security regression, even if the old tool was noisy.
