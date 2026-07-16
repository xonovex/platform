# inventory-generate: Generate an Authoritative Inventory

## Core workflow

1. Resolve the exact subject reference/revision, selected inventory scope and specialization, provider capabilities, and applicable profile requirements.
2. Select deterministic generators and authoritative package, build, model, dataset, cryptographic, service, or environment sources. Record generator identity/version, inputs, exclusions, and failure behavior.
3. Generate component identities, versions/digests, relationships, provenance, observed state, requested/effective permissions where selected, and explicit gaps. Do not guess an unobservable component or version.
4. Optionally use a bounded model to enrich non-authoritative descriptions only. Mark inference, link its source facts, validate its output, and never let it alter identities, versions, digests, relationships, provenance, or completeness claims.
5. Validate the selected interoperable representation through its owning tool/provider and preserve its native reference. Supported specializations include SBOM, AIBOM/AI-SBOM, ML-BOM, CBOM, service inventory, and agent-environment inventory; a profile pins the exact specification/schema where one is required.
6. Publish an Inventory result with exact scope/subject revision, components, relationships, provenance, gaps, generator/evidence origins, optional native BOM references, and follow-up capabilities.

An agent-environment inventory may include skills, plugins, hooks, extensions, MCP servers, CI components, policy bundles, models, tools, secrets access classes, and requested/effective permissions. Missing categories remain gaps, not proof that no components exist.
