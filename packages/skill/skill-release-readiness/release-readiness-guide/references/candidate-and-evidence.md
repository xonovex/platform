# Candidate and Evidence

## Immutable Candidate

Record:

```text
Source revision and build:
Artifact identities, digests, and provenance:
Included and excluded changes:
Dependencies and lock state:
Runtime configuration and feature controls:
Schema, data, protocol, and infrastructure migrations:
Target environments and current revisions:
Release notes and known issues:
```

Verify that the artifact being assessed is the artifact proposed for rollout. Prefer
digest identity and verifiable provenance over mutable tags. Rebuilding creates a new
candidate.

## Evidence Matrix

Map each applicable release criterion to its result, evidence reference, subject
revision, environment, timestamp, owner, limitations, and exception. Include product
acceptance, test strategy, exploratory findings, accessibility, security, privacy,
performance, resilience, compatibility, data migration, observability, support, and
operational readiness as applicable.

Evidence may carry forward only when the exact relevant inputs and environment
assumptions are unchanged and the inheritance rationale is explicit. Record
unresolved defects separately from accepted exceptions and unknown evidence gaps.
