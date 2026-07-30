# Workspace Create

1. Resolve the exact source, source revision, target, supplied context, provider,
   applicable capabilities, and retry identity.
2. Preview the workspace resources that would be created.
3. On explicit `apply`, create only the previewed resources.
4. Verify and return the resulting workspace identity, source revision, canonical
   context, and observed effects.

For provider-native creation, require the supplied idempotency key when the provider
supports one. A local workspace may use its exact target and source revision as a
natural retry identity when repeated creation fails closed.

Do not implement work, merge, publish, abandon, or clean up.
