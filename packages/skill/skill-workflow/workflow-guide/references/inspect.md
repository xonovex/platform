# Inspect an Effective Workflow

## Inputs

- One or more opaque native result references, or an in-memory result/handle.
- Selected or candidate workflow profile.
- Available provider, method, workspace, policy, and learning capabilities.
- Optional governance decisions and evidence references.

## Workflow

1. Resolve each native reference through the owning provider capability; do not infer provider state from paths or naming.
2. Reconstruct the ephemeral handle and identify canonical result kind, native revision, sources, and follow-up capabilities.
3. Resolve every variation axis independently and report explicit, profile, project, detected, or default origin.
4. Expand composite presentation into constituent canonical capabilities.
5. Evaluate result validity, publication boundaries, prerequisites, exit status, evidence freshness, authorization, and cumulative completion separately.
6. Report unsupported operations, missing references, stale evidence, ambiguous selections, and governance guarantee gaps without silently substituting providers or controls.

## Output

```text
Workflow inspection
Profile: <identity/version/source>
Capabilities: <canonical kind, status, native reference/revision>
Axes: <selection and origin per axis>
Publication and pickup boundaries: <effective topology>
Evidence and authorization: <valid/stale/missing>
Completion: <not-evaluated/incomplete/complete with reasons>
Gaps: <actionable findings>
```

Keep native references opaque in the report and redact sensitive provider context. A human-readable report is the default; structured output is an optional view, never the contract itself.
