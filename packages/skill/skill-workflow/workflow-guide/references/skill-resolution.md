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

- [ ] Validate and normalize one `WorkflowRequest`.
- [ ] Ask the host composition runtime to resolve the request against every
      installed plugin root available to it. An explicit inventory may replace
      discovery.
- [ ] Resolve required semantic requirements before preferred support.
- [ ] Select one owner per duplicated check or procedure.
- [ ] Resolve `WorkflowRequest.selection.skillRequirements` against the actual
      installed inventory and retain every preferred failure as degradation.
- [ ] Compose `preferenceOverlays` from global to explicit scope using catalog
      precedence; reject equal or incomparable maximal-scope conflicts and
      non-preference guides.
- [ ] Treat domain or context conflicts as fact conflicts; compare authority,
      freshness, scope, and evidence.
- [ ] Select provider adapters from each resource binding, not from a global provider.
- [ ] Surface unavailable or conflicting selections instead of silently substituting.
- [ ] Block only when a missing selection is required by the operation or binding
      policy; otherwise report degraded advisory coverage.
- [ ] Load each selected guide exactly once in returned `loadOrder`; dependency
      guides precede their consumers. Do not continue the operation while composition
      is blocked.

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
policy. The skill contains only the declarative contract and packaged catalog; it
does not bundle executable resolver code. Xonovex platform hosts can call
`xonovex-workflow-compose` from shared tooling with the packaged catalog and actual
installed inventory. Raw prompt order is not a composition algorithm.
