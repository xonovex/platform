# Automation, Rotation, and Exposure Response

## CI and services

Prefer a job-scoped workload identity or managed identity. Constrain its trust policy to the intended repository, workflow, ref, environment, audience, and subject, then grant only the target permissions needed by that job.

When a stored secret is unavoidable:

- keep it in the CI provider's protected secret store or an external secret manager, never pipeline YAML;
- expose it only to the trusted job and environment that consumes it;
- verify fork, pull-request, protected-branch, reusable-workflow, and rerun behavior with negative probes;
- fetch it at runtime through a non-human identity with read access to that secret only;
- prevent tracing and treat log masking as best-effort;
- record every consumer so rotation and revocation are complete.

Production services should retrieve a secret at startup or through a bounded refresh path. Pin a version when reproducibility requires it; otherwise define how a rotated current version is adopted and how failure rolls back.

## Rotation

Create the replacement, verify it through a read-only probe, update consumers in a controlled overlap window, verify each consumer, revoke the old value, and confirm the old value fails. Automate the sequence when the provider supports it. An expiry date is a backstop, not proof that rotation works.

## Suspected exposure

1. Revoke or disable the credential at its issuer immediately.
2. Preserve timestamps, audit events, affected identity and scopes, and known consumers.
3. Issue a replacement with reduced authority when possible and update consumers from the source of truth.
4. Search logs, artifacts, caches, build outputs, tickets, and repository history for secondary exposure without printing the value again.
5. Remove exposed material only after containment and evidence preservation; coordinate any history rewrite because it changes shared commit identities.
6. Verify the revoked value is rejected, review its audit trail for misuse, and fix the path that exposed it.
