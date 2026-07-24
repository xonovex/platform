# Workspace Merge

1. Resolve the exact workspace revision, destination revision, provider, integration
   criteria, and applicable capabilities.
2. Validate the criteria and preview the integration effect.
3. On explicit `apply`, integrate only the validated workspace.
4. Verify and return the destination revision and every observed effect.

Preserve the workspace, branch, provider reference, and metadata. Cleanup remains a
separate operation after successful integration.
