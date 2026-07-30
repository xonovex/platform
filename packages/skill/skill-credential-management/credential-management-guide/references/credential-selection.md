# Credential Selection

## Selection order

1. Use the provider's interactive browser or device login for a person when its refresh cache is native, protected, and revocable.
2. Use workload identity federation, a managed identity, or another short-lived exchange for automation when the target supports it. Bind trust to exact repository, workflow, ref, environment, audience, and subject claims.
3. Use dynamically issued credentials with a bounded lease when federation is unavailable but a broker can mint them.
4. Use a stored token or client secret only when the operation cannot use the earlier choices. Give it one integration, the shortest practical expiry, and only required resources and operations.

Do not claim two mechanisms are interchangeable. A provider PAT may authorize its own API while an OIDC exchange authorizes a cloud role; use the provider guide to confirm the actual target and operation.

## Decision record

Record the target, operation, identity, credential class, issuer, audience, resource and operation scope, creation authority, lifetime, storage owner, consumers, verification probe, rotation overlap, revocation path, and audit source. State why every stronger short-lived option is unavailable before approving a stored secret.

## Verification

Use a cheap authenticated read against the intended host and resource before any write. Distinguish authentication failure from insufficient permission and wrong host or resource. Never print the credential while diagnosing.
