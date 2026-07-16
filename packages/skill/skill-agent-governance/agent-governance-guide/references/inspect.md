# Inspect Effective Governance

## Workflow

1. Discover effective organization, project, user, session, and external sources through native read-only capabilities.
2. Resolve selected profiles, policy versions, modules, events, capability matrices, actors, exceptions, and evidence references.
3. Calculate effective authority and additive strengthening; flag weakening, conflicts, or unsupported assumptions.
4. Match every mandatory requirement to an adequate enforcement point and explicit failure behavior.
5. Check executor bounds, module provenance, data handling, concurrency/reentrancy, evidence origin, and freshness.
6. Separate policy decision, enforcement action, and evidence in the report.

When an `Inventory` result is available, include installed skills, plugins, hooks, extensions, MCP servers, CI components, policy bundles, models, tools, and permissions by native identity/version and provenance. Treat missing categories as explicit inventory gaps, not proof that no component exists.

## Output

```text
Governance inspection
Scope and authority sources: <native references/versions>
Profiles and policies: <effective versions/applicability>
Modules and executors: <state, bounds, permissions, provenance>
Enforcement: <intent -> native point -> guarantee/failure behavior>
Evidence and exceptions: <origin, freshness, expiry>
Drift and conflicts: <findings>
Verdict: conformant | advisory-only | incomplete | non-conformant
Remediation: <actions without silent mutation>
```

Redact secrets and sensitive content. Report opaque native references and provenance rather than copying raw prompts, tool outputs, or credentials.
