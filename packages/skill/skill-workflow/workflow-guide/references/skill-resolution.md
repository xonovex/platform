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

Record every selected skill with guide, plugin, implementation version, package and
source paths, resolver provenance, catalog contract version and digest, reason,
required status, availability, and an optional matched provision ID/version and
requested range. Resolver provenance is one of explicit, exact dependency, semantic
requirement, or policy. Keep provider/runtime capability records separate. A skill
explains knowledge or procedure; it never grants the corresponding capability.

## Resolution procedure

- [ ] Resolve required semantic requirements before preferred support.
- [ ] Select one owner per duplicated check or procedure.
- [ ] Compose preference overlays only through explicit scope precedence.
- [ ] Treat domain or context conflicts as fact conflicts; compare authority,
      freshness, scope, and evidence.
- [ ] Select provider adapters from each resource binding, not from a global provider.
- [ ] Surface unavailable or conflicting selections instead of silently substituting.
- [ ] Block only when a missing selection is required by the operation or binding
      policy; otherwise report degraded advisory coverage.

Exact overrides belong only in `WorkflowRequest.implementationOverrides` and must
name both identifier and version. Reject missing, ambiguous, incompatible, or
unauthorized overrides. An override changes implementation selection, never semantic
criteria, effect scope, or authority.

## Composition roles

Durable domain and context skills explain meaning and environment. Preference skills
overlay explicit conventions. Procedure skills own methods. Capability-use skills
teach reliable mechanism use. Assurance skills add evidence and checks. Recovery
skills guide failure handling. Communication skills shape the human summary.

The runtime owns dependency resolution, sequence, conflict handling, and mandatory
policy; raw prompt order is not a composition algorithm.
