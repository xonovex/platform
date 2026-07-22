---
name: credential-management-guide
description: "Use when choosing, storing, injecting, rotating, revoking, or responding to exposure of machine credentials and secrets. Triggers on API tokens, PATs, client secrets, keychains, secret managers, masked CI variables, workload identity federation, OIDC, `.env` secret handling, credential rotation, or leaked credentials, even when the user doesn't say 'credential management' and names only a provider-specific token."
---

# Credential Management Guidelines

Own the provider-neutral credential lifecycle. A provider guide still owns its token types, exact scopes, login command, request header, and verification probe.

## Essentials

- **Avoid stored secrets when possible** - Prefer a native login cache for a person and a short-lived workload identity or federated credential for automation, see [references/credential-selection.md](references/credential-selection.md)
- **Use one credential per integration** - Bind it to one non-human identity, environment, purpose, and minimum resource/operation scope
- **Keep a single source of truth** - Store local credentials in the platform secret store and shared credentials in an approved secret manager, see [references/local-storage.md](references/local-storage.md)
- **Resolve at the last responsible moment** - Read the value immediately before use, pass it through the provider-supported channel, and remove it from process state as soon as practical
- **Treat environment variables as transport** - They may be appropriate for a process or CI step, but are not a durable secret store
- **Design the full lifecycle** - Record owner, purpose, consumers, creation, expiry, rotation, revocation, audit source, and recovery before adoption
- **Assume masking can fail** - Never print, transform for debugging, place in URLs, or enable shell tracing around a secret
- **Revoke before cleanup** - On suspected exposure, invalidate the credential first; repository or log cleanup does not stop its use, see [references/automation-and-response.md](references/automation-and-response.md)

## Ownership boundary

Load the relevant provider skill for whether a PAT, app token, OAuth grant, service account, or another credential is valid; its exact permissions; host-specific environment variables; and the cheapest authenticated probe. This skill selects storage, delivery, automation, rotation, and incident behavior after those provider facts are known.

## Gotchas

- A secret encrypted at rest becomes plaintext when consumed; storage encryption does not make logs, child processes, crash dumps, or command arguments safe.
- A CI secret is available to any code that executes in that trusted job. Masking changes display, not authority.
- Fork, pull-request, reusable-workflow, and protected-environment rules differ by CI provider. Verify both allowed and denied paths.
- Rotating a value without updating every consumer causes an outage; updating consumers without revoking the old value leaves two valid credentials.
- Rewriting git history does not revoke a leaked credential and can destroy useful incident evidence if done before containment.

## Progressive Disclosure

- Read [references/credential-selection.md](references/credential-selection.md) - Load when choosing between interactive login, workload identity, dynamic credentials, service identities, and stored tokens
- Read [references/local-storage.md](references/local-storage.md) - Load when storing or retrieving a developer credential on macOS, Linux, Windows, WSL, or a local development fallback
- Read [references/automation-and-response.md](references/automation-and-response.md) - Load when injecting credentials into CI or services, defining rotation, or responding to a suspected leak
