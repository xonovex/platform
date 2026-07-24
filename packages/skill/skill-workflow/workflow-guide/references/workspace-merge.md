# Workspace Merge

1. Resolve the exact workspace revision, destination revision, supplied context,
   provider, integration criteria, applicable capabilities, and retry identity.
2. Validate the criteria and preview the integration effect.
3. On explicit `apply`, integrate only the validated workspace.
4. Verify and return the destination revision, canonical context, and every observed
   effect.

For provider-native integration, require the expected destination revision and
supplied idempotency key when supported. Reconcile an unknown integration result before
retrying.

Preserve the workspace, branch, provider reference, and metadata. Cleanup remains a
separate operation after successful integration.
