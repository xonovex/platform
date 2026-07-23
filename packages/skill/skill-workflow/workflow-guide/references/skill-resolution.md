# Skill and Capability Resolution

## Derived implementation

Resolve skills against the packaged
[composition catalog](../assets/composition-catalog.json). Its exact file-byte SHA-256
is the `catalogDigest`; its `contractVersion` and provision records are the only
taxonomy source for this skill.

Callers select work semantics. Derive implementation from:

- operation and named bindings;
- subject and result kinds;
- selected method and perspectives;
- binding and policy requirements;
- installed skills, provider adapters, and runtime tools.

Record a selected skill with guide, plugin, implementation version, resolver
provenance, catalog contract version and digest, reason, required status, and an
optional matched provision ID/version and requested range. Record package and source
paths only when the installed runtime exposes them. Record an unavailable,
incompatible, ambiguous, or conflicting selection with the requested identity or
requirement, candidates, message, blocking status, and the same catalog provenance;
never invent a concrete implementation. Keep provider/runtime capability records
separate. A skill explains knowledge or procedure; it never grants the corresponding
capability.

## Resolution procedure

- [ ] Resolve required semantic requirements before preferred support.
- [ ] Select one owner per duplicated check or procedure.
- [ ] Resolve `WorkflowRequest.selection.skillRequirements` against the actual
      installed inventory and retain every preferred failure as degradation.
- [ ] Compose `preferenceOverlays` from global to explicit scope using catalog
      precedence; reject equal-scope conflicts and non-preference guides.
- [ ] Treat domain or context conflicts as fact conflicts; compare authority,
      freshness, scope, and evidence.
- [ ] Select provider adapters from each resource binding, not from a global provider.
- [ ] Surface unavailable or conflicting selections instead of silently substituting.
- [ ] Block only when a missing selection is required by the operation or binding
      policy; otherwise report degraded advisory coverage.

Exact overrides belong only in `WorkflowRequest.implementationOverrides` and must
name both identifier and version. Skill overrides name a guide and an exact SemVer;
capability overrides use the provider/runtime's exact version identity. Reject
missing, ambiguous, incompatible, or unauthorized overrides. An override changes
implementation selection, never semantic criteria, effect scope, or authority.

## Composition roles

Durable domain and context skills explain meaning and environment. Preference skills
overlay explicit conventions. Procedure skills own methods. Capability-use skills
teach reliable mechanism use. Assurance skills add evidence and checks. Recovery
skills guide failure handling. Communication skills shape the human summary.

The runtime owns dependency resolution, sequence, conflict handling, and mandatory
policy. Use the shared composition resolver or its `xonovex-skill-compose` executable
with the packaged catalog, actual installed inventory, and normalized request; raw
prompt order is not a composition algorithm.
