# Workspace Abandon

1. Resolve the exact workspace, optional workspace revision, provider, stopping
   reason, and applicable capabilities.
2. Inspect the workspace and collect its partial state and recovery information.
3. Record unresolved work and a safe retry boundary.
4. Return the abandonment record inline.

Preserve the workspace and every external resource. Do not snapshot, delete, prune, or
change provider status.
